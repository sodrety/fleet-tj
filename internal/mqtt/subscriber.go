package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sodrety/fleet-tj/internal/repository"
	"go.uber.org/zap"
)

// ---------------------------------------------------------------------
// 1. Handler – keep the struct (it is never used, but we keep it)
// ---------------------------------------------------------------------
type Handler struct {
	locChan chan<- repository.Location
	logger  *zap.Logger
}

// ---------------------------------------------------------------------
// 2. Global callback – now receives the channel via closure
// ---------------------------------------------------------------------
var onMessage = func(client mqtt.Client, m mqtt.Message, locChan chan<- repository.Location) {
	var loc repository.Location
	if err := json.Unmarshal(m.Payload(), &loc); err != nil {
		zap.L().Error("invalid mqtt payload", zap.Error(err))
		return
	}
	if loc.VehicleID == "" || loc.Latitude == 0 || loc.Longitude == 0 || loc.Timestamp == 0 {
		zap.L().Error("invalid location data", zap.Any("payload", string(m.Payload())))
		return
	}
	select {
	case locChan <- loc:
	default:
		zap.L().Warn("location channel full", zap.Any("location", loc))
	}
}

// ---------------------------------------------------------------------
//  3. Start – subscribe *after* connect and pass the channel to the
//     callback using a closure.
//
// ---------------------------------------------------------------------
func Start(ctx context.Context, brokerURL string, locChan chan<- repository.Location) error {
	opts := mqtt.NewClientOptions().
		AddBroker(brokerURL).
		SetClientID("fleet-backend-" + fmt.Sprintf("%d", time.Now().UnixNano())).
		SetAutoReconnect(true).
		SetConnectTimeout(5 * time.Second).
		SetOnConnectHandler(func(client mqtt.Client) {
			// wrap the global onMessage so it can see the channel
			handler := func(c mqtt.Client, m mqtt.Message) {
				onMessage(c, m, locChan) // <-- locChan is now in scope
			}

			if token := client.Subscribe("/fleet/vehicle/#", 0, handler); token.Wait() && token.Error() != nil {
				zap.L().Error("MQTT subscribe failed", zap.Error(token.Error()))
			} else {
				zap.L().Info("MQTT subscribed to /fleet/vehicle/#")
			}
		}).
		SetConnectionLostHandler(func(client mqtt.Client, err error) {
			zap.L().Warn("MQTT connection lost", zap.Error(err))
		}).
		SetReconnectingHandler(func(client mqtt.Client, _ *mqtt.ClientOptions) {
			zap.L().Info("MQTT reconnecting...")
		})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	<-ctx.Done()
	client.Disconnect(250)
	return nil
}
