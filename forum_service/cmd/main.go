package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	grpcdelivery "github.com/kprf42/dolgova/forum_service/internal/delivery/grpcdel"
	httpdelivery "github.com/kprf42/dolgova/forum_service/internal/delivery/http"
	"github.com/kprf42/dolgova/forum_service/internal/delivery/http/handlers"
	"github.com/kprf42/dolgova/forum_service/internal/delivery/websocket"
	"github.com/kprf42/dolgova/forum_service/internal/repository"
	chat "github.com/kprf42/dolgova/forum_service/internal/usecase"
	comment "github.com/kprf42/dolgova/forum_service/internal/usecase"
	post "github.com/kprf42/dolgova/forum_service/internal/usecase"
	"github.com/kprf42/dolgova/pkg/logger"
	"github.com/kprf42/dolgova/proto/forum"
	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"
)

// JWTClaims кастомная структура claims с реализацией всех необходимых методов
type JWTClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// AuthMiddleware структура для middleware аутентификации
type AuthMiddleware struct {
	JWTSecret string
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *AuthMiddleware) JWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("\n=== JWT Middleware ===\n")
		fmt.Printf("Request URL: %s\n", r.URL.String())
		fmt.Printf("Request Method: %s\n", r.Method)

		if r.Method == "OPTIONS" {
			fmt.Printf("OPTIONS request - skipping auth\n")
			w.WriteHeader(http.StatusOK)
			return
		}

		authHeader := r.Header.Get("Authorization")
		fmt.Printf("Authorization header: '%s'\n", authHeader)

		if authHeader == "" {
			fmt.Printf("ERROR: No Authorization header\n")
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			fmt.Printf("ERROR: No Bearer prefix in token\n")
			http.Error(w, "Bearer token required", http.StatusUnauthorized)
			return
		}
		fmt.Printf("Token string after trim: '%s'\n", tokenString)

		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.JWTSecret), nil
		})

		if err != nil {
			fmt.Printf("ERROR: Token parse error: %v\n", err)
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			fmt.Printf("ERROR: Token is invalid\n")
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			fmt.Printf("ERROR: Invalid token claims type\n")
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		fmt.Printf("Token claims: %+v\n", claims)
		fmt.Printf("User ID from token: %s\n", claims.UserID)

		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		fmt.Printf("Added user_id to context: %s\n", claims.UserID)
		fmt.Printf("=== End JWT Middleware ===\n\n")

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func main() {
	// Инициализация логгера
	log, err := logger.New()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer log.Sync()

	// Загрузка конфигурации
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal("Failed to load config", logger.Error(err))
	}

	// Подключение к существующей базе данных auth сервиса
	dbPath := filepath.Join("..", "auth_service", "auth.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("Failed to connect to database", logger.Error(err))
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error("Failed to close database connection", logger.Error(err))
		}
	}()
	db.SetMaxOpenConns(1)

	// Проверка соединения с БД
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database", logger.Error(err))
	}

	// Применение миграций форумного сервиса
	if err := runForumMigrations(db, log); err != nil {
		log.Fatal("Failed to apply forum migrations", logger.Error(err))
	}

	// Инициализация репозиториев
	postRepo := repository.NewPostRepository(db, log)
	commentRepo := repository.NewCommentRepository(db, log)
	chatRepo := repository.NewChatRepository(db, log)

	// Инициализация use cases
	postUC := post.NewPostUseCase(postRepo, log)
	commentUC := comment.NewCommentUseCase(commentRepo, log)
	chatUC := chat.NewChatUseCase(chatRepo, log)

	// Инициализация WebSocket Hub
	hub := websocket.NewHub(chatUC)
	go hub.Run()

	// Инициализация обработчиков
	postHandlers := handlers.NewPostHandlers(postUC)
	commentHandlers := handlers.NewCommentHandlers(commentUC)
	chatHandlers := handlers.NewChatHandlers(hub, chatUC)

	// Создание HTTP роутера
	router := httpdelivery.NewRouter(postHandlers, commentHandlers, chatHandlers, cfg.JWTSecret)

	// Настройка HTTP сервера
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Настройка gRPC сервера
	grpcServer := grpc.NewServer()
	forum.RegisterForumServiceServer(grpcServer, grpcdelivery.NewForumServer(postUC, commentUC, chatUC))

	// Запуск серверов
	go startHTTPServer(httpServer, cfg.HTTPPort, log)
	go startGRPCServer(grpcServer, cfg.GRPCPort, log)

	// Ожидание сигнала завершения
	waitForShutdownSignal(httpServer, grpcServer, log)
}

type Config struct {
	HTTPPort  int
	GRPCPort  int
	JWTSecret string
}

func loadConfig() (*Config, error) {
	return &Config{
		HTTPPort:  8081,
		GRPCPort:  50051,
		JWTSecret: "your-strong-secret-key",
	}, nil
}

func runForumMigrations(db *sql.DB, log *logger.Logger) error {
	log.Info("Applying forum service migrations")

	// Получаем абсолютный путь к миграциям из auth сервиса
	migrationsPath := filepath.Join("..", "auth_service", "migrations")
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Проверяем существование папки
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("auth service migrations directory does not exist: %s", absPath)
	}

	// Инициализируем драйвер SQLite
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Формируем URL для миграций
	migrationsURL := "file://" + filepath.ToSlash(absPath)

	// Создаем экземпляр мигратора
	m, err := migrate.NewWithDatabaseInstance(
		migrationsURL,
		"sqlite3",
		driver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	// Применяем миграции
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply forum migrations: %w", err)
	}

	log.Info("Forum service migrations applied successfully")
	return nil
}

func startHTTPServer(server *http.Server, port int, log *logger.Logger) {
	log.Info("Starting HTTP server", logger.Int("port", port))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("HTTP server error", logger.Error(err))
	}
}

func startGRPCServer(server *grpc.Server, port int, log *logger.Logger) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal("Failed to listen gRPC", logger.Error(err))
	}

	log.Info("Starting gRPC server", logger.Int("port", port))
	if err := server.Serve(listener); err != nil {
		log.Fatal("gRPC server error", logger.Error(err))
	}
}

func waitForShutdownSignal(httpServer *http.Server, grpcServer *grpc.Server, log *logger.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down servers...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error("HTTP server shutdown error", logger.Error(err))
	}

	grpcServer.GracefulStop()
	log.Info("Servers stopped gracefully")
}

func NewRouter(
	postHandlers *handlers.PostHandlers,
	commentHandlers *handlers.CommentHandlers,
	chatHandlers *handlers.ChatHandlers,
	jwtSecret string,
) *chi.Mux {
	return httpdelivery.NewRouter(postHandlers, commentHandlers, chatHandlers, jwtSecret)
}
