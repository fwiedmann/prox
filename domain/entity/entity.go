package entity

import "github.com/google/uuid"

// ID for an entity
type ID string

// NewID generates a new unique ID
func NewID() ID {
	return ID(uuid.New().String())
}
