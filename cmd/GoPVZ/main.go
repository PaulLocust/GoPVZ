package main

import (
	"GoPVZ/config"
	"GoPVZ/internal/app"
	"log"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
