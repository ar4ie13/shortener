package handler

import (
	"errors"
	"github.com/ar4ie13/shortener/internal/logger"
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
	s Service
	c Config
}

// NewHandler constructs Handler object
func NewHandler(s Service, c Config) *Handler {
	return &Handler{s, c}
}

// ListenAndServe starts web server with specified chi router
func (h Handler) ListenAndServe() error {
	zlog := logger.NewLogger(h.c.GetLogLevel())
	router := chi.NewRouter()

	router.Route("/", func(router chi.Router) {
		router.Post("/", h.postURL)
		router.Get("/{id}", h.getShortURLByID)
	})
	//zlog.Printf("Listening on %v\nURL Template: %v", h.c.GetLocalServerAddr(), h.c.GetShortURLTemplate())
	zlog.Info().Msgf("Listening on %v\nURL Template: %v\nLog Level: %v", h.c.GetLocalServerAddr(), h.c.GetShortURLTemplate(), h.c.GetLogLevel())

	if err := http.ListenAndServe(h.c.GetLocalServerAddr(), h.requestLogger(router)); err != nil {
		return err
	}

	return nil
}

// postURL handles POST requests from clients and receives URL from body to store it in the Repository via Service
func (h Handler) postURL(w http.ResponseWriter, r *http.Request) {
	zlog := logger.NewLogger(h.c.GetLogLevel())
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
		zlog.Error().Msgf("Failed to generate short url: %v", err)
		return
	}

	host := h.c.GetShortURLTemplate() + "/" + id
	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write([]byte(host)); err != nil {
		zlog.Error().Msgf("Failed to write response: %v", err)
	}
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

		zlog := logger.NewLogger(h.c.GetLogLevel())

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

		zlog.
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

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}
