package kafka

import (
	"context"
	"errors"
	"log"

	"github.com/kulikvl/http-log-processor/internal/model"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(broker string, topic, group string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:   []string{broker},
			Topic:     topic,
			GroupID:   group,
			Partition: 0,
		}),
	}
}

func (c *Consumer) Consume(ctx context.Context, logs chan<- model.HttpLogRecord, decode func([]byte) (model.HttpLogRecord, error)) {
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				log.Println("Context cancelled, stopping consume loop...")
				return
			}

			log.Printf("Error while consuming message: %v\n", err)
			continue
		}

		decodedMsg, err := decode(msg.Value)
		if err != nil {
			log.Printf("Error while decoding message: %v\n", err)
			continue
		}

		logs <- decodedMsg
	}

}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
