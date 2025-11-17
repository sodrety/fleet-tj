package api

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rabbitmq/amqp091-go"
	"github.com/sodrety/fleet-tj/internal/repository"
	"go.uber.org/zap"
)

func SetupRoutes(r *gin.Engine, repo *repository.LocationRepo, db *pgxpool.Pool, rabbit *amqp091.Connection, logger *zap.Logger, mqttPub mqtt.Client) {
	v1 := r.Group("/vehicles")
	{
		v1.GET("/:vehicle_id/location", GetLatest(repo))
		v1.GET("/:vehicle_id/history", GetHistory(repo))
		v1.POST("/locate", LocateVehicle(repo, logger, mqttPub))
	}

	r.GET("/health", HealthCheck(db, rabbit))
}
