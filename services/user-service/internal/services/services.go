package services

import (
	"errors"

	"github.com/lucas/user-service/internal/models"
	"github.com/lucas/user-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) RegisterUser(req *models.CreateUserRequest) (*models.User, error) {
	// Validate email
	if !s.isValidEmail(req.Email) {
		return nil, errors.New("invalid email format")
	}

	// Validade password
	if !s.isValidPassword(req.Password) {
		return nil, errors.New("invalid password format")
	}

	// Verify is email is in use
	existingUser, err := s.userRepo.GetUserByEmail(req.Email)
	if err != nil && err.Error() != "user not found" {
		return nil, errors.New("database error")
	}
	if existingUser != nil {
		return nil, errors.New("email already in use")
	}

	// Hash password
	hash, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, errors.New("error hashing password")
	}

	// Create user in the database
	user, err := s.userRepo.CreateUser(req, string(hash))
	if err != nil {
		return nil, errors.New("error creating user")
	}

	// Return the created user
	return user, nil
}

func (s *UserService) isValidEmail(email string) bool {
	return true
}

func (s *UserService) isValidPassword(password string) bool {
	return len(password) >= 8
}

func (s *UserService) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(hashedBytes), err
}
