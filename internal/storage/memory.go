package storage

import (
	"github.com/aifedorov/shortener/internal/logger"
	"github.com/aifedorov/shortener/lib/random"
	"go.uber.org/zap"
)

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
		logger.Log.Debug("short url not found", zap.String("shortURL", shortURL))
		return "", ErrShortURLNotFound
	}
	return targetURL, nil
}

func (ms *MemoryStorage) SaveURL(baseURL, targetURL string) (string, error) {
	shortURL, genErr := random.GenRandomString(targetURL, shortURLSize)
	if genErr != nil {
		logger.Log.Debug("generation of random string failed", zap.Error(genErr))
		return "", ErrGenShortURL
	}

	resURL := baseURL + "/" + shortURL
	if _, ok := ms.PathToURL[shortURL]; ok {
		logger.Log.Debug("url exists", zap.String("shortURL", shortURL))
		return resURL, ErrURLExists
	}

	ms.PathToURL[shortURL] = targetURL
	return resURL, nil
}
