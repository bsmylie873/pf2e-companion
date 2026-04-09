package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUser_TableName(t *testing.T) {
	assert.Equal(t, "users", User{}.TableName())
}

func TestGame_TableName(t *testing.T) {
	assert.Equal(t, "games", Game{}.TableName())
}

func TestGameMembership_TableName(t *testing.T) {
	assert.Equal(t, "game_memberships", GameMembership{}.TableName())
}

func TestSession_TableName(t *testing.T) {
	assert.Equal(t, "sessions", Session{}.TableName())
}

func TestSessionPin_TableName(t *testing.T) {
	assert.Equal(t, "session_pins", SessionPin{}.TableName())
}

func TestPinGroup_TableName(t *testing.T) {
	assert.Equal(t, "pin_groups", PinGroup{}.TableName())
}

func TestGameMap_TableName(t *testing.T) {
	assert.Equal(t, "maps", GameMap{}.TableName())
}

func TestFolder_TableName(t *testing.T) {
	assert.Equal(t, "folders", Folder{}.TableName())
}

func TestNote_TableName(t *testing.T) {
	assert.Equal(t, "notes", Note{}.TableName())
}

func TestCharacter_TableName(t *testing.T) {
	assert.Equal(t, "characters", Character{}.TableName())
}

func TestItem_TableName(t *testing.T) {
	assert.Equal(t, "items", Item{}.TableName())
}

func TestRefreshToken_TableName(t *testing.T) {
	assert.Equal(t, "refresh_tokens", RefreshToken{}.TableName())
}

func TestPasswordResetToken_TableName(t *testing.T) {
	assert.Equal(t, "password_reset_tokens", PasswordResetToken{}.TableName())
}

func TestInviteToken_TableName(t *testing.T) {
	assert.Equal(t, "invite_tokens", InviteToken{}.TableName())
}

func TestUserPreference_TableName(t *testing.T) {
	assert.Equal(t, "user_preferences", UserPreference{}.TableName())
}
