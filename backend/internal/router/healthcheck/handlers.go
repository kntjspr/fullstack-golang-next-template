package healthcheck

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/redis/go-redis/v9"
)

type service struct {
	sqlDB       *sql.DB
	redisClient *redis.Client
}

func newService(sqlDB *sql.DB, redisClient *redis.Client) *service {
	return &service{
		sqlDB:       sqlDB,
		redisClient: redisClient,
	}
}

type componentStatus struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type healthzResponse struct {
	Status     string                     `json:"status"`
	CheckedAt  string                     `json:"checked_at"`
	Components map[string]componentStatus `json:"components"`
}

func getStatus(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{"status": "ok"})
}

func (s *service) getHealthz(w http.ResponseWriter, r *http.Request) {
	response := healthzResponse{
		Status:    "ok",
		CheckedAt: time.Now().UTC().Format(time.RFC3339),
		Components: map[string]componentStatus{
			"app": {
				Status: "up",
			},
		},
	}

	httpStatus := http.StatusOK

	if s.sqlDB != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		err := s.sqlDB.PingContext(ctx)
		cancel()

		if err != nil {
			httpStatus = http.StatusServiceUnavailable
			response.Status = "degraded"
			response.Components["postgres"] = componentStatus{
				Status: "down",
				Error:  err.Error(),
			}
		} else {
			response.Components["postgres"] = componentStatus{
				Status: "up",
			}
		}
	} else {
		response.Components["postgres"] = componentStatus{
			Status: "disabled",
		}
	}

	if s.redisClient != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		err := s.redisClient.Ping(ctx).Err()
		cancel()

		if err != nil {
			httpStatus = http.StatusServiceUnavailable
			response.Status = "degraded"
			response.Components["redis"] = componentStatus{
				Status: "down",
				Error:  err.Error(),
			}
		} else {
			response.Components["redis"] = componentStatus{
				Status: "up",
			}
		}
	} else {
		response.Components["redis"] = componentStatus{
			Status: "disabled",
		}
	}

	render.Status(r, httpStatus)
	render.JSON(w, r, response)
}
