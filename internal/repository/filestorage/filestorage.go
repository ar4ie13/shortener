package filestorage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

// lastUUID is used for storing last used UUID in store
type lastUUID struct {
	UUID int
}

// FileStorage is a main file storage object contains filePath, store struct and last used UUID
type FileStorage struct {
	urlMapping
	lastUUID
	filePath string
	mu       sync.Mutex
}

// urlMapping used to serialize and deserialize json file storage
type urlMapping struct {
	UUID     int    `json:"uuid"`
	ShortURL string `json:"short_url"`
	URL      string `json:"original_url"`
}

// NewFileStorage constructor receives filepath to store data in file and initializes main file storage object
func NewFileStorage(filepath string) *FileStorage {
	return &FileStorage{
		urlMapping: urlMapping{},
		lastUUID:   lastUUID{},
		filePath:   filepath,
		mu:         sync.Mutex{},
	}
}

// Store is method to store UUID, short_url and original_url in jsonl format
func (fs *FileStorage) Store(shortURL string, url string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.urlMapping.UUID = fs.lastUUID.UUID + 1
	fs.urlMapping.ShortURL = shortURL
	fs.urlMapping.URL = url

	file, err := os.OpenFile(fs.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
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
	file, err := os.ReadFile(fs.filePath)

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
			log.Printf("Error decoding JSON: %v\n", err)
			continue
		}

		shortURLMap[fs.urlMapping.ShortURL] = fs.urlMapping.URL
		log.Printf("Read: UUID=%d, ShortURL=%s\n, URL=%s\n", fs.urlMapping.UUID, fs.urlMapping.ShortURL, fs.urlMapping.URL)
	}

	log.Printf("\nMap contains %d items:\n", len(shortURLMap))
	fmt.Println("shurl: ", shortURLMap)

	fs.lastUUID.UUID = len(shortURLMap)

	return shortURLMap, nil
}
