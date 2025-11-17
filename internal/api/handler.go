package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"github.com/sodrety/fleet-tj/internal/repository"
	"go.uber.org/zap"
)

type Payload struct {
	VehicleID string  `json:"vehicle_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longit	ude"`
	Timestamp int64   `json:"timestamp"`
}

func GetLatest(repo *repository.LocationRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		loc, err := repo.GetLatest(c, c.Param("vehicle_id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if loc == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, loc)
	}
}

func GetHistory(repo *repository.LocationRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		start, _ := strconv.ParseInt(c.Query("start"), 10, 64)
		end, _ := strconv.ParseInt(c.Query("end"), 10, 64)
		if start == 0 || end == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "start and end required"})
			return
		}
		locs, err := repo.GetHistory(c, c.Param("vehicle_id"), start, end)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, locs)
	}
}

func LocateVehicle(repo *repository.LocationRepo, logger *zap.Logger, mqttPub mqtt.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read raw request body
		raw, err := c.GetRawData()
		if err != nil {
			c.JSON(400, gin.H{"error": "cannot read body"})
			return
		}

		// (Optional) Log the raw payload
		fmt.Println("RAW:", string(raw))

		var loc repository.Location
		if err := json.Unmarshal(raw, &loc); err != nil {
			c.JSON(400, gin.H{"error": "invalid json"})
			return
		}

		if err := repo.Save(c, loc); err != nil {
			logger.Error("save failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// client := mqtt.NewClient(mqtt.NewClientOptions().
		// 	AddBroker("tcp://localhost:1883").
		// 	SetClientID("mock-" + loc.VehicleID + "-" + fmt.Sprint(time.Now().UnixNano())))
		// client.Connect().Wait()

		payload := Payload{
			VehicleID: loc.VehicleID,
			Latitude:  loc.Latitude,
			Longitude: loc.Longitude,
			Timestamp: loc.Timestamp,
		}

		data, _ := json.Marshal(payload)
		token := mqttPub.Publish("/fleet/vehicle/"+loc.VehicleID+"/location", 0, false, data)
		token.Wait()

		c.JSON(http.StatusOK, "OK")
	}
}
