// internal/delivery/http/auth_handler.go
package http

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kprf42/dolgova/auth_service/internal/entity"
	"github.com/kprf42/dolgova/auth_service/internal/usecase/auth"
	"github.com/kprf42/dolgova/auth_service/internal/usecase/jwt"
)

// AuthHTTPHandler объединяет все HTTP-обработчики аутентификации
type AuthHTTPHandler struct {
	authUC *auth.AuthUseCase
	jwtUC  jwt.JWTUseCase
}

// NewAuthHTTPHandler создает новый экземпляр обработчиков
func NewAuthHTTPHandler(authUC *auth.AuthUseCase, jwtUC jwt.JWTUseCase) *AuthHTTPHandler {
	return &AuthHTTPHandler{
		authUC: authUC,
		jwtUC:  jwtUC,
	}
}

// RegisterRoutes настраивает маршруты для аутентификации
func (h *AuthHTTPHandler) RegisterRoutes(router chi.Router) {
	router.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)
		r.Group(func(r chi.Router) {
			r.Use(h.AuthMiddleware)
		})
	})
}

// RegisterRequest структура запроса регистрации
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterResponse структура ответа регистрации
type RegisterResponse struct {
	UserID string `json:"user_id"`
}

func (h *AuthHTTPHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Decode error: %v", err)
		h.jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Register attempt: %+v", req)

	user, err := h.authUC.Register(r.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		log.Printf("Register error: %v", err)
		h.handleAuthError(w, err)
		return
	}

	h.JsonResponse(w, RegisterResponse{UserID: user.ID}, http.StatusCreated)
}

func (h *AuthHTTPHandler) jsonError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// LoginRequest структура запроса входа
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse структура ответа входа
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// Login обработчик входа пользователя
func (h *AuthHTTPHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tokens, err := h.authUC.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	h.JsonResponse(w, LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.AtExpires,
	}, http.StatusOK)
}

// AuthMiddleware middleware для аутентификации
func (h *AuthHTTPHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Authorization token required", http.StatusUnauthorized)
			return
		}

		claims, err := h.jwtUC.ValidateToken(token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *AuthHTTPHandler) handleAuthError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	var (
		message    string
		statusCode int
	)

	switch {
	case errors.Is(err, entity.ErrUserAlreadyExists):
		message = "User with this email already exists"
		statusCode = http.StatusConflict
	case errors.Is(err, entity.ErrInvalidEmail):
		message = "Invalid email format"
		statusCode = http.StatusBadRequest
	case errors.Is(err, entity.ErrWeakPassword):
		message = "Password must be at least 8 characters"
		statusCode = http.StatusBadRequest
	case errors.Is(err, entity.ErrEmptyUsername):
		message = "Username cannot be empty"
		statusCode = http.StatusBadRequest
	default:
		message = "Internal server error"
		statusCode = http.StatusInternalServerError
		log.Printf("Internal error: %v", err)
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// JsonResponse отправка JSON-ответа (экспортированный метод)
func (h *AuthHTTPHandler) JsonResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// // internal/delivery/http/auth_handler.go
// package http

// import (
// 	"context"
// 	"encoding/json"
// 	"errors"
// 	"log"
// 	"net/http"

// 	"github.com/go-chi/chi/v5"
// 	"github.com/kprf42/dolgova/auth_service/internal/entity"
// 	"github.com/kprf42/dolgova/auth_service/internal/usecase/auth"
// 	"github.com/kprf42/dolgova/auth_service/internal/usecase/jwt"
// )

// // AuthHTTPHandler объединяет все HTTP-обработчики аутентификации
// type AuthHTTPHandler struct {
// 	authUC *auth.AuthUseCase
// 	jwtUC  jwt.JWTUseCase
// }

// // NewAuthHTTPHandler создает новый экземпляр обработчиков
// func NewAuthHTTPHandler(authUC *auth.AuthUseCase, jwtUC jwt.JWTUseCase) *AuthHTTPHandler {
// 	return &AuthHTTPHandler{
// 		authUC: authUC,
// 		jwtUC:  jwtUC,
// 	}
// }

// // RegisterRoutes настраивает маршруты для аутентификации
// func (h *AuthHTTPHandler) RegisterRoutes(router chi.Router) {
// 	router.Route("/auth", func(r chi.Router) {
// 		r.Post("/register", h.Register)
// 		r.Post("/login", h.Login)
// 		r.Group(func(r chi.Router) {
// 			r.Use(h.AuthMiddleware)
// 		})
// 	})
// }

// // RegisterRequest структура запроса регистрации
// type RegisterRequest struct {
// 	Username string `json:"username"`
// 	Email    string `json:"email"`
// 	Password string `json:"password"`
// }

// // RegisterResponse структура ответа регистрации
// type RegisterResponse struct {
// 	UserID string `json:"user_id"`
// }

// func (h *AuthHTTPHandler) Register(w http.ResponseWriter, r *http.Request) {
// 	var req RegisterRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		log.Printf("Decode error: %v", err)
// 		h.jsonError(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	log.Printf("Register attempt: %+v", req)

// 	user, err := h.authUC.Register(r.Context(), req.Username, req.Email, req.Password)
// 	if err != nil {
// 		log.Printf("Register error: %v", err) // Добавлено логирование
// 		h.handleAuthError(w, err)
// 		return
// 	}

// 	h.jsonResponse(w, RegisterResponse{UserID: user.ID}, http.StatusCreated)
// }

// // Добавьте этот новый метод в ваш AuthHTTPHandler
// func (h *AuthHTTPHandler) jsonError(w http.ResponseWriter, message string, statusCode int) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(statusCode)
// 	json.NewEncoder(w).Encode(map[string]string{"error": message})
// }

// // Register обработчик регистрации пользователя
// //func (h *AuthHTTPHandler) Register(w http.ResponseWriter, r *http.Request) {
// // 	var req RegisterRequest
// // 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// // 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// // 		return
// // 	}

// // 	user, err := h.authUC.Register(r.Context(), req.Username, req.Email, req.Password)
// // 	if err != nil {
// // 		h.handleAuthError(w, err)
// // 		return
// // 	}

// // 	h.jsonResponse(w, RegisterResponse{UserID: user.ID}, http.StatusCreated)
// // }

// // LoginRequest структура запроса входа
// type LoginRequest struct {
// 	Email    string `json:"email"`
// 	Password string `json:"password"`
// }

// // LoginResponse структура ответа входа
// type LoginResponse struct {
// 	AccessToken  string `json:"access_token"`
// 	RefreshToken string `json:"refresh_token"`
// 	ExpiresIn    int64  `json:"expires_in"`
// }

// // Login обработчик входа пользователя
// func (h *AuthHTTPHandler) Login(w http.ResponseWriter, r *http.Request) {
// 	var req LoginRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	tokens, err := h.authUC.Login(r.Context(), req.Email, req.Password)
// 	if err != nil {
// 		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
// 		return
// 	}

// 	h.jsonResponse(w, LoginResponse{
// 		AccessToken:  tokens.AccessToken,
// 		RefreshToken: tokens.RefreshToken,
// 		ExpiresIn:    tokens.AtExpires,
// 	}, http.StatusOK)
// }

// // AuthMiddleware middleware для аутентификации
// func (h *AuthHTTPHandler) AuthMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		token := r.Header.Get("Authorization")
// 		if token == "" {
// 			http.Error(w, "Authorization token required", http.StatusUnauthorized)
// 			return
// 		}

// 		claims, err := h.jwtUC.ValidateToken(token)
// 		if err != nil {
// 			http.Error(w, "Invalid token", http.StatusUnauthorized)
// 			return
// 		}

// 		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	})
// }

// // handleAuthError обработка ошибок аутентификации
// //
// //	func (h *AuthHTTPHandler) handleAuthError(w http.ResponseWriter, err error) {
// //		switch err {
// //		case entity.ErrUserAlreadyExists:
// //			http.Error(w, "User already exists", http.StatusConflict)
// //		case entity.ErrInvalidEmail:
// //			http.Error(w, "Invalid email format", http.StatusBadRequest)
// //		case entity.ErrWeakPassword:
// //			http.Error(w, "Password must be at least 8 characters", http.StatusBadRequest)
// //		default:
// //			http.Error(w, "Internal server error", http.StatusInternalServerError)
// //		}
// //	}
// func (h *AuthHTTPHandler) handleAuthError(w http.ResponseWriter, err error) {
// 	w.Header().Set("Content-Type", "application/json")

// 	var (
// 		message    string
// 		statusCode int
// 	)

// 	switch {
// 	case errors.Is(err, entity.ErrUserAlreadyExists):
// 		message = "User with this email already exists"
// 		statusCode = http.StatusConflict
// 	case errors.Is(err, entity.ErrInvalidEmail):
// 		message = "Invalid email format"
// 		statusCode = http.StatusBadRequest
// 	case errors.Is(err, entity.ErrWeakPassword):
// 		message = "Password must be at least 8 characters"
// 		statusCode = http.StatusBadRequest
// 	case errors.Is(err, entity.ErrEmptyUsername):
// 		message = "Username cannot be empty"
// 		statusCode = http.StatusBadRequest
// 	default:
// 		message = "Internal server error"
// 		statusCode = http.StatusInternalServerError
// 		log.Printf("Internal error: %v", err) // Логируем внутренние ошибки
// 	}

// 	w.WriteHeader(statusCode)
// 	json.NewEncoder(w).Encode(map[string]string{"error": message})
// }

// // jsonResponse отправка JSON-ответа
// func (h *AuthHTTPHandler) jsonResponse(w http.ResponseWriter, data interface{}, statusCode int) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(statusCode)
// 	json.NewEncoder(w).Encode(data)
// }

// // Пример инициализации сервера
// /*
// func main() {
// 	// Инициализация репозиториев и use cases
// 	db := initDB() // Ваша функция инициализации БД
// 	userRepo := repository.NewUserRepository(db)

// 	authUC := usecase.NewAuthUseCase(userRepo, "your-secret-key", time.Hour, time.Hour*24*7)
// 	jwtUC := usecase.NewJWTUseCase("your-secret-key", time.Hour, time.Hour*24*7)

// 	// Создание HTTP обработчиков
// 	authHandler := NewAuthHTTPHandler(authUC, jwtUC)

// 	// Настройка маршрутов
// 	r := chi.NewRouter()
// 	authHandler.RegisterRoutes(r)

// 	// Запуск сервера
// 	http.ListenAndServe(":8080", r)
// }
// */
