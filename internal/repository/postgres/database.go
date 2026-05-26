// Пакет postgres реализовывает работу с базами данных
package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Database - обёртка для sqlx.DB
type Database struct {
	*sqlx.DB
}

// NewDatabase создаёт новое подключение к БД
func NewDatabase(dsn string) (*Database, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return &Database{db}, nil
}

// CloseConnect закрывает подключение к БД
func (db *Database) CloseConnect() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// WithTransaction выполняет функцию в транзакции.
// Автоматически коммитит при успехе, откат при ошибке.
func (db *Database) WithTransaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil { // перехватываем panic
			_ = tx.Rollback()
			panic(p) // пробрасываем panic дальше после отката
		}
	}()

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
