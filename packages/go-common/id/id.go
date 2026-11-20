package id

import "github.com/google/uuid"

// New returns a UUIDv7 (time ordered) identifier.
func New() string {
	u, err := uuid.NewV7()
	if err != nil {
		return uuid.NewString()
	}
	return u.String()
}
