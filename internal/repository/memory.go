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
	mu        sync.Mutex
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
	ms.mu.Lock()
	defer ms.mu.Unlock()

	targetURL, exists := ms.PathToURL[shortURL]
	if !exists {
		logger.Log.Debug("memory: short url not found", zap.String("short_url", shortURL))
		return "", ErrShortURLNotFound
	}

	return targetURL, nil
}

func (ms *MemoryRepository) GetAll(baseURL string) ([]URLOutput, error) {
	// TODO: Implement me.
	return nil, nil
}

func (ms *MemoryRepository) Store(baseURL, targetURL string) (string, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	alias, err := ms.Rand.GenRandomString()
	if err != nil {
		logger.Log.Debug("memory: generation of random string failed", zap.Error(err))
		return "", err
	}

	resURL := baseURL + "/" + alias
	if _, exists := ms.PathToURL[resURL]; exists {
		logger.Log.Debug("memory: short url already exists", zap.String("resURL", resURL))
		return resURL, nil
	}

	ms.PathToURL[alias] = targetURL
	return resURL, nil
}

func (ms *MemoryRepository) StoreBatch(baseURL string, urls []BatchURLInput) ([]BatchURLOutput, error) {
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
