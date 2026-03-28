package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
)

// User represents the users table.
type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Username     string         `gorm:"not null;uniqueIndex"                            json:"username"`
	Email        string         `gorm:"not null;uniqueIndex"                            json:"email"`
	PasswordHash string         `gorm:"column:password_hash;not null"                   json:"password_hash"`
	AvatarURL    *string        `gorm:"column:avatar_url"                               json:"avatar_url"`
	Description  *string        `gorm:"column:description"                              json:"description"`
	Location     *string        `gorm:"column:location"                                 json:"location"`
	FoundryData  datatypes.JSON `gorm:"column:foundry_data;type:jsonb"                  json:"foundry_data"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"                                  json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"                                  json:"updated_at"`
}

func (User) TableName() string { return "users" }

// Game represents the games table.
type Game struct {
	ID             uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Title          string         `gorm:"not null"                                       json:"title"`
	Description    *string        `                                                      json:"description"`
	SplashImageURL *string        `gorm:"column:splash_image_url"                        json:"splash_image_url"`
	MapImageURL    *string        `gorm:"column:map_image_url"                           json:"map_image_url"`
	FoundryData    datatypes.JSON `gorm:"column:foundry_data;type:jsonb"                 json:"foundry_data"`
	CreatedAt      time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (Game) TableName() string { return "games" }

// GameMembership represents the game_memberships table.
type GameMembership struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	GameID      uuid.UUID      `gorm:"type:uuid;not null;column:game_id"              json:"game_id"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null;column:user_id"              json:"user_id"`
	IsGM        bool           `gorm:"column:is_gm;not null;default:false"            json:"is_gm"`
	FoundryData datatypes.JSON `gorm:"column:foundry_data;type:jsonb"                 json:"foundry_data"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (GameMembership) TableName() string { return "game_memberships" }

// Session represents the sessions table.
type Session struct {
	ID            uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	GameID        uuid.UUID      `gorm:"type:uuid;not null;column:game_id"              json:"game_id"`
	Title         string         `gorm:"not null"                                       json:"title"`
	SessionNumber *int           `gorm:"column:session_number"                          json:"session_number"`
	ScheduledAt   *time.Time     `gorm:"column:scheduled_at"                            json:"scheduled_at"`
	RuntimeStart  *time.Time     `gorm:"column:runtime_start"                           json:"runtime_start"`
	RuntimeEnd    *time.Time     `gorm:"column:runtime_end"                             json:"runtime_end"`
	Notes         datatypes.JSON `gorm:"column:notes;type:jsonb"                        json:"notes"`
	Version       int            `gorm:"not null;default:1"                             json:"version"`
	FoundryData   datatypes.JSON `gorm:"column:foundry_data;type:jsonb"                 json:"foundry_data"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (Session) TableName() string { return "sessions" }

// SessionPin represents a map pin associated with a session.
type SessionPin struct {
	ID          uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	GameID      uuid.UUID  `gorm:"type:uuid;not null;column:game_id"              json:"game_id"`
	SessionID   *uuid.UUID `gorm:"type:uuid;column:session_id"                    json:"session_id"`
	NoteID      *uuid.UUID `gorm:"type:uuid;column:note_id"                       json:"note_id"`
	Label       string     `gorm:"not null"                                       json:"label"`
	X           float64    `gorm:"type:numeric(6,4);not null"                     json:"x"`
	Y           float64    `gorm:"type:numeric(6,4);not null"                     json:"y"`
	Colour      string     `gorm:"not null;default:'grey'"                        json:"colour"`
	Icon        string     `gorm:"not null;default:'position-marker'"             json:"icon"`
	GroupID     *uuid.UUID `gorm:"type:uuid;column:group_id"                      json:"group_id"`
	Description *string    `                                                       json:"description"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (SessionPin) TableName() string { return "session_pins" }

// PinGroup represents a named cluster of session pins on the map.
type PinGroup struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	GameID    uuid.UUID `gorm:"type:uuid;not null;column:game_id"              json:"game_id"`
	X         float64   `gorm:"type:numeric(6,4);not null"                     json:"x"`
	Y         float64   `gorm:"type:numeric(6,4);not null"                     json:"y"`
	Colour    string    `gorm:"not null;default:'grey'"                        json:"colour"`
	Icon      string    `gorm:"not null;default:'position-marker'"             json:"icon"`
	CreatedAt time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (PinGroup) TableName() string { return "pin_groups" }

// Note represents the notes table.
type Note struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	GameID      uuid.UUID      `gorm:"type:uuid;not null;column:game_id"              json:"game_id"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null;column:user_id"              json:"user_id"`
	SessionID   *uuid.UUID     `gorm:"type:uuid;column:session_id"                    json:"session_id"`
	Title       string         `gorm:"not null"                                       json:"title"`
	Content     datatypes.JSON `gorm:"column:content;type:jsonb"                      json:"content"`
	Visibility  string         `gorm:"not null;default:'private'"                     json:"visibility"`
	Version     int            `gorm:"not null;default:1"                             json:"version"`
	FoundryData datatypes.JSON `gorm:"column:foundry_data;type:jsonb"                 json:"foundry_data"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (Note) TableName() string { return "notes" }

// Character represents the characters table.
type Character struct {
	ID           uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	GameID       uuid.UUID      `gorm:"type:uuid;not null;column:game_id"              json:"game_id"`
	UserID       *uuid.UUID     `gorm:"type:uuid;column:user_id"                       json:"user_id"`
	Name         string         `gorm:"not null"                                       json:"name"`
	IsNPC        bool           `gorm:"column:is_npc;not null;default:false"           json:"is_npc"`
	Ancestry     *string        `                                                      json:"ancestry"`
	Heritage     *string        `                                                      json:"heritage"`
	Class        *string        `gorm:"column:class"                                   json:"class"`
	Background   *string        `                                                      json:"background"`
	Level        int            `gorm:"not null;default:1"                             json:"level"`
	HPMax        *int           `gorm:"column:hp_max"                                  json:"hp_max"`
	HPCurrent    *int           `gorm:"column:hp_current"                              json:"hp_current"`
	AC           *int           `gorm:"column:ac"                                      json:"ac"`
	Strength     *int           `                                                      json:"strength"`
	Dexterity    *int           `                                                      json:"dexterity"`
	Constitution *int           `                                                      json:"constitution"`
	Intelligence *int           `                                                      json:"intelligence"`
	Wisdom       *int           `                                                      json:"wisdom"`
	Charisma     *int           `                                                      json:"charisma"`
	Fortitude    *int           `                                                      json:"fortitude"`
	Reflex       *int           `                                                      json:"reflex"`
	Will         *int           `                                                      json:"will"`
	Skills       datatypes.JSON `gorm:"column:skills;type:jsonb"                       json:"skills"`
	FoundryData  datatypes.JSON `gorm:"column:foundry_data;type:jsonb"                 json:"foundry_data"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (Character) TableName() string { return "characters" }

// Item represents the items table.
type Item struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	GameID      uuid.UUID      `gorm:"type:uuid;not null;column:game_id"              json:"game_id"`
	CharacterID *uuid.UUID     `gorm:"type:uuid;column:character_id"                  json:"character_id"`
	Name        string         `gorm:"not null"                                       json:"name"`
	Description *string        `                                                      json:"description"`
	Level       int            `gorm:"not null;default:0"                             json:"level"`
	PriceGP     *float64       `gorm:"column:price_gp;type:numeric(10,2)"             json:"price_gp"`
	Bulk        *string        `                                                      json:"bulk"`
	Traits      pq.StringArray `gorm:"type:text[]"                                    json:"traits"`
	Quantity    int            `gorm:"not null;default:1"                             json:"quantity"`
	FoundryData datatypes.JSON `gorm:"column:foundry_data;type:jsonb"                 json:"foundry_data"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (Item) TableName() string { return "items" }

// RefreshToken represents a stored refresh token for JWT rotation.
type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;column:user_id"               json:"user_id"`
	TokenHash string    `gorm:"column:token_hash;not null;uniqueIndex"          json:"token_hash"`
	ExpiresAt time.Time `gorm:"column:expires_at;not null"                     json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime"                                  json:"created_at"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }
