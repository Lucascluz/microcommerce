package main

import (
	"context"
	"log"
	"time"

	"github.com/lucas/shared/utils"
	"github.com/segmentio/kafka-go"
)

func main() {
	port := utils.GetEnvOrDefault("PORT", "8082")

	// Start Kafka consumer in a goroutine
	go startKafkaConsumer()

	log.Printf("Catalog service starting on port %s", port)
}

func startKafkaConsumer() {
	broker := utils.GetEnvOrDefault("KAFKA_BROKER", "localhost:9092")
	pingTopic := "service-ping"
	pongTopic := "service-pong"
	groupID := "catalog-service-group"

	log.Printf("Starting Kafka consumer. Broker: %s", broker)

	// Wait for Kafka to be available with retry logic
	for retries := 0; retries < 30; retries++ {
		log.Printf("Attempting to connect to Kafka at %s (attempt %d/30)", broker, retries+1)

		// Test connection by creating a temporary consumer
		conn, err := kafka.DialLeader(context.Background(), "tcp", broker, pingTopic, 0)
		if err == nil {
			conn.Close()
			log.Println("Successfully connected to Kafka")
			break
		}

		log.Printf("Failed to connect to Kafka: %v. Retrying in 5 seconds...", err)
		time.Sleep(5 * time.Second)
	}

	// Create a new Kafka reader
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{broker},
		Topic:       pingTopic,
		GroupID:     groupID,
		StartOffset: kafka.LastOffset,
		MinBytes:    1,
		MaxBytes:    10e6,
	})
	defer r.Close()

	// Create new Kafka writer
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{broker},
		Topic:   pongTopic,
	})
	defer w.Close()

	log.Println("Catalog service Kafka consumer started")

	// Read messages from Kafka
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		m, err := r.ReadMessage(ctx)
		cancel()

		if err != nil {
			if err == context.DeadlineExceeded {
				log.Printf("No messages received in 30 seconds, continuing...")
				continue
			}
			log.Printf("Error reading Kafka message: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		log.Printf("Received ping: %s", string(m.Value))

		// Check if the ping is specifically for this service
		if string(m.Key) == "catalog-service" || string(m.Value) == "ping" {
			// Respond with service status
			resp := []byte(`{"status":"healthy", "service":"catalog-service", "timestamp":"` + time.Now().Format(time.RFC3339) + `"}`)
			ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			err = w.WriteMessages(ctx,
				kafka.Message{
					Key:   []byte("catalog-service"),
					Value: resp,
					Time:  time.Now(),
				},
			)
			cancel()

			if err != nil {
				log.Printf("Error writing kafka pong: %v", err)
			} else {
				log.Printf("Sent pong response")
			}
		}
	}
}
