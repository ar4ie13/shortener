package config

import (
	fileconf "github.com/ar4ie13/shortener/internal/auditor/file/config"
	remoteconf "github.com/ar4ie13/shortener/internal/auditor/remote/config"
)

type AuditConf struct {
	FileConf   fileconf.FileAuditConfig
	RemoteConf remoteconf.RemoteAuditConfig
}
