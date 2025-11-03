package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/ar4ie13/shortener/internal/model"
	"github.com/ar4ie13/shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Service interface interacts with service package
type Service interface {
	GetURL(ctx context.Context, userUUID uuid.UUID, id string) (string, error)
	SaveURL(ctx context.Context, userUUID uuid.UUID, url string) (slug string, err error)
	SaveBatch(ctx context.Context, userUUID uuid.UUID, batch []model.URL) ([]model.URL, error)
	GetUserShortURLs(ctx context.Context, userUUID uuid.UUID) (map[string]string, error)
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
func (h Handler) getStatusCode(err error) (statusCode int) {
	switch {
	case errors.Is(err, service.ErrEmptyURL) || errors.Is(err, service.ErrInvalidURLFormat) ||
		errors.Is(err, service.ErrWrongHTTPScheme) || errors.Is(err, service.ErrMustIncludeHost):
		return http.StatusBadRequest
	case errors.Is(err, service.ErrURLExist):
		return http.StatusConflict
	case errors.Is(err, service.ErrNotFound):
		return http.StatusNoContent
	default:
		return http.StatusInternalServerError
	}
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
		http.Error(w, err.Error(), http.StatusBadRequest)
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
	fmt.Println(userUUID)

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
