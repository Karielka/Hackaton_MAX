package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

func Connect() *gorm.DB {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		host := getenv("POSTGRES_HOST", "localhost") // в docker-compose это может быть "db"
		port := getenv("POSTGRES_PORT", "5433")
		user := getenv("POSTGRES_USER", "app")
		pass := getenv("POSTGRES_PASSWORD", "app")
		name := getenv("POSTGRES_DB", "app")
		ssl  := getenv("POSTGRES_SSLMODE", "disable")

		dsn = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=%s",
			user, pass, host, port, name, ssl,
		)
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

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}