package repository

import (
	"context"
	"errors"
	"github.com/aifedorov/shortener/internal/config"
	"github.com/aifedorov/shortener/pkg/logger"
)

type ConflictError struct {
	ShortURL string
	err      error
}

func NewConflictError(shortURL string, err error) error {
	return &ConflictError{
		ShortURL: shortURL,
		err:      err,
	}
}

func (e *ConflictError) Error() string {
	return e.err.Error()
}

var (
	ErrShortURLNotFound = errors.New("short url not found")
	ErrURLExists        = errors.New("url exists")
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
