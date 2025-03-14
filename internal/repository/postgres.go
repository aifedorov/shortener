package repository

import (
	"context"
	"database/sql"
	"errors"
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
	err = p.createTable()
	if err != nil {
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
	logger.Log.Debug("postgres: generating short url", zap.String("original_url", targetURL))
	newAlias, err := p.rand.GenRandomString()
	if err != nil {
		logger.Log.Error("postgres: generate random string failed", zap.Error(err))
		return "", err
	}

	logger.Log.Debug("postgres: inserting url", zap.String("alias", newAlias), zap.String("original_url", targetURL))
	var alias string
	query := `INSERT INTO urls(cid, alias, original_url)
			VALUES ($1, $2, $3)
			ON CONFLICT (original_url) 
          	DO NOTHING 
          	RETURNING alias;`
	row := p.db.QueryRowContext(p.ctx, query, uuid.NewString(), newAlias, targetURL)
	err = row.Scan(&alias)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Log.Debug("postgres: fetching conflicted url", zap.String("conflict_url", targetURL), zap.Error(err))
		row := p.db.QueryRowContext(p.ctx, "SELECT alias FROM urls WHERE original_url = $1", targetURL)
		err := row.Scan(&alias)
		if err != nil {
			logger.Log.Error("postgres: failed to fetch existed url", zap.String("original_url", targetURL), zap.Error(err))
			return "", err
		}
		return "", NewConflictError(baseURL+"/"+alias, ErrURLExists)
	}
	if err != nil {
		logger.Log.Error("postgres: failed to insert new url", zap.String("alias", newAlias), zap.String("original_url", targetURL), zap.Error(err))
		return "", err
	}

	return baseURL + "/" + alias, nil
}

func (p *PostgresRepository) StoreBatch(baseURL string, urls []URLInput) ([]URLOutput, error) {
	if len(urls) == 0 {
		return nil, nil
	}

	logger.Log.Debug("postgres: begin transaction for storing batch of urls")
	tx, err := p.db.Begin()
	if err != nil {
		logger.Log.Error("postgres: failed to begin transaction", zap.Error(err))
		err := tx.Rollback()
		if err != nil {
			logger.Log.Error("postgres: failed to rollback transaction", zap.Error(err))
			return nil, err
		}
		return nil, err
	}

	logger.Log.Debug("postgres: storing batch of urls", zap.Int("count", len(urls)))
	res := make([]URLOutput, len(urls))
	for i, url := range urls {
		alias, err := p.rand.GenRandomString()
		if err != nil {
			logger.Log.Error("postgres: generate random string failed", zap.Error(err))
			err := tx.Rollback()
			if err != nil {
				logger.Log.Error("postgres: failed to rollback transaction", zap.Error(err))
				return nil, err
			}
			return nil, err
		}

		query := "INSERT INTO urls(cid, alias, original_url) VALUES ($1, $2, $3);"
		_, err = tx.ExecContext(p.ctx, query, url.CID, alias, url.OriginalURL)
		if err != nil {
			logger.Log.Error("postgres: failed to insert url", zap.Error(err))
			err := tx.Rollback()
			if err != nil {
				logger.Log.Error("postgres: failed to rollback transaction", zap.Error(err))
				return nil, err
			}
			return nil, err
		}

		ou := URLOutput{
			CID:      url.CID,
			ShortURL: baseURL + "/" + alias,
		}
		res[i] = ou
		logger.Log.Debug("postgres: url stored", zap.String("cid", ou.CID), zap.String("url", ou.ShortURL))
	}

	logger.Log.Debug("postgres: commiting transaction for storing batch of urls")
	return res, tx.Commit()
}

func (p *PostgresRepository) createTable() error {
	tx, err := p.db.BeginTx(p.ctx, nil)
	if err != nil {
		logger.Log.Error("postgres: failed to create table", zap.Error(err))
		return err
	}

	query := `CREATE TABLE IF NOT EXISTS urls (
    		id SERIAL PRIMARY KEY,
    		cid CHAR(36) NOT NULL,
			alias TEXT NOT NULL UNIQUE,
		 	original_url TEXT NOT NULL UNIQUE,
		 	created TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		`
	ctx, cancel := context.WithTimeout(p.ctx, defaultDBTimeout)
	defer cancel()

	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		logger.Log.Error("postgres: failed to create table", zap.Error(err))
		return tx.Rollback()
	}
	return tx.Commit()
}
