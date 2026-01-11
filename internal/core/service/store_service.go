package service

import (
	"context"

	"github.com/mbaizhakyp/order-tracker/internal/core/entity"
	"github.com/mbaizhakyp/order-tracker/internal/core/ports"
)

type StoreService struct {
	repo ports.StoreRepository
}

func NewStoreService(repo ports.StoreRepository) *StoreService {
	return &StoreService{repo: repo}
}

func (s *StoreService) GetStores(ctx context.Context) ([]entity.Store, error) {
	return s.repo.GetAll(ctx)
}
