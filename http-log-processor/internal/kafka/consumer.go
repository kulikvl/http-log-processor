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
			// Logger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			// 	fmt.Printf("reader log: "+msg+"\n", args...)
			// }),
			// ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			// 	fmt.Printf("reader error: "+msg+"\n", args...)
			// }),
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

		// fmt.Printf("Received message at topic %s, partition %d, offset %d, size of key %d, size of msg %d\n",
		// 	msg.Topic, msg.Partition, msg.Offset, len(msg.Key), len(msg.Value))

		decodedMsg, err := decode(msg.Value)
		if err != nil {
			log.Printf("Error while decoding message: %v\n", err)
			continue
		}

		// fmt.Printf("Decoded message: %v\n", decodedMsg)

		logs <- decodedMsg
	}

}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
