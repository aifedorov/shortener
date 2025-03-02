package storage

import (
	"sync"

	"github.com/aifedorov/shortener/internal/logger"
	"github.com/aifedorov/shortener/lib/random"
	"go.uber.org/zap"
)

type MemoryStorage struct {
	PathToURL sync.Map
	rand      random.Randomizer
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		rand: random.NewService(),
	}
}

func (ms *MemoryStorage) GetURL(shortURL string) (string, error) {
	targetURL, exists := ms.PathToURL.Load(shortURL)
	if !exists {
		logger.Log.Debug("short url not found", zap.String("shortURL", shortURL))
		return "", ErrShortURLNotFound
	}
	if targetURLStr, ok := targetURL.(string); ok {
		return targetURLStr, nil
	}

	return "", ErrShortURLNotFound
}

func (ms *MemoryStorage) SaveURL(baseURL, targetURL string) (string, error) {
	shortURL, genErr := ms.rand.GenRandomString(targetURL)
	if genErr != nil {
		logger.Log.Debug("generation of random string failed", zap.Error(genErr))
		return "", ErrGenShortURL
	}

	resURL := baseURL + "/" + shortURL
	if _, loaded := ms.PathToURL.LoadOrStore(shortURL, targetURL); loaded {
		logger.Log.Debug("url exists", zap.String("shortURL", shortURL))
		return resURL, ErrURLExists
	}

	return resURL, nil
}
