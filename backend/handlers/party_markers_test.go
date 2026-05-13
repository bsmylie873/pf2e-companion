package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	authmw "pf2e-companion/backend/middleware"
	"pf2e-companion/backend/mocks"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/services"
)

// newPartyMarkerHandler builds a handler wired to a fresh hub.
func newPartyMarkerHandler(svc *mocks.MockPartyMarkerService) *PartyMarkerHandler {
	return &PartyMarkerHandler{service: svc, hub: NewGameEventHub()}
}

// partyMarkerCtx builds an echo.Context for /games/:id routes.
func partyMarkerCtx(e *echo.Echo, method, body string, gameID, authUserID uuid.UUID) (echo.Context, *httptest.ResponseRecorder) {
	var bodyReader *strings.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	} else {
		bodyReader = strings.NewReader("")
	}
	req := httptest.NewRequest(method, "/", bodyReader)
	if body != "" {
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)
	return c, rec
}

// -- GetPartyMarker --

func TestPartyMarkerHandler_GetPartyMarker_Success(t *testing.T) {
	mockSvc := &mocks.MockPartyMarkerService{}
	h := newPartyMarkerHandler(mockSvc)
	e := echo.New()

	authUserID := uuid.New()
	gameID := uuid.New()
	marker := &models.PartyMarker{ID: uuid.New(), GameID: gameID}

	mockSvc.On("GetPartyMarker", gameID, authUserID).Return(marker, nil)

	c, rec := partyMarkerCtx(e, http.MethodGet, "", gameID, authUserID)
	err := h.GetPartyMarker(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPartyMarkerHandler_GetPartyMarker_ReturnsNullWhenNone(t *testing.T) {
	mockSvc := &mocks.MockPartyMarkerService{}
	h := newPartyMarkerHandler(mockSvc)
	e := echo.New()

	authUserID := uuid.New()
	gameID := uuid.New()

	mockSvc.On("GetPartyMarker", gameID, authUserID).Return((*models.PartyMarker)(nil), nil)

	c, rec := partyMarkerCtx(e, http.MethodGet, "", gameID, authUserID)
	err := h.GetPartyMarker(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "null")
	mockSvc.AssertExpectations(t)
}

func TestPartyMarkerHandler_GetPartyMarker_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockPartyMarkerService{}
	h := newPartyMarkerHandler(mockSvc)
	e := echo.New()

	authUserID := uuid.New()
	gameID := uuid.New()

	mockSvc.On("GetPartyMarker", gameID, authUserID).Return((*models.PartyMarker)(nil), services.ErrForbidden)

	c, rec := partyMarkerCtx(e, http.MethodGet, "", gameID, authUserID)
	err := h.GetPartyMarker(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

// -- UpsertPartyMarker --

func TestPartyMarkerHandler_UpsertPartyMarker_Creates(t *testing.T) {
	mockSvc := &mocks.MockPartyMarkerService{}
	h := newPartyMarkerHandler(mockSvc)
	e := echo.New()

	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()
	marker := &models.PartyMarker{ID: uuid.New(), GameID: gameID, MapID: mapID, X: 0.5, Y: 0.3}

	body := fmt.Sprintf(`{"map_id":%q,"x":0.5,"y":0.3}`, mapID.String())

	// GetPartyMarker returns nil (not yet created) → 201
	mockSvc.On("GetPartyMarker", gameID, authUserID).Return((*models.PartyMarker)(nil), nil).Once()
	mockSvc.On("UpsertPartyMarker", gameID, authUserID, mapID, 0.5, 0.3).Return(marker, nil).Once()

	c, rec := partyMarkerCtx(e, http.MethodPut, body, gameID, authUserID)
	err := h.UpsertPartyMarker(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPartyMarkerHandler_UpsertPartyMarker_Updates(t *testing.T) {
	mockSvc := &mocks.MockPartyMarkerService{}
	h := newPartyMarkerHandler(mockSvc)
	e := echo.New()

	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()
	existing := &models.PartyMarker{ID: uuid.New(), GameID: gameID}
	updated := &models.PartyMarker{ID: existing.ID, GameID: gameID, MapID: mapID, X: 0.5, Y: 0.3}

	body := fmt.Sprintf(`{"map_id":%q,"x":0.5,"y":0.3}`, mapID.String())

	// GetPartyMarker returns existing → 200
	mockSvc.On("GetPartyMarker", gameID, authUserID).Return(existing, nil).Once()
	mockSvc.On("UpsertPartyMarker", gameID, authUserID, mapID, 0.5, 0.3).Return(updated, nil).Once()

	c, rec := partyMarkerCtx(e, http.MethodPut, body, gameID, authUserID)
	err := h.UpsertPartyMarker(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPartyMarkerHandler_UpsertPartyMarker_InvalidMapID(t *testing.T) {
	mockSvc := &mocks.MockPartyMarkerService{}
	h := newPartyMarkerHandler(mockSvc)
	e := echo.New()

	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"map_id":"not-a-valid-uuid","x":0.5,"y":0.3}`

	c, rec := partyMarkerCtx(e, http.MethodPut, body, gameID, authUserID)
	err := h.UpsertPartyMarker(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPartyMarkerHandler_UpsertPartyMarker_ValidationError(t *testing.T) {
	mockSvc := &mocks.MockPartyMarkerService{}
	h := newPartyMarkerHandler(mockSvc)
	e := echo.New()

	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()

	body := fmt.Sprintf(`{"map_id":%q,"x":150.0,"y":0.3}`, mapID.String())

	// GetPartyMarker succeeds (nil = no existing), then UpsertPartyMarker returns validation error
	mockSvc.On("GetPartyMarker", gameID, authUserID).Return((*models.PartyMarker)(nil), nil).Once()
	valErr := fmt.Errorf("x must be between 0 and 100: %w", services.ErrValidation)
	mockSvc.On("UpsertPartyMarker", gameID, authUserID, mapID, 150.0, 0.3).Return((*models.PartyMarker)(nil), valErr).Once()

	c, rec := partyMarkerCtx(e, http.MethodPut, body, gameID, authUserID)
	err := h.UpsertPartyMarker(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPartyMarkerHandler_UpsertPartyMarker_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockPartyMarkerService{}
	h := newPartyMarkerHandler(mockSvc)
	e := echo.New()

	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()

	body := fmt.Sprintf(`{"map_id":%q,"x":0.5,"y":0.3}`, mapID.String())

	// GetPartyMarker returns forbidden
	mockSvc.On("GetPartyMarker", gameID, authUserID).Return((*models.PartyMarker)(nil), services.ErrForbidden).Once()

	c, rec := partyMarkerCtx(e, http.MethodPut, body, gameID, authUserID)
	err := h.UpsertPartyMarker(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

// -- DeletePartyMarker --

func TestPartyMarkerHandler_DeletePartyMarker_Success(t *testing.T) {
	mockSvc := &mocks.MockPartyMarkerService{}
	h := newPartyMarkerHandler(mockSvc)
	e := echo.New()

	authUserID := uuid.New()
	gameID := uuid.New()

	mockSvc.On("DeletePartyMarker", gameID, authUserID).Return(nil)

	c, rec := partyMarkerCtx(e, http.MethodDelete, "", gameID, authUserID)
	err := h.DeletePartyMarker(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "deleted")
	mockSvc.AssertExpectations(t)
}

func TestPartyMarkerHandler_DeletePartyMarker_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockPartyMarkerService{}
	h := newPartyMarkerHandler(mockSvc)
	e := echo.New()

	authUserID := uuid.New()
	gameID := uuid.New()

	mockSvc.On("DeletePartyMarker", gameID, authUserID).Return(services.ErrForbidden)

	c, rec := partyMarkerCtx(e, http.MethodDelete, "", gameID, authUserID)
	err := h.DeletePartyMarker(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}
