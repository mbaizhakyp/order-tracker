package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/mbaizhakyp/order-tracker/internal/core/service"
	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	reader      *kafka.Reader
	dispatchSvc *service.DispatchService
}

func NewKafkaConsumer(brokers []string, topic string, groupID string, dispatchSvc *service.DispatchService) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &KafkaConsumer{
		reader:      reader,
		dispatchSvc: dispatchSvc,
	}
}

type EventEnvelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type OrderCreatedEvent struct {
	ID      uuid.UUID `json:"id"`
	StoreID uuid.UUID `json:"store_id"`
	// Other fields...
}

func (c *KafkaConsumer) Start(ctx context.Context) error {
	fmt.Println("Starting Kafka Consumer...")
	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			return fmt.Errorf("reader failed: %w", err)
		}

		// Process message
		if err := c.processMessage(ctx, m); err != nil {
			log.Printf("Failed to process message: %v", err)
			continue
		}
	}
}

func (c *KafkaConsumer) processMessage(ctx context.Context, m kafka.Message) error {
	var envelope EventEnvelope
	if err := json.Unmarshal(m.Value, &envelope); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	if envelope.Type == "ORDER_CREATED" {
		var eventData OrderCreatedEvent
		if err := json.Unmarshal(envelope.Data, &eventData); err != nil {
			return fmt.Errorf("failed to unmarshal order data: %w", err)
		}

		log.Printf("[Dispatch] Received ORDER_CREATED for Order %s", eventData.ID)

		// Trigger Dispatch Logic
		if err := c.dispatchSvc.DispatchOrder(ctx, eventData.ID, eventData.StoreID); err != nil {
			return fmt.Errorf("dispatch failed: %w", err)
		}
	}

	return nil
}

func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}
