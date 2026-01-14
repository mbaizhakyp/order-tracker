package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/mbaizhakyp/order-tracker/internal/adapters/event"
	"github.com/mbaizhakyp/order-tracker/internal/adapters/handler"
	"github.com/mbaizhakyp/order-tracker/internal/adapters/middleware"
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

	// Store Service Setup (Moved up for OrderService dependency)
	storeRepo := repository.NewPostgresStoreRepository(dbPool)
	storeSvc := service.NewStoreService(storeRepo)
	storeHandler := handler.NewStoreHandler(storeSvc)

	// Dispatch repositories needed later
	dispatchRepo := repository.NewPostgresDispatchRepository(dbPool)

	orderRepo := repository.NewPostgresOrderRepository(dbPool)
	orderSvc := service.NewOrderService(orderRepo, storeRepo, kafkaPublisher)
	orderHandler := handler.NewOrderHandler(orderSvc)

	// Start Auto-Cancellation Background Job
	go orderSvc.AutoCancelLoop(context.Background())

	locationRepo, err := repository.NewRedisLocationRepository(cfg.Redis.Addr)
	if err != nil {
		log.Fatalf("failed to setup redis: %v", err)
	}
	locationSvc := service.NewLocationService(locationRepo, kafkaPublisher)
	locationHandler := handler.NewLocationHandler(locationSvc)

	// Dispatch Service Setup
	dispatchSvc := service.NewDispatchService(storeRepo, locationRepo, dispatchRepo, orderRepo, kafkaPublisher)

	// Auth Service Setup
	userRepo := repository.NewPostgresUserRepository(dbPool)
	authSvc := service.NewAuthService(userRepo, kafkaPublisher, cfg.Auth.JWTSecret)
	authHandler := handler.NewAuthHandler(authSvc)

	// Start Kafka Consumer (Dispatch Logic)
	dispatchConsumer := event.NewKafkaConsumer(cfg.Kafka.Brokers, "orders.lifecycle", "dispatch-group", dispatchSvc)
	go func() {
		if err := dispatchConsumer.Start(context.Background()); err != nil {
			log.Printf("Kafka consumer (dispatch) stopped: %v", err)
		}
	}()

	// WebSocket Hub
	wsHub := websocket_internal.NewHub()
	go wsHub.Run()

	// Start Kafka Consumer (WebSocket Bridge - Offers)
	wsConsumerOffers := event.NewWebSocketConsumer(cfg.Kafka.Brokers, "offers.dispatch", "websocket-group-offers", wsHub)
	go func() {
		if err := wsConsumerOffers.Start(context.Background()); err != nil {
			log.Printf("Kafka consumer (ws-offers) stopped: %v", err)
		}
	}()

	// Start Kafka Consumer (WebSocket Bridge - Tracker)
	wsConsumerTracker := event.NewWebSocketConsumer(cfg.Kafka.Brokers, "tracker", "websocket-group-tracker", wsHub)
	go func() {
		if err := wsConsumerTracker.Start(context.Background()); err != nil {
			log.Printf("Kafka consumer (ws-tracker) stopped: %v", err)
		}
	}()

	// Start Kafka Consumer (WebSocket Bridge - Order Lifecycle)
	// This was missing! Bridging order status events to frontend.
	wsConsumerLifecycle := event.NewWebSocketConsumer(cfg.Kafka.Brokers, "orders.lifecycle", "websocket-group-lifecycle", wsHub)
	go func() {
		if err := wsConsumerLifecycle.Start(context.Background()); err != nil {
			log.Printf("Kafka consumer (ws-lifecycle) stopped: %v", err)
		}
	}()

	// 4. Setup Router
	r := gin.Default()

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	v1 := r.Group("/api/v1")
	{
		// Auth
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Protected Routes
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg.Auth.JWTSecret))
		{
			protected.POST("/auth/presence", authHandler.Presence)
			protected.POST("/orders", orderHandler.CreateOrder)
			protected.GET("/orders/active", orderHandler.GetShopperActiveOrder)
			// Ideally protect these too, but keeping open for simulator ease for now except creation
			// protected.POST("/orders/:id/claim", orderHandler.ClaimOrder)
		}

		// Public Order Routes (for Simulator/Demo access mostly, or read-only)
		orderGroup := v1.Group("/orders")
		{
			// orderGroup.POST("", orderHandler.CreateOrder) // Moved to protected
			orderGroup.GET("/:id", orderHandler.GetOrder)
			orderGroup.POST("/:id/claim", orderHandler.ClaimOrder)
			orderGroup.POST("/:id/arrive", orderHandler.ArriveAtStore)
			orderGroup.POST("/:id/pickup", orderHandler.PickUpOrder)
			orderGroup.POST("/:id/arrive_customer", orderHandler.ArriveAtCustomer)
			orderGroup.POST("/:id/deliver", orderHandler.DeliverOrder)
			orderGroup.POST("/:id/cancel", orderHandler.CancelOrder)
			orderGroup.GET("/:id/history", orderHandler.GetOrderHistory)
		}
		// List all orders for the logged in user
		v1.GET("/orders", orderHandler.GetUserOrders)

		v1.POST("/location", locationHandler.UpdateLocation)

		demoHandler := handler.NewDemoHandler(storeRepo)
		v1.POST("/demo/location", demoHandler.SetLocation)

		v1.GET("/stores", storeHandler.GetStores)
	}

	r.GET("/ws", func(c *gin.Context) {
		websocket_internal.ServeWs(wsHub, c)
	})

	// 4. Setup Router

	// 5. Run Server
	log.Printf("Starting server on %s", cfg.Server.Port)
	if err := r.Run(cfg.Server.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
