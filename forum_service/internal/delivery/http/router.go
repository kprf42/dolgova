package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kprf42/dolgova/forum_service/internal/delivery/http/handlers"
)

// JWTClaims кастомная структура claims с реализацией всех необходимых методов
type JWTClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type AuthMiddleware struct {
	JWTSecret string
}

func (m *AuthMiddleware) JWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("\n=== JWT Middleware ===\n")
		fmt.Printf("Request URL: %s\n", r.URL.String())
		fmt.Printf("Request Method: %s\n", r.Method)
		fmt.Printf("JWT Secret: %s\n", m.JWTSecret)

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

		parts := strings.Split(tokenString, ".")
		if len(parts) != 3 {
			fmt.Printf("ERROR: Invalid token format - expected 3 parts, got %d\n", len(parts))
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				fmt.Printf("ERROR: Unexpected signing method: %v\n", token.Header["alg"])
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
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

		if claims.ExpiresAt != nil {
			if claims.ExpiresAt.Before(time.Now()) {
				fmt.Printf("ERROR: Token has expired\n")
				http.Error(w, "Token has expired", http.StatusUnauthorized)
				return
			}
		}

		fmt.Printf("Token claims: %+v\n", claims)
		fmt.Printf("User ID from token: %s\n", claims.UserID)

		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		fmt.Printf("Added user_id to context: %s\n", claims.UserID)
		fmt.Printf("=== End JWT Middleware ===\n\n")

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func NewRouter(
	postHandlers *handlers.PostHandlers,
	commentHandlers *handlers.CommentHandlers,
	chatHandlers *handlers.ChatHandlers,
	jwtSecret string,
) *chi.Mux {
	r := chi.NewRouter()

	// Basic middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(enableCORS)

	// Debug middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("\n=== URL Parameters Debug ===\n")
			fmt.Printf("Full URL: %s\n", r.URL.String())
			fmt.Printf("Path: %s\n", r.URL.Path)

			rctx := chi.RouteContext(r.Context())
			if rctx != nil {
				fmt.Printf("Chi Route Pattern: %s\n", rctx.RoutePattern())
				fmt.Printf("Chi URL Params: %+v\n", rctx.URLParams)
			} else {
				fmt.Printf("Chi context is nil\n")
			}
			fmt.Printf("=== End URL Parameters Debug ===\n\n")

			next.ServeHTTP(w, r)
		})
	})

	authMiddleware := &AuthMiddleware{JWTSecret: jwtSecret}

	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Group(func(r chi.Router) {
			r.Get("/posts", postHandlers.GetPosts)
			r.Get("/posts/{postId}", postHandlers.GetPost)
			r.Get("/posts/{postId}/comments", commentHandlers.GetComments)
			r.Get("/chat/messages", chatHandlers.GetMessages)
		})

		// Authenticated routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.JWT)

			r.Post("/posts", postHandlers.CreatePost)
			r.Put("/posts/{postId}", postHandlers.UpdatePost)
			r.Delete("/posts/{postId}", postHandlers.DeletePost)
			r.Post("/posts/{postId}/comments", commentHandlers.CreateComment)
			r.Get("/chat/ws", chatHandlers.Connect)
		})
	})

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return r
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Устанавливаем базовые CORS заголовки
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.Header().Set("Access-Control-Expose-Headers", "Authorization")

		// Обработка preflight запросов
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
