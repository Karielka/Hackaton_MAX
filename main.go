package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("env DATABASE_URL is required, e.g. postgres://user:pass@localhost:5432/db?sslmode=disable")
	}

	// Поднимаем пул подключений
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("parse DSN: %v", err)
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	// Ручка /healthz — пингует БД с таймаутом
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		// Быстрый ping: выполняем простой запрос
		if err := pool.Ping(ctx); err != nil {
			http.Error(w, fmt.Sprintf(`{"ok":false,"error":"%v"}`, err), http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true,"db":"up"}`))
	})

	addr := ":" + getenvDefault("PORT", "8080")
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func getenvDefault(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

