package auditor

import (
	"log"

	"github.com/ar4ie13/shortener/internal/auditor/config"
	"github.com/ar4ie13/shortener/internal/auditor/file"
	"github.com/ar4ie13/shortener/internal/auditor/remote"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type Observer interface {
	Update(action string, userUUID uuid.UUID, url string) error
}
type Auditor struct {
	cfg       config.AuditConf
	observers []Observer
	zlog      zerolog.Logger
}

func NewAuditor(cfg config.AuditConf, zlog zerolog.Logger) *Auditor {
	auditor := &Auditor{
		cfg:  cfg,
		zlog: zlog,
	}
	if auditor.cfg.FileConf.AuditFilePath != "" {
		fileAudit, err := file.NewAuditFileLogger(auditor.cfg.FileConf, auditor.zlog)
		if err != nil {
			log.Fatal(err)
		}
		auditor.Attach(fileAudit)
		zlog.Debug().Msgf("Writing to audit file: %s", auditor.cfg.FileConf.AuditFilePath)
	}
	if auditor.cfg.RemoteConf.RemoteServerURL != "" {
		remoteAudit := remote.NewAuditRemoteLogger(auditor.cfg.RemoteConf, auditor.zlog)
		auditor.Attach(remoteAudit)
		zlog.Debug().Msgf("Writing to audit host: %s", auditor.cfg.RemoteConf.RemoteServerURL)
	}
	return auditor
}

// Attach добавляет наблюдателя
func (s *Auditor) Attach(o Observer) {
	s.observers = append(s.observers, o)
}

// Notify оповещает всех наблюдателей
func (s *Auditor) Notify(action string, userUUID uuid.UUID, url string) {
	for _, observer := range s.observers {
		err := observer.Update(action, userUUID, url)
		if err != nil {
			s.zlog.Error().Msgf("Error updating auditor observer: %v", err)
		}
	}
}
