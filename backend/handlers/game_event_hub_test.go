package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeTestConn creates a matched server/client WebSocket pair for testing.
//
//   - serverConn is what the hub should Register — the hub writes to this end.
//   - clientConn is what the test reads from to verify received messages.
//
// The helper spins up a lightweight httptest.Server whose handler upgrades the
// connection, hands the server-side conn back via a channel, then drains any
// incoming frames so the connection stays alive until cleanup() is called.
func makeTestConn(t *testing.T) (serverConn *websocket.Conn, clientConn *websocket.Conn, cleanup func()) {
	t.Helper()

	ready := make(chan *websocket.Conn, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
		c, err := u.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		ready <- c
		// Drain client→server frames to keep the connection open.
		for {
			if _, _, err := c.ReadMessage(); err != nil {
				break
			}
		}
	}))

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")
	cConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	sConn := <-ready

	cleanup = func() {
		_ = cConn.Close()
		_ = sConn.Close()
		ts.Close()
	}
	return sConn, cConn, cleanup
}

// readEventWithTimeout reads one GameEvent from a client WebSocket connection.
// Returns an error if no message arrives within the given deadline.
func readEventWithTimeout(clientConn *websocket.Conn, timeout time.Duration) (*GameEvent, error) {
	_ = clientConn.SetReadDeadline(time.Now().Add(timeout))
	defer func() { _ = clientConn.SetReadDeadline(time.Time{}) }()

	var ev GameEvent
	if err := clientConn.ReadJSON(&ev); err != nil {
		return nil, err
	}
	return &ev, nil
}

// ---------------------------------------------------------------------------
// 1. Basic construction
// ---------------------------------------------------------------------------

func TestNewGameEventHub(t *testing.T) {
	hub := NewGameEventHub()
	require.NotNil(t, hub)
	assert.NotNil(t, hub.conns)
}

// ---------------------------------------------------------------------------
// 2. ConnCount — empty game
// ---------------------------------------------------------------------------

func TestGameEventHub_ConnCount_Empty(t *testing.T) {
	hub := NewGameEventHub()
	gameID := uuid.New()
	assert.Equal(t, 0, hub.ConnCount(gameID))
}

// ---------------------------------------------------------------------------
// 3. HasConn — unregistered user
// ---------------------------------------------------------------------------

func TestGameEventHub_HasConn_Missing(t *testing.T) {
	hub := NewGameEventHub()
	gameID := uuid.New()
	userID := uuid.New()
	assert.False(t, hub.HasConn(gameID, userID))
}

// ---------------------------------------------------------------------------
// 4. Register increments ConnCount and marks HasConn true
// ---------------------------------------------------------------------------

func TestGameEventHub_Register_And_ConnCount(t *testing.T) {
	hub := NewGameEventHub()
	gameID := uuid.New()
	userID := uuid.New()

	// nil conn is fine for map-management tests — no writes will be attempted.
	hub.Register(gameID, userID, nil, false)
	assert.Equal(t, 1, hub.ConnCount(gameID))
	assert.True(t, hub.HasConn(gameID, userID))

	// A second user in the same game
	user2 := uuid.New()
	hub.Register(gameID, user2, nil, true)
	assert.Equal(t, 2, hub.ConnCount(gameID))
}

// ---------------------------------------------------------------------------
// 5. isGM flag — BroadcastExceptFiltered honours the GM filter
// ---------------------------------------------------------------------------

func TestGameEventHub_Register_With_IsGM(t *testing.T) {
	hub := NewGameEventHub()
	gameID := uuid.New()
	gmUserID := uuid.New()
	playerUserID := uuid.New()

	gmServer, gmClient, cleanupGM := makeTestConn(t)
	defer cleanupGM()
	playerServer, playerClient, cleanupPlayer := makeTestConn(t)
	defer cleanupPlayer()

	hub.Register(gameID, gmUserID, gmServer, true)
	hub.Register(gameID, playerUserID, playerServer, false)

	event := GameEvent{Type: "gm_only", GameID: gameID}
	gmOnlyFilter := BroadcastFilter(func(_ uuid.UUID, isGM bool) bool { return isGM })
	hub.BroadcastExceptFiltered(gameID, uuid.Nil, gmOnlyFilter, event)

	// GM connection SHOULD receive the event.
	ev, err := readEventWithTimeout(gmClient, 400*time.Millisecond)
	require.NoError(t, err, "GM should receive the GM-only broadcast")
	assert.Equal(t, "gm_only", ev.Type)

	// Player connection should NOT receive the event.
	_, err = readEventWithTimeout(playerClient, 200*time.Millisecond)
	assert.Error(t, err, "player should not receive the GM-only broadcast")
}

// ---------------------------------------------------------------------------
// 6. Unregister — matching conn removes the entry
// ---------------------------------------------------------------------------

func TestGameEventHub_Unregister_MatchesConn(t *testing.T) {
	hub := NewGameEventHub()
	gameID := uuid.New()
	userID := uuid.New()

	sConn, _, cleanup := makeTestConn(t)
	defer cleanup()

	hub.Register(gameID, userID, sConn, false)
	require.True(t, hub.HasConn(gameID, userID))

	hub.Unregister(gameID, userID, sConn)

	assert.False(t, hub.HasConn(gameID, userID))
	assert.Equal(t, 0, hub.ConnCount(gameID))
}

// ---------------------------------------------------------------------------
// 7. Unregister — wrong conn leaves the registration intact
// ---------------------------------------------------------------------------

// This mirrors the real-world scenario where a stale goroutine's deferred
// Unregister must not evict the entry created by a fresher reconnect.
func TestGameEventHub_Unregister_WrongConn_DoesNotRemove(t *testing.T) {
	hub := NewGameEventHub()
	gameID := uuid.New()
	userID := uuid.New()

	connA, _, cleanupA := makeTestConn(t)
	defer cleanupA()
	connB, _, cleanupB := makeTestConn(t)
	defer cleanupB()

	hub.Register(gameID, userID, connA, false)

	// Unregister with connB — must NOT remove connA's registration.
	hub.Unregister(gameID, userID, connB)

	assert.True(t, hub.HasConn(gameID, userID),
		"registration should survive an Unregister call that carries the wrong conn pointer")
}

// ---------------------------------------------------------------------------
// 8. Register replaces an existing conn and closes the old one
// ---------------------------------------------------------------------------

func TestGameEventHub_Register_ClosesExistingConn(t *testing.T) {
	hub := NewGameEventHub()
	gameID := uuid.New()
	userID := uuid.New()

	sConn1, cConn1, cleanup1 := makeTestConn(t)
	defer cleanup1()
	sConn2, _, cleanup2 := makeTestConn(t)
	defer cleanup2()

	hub.Register(gameID, userID, sConn1, false)
	// Re-register same user — hub should close sConn1.
	hub.Register(gameID, userID, sConn2, false)

	// cConn1 must detect the close (close frame or connection error).
	_ = cConn1.SetReadDeadline(time.Now().Add(400 * time.Millisecond))
	_, _, err := cConn1.ReadMessage()
	assert.Error(t, err, "old client conn should be closed after re-registration")
}

// ---------------------------------------------------------------------------
// 9. Broadcast on an empty game — must not panic
// ---------------------------------------------------------------------------

func TestGameEventHub_Broadcast_Empty(t *testing.T) {
	hub := NewGameEventHub()
	gameID := uuid.New()
	assert.NotPanics(t, func() {
		hub.Broadcast(gameID, GameEvent{Type: "noop", GameID: gameID})
	})
}

// ---------------------------------------------------------------------------
// 10. BroadcastExcept — excluded user does not receive, others do
// ---------------------------------------------------------------------------

func TestGameEventHub_BroadcastExcept_ExcludesUser(t *testing.T) {
	hub := NewGameEventHub()
	gameID := uuid.New()
	user1 := uuid.New()
	user2 := uuid.New()

	sConn1, cConn1, cleanup1 := makeTestConn(t)
	defer cleanup1()
	sConn2, cConn2, cleanup2 := makeTestConn(t)
	defer cleanup2()

	hub.Register(gameID, user1, sConn1, false)
	hub.Register(gameID, user2, sConn2, false)

	event := GameEvent{Type: "update", GameID: gameID}
	hub.BroadcastExcept(gameID, user1, event)

	// user1 must NOT receive (excluded).
	_, err := readEventWithTimeout(cConn1, 200*time.Millisecond)
	assert.Error(t, err, "excluded user should not receive BroadcastExcept")

	// user2 MUST receive.
	ev, err := readEventWithTimeout(cConn2, 400*time.Millisecond)
	require.NoError(t, err, "non-excluded user should receive BroadcastExcept")
	assert.Equal(t, "update", ev.Type)
}

// ---------------------------------------------------------------------------
// 11. BroadcastExceptFiltered with nil filter — same semantics as BroadcastExcept
// ---------------------------------------------------------------------------

func TestGameEventHub_BroadcastExceptFiltered_NilFilter(t *testing.T) {
	hub := NewGameEventHub()
	gameID := uuid.New()
	user1 := uuid.New()
	user2 := uuid.New()

	sConn1, cConn1, cleanup1 := makeTestConn(t)
	defer cleanup1()
	sConn2, cConn2, cleanup2 := makeTestConn(t)
	defer cleanup2()

	hub.Register(gameID, user1, sConn1, false)
	hub.Register(gameID, user2, sConn2, true) // isGM — but nil filter ignores it

	event := GameEvent{Type: "broadcast", GameID: gameID}
	// nil filter = pass all non-excluded recipients.
	hub.BroadcastExceptFiltered(gameID, user1, nil, event)

	// user1 excluded — must NOT receive.
	_, err := readEventWithTimeout(cConn1, 200*time.Millisecond)
	assert.Error(t, err, "excluded user should not receive when using nil filter")

	// user2 must receive regardless of isGM value.
	ev, err := readEventWithTimeout(cConn2, 400*time.Millisecond)
	require.NoError(t, err, "non-excluded user should receive with nil filter")
	assert.Equal(t, "broadcast", ev.Type)
}

// ---------------------------------------------------------------------------
// 12. SendToUser — delivers only to the target user
// ---------------------------------------------------------------------------

func TestGameEventHub_SendToUser_DeliversSingle(t *testing.T) {
	hub := NewGameEventHub()
	gameID := uuid.New()
	user1 := uuid.New()
	user2 := uuid.New()

	sConn1, cConn1, cleanup1 := makeTestConn(t)
	defer cleanup1()
	sConn2, cConn2, cleanup2 := makeTestConn(t)
	defer cleanup2()

	hub.Register(gameID, user1, sConn1, false)
	hub.Register(gameID, user2, sConn2, false)

	event := GameEvent{Type: "targeted", GameID: gameID}
	hub.SendToUser(gameID, user1, event)

	// user1 MUST receive.
	ev, err := readEventWithTimeout(cConn1, 400*time.Millisecond)
	require.NoError(t, err, "target user should receive SendToUser message")
	assert.Equal(t, "targeted", ev.Type)

	// user2 must NOT receive.
	_, err = readEventWithTimeout(cConn2, 200*time.Millisecond)
	assert.Error(t, err, "non-target user should not receive SendToUser message")
}

// ---------------------------------------------------------------------------
// 13. SendToUser — no registered connection is a no-op (no panic)
// ---------------------------------------------------------------------------

func TestGameEventHub_SendToUser_NoConnection(t *testing.T) {
	hub := NewGameEventHub()
	gameID := uuid.New()
	userID := uuid.New()

	assert.NotPanics(t, func() {
		hub.SendToUser(gameID, userID, GameEvent{Type: "test"})
	})
}

// ---------------------------------------------------------------------------
// 14. SetLogger — callback fires when a write fails on a dead connection
// ---------------------------------------------------------------------------

func TestGameEventHub_SetLogger_CalledOnWriteError(t *testing.T) {
	hub := NewGameEventHub()
	gameID := uuid.New()
	userID := uuid.New()

	var mu sync.Mutex
	var logCalled bool
	hub.SetLogger(func(_ string, _ ...interface{}) {
		mu.Lock()
		logCalled = true
		mu.Unlock()
	})

	sConn, _, cleanup := makeTestConn(t)
	defer cleanup()

	hub.Register(gameID, userID, sConn, false)

	// Close the server-side conn directly — subsequent WriteJSON will fail.
	_ = sConn.Close()

	// Trigger a broadcast; writeOrPrune should detect the error and invoke the logger.
	hub.Broadcast(gameID, GameEvent{Type: "test", GameID: gameID})

	mu.Lock()
	called := logCalled
	mu.Unlock()
	assert.True(t, called, "SetLogger callback should be invoked when a write fails on a closed connection")
}

// ---------------------------------------------------------------------------
// 15. Concurrent access — verified clean under -race
// ---------------------------------------------------------------------------

func TestGameEventHub_ConcurrentAccess(t *testing.T) {
	hub := NewGameEventHub()
	gameID := uuid.New()

	// Part A: concurrent Register / HasConn / ConnCount / Unregister.
	// Each goroutine operates on its own unique userID so there is no
	// intentional key collision; the race detector verifies that the shared
	// map is correctly protected by hub.mu throughout.
	var wg sync.WaitGroup
	const workers = 10
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(isGM bool) {
			defer wg.Done()
			userID := uuid.New()
			hub.Register(gameID, userID, nil, isGM)
			_ = hub.HasConn(gameID, userID)
			_ = hub.ConnCount(gameID)
			hub.Unregister(gameID, userID, nil)
			_ = hub.ConnCount(gameID)
		}(i%2 == 0)
	}
	wg.Wait()

	// Part B: concurrent serialised writes to the same real connection.
	// writeOrPrune must protect concurrent callers via the per-conn mutex.
	sConn, cConn, cleanup := makeTestConn(t)
	defer cleanup()

	subUser := uuid.New()
	hub.Register(gameID, subUser, sConn, false)

	// Drain cConn in the background so the write buffer never fills.
	drainDone := make(chan struct{})
	go func() {
		defer close(drainDone)
		for {
			_ = cConn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			if _, _, err := cConn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	event := GameEvent{Type: "concurrent_write", GameID: gameID}
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			hub.Broadcast(gameID, event)
		}()
	}
	wg.Wait()

	// Signal the drain goroutine to exit and wait for it.
	_ = cConn.Close()
	<-drainDone
}
