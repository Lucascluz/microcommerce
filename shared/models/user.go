package models

import "time"

type RegisterRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserServiceMessage struct {
	CorrelationID string      `json:"correlation_id"`
	Action        string      `json:"action"`
	Data          interface{} `json:"data"`
	Timestamp     time.Time   `json:"timestamp"`
}

type UserServiceResponse struct {
	CorrelationID string      `json:"correlation_id"` // To track the request
	StatusCode    int         `json:"status_code"`
	Data          interface{} `json:"data"`
	Error         string      `json:"error,omitempty"`
}
