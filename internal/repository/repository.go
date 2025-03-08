package repository

import (
	"context"
	"errors"
	"github.com/aifedorov/shortener/internal/config"
	"github.com/aifedorov/shortener/pkg/logger"
)

var (
	ErrShortURLNotFound = errors.New("repository: short url not found")
	ErrURLExists        = errors.New("repository: url exists")
	ErrGenShortURL      = errors.New("repository: generation short url is failed")
)

type Repository interface {
	Run() error
	Ping() error
	Close() error
	Get(shortURL string) (string, error)
	Store(baseURL, targetURL string) (string, error)
	StoreBatch(baseURL string, urls []URLInput) ([]URLOutput, error)
}

type URLInput struct {
	CID         string
	OriginalURL string
}

type URLOutput struct {
	CID      string
	ShortURL string
}

func NewRepository(ctx context.Context, cfg *config.Config) Repository {
	if cfg.DSN != "" {
		logger.Log.Debug("repository: use posgres storage")
		return NewPosgresRepository(ctx, cfg.DSN)
	}
	if cfg.FileStoragePath != "" {
		logger.Log.Debug("repository: use file storage")
		return NewFileRepository(cfg.FileStoragePath)
	}
	logger.Log.Debug("repository: use in memory storage")
	return NewMemoryRepository()
}
