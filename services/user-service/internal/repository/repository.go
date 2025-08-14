package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/lib/pq"
	"github.com/lucas/shared/database"
	"github.com/lucas/user-service/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(user *models.RegisterUserRequest, passwordHash string) (*models.User, error) {
	query := `
		INSERT INTO users (email, name, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, email, name, created_at, updated_at
	`

	var createdUser models.User
	err := r.db.QueryRow(query, user.Email, user.Name, passwordHash).Scan(
		&createdUser.ID,
		&createdUser.Email,
		&createdUser.Name,
		&createdUser.CreatedAt,
		&createdUser.UpdatedAt,
	)
	if err != nil {
		// Handle duplicate email error
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, errors.New("email already exists")
		}
		return nil, err
	}

	return &createdUser, nil
}

func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, email, name, password_hash, created_at, updated_at
		FROM users WHERE email = $1
	`

	var user models.User
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) CreateSession(user *models.User) (*models.Session, string, error) {
	// Generate simple session ID using crypto/rand
	sessionBytes := make([]byte, 16)
	if _, err := rand.Read(sessionBytes); err != nil {
		return nil, "", fmt.Errorf("failed to generate session ID: %w", err)
	}
	sessionID := hex.EncodeToString(sessionBytes)

	// Create session object
	session := &models.Session{
		ID:        sessionID,
		UserID:    strconv.Itoa(user.ID),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hour session
	}

	// Store session in Redis
	redisClient := database.GetRedisClient()
	if redisClient == nil {
		return nil, "", errors.New("redis client not available")
	}

	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal session: %w", err)
	}

	ctx := context.Background()
	err = redisClient.Set(ctx, "session:"+sessionID, sessionJSON, 24*time.Hour).Err()
	if err != nil {
		return nil, "", fmt.Errorf("failed to store session in Redis: %w", err)
	}

	// Generate simple token (for MVP - just the session ID)
	// In production, you'd use a proper JWT library
	token := sessionID

	return session, token, nil
}
