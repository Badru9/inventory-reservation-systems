package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL     string
	Port            string
	ReservationTTL  time.Duration
	SweeperInterval time.Duration
	ShutdownTimeout time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://indico:indico@127.0.0.1:5432/indico?sslmode=disable"),
		Port:            getEnv("PORT", "8080"),
		ReservationTTL:  getDurationEnv("RESERVATION_TTL", 5*time.Minute),
		SweeperInterval: getDurationEnv("SWEEPER_INTERVAL", 15*time.Second),
		ShutdownTimeout: getDurationEnv("SHUTDOWN_TIMEOUT", 10*time.Second),
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	return cfg, nil
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

func getDurationEnv(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		if secs, err := strconv.Atoi(v); err == nil {
			return time.Duration(secs) * time.Second
		}
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
