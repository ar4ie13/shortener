package filestorage

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/ar4ie13/shortener/internal/model"
	fileconf "github.com/ar4ie13/shortener/internal/repository/filestorage/config"
	"github.com/ar4ie13/shortener/internal/repository/memory"
	"github.com/rs/zerolog"
)

// lastUUID is used for storing last used UUID in store
type lastUUID struct {
	UUID int
}

// FileStorage is a main file storage object contains filePath, store struct and last used UUID
type FileStorage struct {
	m          *memory.MemStorage
	urlMapping model.URL
	lastUUID
	filePath fileconf.Config
	zlog     zerolog.Logger
	mu       sync.Mutex
}

// NewFileStorage constructor receives filePath to store data in file and initializes main file storage object
func NewFileStorage(filePath fileconf.Config, zlog zerolog.Logger) *FileStorage {
	return &FileStorage{
		m:          memory.NewMemStorage(),
		urlMapping: model.URL{},
		lastUUID:   lastUUID{},
		filePath:   filePath,
		zlog:       zlog,
		mu:         sync.Mutex{},
	}
}

// Load reads data from JSON file into maps
func (fs *FileStorage) Load() error {
	shortURLMap, err := fs.LoadFile()
	if err != nil {
		return err
	}

	err = fs.m.Load(shortURLMap)
	if err != nil {
		return err
	}

	return nil
}

// Get method is used to get URL (link) from the map
func (fs *FileStorage) Get(ctx context.Context, shortURL string) (string, error) {
	slug, err := fs.m.Get(ctx, shortURL)
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

	fs.urlMapping.ID = fs.lastUUID.UUID + 1
	fs.urlMapping.ShortURL = shortURL
	fs.urlMapping.OriginalURL = url

	file, err := os.OpenFile(fs.filePath.FilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	jsonLine, err := json.Marshal(fs.urlMapping)
	if err != nil {
		panic(err)
	}
	_, err = file.Write(jsonLine)
	if err != nil {
		panic(err)
	}
	_, err = file.WriteString("\n")
	if err != nil {
		panic(err)
	}

	fs.lastUUID.UUID++

	return nil

}

// LoadFile loads json file storage and returns maps for memory storage
func (fs *FileStorage) LoadFile() (shortURLMap map[string]string, err error) {
	shortURLMap = make(map[string]string)
	file, err := os.ReadFile(fs.filePath.FilePath)

	if err != nil {
		if os.IsNotExist(err) {

			return shortURLMap, nil
		}
		return shortURLMap, err
	}

	if len(file) == 0 {
		return shortURLMap, nil
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

		shortURLMap[fs.urlMapping.ShortURL] = fs.urlMapping.OriginalURL
		fs.zlog.Debug().Msgf("read: UUID=%d, ShortURL=%s, URL=%s", fs.urlMapping.ID, fs.urlMapping.ShortURL, fs.urlMapping.OriginalURL)

	}

	fs.zlog.Debug().Msgf("map contains %d items", len(shortURLMap))

	fs.lastUUID.UUID = len(shortURLMap)

	return shortURLMap, nil
}
