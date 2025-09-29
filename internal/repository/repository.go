package repository

import (
	"context"
	"errors"

	"github.com/aifedorov/shortener/internal/config"
	"github.com/aifedorov/shortener/internal/http/middleware/logger"
)

// ConflictError represents an error that occurs when a URL already exists in the repository.
type ConflictError struct {
	ShortURL string
	err      error
}

// NewConflictError creates a new ConflictError with the given short URL and underlying error.
func NewConflictError(shortURL string, err error) error {
	return &ConflictError{
		ShortURL: shortURL,
		err:      err,
	}
}

// Error returns the error message from the underlying error.
func (e *ConflictError) Error() string {
	return e.err.Error()
}

// Repository error definitions
var (
	// ErrShortURLNotFound is returned when a requested short URL does not exist in the repository.
	ErrShortURLNotFound = errors.New("short url not found")
	// ErrURLExists is returned when attempting to store a URL that already exists.
	ErrURLExists = errors.New("url exists")
	// ErrUserHasNoData is returned when a user has no URLs stored in the repository.
	ErrUserHasNoData = errors.New("user has no data")
	// ErrURLDeleted is returned when attempting to access a URL that has been marked as deleted.
	ErrURLDeleted = errors.New("url deleted")
)

// Repository defines the interface for URL storage operations.
type Repository interface {
	// Run initializes the repository and performs any necessary setup.
	Run() error
	// Ping checks the health of the repository connection.
	Ping() error
	// Close closes the repository connection and performs cleanup.
	Close() error
	// Get retrieves the original URL for a given short URL.
	Get(shortURL string) (string, error)
	// GetAll retrieves all URLs belonging to a specific user.
	GetAll(userID, baseURL string) ([]URLOutput, error)
	// Store saves a new URL and returns the generated short URL.
	Store(userID, baseURL, targetURL string) (string, error)
	// StoreBatch saves multiple URLs in a single operation and returns the generated short URLs.
	StoreBatch(userID, baseURL string, urls []BatchURLInput) ([]BatchURLOutput, error)
	// DeleteBatch marks multiple URLs as deleted for a specific user.
	DeleteBatch(userID string, aliases []string) error
}

// NewRepository creates a new repository instance based on the provided configuration.
// It returns a PostgreSQL repository if DSN is configured, a file repository if FileStoragePath is configured,
// or an in-memory repository as fallback.
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
