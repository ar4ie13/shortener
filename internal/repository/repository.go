package repository

import (
	"errors"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrURLExist     = errors.New("URL already exist")
	ErrEmptyIDorURL = errors.New("ID or URL cannot be empty")
	ErrIDExist      = errors.New("ID already exist")
)

// urlLib is used for storing the map id(shortURL):URL
type urlLib map[string]string

// Repository is the main object for the package repository
type Repository struct {
	urlLib
}

// NewRepository is a constructor for Repository object
func NewRepository() *Repository {
	return &Repository{
		urlLib: make(map[string]string),
	}
}

// Get method is used to get URL (link) from the repository map
func (repo *Repository) Get(id string) (string, error) {
	if link, ok := repo.urlLib[id]; ok {
		return link, nil
	}

	return "", ErrNotFound
}

// exists check if URL exist in the map
func (repo *Repository) exists(url string) bool {
	for _, v := range repo.urlLib {
		if v == url {
			return true
		}
	}

	return false
}

// Save saves the id(shortURL):URL pair in the map
func (repo *Repository) Save(id string, url string) error {
	if id == "" || url == "" {
		return ErrEmptyIDorURL
	}
	if repo.exists(url) {
		return ErrURLExist
	}

	if _, ok := repo.urlLib[id]; ok {
		return ErrIDExist
	}

	repo.urlLib[id] = url

	return nil
}
