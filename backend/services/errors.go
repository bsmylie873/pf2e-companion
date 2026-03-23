package services

import "errors"

// ErrForbidden is returned when a user lacks permission for the requested resource.
var ErrForbidden = errors.New("forbidden")
