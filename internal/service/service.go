package service

import (
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrURLExist      = errors.New("URL already exist")
	ErrEmptyIDorURL  = errors.New("ID or URL cannot be empty")
	ErrShortURLExist = errors.New("ID already exist")

	ErrEmptyURL         = errors.New("URL template cannot be empty")
	ErrWrongHTTPScheme  = errors.New("URL template must use http or https scheme")
	ErrMustIncludeHost  = errors.New("URL template must include a host")
	ErrInvalidURLFormat = errors.New("invalid URL format")

	errEmptyID        = errors.New("short url cannot be empty")
	errShortURLLength = errors.New("short url length is too small")
)

const (
	randGenerateSymbols = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	shortURLLen         = 8
)

// Repository interface used to interact with repository package to store or retrieve values
type Repository interface {
	Get(shortURL string) (string, error)
	Save(shortURL string, url string) error
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
func (s Service) GetURL(shortURL string) (string, error) {
	if shortURL == "" {
		return "", errEmptyID
	}

	idURL, err := s.r.Get(shortURL)
	if err != nil {
		if errors.Is(err, ErrNotFound) || errors.Is(err, ErrEmptyIDorURL) {
			return "", fmt.Errorf("failed to get URL: %w", err)
		}
	}

	return idURL, nil
}

// GenerateShortURL generates shortURL for non-existent URL and stores it in the Repository
func (s Service) GenerateShortURL(urlLink string) (slug string, err error) {

	urlLink = strings.TrimRight(urlLink, "/")

	if urlLink == "" {
		return "", ErrEmptyURL
	}

	// Validate the URL format
	parsedURL, err := url.Parse(urlLink)
	if err != nil {
		return "", ErrInvalidURLFormat
	}

	// Ensure the scheme is http or https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", ErrWrongHTTPScheme
	}

	// Ensure the host is not empty
	if parsedURL.Host == "" {
		return "", ErrMustIncludeHost
	}

	for attempt := 1; attempt <= 3; attempt++ {
		slug, err = generateShortURL(shortURLLen)

		if err != nil {
			if attempt == 3 {
				return "", err
			}
			continue
		}

		err = s.r.Save(slug, urlLink)

		if err == nil {
			return slug, nil
		}

		if errors.Is(err, ErrURLExist) {
			return "", err
		}

		if attempt == 3 {
			if errors.Is(err, ErrShortURLExist) {
				return "", fmt.Errorf("failed to save URL to repository: %w", err)
			}
		}
	}
	return "", err
}

// generateShortURL is a sub-function for GenerateShortURL
func generateShortURL(length int) (string, error) {
	if length <= 0 {
		return "", errShortURLLength
	}

	shortURL := make([]byte, length)
	for i := range shortURL {
		shortURL[i] = randGenerateSymbols[rand.Intn(len(randGenerateSymbols))]
	}

	return string(shortURL), nil
}

var ()
