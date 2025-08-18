package service

import (
	"errors"
	"math/rand"
)

var (
	ErrNotFound = errors.New("not found")
	ErrURLExist = errors.New("URL already exist")
)

const (
	randGenerateSymbols = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	shortURLLen         = 8
)

type Repository interface {
	Get(id string) (string, error)
	Save(id string, url string) error
}

type Service struct {
	r Repository
}

func NewService(r Repository) *Service {
	return &Service{r}
}

func (s Service) Get(id string) (string, error) {
	id, err := s.r.Get(id)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (s Service) GenerateShortURL(url string) (string, error) {
	id := generateShortURL(shortURLLen)
	if err := s.r.Save(id, url); err != nil {
		return id, err
	}

	return id, nil
}

func generateShortURL(length int) string {
	shortURL := make([]byte, length)
	for i := range shortURL {
		shortURL[i] = randGenerateSymbols[rand.Intn(len(randGenerateSymbols))]
	}

	return string(shortURL)
}
