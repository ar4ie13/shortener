package config

import (
	"strings"
	"testing"
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
