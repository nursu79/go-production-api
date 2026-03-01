package domain

import "errors"

var (
	// ErrDuplicateEmail is returned when attempting to register an email that already exists.
	ErrDuplicateEmail = errors.New("email already registered")

	// ErrValidation is returned when input validation fails.
	ErrValidation = errors.New("validation failed")

	// ErrNotFound is returned when an entity is not found.
	ErrNotFound = errors.New("not found")
)
