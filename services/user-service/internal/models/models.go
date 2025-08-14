package models

import (
	"time"
)

type User struct {
	ID           int       `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	Name         string    `json:"name" db:"name"`
	PasswordHash string    `json:"-" db:"password_hash"` // "-" excludes from JSON
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type Session struct {
    ID        string    `json:"session_id"`
    UserID    string    `json:"user_id"`
    CreatedAt time.Time `json:"created_at"`
    ExpiresAt time.Time `json:"expires_at"`
    IPAddress string    `json:"ip_address,omitempty"`
    UserAgent string    `json:"user_agent,omitempty"`
}

type RegisterUserRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"` // Plain password, will be hashed
}

type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"` // Plain password, will be hashed
}

// Just for testing, delete later
type GetProfileRequest struct {
	Email string `json:"email"`
}
