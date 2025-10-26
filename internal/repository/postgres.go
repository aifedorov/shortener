package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"github.com/aifedorov/shortener/internal/pkg/random"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

// PostgresRepository provides a PostgreSQL-based implementation of the Repository interface.
// It stores URL mappings in a PostgreSQL database with full ACID compliance.
type PostgresRepository struct {
	// ctx is the context for database operations.
	ctx context.Context
	// db is the PostgreSQL database connection.
	db *sql.DB
	// dsn is the PostgreSQL connection string.
	dsn string
	// rand is used for generating random short URL identifiers.
	rand random.Randomizer
}

// Model represents a URL mapping model for database operations.
// It contains all fields needed for storing and retrieving URL mappings.
type Model struct {
	// userID is the ID of the user who created the URL mapping.
	userID string
	// cid is the correlation ID for batch operations.
	cid string
	// alias is the short URL path/alias.
	alias string
	// originalURL is the original URL that was shortened.
	originalURL string
	// baseURL is the base URL used for generating short URLs.
	baseURL string
	// isDeleted indicates if the URL has been marked as deleted.
	isDeleted bool
}

// NewPosgresRepository creates a new PostgreSQL repository instance.
// The repository will use the provided context and DSN for database operations.
func NewPosgresRepository(ctx context.Context, dsn string) *PostgresRepository {
	return &PostgresRepository{
		ctx:  ctx,
		dsn:  dsn,
		rand: random.NewService(),
	}
}

// Run initializes the PostgreSQL repository by opening the database connection.
func (p *PostgresRepository) Run() error {
	logger.Log.Debug("postgres: opening db", zap.String("dsn", p.dsn))
	db, err := sql.Open("pgx", p.dsn)
	if err != nil {
		logger.Log.Error("postgres: failed to open", zap.Error(err))
		return err
	}
	p.db = db

	err = p.Ping()
	if err != nil {
		logger.Log.Error("postgres: failed to ping", zap.Error(err))
		return fmt.Errorf("failed to ping: %w", err)
	}

	return nil
}

// defaultDBTimeout defines the default timeout for database operations.
const defaultDBTimeout = 3 * time.Second

// Ping checks the health of the PostgreSQL database connection.
func (p *PostgresRepository) Ping() error {
	ctx, cancel := context.WithTimeout(p.ctx, defaultDBTimeout)
	defer cancel()

	if err := p.db.PingContext(ctx); err != nil {
		logger.Log.Error("postgres: failed to ping", zap.Error(err))
		return err
	}
	return nil
}

// Close closes the PostgreSQL repository connection and performs cleanup.
func (p *PostgresRepository) Close() error {
	return p.db.Close()
}

// Get retrieves the original URL for a given short URL from the PostgreSQL database.
func (p *PostgresRepository) Get(shortURL string) (string, error) {
	oURL, err := p.fetchOriginalURL(shortURL)
	if errors.Is(err, ErrShortURLNotFound) {
		return "", ErrShortURLNotFound
	}
	if errors.Is(err, ErrURLDeleted) {
		return "", ErrURLDeleted
	}
	if err != nil {
		return "", errors.New("failed to get original URL")
	}
	return oURL, nil
}

// GetAll retrieves all URLs belonging to a specific user from the PostgreSQL database.
func (p *PostgresRepository) GetAll(userID, baseURL string) ([]URLOutput, error) {
	res, err := p.fetchURs(userID, baseURL)
	if errors.Is(err, ErrUserHasNoData) {
		return nil, ErrUserHasNoData
	}
	if err != nil {
		return nil, errors.New("failed to fetch URLs")
	}
	return res, nil
}

// Store saves a new URL to the PostgreSQL database and returns the generated short URL.
func (p *PostgresRepository) Store(userID, baseURL, targetURL string) (string, error) {
	return p.store(userID, baseURL, targetURL)
}

// StoreBatch saves multiple URLs to the PostgreSQL database in a single operation.
func (p *PostgresRepository) StoreBatch(userID, baseURL string, urls []BatchURLInput) ([]BatchURLOutput, error) {
	return p.storeBatch(userID, baseURL, urls)
}

// DeleteBatch marks multiple URLs as deleted for a specific user in the PostgreSQL database.
func (p *PostgresRepository) DeleteBatch(userID string, aliases []string) error {
	return p.deleteBatch(userID, aliases)
}

func (p *PostgresRepository) GetStats() (StatsOutput, error) {
	return p.fetchStats()
}
