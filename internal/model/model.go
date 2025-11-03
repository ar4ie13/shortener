package model

import "github.com/google/uuid"

// URL struct is a main struct used by service
type URL struct {
	UUID        uuid.UUID `json:"uuid" db:"uuid"`
	UserUUID    uuid.UUID `json:"user_uuid" db:"user_uuid"`
	ShortURL    string    `json:"short_url" db:"short_url"`
	OriginalURL string    `json:"original_url" db:"original_url"`
}
