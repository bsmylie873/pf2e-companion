package handlers

import (
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// MapEvent represents a real-time event broadcast to game subscribers.
type MapEvent struct {
	Type   string      `json:"type"`
	MapID  uuid.UUID   `json:"map_id"`
	GameID uuid.UUID   `json:"game_id"`
	Data   interface{} `json:"data"`
}

// MapEventHub manages WebSocket connections per game.
type MapEventHub struct {
	mu    sync.RWMutex
	conns map[uuid.UUID]map[*websocket.Conn]struct{} // gameID → set of conns
}

// NewMapEventHub creates a new MapEventHub.
func NewMapEventHub() *MapEventHub {
	return &MapEventHub{
		conns: make(map[uuid.UUID]map[*websocket.Conn]struct{}),
	}
}

// Register adds a WebSocket connection as a subscriber to a game's map events.
func (h *MapEventHub) Register(gameID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.conns[gameID] == nil {
		h.conns[gameID] = make(map[*websocket.Conn]struct{})
	}
	h.conns[gameID][conn] = struct{}{}
}

// Unregister removes a WebSocket connection from game subscribers.
func (h *MapEventHub) Unregister(gameID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.conns[gameID], conn)
	if len(h.conns[gameID]) == 0 {
		delete(h.conns, gameID)
	}
}

// Broadcast sends a MapEvent to all subscribers of the given game.
func (h *MapEventHub) Broadcast(gameID uuid.UUID, event MapEvent) {
	h.mu.RLock()
	conns := make([]*websocket.Conn, 0, len(h.conns[gameID]))
	for conn := range h.conns[gameID] {
		conns = append(conns, conn)
	}
	h.mu.RUnlock()
	for _, conn := range conns {
		_ = conn.WriteJSON(event)
	}
}
