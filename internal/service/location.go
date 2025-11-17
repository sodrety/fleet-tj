package service

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sodrety/fleet-tj/internal/publisher"
	"github.com/sodrety/fleet-tj/internal/repository"
	"github.com/sodrety/fleet-tj/pkg/config"
	"go.uber.org/zap"
)

type LocationService struct {
	repo     *repository.LocationRepo
	rabbit   *amqp.Connection
	geofence *config.Geofence
	logger   *zap.Logger
}

func NewLocationService(repo *repository.LocationRepo, rabbit *amqp.Connection, gf *config.Geofence, logger *zap.Logger) *LocationService {
	return &LocationService{
		repo:     repo,
		rabbit:   rabbit,
		geofence: gf,
		logger:   logger,
	}
}

func (s *LocationService) Process(ctx context.Context, loc repository.Location) {
	if err := s.repo.Save(ctx, loc); err != nil {
		s.logger.Error("save failed", zap.Error(err))
	}

	center := s.geofence.Center
	s.logger.Info("geofence", zap.Any("center", s.geofence))
	if IsInGeofence(loc.Latitude, loc.Longitude, center.Latitude, center.Longitude, s.geofence.RadiusMeters) {
		event := publisher.GeofenceEvent{
			VehicleID: loc.VehicleID,
			Event:     "geofence_entry",
			Timestamp: loc.Timestamp,
		}
		event.Location.Latitude = loc.Latitude
		event.Location.Longitude = loc.Longitude

		if err := publisher.PublishGeofence(ctx, s.rabbit, event); err != nil {
			s.logger.Error("publish failed", zap.Error(err))
		} else {
			s.logger.Info("geofence event published", zap.String("vehicle", loc.VehicleID))
		}
	}
}
