package config

import (
	"fmt"
	"os"

	"github.com/go-kit/log"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
	DBPort     string
	SSLMode    string
	TimeZone   string

	ServerPort string
}

func LoadConfig(logger log.Logger) (*Config, error) {
	_ = godotenv.Load() // For local development, but env vars take precedence in production

	viper.SetDefault("APP_PORT", "8089")
	viper.AutomaticEnv()

	var cfg Config
	var missingCriticalConfig []string
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if cfg.DBHost == "" {
		cfg.DBHost = os.Getenv("DB_HOST")
		if cfg.DBHost == "" {
			missingCriticalConfig = append(missingCriticalConfig, "DB_HOST")
		}
	}

	if cfg.DBUser == "" {
		cfg.DBUser = os.Getenv("DB_USER")
		if cfg.DBUser == "" {
			missingCriticalConfig = append(missingCriticalConfig, "DB_USER")
		}
	}

	if cfg.DBPassword == "" {
		cfg.DBPassword = os.Getenv("DB_PASSWORD")
		if cfg.DBPassword == "" {
			missingCriticalConfig = append(missingCriticalConfig, "DB_PASSWORD")
		}
	}

	cfg.DBName = os.Getenv("DB_NAME")
	if cfg.DBName == "" {
		cfg.DBName = os.Getenv("DB_NAME")
		if cfg.DBName == "" {
			missingCriticalConfig = append(missingCriticalConfig, "DB_NAME")
		}
	}

	if cfg.DBPort == "" {
		cfg.DBPort = os.Getenv("DB_PORT")
		if cfg.DBPort == "" {
			logger.Log("warning", "DB_PORT environment variable not set, using default port 5432.")
			cfg.DBPort = "5432"
		}
	}

	if cfg.SSLMode == "" {
		cfg.SSLMode = os.Getenv("DB_SSLMODE")
		if cfg.SSLMode == "" {
			logger.Log("warning", "DB_SSLMODE environment variable not set, using default sslmode disable.")
			cfg.SSLMode = "disable"
		}
	}

	if cfg.TimeZone == "" {
		cfg.TimeZone = os.Getenv("DB_TIMEZONE")
		if cfg.TimeZone == "" {
			logger.Log("warning", "DB_TIMEZONE environment variable not set, using default TimeZone UTC.")
			cfg.TimeZone = "UTC"
		}
	}

	if cfg.ServerPort == "" {
		cfg.ServerPort = os.Getenv("SERVER_PORT")
		if cfg.ServerPort == "" {
			logger.Log("info", "SERVER_PORT environment variable not set, using default port 8088.")
			cfg.ServerPort = "8088"
		}
	}

	if len(missingCriticalConfig) > 0 {
		return nil, fmt.Errorf("missing critical environment variables: %v", missingCriticalConfig)
	}

	return &cfg, nil
}
