package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/ar4ie13/shortener/internal/model"
	"github.com/ar4ie13/shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"io"
	"net/http"
	"time"
)

// Service interface interacts with service package
type Service interface {
	GetURL(id string) (string, error)
	GenerateShortURL(url string) (slug string, err error)
}

// Config interface gets configuration flags from config package
type Config interface {
	GetLocalServerAddr() string
	GetShortURLTemplate() string
	GetLogLevel() zerolog.Level
}

// Handler is a main object for package handler
type Handler struct {
	s    Service
	c    Config
	zlog zerolog.Logger
}

// NewHandler constructs Handler object
func NewHandler(s Service, c Config, zlog zerolog.Logger) *Handler {
	return &Handler{s, c, zlog}
}

// ListenAndServe starts web server with specified chi router
func (h Handler) ListenAndServe() error {
	router := chi.NewRouter()

	router.Route("/", func(router chi.Router) {
		router.Post("/", h.postURL)
		router.Get("/{id}", h.getShortURLByID)
		router.Route("/api", func(router chi.Router) {
			router.Post("/shorten", h.postURLJSON)
		})
	})
	h.zlog.Info().Msgf("Listening on %v\nURL Template: %v\nLog Level: %v", h.c.GetLocalServerAddr(), h.c.GetShortURLTemplate(), h.c.GetLogLevel())

	if err := http.ListenAndServe(h.c.GetLocalServerAddr(), h.requestLogger(router)); err != nil {
		return err
	}

	return nil
}

// postURL handles POST requests from clients and receives URL from body to store it in the Repository via Service
func (h Handler) postURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := h.s.GenerateShortURL(string(body))
	if err != nil {
		if errors.Is(err, service.ErrEmptyURL) || errors.Is(err, service.ErrInvalidURLFormat) ||
			errors.Is(err, service.ErrWrongHTTPScheme) || errors.Is(err, service.ErrMustIncludeHost) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		h.zlog.Error().Msgf("Failed to generate short url: %v", err)
		return
	}

	host := h.c.GetShortURLTemplate() + "/" + id
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
	var req model.LongURL
	dec := json.NewDecoder(buf)
	if err := dec.Decode(&req); err != nil {
		h.zlog.Debug().Msgf("cannot decode request JSON body: %v", h.zlog.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id, err := h.s.GenerateShortURL(req.LongURL)
	if err != nil {
		if errors.Is(err, service.ErrEmptyURL) || errors.Is(err, service.ErrInvalidURLFormat) ||
			errors.Is(err, service.ErrWrongHTTPScheme) || errors.Is(err, service.ErrMustIncludeHost) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		h.zlog.Error().Msgf("Failed to generate short url: %v", err)
		return
	}

	resp := model.ShortURL{
		ShortURL: h.c.GetShortURLTemplate() + "/" + id,
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	if err = enc.Encode(resp); err != nil {
		h.zlog.Debug().Msgf("error encoding response: %v", err)
		return
	}
	h.zlog.Debug().Msg("sending HTTP 200 response")
}

// getShortURLByID handles get requests and redirects to the URL by provided shortURL if it is found in Repository
func (h Handler) getShortURLByID(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	url, err := h.s.GetURL(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// loggingResponseWriter structure for logging size and status code of responses
type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

// middleware logger using zerolog
func (h Handler) requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}

		next.ServeHTTP(&lw, r)

		h.zlog.
			Info().
			Str("method", r.Method).
			Str("url", r.RequestURI).
			Str("user_agent", r.UserAgent()).
			Int("size", responseData.size).
			Dur("elapsed_ms", time.Since(start)).
			Int("status", responseData.status).
			Msg("incoming request")
	})
}

// Write is redeclared method for embedding size into response logging
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader is redeclared method for embedding status code into response logging
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
