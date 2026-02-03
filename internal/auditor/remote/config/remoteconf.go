// Package config contains remote host configuration for auditor service.
package config

// RemoteAuditConfig contains remote server URL
type RemoteAuditConfig struct {
	RemoteServerURL string `json:"remote_server_url,omitempty"`
}
