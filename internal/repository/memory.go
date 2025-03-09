package repository

import (
	"github.com/aifedorov/shortener/pkg/logger"
	"github.com/aifedorov/shortener/pkg/random"
	"go.uber.org/zap"
	"sync"
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
		logger.Log.Debug("short url not found", zap.String("shortURL", shortURL))
		return "", ErrShortURLNotFound
	}

	return targetURL, nil
}

func (ms *MemoryRepository) Store(baseURL, targetURL string) (string, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	alias, genErr := ms.Rand.GenRandomString(targetURL)
	if genErr != nil {
		logger.Log.Debug("generation of random string failed", zap.Error(genErr))
		return "", ErrGenShortURL
	}

	resURL := baseURL + "/" + alias

	if _, exists := ms.PathToURL[resURL]; exists {
		logger.Log.Debug("short url already exists", zap.String("resURL", resURL))
		return resURL, nil
	}

	ms.PathToURL[resURL] = targetURL
	return resURL, nil
}

func (ms *MemoryRepository) StoreBatch(baseURL string, urls []URLInput) ([]URLOutput, error) {
	if len(urls) == 0 {
		return nil, nil
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	logger.Log.Debug("memory: storing batch of urls", zap.Int("count", len(urls)))
	res := make([]URLOutput, len(urls))
	for i, url := range urls {
		alias, genErr := ms.Rand.GenRandomString(url.OriginalURL)
		if genErr != nil {
			logger.Log.Debug("memory: generation of random string failed", zap.Error(genErr))
			return nil, ErrGenShortURL
		}

		resURL := baseURL + "/" + alias
		ou := URLOutput{
			CID:      url.CID,
			ShortURL: resURL,
		}
		res[i] = ou
		ms.PathToURL[alias] = url.OriginalURL
	}

	logger.Log.Debug("memory: store updated", zap.Any("store", ms.PathToURL))
	return res, nil
}
