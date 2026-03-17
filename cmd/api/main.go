package main

import (
	"log"

	"tramplin/internal/app"
	"tramplin/internal/config"
)

// @title API Трамплин
// @version 1.0
// @description API платформы Трамплин.
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите JWT в формате: Bearer {token}
func main() {
	cfg := config.Load()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("starting %s on :%s", cfg.AppName, cfg.Port)

	if err := application.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
