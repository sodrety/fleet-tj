package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LocationRepo struct {
	pool *pgxpool.Pool
}

type Location struct {
	VehicleID string  `json:"vehicle_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64   `json:"timestamp"`
}

func NewLocationRepo(pool *pgxpool.Pool) *LocationRepo {
	return &LocationRepo{
		pool: pool,
	}
}

func (r *LocationRepo) Save(ctx context.Context, loc Location) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO vehicle_locations (vehicle_id, latitude, longitude, timestamp)
		VALUES ($1, $2, $3, $4)`,
		loc.VehicleID, loc.Latitude, loc.Longitude, loc.Timestamp)
	return err
}

func (r *LocationRepo) GetLatest(ctx context.Context, vehicleId string) (*Location, error) {
	var loc Location
	err := r.pool.QueryRow(ctx,
		`SELECT vehicle_id, latitude, longitude, timestamp
		FROM vehicle_locations
		WHERE vehicle_id = $1
		ORDER BY timestamp DESC
		LIMIT 1`,
		vehicleId).Scan(&loc.VehicleID, &loc.Latitude, &loc.Longitude, &loc.Timestamp)
	if err == pgx.ErrNoRows {
		return nil, err
	}
	return &loc, nil
}

func (r *LocationRepo) GetHistory(ctx context.Context, vehicleId string, start, end int64) ([]Location, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT vehicle_id, latitude, longitude, timestamp
		FROM vehicle_locations
		WHERE vehicle_id = $1 AND timestamp BETWEEN $2 AND $3
		ORDER BY timestamp ASC`,
		vehicleId, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []Location
	for rows.Next() {
		var loc Location
		if err := rows.Scan(&loc.VehicleID, &loc.Latitude, &loc.Longitude, &loc.Timestamp); err != nil {
			return nil, err
		}
		locations = append(locations, loc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return locations, nil
}
