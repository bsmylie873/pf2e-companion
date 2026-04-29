package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	authpkg "pf2e-companion/backend/auth"
	"pf2e-companion/backend/ot"
	"pf2e-companion/backend/repositories"
)

// GameWebSocket handles GET /games/:id/ws.
// Authenticates via access_token cookie (inline JWT check, since WS upgrade needs special handling).
func GameWebSocket(hub *GameEventHub, otStore *ot.DocumentStore, membershipRepo repositories.MembershipRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("access_token")
		if err != nil || cookie.Value == "" {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"code": 401, "message": "unauthorized"})
		}
		claims, err := authpkg.ValidateToken(cookie.Value)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"code": 401, "message": "unauthorized"})
		}

		userID := claims.UserID
		gameID, err := ParseUUID(c, "id")
		if err != nil {
			return nil
		}

		m, err := membershipRepo.FindByUserAndGameID(userID, gameID)
		if err != nil {
			return c.JSON(http.StatusForbidden, map[string]interface{}{"code": 403, "message": "forbidden"})
		}

		ws, err := wsUpgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}
		defer ws.Close()

		hub.Register(gameID, userID, ws, m.IsGM)
		defer hub.Unregister(gameID, userID, ws)

		// Read loop: handle OT messages from this client.
		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				break
			}

			// Peek at the message type.
			var envelope struct {
				Type string `json:"type"`
			}
			if err := json.Unmarshal(msg, &envelope); err != nil {
				continue
			}

			switch envelope.Type {
			case "ot_steps":
				var req struct {
					EntityID uuid.UUID `json:"entity_id"`
					Version  int       `json:"version"`
					Steps    []ot.Step `json:"steps"`
					ClientID string    `json:"client_id"`
				}
				if err := json.Unmarshal(msg, &req); err != nil {
					continue
				}

				doc := otStore.GetOrCreate(req.EntityID, req.Version)
				newVersion, applyErr := doc.ApplySteps(req.Version, req.Steps)

				if errors.Is(applyErr, ot.ErrVersionMismatch) {
					// Send missed steps back to the client so it can rebase.
					missedSteps := doc.StepsSince(req.Version)
					hub.SendToUser(gameID, userID, GameEvent{
						Type:   "ot_rebase",
						GameID: gameID,
						Data: map[string]interface{}{
							"entity_id": req.EntityID,
							"version":   newVersion,
							"steps":     missedSteps,
						},
					})
					continue
				}
				if applyErr != nil {
					continue
				}

				// Acknowledge to the sender.
				hub.SendToUser(gameID, userID, GameEvent{
					Type:   "ot_ack",
					GameID: gameID,
					Data: map[string]interface{}{
						"entity_id": req.EntityID,
						"version":   newVersion,
					},
				})

				// Broadcast the accepted steps to all other subscribers.
				hub.BroadcastExcept(gameID, userID, GameEvent{
					Type:   "ot_steps",
					GameID: gameID,
					Data: map[string]interface{}{
						"entity_id": req.EntityID,
						"version":   newVersion,
						"steps":     req.Steps,
						"client_id": req.ClientID,
					},
				})

			default:
				// Ignore unknown message types.
			}
		}
		return nil
	}
}
