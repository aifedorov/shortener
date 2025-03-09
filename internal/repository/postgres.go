package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/aifedorov/shortener/pkg/logger"
	"github.com/aifedorov/shortener/pkg/random"
	"github.com/google/uuid"
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
	logger.Log.Debug("postgres: generating short url", zap.String("url", targetURL))
	alias, genErr := p.rand.GenRandomString(targetURL)
	if genErr != nil {
		logger.Log.Error("postgres: generate random string failed", zap.Error(genErr))
		return "", ErrGenShortURL
	}

	logger.Log.Debug("postgres: inserting url", zap.String("alias", alias), zap.String("url", targetURL))
	query := "INSERT INTO urls(cid, alias, original_url) VALUES ($1, $2, $3);"
	_, err := p.db.ExecContext(p.ctx, query, uuid.NewString(), alias, targetURL)
	if err != nil {
		logger.Log.Error("postgres: failed to insert url", zap.Error(err))
		return "", err
	}

	shortURL := baseURL + "/" + alias
	return shortURL, nil
}

func (p *PostgresRepository) StoreBatch(baseURL string, urls []URLInput) ([]URLOutput, error) {
	if len(urls) == 0 {
		return nil, nil
	}

	logger.Log.Debug("postgres: begin transaction for storing batch of urls")
	tx, err := p.db.Begin()
	if err != nil {
		logger.Log.Error("postgres: failed to begin transaction", zap.Error(err))
		_ = tx.Rollback()
		return nil, err
	}

	logger.Log.Debug("postgres: storing batch of urls", zap.Int("count", len(urls)))
	res := make([]URLOutput, len(urls))
	for i, url := range urls {
		alias, genErr := p.rand.GenRandomString(url.OriginalURL)
		if genErr != nil {
			logger.Log.Error("postgres: generate random string failed", zap.Error(genErr))
			_ = tx.Rollback()
			return nil, ErrGenShortURL
		}

		query := "INSERT INTO urls(cid, alias, original_url) VALUES ($1, $2, $3);"
		_, err := tx.ExecContext(p.ctx, query, url.CID, alias, url.OriginalURL)
		if err != nil {
			logger.Log.Error("postgres: failed to insert url", zap.Error(err))
			_ = tx.Rollback()
			return nil, err
		}

		ou := URLOutput{
			CID:      url.CID,
			ShortURL: baseURL + "/" + alias,
		}
		res[i] = ou
		logger.Log.Debug("postgres: url stored", zap.String("cid", ou.CID), zap.String("url", ou.ShortURL))
	}

	logger.Log.Debug("postgres: urls stored", zap.Int("count", len(res)))
	logger.Log.Debug("postgres: commiting transaction for storing batch of urls")
	return res, tx.Commit()
}

func createTable(ctx context.Context, db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS urls (
    		id SERIAL PRIMARY KEY,
    		cid CHAR(36) NOT NULL,
			alias TEXT NOT NULL,
		 	original_url TEXT NOT NULL,
		 	created TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
