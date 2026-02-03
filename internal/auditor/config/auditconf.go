// Package config used for configuration of auditor service.
package config

import (
	fileconf "github.com/ar4ie13/shortener/internal/auditor/file/config"
	remoteconf "github.com/ar4ie13/shortener/internal/auditor/remote/config"
)

// AuditConf contains configuration for both file and remote host audits
type AuditConf struct {
	Enabled    bool                         `json:"enabled,omitempty"`
	FileConf   fileconf.FileAuditConfig     `json:"file_audit_config,omitempty"`
	RemoteConf remoteconf.RemoteAuditConfig `json:"remote_audit_config,omitempty"`
}
