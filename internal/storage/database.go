package storage

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Database struct {
	*sqlx.DB
}

// Новое подключение к БД
func NewDatabase(dsn string) (*Database, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return &Database{db}, nil
}

// Закрывает подключение к БД
func (db *Database) CloseConnect() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
