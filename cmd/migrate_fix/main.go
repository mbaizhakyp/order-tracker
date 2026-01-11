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
		ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_status_check;
		ALTER TABLE orders ADD CONSTRAINT orders_status_check CHECK (status IN ('CREATED', 'OFFERED', 'CLAIMED', 'ARRIVED_AT_STORE', 'PICKED_UP', 'ARRIVED_AT_CUSTOMER', 'DELIVERED', 'CANCELLED'));
	`

	_, err = dbPool.Exec(context.Background(), query)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Successfully updated orders_status_check constraint.")
}
