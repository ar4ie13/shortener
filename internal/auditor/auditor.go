// Package auditor is used to perform audit to file or/and remote server depending on specified
// flags or environment variables.
package auditor

import (
	"log"

	"github.com/ar4ie13/shortener/internal/auditor/config"
	"github.com/ar4ie13/shortener/internal/auditor/file"
	"github.com/ar4ie13/shortener/internal/auditor/remote"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Observer is an observer interfae
type Observer interface {
	Update(action string, userUUID uuid.UUID, url string) error
}

// Auditor implements Observer interface
type Auditor struct {
	enabled   bool
	cfg       config.AuditConf
	observers []Observer
	zlog      zerolog.Logger
}

// NewAuditor creates new Auditor instance
func NewAuditor(cfg config.AuditConf, zlog zerolog.Logger) *Auditor {
	auditor := &Auditor{
		enabled: false,
		cfg:     cfg,
		zlog:    zlog,
	}
	if auditor.cfg.FileConf.AuditFilePath != "" {
		fileAudit, err := file.NewAuditFileLogger(auditor.cfg.FileConf, auditor.zlog)
		if err != nil {
			log.Fatal(err)
		}
		auditor.Attach(fileAudit)
		zlog.Info().Msgf("Writing to audit file: %s", auditor.cfg.FileConf.AuditFilePath)
		auditor.enabled = true
	}
	if auditor.cfg.RemoteConf.RemoteServerURL != "" {
		remoteAudit := remote.NewAuditRemoteLogger(auditor.cfg.RemoteConf, auditor.zlog)
		auditor.Attach(remoteAudit)
		zlog.Info().Msgf("Writing to audit host: %s", auditor.cfg.RemoteConf.RemoteServerURL)
		auditor.enabled = true
	}
	return auditor
}

// Attach adds new observer
func (s *Auditor) Attach(o Observer) {
	s.observers = append(s.observers, o)
}

// Notify sends notification to all observers when the event occurs
func (s *Auditor) Notify(action string, userUUID uuid.UUID, url string) {
	if s.enabled {
		for _, observer := range s.observers {
			err := observer.Update(action, userUUID, url)
			if err != nil {
				s.zlog.Error().Msgf("Error updating auditor observer: %v", err)
			}
		}
	}
}
