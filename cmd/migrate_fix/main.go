package main

import (
	"context"
	"log"

	"github.com/mbaizhakyp/order-tracker/internal/adapters/repository"
	"github.com/mbaizhakyp/order-tracker/internal/core/config"
)

func main() {
	log.Println("Starting migration fix...")

	// 1. Load Config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 2. Connect to DB
	dbPool, err := repository.NewDB(cfg.Database.DSN)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer dbPool.Close()

	// 3. Execute Migration
	query := `
		ALTER TABLE orders 
		ADD COLUMN IF NOT EXISTS delivery_lat FLOAT DEFAULT 0,
		ADD COLUMN IF NOT EXISTS delivery_lng FLOAT DEFAULT 0;
	`

	_, err = dbPool.Exec(context.Background(), query)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Successfully added delivery_lat and delivery_lng columns to orders table.")
}
