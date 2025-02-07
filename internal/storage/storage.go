package storage

import (
	"errors"
	"github.com/aifedorov/shortener/lib/random"
)

var (
	ErrURLNotFound = errors.New("storage: url not found")
	ErrURLExists   = errors.New("storage: url exists")
	ErrGenShortURL = errors.New("storage: generation short url is failed")
)

type Storage interface {
	GetURL(shortURL string) (string, error)
	SaveURL(baseURL, targetURL string) (string, error)
}

type MemoryStorage struct {
	PathToURL map[string]string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		PathToURL: make(map[string]string),
	}
}

const shortURLSize = 8

func (ms *MemoryStorage) GetURL(shortURL string) (string, error) {
	targetURL, exists := ms.PathToURL[shortURL]
	if !exists {
		return "", ErrURLNotFound
	}
	return targetURL, nil
}

func (ms *MemoryStorage) SaveURL(baseURL, targetURL string) (string, error) {
	shortURL, genErr := random.GenRandomString(targetURL, shortURLSize)
	if genErr != nil {
		return "", ErrGenShortURL
	}

	resURL := baseURL + "/" + shortURL
	if _, ok := ms.PathToURL[shortURL]; ok {
		return resURL, ErrURLExists
	}

	ms.PathToURL[shortURL] = targetURL
	return resURL, nil
}
