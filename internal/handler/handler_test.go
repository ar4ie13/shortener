package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// This part will test two main handlers for POST and GET methods
// MockService mocks the Service interface for testing
type MockService struct {
	urlLib map[string]string
	err    error
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
			expectedStatus: http.StatusBadRequest,
			expectedHeader: "",
		},
		{
			name:           "Root path",
			path:           "/",
			storageURLs:    map[string]string{},
			storageErr:     nil,
			expectedStatus: http.StatusBadRequest,
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
			handler := NewHandler(mockService)

			// Create HTTP request
			req, err := http.NewRequest(http.MethodGet, tt.path, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.getShortURLByID(rr, req)
			fmt.Println(tt.name, mockService.urlLib)
			fmt.Println(req.URL)
			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Check Location header for valid case
			if tt.expectedHeader != "" {
				location := rr.Header().Get("Location")
				if location != tt.expectedHeader {
					t.Errorf("Expected Location header %q, got %q", tt.expectedHeader, location)
				}
			} else if rr.Header().Get("Location") != "" {
				t.Errorf("Expected no Location header, got %q", rr.Header().Get("Location"))
			}
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
			storageURLs:    map[string]string{},
			storageErr:     nil,
			body:           "www.ya.ru",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
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
			expectedStatus: http.StatusBadRequest,
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
			handler := NewHandler(mockService)

			// Create HTTP request
			req, err := http.NewRequest(http.MethodPost, tt.path, strings.NewReader(tt.body))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
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
