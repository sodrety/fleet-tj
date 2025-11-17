package mqtt

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var Publisher mqtt.Client

func InitPublisher(brokerURL string) error {
	opts := mqtt.NewClientOptions().
		AddBroker(brokerURL).
		SetClientID("fleet-publisher-" + fmt.Sprint(time.Now().UnixNano())).
		SetAutoReconnect(true).
		SetConnectTimeout(5 * time.Second)

	Publisher = mqtt.NewClient(opts)

	if token := Publisher.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}
