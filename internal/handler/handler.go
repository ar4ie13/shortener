package handler

import (
	"errors"
	"github.com/ar4ie13/shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"log"
	"net/http"
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
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Route("/", func(router chi.Router) {
		router.Post("/", h.postURL)
		router.Get("/{id}", h.getShortURLByID)
	})
	log.Printf("Listening on %v\nURL Template: %v", h.c.GetLocalServerAddr(), h.c.GetShortURLTemplate())
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

	id, err := h.s.GenerateShortURL(string(body))
	if err != nil {
		if errors.Is(err, service.ErrEmptyURL) || errors.Is(err, service.ErrInvalidURLFormat) ||
			errors.Is(err, service.ErrWrongHTTPScheme) || errors.Is(err, service.ErrMustIncludeHost) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	host := h.c.GetShortURLTemplate() + "/" + id
	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write([]byte(host)); err != nil {
		log.Println(err)
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
