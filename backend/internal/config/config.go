package config

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

// Config struct for describe configuration of the app.
type Config struct {
	Server   *ServerConfig
	Logger   *LoggerConfig
	Database *DatabaseConfig
	Redis    *RedisConfig
	Sentry   *SentryConfig
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

var (
	once     sync.Once // create sync.Once primitive
	instance *Config   // create nil Config struct
)

// NewConfig function to prepare config variables from .env file and return config.
func NewConfig() *Config {
	// Configuring config one time.
	once.Do(func() {
		// Server host (should be string):
		host := getEnv("SERVER_HOST", "0.0.0.0")
		// Server port (should be int):
		port, err := strconv.Atoi(getEnv("SERVER_PORT", "5000"))
		if err != nil {
			panic("wrong server port (check your .env)")
		}
		// Server read timeout (should be int):
		readTimeout, err := strconv.Atoi(getEnv("SERVER_READ_TIMEOUT", "5"))
		if err != nil {
			panic("wrong server read timeout (check your .env)")
		}
		// Server write timeout (should be int):
		writeTimeout, err := strconv.Atoi(getEnv("SERVER_WRITE_TIMEOUT", "10"))
		if err != nil {
			panic("wrong server write timeout (check your .env)")
		}
		// Server idle timeout (should be int):
		idleTimeout, err := strconv.Atoi(getEnv("SERVER_IDLE_TIMEOUT", "120"))
		if err != nil {
			panic("wrong server idle timeout (check your .env)")
		}
		// Logger pretty format (should be bool):
		loggerPretty, err := strconv.ParseBool(getEnv("LOGGER_PRETTY", "false"))
		if err != nil {
			panic("wrong logger pretty flag (check your .env)")
		}
		// Database enable flag (should be bool):
		dbEnable, err := strconv.ParseBool(getEnv("DB_ENABLE", "false"))
		if err != nil {
			panic("wrong database enable flag (check your .env)")
		}
		// Database port (should be int):
		dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
		if err != nil {
			panic("wrong database port (check your .env)")
		}
		// Redis enable flag (should be bool):
		redisEnable, err := strconv.ParseBool(getEnv("REDIS_ENABLE", "false"))
		if err != nil {
			panic("wrong redis enable flag (check your .env)")
		}
		// Redis port (should be int):
		redisPort, err := strconv.Atoi(getEnv("REDIS_PORT", "6379"))
		if err != nil {
			panic("wrong redis port (check your .env)")
		}
		// Redis DB index (should be int):
		redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
		if err != nil {
			panic("wrong redis db index (check your .env)")
		}
		// Sentry traces sample rate (should be float):
		sentryTracesSampleRate, err := strconv.ParseFloat(getEnv("SENTRY_TRACES_SAMPLE_RATE", "0"), 64)
		if err != nil {
			panic("wrong sentry traces sample rate (check your .env)")
		}
		sentryDSN := getEnv("SENTRY_DSN", "")

		// Set all variables to the config instance.
		instance = &Config{
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
		}
	})

	// Return configured config instance.
	return instance
}

func getEnv(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}

	return fallback
}
