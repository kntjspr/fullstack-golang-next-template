package database

import (
	"database/sql"

	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	"github.com/create-go-app/chi-go-template/internal/config"
)

// OpenPostgres creates and validates a PostgreSQL connection for GORM.
func OpenPostgres(cfg *config.DatabaseConfig, log zerolog.Logger) (*gorm.DB, *sql.DB, error) {
	db, err := gorm.Open(
		postgres.Open(cfg.DSN()),
		&gorm.Config{
			Logger: gormLogger.Default.LogMode(gormLogger.Warn),
		},
	)
	if err != nil {
		return nil, nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, nil, err
	}

	log.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Str("database", cfg.Name).
		Msg("postgres connection established")

	return db, sqlDB, nil
}
