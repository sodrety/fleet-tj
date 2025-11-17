// internal/api/health.go
package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rabbitmq/amqp091-go"
)

type Health struct {
	DB       string `json:"db"`
	RabbitMQ string `json:"rabbitmq"`
}

func HealthCheck(db *pgxpool.Pool, rabbit *amqp091.Connection) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		status := Health{DB: "ok", RabbitMQ: "ok"}

		// DB
		if err := db.Ping(ctx); err != nil {
			status.DB = "error: " + err.Error()
		}

		// RabbitMQ
		if rabbit.IsClosed() {
			status.RabbitMQ = "closed"
		}

		if status.DB == "ok" && status.RabbitMQ == "ok" {
			c.JSON(http.StatusOK, status)
		} else {
			c.JSON(http.StatusServiceUnavailable, status)
		}
	}
}
