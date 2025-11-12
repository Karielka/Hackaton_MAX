package db

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

// Connect открывает соединение к Postgres по DATABASE_URL
func Connect() *gorm.DB {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("env DATABASE_URL is required, e.g. postgres://user:pass@host:5432/db?sslmode=disable")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: glogger.Default.LogMode(glogger.Warn),
	})
	if err != nil {
		log.Fatalf("gorm open: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("db.DB(): %v", err)
	}

	// Пул соединений
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(45 * time.Minute)

	return db
}
