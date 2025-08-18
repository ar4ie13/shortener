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
		h.GetShortURLById(w, r)
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
	//TODO: remove when autotests will be completed successfully
	//responseText := fmt.Sprintf("%s %d Created\nContent-Type: %s\nContent-Length: %d\n\n%s", r.Proto, http.StatusCreated, r.Header.Get("Content-Type"), len(host), host)
	if _, err := w.Write([]byte(host)); err != nil {
		log.Println(err)
	}

}

func (h Handler) GetShortURLById(w http.ResponseWriter, r *http.Request) {
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
	/*TODO: remove when autotests will be completed successfully
		w.WriteHeader(http.StatusTemporaryRedirect)
	responseText := fmt.Sprintf("%s %d Temporary Redirect\n\nLocation: %s", r.Proto, http.StatusTemporaryRedirect, url)
	if _, err := w.Write([]byte("Location: " + url)); err != nil {
		log.Println(err)
	}
	*/
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
