package config

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDatabase(config *Config) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		config.DBHost, config.DBUser, config.DBPassword, config.DBName, config.DBPort, config.DBSSLMode,
	)

	maxRetries := 10
	retryInterval := 10 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("Mencoba menghubungkan ke database (percobaan %d/%d)...", attempt, maxRetries)

		var err error
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})

		if err == nil {
			// Connection successful, verify it works
			sqlDB, err := DB.DB()
			if err == nil {
				err = sqlDB.Ping()
				if err == nil {
					log.Println("✓ Database berhasil terhubung")
					return
				}
			}
			log.Printf("✗ Verifikasi koneksi database gagal: %v", err)
		} else {
			log.Printf("✗ Gagal menghubungkan ke database: %v", err)
		}

		if attempt < maxRetries {
			log.Printf("⏳ Menunggu %v sebelum mencoba lagi...", retryInterval)
			time.Sleep(retryInterval)
		}
	}

	log.Fatal("❌ Gagal menghubungkan ke database setelah percobaan maksimum")
}

func GetDB() *gorm.DB {
	return DB
}
