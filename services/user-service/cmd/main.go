package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lucas/shared/utils"
	"github.com/segmentio/kafka-go"
)

func main() {
	port := utils.GetEnvOrDefault("PORT", "8083")

	// Start Kafka consumer in a goroutine
	go startKafkaConsumer()

	// Create HTTP server for health checks
	r := gin.Default()

	// Health check endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "user-service",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Health check endpoint (alternative path)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "user-service",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	log.Printf("User service starting on port %s", port)

	// Start HTTP server (this will block and keep the service running)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func startKafkaConsumer() {
	broker := utils.GetEnvOrDefault("KAFKA_BROKER", "kafka:9092")
	pingTopic := "service-ping"
	pongTopic := "service-pong"
	groupID := "user-service-group"

	log.Printf("Starting Kafka consumer. Broker: %s", broker)

	// Wait for Kafka to be available with retry logic
	var connected bool
	for retries := 0; retries < 30; retries++ {
		log.Printf("Attempting to connect to Kafka at %s (attempt %d/30)", broker, retries+1)

		// Test connection by creating a temporary consumer
		conn, err := kafka.DialLeader(context.Background(), "tcp", broker, pingTopic, 0)
		if err == nil {
			conn.Close()
			log.Println("Successfully connected to Kafka")
			connected = true
			break
		}

		log.Printf("Failed to connect to Kafka: %v. Retrying in 5 seconds...", err)
		time.Sleep(5 * time.Second)
	}

	if !connected {
		log.Printf("Failed to connect to Kafka after 30 attempts. Running without Kafka consumer.")
		return
	}

	// Create a new Kafka reader
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{broker},
		Topic:       pingTopic,
		GroupID:     groupID,
		StartOffset: kafka.FirstOffset,
		MinBytes:    1,
		MaxBytes:    10e6,
		MaxWait:     time.Second * 5, // Don't wait too long for messages
	})
	defer r.Close()

	// Create new Kafka writer
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{broker},
		Topic:    pongTopic,
		Balancer: &kafka.LeastBytes{}, // Better load balancing
	})
	defer w.Close()

	log.Println("User service Kafka consumer started")

	// Read messages from Kafka
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		m, err := r.ReadMessage(ctx)
		cancel()

		if err != nil {
			if err == context.DeadlineExceeded {
				// This is normal - no messages received, continue quietly
				continue
			}
			log.Printf("Error reading Kafka message: %v", err)
			time.Sleep(5 * time.Second) // Wait longer on error
			continue
		}

		log.Printf("Received ping: %s", string(m.Value))

		// Check if the ping is specifically for this service
		if string(m.Key) == "user-service" || string(m.Value) == "ping" {
			// Respond with service status
			resp := []byte(`{"status":"healthy", "service":"user-service", "timestamp":"` + time.Now().Format(time.RFC3339) + `"}`)
			ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			err = w.WriteMessages(ctx,
				kafka.Message{
					Key:   []byte("user-service"),
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
