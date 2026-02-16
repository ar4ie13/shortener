package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/ar4ie13/shortener/internal/handlers"
	"github.com/ar4ie13/shortener/internal/model"
	"github.com/ar4ie13/shortener/internal/myerrors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// --- Mock dependencies ---
type mockService struct{}

func (m mockService) GetURL(_ context.Context, _ uuid.UUID, id string) (string, error) {
	if id == "abc123" {
		return "https://example.com", nil
	}
	return "", myerrors.ErrNotFound
}

func (m mockService) SaveURL(_ context.Context, _ uuid.UUID, url string) (slug string, err error) {
	if url == "https://example.com" {
		return "abc123", nil
	}
	if url == "https://duplicate.com" {
		return "dup456", myerrors.ErrURLExist
	}
	return "new789", nil
}

func (m mockService) SaveBatch(_ context.Context, _ uuid.UUID, batch []model.URL) ([]model.URL, error) {
	res := make([]model.URL, len(batch))
	for i, u := range batch {
		res[i] = model.URL{
			UUID:        u.UUID,
			OriginalURL: u.OriginalURL,
			ShortURL:    "slug" + fmt.Sprint(i+1),
		}
	}
	return res, nil
}

func (m mockService) GetUserShortURLs(_ context.Context, _ uuid.UUID) (map[string]string, error) {
	return map[string]string{
		"abc123": "https://example.com",
		"new789": "https://test.com",
	}, nil
}

func (m mockService) SendShortURLForDelete(_ context.Context, _ uuid.UUID, _ []string) {}

func (m mockService) GetStats(_ context.Context) (int, int, error) {
	return 10, 5, nil
}

type mockAuth struct{}

func (m mockAuth) GenerateUserUUID() uuid.UUID {
	return uuid.MustParse("12345678-1234-1234-1234-123456789abc")
}
func (m mockAuth) BuildJWTString(_ uuid.UUID) (string, error) { return "fake-jwt", nil }
func (m mockAuth) ValidateUserUUID(_ string) (uuid.UUID, error) {
	return uuid.MustParse("12345678-1234-1234-1234-123456789abc"), nil
}

type mockAuditor struct{}

func (m mockAuditor) Notify(_ string, _ uuid.UUID, _ string) {}

type mockConfig struct{}

func (m mockConfig) GetLocalServerAddr() string {
	return ":8080"
}
func (m mockConfig) GetShortURLTemplate() string {
	return "http://localhost:8080"
}
func (m mockConfig) GetLogLevel() zerolog.Level {
	return zerolog.InfoLevel
}
func (m mockConfig) CheckPostgresConnection(_ context.Context) error {
	return nil
}
func (m mockConfig) GetHTTPS() bool {
	return false
}
func (m mockConfig) GetTLSCertPath() string {
	return ""
}
func (m mockConfig) GetTLSKeyPath() string {
	return ""
}
func (m mockConfig) GetTrustedSubnet() string {
	return "192.168.31.0/24"
}

// --- Helper to create test server with a known user context ---

func newTestServer() *httptest.Server {
	service := mockService{}
	cfg := mockConfig{}
	auth := mockAuth{}
	auditor := mockAuditor{}
	logger := zerolog.Nop()

	handler := handlers.NewHandler(service, cfg, auth, auditor, logger)

	router := chi.NewRouter()
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate auth middleware: inject user UUID as string
			ctx := context.WithValue(r.Context(), handlers.UserUUIDKey, "12345678-1234-1234-1234-123456789abc")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	router.Post("/", handler.PostURL)
	router.Get("/{id}", handler.GetURL)
	router.Post("/api/shorten", handler.PostURLJSON)
	router.Get("/api/user/urls", handler.GetUsersShortURL)

	return httptest.NewServer(router)
}

// Example_postURL demonstrates how to post URL and receive the response containing slug for the provided URL.
func Example_postURL() {
	srv := newTestServer()
	defer srv.Close()

	// Send plain-text URL
	resp, err := http.Post(srv.URL, "text/plain", strings.NewReader("https://example.com"))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body := make([]byte, 128)
	n, _ := resp.Body.Read(body)
	shortURL := string(body[:n])

	fmt.Println("Status:", resp.StatusCode)
	fmt.Println("Short URL:", shortURL)

	// Output:
	// Status: 201
	// Short URL: http://localhost:8080/abc123
}

// Example_postURLJSON shows how to post url in json.
func Example_postURLJSON() {
	srv := newTestServer()
	defer srv.Close()

	reqBody, _ := json.Marshal(map[string]string{"url": "https://example.com"})
	resp, err := http.Post(srv.URL+"/api/shorten", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var result struct {
		ShortURL string `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	fmt.Println("Status:", resp.StatusCode)
	fmt.Println("Short URL:", result.ShortURL)

	// Output:
	// Status: 201
	// Short URL: http://localhost:8080/abc123
}

// Example_postURLJSON_duplicate shows that if the duplicated URL is attempting to ne stored - status
// Conflict will be returned.
func Example_postURLJSON_duplicate() {
	srv := newTestServer()
	defer srv.Close()

	reqBody, _ := json.Marshal(map[string]string{"url": "https://duplicate.com"})
	resp, err := http.Post(srv.URL+"/api/shorten", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var result struct {
		ShortURL string `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	fmt.Println("Status:", resp.StatusCode)
	fmt.Println("Existing Short URL:", result.ShortURL)

	// Output:
	// Status: 409
	// Existing Short URL: http://localhost:8080/dup456
}

// Example_getURL returns saved URL by the provided slug
func Example_getURL() {
	srv := newTestServer()
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/abc123")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("Status:", resp.StatusCode)
	fmt.Println("Get URL:", resp.Request.URL.String())

	// Output:
	// Status: 200
	// Get URL: https://example.com
}

// Example_getUsersShortURL demonstrates that all saved URLs will be returned to the owner
func Example_getUsersShortURL() {
	srv := newTestServer()
	defer srv.Close()

	req, _ := http.NewRequest("GET", srv.URL+"/api/user/urls", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var urls []struct {
		ShortURL string `json:"short_url"`
		LongURL  string `json:"original_url"`
	}
	json.NewDecoder(resp.Body).Decode(&urls)

	fmt.Println("Status:", resp.StatusCode)
	fmt.Println("Found", len(urls), "URLs")

	// Output:
	// Status: 200
	// Found 2 URLs
}
