package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kulikvl/http-log-processor/internal/capnproto"
	"github.com/kulikvl/http-log-processor/internal/database"
	"github.com/kulikvl/http-log-processor/internal/database/entity"
	"github.com/kulikvl/http-log-processor/internal/kafka"
	"github.com/kulikvl/http-log-processor/internal/mapper"
	"github.com/kulikvl/http-log-processor/internal/model"
	"github.com/kulikvl/http-log-processor/internal/utils"
)

var (
	broker           = utils.GetEnv("KAFKA_BROKER", "localhost:9092")
	topic            = utils.GetEnv("KAFKA_TOPIC", "http_log")
	group            = utils.GetEnv("KAFKA_GROUP", "data-engineering-task-reader")
	db               = utils.GetEnv("DB", "localhost:9000")
	dbDriver         = utils.GetEnv("DB_DRIVER", "clickhouse")
	writerBufferSize = utils.GetEnvAsInt("WRITER_BUFFER_SIZE", 1024)                                // RAM buffer size for logs
	writeInterval    = time.Duration(utils.GetEnvAsInt("WRITE_INTERVAL_SECONDS", 10)) * time.Second // insert logs every 10 seconds
)

func main() {
	// Gracefully handle SIGINT and SIGTERM
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Kafka logs channel, what we consume
	kafkaLogs := make(chan model.HttpLogRecord)
	// ClickHouse logs channel, what we write
	dbLogs := make(chan entity.HttpLogRecord)

	// Transformer goroutine, that maps (anonymize IPs, map data types, ...) Kafka logs to ClickHouse logs
	go func() {
		for l := range kafkaLogs {
			dbLogs <- mapper.ToDb(l)
		}
	}()

	// Context to stop r/w loop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // ensure cancel is called on main function exit

	// Setup and run ClickHouse writer
	writer, err := database.NewWriter(dbDriver, db, writerBufferSize, writeInterval)
	if err != nil {
		log.Fatalf("failed to create ClickHouse writer: %v\n", err)
	}
	defer func() {
		if err := writer.Close(); err != nil {
			log.Printf("failed to close ClickHouse writer: %v\n", err)
		} else {
			log.Println("ClickHouse writer closed successfully.")
		}
	}()
	go writer.Write(ctx, dbLogs)

	// Setup and run Kafka consumer
	consumer := kafka.NewConsumer(broker, topic, group)
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Printf("failed to close Kafka consumer: %v", err)
		} else {
			log.Println("Kafka consumer closed successfully.")
		}
	}()
	go consumer.Consume(ctx, kafkaLogs, capnproto.Decode)

	<-signals
	cancel()
	log.Println("Received shutdown signal, initiating shutdown...")
}
