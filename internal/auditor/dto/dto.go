package dto

import (
	"github.com/google/uuid"
)

type AuditRequest struct {
	TS     int64     `json:"ts"`
	Action string    `json:"action"`
	UserID uuid.UUID `json:"user_id"`
	URL    string    `json:"url"`
}
