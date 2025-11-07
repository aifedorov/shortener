package repository

import (
	"database/sql"
	"errors"

	"github.com/aifedorov/shortener/internal/http/middleware/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (p *PostgresRepository) store(userID, baseURL, targetURL string) (string, error) {
	if targetURL == "" {
		logger.Log.Error("postgres: target URL is empty")
		return "", errors.New("target URL is empty")
	}

	logger.Log.Debug("postgres: generating alias", zap.String("original_url", targetURL))
	alias, err := p.rand.GenRandomString()
	if err != nil {
		logger.Log.Error("postgres: generate random string failed", zap.Error(err))
		return "", errors.New("failed to generate random string")
	}

	shortURL, err := p.insert(Model{
		userID:      userID,
		cid:         uuid.NewString(),
		alias:       alias,
		originalURL: targetURL,
		baseURL:     baseURL,
	})
	var cErr *ConflictError
	if errors.As(err, &cErr) {
		return "", cErr
	}
	if err != nil {
		return "", errors.New("failed to insert alias")
	}
	return shortURL, nil
}

func (p *PostgresRepository) storeBatch(userID, baseURL string, urls []BatchURLInput) ([]BatchURLOutput, error) {
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
		shortURL, err := p.store(userID, baseURL, url.OriginalURL)
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
			return nil, errors.New("failed to storing batch of urls")
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

func (p *PostgresRepository) fetchAliasWithUserID(userID, originalURL string) (string, error) {
	var alias string
	row := p.db.QueryRowContext(p.ctx, "SELECT alias FROM urls WHERE original_url = $1 AND user_id = $2", originalURL, userID)
	err := row.Scan(&alias)

	if errors.Is(err, sql.ErrNoRows) {
		logger.Log.Debug("postgres: alias not found for user_id", zap.String("original_url", originalURL), zap.String("user_id", userID))
		return "", ErrShortURLNotFound
	}
	if err != nil {
		logger.Log.Error("postgres: failed to fetch data", zap.Error(err))
		return "", errors.New("failed to fetch data")
	}
	return alias, nil
}

func (p *PostgresRepository) fetchOriginalURL(alias string) (string, error) {
	query := "SELECT original_url, is_deleted FROM urls WHERE alias = $1"
	row := p.db.QueryRowContext(p.ctx, query, alias)

	var model Model
	err := row.Scan(&model.originalURL, &model.isDeleted)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Log.Error("postgres: original url not found", zap.String("alias", alias))
		return "", ErrShortURLNotFound
	}
	if err != nil {
		logger.Log.Error("postgres: failed to fetch original url", zap.Error(err))
		return "", errors.New("failed to fetch original url")
	}
	if model.isDeleted {
		return "", ErrURLDeleted
	}
	return model.originalURL, nil
}

func (p *PostgresRepository) fetchURs(userID, baseURL string) ([]URLOutput, error) {
	query := "SELECT alias, original_url FROM urls WHERE user_id = $1 AND NOT is_deleted"
	rows, err := p.db.QueryContext(p.ctx, query, userID)
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
	if len(res) == 0 {
		return nil, ErrUserHasNoData
	}
	return res, nil
}

func (p *PostgresRepository) insert(model Model) (string, error) {
	var alias string
	query := `INSERT INTO urls(user_id, cid, alias, original_url)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (original_url)
          	DO NOTHING 
          	RETURNING alias;`
	row := p.db.QueryRowContext(p.ctx, query, model.userID, model.cid, model.alias, model.originalURL)

	err := row.Scan(&alias)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Log.Debug("postgres: fetching conflicted url", zap.Error(err))
		alias, err := p.fetchAliasWithUserID(model.userID, model.originalURL)
		if err != nil {
			logger.Log.Error("postgres: failed to fetch existed url", zap.Error(err))
			return "", errors.New("postgres: failed to fetch existed url")
		}
		return "", NewConflictError(model.baseURL+"/"+alias, ErrURLExists)
	}
	if err != nil {
		logger.Log.Error("postgres: failed to insert new url", zap.Error(err))
		return "", errors.New("postgres: failed to insert new url")
	}
	return model.baseURL + "/" + model.alias, nil
}

func (p *PostgresRepository) deleteBatch(userID string, aliases []string) error {
	if len(aliases) == 0 {
		return errors.New("postgres: aliases is empty")
	}

	query := "UPDATE urls SET is_deleted = true WHERE user_id = $1 AND alias = ANY($2)"
	_, err := p.db.ExecContext(p.ctx, query, userID, aliases)
	if err != nil {
		logger.Log.Error("postgres: failed to delete batch of urls", zap.Error(err))
		return errors.New("failed to delete batch of urls")
	}
	return nil
}

func (p *PostgresRepository) fetchStats() (StatsOutput, error) {
	query := `SELECT count(DISTINCT original_url) AS uniq_urls, count(DISTINCT user_id) AS uniq_users FROM urls;`
	row := p.db.QueryRowContext(p.ctx, query)

	var stats StatsOutput
	err := row.Scan(&stats.TotalURLs, &stats.TotalUsers)
	if err != nil {
		logger.Log.Error("postgres: failed to fetch stats", zap.Error(err))
		return StatsOutput{}, errors.New("failed to fetch stats")
	}
	return stats, nil
}
