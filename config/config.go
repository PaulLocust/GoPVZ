package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type (
	Config struct {
		Log   Log
		DB    DB
		HTTP  HTTP
		GRPC  GRPC
		JWT   JWT
		PGURL PGURL
	}

	HTTP struct {
		Port string `env:"HTTP_PORT" envDefault:"8080"`
	}

	GRPC struct {
		Port string `env:"GRPC_PORT" envDefault:"3000"`
	}

	Log struct {
		Level string `env:"LOG_LEVEL,required"`
	}

	DB struct {
		Host     string `env:"DB_HOST,required"`
		User     string `env:"DB_USER,required"`
		Name     string `env:"DB_NAME,required"`
		Port     string `env:"DB_PORT,required"`
		Password string `env:"DB_PASSWORD,required"`
		SSLMode  string `env:"DB_SSL,required"`
	}

	PGURL struct {
		URL string `env:"PG_URL"`
	}

	JWT struct {
		Secret string `env:"JWT_SECRET,required"`
	}
)

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	return cfg, nil
}
