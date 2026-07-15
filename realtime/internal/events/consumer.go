package events

import (
	"context"
	"net"
	"strconv"
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
	brokerList := strings.Split(brokers, ",")
	ensureTopics(brokerList, topics)
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     brokerList,
			GroupTopics: topics,
			GroupID:     group,
		}),
	}
}

func ensureTopics(brokers, topics []string) {
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return
	}
	defer conn.Close()
	controller, err := conn.Controller()
	if err != nil {
		return
	}
	ctrl, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return
	}
	defer ctrl.Close()
	configs := make([]kafka.TopicConfig, 0, len(topics))
	for _, t := range topics {
		configs = append(configs, kafka.TopicConfig{Topic: t, NumPartitions: 1, ReplicationFactor: 1})
	}
	_ = ctrl.CreateTopics(configs...)
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
