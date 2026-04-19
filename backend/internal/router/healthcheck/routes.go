package healthcheck

import (
	"database/sql"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

const (
	groupURL  = "/hc"
	statusURL = "/status"
)

// Routes function to create router.
func Routes(m *chi.Mux, sqlDB *sql.DB, redisClient *redis.Client) {
	checks := newService(sqlDB, redisClient)

	// Create group.
	m.Route(groupURL, func(r chi.Router) {
		r.Get(statusURL, getStatus) // get status route
	})

	m.Get("/healthz", checks.getHealthz)
}
