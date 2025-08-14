package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lucas/shared/models"
	"github.com/lucas/shared/utils"
	"github.com/segmentio/kafka-go"
)

type UserHandler struct {
	kafkaWriter *kafka.Writer
	kafkaReader *kafka.Reader
}

func NewUserHandler() *UserHandler {
	broker := utils.GetEnvOrDefault("KAFKA_BROKER", "kafka:9092")

	log.Printf("Initializing UserHandler with Kafka broker: %s", broker)

	return &UserHandler{
		kafkaWriter: &kafka.Writer{
			Addr:     kafka.TCP(broker),
			Topic:    "user-requests",
			Balancer: &kafka.LeastBytes{},
		},
		kafkaReader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     []string{broker},
			Topic:       "user-responses",
			GroupID:     "api-gateway-user-group",
			StartOffset: kafka.FirstOffset, // Changed to FirstOffset to ensure we don't miss messages
		}),
	}
}

func (u *UserHandler) Register(c *gin.Context) {

	log.Printf("Register request received")

	// 1. Validate request structure (Gateway responsibility)
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// 2. Basic validation (Gateway responsibility)
	if req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and password are required"})
		return
	}

	// 3. Create message with correlation ID for response tracking
	correlationID := uuid.New().String()
	message := models.UserServiceMessage{
		CorrelationID: correlationID,
		Action:        "register",
		Data:          req,
		Timestamp:     time.Now(),
	}

	// 4. Send message to user-service (Gateway responsibility)
	messageBytes, _ := json.Marshal(message)
	err := u.kafkaWriter.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(correlationID),
			Value: messageBytes,
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
		return
	}

	log.Printf("Register request sent via kafka: %s", string(messageBytes))

	// 5. Await response (Gateway responsibility)
	response, err := u.waitForResponse(correlationID, 2*time.Minute)
	if err != nil {
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Request timeout"})
		return
	}

	// 6. Forward response (Gateway responsibility)
	c.JSON(response.StatusCode, response.Data)

	log.Printf("Register response received: %s", response.Data)
}

func (u *UserHandler) Login(c *gin.Context) {

	log.Printf("Login request received")

	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and password are required"})
		return
	}

	correlationID := uuid.New().String()
	message := models.UserServiceMessage{
		CorrelationID: correlationID,
		Action:        "login",
		Data:          req,
		Timestamp:     time.Now(),
	}

	log.Printf("Sending login request message via kafka")
	messageBytes, _ := json.Marshal(message)
	err := u.kafkaWriter.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(correlationID),
			Value: messageBytes,
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
		return
	}

	log.Printf("Waiting login response from kafka")
	response, err := u.waitForResponse(correlationID, 30*time.Second)
	if err != nil {
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Request timeout"})
		return
	}

	c.JSON(response.StatusCode, response.Data)
	log.Printf("Login response received: %s", response.Data)
}

func (u *UserHandler) GetProfile(c *gin.Context) {
	// Extract user ID from JWT token or path parameter
	userID := c.Param("id")
	if userID == "" {
		// Try to get from JWT token (you'd implement JWT middleware)
		userID = c.GetString("user_id")
	}

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	correlationID := uuid.New().String()
	message := models.UserServiceMessage{
		CorrelationID: correlationID,
		Action:        "get_profile",
		Data:          map[string]string{"user_id": userID},
		Timestamp:     time.Now(),
	}

	messageBytes, _ := json.Marshal(message)
	err := u.kafkaWriter.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(correlationID),
			Value: messageBytes,
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
		return
	}

	response, err := u.waitForResponse(correlationID, 120*time.Second)
	if err != nil {
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Request timeout"})
		return
	}

	c.JSON(response.StatusCode, response.Data)
}

func (u *UserHandler) Logout(c *gin.Context) {
	// For logout, you might just invalidate the token locally
	// or send a message to revoke refresh tokens
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// Helper method to wait for Kafka response
func (u *UserHandler) waitForResponse(correlationID string, timeout time.Duration) (*models.UserServiceResponse, error) {

	log.Printf("Waiting for response with correlationID: %s", correlationID)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Timeout waiting for response with correlationID: %s", correlationID)
			return nil, fmt.Errorf("timeout waiting for response")
		default:
			message, err := u.kafkaReader.FetchMessage(ctx)
			if err != nil {
				log.Printf("Error fetching message: %v", err)
				continue
			}

			log.Printf("Received kafka message - Key: %s, Value: %s", string(message.Key), string(message.Value))

			// Check if this message is for us
			var response models.UserServiceResponse
			if err := json.Unmarshal(message.Value, &response); err != nil {
				log.Printf("Failed to unmarshal response: %v", err)
				u.kafkaReader.CommitMessages(context.Background(), message)
				continue
			}

			log.Printf("Parsed response - CorrelationID: %s, StatusCode: %d", response.CorrelationID, response.StatusCode)

			if response.CorrelationID == correlationID {
				log.Printf("Found matching response for correlationID: %s", correlationID)
				u.kafkaReader.CommitMessages(context.Background(), message)
				return &response, nil
			}

			// Not our message, commit and continue
			log.Printf("Message not for us (expected: %s, got: %s), continuing...", correlationID, response.CorrelationID)
			u.kafkaReader.CommitMessages(context.Background(), message)
		}
	}
}
