package websocket

import (
	"context"
	"log"

	"github.com/kprf42/dolgova/forum_service/internal/entity"
)

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan *entity.ChatMessage
	register   chan *Client
	unregister chan *Client
	chatUC     ChatUseCase
}

type ChatUseCase interface {
	SaveMessage(ctx context.Context, msg *entity.ChatMessage) error
	GetMessages(ctx context.Context, limit, offset int) ([]*entity.ChatMessage, error)
}

func NewHub(chatUC ChatUseCase) *Hub {
	return &Hub{
		broadcast:  make(chan *entity.ChatMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		chatUC:     chatUC,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

			// Отправляем историю сообщений новому клиенту
			messages, err := h.chatUC.GetMessages(context.Background(), 100, 0)
			if err == nil {
				for _, msg := range messages {
					client.send <- msg
				}
			}

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}

		case message := <-h.broadcast:
			// Сохраняем сообщение в БД
			if err := h.chatUC.SaveMessage(context.Background(), message); err != nil {
				log.Printf("Error saving message: %v", err)
				continue
			}

			// Рассылаем сообщение всем клиентам
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
