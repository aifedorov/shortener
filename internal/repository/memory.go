package repository

import (
	"context"
	"github.com/aifedorov/shortener/pkg/logger"
	"sync"

	"github.com/aifedorov/shortener/pkg/random"
	"go.uber.org/zap"
)

type MemoryRepository struct {
	PathToURL sync.Map
	rand      random.Randomizer
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		rand: random.NewService(),
	}
}

func (ms *MemoryRepository) Run(_ context.Context) error {
	return nil
}

func (ms *MemoryRepository) Ping(_ context.Context) error {
	return nil
}

func (ms *MemoryRepository) Close() error {
	return nil
}

func (ms *MemoryRepository) Get(shortURL string) (string, error) {
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

func (ms *MemoryRepository) Store(baseURL, targetURL string) (string, error) {
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
