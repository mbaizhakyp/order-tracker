package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mbaizhakyp/order-tracker/internal/adapters/websocket"
	"github.com/mbaizhakyp/order-tracker/internal/core/entity"
	kafka "github.com/segmentio/kafka-go"
)

type WebSocketConsumer struct {
	reader *kafka.Reader
	hub    *websocket.Hub
}

func NewWebSocketConsumer(brokers []string, topic string, groupID string, hub *websocket.Hub) *WebSocketConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &WebSocketConsumer{
		reader: reader,
		hub:    hub,
	}
}

func (c *WebSocketConsumer) Start(ctx context.Context) error {
	fmt.Println("Starting WebSocket Consumer (offers.dispatch)...")
	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			return fmt.Errorf("reader failed: %w", err)
		}

		// Process message
		if err := c.processMessage(ctx, m); err != nil {
			log.Printf("Failed to process dispatch message: %v", err)
			continue
		}
	}
}

func (c *WebSocketConsumer) processMessage(ctx context.Context, m kafka.Message) error {
	var envelope EventEnvelope
	if err := json.Unmarshal(m.Value, &envelope); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	if envelope.Type == "OFFER_CREATED" {
		var offer entity.DispatchOffer
		if err := json.Unmarshal(envelope.Data, &offer); err != nil {
			return fmt.Errorf("failed to unmarshal offer data: %w", err)
		}

		log.Printf("[WS-Consumer] Received OFFER_CREATED for Shopper %s", offer.ShopperID)

		// Prepare Map to merge Offer fields + Payload (coords)
		finalPayload := make(map[string]interface{})

		// 1. Add Offer fields (using JSON struct tags would be cleaner, but manual map is fine for now)
		// Or just re-marshal offer to map
		offerBytes, _ := json.Marshal(offer)
		_ = json.Unmarshal(offerBytes, &finalPayload)

		// 2. Add Payload fields (Coords)
		if len(envelope.Payload) > 0 {
			var extraData map[string]interface{}
			if err := json.Unmarshal(envelope.Payload, &extraData); err == nil {
				for k, v := range extraData {
					finalPayload[k] = v
				}
			}
		}

		// Push to Shopper via Hub
		msg := map[string]interface{}{
			"type":    "NEW_OFFER",
			"payload": finalPayload,
		}
		c.hub.SendToUser(offer.ShopperID.String(), msg)
	} else if envelope.Type == "SHOPPER_MOVED" {
		// Just pass the whole thing to frontend
		// Data is { "shopper_id": ..., "lat": ..., "lng": ... }
		var locationData map[string]interface{}
		if err := json.Unmarshal(envelope.Data, &locationData); err != nil {
			return fmt.Errorf("failed to unmarshal location data: %w", err)
		}

		msg := map[string]interface{}{
			"type":    "SHOPPER_MOVED",
			"payload": locationData,
		}
		c.hub.Broadcast(msg)
	} else {
		// Generic handling for Order Lifecycle events
		// (ORDER_CLAIMED, ORDER_ARRIVED_AT_STORE, ORDER_PICKED_UP, ORDER_ARRIVED_AT_CUSTOMER, ORDER_DELIVERED)
		var orderData map[string]interface{}
		if err := json.Unmarshal(envelope.Data, &orderData); err != nil {
			// It might not be an order event, just log warn and ignore
			log.Printf("Warn: could not unmarshal data for type %s", envelope.Type)
			return nil
		}

		msg := map[string]interface{}{
			"type": envelope.Type,
			"data": orderData,
		}

		// Broadcast to all (simplest for MVP to ensure Customer and Courier getting updates)
		c.hub.Broadcast(msg)
	}

	return nil
}
