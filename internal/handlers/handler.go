package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/ar4ie13/shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

// Service interface interacts with service package
type Service interface {
	GetURL(ctx context.Context, id string) (string, error)
	GenerateShortURL(ctx context.Context, url string) (slug string, err error)
}

// Config interface gets configuration flags from config package
type Config interface {
	GetLocalServerAddr() string
	GetShortURLTemplate() string
	GetLogLevel() zerolog.Level
	CheckPostgresConnection() error
}

// Handler is a main object for package handlers
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
	h.zlog.Info().Msgf("listening on %v\nURL Template: %v\nLog Level: %v", h.c.GetLocalServerAddr(), h.c.GetShortURLTemplate(), h.c.GetLogLevel())

	if err := http.ListenAndServe(h.c.GetLocalServerAddr(), router); err != nil {
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

	id, err := h.s.GenerateShortURL(r.Context(), string(body))
	if err != nil {
		if errors.Is(err, service.ErrEmptyURL) || errors.Is(err, service.ErrInvalidURLFormat) ||
			errors.Is(err, service.ErrWrongHTTPScheme) || errors.Is(err, service.ErrMustIncludeHost) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	var req LongURLReq
	dec := json.NewDecoder(buf)
	if err = dec.Decode(&req); err != nil {
		h.zlog.Debug().Msgf("cannot decode request JSON body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := h.s.GenerateShortURL(r.Context(), req.LongURL)
	if err != nil {
		if errors.Is(err, service.ErrEmptyURL) || errors.Is(err, service.ErrInvalidURLFormat) ||
			errors.Is(err, service.ErrWrongHTTPScheme) || errors.Is(err, service.ErrMustIncludeHost) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.zlog.Error().Msgf("Failed to generate short url: %v", err)
		return
	}

	resp := ShortURLResp{
		ShortURL: h.c.GetShortURLTemplate() + "/" + id,
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
	url, err := h.s.GetURL(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h Handler) checkPostgresConnection(w http.ResponseWriter, _ *http.Request) {
	err := h.c.CheckPostgresConnection()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

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
		req  []BathRequest
		resp []BathResponse
	)

	dec := json.NewDecoder(buf)
	if err = dec.Decode(&req); err != nil {
		h.zlog.Debug().Msgf("cannot decode bacth request JSON body: %v", h.zlog.Err(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for i := range req {
		var res BathResponse
		fmt.Println(req[i])
		id, err := h.s.GenerateShortURL(r.Context(), req[i].LongURL)
		if err != nil {
			if errors.Is(err, service.ErrEmptyURL) || errors.Is(err, service.ErrInvalidURLFormat) ||
				errors.Is(err, service.ErrWrongHTTPScheme) || errors.Is(err, service.ErrMustIncludeHost) {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			h.zlog.Error().Msgf("Failed to generate short url: %v", err)
			return
		}
		res = BathResponse{
			UUID:     req[i].UUID,
			ShortURL: h.c.GetShortURLTemplate() + "/" + id,
		}
		resp = append(resp, res)
	}
	fmt.Printf("req: %v\n", req)
	fmt.Printf("resp: %v\n", resp)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(w)
	if err = enc.Encode(resp); err != nil {
		h.zlog.Debug().Msgf("error encoding batch response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
