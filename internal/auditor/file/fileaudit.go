package file

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ar4ie13/shortener/internal/auditor/dto"
	cfg "github.com/ar4ie13/shortener/internal/auditor/file/config"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// AuditFileLogger implements Observer interface
type AuditFileLogger struct {
	file *os.File
	mu   sync.Mutex
	zlog zerolog.Logger
}

// NewAuditFileLogger creates new File Audit instance
func NewAuditFileLogger(conf cfg.FileAuditConfig, zlog zerolog.Logger) (*AuditFileLogger, error) {
	file, err := os.OpenFile(conf.AuditFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &AuditFileLogger{
		file: file,
		mu:   sync.Mutex{},
		zlog: zlog,
	}, nil
}

// Update sends jsonl audit string to remote audit host
func (a *AuditFileLogger) Update(action string, userUUID uuid.UUID, url string) error {
	entry := dto.AuditRequest{
		TS:     time.Now().Unix(),
		Action: action,
		UserID: userUUID,
		URL:    url,
	}

	jsonLine, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("cannot marshal audit entry to json: %w", err)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	_, err = a.file.Write(jsonLine)
	if err != nil {
		a.zlog.Error().Msgf("Audit log write error: %v", err)
		return fmt.Errorf("cannot write to file: %w", err)
	}
	_, err = a.file.WriteString("\n")
	if err != nil {
		return fmt.Errorf("cannot write to file: %w", err)
	}
	return nil
}
