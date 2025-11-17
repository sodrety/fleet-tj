package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sodrety/fleet-tj/internal/api"
	"github.com/sodrety/fleet-tj/internal/mqtt"
	"github.com/sodrety/fleet-tj/internal/repository"
	"github.com/sodrety/fleet-tj/internal/service"
	"github.com/sodrety/fleet-tj/internal/worker"
	"github.com/sodrety/fleet-tj/pkg/config"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Config
	gf, err := config.LoadGeofence()
	if err != nil {
		logger.Fatal("load geofence", zap.Error(err))
	}

	// DB
	pool, err := pgxpool.New(context.Background(), os.Getenv("DB_DSN"))
	if err != nil {
		logger.Fatal("db connect", zap.Error(err))
	}
	defer pool.Close()

	// RabbitMQ
	rabbitConn, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		logger.Fatal("rabbitmq", zap.Error(err))
	}
	defer rabbitConn.Close()

	// MQTT
	locChan := make(chan repository.Location, 1000)
	go func() {
		if err := mqtt.Start(context.Background(), os.Getenv("MQTT_BROKER"), locChan); err != nil {
			logger.Fatal("mqtt", zap.Error(err))
		}
	}()

	// MQTT Publisher
	if err := mqtt.InitPublisher(os.Getenv("MQTT_BROKER")); err != nil {
		logger.Fatal("mqtt publisher init failed", zap.Error(err))
	}

	// Services
	repo := repository.NewLocationRepo(pool)
	locService := service.NewLocationService(repo, rabbitConn, gf, logger)

	// Worker Pool
	ctx, cancel := context.WithCancel(context.Background())
	for i := 0; i < 5; i++ {
		go func() {
			for loc := range locChan {
				locService.Process(ctx, loc)
			}
		}()
	}

	// Geofence Consumer
	go func() {
		if err := worker.StartGeofenceConsumer(ctx, rabbitConn, logger); err != nil {
			logger.Error("geofence worker failed", zap.Error(err))
		}
	}()

	// API
	r := gin.Default()

	// Cors
	r.Use(cors.Default())

	api.SetupRoutes(r, repo, pool, rabbitConn, logger, mqtt.Publisher)
	go r.Run(":8080")

	// Graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	cancel()
	time.Sleep(2 * time.Second)
}
