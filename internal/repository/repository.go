package repository

import (
	"github.com/ar4ie13/shortener/internal/service"
)

// urlLib is used for storing the map id(shortURL):URL
type urlLib map[string]string

// Repository is the main object for the package repository
type Repository struct {
	urlLib
}

// NewRepository is a contructor for Repository object
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

	return "", service.ErrNotFound
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
		return service.ErrInvalidIDorURL
	}

	if repo.exists(url) {
		return service.ErrURLExist
	}

	repo.urlLib[id] = url

	return nil
}
