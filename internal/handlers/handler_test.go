package handlers

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This part will test two main handlers for POST and GET methods
// MockService mocks the Service interface for testing

type MockConfig struct {
	LocalServerAddr  string
	ShortURLTemplate string
	LogLevel         zerolog.Level
}

func (c *MockConfig) CheckPostgresConnection() error {
	return nil
}

func (c *MockConfig) GetLocalServerAddr() string {
	c.LocalServerAddr = "localhost:8080"
	return c.LocalServerAddr
}

func (c *MockConfig) GetShortURLTemplate() string {
	c.ShortURLTemplate = "http://localhost:8080"
	return c.ShortURLTemplate
}

func (c *MockConfig) GetLogLevel() zerolog.Level {
	c.LogLevel = zerolog.InfoLevel
	return c.LogLevel
}

type MockService struct {
	urlLib map[string]string
	err    error
}

type MockLogger struct {
	logger zerolog.Logger
}

func NewLogger(level zerolog.Level) *MockLogger {
	return &MockLogger{
		logger: zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).With().Timestamp().Logger().Level(level),
	}
}

// GetURL mocks the Service GetURL method
func (m *MockService) GetURL(id string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	url, exists := m.urlLib[id]
	if !exists {
		return "", errors.New("url not found")
	}
	return url, nil
}

// GenerateShortURL mocks the Service GenerateShortURL method
func (m *MockService) GenerateShortURL(url string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	for key := range m.urlLib {
		if m.urlLib[key] == url {
			return "", errors.New("url already exists")
		}

	}
	return "abc123", nil
}

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Prevents automatic redirects
		},
	}

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestGetShortURLByID(t *testing.T) {
	// Test cases
	tests := []struct {
		name           string
		path           string
		storageURLs    map[string]string
		storageErr     error
		expectedStatus int
		expectedHeader string
	}{
		{
			name:           "Empty path",
			path:           "",
			storageURLs:    map[string]string{},
			storageErr:     nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedHeader: "",
		},
		{
			name:           "Root path",
			path:           "/",
			storageURLs:    map[string]string{},
			storageErr:     nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedHeader: "",
		},
		{
			name:           "Non-existent ID",
			path:           "/nonexistent",
			storageURLs:    map[string]string{},
			storageErr:     errors.New("url not found"),
			expectedStatus: http.StatusBadRequest,
			expectedHeader: "",
		},
		{
			name:           "Valid ID",
			path:           "/abc123",
			storageURLs:    map[string]string{"abc123": "https://example.com"},
			storageErr:     nil,
			expectedStatus: http.StatusTemporaryRedirect,
			expectedHeader: "https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := &MockService{
				urlLib: tt.storageURLs,
				err:    tt.storageErr,
			}
			mockConfig := &MockConfig{
				LocalServerAddr: "localhost:8080",
				LogLevel:        zerolog.InfoLevel,
			}

			mockLogger := NewLogger(mockConfig.LogLevel)
			h := NewHandler(mockService, mockConfig, mockLogger.logger)

			router := chi.NewRouter()
			router.Route("/", func(router chi.Router) {
				router.Post("/", h.postURL)
				router.Get("/{id}", h.getShortURLByID)
			})
			ts := httptest.NewServer(router)
			defer ts.Close()
			resp, _ := testRequest(t, ts, "GET", tt.path)
			defer resp.Body.Close()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			assert.Equal(t, tt.expectedHeader, resp.Header.Get("Location"))

		})
	}
}

func TestPostURL(t *testing.T) {
	// Test cases
	tests := []struct {
		name           string
		path           string
		storageURLs    map[string]string
		storageErr     error
		body           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Empty path",
			path:           "",
			storageURLs:    map[string]string{"a123": "https://example.com"},
			storageErr:     nil,
			body:           "https://ya.com",
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost:8080/abc123",
		},
		{
			name:           "Positive POST request",
			path:           "/",
			storageURLs:    map[string]string{},
			storageErr:     nil,
			body:           "www.ya.ru",
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost:8080/abc123",
		},
		{
			name:           "Existent URL",
			path:           "/",
			storageURLs:    map[string]string{"abc123": "www.ya.ru"},
			storageErr:     errors.New("id already exists"),
			body:           "www.ya.ru",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := &MockService{
				urlLib: tt.storageURLs,
				err:    tt.storageErr,
			}
			mockConfig := &MockConfig{}

			mockLogger := NewLogger(mockConfig.LogLevel)
			handler := NewHandler(mockService, mockConfig, mockLogger.logger)

			// Create HTTP request
			req, err := http.NewRequest(http.MethodPost, tt.path, strings.NewReader(tt.body))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handlers
			handler.postURL(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			body, err := io.ReadAll(rr.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			// Check body for valid case
			if tt.expectedBody != string(body) {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, body)
			}

		})
	}
}
