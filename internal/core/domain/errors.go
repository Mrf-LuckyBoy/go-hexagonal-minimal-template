package domain

import "errors"

// Domain errors shared by every adapter. Services and driven adapters wrap
// them (fmt.Errorf("...: %w", domain.ErrNotFound)); driving adapters check
// them with errors.Is to pick a response, e.g. HTTP 404 / 400 / 409.
var (
	ErrNotFound     = errors.New("not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrConflict     = errors.New("conflict")
)
