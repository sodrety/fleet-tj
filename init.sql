CREATE TABLE IF NOT EXISTS vehicle_locations (
    id SERIAL PRIMARY KEY,
    vehicle_id TEXT NOT NULL,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    timestamp BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_vehicle_timestamp
ON vehicle_locations(vehicle_id, timestamp DESC);
