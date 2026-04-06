package models

import (
	"time"

	"github.com/google/uuid"
)

// BackupFile is the top-level container for an export/import backup JSON file.
type BackupFile struct {
	SchemaVersion string    `json:"schema_version"`
	GameID        uuid.UUID `json:"game_id"`
	ExportedAt    time.Time `json:"exported_at"`
	Sessions      []Session `json:"sessions"`
	Notes         []Note    `json:"notes"`
}

// ImportSummary reports how many records were created, skipped, or overwritten during import.
type ImportSummary struct {
	SessionsCreated     int `json:"sessions_created"`
	SessionsSkipped     int `json:"sessions_skipped"`
	SessionsOverwritten int `json:"sessions_overwritten"`
	NotesCreated        int `json:"notes_created"`
	NotesSkipped        int `json:"notes_skipped"`
	NotesOverwritten    int `json:"notes_overwritten"`
}
