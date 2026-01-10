package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mbaizhakyp/order-tracker/internal/core/entity"
	"github.com/mbaizhakyp/order-tracker/internal/core/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---

type MockStoreRepo struct {
	mock.Mock
}

func (m *MockStoreRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Store, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Store), args.Error(1)
}

type MockDispatchRepo struct {
	mock.Mock
}

func (m *MockDispatchRepo) CreateOffer(ctx context.Context, offer *entity.DispatchOffer) error {
	args := m.Called(ctx, offer)
	return args.Error(0)
}

// MockLocationRepo and MockOrderRepo reused/extended from previous tests...
// But since they are in `service_test` package, I need to redefine or make them shared.
// For simplicity in this file, I'll redefine them locally for this test context.

type MockLocationRepoDispatch struct {
	mock.Mock
}

func (m *MockLocationRepoDispatch) UpdateShopperLocation(ctx context.Context, loc *entity.ShopperLocation) error {
	return nil
}

func (m *MockLocationRepoDispatch) GetShoppersWithinRadius(ctx context.Context, lat, lng, radiusKm float64) ([]entity.ShopperLocation, error) {
	args := m.Called(ctx, lat, lng, radiusKm)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.ShopperLocation), args.Error(1)
}

// --- Tests ---

func TestDispatchOrder(t *testing.T) {
	orderID := uuid.New()
	storeID := uuid.New()
	shopper1 := uuid.New()
	shopper2 := uuid.New()

	tests := []struct {
		name          string
		setupMocks    func(*MockStoreRepo, *MockLocationRepoDispatch, *MockDispatchRepo, *MockOrderRepo)
		expectedError string
	}{
		{
			name: "Success - Dispatched to 2 shoppers",
			setupMocks: func(s *MockStoreRepo, l *MockLocationRepoDispatch, d *MockDispatchRepo, o *MockOrderRepo) {
				// 1. Get Store
				s.On("GetByID", mock.Anything, storeID).Return(&entity.Store{
					ID: storeID, Lat: 30.0, Lng: -97.0,
				}, nil)

				// 2. Find Shoppers
				l.On("GetShoppersWithinRadius", mock.Anything, 30.0, -97.0, 15.0).Return([]entity.ShopperLocation{
					{ShopperID: shopper1},
					{ShopperID: shopper2},
				}, nil)

				// 3. Create Offers
				d.On("CreateOffer", mock.Anything, mock.MatchedBy(func(offer *entity.DispatchOffer) bool {
					return offer.OrderID == orderID && (offer.ShopperID == shopper1 || offer.ShopperID == shopper2)
				})).Return(nil).Times(2)

				// 4. Update Order
				o.On("UpdateSTATUS", mock.Anything, orderID, entity.OrderStatusOffered).Return(nil)
			},
			expectedError: "",
		},
		{
			name: "No Shoppers Found",
			setupMocks: func(s *MockStoreRepo, l *MockLocationRepoDispatch, d *MockDispatchRepo, o *MockOrderRepo) {
				s.On("GetByID", mock.Anything, storeID).Return(&entity.Store{Lat: 30.0, Lng: -97.0}, nil)
				l.On("GetShoppersWithinRadius", mock.Anything, 30.0, -97.0, 15.0).Return([]entity.ShopperLocation{}, nil)
				// CreateOffer and UpdateStatus should NOT be called
			},
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			storeRepo := new(MockStoreRepo)
			locRepo := new(MockLocationRepoDispatch)
			dispatchRepo := new(MockDispatchRepo)
			orderRepo := new(MockOrderRepo)
			// Reuse MockEventPublisher from order_service_test.go if accessible, or define local
			// For simplicity/speed in this context, defining local mock or using `mock.Anything` if passed

			// We need a specific mock for Publisher
			publisher := new(MockEventPublisher) // We need to define this if not in same package or export it
			publisher.On("Publish", mock.Anything, "offers.dispatch", mock.Anything, mock.Anything).Return(nil).Maybe()

			tc.setupMocks(storeRepo, locRepo, dispatchRepo, orderRepo)

			svc := service.NewDispatchService(storeRepo, locRepo, dispatchRepo, orderRepo, publisher)

			err := svc.DispatchOrder(context.Background(), orderID, storeID)

			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}

			storeRepo.AssertExpectations(t)
			locRepo.AssertExpectations(t)
			dispatchRepo.AssertExpectations(t)
			orderRepo.AssertExpectations(t)
		})
	}
}
