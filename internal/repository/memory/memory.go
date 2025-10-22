package memory

import (
	"context"
	"fmt"

	"github.com/ar4ie13/shortener/internal/service"
	"github.com/google/uuid"
)

// SlugMemStore stores slug:URL
type SlugMemStore map[string]string

// URLMemStore stores URL:slug
type URLMemStore map[string]string

// UUIDMemStore stores uuid:slug
type UUIDMemStore map[string]string

// MemStorage is the main object for the package repository
type MemStorage struct {
	SlugMemStore
	URLMemStore
	UUIDMemStore
}

// NewMemStorage is a constructor for MemStorage object
func NewMemStorage() *MemStorage {
	return &MemStorage{
		SlugMemStore: make(map[string]string),
		URLMemStore:  make(map[string]string),
		UUIDMemStore: make(map[string]string),
	}
}

// Get method is used to get URL (link) from the repository map
func (repo *MemStorage) Get(_ context.Context, shortURL string) (string, error) {
	if v, ok := repo.SlugMemStore[shortURL]; ok {
		return v, nil
	}

	return "", service.ErrNotFound
}

// existsURL check if URL exist in the map
func (repo *MemStorage) existsURL(url string) bool {
	if _, ok := repo.URLMemStore[url]; ok {
		return true
	}

	return false
}

// existsShortURL check if URL exist in the map
func (repo *MemStorage) existsShortURL(shortURL string) bool {
	if _, ok := repo.SlugMemStore[shortURL]; ok {
		return true
	}

	return false
}

// Save saves shortURL, URL and UUID to the correlated maps
func (repo *MemStorage) Save(_ context.Context, shortURL string, url string) error {

	if shortURL == "" || url == "" {
		return service.ErrEmptyShortURLorURL
	}

	if repo.existsURL(url) {
		return fmt.Errorf("%w :%s", service.ErrURLExist, url)
	}

	if repo.existsShortURL(shortURL) {
		return fmt.Errorf("%w :%s", service.ErrShortURLExist, shortURL)
	}

	repo.SlugMemStore[shortURL] = url
	repo.URLMemStore[url] = shortURL
	repo.UUIDMemStore[shortURL] = uuid.New().String()

	return nil
}
