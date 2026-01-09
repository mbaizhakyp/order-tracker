package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/mbaizhakyp/order-tracker/internal/core/entity"
	"github.com/mbaizhakyp/order-tracker/internal/core/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---

type MockLocationRepo struct {
	mock.Mock
}

func (m *MockLocationRepo) UpdateShopperLocation(ctx context.Context, loc *entity.ShopperLocation) error {
	args := m.Called(ctx, loc)
	return args.Error(0)
}

func (m *MockLocationRepo) GetShoppersWithinRadius(ctx context.Context, lat, lng, radiusKm float64) ([]entity.ShopperLocation, error) {
	args := m.Called(ctx, lat, lng, radiusKm)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.ShopperLocation), args.Error(1)
}

// --- Tests ---

func TestUpdateLocation(t *testing.T) {
	shopperID := uuid.New()

	tests := []struct {
		name          string
		req           service.UpdateLocationRequest
		setupMocks    func(*MockLocationRepo)
		expectedError string
	}{
		{
			name: "Success",
			req: service.UpdateLocationRequest{
				ShopperID: shopperID,
				Lat:       30.2672,
				Lng:       -97.7431,
			},
			setupMocks: func(m *MockLocationRepo) {
				m.On("UpdateShopperLocation", mock.Anything, mock.MatchedBy(func(l *entity.ShopperLocation) bool {
					return l.ShopperID == shopperID && l.Location.Lat == 30.2672
				})).Return(nil)
			},
			expectedError: "",
		},
		{
			name: "Invalid Latitude (Low)",
			req: service.UpdateLocationRequest{
				ShopperID: shopperID,
				Lat:       -91.0,
				Lng:       -97.7431,
			},
			setupMocks: func(m *MockLocationRepo) {
				// No calls
			},
			expectedError: "invalid latitude",
		},
		{
			name: "Invalid Latitude (High)",
			req: service.UpdateLocationRequest{
				ShopperID: shopperID,
				Lat:       91.0,
				Lng:       -97.7431,
			},
			setupMocks: func(m *MockLocationRepo) {
				// No calls
			},
			expectedError: "invalid latitude",
		},
		{
			name: "Invalid Longitude",
			req: service.UpdateLocationRequest{
				ShopperID: shopperID,
				Lat:       30.2672,
				Lng:       181.0,
			},
			setupMocks: func(m *MockLocationRepo) {
				// No calls
			},
			expectedError: "invalid longitude",
		},
		{
			name: "Repository Error",
			req: service.UpdateLocationRequest{
				ShopperID: shopperID,
				Lat:       30.2672,
				Lng:       -97.7431,
			},
			setupMocks: func(m *MockLocationRepo) {
				m.On("UpdateShopperLocation", mock.Anything, mock.Anything).Return(errors.New("redis error"))
			},
			expectedError: "failed to update location: redis error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Init Mocks
			mockRepo := new(MockLocationRepo)
			tc.setupMocks(mockRepo)

			// Init Service
			svc := service.NewLocationService(mockRepo)

			// Execute
			err := svc.UpdateLocation(context.Background(), tc.req)

			// Assert
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// Verify Mocks
			mockRepo.AssertExpectations(t)
		})
	}
}
