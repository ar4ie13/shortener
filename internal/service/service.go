package service

import (
	"errors"
	"fmt"
	"math/rand"
	"net/url"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrURLExist        = errors.New("URL already exist")
	ErrInvalidIDorURL  = errors.New("invalid ID or URL")
	ErrEmptyURL        = errors.New("URL template cannot be empty")
	ErrWrongHTTPScheme = errors.New("URL template must use http or https scheme")
	ErrMustIncludeHost = errors.New("URL template must include a host")
	ErrEmptyID         = errors.New("short url cannot be empty")
	ErrShortURLLength  = errors.New("short url length is too small")
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
	if id == "" {
		return "", ErrEmptyID
	}
	id, err := s.r.Get(id)
	if err != nil {

		return "", err
	}

	return id, nil
}

// GenerateShortURL generates shortURL for non-existent URL and stores it in the Repository
func (s Service) GenerateShortURL(urlLink string) (string, error) {
	if urlLink == "" {
		return "", ErrEmptyURL
	}

	// Validate the URL format
	parsedURL, err := url.Parse(urlLink)
	if err != nil {
		return "", fmt.Errorf("invalid URL format: %v", err)
	}

	// Ensure the scheme is http or https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", ErrWrongHTTPScheme
	}

	// Ensure the host is not empty
	if parsedURL.Host == "" {
		return "", ErrMustIncludeHost
	}
	id, err := generateShortURL(shortURLLen)
	if err != nil {
		return "", err
	}
	if err = s.r.Save(id, urlLink); err != nil {
		return id, err
	}

	return id, nil
}

// generateShortURL is a sub-function for GenerateShortURL
func generateShortURL(length int) (string, error) {
	if length <= 0 {
		return "", ErrShortURLLength
	}
	shortURL := make([]byte, length)
	for i := range shortURL {
		shortURL[i] = randGenerateSymbols[rand.Intn(len(randGenerateSymbols))]
	}

	return string(shortURL), nil
}
