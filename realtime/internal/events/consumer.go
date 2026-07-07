package events

import (
	"context"
	"strings"

	"github.com/segmentio/kafka-go"
)

type Handler func(ctx context.Context, key, value []byte) error

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers string, topics []string, group string) *Consumer {
	if brokers == "" || len(topics) == 0 {
		return &Consumer{}
	}
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     strings.Split(brokers, ","),
			GroupTopics: topics,
			GroupID:     group,
		}),
	}
}

func (c *Consumer) Run(ctx context.Context, handle Handler) error {
	if c.reader == nil {
		<-ctx.Done()
		return ctx.Err()
	}
	defer c.reader.Close()
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			return err
		}
		_ = handle(ctx, m.Key, m.Value)
		if err := c.reader.CommitMessages(ctx, m); err != nil {
			return err
		}
	}
}
