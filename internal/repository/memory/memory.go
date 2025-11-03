package memory

import (
	"context"
	"fmt"

	"github.com/ar4ie13/shortener/internal/model"
	"github.com/ar4ie13/shortener/internal/service"
	"github.com/google/uuid"
)

// SlugMemStore stores slug:URL
type SlugMemStore map[string]string

// URLMemStore stores URL:slug
type URLMemStore map[string]string

// UserUUIDURLMemStore stores UserUUID:URL:slug
type UserUUIDURLMemStore map[uuid.UUID]URLMemStore

// UUIDMemStore stores UUID:slug
type UUIDMemStore map[uuid.UUID]string

// UserUUIDSlugMemStore stores UserUUID:shortURL:URL
type UserUUIDSlugMemStore map[uuid.UUID]SlugMemStore

// MemStorage is the main object for the package repository
type MemStorage struct {
	SlugMemStore
	UserUUIDURLMemStore
	UUIDMemStore
	UserUUIDSlugMemStore
}

// NewMemStorage is a constructor for MemStorage object
func NewMemStorage() *MemStorage {
	return &MemStorage{
		SlugMemStore:         make(map[string]string),
		UserUUIDURLMemStore:  make(map[uuid.UUID]URLMemStore),
		UUIDMemStore:         make(map[uuid.UUID]string),
		UserUUIDSlugMemStore: make(map[uuid.UUID]SlugMemStore),
	}
}

// GetURL method is used to get URL (link) from the repository map
func (repo *MemStorage) GetURL(_ context.Context, userUUID uuid.UUID, shortURL string) (string, error) {
	if v, ok := repo.UserUUIDSlugMemStore[userUUID][shortURL]; ok {
		return v, nil
	}

	return "", service.ErrNotFound
}

// GetShortURL method is used to get shortURL from the repository map
func (repo *MemStorage) GetShortURL(_ context.Context, userUUID uuid.UUID, originalURL string) (string, error) {
	if v, ok := repo.UserUUIDURLMemStore[userUUID][originalURL]; ok {
		return v, nil
	}

	return "", service.ErrNotFound
}

// existsURL check if URL exist in the map
func (repo *MemStorage) existsURL(userUUID uuid.UUID, url string) bool {
	if _, ok := repo.UserUUIDURLMemStore[userUUID][url]; ok {
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
func (repo *MemStorage) Save(_ context.Context, userUUID uuid.UUID, shortURL string, url string) error {

	if shortURL == "" || url == "" {
		return service.ErrEmptyShortURLorURL
	}

	if repo.existsURL(userUUID, url) {
		return fmt.Errorf("%w :%s", service.ErrURLExist, url)
	}

	if repo.existsShortURL(shortURL) {
		return fmt.Errorf("%w :%s", service.ErrShortURLExist, shortURL)
	}

	repo.SlugMemStore[shortURL] = url
	if repo.UserUUIDURLMemStore[userUUID] == nil {
		repo.UserUUIDURLMemStore[userUUID] = make(URLMemStore)
	}
	repo.UserUUIDURLMemStore[userUUID][url] = shortURL
	repo.UUIDMemStore[uuid.New()] = shortURL
	if repo.UserUUIDSlugMemStore[userUUID] == nil {
		repo.UserUUIDSlugMemStore[userUUID] = make(SlugMemStore)
	}
	repo.UserUUIDSlugMemStore[userUUID][shortURL] = url

	return nil
}

// SaveBatch saves slice of shortURL, URL and UUID to the correlated maps
func (repo *MemStorage) SaveBatch(_ context.Context, userUUID uuid.UUID, batch []model.URL) error {

	result := make([]model.URL, len(batch))
	for i := range batch {
		switch {
		case batch[i].ShortURL == "" || batch[i].OriginalURL == "":
			return service.ErrEmptyShortURLorURL
		case repo.existsURL(userUUID, batch[i].OriginalURL):
			return fmt.Errorf("%w: %s", service.ErrURLExist, batch[i].OriginalURL)
		case repo.existsShortURL(batch[i].ShortURL):
			return fmt.Errorf("%w: %s", service.ErrShortURLExist, batch[i].ShortURL)
		}
		result[i] = batch[i]
	}
	if repo.UserUUIDURLMemStore[userUUID] == nil {
		repo.UserUUIDURLMemStore[userUUID] = make(URLMemStore)
	}
	if repo.UserUUIDSlugMemStore[userUUID] == nil {
		repo.UserUUIDSlugMemStore[userUUID] = make(SlugMemStore)
	}
	for i := range result {
		repo.SlugMemStore[result[i].ShortURL] = batch[i].OriginalURL
		repo.UserUUIDSlugMemStore[userUUID][result[i].ShortURL] = batch[i].OriginalURL
		repo.UserUUIDURLMemStore[userUUID][batch[i].OriginalURL] = batch[i].ShortURL
		repo.UUIDMemStore[batch[i].UUID] = batch[i].ShortURL
	}

	return nil
}

func (repo *MemStorage) GetUserShortURLs(_ context.Context, userUUID uuid.UUID) (map[string]string, error) {

	if v, ok := repo.UserUUIDSlugMemStore[userUUID]; ok {
		return v, nil
	}

	return nil, service.ErrNotFound
}
