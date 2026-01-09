package ports

import "context"

type EventPublisher interface {
	Publish(ctx context.Context, topic string, key string, payload interface{}) error
}
