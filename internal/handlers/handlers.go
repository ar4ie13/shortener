package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/ar4ie13/shortener/internal/model"
	"github.com/ar4ie13/shortener/internal/myerrors"
	"github.com/go-chi/chi/v5"
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

// Config interface gets configuration flags from config package
type Config interface {
	GetLocalServerAddr() string
	GetShortURLTemplate() string
	GetLogLevel() zerolog.Level
	CheckPostgresConnection(ctx context.Context) error
}

// Handler is a main object for package handlers
type Handler struct {
	service Service
	cfg     Config
	auth    Auth
	zlog    zerolog.Logger
}

// NewHandler constructs Handler object
func NewHandler(s Service, c Config, a Auth, zlog zerolog.Logger) *Handler {
	return &Handler{s, c, a, zlog}
}

// ListenAndServe starts web server with specified chi router
func (h Handler) ListenAndServe() error {
	router := chi.NewRouter()

	// middleware for router
	router.Use(h.requestLogger)
	router.Use(h.authMiddleware)
	router.Use(h.gzipMiddleware)

	router.Route("/", func(router chi.Router) {
		router.Post("/", h.postURL)
		router.Get("/{id}", h.getURL)
		router.Get("/ping", h.checkPostgresConnection)
		router.Route("/api", func(router chi.Router) {
			router.Post("/shorten", h.postURLJSON)
			router.Post("/shorten/batch", h.postURLJSONBatch)
			router.Get("/user/urls", h.getUsersShortURL)
			router.Delete("/user/urls", h.deleteUsersShortURL)
		})
	})
	h.zlog.Info().Msgf("listening on %v\nURL Template: %v\nLog Level: %v", h.cfg.GetLocalServerAddr(), h.cfg.GetShortURLTemplate(), h.cfg.GetLogLevel())

	if err := http.ListenAndServe(h.cfg.GetLocalServerAddr(), router); err != nil {
		return err
	}

	return nil
}

// getUserUID
func (h Handler) getUserUUIDFromRequest(r *http.Request) (uuid.UUID, error) {
	userUUID, err := uuid.Parse(r.Context().Value(userUUIDKey).(string))
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

// postURL handles POST requests from clients and receives URL from body to store it in the Repository via Service
func (h Handler) postURL(w http.ResponseWriter, r *http.Request) {
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
}

func (h Handler) postURLJSON(w http.ResponseWriter, r *http.Request) {
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
}

// getURL handles get requests and redirects to the URL by provided shortURL if it is found in Repository
func (h Handler) getURL(w http.ResponseWriter, r *http.Request) {
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
}

// checkPostgresConnection used in /ping GET request
func (h Handler) checkPostgresConnection(w http.ResponseWriter, r *http.Request) {
	err := h.cfg.CheckPostgresConnection(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// postURLJSONBatch handles bath request in JSON
func (h Handler) postURLJSONBatch(w http.ResponseWriter, r *http.Request) {
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

	h.zlog.Debug().Msg("decoding batch request")
	var (
		req  []BatchRequest
		resp []BatchResponse
	)

	dec := json.NewDecoder(buf)
	if err = dec.Decode(&req); err != nil {
		h.zlog.Debug().Msgf("cannot decode bacth request JSON body: %v", h.zlog.Err(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.zlog.Debug().Msg("batch request decoded successfully")
	var URLs []model.URL
	for i := range req {
		URLs = append(URLs, model.URL{UUID: req[i].UUID, OriginalURL: req[i].LongURL})
	}

	serviceResp, err := h.service.SaveBatch(r.Context(), userUUID, URLs)
	if err != nil {
		statusCode := h.getStatusCode(err)
		switch statusCode {
		case http.StatusBadRequest:
			http.Error(w, err.Error(), statusCode)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.zlog.Error().Msgf("error handling batch: %v", err)
		return
	}
	for i := range serviceResp {
		resp = append(resp, BatchResponse{UUID: serviceResp[i].UUID, ShortURL: h.cfg.GetShortURLTemplate() + "/" + serviceResp[i].ShortURL})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(w)
	if err = enc.Encode(resp); err != nil {
		h.zlog.Debug().Msgf("error encoding batch response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getURL handles get requests and redirects to the URL by provided shortURL if it is found in Repository
func (h Handler) getUsersShortURL(w http.ResponseWriter, r *http.Request) {
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

// deleteUsersShortURL handles users short url deletion and places slugs into the channel in service
func (h Handler) deleteUsersShortURL(w http.ResponseWriter, r *http.Request) {
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
