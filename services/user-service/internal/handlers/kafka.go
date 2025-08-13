package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/lucas/shared/models"
	usermodels "github.com/lucas/user-service/internal/models"
	"github.com/lucas/user-service/internal/services"
	"github.com/segmentio/kafka-go"
)

type KafkaHandler struct {
	userService *services.UserService
	writer      *kafka.Writer
}

func NewKafkaHandler(userService *services.UserService, writer *kafka.Writer) *KafkaHandler {
	return &KafkaHandler{
		userService: userService,
		writer:      writer,
	}
}

func (h *KafkaHandler) HandleUserMessage(message kafka.Message) {
	var userMsg models.UserServiceMessage
	if err := json.Unmarshal(message.Value, &userMsg); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return
	}

	switch userMsg.Action {
	case "register":
		h.handleRegister(userMsg)
	case "login":
		h.handleLogin(userMsg)
	case "get_profile":
		h.handleGetProfile(userMsg)
	default:
		log.Printf("Unknown action: %s", userMsg.Action)
	}
}

func (h *KafkaHandler) handleRegister(userMsg models.UserServiceMessage) {
	
	// Parse register request
	reqBytes, _ := json.Marshal(userMsg.Data)
	var registerReq models.RegisterRequest
	if err := json.Unmarshal(reqBytes, &registerReq); err != nil {
		h.sendErrorResponse(userMsg.CorrelationID, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Convert to internal model
	createUserReq := &usermodels.CreateUserRequest{
		Email:    registerReq.Email,
		Name:     registerReq.Name,
		Password: registerReq.Password,
	}

	// Register user
	user, err := h.userService.RegisterUser(createUserReq)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "email already registered" || err.Error() == "invalid email format" {
			statusCode = http.StatusBadRequest
		}
		h.sendErrorResponse(userMsg.CorrelationID, statusCode, err.Error())
		return
	}

	// Send success response
	h.sendSuccessResponse(userMsg.CorrelationID, http.StatusCreated, map[string]any{
		"user":    user,
		"message": "User registered successfully",
	})
}

func (h *KafkaHandler) handleLogin(message models.UserServiceMessage) {
	log.Printf("Method not implemented")
}

func (h *KafkaHandler) handleLogout(message models.UserServiceMessage) {
	log.Printf("Method not implemented")

}

func (h *KafkaHandler) handleGetProfile(message models.UserServiceMessage) {
	log.Printf("Method not implemented")
}

func (h *KafkaHandler) sendSuccessResponse(correlationID string, statusCode int, data any) {
	response := models.UserServiceResponse{
		CorrelationID: correlationID,
		StatusCode:    statusCode,
		Data:          data,
	}
	h.sendResponse(response)
}

func (h *KafkaHandler) sendErrorResponse(correlationID string, statusCode int, errorMsg string) {
	response := models.UserServiceResponse{
		CorrelationID: correlationID,
		StatusCode:    statusCode,
		Data:          map[string]string{"error": errorMsg},
		Error:         errorMsg,
	}
	h.sendResponse(response)
}

func (h *KafkaHandler) sendResponse(response models.UserServiceResponse) {
	responseBytes, _ := json.Marshal(response)

	err := h.writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(response.CorrelationID),
			Value: responseBytes,
		},
	)

	if err != nil {
		log.Printf("Failed to send response: %v", err)
	}
}
