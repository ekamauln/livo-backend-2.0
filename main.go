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
// @description Layanan backend manajemen pengguna yang komprehensif dengan autentikasi JWT dan kontrol akses berbasis role
// @contact.name Saya Livotech Support
// @contact.email support@livotech.com
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Ketik "Bearer" diikuti oleh spasi dan token JWT.
func main() {
	log.Println("🚀 Memulai Livotech Backend Service...")

	// Load configuration
	log.Println("📝 Memuat konfigurasi...")
	cfg := config.LoadConfig()
	log.Println("✓ Konfigurasi berhasil dimuat")

	// Connect to database with retry logic
	log.Println("🔌 Menghubungkan ke database...")
	config.ConnectDatabase(cfg)

	// Run migrations
	log.Println("🔄 Menjalankan migrasi database...")
	db := config.GetDB()
	migrations.AutoMigrate(db) // No error handling needed, it's handled inside the function

	// Initialize controllers
	log.Println("🎮 Menginisialisasi controller...")
	authController := controllers.NewAuthController(db, cfg)
	log.Println("✓ Berhasil memuat controller")

	// Setup routes
	log.Println("🛣️  Menyiapkan rute...")
	router := routes.SetupRoutes(cfg, authController)
	log.Println("✓ Rute berhasil dikonfigurasi")

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
		log.Fatal("❌ Gagal memulai server:", err)
	}
}
