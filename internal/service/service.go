package service

import (
	"errors"
	"math/rand"
)

var (
	ErrNotFound = errors.New("not found")
	ErrURLExist = errors.New("URL already exist")
)

const (
	randGenerateSymbols = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	shortURLLen         = 8
)

// Repository interface used to interact with repository package to store or retrieve values
type Repository interface {
	Get(id string) (string, error)
	Save(id string, url string) error
}

// Service is a main object of the package that implements Repository interface
type Service struct {
	r Repository
}

// NewService is a constructor for Service object
func NewService(r Repository) *Service {
	return &Service{r}
}

// GetURL method gets URL by provided id
func (s Service) GetURL(id string) (string, error) {
	id, err := s.r.Get(id)
	if err != nil {

		return "", err
	}

	return id, nil
}

// GenerateShortURL generates shortURL for non-existent URL and stores it in the Repository
func (s Service) GenerateShortURL(url string) (string, error) {
	id := generateShortURL(shortURLLen)
	if err := s.r.Save(id, url); err != nil {
		return id, err
	}

	return id, nil
}

// generateShortURL is a sub-function for GenerateShortURL
func generateShortURL(length int) string {
	shortURL := make([]byte, length)
	for i := range shortURL {
		shortURL[i] = randGenerateSymbols[rand.Intn(len(randGenerateSymbols))]
	}

	return string(shortURL)
}
