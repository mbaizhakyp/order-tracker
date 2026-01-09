package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/mbaizhakyp/order-tracker/internal/adapters/event"
	"github.com/mbaizhakyp/order-tracker/internal/adapters/handler"
	"github.com/mbaizhakyp/order-tracker/internal/adapters/repository"
	websocket_internal "github.com/mbaizhakyp/order-tracker/internal/adapters/websocket"
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

	// Dispatch Service Setup
	storeRepo := repository.NewPostgresStoreRepository(dbPool)
	dispatchRepo := repository.NewPostgresDispatchRepository(dbPool)
	dispatchSvc := service.NewDispatchService(storeRepo, locationRepo, dispatchRepo, orderRepo)

	// Start Kafka Consumer
	consumer := event.NewKafkaConsumer(cfg.Kafka.Brokers, "orders.lifecycle", "dispatch-group", dispatchSvc)
	go func() {
		if err := consumer.Start(context.Background()); err != nil {
			log.Printf("Kafka consumer stopped: %v", err)
		}
	}()

	// WebSocket Hub
	wsHub := websocket_internal.NewHub()
	go wsHub.Run()

	// 4. Setup Router
	r := gin.Default()
	v1 := r.Group("/api/v1")
	{
		v1.POST("/orders", orderHandler.CreateOrder)
		v1.GET("/orders/:id", orderHandler.GetOrder)

		v1.POST("/location", locationHandler.UpdateLocation)
	}

	r.GET("/ws", func(c *gin.Context) {
		websocket_internal.ServeWs(wsHub, c)
	})

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
