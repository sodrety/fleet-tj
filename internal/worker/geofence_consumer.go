package worker

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type GeofenceEvent struct {
	VehicleID string `json:"vehicle_id"`
	Event     string `json:"event"`
	Location  struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"location"`
	Timestamp int64 `json:"timestamp"`
}

func StartGeofenceConsumer(ctx context.Context, conn *amqp091.Connection, logger *zap.Logger) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// Declare exchange and queue (idempotent)
	if err := ch.ExchangeDeclare("fleet.events", "direct", true, false, false, false, nil); err != nil {
		return err
	}
	q, err := ch.QueueDeclare("geofence_alerts", true, false, false, false, nil)
	if err != nil {
		return err
	}
	if err := ch.QueueBind(q.Name, "geofence.entry", "fleet.events", false, nil); err != nil {
		return err
	}

	// Fair dispatch
	if err := ch.Qos(1, 0, false); err != nil {
		return err
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	logger.Info("geofence worker started", zap.String("queue", q.Name))

	go func() {
		for msg := range msgs {
			var event GeofenceEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				logger.Error("invalid geofence event", zap.Error(err))
				msg.Nack(false, true)
				continue
			}

			logger.Info("geofence entry detected",
				zap.String("vehicle", event.VehicleID),
				zap.Float64("lat", event.Location.Latitude),
				zap.Float64("lon", event.Location.Longitude),
			)

			// testing purpose
			msg.Ack(false)
		}
	}()

	// Block until shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	return nil
}
