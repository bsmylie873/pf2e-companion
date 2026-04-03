package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// UserResponse is a DTO that mirrors User but omits the password_hash field.
type UserResponse struct {
	ID          uuid.UUID      `json:"id"`
	Username    string         `json:"username"`
	Email       string         `json:"email"`
	AvatarURL   *string        `json:"avatar_url"`
	Description *string        `json:"description"`
	Location    *string        `json:"location"`
	FoundryData datatypes.JSON `json:"foundry_data"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// FromUser converts a User model to a UserResponse, stripping the password hash.
func FromUser(user User) UserResponse {
	return UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		AvatarURL:   user.AvatarURL,
		Description: user.Description,
		Location:    user.Location,
		FoundryData: user.FoundryData,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

// UserPublicResponse is a minimal public user DTO (no email, foundry_data, etc.)
type UserPublicResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	AvatarURL *string   `json:"avatar_url"`
}

// FromUserPublic converts a User to a UserPublicResponse.
func FromUserPublic(user User) UserPublicResponse {
	return UserPublicResponse{ID: user.ID, Username: user.Username, AvatarURL: user.AvatarURL}
}

// LoginRequest is the DTO for the POST /auth/login endpoint.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterRequest is the DTO for the POST /auth/register endpoint.
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TokenPair holds an access token and a refresh token.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// PinGroupResponse is the DTO returned for pin group endpoints.
type PinGroupResponse struct {
	ID        uuid.UUID    `json:"id"`
	GameID    uuid.UUID    `json:"game_id"`
	X         float64      `json:"x"`
	Y         float64      `json:"y"`
	Colour    string       `json:"colour"`
	Icon      string       `json:"icon"`
	PinCount  int          `json:"pin_count"`
	Pins      []SessionPin `json:"pins"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// UserPreferenceResponse is the DTO returned for preference endpoints.
type UserPreferenceResponse struct {
	DefaultGameID    *uuid.UUID     `json:"default_game_id"`
	DefaultPinColour *string        `json:"default_pin_colour"`
	DefaultPinIcon   *string        `json:"default_pin_icon"`
	SidebarState     datatypes.JSON `json:"sidebar_state"`
	DefaultViewMode  datatypes.JSON `json:"default_view_mode"`
	MapEditorMode    string         `json:"map_editor_mode"`
	PageSize         datatypes.JSON `json:"page_size"`
}

// FromUserPreference converts a UserPreference model to a UserPreferenceResponse.
func FromUserPreference(pref UserPreference) UserPreferenceResponse {
	return UserPreferenceResponse{
		DefaultGameID:    pref.DefaultGameID,
		DefaultPinColour: pref.DefaultPinColour,
		DefaultPinIcon:   pref.DefaultPinIcon,
		SidebarState:     pref.SidebarState,
		DefaultViewMode:  pref.DefaultViewMode,
		MapEditorMode:    pref.MapEditorMode,
		PageSize:         pref.PageSize,
	}
}
