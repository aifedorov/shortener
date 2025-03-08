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
	db   *sql.DB
	dsn  string
	rand random.Randomizer
}

func NewPosgresRepository(dsn string) *PostgresRepository {
	return &PostgresRepository{
		dsn:  dsn,
		rand: random.NewService(),
	}
}

func (repo *PostgresRepository) Run(ctx context.Context) error {
	logger.Log.Debug("postgres: opening db", zap.String("dsn", repo.dsn))
	db, err := sql.Open("pgx", repo.dsn)
	if err != nil {
		logger.Log.Error("postgres: failed to open", zap.Error(err))
		return err
	}
	repo.db = db
	logger.Log.Debug("postgres: opened db")

	logger.Log.Debug("postgres: creating table")
	ctErr := createTable(ctx, db)
	if ctErr != nil {
		logger.Log.Error("postgres: failed to create table", zap.Error(err))
		return err
	}

	logger.Log.Debug("postgres: table created")
	return nil
}

const defaultDBTimeout = 3 * time.Second

func (repo *PostgresRepository) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer cancel()

	if err := repo.db.PingContext(ctx); err != nil {
		logger.Log.Error("postgres: failed to ping", zap.Error(err))
		return err
	}

	logger.Log.Debug("postgres: ping succeeded")
	return nil
}

func (repo *PostgresRepository) Close() error {
	logger.Log.Debug("postgres: closing repository")
	return repo.db.Close()
}

func (repo *PostgresRepository) Get(_ string) (string, error) {
	// TODO: implement me
	return "", nil
}

func (repo *PostgresRepository) Store(_, _ string) (string, error) {
	// TODO: implement me
	return "", nil
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
