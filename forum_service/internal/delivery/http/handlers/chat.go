package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/kprf42/dolgova/forum_service/internal/delivery/websocket"
	"github.com/kprf42/dolgova/forum_service/internal/entity"
	chat "github.com/kprf42/dolgova/forum_service/internal/usecase"
)

type ChatHandlers struct {
	hub    *websocket.Hub
	chatUC *chat.ChatUseCase
}

func NewChatHandlers(hub *websocket.Hub, chatUC *chat.ChatUseCase) *ChatHandlers {
	return &ChatHandlers{
		hub:    hub,
		chatUC: chatUC,
	}
}

func (h *ChatHandlers) Connect(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*entity.Claims)
	websocket.ServeWs(h.hub, w, r, claims.UserID)
}

func (h *ChatHandlers) GetMessages(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	messages, err := h.chatUC.GetMessages(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
