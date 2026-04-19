package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config struct for describe configuration of the app.
type Config struct {
	StageStatus string
	Server      *ServerConfig
	Logger      *LoggerConfig
	Database    *DatabaseConfig
	Redis       *RedisConfig
	Sentry      *SentryConfig
}

type ServerConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type LoggerConfig struct {
	Level  string
	Pretty bool
}

type DatabaseConfig struct {
	Enable   bool
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
	TimeZone string
}

type RedisConfig struct {
	Enable   bool
	Host     string
	Port     int
	Password string
	DB       int
}

func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type SentryConfig struct {
	Enable           bool
	DSN              string
	Environment      string
	Release          string
	TracesSampleRate float64
}

// DSN returns PostgreSQL connection string for GORM.
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		d.Host,
		d.Port,
		d.User,
		d.Password,
		d.Name,
		d.SSLMode,
		d.TimeZone,
	)
}

// NewConfig function to prepare config variables from .env file and return config.
func NewConfig() (*Config, error) {
	host := getEnv("SERVER_HOST", "0.0.0.0")

	port, err := intEnv("SERVER_PORT", "5000")
	if err != nil {
		return nil, err
	}
	readTimeout, err := intEnv("SERVER_READ_TIMEOUT", "5")
	if err != nil {
		return nil, err
	}
	writeTimeout, err := intEnv("SERVER_WRITE_TIMEOUT", "10")
	if err != nil {
		return nil, err
	}
	idleTimeout, err := intEnv("SERVER_IDLE_TIMEOUT", "120")
	if err != nil {
		return nil, err
	}
	loggerPretty, err := boolEnv("LOGGER_PRETTY", "false")
	if err != nil {
		return nil, err
	}
	dbEnable, err := boolEnv("DB_ENABLE", "false")
	if err != nil {
		return nil, err
	}
	dbPort, err := intEnv("DB_PORT", "5432")
	if err != nil {
		return nil, err
	}
	redisEnable, err := boolEnv("REDIS_ENABLE", "false")
	if err != nil {
		return nil, err
	}
	redisPort, err := intEnv("REDIS_PORT", "6379")
	if err != nil {
		return nil, err
	}
	redisDB, err := intEnv("REDIS_DB", "0")
	if err != nil {
		return nil, err
	}
	sentryTracesSampleRate, err := float64Env("SENTRY_TRACES_SAMPLE_RATE", "0")
	if err != nil {
		return nil, err
	}

	stageStatus := strings.ToLower(strings.TrimSpace(getEnv("STAGE_STATUS", "dev")))
	switch stageStatus {
	case "dev", "prod":
	default:
		return nil, errors.New("wrong STAGE_STATUS (expected dev or prod)")
	}

	sentryDSN := getEnv("SENTRY_DSN", "")

	return &Config{
		StageStatus: stageStatus,
		Server: &ServerConfig{
			Addr:         fmt.Sprintf("%s:%d", host, port),
			ReadTimeout:  time.Duration(readTimeout) * time.Second,
			WriteTimeout: time.Duration(writeTimeout) * time.Second,
			IdleTimeout:  time.Duration(idleTimeout) * time.Second,
		},
		Logger: &LoggerConfig{
			Level:  getEnv("LOGGER_LEVEL", "info"),
			Pretty: loggerPretty,
		},
		Database: &DatabaseConfig{
			Enable:   dbEnable,
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "postgres"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
			TimeZone: getEnv("DB_TIMEZONE", "UTC"),
		},
		Redis: &RedisConfig{
			Enable:   redisEnable,
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     redisPort,
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		Sentry: &SentryConfig{
			Enable:           sentryDSN != "",
			DSN:              sentryDSN,
			Environment:      getEnv("SENTRY_ENVIRONMENT", "development"),
			Release:          getEnv("SENTRY_RELEASE", ""),
			TracesSampleRate: sentryTracesSampleRate,
		},
	}, nil
}

func getEnv(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}

	return fallback
}

func intEnv(name, fallback string) (int, error) {
	value, err := strconv.Atoi(getEnv(name, fallback))
	if err != nil {
		return 0, fmt.Errorf("wrong %s (check your .env): %w", name, err)
	}
	return value, nil
}

func boolEnv(name, fallback string) (bool, error) {
	value, err := strconv.ParseBool(getEnv(name, fallback))
	if err != nil {
		return false, fmt.Errorf("wrong %s (check your .env): %w", name, err)
	}
	return value, nil
}

func float64Env(name, fallback string) (float64, error) {
	value, err := strconv.ParseFloat(getEnv(name, fallback), 64)
	if err != nil {
		return 0, fmt.Errorf("wrong %s (check your .env): %w", name, err)
	}
	return value, nil
}
