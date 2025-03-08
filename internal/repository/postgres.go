package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/aifedorov/shortener/pkg/logger"
	"github.com/aifedorov/shortener/pkg/random"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type PostgresRepository struct {
	ctx  context.Context
	db   *sql.DB
	dsn  string
	rand random.Randomizer
}

func NewPosgresRepository(ctx context.Context, dsn string) *PostgresRepository {
	return &PostgresRepository{
		ctx:  ctx,
		dsn:  dsn,
		rand: random.NewService(),
	}
}

func (p *PostgresRepository) Run() error {
	logger.Log.Debug("postgres: opening db", zap.String("dsn", p.dsn))
	db, err := sql.Open("pgx", p.dsn)
	if err != nil {
		logger.Log.Error("postgres: failed to open", zap.Error(err))
		return err
	}
	p.db = db

	logger.Log.Debug("postgres: creating table")
	ctErr := createTable(p.ctx, db)
	if ctErr != nil {
		logger.Log.Error("postgres: failed to create table", zap.Error(err))
		return err
	}

	return nil
}

const defaultDBTimeout = 3 * time.Second

func (p *PostgresRepository) Ping() error {
	ctx, cancel := context.WithTimeout(p.ctx, defaultDBTimeout)
	defer cancel()

	if err := p.db.PingContext(ctx); err != nil {
		logger.Log.Error("postgres: failed to ping", zap.Error(err))
		return err
	}

	logger.Log.Debug("postgres: ping succeeded")
	return nil
}

func (p *PostgresRepository) Close() error {
	logger.Log.Debug("postgres: closing repository")
	return p.db.Close()
}

func (p *PostgresRepository) Get(shortURL string) (string, error) {
	query := "SELECT original_url FROM urls WHERE alias = $1"
	row := p.db.QueryRowContext(p.ctx, query, shortURL)
	var originalURL string

	err := row.Scan(&originalURL)
	if err != nil {
		return "", err
	}
	return originalURL, nil
}

func (p *PostgresRepository) Store(baseURL, targetURL string) (string, error) {
	alias, genErr := p.rand.GenRandomString(targetURL)
	if genErr != nil {
		logger.Log.Error("postgres: generate random string failed", zap.Error(genErr))
		return "", ErrGenShortURL
	}

	shortURL := baseURL + "/" + alias
	query := "INSERT INTO urls(alias, original_url) VALUES ($1, $2);"
	_, err := p.db.ExecContext(p.ctx, query, alias, targetURL)
	if err != nil {
		logger.Log.Error("postgres: failed to insert url", zap.Error(err))
		return "", err
	}

	logger.Log.Debug("postgres: stored url", zap.String("short_url", shortURL), zap.String("original_url", targetURL))
	return shortURL, nil
}

func createTable(ctx context.Context, db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS urls (
    		id SERIAL,
			alias TEXT NOT NULL,
		 	original_url TEXT NOT NULL
		);
		`
	ctx, cancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer cancel()
	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return err
	}
	return nil
}
