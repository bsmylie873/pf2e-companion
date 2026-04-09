package handlers

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGameEventHub(t *testing.T) {
	hub := NewGameEventHub()
	require.NotNil(t, hub)
	assert.NotNil(t, hub.conns)
}

func TestGameEventHub_Register_And_Unregister(t *testing.T) {
	hub := NewGameEventHub()
	gameID := uuid.New()
	userID := uuid.New()

	// Register nil conn — exercises the map initialization path
	hub.Register(gameID, userID, nil)

	// Unregister — should not panic
	hub.Unregister(gameID, userID)
}

func TestGameEventHub_Unregister_NonExistent(t *testing.T) {
	hub := NewGameEventHub()
	gameID := uuid.New()
	userID := uuid.New()

	// Unregistering a user that was never registered should not panic
	assert.NotPanics(t, func() {
		hub.Unregister(gameID, userID)
	})
}

func TestGameEventHub_Register_MultipleUsers_SameGame(t *testing.T) {
	hub := NewGameEventHub()
	gameID := uuid.New()
	user1 := uuid.New()
	user2 := uuid.New()

	// Register two users in the same game
	hub.Register(gameID, user1, nil)
	hub.Register(gameID, user2, nil)

	// Remove first user - game should still be tracked
	hub.Unregister(gameID, user1)

	// Remove second user - game entry should be cleaned up
	hub.Unregister(gameID, user2)
}

func TestGameEventHub_Broadcast_EmptyGame(t *testing.T) {
	hub := NewGameEventHub()
	gameID := uuid.New()

	// Broadcast with no connections — should not panic
	assert.NotPanics(t, func() {
		hub.Broadcast(gameID, GameEvent{Type: "test", GameID: gameID})
	})
}

func TestGameEventHub_BroadcastExcept_EmptyGame(t *testing.T) {
	hub := NewGameEventHub()
	gameID := uuid.New()
	userID := uuid.New()

	// BroadcastExcept with no connections — should not panic
	assert.NotPanics(t, func() {
		hub.BroadcastExcept(gameID, userID, GameEvent{Type: "test", GameID: gameID})
	})
}

func TestGameEventHub_SendToUser_NoConnection(t *testing.T) {
	hub := NewGameEventHub()
	gameID := uuid.New()
	userID := uuid.New()

	// SendToUser when user has no connection — should not panic
	assert.NotPanics(t, func() {
		hub.SendToUser(gameID, userID, GameEvent{Type: "test"})
	})
}

func TestGameEventHub_Register_OverwritesExistingConn(t *testing.T) {
	hub := NewGameEventHub()
	gameID := uuid.New()
	userID := uuid.New()

	// Register twice for same user (e.g., reconnect) — should not panic
	hub.Register(gameID, userID, nil)
	hub.Register(gameID, userID, nil) // overwrite with new conn
	hub.Unregister(gameID, userID)
}

func TestGameEventHub_MultipleGames(t *testing.T) {
	hub := NewGameEventHub()
	game1 := uuid.New()
	game2 := uuid.New()
	user1 := uuid.New()
	user2 := uuid.New()

	hub.Register(game1, user1, nil)
	hub.Register(game2, user2, nil)

	// Unregister from game1 — game2 should remain
	hub.Unregister(game1, user1)

	// BroadcastExcept game2 should not panic (user2 has nil conn but Broadcast
	// only iterates over conns for that gameID, and since we only call Broadcast
	// on empty hub entries here, it is safe)
	assert.NotPanics(t, func() {
		hub.Broadcast(game1, GameEvent{Type: "noop", GameID: game1})
	})

	hub.Unregister(game2, user2)
}
