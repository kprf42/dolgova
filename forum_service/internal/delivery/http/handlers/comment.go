package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kprf42/dolgova/forum_service/internal/entity"
	comment "github.com/kprf42/dolgova/forum_service/internal/usecase"
)

type CommentHandlers struct {
	uc *comment.CommentUseCase
}

func NewCommentHandlers(uc *comment.CommentUseCase) *CommentHandlers {
	return &CommentHandlers{uc: uc}
}

func (h *CommentHandlers) CreateComment(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("\n=== CreateComment Handler ===\n")
	fmt.Printf("Request URL: %s\n", r.URL.String())

	// Получаем postID из URL
	postID := chi.URLParam(r, "postId")
	fmt.Printf("Post ID from URL: '%s'\n", postID)

	// Проверяем UUID
	if _, err := uuid.Parse(postID); err != nil {
		fmt.Printf("ERROR: Invalid UUID format. Input: '%s', Error: %v\n", postID, err)
		fmt.Printf("Expected format example: 550e8400-e29b-41d4-a716-446655440000\n")
		http.Error(w, fmt.Sprintf("invalid post id format: must be a valid UUID"), http.StatusBadRequest)
		return
	}

	// Декодируем тело запроса
	var req entity.CommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Printf("ERROR: Failed to decode request body: %v\n", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.PostID = postID
	fmt.Printf("Request body decoded: %+v\n", req)

	// Получаем user_id из контекста
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		fmt.Printf("ERROR: Failed to get user_id from context\n")
		http.Error(w, "unauthorized: missing user_id", http.StatusUnauthorized)
		return
	}
	fmt.Printf("User ID from context: %s\n", userID)

	// Создаем комментарий
	comment, err := h.uc.Create(r.Context(), &req, userID)
	if err != nil {
		fmt.Printf("ERROR: Failed to create comment: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("Successfully created comment: %+v\n", comment)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(comment); err != nil {
		fmt.Printf("ERROR: Failed to encode response: %v\n", err)
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}

	fmt.Printf("=== End CreateComment Handler ===\n\n")
}

func (h *CommentHandlers) GetComments(w http.ResponseWriter, r *http.Request) {
	// Добавьте отладочный вывод
	fmt.Println("\n=== GetComments Handler ===")
	fmt.Printf("Request URL: %s\n", r.URL.String())

	// Получаем postID из URL
	postID := chi.URLParam(r, "postId")
	fmt.Printf("Extracted postID: '%s'\n", postID)

	// Проверяем UUID
	if _, err := uuid.Parse(postID); err != nil {
		fmt.Printf("Invalid UUID: %v\n", err)
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}

	// Обработка параметров запроса
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	fmt.Printf("Query params: limit=%d, offset=%d\n", limit, offset)

	// Получаем комментарии
	comments, total, err := h.uc.GetByPostID(r.Context(), postID, limit, offset)
	if err != nil {
		fmt.Printf("Error getting comments: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("Found %d comments (total: %d)\n", len(comments), total)

	// Формируем ответ
	response := struct {
		Comments []*entity.Comment `json:"comments"`
		Total    int               `json:"total"`
	}{
		Comments: comments,
		Total:    total,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("Error encoding response: %v\n", err)
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}

	fmt.Println("=== End GetComments Handler ===")
}

// func (h *CommentHandlers) GetComments(w http.ResponseWriter, r *http.Request) {
// 	postID := chi.URLParam(r, "id")
// 	if _, err := uuid.Parse(postID); err != nil {
// 		http.Error(w, "invalid post id", http.StatusBadRequest)
// 		return
// 	}

// 	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
// 	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

// 	if limit <= 0 {
// 		limit = 10
// 	}
// 	if offset < 0 {
// 		offset = 0
// 	}

// 	comments, total, err := h.uc.GetByPostID(r.Context(), postID, limit, offset)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	response := struct {
// 		Comments []*entity.Comment `json:"comments"`
// 		Total    int               `json:"total"`
// 	}{
// 		Comments: comments,
// 		Total:    total,
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(response)
// }
