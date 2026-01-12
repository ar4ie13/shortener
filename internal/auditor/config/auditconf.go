package config

import (
	fileconf "github.com/ar4ie13/shortener/internal/auditor/file/config"
	remoteconf "github.com/ar4ie13/shortener/internal/auditor/remote/config"
)

// AuditConf contains configuration for both file and remote host audits
type AuditConf struct {
	Enabled    bool
	FileConf   fileconf.FileAuditConfig
	RemoteConf remoteconf.RemoteAuditConfig
}
