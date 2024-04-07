package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"

	"github.com/kulikvl/http-log-processor/internal/database/entity"
)

func createTable(db *sql.DB) error {
	// Create main http logs table.
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS http_log (
            timestamp           DateTime,
            resource_id         UInt64,
            bytes_sent          UInt64,
			request_time_milli  UInt64,
			response_status     UInt16,
            cache_status        LowCardinality(String),
			method              LowCardinality(String),
			remote_addr         String,
			url                 String
        ) ENGINE = MergeTree() ORDER BY timestamp
    `)

	if err != nil {
		return err
	}

	// Create materialized view that represents aggregated http logs data. It is automatically updated when new data is inserted into the main table.
	_, err = db.Exec(`
		CREATE MATERIALIZED VIEW IF NOT EXISTS http_log_aggregated 
		ENGINE = SummingMergeTree() 
		ORDER BY (resource_id, response_status, cache_status, remote_addr) 
		AS 
		SELECT 
			resource_id, 
			response_status, 
			cache_status, 
			remote_addr,
			sum(bytes_sent) as total_bytes_sent,
			count() as request_count 
		FROM http_log 
		GROUP BY resource_id, response_status, cache_status, remote_addr
    `)

	return err
}

// Inserts logs into database using efficient batch insertion.
func insertLogs(db *sql.DB, logs []entity.HttpLogRecord) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}

	stmt, err := tx.Prepare("INSERT INTO http_log (timestamp, resource_id, bytes_sent, request_time_milli, response_status, cache_status, method, remote_addr, url) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	for _, l := range logs {
		if _, err := stmt.Exec(
			l.Timestamp,
			l.ResourceID,
			l.BytesSent,
			l.RequestTimeMilli,
			l.ResponseStatus,
			l.CacheStatus,
			l.Method,
			l.RemoteAddr,
			l.URL,
		); err != nil {
			return fmt.Errorf("failed to execute statement (insert): %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction (batch insert): %v", err)
	}

	return nil
}

type Writer struct {
	db            *sql.DB
	bufferSize    int
	writeInterval time.Duration
}

func NewWriter(driver, dbAddr string, bufferSize int, writeInterval time.Duration) (*Writer, error) {
	dataSource := fmt.Sprintf("tcp://%s?debug=false", dbAddr)

	db, err := sql.Open(driver, dataSource)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	log.Print("Successfully connected to the database.\n")

	if err := createTable(db); err != nil {
		return nil, fmt.Errorf("failed to create http logs table: %v", err)
	}

	return &Writer{
		db:            db,
		bufferSize:    bufferSize,
		writeInterval: writeInterval,
	}, nil
}

func (w *Writer) Close() error {
	return w.db.Close()
}

func (w *Writer) Write(ctx context.Context, logs <-chan entity.HttpLogRecord) {
	ticker := time.NewTicker(w.writeInterval)
	defer ticker.Stop()

	buffer := make([]entity.HttpLogRecord, 0, w.bufferSize)

	// Goroutine to write logs to the database in batches using retry mechanism with exponential backoff.
	logBatchChannel := make(chan []entity.HttpLogRecord)
	go func() {
		for batch := range logBatchChannel {
			log.Printf("Inserting logs buffer with %d log entries to the database...\n", len(batch))

			for i := 0; true; i++ {
				if err := insertLogs(w.db, batch); err != nil {
					seconds := math.Pow(2, float64(i))
					waitTime := time.Duration(seconds) * time.Second

					log.Printf("Failed to insert logs: %v. Retrying after %f seconds...\n", err, seconds)
					time.Sleep(waitTime)
				} else {
					log.Printf("Logs (%d) were successfully inserted into the database!\n", len(batch))
					break
				}
			}
		}
	}()

	// Copies current logs buffer and sends it to the logBatchChannel.
	flushBuffer := func() {
		if len(buffer) == 0 {
			return
		}

		bufferToSend := make([]entity.HttpLogRecord, len(buffer))
		copy(bufferToSend, buffer)

		select {
		case logBatchChannel <- bufferToSend:
			buffer = buffer[:0]
		case <-ctx.Done():
			return
		default:
			log.Println("Log channel for writing to database is full; will retry on next tick...")
		}
	}

	// Main loop to handle logs and flush buffer.
	for {
		select {
		case l := <-logs:
			buffer = append(buffer, l)
			if len(buffer) >= w.bufferSize {
				log.Printf("Buffer is full (%d), waiting for tick...\n", len(buffer))
				<-ticker.C
				flushBuffer()
			}
		case <-ticker.C:
			flushBuffer()
		case <-ctx.Done():
			return
		}
	}
}
