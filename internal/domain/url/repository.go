package url

import "context"

// Repository defines the interface for working with URL storage.
// This is a port in Hexagonal Architecture terminology.
// Implementations will be located in the infrastructure layer.
type Repository interface {
	// Save saves a URL to storage.
	// Returns ConflictError if URL already exists.
	Save(ctx context.Context, url *URL) error

	// SaveBatch saves multiple URLs in a single operation.
	// Must be an atomic operation (all or nothing).
	SaveBatch(ctx context.Context, urls []*URL) error

	// FindByAlias finds a URL by its short alias.
	// Returns ErrShortURLNotFound if not found.
	FindByAlias(ctx context.Context, alias string) (*URL, error)

	// FindByOriginalURL finds a URL by its original URL.
	// Returns nil if not found (not an error, as this is a valid case).
	FindByOriginalURL(ctx context.Context, userID, originalURL string) (*URL, error)

	// FindByUserID returns all URLs belonging to a user.
	// Returns an empty slice if no URLs found.
	FindByUserID(ctx context.Context, userID string) ([]*URL, error)

	// DeleteBatch marks multiple URLs as deleted (soft delete).
	// Only deletes URLs belonging to the specified user.
	DeleteBatch(ctx context.Context, userID string, aliases []string) error

	// Ping checks storage availability.
	Ping(ctx context.Context) error

	// Close closes the connection to storage.
	Close() error
}

// BatchItem represents an item for batch saving.
type BatchItem struct {
	CorrelationID string // ID for linking request and response
	URL           *URL
}
