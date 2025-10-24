package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/ar4ie13/shortener/internal/model"
	"github.com/ar4ie13/shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

// Service interface interacts with service package
type Service interface {
	GetURL(ctx context.Context, id string) (string, error)
	SaveURL(ctx context.Context, url string) (slug string, err error)
	SaveBatch(ctx context.Context, batch []model.URL) ([]model.URL, error)
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
	zlog    zerolog.Logger
}

// NewHandler constructs Handler object
func NewHandler(s Service, c Config, zlog zerolog.Logger) *Handler {
	return &Handler{s, c, zlog}
}

// ListenAndServe starts web server with specified chi router
func (h Handler) ListenAndServe() error {
	router := chi.NewRouter()

	// middleware for router
	router.Use(h.requestLogger)
	router.Use(gzipMiddleware)

	router.Route("/", func(router chi.Router) {
		router.Post("/", h.postURL)
		router.Get("/{id}", h.getShortURLByID)
		router.Get("/ping", h.checkPostgresConnection)
		router.Route("/api", func(router chi.Router) {
			router.Post("/shorten", h.postURLJSON)
			router.Post("/shorten/batch", h.postURLJSONBatch)
		})
	})
	h.zlog.Info().Msgf("listening on %v\nURL Template: %v\nLog Level: %v", h.cfg.GetLocalServerAddr(), h.cfg.GetShortURLTemplate(), h.cfg.GetLogLevel())

	if err := http.ListenAndServe(h.cfg.GetLocalServerAddr(), router); err != nil {
		return err
	}

	return nil
}

// errorProcessing process error and return the correlated status code
func (h Handler) errorProcessing(err error) (statusCode int) {
	switch {
	case errors.Is(err, service.ErrEmptyURL) || errors.Is(err, service.ErrInvalidURLFormat) ||
		errors.Is(err, service.ErrWrongHTTPScheme) || errors.Is(err, service.ErrMustIncludeHost):
		return http.StatusBadRequest
	case errors.Is(err, service.ErrURLExist):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// postURL handles POST requests from clients and receives URL from body to store it in the Repository via Service
func (h Handler) postURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	slug, err := h.service.SaveURL(r.Context(), string(body))
	if err != nil {
		statusCode := h.errorProcessing(err)
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

	slug, err := h.service.SaveURL(r.Context(), req.LongURL)
	if err != nil {
		statusCode := h.errorProcessing(err)
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

// getShortURLByID handles get requests and redirects to the URL by provided shortURL if it is found in Repository
func (h Handler) getShortURLByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	url, err := h.service.GetURL(r.Context(), id)
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

	serviceResp, err := h.service.SaveBatch(r.Context(), URLs)
	if err != nil {
		statusCode := h.errorProcessing(err)
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
