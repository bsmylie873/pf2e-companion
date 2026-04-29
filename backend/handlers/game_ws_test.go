package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	authpkg "pf2e-companion/backend/auth"
	"pf2e-companion/backend/mocks"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/ot"
)

// setupGameWSServer starts an httptest.Server running the GameWebSocket handler
// wired into an Echo instance. Returns the ws:// URL for the route and a cleanup func.
func setupGameWSServer(
	t *testing.T,
	hub *GameEventHub,
	membershipRepo *mocks.MockMembershipRepository,
	gameID uuid.UUID,
) (wsURL string, cleanup func()) {
	t.Helper()
	e := echo.New()
	otStore := ot.NewDocumentStore()
	handler := GameWebSocket(hub, otStore, membershipRepo)
	e.GET("/games/:id/ws", handler)
	ts := httptest.NewServer(e)
	wsURL = "ws" + ts.URL[4:] + "/games/" + gameID.String() + "/ws"
	return wsURL, ts.Close
}

// dialWithToken dials a WebSocket URL carrying the given JWT in an access_token cookie.
func dialWithToken(t *testing.T, wsURL, token string) (*websocket.Conn, *http.Response, error) {
	t.Helper()
	header := http.Header{}
	header.Add("Cookie", "access_token="+token)
	return websocket.DefaultDialer.Dial(wsURL, header)
}

// ---------------------------------------------------------------------------
// 1. Non-member is rejected before the WS upgrade
// ---------------------------------------------------------------------------

func TestGameWebSocket_RejectsNonMembers(t *testing.T) {
	hub := NewGameEventHub()
	userID := uuid.New()
	gameID := uuid.New()

	mockRepo := &mocks.MockMembershipRepository{}
	mockRepo.On("FindByUserAndGameID", userID, gameID).
		Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	wsURL, cleanup := setupGameWSServer(t, hub, mockRepo, gameID)
	defer cleanup()

	token, err := authpkg.GenerateAccessToken(userID)
	require.NoError(t, err)

	_, resp, err := dialWithToken(t, wsURL, token)
	// Server returns 403 before upgrading — gorilla Dial should surface an error.
	assert.Error(t, err, "WS handshake should fail for a non-member")
	if resp != nil {
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	}
	mockRepo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// 2. Valid member is registered in the hub with the correct isGM flag
// ---------------------------------------------------------------------------

func TestGameWebSocket_RegistersWithIsGMFromMembership(t *testing.T) {
	hub := NewGameEventHub()
	userID := uuid.New()
	gameID := uuid.New()

	mockRepo := &mocks.MockMembershipRepository{}
	mockRepo.On("FindByUserAndGameID", userID, gameID).
		Return(models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}, nil)

	wsURL, cleanup := setupGameWSServer(t, hub, mockRepo, gameID)
	defer cleanup()

	token, err := authpkg.GenerateAccessToken(userID)
	require.NoError(t, err)

	conn, _, err := dialWithToken(t, wsURL, token)
	require.NoError(t, err)
	defer conn.Close()

	// Give the handler goroutine time to call hub.Register.
	time.Sleep(50 * time.Millisecond)

	assert.True(t, hub.HasConn(gameID, userID), "user should be registered after successful WS connect")
	assert.Equal(t, 1, hub.ConnCount(gameID))
	mockRepo.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// 3. Second connection for the same user closes the first (tab replacement)
// ---------------------------------------------------------------------------

func TestGameWebSocket_TabReplacement_OldSocketClosed(t *testing.T) {
	hub := NewGameEventHub()
	userID := uuid.New()
	gameID := uuid.New()

	mockRepo := &mocks.MockMembershipRepository{}
	// FindByUserAndGameID is called once per connection attempt.
	mockRepo.On("FindByUserAndGameID", userID, gameID).
		Return(models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}, nil)

	wsURL, cleanup := setupGameWSServer(t, hub, mockRepo, gameID)
	defer cleanup()

	token, err := authpkg.GenerateAccessToken(userID)
	require.NoError(t, err)

	// --- first connection ---
	connA, _, err := dialWithToken(t, wsURL, token)
	require.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, hub.ConnCount(gameID), "should have exactly one connection after first dial")

	// --- second connection (tab replacement) ---
	connB, _, err := dialWithToken(t, wsURL, token)
	require.NoError(t, err)
	defer connB.Close()
	// Give hub.Register time to close the server-side of connA.
	time.Sleep(50 * time.Millisecond)

	// connA's server side was closed — the client end must observe an error.
	_ = connA.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, _, err = connA.ReadMessage()
	assert.Error(t, err, "connA should be closed after tab replacement")

	// Hub should still hold exactly one registration (connB).
	assert.Equal(t, 1, hub.ConnCount(gameID))
	assert.True(t, hub.HasConn(gameID, userID))
}
