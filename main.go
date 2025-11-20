package main

import (
	"fmt"
	"log"

	"livo-backend-2.0/config"
	"livo-backend-2.0/controllers"
	_ "livo-backend-2.0/docs" // This is required for Swagger
	"livo-backend-2.0/migrations"
	"livo-backend-2.0/routes"
)

// @title Livotech Backend Service
// @version 2.0
// @description Comprehensive backend service for Livotech platform with JWT authentication and role-based access control. Authentication: This endpoint uses Bearer token authentication. Include your JWT token in the Authorization header in the format: Bearer your-access-token
// @contact.name Saya
// @contact.email support@livotech.com
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT.
func main() {
	log.Println("🚀 Starting Livotech Backend Service...")

	// Load configuration
	log.Println("📝 Loading configuration...")
	cfg := config.LoadConfig()
	log.Println("✓ Configuration loaded successfully")

	// Connect to database with retry logic
	log.Println("🔌 Connecting to database...")
	config.ConnectDatabase(cfg)

	// Run migrations
	log.Println("🔄 Running database migrations...")
	db := config.GetDB()
	migrations.AutoMigrate(db) // No error handling needed, it's handled inside the function

	// Initialize controllers
	log.Println("🎮 Initializing controllers...")
	authController := controllers.NewAuthController(db, cfg)
	userManagerController := controllers.NewUserManagerController(db)
	boxController := controllers.NewBoxController(db)
	channelController := controllers.NewChannelController(db)
	expeditionController := controllers.NewExpeditionController(db)
	storeController := controllers.NewStoreController(db)
	orderController := controllers.NewOrderController(db)
	log.Println("✓ Controllers initialized successfully")

	// Setup routes
	log.Println("🛣️  Setting up routes...")
	router := routes.SetupRoutes(cfg, authController, userManagerController, boxController, channelController, expeditionController, storeController, orderController)
	log.Println("✓ Routes configured successfully")

	// Build API URL from config
	apiURL := fmt.Sprintf("http://%s:%s", cfg.APIHost, cfg.Port)

	// Start server
	log.Println("════════════════════════════════════════════════════════════")
	log.Printf("✓ Server sudah berjalan di port %s", cfg.Port)
	log.Printf("📊 Cek kesehatan: %s/health", apiURL)
	log.Printf("📚 Dokumentasi API: %s/docs", apiURL)
	log.Printf("📖 Swagger UI: %s/swagger/index.html", apiURL)
	log.Println("════════════════════════════════════════════════════════════")

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal("❌ Server Initialization Failed:", err)
	}
}
