package repository

import (
	"errors"
	"sync"

	"github.com/aifedorov/shortener/internal/pkg/random"
	"go.uber.org/zap"

	"github.com/aifedorov/shortener/internal/http/middleware/logger"
)

// MemoryRepository provides an in-memory implementation of the Repository interface.
// It stores URL mappings in a map with thread-safe access using read-write mutex.
type MemoryRepository struct {
	// PathToURL maps short URL paths to original URLs.
	PathToURL map[string]string
	// Rand is used for generating random short URL identifiers.
	Rand random.Randomizer
	// mu provides thread-safe access to the PathToURL map.
	mu sync.RWMutex
}

// NewMemoryRepository creates a new in-memory repository instance.
// The repository is ready to use immediately after creation.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		PathToURL: make(map[string]string),
		Rand:      random.NewService(),
	}
}

// Run initializes the memory repository.
func (ms *MemoryRepository) Run() error {
	return nil
}

// Ping checks the health of the memory repository connection.
func (ms *MemoryRepository) Ping() error {
	return nil
}

// Close closes the memory repository connection and performs cleanup.
func (ms *MemoryRepository) Close() error {
	return nil
}

// Get retrieves the original URL for a given short URL from memory storage.
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

// GetAll retrieves all URLs belonging to a specific user from memory storage.
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

// Store saves a new URL to memory storage and returns the generated short URL.
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

// StoreBatch saves multiple URLs to memory storage in a single operation.
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

// DeleteBatch marks multiple URLs as deleted for a specific user in memory storage.
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
