package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lucas/shared/utils"
	"github.com/segmentio/kafka-go"
)

type GatewayHandler struct {
	kafkaWriter *kafka.Writer
	kafkaReader *kafka.Reader
}

func NewGatewayHandler() *GatewayHandler {
	broker := utils.GetEnvOrDefault("KAFKA_BROKER", "localhost:9092")
	return &GatewayHandler{
		kafkaWriter: &kafka.Writer{
			Addr:     kafka.TCP(broker),
			Topic:    "service-ping",
			Balancer: &kafka.LeastBytes{},
		},
		kafkaReader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     []string{broker},
			Topic:       "service-pong",
			GroupID:     "api-gateway-group",
			StartOffset: kafka.FirstOffset,
		}),
	}
}

func (h *GatewayHandler) PingServices(c *gin.Context) {
	services := []string{"catalog-service", "transaction-service", "user-service", "notifications-service", "visualization-service"} // Updated service list

	log.Printf("Pinging %d services", len(services))

	// Send ping to each service
	for _, svc := range services {
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
	}

	// Collect responses with longer timeout
	responses := make([]map[string]interface{}, 0)
	receivedServices := make(map[string]bool)
	timeout := time.After(15 * time.Second)

responseLoop:
	for len(responses) < len(services) {
		select {
		case <-timeout:
			log.Printf("Timeout waiting for service responses. Got %d/%d responses", len(responses), len(services))
			break responseLoop
		default:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			m, err := h.kafkaReader.ReadMessage(ctx)
			cancel()

			if err != nil {
				if err != context.DeadlineExceeded {
					log.Printf("Error reading pong: %v", err)
				}
				continue
			}

			log.Printf("Received response: %s", string(m.Value))
			var resp map[string]interface{}
			if err := json.Unmarshal(m.Value, &resp); err == nil {
				// Check if we already received response from this service
				if serviceName, ok := resp["service"].(string); ok {
					if !receivedServices[serviceName] {
						receivedServices[serviceName] = true
						responses = append(responses, resp)
					}
				}
			}
		}
	}

	// Respond with collected service statuses
	c.JSON(http.StatusOK, gin.H{
		"services":            responses,
		"total_services":      len(services),
		"responding_services": len(responses),
	})
}
