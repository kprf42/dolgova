package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/kprf42/dolgova/auth_service/internal/config"
	myHttp "github.com/kprf42/dolgova/auth_service/internal/delivery/http"
	"github.com/kprf42/dolgova/auth_service/internal/repository"
	"github.com/kprf42/dolgova/auth_service/internal/usecase/auth"
	"github.com/kprf42/dolgova/auth_service/internal/usecase/jwt"
	"github.com/kprf42/dolgova/pkg/logger"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Инициализация логгера
	log, err := logger.New()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer log.Sync()

	log.Info("Starting auth service initialization")

	// Загрузка конфигурации
	cfg, err := config.New()
	if err != nil {
		log.Fatal("Failed to load config", logger.Error(err))
	}

	// Инициализация базы данных
	db, err := sql.Open("sqlite3", cfg.DBPath)
	if err != nil {
		log.Fatal("Failed to open database", logger.Error(err))
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error("Failed to close database", logger.Error(err))
		}
	}()

	// Проверка соединения с БД
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to connect to database", logger.Error(err))
	}

	// Применение миграций
	if err := applyMigrations(db); err != nil {
		log.Fatal("Failed to apply migrations", logger.Error(err))
	}

	// Инициализация репозиториев
	userRepo := repository.NewUserRepository(db, log)

	// Настройка времени жизни токенов
	accessExpiry := 15 * time.Minute
	refreshExpiry := 7 * 24 * time.Hour

	// Инициализация use cases
	authUC := auth.NewAuthUseCase(*userRepo, cfg.JWTSecret, accessExpiry, refreshExpiry, log)
	jwtService := jwt.NewJWTService(cfg.JWTSecret, accessExpiry, refreshExpiry)

	// Инициализация HTTP обработчиков
	authHandler := myHttp.NewAuthHTTPHandler(authUC, jwtService)

	// Настройка роутера
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Маршруты аутентификации
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
	})

	// Защищенные маршруты
	r.Group(func(r chi.Router) {
		r.Use(authHandler.AuthMiddleware)
		r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value("user_id").(string)
			authHandler.JsonResponse(w,
				map[string]string{"message": "Authenticated user: " + userID},
				http.StatusOK)
		})
	})

	// Настройка сервера
	server := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	log.Info("Starting server on :8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Server failed", logger.Error(err))
	}
}

func applyMigrations(db *sql.DB) error {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"sqlite3", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

// package main

// import (
// 	"database/sql"
// 	"log"
// 	"net/http"
// 	"time"

// 	"github.com/go-chi/chi/v5"
// 	"github.com/go-chi/cors"
// 	"github.com/golang-migrate/migrate/v4"
// 	"github.com/golang-migrate/migrate/v4/database/sqlite3"
// 	_ "github.com/golang-migrate/migrate/v4/source/file"
// 	"github.com/kprf42/dolgova/auth_service/internal/config"
// 	myHttp "github.com/kprf42/dolgova/auth_service/internal/delivery/http"
// 	"github.com/kprf42/dolgova/auth_service/internal/repository"
// 	"github.com/kprf42/dolgova/auth_service/internal/usecase/auth"
// 	"github.com/kprf42/dolgova/auth_service/internal/usecase/jwt"
// 	_ "github.com/mattn/go-sqlite3"
// )

// func main() {
// 	cfg, err := config.New()
// 	if err != nil {
// 		log.Fatalf("Failed to load config: %v", err)
// 	}

// 	// Initialize database
// 	db, err := sql.Open("sqlite3", cfg.DBPath)
// 	if err != nil {
// 		log.Fatalf("Failed to open database: %v", err)
// 	}
// 	defer db.Close()

// 	// Check database connection
// 	if err := db.Ping(); err != nil {
// 		log.Fatalf("Failed to connect to database: %v", err)
// 	}

// 	// Apply migrations
// 	if err := applyMigrations(db); err != nil {
// 		log.Fatalf("Failed to apply migrations: %v", err)
// 	}

// 	// Initialize repositories
// 	userRepo := repository.NewUserRepository(db)

// 	// Set token expiration to 15 minutes
// 	accessExpiry := 15 * time.Minute
// 	refreshExpiry := 7 * 24 * time.Hour // Refresh token lasts 7 days

// 	// Initialize use cases with 15-minute access token expiration
// 	authUC := auth.NewAuthUseCase(*userRepo, cfg.JWTSecret, accessExpiry, refreshExpiry)
// 	jwtService := jwt.NewJWTService(cfg.JWTSecret, accessExpiry, refreshExpiry)

// 	// Initialize HTTP handlers
// 	authHandler := myHttp.NewAuthHTTPHandler(authUC, jwtService)

// 	// Setup router
// 	r := chi.NewRouter()
// 	r.Use(cors.Handler(cors.Options{
// 		AllowedOrigins:   []string{"http://localhost:3000"}, // Укажите адрес вашего фронтенда
// 		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
// 		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
// 		ExposedHeaders:   []string{"Link"},
// 		AllowCredentials: true,
// 		MaxAge:           300,
// 	}))
// 	// Authentication routes
// 	r.Route("/auth", func(r chi.Router) {
// 		r.Post("/register", authHandler.Register)
// 		r.Post("/login", authHandler.Login)
// 	})

// 	// Protected routes
// 	r.Group(func(r chi.Router) {
// 		r.Use(authHandler.AuthMiddleware)
// 		r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
// 			userID := r.Context().Value("user_id").(string)
// 			authHandler.JsonResponse(w,
// 				map[string]string{"message": "Authenticated user: " + userID},
// 				http.StatusOK)
// 		})
// 	})

// 	// Start server
// 	server := &http.Server{
// 		Addr:         ":8080",
// 		Handler:      r,
// 		ReadTimeout:  5 * time.Second,
// 		WriteTimeout: 10 * time.Second,
// 		IdleTimeout:  15 * time.Second,
// 	}

// 	log.Println("Starting server on :8080")
// 	if err := server.ListenAndServe(); err != nil {
// 		log.Fatalf("Server failed: %v", err)
// 	}
// }

// func applyMigrations(db *sql.DB) error {
// 	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
// 	if err != nil {
// 		return err
// 	}

// 	m, err := migrate.NewWithDatabaseInstance(
// 		"file://migrations",
// 		"sqlite3", driver)
// 	if err != nil {
// 		return err
// 	}

// 	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
// 		return err
// 	}

// 	log.Println("Migrations applied successfully")
// 	return nil
// }
