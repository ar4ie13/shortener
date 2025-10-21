package memory

import (
	"context"

	"github.com/ar4ie13/shortener/internal/service"
)

// slugMemStore stores slug:URL
type slugMemStore map[string]string

// urlMemStore stores URL:slug
type urlMemStore map[string]string

// MemStorage is the main object for the package repository
type MemStorage struct {
	slugMemStore
	urlMemStore
}

// NewMemStorage is a constructor for MemStorage object
func NewMemStorage() *MemStorage {
	return &MemStorage{
		slugMemStore: make(map[string]string),
		urlMemStore:  make(map[string]string),
	}
}

// Get method is used to get URL (link) from the repository map
func (repo *MemStorage) Get(_ context.Context, shortURL string) (string, error) {
	if v, ok := repo.slugMemStore[shortURL]; ok {
		return v, nil
	}

	return "", service.ErrNotFound
}

// existsURL check if URL exist in the map
func (repo *MemStorage) existsURL(url string) bool {
	if _, ok := repo.urlMemStore[url]; ok {
		return true
	}

	return false
}

// existsShortURL check if URL exist in the map
func (repo *MemStorage) existsShortURL(shortURL string) bool {
	if _, ok := repo.slugMemStore[shortURL]; ok {
		return true
	}

	return false
}

// Save saves the slug(shortURL):URL pair in the map
func (repo *MemStorage) Save(_ context.Context, shortURL string, url string) error {

	if shortURL == "" || url == "" {
		return service.ErrEmptyIDorURL
	}

	if repo.existsURL(url) {
		return service.ErrURLExist
	}

	if repo.existsShortURL(shortURL) {
		return service.ErrShortURLExist
	}

	repo.slugMemStore[shortURL] = url
	repo.urlMemStore[url] = shortURL

	return nil
}

// Load gets maps from file storage into memory storage maps
func (repo *MemStorage) Load(shortURLMap map[string]string) error {
	for k, v := range shortURLMap {
		repo.slugMemStore[k] = shortURLMap[k]
		repo.urlMemStore[v] = k
	}

	return nil
}
