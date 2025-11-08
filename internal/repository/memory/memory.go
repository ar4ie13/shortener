package memory

import (
	"context"
	"fmt"

	"github.com/ar4ie13/shortener/internal/model"
	"github.com/ar4ie13/shortener/internal/myerrors"
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

type IsSlugDeletedMemStore map[string]bool

// MemStorage is the main object for the package repository
type MemStorage struct {
	SlugMemStore
	URLMemStore
	UserUUIDURLMemStore
	UUIDMemStore
	UserUUIDSlugMemStore
	IsSlugDeletedMemStore
}

// NewMemStorage is a constructor for MemStorage object
func NewMemStorage() *MemStorage {
	return &MemStorage{
		SlugMemStore:          make(map[string]string),
		URLMemStore:           make(map[string]string),
		UserUUIDURLMemStore:   make(map[uuid.UUID]URLMemStore),
		UUIDMemStore:          make(map[uuid.UUID]string),
		UserUUIDSlugMemStore:  make(map[uuid.UUID]SlugMemStore),
		IsSlugDeletedMemStore: make(map[string]bool),
	}
}

// GetURL method is used to get URL (link) from the repository map
func (repo *MemStorage) GetURL(_ context.Context, shortURL string) (string, error) {
	if v, ok := repo.SlugMemStore[shortURL]; ok {
		if repo.IsSlugDeletedMemStore[shortURL] {
			return "", myerrors.ErrShortURLIsDeleted
		} else {
			return v, nil
		}
	}

	return "", myerrors.ErrNotFound
}

// GetShortURL method is used to get shortURL from the repository map
func (repo *MemStorage) GetShortURL(_ context.Context, originalURL string) (string, error) {
	if v, ok := repo.URLMemStore[originalURL]; ok {
		if !repo.IsSlugDeletedMemStore[v] {
			return v, nil
		}

	}

	return "", myerrors.ErrNotFound
}

// existsURL check if URL exist in the map
func (repo *MemStorage) existsURL(url string) bool {
	if v, ok := repo.URLMemStore[url]; ok {
		if !repo.IsSlugDeletedMemStore[v] {
			return true
		}

	}

	return false
}

// existsShortURL check if URL exist in the map
func (repo *MemStorage) existsShortURL(shortURL string) bool {
	if _, ok := repo.SlugMemStore[shortURL]; ok {
		if !repo.IsSlugDeletedMemStore[shortURL] {
			return true
		}
	}

	return false
}

// Save saves shortURL, URL and UUID to the correlated maps
func (repo *MemStorage) Save(_ context.Context, userUUID uuid.UUID, shortURL string, url string) error {

	if shortURL == "" || url == "" {
		return myerrors.ErrEmptyShortURLorURL
	}

	if repo.existsURL(url) {
		return fmt.Errorf("%w :%s", myerrors.ErrURLExist, url)
	}

	if repo.existsShortURL(shortURL) {
		return fmt.Errorf("%w :%s", myerrors.ErrShortURLExist, shortURL)
	}

	repo.SlugMemStore[shortURL] = url
	repo.URLMemStore[url] = shortURL
	if repo.UserUUIDURLMemStore[userUUID] == nil {
		repo.UserUUIDURLMemStore[userUUID] = make(URLMemStore)
	}
	repo.UserUUIDURLMemStore[userUUID][url] = shortURL
	repo.UUIDMemStore[uuid.New()] = shortURL
	if repo.UserUUIDSlugMemStore[userUUID] == nil {
		repo.UserUUIDSlugMemStore[userUUID] = make(SlugMemStore)
	}
	repo.UserUUIDSlugMemStore[userUUID][shortURL] = url
	repo.IsSlugDeletedMemStore[shortURL] = false

	return nil
}

// SaveBatch saves slice of shortURL, URL and UUID to the correlated maps
func (repo *MemStorage) SaveBatch(_ context.Context, userUUID uuid.UUID, batch []model.URL) error {

	result := make([]model.URL, len(batch))
	for i := range batch {
		switch {
		case batch[i].ShortURL == "" || batch[i].OriginalURL == "":
			return myerrors.ErrEmptyShortURLorURL
		case repo.existsURL(batch[i].OriginalURL):
			return fmt.Errorf("%w: %s", myerrors.ErrURLExist, batch[i].OriginalURL)
		case repo.existsShortURL(batch[i].ShortURL):
			return fmt.Errorf("%w: %s", myerrors.ErrShortURLExist, batch[i].ShortURL)
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
		repo.URLMemStore[result[i].OriginalURL] = result[i].ShortURL
		repo.SlugMemStore[result[i].ShortURL] = batch[i].OriginalURL
		repo.UserUUIDSlugMemStore[userUUID][result[i].ShortURL] = batch[i].OriginalURL
		repo.UserUUIDURLMemStore[userUUID][batch[i].OriginalURL] = batch[i].ShortURL
		repo.UUIDMemStore[batch[i].UUID] = batch[i].ShortURL
		repo.IsSlugDeletedMemStore[batch[i].ShortURL] = false
	}

	return nil
}

// GetUserShortURLs return short URLs for specified user
func (repo *MemStorage) GetUserShortURLs(_ context.Context, userUUID uuid.UUID) (map[string]string, error) {
	result := make(SlugMemStore)
	if _, ok := repo.UserUUIDSlugMemStore[userUUID]; !ok {
		return nil, myerrors.ErrNotFound
	}

	for slug, url := range repo.UserUUIDSlugMemStore[userUUID] {
		if !repo.IsSlugDeletedMemStore[slug] {
			result[slug] = url
		}
	}

	return result, nil
}

// DeleteUserShortURLs mark short URLs as Deleted in storage
func (repo *MemStorage) DeleteUserShortURLs(_ context.Context, shortURLsToDelete map[uuid.UUID][]string) error {
	for userUUID, slugs := range shortURLsToDelete {

		if _, ok := repo.UserUUIDSlugMemStore[userUUID]; !ok {
			return myerrors.ErrInvalidUserUUID
		}
		for _, slug := range slugs {
			if repo.UserUUIDSlugMemStore[userUUID][slug] != "" {
				if _, ok := repo.UserUUIDSlugMemStore[userUUID][slug]; ok {
					repo.IsSlugDeletedMemStore[slug] = true
				}
			}
		}
	}

	return nil
}
