package id

import (
	"github.com/google/uuid"
)

// New creates a new unique id.
func New() string {
	return uuid.New().String()
}
