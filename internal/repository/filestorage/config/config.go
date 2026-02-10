// Package config contains file storage configuration.
package config

// Config contains filepath to file storage
type Config struct {
	FilePath string `json:"file_path,omitempty"`
}
