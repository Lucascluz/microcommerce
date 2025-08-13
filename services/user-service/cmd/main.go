package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lucas/shared/database"
	"github.com/lucas/shared/utils"
	"github.com/lucas/user-service/internal/handlers"
	"github.com/lucas/user-service/internal/repository"
	"github.com/lucas/user-service/internal/services"
	"github.com/segmentio/kafka-go"
)

func main() {
	// 1. Initialize database
	if err := initDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 2. Set up dependencies
	userRepo := repository.NewUserRepository(database.GetDB())
	userService := services.NewUserService(userRepo)

	// 3. Set up Kafka
	kafkaWriter := &kafka.Writer{
		Addr:     kafka.TCP(utils.GetEnvOrDefault("KAFKA_BROKER", "kafka:9092")),
		Topic:    "user-responses",
		Balancer: &kafka.LeastBytes{},
	}
	defer kafkaWriter.Close()

	// Kafka writer for health check responses
	healthWriter := &kafka.Writer{
		Addr:     kafka.TCP(utils.GetEnvOrDefault("KAFKA_BROKER", "kafka:9092")),
		Topic:    "service-pong",
		Balancer: &kafka.LeastBytes{},
	}
	defer healthWriter.Close()

	kafkaHandler := handlers.NewKafkaHandler(userService, kafkaWriter)

	// 4. Start Kafka consumers
	go startUserRequestsConsumer(kafkaHandler)
	go startHealthCheckConsumer(healthWriter)

	// 5. Start HTTP server for health checks
	startHTTPServer()
}

func initDatabase() error {
	config := database.GetPostgreSQLConfig()
	if err := database.ConnectPostgreSQL(config); err != nil {
		return err
	}

	// Run migrations
	if err := database.RunMigrations("./migrations"); err != nil {
		return err
	}

	return nil
}

func startUserRequestsConsumer(handler *handlers.KafkaHandler) {
	broker := utils.GetEnvOrDefault("KAFKA_BROKER", "kafka:9092")

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{broker},
		Topic:       "user-requests",
		GroupID:     "user-service-group",
		StartOffset: kafka.LastOffset,
	})
	defer reader.Close()

	log.Println("User requests consumer started")

	for {
		message, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}

		handler.HandleUserMessage(message)
	}
}

func startHTTPServer() {
	port := utils.GetEnvOrDefault("PORT", "8083")
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "user-service",
		})
	})

	log.Printf("User service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func startHealthCheckConsumer(writer *kafka.Writer) {
	broker := utils.GetEnvOrDefault("KAFKA_BROKER", "kafka:9092")

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{broker},
		Topic:       "service-ping",
		GroupID:     "user-service-health-group",
		StartOffset: kafka.LastOffset,
	})
	defer reader.Close()

	log.Println("Health check consumer started")

	for {
		message, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading health check message: %v", err)
			continue
		}

		// Respond to health check pings
		if string(message.Key) == "user-service" || string(message.Value) == "ping" {
			response := map[string]interface{}{
				"status":    "healthy",
				"service":   "user-service",
				"timestamp": time.Now().Format(time.RFC3339),
			}

			responseBytes, _ := json.Marshal(response)

			err = writer.WriteMessages(context.Background(),
				kafka.Message{
					Key:   []byte("user-service"),
					Value: responseBytes,
				},
			)

			if err != nil {
				log.Printf("Failed to send health check response: %v", err)
			} else {
				log.Printf("Sent health check response")
			}
		}
	}
}
