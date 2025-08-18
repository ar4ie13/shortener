package handler

import (
	"io"
	"log"
	"net/http"
	"strings"
)

const (
	ServerAddr = "localhost:8080"
)

type Service interface {
	Get(id string) (string, error)
	GenerateShortURL(url string) (string, error)
}

type Handler struct {
	s Service
}

func NewHandler(s Service) *Handler {
	return &Handler{s}
}

func (h Handler) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", h.RequestRouter)
	if err := http.ListenAndServe(ServerAddr, mux); err != nil {
		return err
	}

	return nil
}

func (h Handler) RequestRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.PostURL(w, r)
	case http.MethodGet:
		h.GetShortURLByID(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (h Handler) PostURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	id, err := h.s.GenerateShortURL(string(body))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}
	host := "http://" + r.Host + "/" + id
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write([]byte(host)); err != nil {
		log.Println(err)
	}

}

func (h Handler) GetShortURLByID(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || r.URL.Path == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	id := strings.TrimLeft(r.URL.Path, "/")
	url, err := h.s.Get(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
