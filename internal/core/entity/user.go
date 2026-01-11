package entity

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleCustomer UserRole = "CUSTOMER"
	RoleShopper  UserRole = "SHOPPER"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never return password hash in JSON
	Role         UserRole  `json:"role"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
}
