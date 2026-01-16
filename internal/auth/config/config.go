// Package config contains authentication service configuration for auditor service.
package config

import "time"

type Config struct {
	SecretKey       string
	TokenExpiration time.Duration
}
