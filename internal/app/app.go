package app

import (
	"github.com/gofiber/fiber/v2"

	"tramplin/internal/config"
	"tramplin/internal/repository/memory"
	"tramplin/internal/service"
	httptransport "tramplin/internal/transport/http"
	"tramplin/internal/transport/http/handlers"
)

func New(cfg config.Config) *fiber.App {
	application := fiber.New(fiber.Config{
		AppName: cfg.AppName,
	})

	repo := memory.NewRepository()
	services := service.New(repo)
	httpHandlers := handlers.New(services)

	httptransport.RegisterRoutes(application, httpHandlers)

	return application
}
