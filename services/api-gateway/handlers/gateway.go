package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lucas/shared/utils"
	"github.com/segmentio/kafka-go"
)

type GatewayHandler struct {
	kafkaWriter *kafka.Writer
}

func NewGatewayHandler() *GatewayHandler {
	broker := utils.GetEnvOrDefault("KAFKA_BROKER", "kafka:9092")

	return &GatewayHandler{
		kafkaWriter: &kafka.Writer{
			Addr:     kafka.TCP(broker),
			Topic:    "service-ping",
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (h *GatewayHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "api-gateway",
	})
}

func (h *GatewayHandler) PingServices(c *gin.Context) {
	broker := utils.GetEnvOrDefault("KAFKA_BROKER", "kafka:9092")
	services := []string{"catalog-service", "transaction-service", "user-service", "notification-service", "visualization-service"}

	log.Printf("Pinging %d services", len(services))

	// Create a unique consumer group ID for this request to ensure fresh reads
	requestID := time.Now().UnixNano()
	groupID := fmt.Sprintf("api-gateway-ping-%d", requestID)

	// Create reader for this specific request
	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{broker},
		Topic:       "service-pong",
		GroupID:     groupID,
		StartOffset: kafka.LastOffset, // Read only new messages
		MinBytes:    1,
		MaxBytes:    10e6,
		MaxWait:     time.Second * 1,
	})
	defer kafkaReader.Close()

	// Send ping to each service
	for i, svc := range services {
		msg := kafka.Message{
			Key:   []byte(svc),
			Value: []byte("ping"),
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := h.kafkaWriter.WriteMessages(ctx, msg); err != nil {
			log.Printf("Error sending ping to %s: %v", svc, err)
			cancel()
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Failed to communicate with services",
				"details": err.Error(),
			})
			return
		}
		cancel()
		log.Printf("Sent ping to %s", svc)

		// Small delay between messages
		if i < len(services)-1 {
			time.Sleep(50 * time.Millisecond)
		}
	}

	// Wait a moment for services to process and respond
	log.Printf("Waiting 1 second for services to process ping messages...")
	time.Sleep(1 * time.Second)

	// Collect responses
	log.Printf("Starting to collect responses...")
	responses := make([]map[string]interface{}, 0)
	receivedServices := make(map[string]bool)
	timeout := time.After(6 * time.Second) // Reasonable timeout

responseLoop:
	for len(responses) < len(services) {
		select {
		case <-timeout:
			log.Printf("Timeout waiting for service responses. Got %d/%d responses", len(responses), len(services))
			break responseLoop
		default:
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			m, err := kafkaReader.ReadMessage(ctx)
			cancel()

			if err != nil {
				if err != context.DeadlineExceeded {
					log.Printf("Error reading pong: %v", err)
				} else {
					log.Printf("No message received in 1 second, continuing...")
				}
				continue
			}

			log.Printf("Received response from %s: %s", string(m.Key), string(m.Value))
			var resp map[string]interface{}
			if err := json.Unmarshal(m.Value, &resp); err == nil {
				// Check if we already received response from this service
				if serviceName, ok := resp["service"].(string); ok {
					if !receivedServices[serviceName] {
						receivedServices[serviceName] = true
						responses = append(responses, resp)
						log.Printf("Collected response from %s (%d/%d)", serviceName, len(responses), len(services))
					} else {
						log.Printf("Duplicate response from %s, ignoring", serviceName)
					}
				} else {
					log.Printf("Response missing 'service' field: %v", resp)
				}
			} else {
				log.Printf("Failed to unmarshal response: %v", err)
			}
		}
	}

	status := http.StatusOK
	if len(responses) < len(services) {
		status = http.StatusPartialContent
	}

	// Respond with collected service statuses
	c.JSON(status, gin.H{
		"services":            responses,
		"total_services":      len(services),
		"responding_services": len(responses),
		"timestamp":           time.Now().Format(time.RFC3339),
	})
}
