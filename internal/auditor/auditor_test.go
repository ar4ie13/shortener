package auditor

import (
	"errors"
	"testing"

	"github.com/ar4ie13/shortener/internal/auditor/config"
	fileconf "github.com/ar4ie13/shortener/internal/auditor/file/config"
	remoteconf "github.com/ar4ie13/shortener/internal/auditor/remote/config"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockObserver is a mock implementation of the Observer interface for testing
type mockObserver struct {
	updateCalled bool
	action       string
	userUUID     uuid.UUID
	url          string
	err          error
}

func (m *mockObserver) Update(action string, userUUID uuid.UUID, url string) error {
	m.updateCalled = true
	m.action = action
	m.userUUID = userUUID
	m.url = url
	return m.err
}

func TestNewAuditor_Disabled(t *testing.T) {
	cfg := config.AuditConf{
		Enabled: false,
	}
	zlog := zerolog.Nop()

	auditor := NewAuditor(cfg, zlog)

	require.NotNil(t, auditor)
	assert.False(t, auditor.cfg.Enabled)
	assert.Empty(t, auditor.observers)
}

func TestNewAuditor_EnabledNoObservers(t *testing.T) {
	cfg := config.AuditConf{
		Enabled:    true,
		FileConf:   fileconf.FileAuditConfig{AuditFilePath: ""},
		RemoteConf: remoteconf.RemoteAuditConfig{RemoteServerURL: ""},
	}
	zlog := zerolog.Nop()

	auditor := NewAuditor(cfg, zlog)

	require.NotNil(t, auditor)
	assert.True(t, auditor.cfg.Enabled)
	assert.Empty(t, auditor.observers)
}

func TestNewAuditor_WithFileAudit(t *testing.T) {
	tmpFile := t.TempDir() + "/audit.log"
	cfg := config.AuditConf{
		Enabled:    true,
		FileConf:   fileconf.FileAuditConfig{AuditFilePath: tmpFile},
		RemoteConf: remoteconf.RemoteAuditConfig{RemoteServerURL: ""},
	}
	zlog := zerolog.Nop()

	auditor := NewAuditor(cfg, zlog)

	require.NotNil(t, auditor)
	assert.True(t, auditor.cfg.Enabled)
	assert.Len(t, auditor.observers, 1)
}

func TestNewAuditor_WithRemoteAudit(t *testing.T) {
	cfg := config.AuditConf{
		Enabled:    true,
		FileConf:   fileconf.FileAuditConfig{AuditFilePath: ""},
		RemoteConf: remoteconf.RemoteAuditConfig{RemoteServerURL: "http://localhost:8080/audit"},
	}
	zlog := zerolog.Nop()

	auditor := NewAuditor(cfg, zlog)

	require.NotNil(t, auditor)
	assert.True(t, auditor.cfg.Enabled)
	assert.Len(t, auditor.observers, 1)
}

func TestNewAuditor_WithBothAuditors(t *testing.T) {
	tmpFile := t.TempDir() + "/audit.log"
	cfg := config.AuditConf{
		Enabled:    true,
		FileConf:   fileconf.FileAuditConfig{AuditFilePath: tmpFile},
		RemoteConf: remoteconf.RemoteAuditConfig{RemoteServerURL: "http://localhost:8080/audit"},
	}
	zlog := zerolog.Nop()

	auditor := NewAuditor(cfg, zlog)

	require.NotNil(t, auditor)
	assert.True(t, auditor.cfg.Enabled)
	assert.Len(t, auditor.observers, 2)
}

func TestAttach(t *testing.T) {
	cfg := config.AuditConf{Enabled: true}
	zlog := zerolog.Nop()
	auditor := &Auditor{
		cfg:  cfg,
		zlog: zlog,
	}

	mock1 := &mockObserver{}
	mock2 := &mockObserver{}

	auditor.Attach(mock1)
	assert.Len(t, auditor.observers, 1)

	auditor.Attach(mock2)
	assert.Len(t, auditor.observers, 2)
}

func TestNotify_Enabled(t *testing.T) {
	cfg := config.AuditConf{Enabled: true}
	zlog := zerolog.Nop()
	auditor := &Auditor{
		cfg:  cfg,
		zlog: zlog,
	}

	mock := &mockObserver{}
	auditor.Attach(mock)

	userID := uuid.New()
	auditor.Notify("create", userID, "http://example.com")

	assert.True(t, mock.updateCalled)
	assert.Equal(t, "create", mock.action)
	assert.Equal(t, userID, mock.userUUID)
	assert.Equal(t, "http://example.com", mock.url)
}

func TestNotify_Disabled(t *testing.T) {
	cfg := config.AuditConf{Enabled: false}
	zlog := zerolog.Nop()
	auditor := &Auditor{
		cfg:  cfg,
		zlog: zlog,
	}

	mock := &mockObserver{}
	auditor.Attach(mock)

	userID := uuid.New()
	auditor.Notify("create", userID, "http://example.com")

	assert.False(t, mock.updateCalled)
}

func TestNotify_MultipleObservers(t *testing.T) {
	cfg := config.AuditConf{Enabled: true}
	zlog := zerolog.Nop()
	auditor := &Auditor{
		cfg:  cfg,
		zlog: zlog,
	}

	mock1 := &mockObserver{}
	mock2 := &mockObserver{}
	auditor.Attach(mock1)
	auditor.Attach(mock2)

	userID := uuid.New()
	auditor.Notify("delete", userID, "http://test.com")

	assert.True(t, mock1.updateCalled)
	assert.Equal(t, "delete", mock1.action)
	assert.Equal(t, userID, mock1.userUUID)
	assert.Equal(t, "http://test.com", mock1.url)

	assert.True(t, mock2.updateCalled)
	assert.Equal(t, "delete", mock2.action)
	assert.Equal(t, userID, mock2.userUUID)
	assert.Equal(t, "http://test.com", mock2.url)
}

func TestNotify_ObserverError(t *testing.T) {
	cfg := config.AuditConf{Enabled: true}
	zlog := zerolog.Nop()
	auditor := &Auditor{
		cfg:  cfg,
		zlog: zlog,
	}

	mock1 := &mockObserver{err: errors.New("observer error")}
	mock2 := &mockObserver{}
	auditor.Attach(mock1)
	auditor.Attach(mock2)

	userID := uuid.New()
	auditor.Notify("update", userID, "http://error.com")

	// Both observers should still be called even if one errors
	assert.True(t, mock1.updateCalled)
	assert.True(t, mock2.updateCalled)
}

func TestNotify_EmptyObservers(t *testing.T) {
	cfg := config.AuditConf{Enabled: true}
	zlog := zerolog.Nop()
	auditor := &Auditor{
		cfg:  cfg,
		zlog: zlog,
	}

	userID := uuid.New()
	// Should not panic with no observers
	auditor.Notify("create", userID, "http://example.com")
}
