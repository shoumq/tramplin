package app

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
	_ "tramplin/docs"

	"tramplin/internal/authjwt"
	"tramplin/internal/config"
	"tramplin/internal/database"
	"tramplin/internal/repository/postgres"
	"tramplin/internal/service"
	"tramplin/internal/storage"
	miniostorage "tramplin/internal/storage/minio"
	httptransport "tramplin/internal/transport/http"
	"tramplin/internal/transport/http/handlers"
)

func New(cfg config.Config) (*fiber.App, error) {
	ctx := context.Background()
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	if err := database.WaitForDB(ctx, db, 20, 2*time.Second); err != nil {
		return nil, fmt.Errorf("wait for database: %w", err)
	}
	if err := database.RunMigrations(ctx, db, cfg.MigrationsDir); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	application := fiber.New(fiber.Config{
		AppName: cfg.AppName,
	})

	repo, err := postgres.NewRepository(ctx, cfg.DatabaseURL, cfg.S3PublicURL, cfg.S3Bucket)
	if err != nil {
		return nil, fmt.Errorf("create repository: %w", err)
	}
	var objectStorage storage.Storage = storage.NoopStorage{}
	if cfg.S3Endpoint != "" && cfg.S3Bucket != "" {
		objectStorage, err = miniostorage.New(ctx, miniostorage.Config{
			Endpoint:      cfg.S3Endpoint,
			AccessKey:     cfg.S3AccessKey,
			SecretKey:     cfg.S3SecretKey,
			UseSSL:        cfg.S3UseSSL,
			Bucket:        cfg.S3Bucket,
			PublicBaseURL: cfg.S3PublicURL,
		})
		if err != nil {
			return nil, fmt.Errorf("create object storage: %w", err)
		}
	}

	jwtManager := authjwt.New(cfg.JWTSecret, cfg.JWTTTL)

	services := service.New(repo, objectStorage, jwtManager)
	httpHandlers := handlers.New(services, jwtManager)

	httptransport.RegisterRoutes(application, httpHandlers, jwtManager)

	return application, nil
}
