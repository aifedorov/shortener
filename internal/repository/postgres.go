package repository

import "database/sql"

type PostgresRepository struct {
	DB *sql.DB
}
