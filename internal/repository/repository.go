package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrURLExist      = errors.New("URL already exist")
	ErrEmptyIDorURL  = errors.New("ID or URL cannot be empty")
	ErrShortURLExist = errors.New("ID already exist")
	ErrFileStorage   = errors.New("file storage error")
)

// store used to serialize and deserialize json file storage
type store struct {
	UUID     int    `json:"uuid"`
	ShortURL string `json:"short_url"`
	URL      string `json:"original_url"`
}

// urlLib is used for storing the slice []store{UUID, ShortURL, URL}
type urlLib []store

// fileStoreJSON is used to get file path for json storage
type fileStoreJSON struct {
	filepath string
}

// lastUUID is used for storing last used UUID in urlLib
type lastUUID struct {
	UUID int
}

// Repository is the main object for the package repository
type Repository struct {
	mu sync.Mutex
	urlLib
	fileStoreJSON
	lastUUID
}

// NewRepository is a constructor for Repository object
func NewRepository(filename string) (*Repository, error) {

	repo := &Repository{
		mu:            sync.Mutex{},
		urlLib:        make(urlLib, 0),
		fileStoreJSON: fileStoreJSON{filepath: filename},
	}
	if err := repo.load(); err != nil {
		return nil, err
	}
	return repo, nil
}

// getLastID method is used to get last UUID stored in repo
func (repo *Repository) getLastID() int {
	if len(repo.urlLib) == 0 {
		return 1
	}
	return repo.lastUUID.UUID
}

// Get method is used to get URL (link) from the repository map
func (repo *Repository) Get(shortURL string) (string, error) {
	for _, v := range repo.urlLib {
		if v.ShortURL == shortURL {
			return v.URL, nil
		}
	}

	return "", ErrNotFound
}

// existsURL check if URL exist in the map
func (repo *Repository) existsURL(url string) bool {
	for _, v := range repo.urlLib {
		if v.URL == url {
			return true
		}
	}

	return false
}

// existsShortURL check if URL exist in the map
func (repo *Repository) existsShortURL(shortURL string) bool {
	for _, v := range repo.urlLib {
		if v.ShortURL == shortURL {
			return true
		}
	}

	return false
}

// Save saves the id(shortURL):URL pair in the map
func (repo *Repository) Save(shortURL string, url string) error {

	if shortURL == "" || url == "" {
		return ErrEmptyIDorURL
	}

	if repo.existsURL(url) {
		return ErrURLExist
	}

	if repo.existsShortURL(shortURL) {
		return ErrShortURLExist
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()
	lUUID := repo.lastUUID.UUID + 1
	repo.urlLib = append(repo.urlLib, store{lUUID, shortURL, url})

	file, err := json.MarshalIndent(repo.urlLib, "", "  ")
	if err != nil {
		return err
	}
	repo.lastUUID.UUID = lUUID

	err = os.WriteFile(repo.fileStoreJSON.filepath, file, 0644)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrFileStorage, err)
	}
	return nil
}

// load reads data from JSON file
func (repo *Repository) load() error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	file, err := os.ReadFile(repo.fileStoreJSON.filepath)

	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if len(file) == 0 {
		return nil
	}

	err = json.Unmarshal(file, &repo.urlLib)
	if err != nil {
		return fmt.Errorf("json unmarshal error of file:%s : %w", repo.fileStoreJSON.filepath, err)
	}

	repo.lastUUID.UUID = len(repo.urlLib)

	return nil
}
