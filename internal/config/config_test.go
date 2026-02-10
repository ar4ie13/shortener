package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestShortURLTemplate_Set(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
		errorMsg    string
	}{
		{
			name:     "valid http URL",
			input:    "http://example.com",
			expected: "http://example.com",
		},
		{
			name:     "valid https URL",
			input:    "https://example.com",
			expected: "https://example.com",
		},
		{
			name:        "empty URL",
			input:       "",
			expectError: true,
			errorMsg:    "URL template cannot be empty",
		},
		{
			name:        "wrong scheme",
			input:       "ftp://example.com",
			expectError: true,
			errorMsg:    "URL template must use http or https scheme",
		},
		{
			name:        "no host",
			input:       "http://",
			expectError: true,
			errorMsg:    "URL template must include a host",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u ShortURLTemplate
			err := u.Set(tt.input)
			if tt.expectError {
				if err == nil {
					t.Error("Set() expected an error, but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Set() error expected to contain %q, got %q", tt.errorMsg, err.Error())
				}
				return
			}
			if err != nil {
				t.Errorf("Set() unexpected error: %v", err)
			}
			if string(u) != tt.expected {
				t.Errorf("Set() expected ShortURLTemplate %q, got %q", tt.expected, string(u))
			}
		})
	}
}

func TestConfig_GetLocalServerAddr(t *testing.T) {
	tests := []struct {
		name      string
		localAddr string
		expected  string
	}{
		{
			name:      "default address",
			localAddr: "localhost:8080",
			expected:  "localhost:8080",
		},
		{
			name:      "custom address",
			localAddr: "127.0.0.1:9090",
			expected:  "127.0.0.1:9090",
		},
		{
			name:      "empty address",
			localAddr: "",
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				LocalServerAddr: tt.localAddr,
			}
			got := cfg.GetLocalServerAddr()
			if got != tt.expected {
				t.Errorf("GetLocalServerAddr() expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestConfig_GetShortURLTemplate(t *testing.T) {
	tests := []struct {
		name             string
		shortURLTemplate ShortURLTemplate
		expected         string
	}{
		{
			name:             "valid http URL",
			shortURLTemplate: "http://example.com",
			expected:         "http://example.com",
		},
		{
			name:             "valid https URL",
			shortURLTemplate: "https://example.com",
			expected:         "https://example.com",
		},
		{
			name:             "empty URL",
			shortURLTemplate: "",
			expected:         "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				ShortURLTemplate: tt.shortURLTemplate,
			}
			got := cfg.GetShortURLTemplate()
			if got != tt.expected {
				t.Errorf("GetShortURLTemplate() expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestShortURLTemplate_String(t *testing.T) {
	tests := []struct {
		name     string
		input    ShortURLTemplate
		expected string
	}{
		{
			name:     "valid string",
			input:    ShortURLTemplate("http://example.com"),
			expected: "http://example.com",
		},
		{
			name:     "empty string",
			input:    ShortURLTemplate(""),
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.input.String(); got != tt.expected {
				t.Errorf("String() expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		level    zerolog.Level
		expected string
	}{
		{
			name:     "debug level",
			level:    zerolog.DebugLevel,
			expected: "debug",
		},
		{
			name:     "info level",
			level:    zerolog.InfoLevel,
			expected: "info",
		},
		{
			name:     "warn level",
			level:    zerolog.WarnLevel,
			expected: "warn",
		},
		{
			name:     "error level",
			level:    zerolog.ErrorLevel,
			expected: "error",
		},
		{
			name:     "fatal level",
			level:    zerolog.FatalLevel,
			expected: "fatal",
		},
		{
			name:     "panic level",
			level:    zerolog.PanicLevel,
			expected: "panic",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := LogLevel{Level: tt.level}
			if got := l.String(); got != tt.expected {
				t.Errorf("String() expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestLogLevel_Set(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    zerolog.Level
		expectError bool
	}{
		{
			name:     "debug level",
			input:    "debug",
			expected: zerolog.DebugLevel,
		},
		{
			name:     "info level",
			input:    "info",
			expected: zerolog.InfoLevel,
		},
		{
			name:     "warn level",
			input:    "warn",
			expected: zerolog.WarnLevel,
		},
		{
			name:     "error level",
			input:    "error",
			expected: zerolog.ErrorLevel,
		},
		{
			name:     "fatal level",
			input:    "fatal",
			expected: zerolog.FatalLevel,
		},
		{
			name:     "panic level",
			input:    "panic",
			expected: zerolog.PanicLevel,
		},
		{
			name:     "uppercase INFO",
			input:    "INFO",
			expected: zerolog.InfoLevel,
		},
		{
			name:     "mixed case Debug",
			input:    "Debug",
			expected: zerolog.DebugLevel,
		},
		{
			name:        "invalid level",
			input:       "invalid",
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var l LogLevel
			err := l.Set(tt.input)
			if tt.expectError {
				if err == nil {
					t.Error("Set() expected an error, but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("Set() unexpected error: %v", err)
			}
			if l.Level != tt.expected {
				t.Errorf("Set() expected level %v, got %v", tt.expected, l.Level)
			}
		})
	}
}

func TestConfig_GetLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    zerolog.Level
		expected zerolog.Level
	}{
		{
			name:     "debug level",
			level:    zerolog.DebugLevel,
			expected: zerolog.DebugLevel,
		},
		{
			name:     "info level",
			level:    zerolog.InfoLevel,
			expected: zerolog.InfoLevel,
		},
		{
			name:     "error level",
			level:    zerolog.ErrorLevel,
			expected: zerolog.ErrorLevel,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				LogLevel: LogLevel{Level: tt.level},
			}
			if got := cfg.GetLogLevel(); got != tt.expected {
				t.Errorf("GetLogLevel() expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestConfig_GetHTTPS(t *testing.T) {
	tests := []struct {
		name     string
		https    bool
		expected bool
	}{
		{
			name:     "https enabled",
			https:    true,
			expected: true,
		},
		{
			name:     "https disabled",
			https:    false,
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				HTTPS: tt.https,
			}
			if got := cfg.GetHTTPS(); got != tt.expected {
				t.Errorf("GetHTTPS() expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestConfig_GetTLSCertPath(t *testing.T) {
	tests := []struct {
		name     string
		certPath string
		expected string
	}{
		{
			name:     "valid path",
			certPath: "/path/to/cert.pem",
			expected: "/path/to/cert.pem",
		},
		{
			name:     "empty path",
			certPath: "",
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				TLSCertPath: tt.certPath,
			}
			if got := cfg.GetTLSCertPath(); got != tt.expected {
				t.Errorf("GetTLSCertPath() expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestConfig_GetTLSKeyPath(t *testing.T) {
	tests := []struct {
		name     string
		keyPath  string
		expected string
	}{
		{
			name:     "valid path",
			keyPath:  "/path/to/key.pem",
			expected: "/path/to/key.pem",
		},
		{
			name:     "empty path",
			keyPath:  "",
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				TLSKeyPath: tt.keyPath,
			}
			if got := cfg.GetTLSKeyPath(); got != tt.expected {
				t.Errorf("GetTLSKeyPath() expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestConfig_loadConfigFile(t *testing.T) {
	tests := []struct {
		name        string
		fileContent string
		setup       func(t *testing.T) string
		validate    func(t *testing.T, cfg *Config)
		expectError bool
	}{
		{
			name: "valid config file",
			fileContent: `{
				"local_server_addr": "0.0.0.0:9090",
				"short_url_template": "https://short.url",
				"https_enabled": true,
				"tls_cert_path": "/certs/cert.pem",
				"tls_key_path": "/certs/key.pem"
			}`,
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "config.json")
				err := os.WriteFile(path, []byte(`{
					"local_server_addr": "0.0.0.0:9090",
					"short_url_template": "https://short.url",
					"https_enabled": true,
					"tls_cert_path": "/certs/cert.pem",
					"tls_key_path": "/certs/key.pem"
				}`), 0644)
				if err != nil {
					t.Fatal(err)
				}
				return path
			},
			validate: func(t *testing.T, cfg *Config) {
				if cfg.LocalServerAddr != "0.0.0.0:9090" {
					t.Errorf("expected LocalServerAddr %q, got %q", "0.0.0.0:9090", cfg.LocalServerAddr)
				}
				if string(cfg.ShortURLTemplate) != "https://short.url" {
					t.Errorf("expected ShortURLTemplate %q, got %q", "https://short.url", cfg.ShortURLTemplate)
				}
				if !cfg.HTTPS {
					t.Error("expected HTTPS to be true")
				}
				if cfg.TLSCertPath != "/certs/cert.pem" {
					t.Errorf("expected TLSCertPath %q, got %q", "/certs/cert.pem", cfg.TLSCertPath)
				}
				if cfg.TLSKeyPath != "/certs/key.pem" {
					t.Errorf("expected TLSKeyPath %q, got %q", "/certs/key.pem", cfg.TLSKeyPath)
				}
			},
		},
		{
			name: "empty path returns nil",
			setup: func(t *testing.T) string {
				return ""
			},
			validate: func(t *testing.T, cfg *Config) {
				// Config should remain unchanged
			},
		},
		{
			name: "non-existent file returns nil",
			setup: func(t *testing.T) string {
				return "/non/existent/path/config.json"
			},
			validate: func(t *testing.T, cfg *Config) {
				// Config should remain unchanged
			},
		},
		{
			name: "invalid JSON returns error",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "invalid.json")
				err := os.WriteFile(path, []byte(`{invalid json}`), 0644)
				if err != nil {
					t.Fatal(err)
				}
				return path
			},
			expectError: true,
		},
		{
			name: "partial config file",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "partial.json")
				err := os.WriteFile(path, []byte(`{"local_server_addr": "127.0.0.1:3000"}`), 0644)
				if err != nil {
					t.Fatal(err)
				}
				return path
			},
			validate: func(t *testing.T, cfg *Config) {
				if cfg.LocalServerAddr != "127.0.0.1:3000" {
					t.Errorf("expected LocalServerAddr %q, got %q", "127.0.0.1:3000", cfg.LocalServerAddr)
				}
			},
		},
		{
			name: "config with auth settings",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "auth.json")
				err := os.WriteFile(path, []byte(`{
					"auth_config": {
						"secret_key": "test-secret-key",
						"token_expiration": 3600000000000
					}
				}`), 0644)
				if err != nil {
					t.Fatal(err)
				}
				return path
			},
			validate: func(t *testing.T, cfg *Config) {
				if cfg.AuthConf.SecretKey != "test-secret-key" {
					t.Errorf("expected SecretKey %q, got %q", "test-secret-key", cfg.AuthConf.SecretKey)
				}
				if cfg.AuthConf.TokenExpiration != time.Hour {
					t.Errorf("expected TokenExpiration %v, got %v", time.Hour, cfg.AuthConf.TokenExpiration)
				}
			},
		},
		{
			name: "config with database settings",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "db.json")
				err := os.WriteFile(path, []byte(`{
					"postgres_config": {
						"database_dsn": "postgres://user:pass@localhost:5432/db"
					}
				}`), 0644)
				if err != nil {
					t.Fatal(err)
				}
				return path
			},
			validate: func(t *testing.T, cfg *Config) {
				if cfg.PostgresDSN.DatabaseDSN != "postgres://user:pass@localhost:5432/db" {
					t.Errorf("expected DatabaseDSN %q, got %q", "postgres://user:pass@localhost:5432/db", cfg.PostgresDSN.DatabaseDSN)
				}
			},
		},
		{
			name: "config with file storage settings",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "file.json")
				err := os.WriteFile(path, []byte(`{
					"file_config": {
						"file_path": "/data/storage.json"
					}
				}`), 0644)
				if err != nil {
					t.Fatal(err)
				}
				return path
			},
			validate: func(t *testing.T, cfg *Config) {
				if cfg.FilePath.FilePath != "/data/storage.json" {
					t.Errorf("expected FilePath %q, got %q", "/data/storage.json", cfg.FilePath.FilePath)
				}
			},
		},
		{
			name: "config with audit settings",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "audit.json")
				err := os.WriteFile(path, []byte(`{
					"audit_config": {
						"file_audit_config": {
							"audit_file_path": "/var/log/audit.log"
						},
						"remote_audit_config": {
							"remote_server_url": "http://audit.server:8080"
						}
					}
				}`), 0644)
				if err != nil {
					t.Fatal(err)
				}
				return path
			},
			validate: func(t *testing.T, cfg *Config) {
				if cfg.AuditConf.FileConf.AuditFilePath != "/var/log/audit.log" {
					t.Errorf("expected AuditFilePath %q, got %q", "/var/log/audit.log", cfg.AuditConf.FileConf.AuditFilePath)
				}
				if cfg.AuditConf.RemoteConf.RemoteServerURL != "http://audit.server:8080" {
					t.Errorf("expected RemoteServerURL %q, got %q", "http://audit.server:8080", cfg.AuditConf.RemoteConf.RemoteServerURL)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{}
			path := tt.setup(t)
			err := cfg.loadConfigFile(path)

			if tt.expectError {
				if err == nil {
					t.Error("loadConfigFile() expected an error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("loadConfigFile() unexpected error: %v", err)
				return
			}

			if tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name     string
		expected struct {
			localAddr string
			shortURL  string
		}
	}{
		{
			name: "default config",
			expected: struct {
				localAddr string
				shortURL  string
			}{
				localAddr: "localhost:8080",
				shortURL:  "http://localhost:8080"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig()

			if cfg.LocalServerAddr != tt.expected.localAddr {
				t.Errorf("InitConfig() expected %q, got %q", tt.expected.localAddr, cfg.LocalServerAddr)

			}

			if string(cfg.ShortURLTemplate) != tt.expected.shortURL {
				t.Errorf("InitConfig() expected %q, got %q", tt.expected.shortURL, cfg.ShortURLTemplate)

			}
		})
	}
}
