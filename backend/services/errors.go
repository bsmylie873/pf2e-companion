package services

import "errors"

// ErrForbidden is returned when a user lacks permission for the requested resource.
var ErrForbidden = errors.New("forbidden")

// ErrGroupedPinMove is returned when a user tries to reposition a pin that belongs to a group.
var ErrGroupedPinMove = errors.New("pin belongs to a group; remove it from the group before repositioning")

// ErrValidation is returned when input fails business-rule validation.
var ErrValidation = errors.New("validation error")

// ErrConflict is returned when a unique constraint would be violated.
var ErrConflict = errors.New("conflict")

// ErrSessionFoldersReadOnly is returned when a player tries to modify session folders.
var ErrSessionFoldersReadOnly = errors.New("session_folders_read_only")
