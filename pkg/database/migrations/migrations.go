package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-kit/log"
)

const (
	createUsersTableSQL = `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			first_name VARCHAR(255) NOT NULL,
			last_name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			phone VARCHAR(50) NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			status VARCHAR(50) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`
)

func ApplyMigrations(db *sql.DB, logger log.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Log("level", "info", "msg", "Starting database schema migrations...")
	if _, err := db.ExecContext(ctx, createUsersTableSQL); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}
	logger.Log("level", "info", "msg", "Database schema migrations completed (tables created/verified).")

	return nil
}
