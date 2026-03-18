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
		FoundryData: user.FoundryData,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}
