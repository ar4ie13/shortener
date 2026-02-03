// Package config contains authentication service configuration for auditor service.
package config

import "time"

type Config struct {
	SecretKey       string        `json:"secret_key,omitempty"`
	TokenExpiration time.Duration `json:"token_expiration,omitempty"`
}
