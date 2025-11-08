package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/ar4ie13/shortener/internal/model"
	"github.com/ar4ie13/shortener/internal/myerrors"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const (
	randGenerateSymbols  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	shortURLLen          = 8
	timeToCollectDeleted = 2 * time.Second
)

// Repository interface used to interact with repository package to store or retrieve values
type Repository interface {
	GetURL(ctx context.Context, shortURL string) (string, error)
	GetShortURL(ctx context.Context, originalURL string) (string, error)
	Save(ctx context.Context, userUUID uuid.UUID, shortURL string, url string) error
	SaveBatch(ctx context.Context, userUUID uuid.UUID, batch []model.URL) error
	GetUserShortURLs(ctx context.Context, userUUID uuid.UUID) (map[string]string, error)
	DeleteUserShortURLs(ctx context.Context, shortURLsToDelete map[uuid.UUID][]string) error
}

// Service is a main object of the package that implements Repository interface
type Service struct {
	repo         Repository
	toDeleteChan []chan map[uuid.UUID][]string
	zlog         zerolog.Logger
}

// NewService is a constructor for Service object
func NewService(r Repository, zlog zerolog.Logger) *Service {
	srv := &Service{
		repo:         r,
		toDeleteChan: []chan map[uuid.UUID][]string{},
		zlog:         zlog,
	}
	go srv.deleteShortURLs()
	return srv
}

// GetURL method gets URL by provided id
func (s *Service) GetURL(ctx context.Context, userUUID uuid.UUID, shortURL string) (string, error) {
	if shortURL == "" {
		return "", myerrors.ErrEmptyID
	}

	getURL, err := s.repo.GetURL(ctx, shortURL)
	if getURL == "" || err != nil {
		return "", fmt.Errorf("failed to get URL: %w", err)
	}

	return getURL, nil
}

// GetUserShortURLs method gets all shortURLs and URL saved by user
func (s *Service) GetUserShortURLs(ctx context.Context, userUUID uuid.UUID) (map[string]string, error) {
	result, err := s.repo.GetUserShortURLs(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get short urls: %w", err)
	}

	return result, nil
}

// SaveURL generates shortURL for non-existent URL and stores it in the Repository
func (s *Service) SaveURL(ctx context.Context, userUUID uuid.UUID, urlLink string) (slug string, err error) {
	urlLink = strings.TrimRight(urlLink, "/")

	if urlLink == "" {
		return "", myerrors.ErrEmptyURL
	}

	// Validate the URL format
	parsedURL, err := url.Parse(urlLink)
	if err != nil {
		return "", myerrors.ErrInvalidURLFormat
	}

	// Ensure the scheme is http or https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", myerrors.ErrWrongHTTPScheme
	}

	// Ensure the host is not empty
	if parsedURL.Host == "" {
		return "", myerrors.ErrMustIncludeHost
	}

	for attempt := 1; attempt <= 3; attempt++ {
		slug, err = generateShortURL(shortURLLen)

		if err != nil {
			if attempt == 3 {
				return "", err
			}
			continue
		}

		err = s.repo.Save(ctx, userUUID, slug, urlLink)

		if err == nil {
			return slug, nil
		}

		if errors.Is(err, myerrors.ErrURLExist) {
			slug, err = s.repo.GetShortURL(ctx, urlLink)
			if err != nil {
				return "", myerrors.ErrNotFound
			}
			return slug, myerrors.ErrURLExist
		}

		if attempt == 3 {
			if errors.Is(err, myerrors.ErrShortURLExist) {
				return "", fmt.Errorf("failed to save URL to repository: %w", err)
			}
		}
	}
	return "", err
}

// SaveBatch saves batch of jsonl rows to the repository
func (s *Service) SaveBatch(ctx context.Context, userUUID uuid.UUID, batch []model.URL) ([]model.URL, error) {
	result := make([]model.URL, len(batch))
	for i := range batch {
		slug, err := generateShortURL(shortURLLen)
		if err != nil {
			return nil, fmt.Errorf("failed to generate short url: %w", err)
		}

		urlLink := strings.TrimRight(batch[i].OriginalURL, "/")

		if urlLink == "" {
			return nil, fmt.Errorf("%w: %s", myerrors.ErrEmptyURL, batch[i].OriginalURL)
		}

		// Validate the URL format
		parsedURL, err := url.Parse(batch[i].OriginalURL)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", myerrors.ErrInvalidURLFormat, batch[i].OriginalURL)
		}

		// Ensure the scheme is http or https
		if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
			return nil, myerrors.ErrWrongHTTPScheme
		}

		// Ensure the host is not empty
		if parsedURL.Host == "" {
			return nil, myerrors.ErrMustIncludeHost
		}
		result[i] = model.URL{
			ShortURL:    slug,
			OriginalURL: urlLink,
			UUID:        batch[i].UUID,
		}
	}

	err := s.repo.SaveBatch(ctx, userUUID, result)
	if err != nil {
		return nil, fmt.Errorf("failed to save batch: %w", err)
	}

	return result, nil
}

// generateShortURL is a sub-function for SaveURL
func generateShortURL(length int) (string, error) {
	if length <= 0 {
		return "", myerrors.ErrShortURLLength
	}

	shortURL := make([]byte, length)
	for i := range shortURL {
		shortURL[i] = randGenerateSymbols[rand.Intn(len(randGenerateSymbols))]
	}

	return string(shortURL), nil
}

// SendShortURLForDelete writes userUUID and slugs for deletion to toDeleteChan slice of channels
func (s *Service) SendShortURLForDelete(_ context.Context, userUUID uuid.UUID, shortURLs []string) {
	data := make(chan map[uuid.UUID][]string, 1)
	defer close(data)
	data <- map[uuid.UUID][]string{userUUID: shortURLs}
	s.toDeleteChan = append(s.toDeleteChan, data)
}

// collectShortURLs aggregates all channels from toDeleteChan slice into resulting one
func (s *Service) collectShortURLs() chan map[uuid.UUID][]string {
	finalCh := make(chan map[uuid.UUID][]string)

	var wg sync.WaitGroup

	for _, ch := range s.toDeleteChan {
		chClosure := ch
		wg.Add(1)
		go func() {
			defer wg.Done()
			for data := range chClosure {
				finalCh <- data
			}
		}()
	}
	go func() {
		wg.Wait()
		close(finalCh)
	}()
	return finalCh
}

// deleteShortURLs runs as a separate goroutine and periodically send batch of slugs for deletion to repository
func (s *Service) deleteShortURLs() {
	ticker := time.NewTicker(timeToCollectDeleted)
	defer ticker.Stop()

	shortURLsForDelete := make(map[uuid.UUID][]string)
	var ch = make(chan map[uuid.UUID][]string)
	for {
		select {
		case toDelete := <-ch:
			for k, v := range toDelete {
				shortURLsForDelete[k] = append(shortURLsForDelete[k], v...)
			}

		case <-ticker.C:
			ch = s.collectShortURLs()
			if len(shortURLsForDelete) == 0 {
				continue
			}
			err := s.repo.DeleteUserShortURLs(context.TODO(), shortURLsForDelete)
			if err != nil {
				s.zlog.Err(err).Msg("failed to delete short urls")
			}
			shortURLsForDelete = make(map[uuid.UUID][]string)
		}
	}
}
