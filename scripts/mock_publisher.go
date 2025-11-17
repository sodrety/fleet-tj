package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Payload struct {
	VehicleID string  `json:"vehicle_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64   `json:"timestamp"`
}

func main() {
	vehicle := flag.String("vehicle", "B1234XYZ", "vehicle ID")
	flag.Parse()

	client := mqtt.NewClient(mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1883").
		SetClientID("mock-" + *vehicle))
	client.Connect().Wait()

	ticker := time.NewTicker(2 * time.Second)
	lat, lon := -6.2549, 106.7183
	for {
		<-ticker.C
		payload := Payload{
			VehicleID: *vehicle,
			Latitude:  lat + rand.Float64()*0.0005,
			Longitude: lon + rand.Float64()*0.0005,
			Timestamp: time.Now().Unix(),
		}
		data, _ := json.Marshal(payload)
		client.Publish("/fleet/vehicle/"+*vehicle+"/location", 0, false, data)
		fmt.Println("Published location update for vehicle", payload)
	}
}
