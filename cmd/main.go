package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/tajious/heimdall/internal/api/handlers"
	"github.com/tajious/heimdall/internal/api/router"
	"github.com/tajious/heimdall/internal/config"
	"github.com/tajious/heimdall/internal/middleware"
	"github.com/tajious/heimdall/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	var store storage.Storage
	if cfg.Server.Environment == "development" {
		log.Println("Using in-memory storage for development")
		store = storage.NewInMemoryStorage()
	} else {
		log.Println("Using PostgreSQL storage for production")
		dsn := storage.BuildDSN(cfg.Database)
		store, err = storage.NewPostgresStorage(dsn)
		if err != nil {
			log.Fatalf("Failed to initialize storage: %v", err)
		}
	}

	app := fiber.New(fiber.Config{
		AppName: "Heimdall",
	})

	app.Use(cors.New())
	app.Use(logger.New())

	authHandler := handlers.NewAuthHandler(store, cfg.JWT.Secret, cfg.JWT.AccessExpiration)
	tenantHandler := handlers.NewTenantHandler(store)
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWT.Secret)
	rateLimiter := middleware.NewRateLimiter(middleware.NewMemoryStore(), true)

	apiRouter := router.NewRouter(
		app,
		authHandler,
		tenantHandler,
		authMiddleware,
		rateLimiter,
	)

	apiRouter.SetupRoutes()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
