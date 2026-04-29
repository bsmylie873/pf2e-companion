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

// gameConn is a per-connection record. The mu serialises writes
// (gorilla/websocket requires a single writer at a time).
type gameConn struct {
	conn *websocket.Conn
	isGM bool
	mu   sync.Mutex
}

// BroadcastFilter returns true if the recipient should receive the event.
type BroadcastFilter func(userID uuid.UUID, isGM bool) bool

// GameEventHub manages WebSocket connections per game, keyed by userID.
type GameEventHub struct {
	mu    sync.RWMutex
	conns map[uuid.UUID]map[uuid.UUID]*gameConn // gameID → userID → record
	log   func(format string, args ...interface{})
}

// NewGameEventHub creates a new GameEventHub.
func NewGameEventHub() *GameEventHub {
	return &GameEventHub{
		conns: make(map[uuid.UUID]map[uuid.UUID]*gameConn),
	}
}

// SetLogger optionally attaches a logger for write-failure diagnostics.
func (h *GameEventHub) SetLogger(fn func(string, ...interface{})) {
	h.log = fn
}

// Register adds a WebSocket connection for a user in a game.
// If an existing connection is present for the same user, it is closed first (tab replacement).
func (h *GameEventHub) Register(gameID, userID uuid.UUID, conn *websocket.Conn, isGM bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.conns[gameID] == nil {
		h.conns[gameID] = make(map[uuid.UUID]*gameConn)
	}
	if existing, ok := h.conns[gameID][userID]; ok && existing.conn != conn {
		_ = existing.conn.Close()
	}
	h.conns[gameID][userID] = &gameConn{conn: conn, isGM: isGM}
}

// Unregister removes a user's WebSocket connection from a game,
// but ONLY if the stored connection matches the provided conn.
// This prevents a stale read-loop's deferred Unregister from
// wiping out a fresh reconnect's registration.
func (h *GameEventHub) Unregister(gameID, userID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if entry, ok := h.conns[gameID][userID]; ok && entry.conn == conn {
		delete(h.conns[gameID], userID)
		if len(h.conns[gameID]) == 0 {
			delete(h.conns, gameID)
		}
	}
}

// writeOrPrune serialises the write via the per-conn mutex.
// On error it logs, unregisters, and closes the dead connection.
func (h *GameEventHub) writeOrPrune(gameID, userID uuid.UUID, entry *gameConn, event GameEvent) {
	entry.mu.Lock()
	err := entry.conn.WriteJSON(event)
	entry.mu.Unlock()
	if err != nil {
		if h.log != nil {
			h.log("game-ws broadcast write failed (game=%s user=%s): %v", gameID, userID, err)
		}
		h.Unregister(gameID, userID, entry.conn)
		_ = entry.conn.Close()
	}
}

// Broadcast sends a GameEvent to ALL subscribers of the given game.
func (h *GameEventHub) Broadcast(gameID uuid.UUID, event GameEvent) {
	h.broadcastFiltered(gameID, uuid.Nil, nil, event)
}

// BroadcastExcept sends a GameEvent to all subscribers except the specified user.
func (h *GameEventHub) BroadcastExcept(gameID, excludeUserID uuid.UUID, event GameEvent) {
	h.broadcastFiltered(gameID, excludeUserID, nil, event)
}

// BroadcastExceptFiltered sends a GameEvent to subscribers that pass the filter,
// excluding the specified user.
func (h *GameEventHub) BroadcastExceptFiltered(
	gameID, excludeUserID uuid.UUID, filter BroadcastFilter, event GameEvent,
) {
	h.broadcastFiltered(gameID, excludeUserID, filter, event)
}

func (h *GameEventHub) broadcastFiltered(
	gameID, excludeUserID uuid.UUID, filter BroadcastFilter, event GameEvent,
) {
	type target struct {
		uid   uuid.UUID
		entry *gameConn
	}
	h.mu.RLock()
	targets := make([]target, 0, len(h.conns[gameID]))
	for uid, entry := range h.conns[gameID] {
		if uid == excludeUserID {
			continue
		}
		if filter != nil && !filter(uid, entry.isGM) {
			continue
		}
		targets = append(targets, target{uid, entry})
	}
	h.mu.RUnlock()
	for _, t := range targets {
		h.writeOrPrune(gameID, t.uid, t.entry, event)
	}
}

// SendToUser sends a GameEvent only to the specified user in the game.
func (h *GameEventHub) SendToUser(gameID, userID uuid.UUID, event GameEvent) {
	h.mu.RLock()
	entry := h.conns[gameID][userID]
	h.mu.RUnlock()
	if entry != nil {
		h.writeOrPrune(gameID, userID, entry, event)
	}
}

// ConnCount returns the number of registered connections for a game (test helper).
func (h *GameEventHub) ConnCount(gameID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.conns[gameID])
}

// HasConn returns true if the given user is registered in the game (test helper).
func (h *GameEventHub) HasConn(gameID, userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.conns[gameID][userID]
	return ok
}
