package filestorage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/ar4ie13/shortener/internal/model"
	fileconf "github.com/ar4ie13/shortener/internal/repository/filestorage/config"
	"github.com/ar4ie13/shortener/internal/repository/memory"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// FileStorage is a main file storage object contains filePath, store struct and last used UUID
type FileStorage struct {
	m          *memory.MemStorage
	urlMapping model.URL
	filePath   fileconf.Config
	zlog       zerolog.Logger
	mu         sync.Mutex
}

// NewFileStorage constructor receives filePath to store data in file and initializes main file storage object
func NewFileStorage(filePath fileconf.Config, zlog zerolog.Logger) *FileStorage {
	return &FileStorage{
		m:          memory.NewMemStorage(),
		urlMapping: model.URL{},
		filePath:   filePath,
		zlog:       zlog,
		mu:         sync.Mutex{},
	}
}

// Load reads data from JSON file into maps
func (fs *FileStorage) Load() error {
	err := fs.LoadFile()
	if err != nil {
		return err
	}

	return nil
}

// GetURL method is used to get URL (link) from the map
func (fs *FileStorage) GetURL(ctx context.Context, shortURL string) (string, error) {
	urlLink, err := fs.m.GetURL(ctx, shortURL)
	if err != nil {
		return "", err
	}

	return urlLink, nil
}

// GetShortURL method is used to get URL (link) from the map
func (fs *FileStorage) GetShortURL(ctx context.Context, originalURL string) (string, error) {
	slug, err := fs.m.GetShortURL(ctx, originalURL)
	if err != nil {
		return "", err
	}

	return slug, nil
}

// Save is a method used to save short url and original url
func (fs *FileStorage) Save(ctx context.Context, shortURL string, url string) error {
	if err := fs.m.Save(ctx, shortURL, url); err != nil {
		return err
	}

	if err := fs.Store(shortURL, url); err != nil {
		return err
	}

	return nil
}

// Store is method to store UUID, short_url and original_url in jsonl format
func (fs *FileStorage) Store(shortURL string, url string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.urlMapping.UUID = uuid.New()
	fs.urlMapping.ShortURL = shortURL
	fs.urlMapping.OriginalURL = url

	file, err := os.OpenFile(fs.filePath.FilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	jsonLine, err := json.Marshal(fs.urlMapping)
	if err != nil {
		return fmt.Errorf("cannot marshal json: %w", err)
	}
	_, err = file.Write(jsonLine)
	if err != nil {
		return fmt.Errorf("cannot write to file: %w", err)
	}
	_, err = file.WriteString("\n")
	if err != nil {
		return fmt.Errorf("cannot write to file: %w", err)
	}

	return nil
}

// LoadFile loads json file storage and returns maps for memory storage
func (fs *FileStorage) LoadFile() error {

	file, err := os.ReadFile(fs.filePath.FilePath)

	if err != nil {
		if os.IsNotExist(err) {

			return nil
		}
		return err
	}

	if len(file) == 0 {
		return nil
	}

	buf := bytes.NewBuffer(file)
	decoder := json.NewDecoder(buf)

	for {

		err = decoder.Decode(&fs.urlMapping)

		if err != nil {
			// Check for EOF
			if err.Error() == "EOF" {
				break
			}
			fs.zlog.Debug().Msgf("error decoding JSON: %v\n", err)
			continue
		}

		fs.m.SlugMemStore[fs.urlMapping.ShortURL] = fs.urlMapping.OriginalURL
		fs.m.URLMemStore[fs.urlMapping.OriginalURL] = fs.urlMapping.ShortURL
		fs.m.UUIDMemStore[fs.urlMapping.UUID] = fs.urlMapping.ShortURL
		fs.zlog.Debug().Msgf("read: UUID=%s, ShortURL=%s, URL=%s", fs.urlMapping.UUID, fs.urlMapping.ShortURL, fs.urlMapping.OriginalURL)

	}

	fs.zlog.Debug().Msgf("filestorage red successfully, map contains %d items", len(fs.m.SlugMemStore))

	return nil
}

// SaveBatch used to save batch of short urls and URL to the file storage
func (fs *FileStorage) SaveBatch(ctx context.Context, batch []model.URL) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	err := fs.m.SaveBatch(ctx, batch)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(fs.filePath.FilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	for i := range batch {
		fs.urlMapping.UUID = batch[i].UUID
		fs.urlMapping.ShortURL = batch[i].ShortURL
		fs.urlMapping.OriginalURL = batch[i].OriginalURL

		jsonLine, err := json.Marshal(fs.urlMapping)
		if err != nil {
			return fmt.Errorf("cannot marshal json: %w", err)
		}
		_, err = file.Write(jsonLine)
		if err != nil {
			return fmt.Errorf("cannot write to file: %w", err)
		}
		_, err = file.WriteString("\n")
		if err != nil {
			return fmt.Errorf("cannot write to file: %w", err)
		}
	}

	return nil
}
