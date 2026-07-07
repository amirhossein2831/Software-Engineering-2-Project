package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

type Publisher struct {
	writer *kafka.Writer
}

func NewPublisher(brokers string) *Publisher {
	if brokers == "" {
		return &Publisher{}
	}
	return &Publisher{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(brokers),
			Balancer:               &kafka.Hash{},
			Async:                  true,
			AllowAutoTopicCreation: true,
		},
	}
}

func (p *Publisher) Publish(ctx context.Context, topic, key string, payload any) error {
	if p.writer == nil {
		return nil
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: b,
		Time:  time.Now(),
	})
}

func (p *Publisher) Close() error {
	if p.writer == nil {
		return nil
	}
	return p.writer.Close()
}
