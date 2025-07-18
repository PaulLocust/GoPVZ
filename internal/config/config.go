package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env  string     `yaml:"env" env-default:"local"`
	DB   Database   `yaml:"database"`
	REST RESTConfig `yaml:"rest"`
	GRPC GRPCConfig `yaml:"grpc"`
}

type Database struct {
	Host     string `yaml:"host" env-default:"localhost"`
	User     string `yaml:"user" env-default:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name" env-default:"postgres"`
	Port     int    `yaml:"port" env-default:"5432"`
	SSLMode  string `yaml:"ssl_mode" env-default:"disable"`
}

type RESTConfig struct {
	Port      int           `yaml:"port"`
	Timeout   time.Duration `yaml:"timeout"`
	JWTSecret string        `yaml:"jwt_secret"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

func MustLoad() Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return cfg
}
