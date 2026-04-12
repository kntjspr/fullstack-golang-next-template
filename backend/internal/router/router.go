package router

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"gorm.io/gorm"

	"github.com/kntjspr/fullstack-golang-next-template/internal/router/healthcheck"
	"github.com/kntjspr/fullstack-golang-next-template/internal/swagger"
)

// GetRoutes function for getting routes.
func GetRoutes(m *chi.Mux, sqlDB *sql.DB, redisClient *redis.Client, gormDB *gorm.DB) {
	healthcheck.Routes(m, sqlDB, redisClient) // health check routes
	RegisterAuthRoutes(m, gormDB)
	UsersRoutes(m, gormDB)
	m.Get("/swagger/spec", swagger.SpecHandler)
	m.Get("/swagger/ui", swagger.UIHandler)
	m.Get("/openapi.yaml", swagger.SpecHandler) // backward-compatible alias
	m.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/spec")))
	m.NotFound(http.NotFound) // not found routes
}
