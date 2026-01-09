package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/mbaizhakyp/order-tracker/internal/core/ports"
	"github.com/segmentio/kafka-go"
)

type KafkaEventPublisher struct {
	writer *kafka.Writer
}

func NewKafkaEventPublisher(brokers []string) ports.EventPublisher {
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond, // Send quickly for this demo
	}
	return &KafkaEventPublisher{writer: w}
}

func (p *KafkaEventPublisher) Publish(ctx context.Context, topic string, key string, payload interface{}) error {
	value, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: value,
		Time:  time.Now(),
	}

	// Create topic if not exists (auto-create is often on, but good to be explicit or handle errors)
	// kafka-go writer handles topic auto-creation if configured on broker

	err = p.writer.WriteMessages(ctx, msg)
	if err != nil {
		log.Printf("Failed to write to kafka: %v", err)
		return err
	}

	return nil
}

func (p *KafkaEventPublisher) Close() error {
	return p.writer.Close()
}
