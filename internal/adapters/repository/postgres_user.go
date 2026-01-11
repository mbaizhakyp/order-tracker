package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mbaizhakyp/order-tracker/internal/core/entity"
	"github.com/mbaizhakyp/order-tracker/internal/core/ports"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) ports.UserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, name, email, password_hash, role, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	_, err := r.db.Exec(ctx, query,
		user.ID,
		user.Name,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.CreatedAt,
	)
	return err
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `SELECT id, name, email, password_hash, role, created_at FROM users WHERE email = $1`
	var user entity.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // Return nil if not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	query := `SELECT id, name, email, password_hash, role, created_at FROM users WHERE id = $1`
	var user entity.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // Return nil if not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}
