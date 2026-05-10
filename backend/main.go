package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/trustwired/internal-t/internal/config"
	appcron "github.com/trustwired/internal-t/internal/cron"
	"github.com/trustwired/internal-t/internal/database"
	"github.com/trustwired/internal-t/internal/handlers"
	"github.com/trustwired/internal-t/internal/routes"
	"github.com/trustwired/internal-t/internal/services"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate on startup (or if --migrate flag passed)
	if len(os.Args) > 1 && os.Args[1] == "--migrate" {
		log.Println("Running database migrations...")
		if err := database.AutoMigrate(db); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("Migrations completed successfully")
		return
	}

	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Auto-migration failed: %v", err)
	}

	// Initialize services
	auditService := services.NewAuditService(db)
	emailService := services.NewEmailService(cfg)
	storageService, err := services.NewStorageService(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	slaService := services.NewSLAService()

	// Initialize handler
	h := handlers.NewHandler(db, cfg, auditService, emailService, storageService, slaService)

	// Start cron
	cronRunner := appcron.StartCron(db, slaService)
	defer cronRunner.Stop()

	// Setup Gin
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	routes.SetupRoutes(r, h, db, cfg)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s (env: %s)", port, cfg.AppEnv)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
