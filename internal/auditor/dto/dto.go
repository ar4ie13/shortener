// Package dto contains struct used by auditor service
package dto

import (
	"github.com/google/uuid"
)

// AuditRequest is a dto for audit entries
type AuditRequest struct {
	TS     int64     `json:"ts"`
	Action string    `json:"action"`
	UserID uuid.UUID `json:"user_id"`
	URL    string    `json:"url"`
}
