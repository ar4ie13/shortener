package repository

import (
	"github.com/ar4ie13/shortener/internal/repository/filestorage"
	"github.com/ar4ie13/shortener/internal/repository/memory"
)

// Repository is a main repository object contains both memory and file storage
type Repository struct {
	m *memory.MemStorage
	f *filestorage.FileStorage
}

// NewRepository constructs repository object
func NewRepository(filepath string) (*Repository, error) {
	repo := &Repository{
		m: memory.NewMemStorage(),
		f: filestorage.NewFileStorage(filepath),
	}
	err := repo.Load()
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// Save is a method used to save short url and original url
func (repo *Repository) Save(shortURL string, url string) error {
	if err := repo.m.Save(shortURL, url); err != nil {
		return err
	}

	if err := repo.f.Store(shortURL, url); err != nil {
		return err
	}

	return nil
}

// Get method is used to get URL (link) from the map
func (repo *Repository) Get(shortURL string) (string, error) {
	slug, err := repo.m.Get(shortURL)
	if err != nil {
		return "", err
	}

	return slug, nil
}

// Load reads data from JSON file into maps
func (repo *Repository) Load() error {
	shortURLMap, err := repo.f.LoadFile()
	if err != nil {
		return err
	}

	err = repo.m.Load(shortURLMap)
	if err != nil {
		return err
	}

	return nil
}
