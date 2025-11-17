package config

import "github.com/spf13/viper"

type Geofence struct {
	Center struct {
		Latitude  float64 `mapstructure:"latitude"`
		Longitude float64 `mapstructure:"longitude"`
	} `mapstructure:"center"`
	RadiusMeters float64 `mapstructure:"radius_meters"`
}

type RootConfig struct {
	Geofence Geofence `mapstructure:"geofence"`
}

func LoadGeofence() (*Geofence, error) {
	viper.SetConfigFile("config/geofence.yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg RootConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg.Geofence, nil
}
