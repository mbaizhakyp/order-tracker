package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mbaizhakyp/order-tracker/internal/core/entity"
	"github.com/mbaizhakyp/order-tracker/internal/core/ports"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  ports.UserRepository
	publisher ports.EventPublisher
	jwtSecret string
}

func NewAuthService(repo ports.UserRepository, publisher ports.EventPublisher, secret string) *AuthService {
	return &AuthService{
		userRepo:  repo,
		publisher: publisher,
		jwtSecret: secret,
	}
}

func (s *AuthService) Register(ctx context.Context, name, email, password string, role entity.UserRole) (*entity.User, error) {
	// Check if user exists
	existing, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	// Hash Password
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &entity.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hashed),
		Role:         role,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, *entity.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", nil, err
	}
	if user == nil {
		return "", nil, errors.New("invalid credentials")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	// Generate Token
	token, err := s.generateToken(user)
	if err != nil {
		return "", nil, err
	}

	// Publish SHOPPER_ONLINE event if user is a shopper
	if user.Role == "SHOPPER" {
		eventPayload := map[string]interface{}{
			"type": "SHOPPER_ONLINE",
			"data": map[string]interface{}{
				"shopper_id": user.ID.String(),
				"name":       user.Name,
			},
		}
		// Using "orders.lifecycle" as the common topic for simulation events suitable for Simulator to consume
		if err := s.publisher.Publish(ctx, "orders.lifecycle", user.ID.String(), eventPayload); err != nil {
			fmt.Printf("WARNING: Failed to publish SHOPPER_ONLINE: %v\n", err)
		}
	}

	return token, user, nil
}

func (s *AuthService) generateToken(user *entity.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":  user.ID.String(),
		"role": user.Role,
		"exp":  time.Now().Add(time.Hour * 24).Unix(), // 24 hour expiry
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) AnnounceShopperOnline(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	if user.Role != "SHOPPER" {
		return nil // Only shoppers need to announce presence for simulation
	}

	eventPayload := map[string]interface{}{
		"type": "SHOPPER_ONLINE",
		"data": map[string]interface{}{
			"shopper_id": user.ID.String(),
			"name":       user.Name,
		},
	}
	if err := s.publisher.Publish(ctx, "orders.lifecycle", user.ID.String(), eventPayload); err != nil {
		return fmt.Errorf("failed to publish presence: %w", err)
	}
	return nil
}
