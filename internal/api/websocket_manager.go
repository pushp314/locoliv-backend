package api

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now (adjust for production)
	},
}

type Client struct {
	ID     uuid.UUID
	Conn   *websocket.Conn
	Send   chan []byte
	UserID uuid.UUID
}

type WebSocketManager struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	// Map userID to list of active clients (for multi-device support)
	userClients map[uuid.UUID]map[*Client]bool
	mu          sync.RWMutex
	logger      *zap.Logger
}

func NewWebSocketManager(logger *zap.Logger) *WebSocketManager {
	return &WebSocketManager{
		clients:     make(map[*Client]bool),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan []byte),
		userClients: make(map[uuid.UUID]map[*Client]bool),
		logger:      logger,
	}
}

func (m *WebSocketManager) Run() {
	for {
		select {
		case client := <-m.register:
			m.mu.Lock()
			m.clients[client] = true
			if _, ok := m.userClients[client.UserID]; !ok {
				m.userClients[client.UserID] = make(map[*Client]bool)
			}
			m.userClients[client.UserID][client] = true
			m.mu.Unlock()
			m.logger.Debug("Client registered", zap.String("userID", client.UserID.String()))

		case client := <-m.unregister:
			m.mu.Lock()
			if _, ok := m.clients[client]; ok {
				delete(m.clients, client)
				if userMap, ok := m.userClients[client.UserID]; ok {
					delete(userMap, client)
					if len(userMap) == 0 {
						delete(m.userClients, client.UserID)
					}
				}
				close(client.Send)
				m.logger.Debug("Client unregistered", zap.String("userID", client.UserID.String()))
			}
			m.mu.Unlock()

		case message := <-m.broadcast:
			// Broadcast to all (if needed, though we usually target specific users)
			m.mu.RLock()
			for client := range m.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(m.clients, client)
				}
			}
			m.mu.RUnlock()
		}
	}
}

// SendToUser sends a message to a specific user's connected clients
func (m *WebSocketManager) SendToUser(userID uuid.UUID, message interface{}) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clients, ok := m.userClients[userID]
	if !ok {
		return
	}

	jsonMsg, err := json.Marshal(message)
	if err != nil {
		m.logger.Error("Failed to marshal message", zap.Error(err))
		return
	}

	for client := range clients {
		select {
		case client.Send <- jsonMsg:
		default:
			// If buffer is full, we assume client is dead/slow and unregister via loop check
			// Ideally we don't block here
		}
	}
}

// WebSocket Event types
type WSEvent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func (c *Client) ReadPump(manager *WebSocketManager) {
	defer func() {
		manager.unregister <- c
		c.Conn.Close()
	}()

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// log error
			}
			break
		}
		// For now, we only push data server->client.
		// If we want client->server via WS, handle messages here.
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()

	for message := range c.Send {
		w, err := c.Conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		w.Write(message)

		// Add queued chat messages to the current websocket message.
		n := len(c.Send)
		for i := 0; i < n; i++ {
			w.Write(<-c.Send)
		}

		if err := w.Close(); err != nil {
			return
		}
	}
	c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
}
