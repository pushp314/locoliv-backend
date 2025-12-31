package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/locolive/backend/internal/domain"
	"github.com/locolive/backend/internal/middleware"
	"github.com/locolive/backend/pkg/response"
	"go.uber.org/zap"
)

type ChatHandler struct {
	chatService *domain.ChatService
	wsManager   *WebSocketManager
	logger      *zap.Logger
}

func NewChatHandler(chatService *domain.ChatService, wsManager *WebSocketManager, logger *zap.Logger) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		wsManager:   wsManager,
		logger:      logger,
	}
}

// HandleWebSocket upgrades HTTP connection to WebSocket
func (h *ChatHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		// WebSocket auth usually happens via query param ticket or similar if headers not supported by client lib
		// For MVP, we'll assume AuthMiddleware worked (cookie/header)
		response.Unauthorized(w, "not authenticated")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("WebSocket upgrade failed", zap.Error(err))
		return
	}

	client := &Client{
		ID:     uuid.New(),
		Conn:   conn,
		Send:   make(chan []byte, 256),
		UserID: userID,
	}

	h.wsManager.register <- client

	go client.WritePump()
	go client.ReadPump(h.wsManager)
}

// CreateChat starts a new chat with a user
func (h *ChatHandler) CreateChat(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "not authenticated")
		return
	}

	var req struct {
		TargetUserID string `json:"target_user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request")
		return
	}

	targetID, err := uuid.Parse(req.TargetUserID)
	if err != nil {
		response.BadRequest(w, "invalid target user id")
		return
	}

	chat, err := h.chatService.CreateChat(r.Context(), userID, targetID)
	if err != nil {
		h.logger.Error("failed to create chat", zap.Error(err))
		response.InternalError(w, "failed to create chat")
		return
	}

	response.OK(w, chat)
}

// GetChats returns list of user's chats
func (h *ChatHandler) GetChats(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "not authenticated")
		return
	}

	chats, err := h.chatService.GetUserChats(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get chats", zap.Error(err))
		response.InternalError(w, "failed to get chats")
		return
	}

	response.OK(w, chats)
}

// GetMessages returns messages for a chat
func (h *ChatHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	chatIDStr := chi.URLParam(r, "chatId")
	chatID, err := uuid.Parse(chatIDStr)
	if err != nil {
		response.BadRequest(w, "invalid chat id")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	messages, err := h.chatService.GetMessages(r.Context(), chatID, limit, offset)
	if err != nil {
		h.logger.Error("failed to get messages", zap.Error(err))
		response.InternalError(w, "failed to get messages")
		return
	}

	response.OK(w, messages)
}

// SendMessage sends a message to a chat (HTTP fallback + WebSocket broadcast)
func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "not authenticated")
		return
	}

	chatIDStr := chi.URLParam(r, "chatId")
	chatID, err := uuid.Parse(chatIDStr)
	if err != nil {
		response.BadRequest(w, "invalid chat id")
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request")
		return
	}

	msg, err := h.chatService.SendMessage(r.Context(), chatID, userID, req.Content)
	if err != nil {
		h.logger.Error("failed to send message", zap.Error(err))
		response.InternalError(w, "failed to send message")
		return
	}

	// Broadcast via WebSocket
	// 1. Get chat participants to know who to notify
	chat, err := h.chatService.GetChat(r.Context(), chatID)
	if err == nil {
		event := WSEvent{
			Type:    "new_message",
			Payload: msg,
		}
		for _, u := range chat.Users {
			h.wsManager.SendToUser(u.ID, event)
		}
	}

	response.OK(w, msg)
}
