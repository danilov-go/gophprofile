// Package db реализует хранилище данных в базе PostgreSQL.
package db

import (
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type storageDB struct {
	db *sql.DB
}

// NewStorageDB создает новый экземпляр storageDB.
func NewStorageDB(sql *sql.DB) *storageDB {
	return &storageDB{
		db: sql,
	}
}

// InitDB инициализирует подключение к базе данных PostgreSQL по переданной строке соединения.
func InitDB(ps string) (*storageDB, error) {
	db, err := sql.Open("pgx", ps)
	if err != nil {
		return nil, err
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, errors.Join(err, closeErr)
		}
		return nil, err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, errors.Join(err, closeErr)
		}
		return nil, err
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		if closeErr := db.Close(); closeErr != nil {
			return nil, errors.Join(err, closeErr)
		}
		return nil, err
	}
	return &storageDB{
		db: db,
	}, nil
}
