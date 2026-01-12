package file

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ar4ie13/shortener/internal/auditor/dto"
	cfg "github.com/ar4ie13/shortener/internal/auditor/file/config"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuditFileLogger(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := tmpDir + "/audit.log"

	conf := cfg.FileAuditConfig{AuditFilePath: logPath}
	logger := zerolog.New(log.Logger)

	// Act
	auditLogger, err := NewAuditFileLogger(conf, logger)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, auditLogger)
	require.NotNil(t, auditLogger.file)

	// Ensure file was created
	_, err = os.Stat(logPath)
	require.NoError(t, err)

	// Close to avoid fd leak
	auditLogger.file.Close()
}

func TestNewAuditFileLogger_FileOpenError(t *testing.T) {
	conf := cfg.FileAuditConfig{AuditFilePath: "/invalid/path/audit.log"}
	logger := zerolog.Nop()

	_, err := NewAuditFileLogger(conf, logger)
	require.Error(t, err)
}

func TestAuditFileLogger_Update_Success(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := tmpDir + "/audit_update.log"

	conf := cfg.FileAuditConfig{AuditFilePath: logPath}
	logger := zerolog.Nop()

	auditLogger, err := NewAuditFileLogger(conf, logger)
	require.NoError(t, err)
	defer auditLogger.file.Close()

	userUUID := uuid.New()
	action := "shorten"
	url := "https://example.com/test"

	// Act
	err = auditLogger.Update(action, userUUID, url)
	require.NoError(t, err)

	// Read file content
	content, err := os.ReadFile(logPath)
	require.NoError(t, err)

	// Parse JSON line
	var entry dto.AuditRequest
	err = json.Unmarshal(content, &entry)
	require.NoError(t, err)

	// Validate fields
	assert.Equal(t, action, entry.Action)
	assert.Equal(t, userUUID, entry.UserID)
	assert.Equal(t, url, entry.URL)
	assert.Greater(t, entry.TS, int64(0))

	// Check it's valid JSONL (ends with newline)
	assert.True(t, strings.HasSuffix(string(content), "\n"))
}

func TestAuditFileLogger_Update_Concurrent(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := tmpDir + "/audit_concurrent.log"

	conf := cfg.FileAuditConfig{AuditFilePath: logPath}
	logger := zerolog.Nop()

	auditLogger, err := NewAuditFileLogger(conf, logger)
	require.NoError(t, err)
	defer auditLogger.file.Close()

	const goroutines = 10
	const entriesPerGoroutine = 5

	var wg sync.WaitGroup
	wg.Add(goroutines)

	startTime := time.Now().Unix()

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			userUUID := uuid.New()
			for j := 0; j < entriesPerGoroutine; j++ {
				action := "action-" + string(rune('a'+id))
				url := "https://test.com/" + string(rune('a'+id)) + "/" + string(rune('0'+j))
				err := auditLogger.Update(action, userUUID, url)
				assert.NoError(t, err)
			}
		}(i)
	}

	wg.Wait()

	// Validate total lines
	content, err := os.ReadFile(logPath)
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	assert.Len(t, lines, goroutines*entriesPerGoroutine)

	// Validate each line is valid JSON and has expected structure
	for _, line := range lines {
		var entry dto.AuditRequest
		err := json.Unmarshal([]byte(line), &entry)
		assert.NoError(t, err)
		assert.Greater(t, entry.TS, startTime-10) // reasonable timestamp
		assert.NotEmpty(t, entry.Action)
		assert.NotEmpty(t, entry.URL)
		assert.NotEqual(t, uuid.Nil, entry.UserID)
	}
}

func TestAuditFileLogger_Update_JSONMarshalError(t *testing.T) {
	t.Skip("JSON marshal error is not realistically triggerable with valid dto.AuditRequest")
}

func TestAuditFileLogger_Update_WriteError(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := tmpDir + "/audit_write_error.log"

	conf := cfg.FileAuditConfig{AuditFilePath: logPath}
	logger := zerolog.Nop()

	auditLogger, err := NewAuditFileLogger(conf, logger)
	require.NoError(t, err)

	// Close the file manually to simulate write failure
	auditLogger.file.Close()

	userUUID := uuid.New()
	err = auditLogger.Update("test", userUUID, "https://fail.com")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot write to file")
}
