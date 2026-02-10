// Package handlers contains API and middleware used by web server.
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ar4ie13/shortener/internal/model"
	"github.com/ar4ie13/shortener/internal/myerrors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// errorStatusMap used for fast error check in get
var errorStatusMap = map[error]int{
	myerrors.ErrEmptyURL:          http.StatusBadRequest,
	myerrors.ErrInvalidURLFormat:  http.StatusBadRequest,
	myerrors.ErrWrongHTTPScheme:   http.StatusBadRequest,
	myerrors.ErrMustIncludeHost:   http.StatusBadRequest,
	myerrors.ErrURLExist:          http.StatusConflict,
	myerrors.ErrNotFound:          http.StatusNoContent,
	myerrors.ErrShortURLIsDeleted: http.StatusGone,
}

// Service interface interacts with service package
type Service interface {
	GetURL(ctx context.Context, userUUID uuid.UUID, id string) (string, error)
	SaveURL(ctx context.Context, userUUID uuid.UUID, url string) (slug string, err error)
	SaveBatch(ctx context.Context, userUUID uuid.UUID, batch []model.URL) ([]model.URL, error)
	GetUserShortURLs(ctx context.Context, userUUID uuid.UUID) (map[string]string, error)
	SendShortURLForDelete(ctx context.Context, userUUID uuid.UUID, shortURLs []string)
}

// Auth used for authentication
type Auth interface {
	GenerateUserUUID() uuid.UUID
	BuildJWTString(userUUID uuid.UUID) (string, error)
	ValidateUserUUID(tokenString string) (uuid.UUID, error)
}

type Auditor interface {
	Notify(action string, userUUID uuid.UUID, url string)
}

// Config interface gets configuration flags from config package
type Config interface {
	GetLocalServerAddr() string
	GetShortURLTemplate() string
	GetLogLevel() zerolog.Level
	CheckPostgresConnection(ctx context.Context) error
	GetHTTPS() bool
	GetTLSCertPath() string
	GetTLSKeyPath() string
}

// Handler is a main object for package handlers
type Handler struct {
	service  Service
	cfg      Config
	auth     Auth
	observer Auditor
	zlog     zerolog.Logger
}

// NewHandler constructs Handler object
func NewHandler(s Service, c Config, a Auth, o Auditor, zlog zerolog.Logger) *Handler {
	return &Handler{s, c, a, o, zlog}
}

func (h Handler) RegisterRoutes() *chi.Mux {
	router := chi.NewRouter()

	// adding pprof to /debug
	router.Group(func(router chi.Router) {
		router.Mount("/debug", middleware.Profiler())
	})

	// main routes
	router.Group(func(router chi.Router) {

		// middleware for router
		router.Use(h.requestLogger)
		router.Use(h.authMiddleware)
		router.Use(h.gzipMiddleware)

		router.Route("/", func(router chi.Router) {
			router.Post("/", h.PostURL)
			router.Get("/{id}", h.GetURL)
			router.Get("/ping", h.CheckPostgresConnection)
			router.Route("/api", func(router chi.Router) {
				router.Post("/shorten", h.PostURLJSON)
				router.Post("/shorten/batch", h.PostURLJSONBatch)
				router.Get("/user/urls", h.GetUsersShortURL)
				router.Delete("/user/urls", h.DeleteUsersShortURL)
			})
		})
	})
	return router
}

// ListenAndServe starts web server with specified chi router
func (h Handler) ListenAndServe() error {

	srv := &http.Server{
		Addr:    h.cfg.GetLocalServerAddr(),
		Handler: h.RegisterRoutes(),
	}

	// Graceful shutdown
	idleConnsClosed := make(chan struct{})

	sigint := make(chan os.Signal, 1)

	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sigint
		if err := srv.Shutdown(context.Background()); err != nil {
			h.zlog.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	h.zlog.Info().Msgf("listening on %v\nURL Template: %v\nLog Level: %v", h.cfg.GetLocalServerAddr(), h.cfg.GetShortURLTemplate(), h.cfg.GetLogLevel())
	switch h.cfg.GetHTTPS() {
	case true:
		if err := srv.ListenAndServeTLS(h.cfg.GetTLSCertPath(), h.cfg.GetTLSKeyPath()); !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	default:
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			return err
		}

	}
	<-idleConnsClosed

	h.zlog.Info().Msgf("shortener server shutdown gracefully")
	return nil
}

// getUserUID
func (h Handler) getUserUUIDFromRequest(r *http.Request) (uuid.UUID, error) {
	userUUID, err := uuid.Parse(r.Context().Value(UserUUIDKey).(string))
	if err != nil {
		h.zlog.Debug().Msgf("cannot parse user UUID: %v", err)
		return uuid.Nil, err
	}

	return userUUID, nil
}

// getStatusCode process error and return the correlated status code
func (h Handler) getStatusCode(err error) int {
	// fast error check
	if status, exists := errorStatusMap[err]; exists {
		return status
	}

	// For wrapped errors
	for errType, status := range errorStatusMap {
		if errors.Is(err, errType) {
			return status
		}
	}

	return http.StatusInternalServerError
}

// PostURL handles POST requests from clients and receives URL from body to store it in the Repository via Service
func (h Handler) PostURL(w http.ResponseWriter, r *http.Request) {
	userUUID, err := h.getUserUUIDFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	slug, err := h.service.SaveURL(r.Context(), userUUID, string(body))
	if err != nil {
		statusCode := h.getStatusCode(err)
		switch statusCode {
		case http.StatusConflict:
			w.WriteHeader(statusCode)
			_, err = w.Write([]byte(h.cfg.GetShortURLTemplate() + "/" + slug))
			if err != nil {
				h.zlog.Error().Err(err).Msg("failed to write response body")
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		case http.StatusBadRequest:
			http.Error(w, err.Error(), statusCode)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.zlog.Error().Msgf("Failed to generate short url: %v", err)
		return
	}

	host := h.cfg.GetShortURLTemplate() + "/" + slug
	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write([]byte(host)); err != nil {
		h.zlog.Error().Msgf("Failed to write response: %v", err)
	}

	h.observer.Notify("shorten", userUUID, string(body))
}

func (h Handler) PostURLJSON(w http.ResponseWriter, r *http.Request) {
	userUUID, err := h.getUserUUIDFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
	}

	buf := new(bytes.Buffer)
	n, err := buf.ReadFrom(r.Body)
	if err != nil || n == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.zlog.Debug().Msg("decoding request")
	var req LongURLReq
	dec := json.NewDecoder(buf)
	if err = dec.Decode(&req); err != nil {
		h.zlog.Debug().Msgf("cannot decode request JSON body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.zlog.Debug().Msg("request decoded successfully")

	slug, err := h.service.SaveURL(r.Context(), userUUID, req.LongURL)
	if err != nil {
		statusCode := h.getStatusCode(err)
		switch statusCode {
		case http.StatusConflict:
			resp := ShortURLResp{
				ShortURL: h.cfg.GetShortURLTemplate() + "/" + slug,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			enc := json.NewEncoder(w)
			if err = enc.Encode(resp); err != nil {
				h.zlog.Debug().Msgf("error encoding response: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		case http.StatusBadRequest:
			http.Error(w, err.Error(), statusCode)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.zlog.Error().Msgf("Failed to generate short url: %v", err)
		return
	}

	resp := ShortURLResp{
		ShortURL: h.cfg.GetShortURLTemplate() + "/" + slug,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(w)
	if err = enc.Encode(resp); err != nil {
		h.zlog.Debug().Msgf("error encoding response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.observer.Notify("shorten", userUUID, req.LongURL)
}

// GetURL handles get requests and redirects to the URL by provided shortURL if it is found in Repository
func (h Handler) GetURL(w http.ResponseWriter, r *http.Request) {
	userUUID, err := h.getUserUUIDFromRequest(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	id := chi.URLParam(r, "id")
	url, err := h.service.GetURL(r.Context(), userUUID, id)
	if err != nil {
		statusCode := h.getStatusCode(err)
		http.Error(w, err.Error(), statusCode)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)

	h.observer.Notify("follow", userUUID, url)
}

// CheckPostgresConnection used in /ping GET request
func (h Handler) CheckPostgresConnection(w http.ResponseWriter, r *http.Request) {
	err := h.cfg.CheckPostgresConnection(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// PostURLJSONBatch handles batch request in JSON
func (h Handler) PostURLJSONBatch(w http.ResponseWriter, r *http.Request) {
	userUUID, err := h.getUserUUIDFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return // ✅ Added return
	}

	// Optional: relaxed Content-Type check
	if contentType := r.Header.Get("Content-Type"); contentType != "" {
		if !strings.HasPrefix(contentType, "application/json") {
			http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
			return
		}
	}

	h.zlog.Debug().Msg("decoding batch request")

	var req []BatchRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.zlog.Debug().Err(err).Msg("failed to decode batch request JSON")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(req) == 0 {
		http.Error(w, "Empty batch", http.StatusBadRequest)
		return
	}

	// Build input for service (avoid extra copy if possible)
	URLs := make([]model.URL, len(req))
	for i := range req {
		URLs[i] = model.URL{
			UUID:        req[i].UUID,
			OriginalURL: req[i].LongURL,
		}
	}

	serviceResp, err := h.service.SaveBatch(r.Context(), userUUID, URLs)
	if err != nil {
		statusCode := h.getStatusCode(err)
		http.Error(w, err.Error(), statusCode)
		if statusCode == http.StatusInternalServerError {
			h.zlog.Error().Err(err).Msg("error in batch save")
		}
		return
	}

	// Pre-allocate response
	resp := make([]BatchResponse, len(serviceResp))
	baseURL := h.cfg.GetShortURLTemplate()
	for i := range serviceResp {
		resp[i] = BatchResponse{
			UUID:     serviceResp[i].UUID,
			ShortURL: baseURL + "/" + serviceResp[i].ShortURL,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err = json.NewEncoder(w).Encode(resp); err != nil {
		h.zlog.Warn().Err(err).Msg("failed to encode batch response")
		// Note: Can't send error to client after headers written
	}
}

// GetUsersShortURL handles get requests and redirects to the URL by provided shortURL if it is found in Repository
func (h Handler) GetUsersShortURL(w http.ResponseWriter, r *http.Request) {
	userUUID, err := h.getUserUUIDFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	userSlugs, err := h.service.GetUserShortURLs(r.Context(), userUUID)
	if err != nil {
		statusCode := h.getStatusCode(err)
		switch statusCode {
		case http.StatusBadRequest:
			http.Error(w, err.Error(), statusCode)
			return

		case http.StatusNoContent:
			http.Error(w, err.Error(), statusCode)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.zlog.Error().Msgf("error handling request: %v", err)
		return
	}

	var resp []UserShortURLs

	for k, v := range userSlugs {
		resp = append(resp, UserShortURLs{ShortURL: h.cfg.GetShortURLTemplate() + "/" + k, LongURL: v})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	if err = enc.Encode(resp); err != nil {
		h.zlog.Debug().Msgf("error encoding response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// DeleteUsersShortURL handles users short url deletion and places slugs into the channel in service
func (h Handler) DeleteUsersShortURL(w http.ResponseWriter, r *http.Request) {
	userUUID, err := h.getUserUUIDFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		h.zlog.Debug().Msgf("error getting user UUID: %v", err)
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		h.zlog.Debug().Msgf("Error reading body: %v", err)
		return
	}
	defer r.Body.Close() // Ensure the body is closed

	// Declare a slice to unmarshal the JSON into
	var shortURLs []string

	// Unmarshal the JSON bytes into the Go slice
	err = json.Unmarshal(bodyBytes, &shortURLs)
	if err != nil {
		http.Error(w, "Error unmarshalling JSON", http.StatusBadRequest)
		h.zlog.Debug().Msgf("Error unmarshalling JSON: %v", err)
		return
	}

	h.service.SendShortURLForDelete(r.Context(), userUUID, shortURLs)

	w.WriteHeader(http.StatusAccepted)
}
