package main

import (
	"log"

	"tramplin/internal/app"
	"tramplin/internal/config"
)

func main() {
	cfg := config.Load()

	application := app.New(cfg)

	log.Printf("starting %s on :%s", cfg.AppName, cfg.Port)

	if err := application.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
