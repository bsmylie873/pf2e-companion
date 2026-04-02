package handlers

import (
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// GameEvent represents a real-time event broadcast to game subscribers.
type GameEvent struct {
	Type   string      `json:"type"`
	GameID uuid.UUID   `json:"game_id"`
	Data   interface{} `json:"data"`
}

// GameEventHub manages WebSocket connections per game, keyed by userID.
type GameEventHub struct {
	mu    sync.RWMutex
	conns map[uuid.UUID]map[uuid.UUID]*websocket.Conn // gameID → userID → conn
}

// NewGameEventHub creates a new GameEventHub.
func NewGameEventHub() *GameEventHub {
	return &GameEventHub{
		conns: make(map[uuid.UUID]map[uuid.UUID]*websocket.Conn),
	}
}

// Register adds a WebSocket connection for a user in a game.
func (h *GameEventHub) Register(gameID, userID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.conns[gameID] == nil {
		h.conns[gameID] = make(map[uuid.UUID]*websocket.Conn)
	}
	h.conns[gameID][userID] = conn
}

// Unregister removes a user's WebSocket connection from a game.
func (h *GameEventHub) Unregister(gameID, userID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.conns[gameID], userID)
	if len(h.conns[gameID]) == 0 {
		delete(h.conns, gameID)
	}
}

// Broadcast sends a GameEvent to ALL subscribers of the given game.
func (h *GameEventHub) Broadcast(gameID uuid.UUID, event GameEvent) {
	h.mu.RLock()
	conns := make([]*websocket.Conn, 0, len(h.conns[gameID]))
	for _, conn := range h.conns[gameID] {
		conns = append(conns, conn)
	}
	h.mu.RUnlock()
	for _, conn := range conns {
		_ = conn.WriteJSON(event)
	}
}

// BroadcastExcept sends a GameEvent to all subscribers except the specified user.
func (h *GameEventHub) BroadcastExcept(gameID, excludeUserID uuid.UUID, event GameEvent) {
	h.mu.RLock()
	conns := make([]*websocket.Conn, 0)
	for uid, conn := range h.conns[gameID] {
		if uid != excludeUserID {
			conns = append(conns, conn)
		}
	}
	h.mu.RUnlock()
	for _, conn := range conns {
		_ = conn.WriteJSON(event)
	}
}

// SendToUser sends a GameEvent only to the specified user in the game.
func (h *GameEventHub) SendToUser(gameID, userID uuid.UUID, event GameEvent) {
	h.mu.RLock()
	conn := h.conns[gameID][userID]
	h.mu.RUnlock()
	if conn != nil {
		_ = conn.WriteJSON(event)
	}
}
