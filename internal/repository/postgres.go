package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/aifedorov/shortener/internal/middleware/logger"
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
	oURL, err := p.fetchOriginalURL(shortURL)
	if err != nil {
		return "", errors.New("failed to get original URL")
	}
	return oURL, nil
}

func (p *PostgresRepository) GetWithUserID(userID, baseURL string) ([]URLOutput, error) {
	res, err := p.fetchURs(userID, baseURL)
	if err != nil {
		return nil, errors.New("failed to fetch URLs")
	}
	return res, nil
}

// TODO: Remove `Store` and use `StoreBatch` instead.
func (p *PostgresRepository) Store(baseURL, targetURL string) (string, error) {
	logger.Log.Debug("postgres: generating alias", zap.String("original_url", targetURL))
	alias, err := p.rand.GenRandomString()
	if err != nil {
		logger.Log.Error("postgres: generate random string failed", zap.Error(err))
		return "", errors.New("failed to generate random string")
	}

	shortURL, err := p.insertAlias(alias, baseURL, targetURL)
	var cErr *ConflictError
	if errors.As(err, &cErr) {
		return "", cErr
	}
	if err != nil {
		return "", errors.New("failed to insert alias")
	}
	return shortURL, nil
}

func (p *PostgresRepository) StoreBatch(baseURL string, urls []BatchURLInput) ([]BatchURLOutput, error) {
	if len(urls) == 0 {
		return nil, errors.New("urls is empty")
	}

	logger.Log.Debug("postgres: begin transaction for storing batch of urls")
	tx, err := p.db.Begin()
	if err != nil {
		logger.Log.Error("postgres: failed to begin transaction", zap.Error(err))
		err := tx.Rollback()
		if err != nil {
			logger.Log.Error("postgres: failed to rollback transaction", zap.Error(err))
			return nil, errors.New("failed to rollback transaction")
		}
		return nil, errors.New("failed to begin transaction")
	}

	logger.Log.Debug("postgres: storing batch of urls", zap.Int("count", len(urls)))
	res := make([]BatchURLOutput, len(urls))
	for i, url := range urls {
		shortURL, err := p.Store(baseURL, url.OriginalURL)
		var cErr *ConflictError
		if errors.As(err, &cErr) {
			return nil, cErr
		}
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				logger.Log.Error("postgres: failed to rollback transaction", zap.Error(err))
				return nil, errors.New("failed to rollback transaction")
			}
		}

		ou := BatchURLOutput{
			CID:      url.CID,
			ShortURL: shortURL,
		}
		res[i] = ou
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

func (p *PostgresRepository) fetchAlias(originalURL string) (string, error) {
	var alias string
	row := p.db.QueryRowContext(p.ctx, "SELECT alias FROM urls WHERE original_url = $1", originalURL)
	err := row.Scan(&alias)
	if err != nil {
		logger.Log.Error("postgres: failed to fetch alias", zap.String("original_url", originalURL), zap.Error(err))
		return "", err
	}
	return alias, nil
}

func (p *PostgresRepository) fetchOriginalURL(alias string) (string, error) {
	query := "SELECT original_url FROM urls WHERE alias = $1"
	row := p.db.QueryRowContext(p.ctx, query, alias)

	var originalURL string
	err := row.Scan(&originalURL)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Log.Error("postgres: original url not found", zap.String("alias", alias))
		return "", ErrShortURLNotFound
	}
	if err != nil {
		logger.Log.Error("postgres: failed to fetch original url", zap.String("alias", alias), zap.Error(err))
		return "", err
	}
	return originalURL, nil
}

func (p *PostgresRepository) fetchURs(userID, baseURL string) ([]URLOutput, error) {
	query := "SELECT alias, original_url FROM urls WHERE user_id = $1"
	rows, err := p.db.QueryContext(p.ctx, query, userID)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Log.Info("postgres: user doesn't have any urls", zap.String("user_id", userID))
		return nil, ErrUserHasNoData
	}
	if err != nil {
		logger.Log.Error("postgres: failed to fetch urls", zap.String("user_id", userID), zap.Error(err))
		return nil, errors.New("failed to fetch urls")
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			logger.Log.Error("postgres: failed to close rows", zap.Error(err))
			return
		}
	}()

	res := make([]URLOutput, 0)
	for rows.Next() {
		var record struct {
			alias       string
			originalURL string
		}

		err := rows.Scan(&record.alias, &record.originalURL)
		if err != nil {
			logger.Log.Error("postgres: failed to fetch all urls", zap.Error(err))
			return nil, errors.New("failed to fetch all urls")
		}
		model := URLOutput{
			ShortURL:    baseURL + "/" + record.alias,
			OriginalURL: record.originalURL,
		}
		res = append(res, model)
	}

	err = rows.Err()
	if err != nil {
		logger.Log.Error("postgres: failed to fetch all urls", zap.Error(err))
		return nil, errors.New("failed to fetch all urls")
	}
	return res, nil
}

func (p *PostgresRepository) insertAlias(alias, baseURL, originalURL string) (string, error) {
	logger.Log.Debug("postgres: inserting alias", zap.String("alias", alias), zap.String("original_url", originalURL))
	var data string
	query := `INSERT INTO urls(cid, alias, original_url)
			VALUES ($1, $2, $3)
			ON CONFLICT (original_url) 
          	DO NOTHING 
          	RETURNING alias;`
	row := p.db.QueryRowContext(p.ctx, query, uuid.NewString(), alias, originalURL)
	err := row.Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Log.Debug("postgres: fetching conflicted url", zap.String("conflict_url", originalURL), zap.Error(err))
		alias, err = p.fetchAlias(originalURL)
		if err != nil {
			logger.Log.Error("postgres: failed to fetch existed url", zap.String("original_url", originalURL), zap.Error(err))
			return "", errors.New("failed to fetch existed url")
		}
		return "", NewConflictError(baseURL+"/"+alias, ErrURLExists)
	}
	if err != nil {
		logger.Log.Error("postgres: failed to insert new url", zap.String("alias", alias), zap.String("original_url", originalURL), zap.Error(err))
		return "", errors.New("failed to insert new url")
	}
	return baseURL + "/" + alias, nil
}
