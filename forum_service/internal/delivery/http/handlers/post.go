package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kprf42/dolgova/forum_service/internal/entity"
	post "github.com/kprf42/dolgova/forum_service/internal/usecase"
)

// JWTClaims кастомная структура claims с реализацией всех необходимых методов
type JWTClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type PostHandlers struct {
	uc *post.PostUseCase
}

func NewPostHandlers(uc *post.PostUseCase) *PostHandlers {
	return &PostHandlers{uc: uc}
}

func (h *PostHandlers) CreatePost(w http.ResponseWriter, r *http.Request) {
	var req entity.PostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Printf("Error decoding request: %v\n", err)
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	fmt.Printf("Received request: %+v\n", req)

	// Проверяем, что category_id является числом от 1 до 3
	categoryID := req.CategoryID
	if categoryID != "1" && categoryID != "2" && categoryID != "3" {
		fmt.Printf("Invalid category_id: %s\n", categoryID)
		http.Error(w, "invalid category_id: must be 1, 2 or 3", http.StatusBadRequest)
		return
	}

	// Получаем claims из контекста
	claimsValue := r.Context().Value("claims")
	fmt.Printf("Claims from context: %v (type: %T)\n", claimsValue, claimsValue)

	claims, ok := claimsValue.(map[string]interface{})
	if !ok {
		fmt.Printf("Failed to get claims from context\n")
		http.Error(w, "unauthorized: invalid claims", http.StatusUnauthorized)
		return
	}

	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		fmt.Printf("Failed to get user_id from claims. ok: %v, userID: %s\n", ok, userID)
		http.Error(w, "unauthorized: missing user_id", http.StatusUnauthorized)
		return
	}

	fmt.Printf("Creating post for user: %s\n", userID)

	response, err := h.uc.Create(r.Context(), &req, userID)
	if err != nil {
		fmt.Printf("Error creating post: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *PostHandlers) GetPost(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("\n=== GetPost Handler ===\n")

	// Получаем параметры несколькими способами для отладки
	rctx := chi.RouteContext(r.Context())
	var postID string

	if rctx != nil {
		// Способ 1: Через URLParam
		postID = chi.URLParam(r, "postId")
		fmt.Printf("Method 1 - URLParam: '%s'\n", postID)

		// Способ 2: Напрямую из контекста
		if len(rctx.URLParams.Keys) > 0 && len(rctx.URLParams.Values) > 0 {
			for i, key := range rctx.URLParams.Keys {
				fmt.Printf("URL Param %s: %s\n", key, rctx.URLParams.Values[i])
				if key == "postId" {
					postID = rctx.URLParams.Values[i]
				}
			}
		}

		// Способ 3: Через pattern matching
		fmt.Printf("Route Pattern: %s\n", rctx.RoutePattern())
	} else {
		fmt.Printf("ERROR: Chi context is nil\n")
	}

	// Способ 4: Парсим URL напрямую
	urlPath := r.URL.Path
	fmt.Printf("URL Path: %s\n", urlPath)
	pathParts := strings.Split(urlPath, "/")
	if len(pathParts) > 4 { // /api/v1/posts/{postId}
		fmt.Printf("PostID from URL split: %s\n", pathParts[4])
		if postID == "" {
			postID = pathParts[4]
		}
	}

	fmt.Printf("Final PostID: '%s'\n", postID)

	// Проверяем, не пустой ли ID
	if postID == "" {
		fmt.Printf("ERROR: Post ID is empty\n")
		http.Error(w, "post id is required", http.StatusBadRequest)
		return
	}

	// Пытаемся распарсить UUID
	parsedUUID, err := uuid.Parse(postID)
	if err != nil {
		fmt.Printf("ERROR: Invalid UUID format. Input: '%s', Error: %v\n", postID, err)
		fmt.Printf("Expected format example: 550e8400-e29b-41d4-a716-446655440000\n")
		http.Error(w, fmt.Sprintf("invalid post id format: must be a valid UUID (example: 550e8400-e29b-41d4-a716-446655440000)"), http.StatusBadRequest)
		return
	}

	fmt.Printf("Successfully parsed UUID: %s\n", parsedUUID.String())

	post, err := h.uc.GetByID(r.Context(), postID)
	if err != nil {
		fmt.Printf("ERROR: Failed to get post from database: %v\n", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	fmt.Printf("Successfully retrieved post from database: %+v\n", post)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(post); err != nil {
		fmt.Printf("ERROR: Failed to encode response: %v\n", err)
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Successfully sent response for post ID: %s\n", post.ID)
	fmt.Printf("=== End GetPost Handler ===\n\n")
}

func (h *PostHandlers) GetPosts(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	categoryID := r.URL.Query().Get("category_id")

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	posts, total, err := h.uc.GetAll(r.Context(), limit, offset, categoryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Posts []*entity.PostResponse `json:"posts"`
		Total int                    `json:"total"`
	}{
		Posts: posts,
		Total: total,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *PostHandlers) UpdatePost(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("\n=== UpdatePost Handler ===\n")

	// Получаем параметры несколькими способами для отладки
	rctx := chi.RouteContext(r.Context())
	var postID string

	if rctx != nil {
		// Способ 1: Через URLParam
		postID = chi.URLParam(r, "postId")
		fmt.Printf("Method 1 - URLParam: '%s'\n", postID)

		// Способ 2: Напрямую из контекста
		if len(rctx.URLParams.Keys) > 0 && len(rctx.URLParams.Values) > 0 {
			for i, key := range rctx.URLParams.Keys {
				fmt.Printf("URL Param %s: %s\n", key, rctx.URLParams.Values[i])
				if key == "postId" {
					postID = rctx.URLParams.Values[i]
				}
			}
		}

		// Способ 3: Через pattern matching
		fmt.Printf("Route Pattern: %s\n", rctx.RoutePattern())
	} else {
		fmt.Printf("ERROR: Chi context is nil\n")
	}

	// Способ 4: Парсим URL напрямую
	urlPath := r.URL.Path
	fmt.Printf("URL Path: %s\n", urlPath)
	pathParts := strings.Split(urlPath, "/")
	if len(pathParts) > 4 { // /api/v1/posts/{postId}
		fmt.Printf("PostID from URL split: %s\n", pathParts[4])
		if postID == "" {
			postID = pathParts[4]
		}
	}

	fmt.Printf("Final PostID: '%s'\n", postID)

	// Проверяем, не пустой ли ID
	if postID == "" {
		fmt.Printf("ERROR: Post ID is empty\n")
		http.Error(w, "post id is required", http.StatusBadRequest)
		return
	}

	// Пытаемся распарсить UUID
	_, err := uuid.Parse(postID)
	if err != nil {
		fmt.Printf("ERROR: Invalid UUID format. Input: '%s', Error: %v\n", postID, err)
		fmt.Printf("Expected format example: 550e8400-e29b-41d4-a716-446655440000\n")
		http.Error(w, fmt.Sprintf("invalid post id format: must be a valid UUID (example: 550e8400-e29b-41d4-a716-446655440000)"), http.StatusBadRequest)
		return
	}

	// Декодируем тело запроса
	var req entity.PostUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Printf("ERROR: Failed to decode request body: %v\n", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	fmt.Printf("Request body decoded: %+v\n", req)

	// Получаем user_id из контекста
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		fmt.Printf("ERROR: Failed to get user_id from context\n")
		http.Error(w, "unauthorized: missing user_id", http.StatusUnauthorized)
		return
	}
	fmt.Printf("User ID from context: %s\n", userID)

	// Обновляем пост
	response, err := h.uc.Update(r.Context(), postID, &req, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "unauthorized" {
			status = http.StatusUnauthorized
		}
		fmt.Printf("ERROR: Failed to update post: %v\n", err)
		http.Error(w, err.Error(), status)
		return
	}

	fmt.Printf("Successfully updated post: %+v\n", response)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("ERROR: Failed to encode response: %v\n", err)
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}

	fmt.Printf("=== End UpdatePost Handler ===\n\n")
}

func (h *PostHandlers) DeletePost(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("\n=== DeletePost Handler ===\n")

	// Получаем параметры несколькими способами для отладки
	rctx := chi.RouteContext(r.Context())
	var postID string

	if rctx != nil {
		// Способ 1: Через URLParam
		postID = chi.URLParam(r, "postId")
		fmt.Printf("Method 1 - URLParam: '%s'\n", postID)

		// Способ 2: Напрямую из контекста
		if len(rctx.URLParams.Keys) > 0 && len(rctx.URLParams.Values) > 0 {
			for i, key := range rctx.URLParams.Keys {
				fmt.Printf("URL Param %s: %s\n", key, rctx.URLParams.Values[i])
				if key == "postId" {
					postID = rctx.URLParams.Values[i]
				}
			}
		}

		// Способ 3: Через pattern matching
		fmt.Printf("Route Pattern: %s\n", rctx.RoutePattern())
	} else {
		fmt.Printf("ERROR: Chi context is nil\n")
	}

	// Способ 4: Парсим URL напрямую
	urlPath := r.URL.Path
	fmt.Printf("URL Path: %s\n", urlPath)
	pathParts := strings.Split(urlPath, "/")
	if len(pathParts) > 4 { // /api/v1/posts/{postId}
		fmt.Printf("PostID from URL split: %s\n", pathParts[4])
		if postID == "" {
			postID = pathParts[4]
		}
	}

	fmt.Printf("Final PostID: '%s'\n", postID)

	// Проверяем, не пустой ли ID
	if postID == "" {
		fmt.Printf("ERROR: Post ID is empty\n")
		http.Error(w, "post id is required", http.StatusBadRequest)
		return
	}

	// Пытаемся распарсить UUID
	_, err := uuid.Parse(postID)
	if err != nil {
		fmt.Printf("ERROR: Invalid UUID format. Input: '%s', Error: %v\n", postID, err)
		fmt.Printf("Expected format example: 550e8400-e29b-41d4-a716-446655440000\n")
		http.Error(w, fmt.Sprintf("invalid post id format: must be a valid UUID (example: 550e8400-e29b-41d4-a716-446655440000)"), http.StatusBadRequest)
		return
	}

	// Получаем user_id из контекста
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		fmt.Printf("ERROR: Failed to get user_id from context\n")
		http.Error(w, "unauthorized: missing user_id", http.StatusUnauthorized)
		return
	}
	fmt.Printf("User ID from context: %s\n", userID)

	// Удаляем пост
	if err := h.uc.Delete(r.Context(), postID, userID); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "unauthorized" {
			status = http.StatusUnauthorized
		}
		fmt.Printf("ERROR: Failed to delete post: %v\n", err)
		http.Error(w, err.Error(), status)
		return
	}

	fmt.Printf("Successfully deleted post\n")
	fmt.Printf("=== End DeletePost Handler ===\n\n")

	w.WriteHeader(http.StatusNoContent)
}
