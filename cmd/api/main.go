package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/mbaizhakyp/order-tracker/internal/adapters/event"
	"github.com/mbaizhakyp/order-tracker/internal/adapters/handler"
	"github.com/mbaizhakyp/order-tracker/internal/adapters/repository"
	"github.com/mbaizhakyp/order-tracker/internal/core/config"
	"github.com/mbaizhakyp/order-tracker/internal/core/service"
)

func main() {
	// 1. Load Config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 2. Setup Database
	dbPool, err := repository.NewDB(cfg.Database.DSN)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer dbPool.Close()

	// 3. Setup Dependencies
	kafkaPublisher := event.NewKafkaEventPublisher(cfg.Kafka.Brokers)

	orderRepo := repository.NewPostgresOrderRepository(dbPool)
	orderSvc := service.NewOrderService(orderRepo, kafkaPublisher)
	orderHandler := handler.NewOrderHandler(orderSvc)

	locationRepo, err := repository.NewRedisLocationRepository(cfg.Redis.Addr)
	if err != nil {
		log.Fatalf("failed to setup redis: %v", err)
	}
	locationSvc := service.NewLocationService(locationRepo)
	locationHandler := handler.NewLocationHandler(locationSvc)

	// 4. Setup Router
	r := gin.Default()
	v1 := r.Group("/api/v1")
	{
		v1.POST("/orders", orderHandler.CreateOrder)
		v1.GET("/orders/:id", orderHandler.GetOrder)

		v1.POST("/location", locationHandler.UpdateLocation)
	}

	// 5. Run Server
	log.Printf("Starting server on %s", cfg.Server.Port)
	if err := r.Run(cfg.Server.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
