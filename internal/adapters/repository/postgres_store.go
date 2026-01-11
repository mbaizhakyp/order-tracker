package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mbaizhakyp/order-tracker/internal/core/entity"
	"github.com/mbaizhakyp/order-tracker/internal/core/ports"
)

type PostgresStoreRepository struct {
	db *pgxpool.Pool
}

func NewPostgresStoreRepository(db *pgxpool.Pool) ports.StoreRepository {
	return &PostgresStoreRepository{db: db}
}

func (r *PostgresStoreRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Store, error) {
	query := `
		SELECT id, name, ST_X(location::geometry), ST_Y(location::geometry)
		FROM stores
		WHERE id = $1
	`
	// Note: ST_X is Longitude, ST_Y is Latitude for SRID 4326

	store := &entity.Store{}
	err := r.db.QueryRow(ctx, query, id).Scan(&store.ID, &store.Name, &store.Lng, &store.Lat)
	if err != nil {
		return nil, fmt.Errorf("failed to get store: %w", err)
	}
	return store, nil
}

func (r *PostgresStoreRepository) GetAll(ctx context.Context) ([]entity.Store, error) {
	query := `
		SELECT id, name, ST_X(location::geometry), ST_Y(location::geometry)
		FROM stores
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list stores: %w", err)
	}
	defer rows.Close()

	var stores []entity.Store
	for rows.Next() {
		var s entity.Store
		if err := rows.Scan(&s.ID, &s.Name, &s.Lng, &s.Lat); err != nil {
			return nil, err
		}
		stores = append(stores, s)
	}
	return stores, nil
}

func (r *PostgresStoreRepository) UpdateLocation(ctx context.Context, id string, lat, lng float64) error {
	query := `
		UPDATE stores 
		SET location = ST_SetSRID(ST_MakePoint($1, $2), 4326)
		WHERE id = $3
	`
	_, err := r.db.Exec(ctx, query, lng, lat, id)
	return err
}
