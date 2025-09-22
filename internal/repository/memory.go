package repository

import (
	"errors"
	"sync"

	"github.com/aifedorov/shortener/internal/pkg/random"
	"go.uber.org/zap"

	"github.com/aifedorov/shortener/internal/http/middleware/logger"
)

type MemoryRepository struct {
	PathToURL map[string]string
	Rand      random.Randomizer
	mu        sync.RWMutex
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		PathToURL: make(map[string]string),
		Rand:      random.NewService(),
	}
}

func (ms *MemoryRepository) Run() error {
	return nil
}

func (ms *MemoryRepository) Ping() error {
	return nil
}

func (ms *MemoryRepository) Close() error {
	return nil
}

func (ms *MemoryRepository) Get(shortURL string) (string, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	targetURL, exists := ms.PathToURL[shortURL]
	if !exists {
		logger.Log.Debug("memory: short url not found", zap.String("short_url", shortURL))
		return "", ErrShortURLNotFound
	}

	return targetURL, nil
}

func (ms *MemoryRepository) GetAll(_, baseURL string) ([]URLOutput, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	res := make([]URLOutput, len(ms.PathToURL))
	i := 0
	for alias, url := range ms.PathToURL {
		res[i] = URLOutput{
			ShortURL:    baseURL + "/" + alias,
			OriginalURL: url,
		}
		i++
	}
	return res, nil
}

func (ms *MemoryRepository) Store(_, baseURL, targetURL string) (string, error) {
	alias, err := ms.Rand.GenRandomString()
	if err != nil {
		logger.Log.Debug("memory: generation of random string failed", zap.Error(err))
		return "", err
	}

	resURL := baseURL + "/" + alias
	ms.mu.RLock()
	if _, exists := ms.PathToURL[resURL]; exists {
		logger.Log.Debug("memory: short url already exists", zap.String("resURL", resURL))
		return resURL, nil
	}
	ms.mu.RUnlock()

	ms.mu.Lock()
	ms.PathToURL[alias] = targetURL
	ms.mu.Unlock()

	return resURL, nil
}

func (ms *MemoryRepository) StoreBatch(_, baseURL string, urls []BatchURLInput) ([]BatchURLOutput, error) {
	if len(urls) == 0 {
		return nil, nil
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	logger.Log.Debug("memory: storing batch of urls", zap.Int("count", len(urls)))
	res := make([]BatchURLOutput, len(urls))
	for i, url := range urls {
		alias, err := ms.Rand.GenRandomString()
		if err != nil {
			logger.Log.Debug("memory: generation of random string failed", zap.Error(err))
			return nil, err
		}

		resURL := baseURL + "/" + alias
		ou := BatchURLOutput{
			CID:      url.CID,
			ShortURL: resURL,
		}
		res[i] = ou
		ms.PathToURL[alias] = url.OriginalURL
	}

	logger.Log.Debug("memory: store updated", zap.Any("store", ms.PathToURL))
	return res, nil
}

func (ms *MemoryRepository) DeleteBatch(_ string, aliases []string) error {
	if len(aliases) == 0 {
		return errors.New("memory: aliases is empty")
	}

	ms.mu.Lock()
	for _, alias := range aliases {
		delete(ms.PathToURL, alias)
	}
	ms.mu.Unlock()
	return nil
}
