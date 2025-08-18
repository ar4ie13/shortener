package repository

import (
	"fmt"
	"github.com/ar4ie13/shortener/internal/service"
)

type urlLib map[string]string

type Repository struct {
	urlLib
}

func NewRepository() *Repository {

	return &Repository{
		urlLib: make(map[string]string),
	}
}

func (repo *Repository) Get(id string) (string, error) {
	if link, ok := repo.urlLib[id]; ok {
		return link, nil
	}

	return "", service.ErrNotFound
}

func (repo *Repository) exists(url string) bool {
	for _, v := range repo.urlLib {
		if v == url {
			return true
		}
	}

	return false
}

func (repo *Repository) Save(id string, url string) error {
	if repo.exists(url) {
		return service.ErrURLExist
	}
	repo.urlLib[id] = url
	fmt.Println(repo.urlLib)

	return nil
}
