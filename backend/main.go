package main

import (
	"database/sql"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/kntjspr/fullstack-golang-next-template/cmd"
	"github.com/kntjspr/fullstack-golang-next-template/internal/auth"
	"github.com/kntjspr/fullstack-golang-next-template/internal/cache"
	"github.com/kntjspr/fullstack-golang-next-template/internal/config"
	"github.com/kntjspr/fullstack-golang-next-template/internal/database"
	"github.com/kntjspr/fullstack-golang-next-template/internal/logger"
	"github.com/kntjspr/fullstack-golang-next-template/internal/router"
	"github.com/kntjspr/fullstack-golang-next-template/internal/telemetry"
	"github.com/kntjspr/fullstack-golang-next-template/middleware"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	// Create config.
	c, err := config.NewConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}
	log.Logger = logger.New(c.Logger.Level, c.Logger.Pretty)

	if err := auth.RequireJWTSecret(); err != nil {
		log.Fatal().Err(err).Msg("JWT_SECRET is required")
	}

	sentryMiddleware, err := telemetry.InitSentry(c.Sentry)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot initialize sentry")
	}
	if sentryMiddleware != nil {
		defer telemetry.FlushSentry()
	}

	var (
		sqlDB       = (*sql.DB)(nil)
		redisClient = (*redis.Client)(nil)
		gormDB      = (*gorm.DB)(nil)
	)

	if c.Database.Enable {
		openedGormDB, openedDB, err := database.OpenPostgres(c.Database, log.Logger)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot establish postgres connection")
		}
		gormDB = openedGormDB
		sqlDB = openedDB

		defer func() {
			if err := sqlDB.Close(); err != nil {
				log.Error().Err(err).Msg("cannot close postgres connection")
			}
		}()
	}

	if c.Redis.Enable {
		openedRedis, err := cache.OpenRedis(c.Redis, log.Logger)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot establish redis connection")
		}
		redisClient = openedRedis

		defer func() {
			if err := redisClient.Close(); err != nil {
				log.Error().Err(err).Msg("cannot close redis connection")
			}
		}()
	}

	// Create router.
	r := chi.NewRouter()

	if sentryMiddleware != nil {
		r.Use(telemetry.Middleware(sentryMiddleware))
	}

	r.Use(middleware.SecurityHeaders)

	// Set a logger middleware.
	r.Use(chimiddleware.Logger)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(chimiddleware.Timeout(c.Server.ReadTimeout))

	// Get router with all routes.
	router.GetRoutes(r, sqlDB, redisClient, gormDB)

	// Run server instance.
	if err := cmd.Run(c, r); err != nil {
		log.Fatal().Err(err).Msg("cannot run server")
	}
}
