package publisher

import (
	"context"
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
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

func PublishGeofence(ctx context.Context, conn *amqp091.Connection, event GeofenceEvent) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare("fleet.events", "direct", true, false, false, false, nil)
	if err != nil {
		return err
	}

	body, _ := json.Marshal(event)
	return ch.PublishWithContext(ctx, "fleet.events", "geofence.entry", false, false, amqp091.Publishing{ContentType: "application/json", Body: body})
}
