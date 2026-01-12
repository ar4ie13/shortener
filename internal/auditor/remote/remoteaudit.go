package remote

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ar4ie13/shortener/internal/auditor/dto"
	"github.com/ar4ie13/shortener/internal/auditor/remote/config"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// AuditRemoteLogger implements observer interface for file audit logging
type AuditRemoteLogger struct {
	cfg  config.RemoteAuditConfig
	zlog zerolog.Logger
}

// NewAuditRemoteLogger create new Remote Audit instance
func NewAuditRemoteLogger(conf config.RemoteAuditConfig, zlog zerolog.Logger) *AuditRemoteLogger {
	return &AuditRemoteLogger{
		cfg:  conf,
		zlog: zlog,
	}
}

// Update writes jsonl audit string to audit file
func (a *AuditRemoteLogger) Update(action string, userUUID uuid.UUID, url string) error {
	entry := dto.AuditRequest{
		TS:     time.Now().Unix(),
		Action: action,
		UserID: userUUID,
		URL:    url,
	}

	jsonLine, err := json.Marshal(entry)
	if err != nil {
		a.zlog.Error().Msgf("Audit log write error: %v", err)
		return fmt.Errorf("cannot marshal audit entry to json: %w", err)
	}

	client := &http.Client{}
	request, err := http.NewRequest(http.MethodPost, a.cfg.RemoteServerURL, strings.NewReader(string(jsonLine)))
	if err != nil {
		a.zlog.Error().Msgf("Cannot create audit entry to remote server: %v", err)
		return fmt.Errorf("cannot create audit entry to remote server: %w", err)

	}
	request.Header.Add("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		a.zlog.Error().Msgf("Cannot send audit entry to remote server: %v, response status: %v", err, response.Status)
		return fmt.Errorf("cannot send audit entry to remote server: %w", err)
	}
	defer response.Body.Close()
	return nil
}
