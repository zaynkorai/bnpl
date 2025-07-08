package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"bnpl/pkg/config"
	"bnpl/pkg/database/migrations"

	"github.com/go-kit/log"
)

func NewPostgresDB(cfg *config.Config, logger log.Logger) (*sql.DB, error) {

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.SSLMode, cfg.TimeZone)

	var db *sql.DB
	maxAttempts := 5
	for i := 0; i < maxAttempts; i++ {
		var openErr error
		db, openErr = sql.Open("postgres", dsn)
		if openErr != nil {
			logger.Log("level", "error", "msg", fmt.Sprintf("attempt %d/%d: failed to open database connection", i+1, maxAttempts), "err", openErr)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		pingErr := db.PingContext(ctx)
		cancel()

		if pingErr != nil {
			logger.Log("level", "error", "msg", fmt.Sprintf("attempt %d/%d: failed to ping database", i+1, maxAttempts), "err", pingErr)
			db.Close()
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		logger.Log("level", "info", "msg", "Database connection established.")
		break
	}

	if db == nil {
		return nil, fmt.Errorf("failed to connect to database after %d retries", maxAttempts)
	}

	if err := migrations.ApplyMigrations(db, logger); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	return db, nil
}
