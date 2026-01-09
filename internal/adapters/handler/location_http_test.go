package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mbaizhakyp/order-tracker/internal/adapters/handler"
	"github.com/mbaizhakyp/order-tracker/internal/core/entity"
	"github.com/mbaizhakyp/order-tracker/internal/core/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock Repo ---
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
	gin.SetMode(gin.TestMode)

	shopperID := uuid.New()

	tests := []struct {
		name           string
		payload        interface{}
		setupMocks     func(*MockLocationRepo)
		expectedStatus int
	}{
		{
			name: "Success",
			payload: service.UpdateLocationRequest{
				ShopperID: shopperID,
				Lat:       30.2672,
				Lng:       -97.7431,
			},
			setupMocks: func(m *MockLocationRepo) {
				m.On("UpdateShopperLocation", mock.Anything, mock.MatchedBy(func(l *entity.ShopperLocation) bool {
					return l.ShopperID == shopperID && l.Location.Lat == 30.2672
				})).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Invalid JSON",
			payload: "invalid-json",
			setupMocks: func(m *MockLocationRepo) {
				// No calls
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid Coordinates (Service Validation)",
			payload: service.UpdateLocationRequest{
				ShopperID: shopperID,
				Lat:       200.0, // Invalid
				Lng:       0,
			},
			setupMocks: func(m *MockLocationRepo) {
				// Service should block this before repo call, or returning error
			},
			expectedStatus: http.StatusInternalServerError, // Or Bad Request if we mapped errors better
		},
		{
			name: "Redis Error",
			payload: service.UpdateLocationRequest{
				ShopperID: shopperID,
				Lat:       30.2672,
				Lng:       -97.7431,
			},
			setupMocks: func(m *MockLocationRepo) {
				m.On("UpdateShopperLocation", mock.Anything, mock.Anything).Return(errors.New("redis down"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockRepo := new(MockLocationRepo)
			tc.setupMocks(mockRepo)
			svc := service.NewLocationService(mockRepo)
			h := handler.NewLocationHandler(svc)

			r := gin.New()
			r.POST("/location", h.UpdateLocation)

			// Request
			var body []byte
			if str, ok := tc.payload.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tc.payload)
			}

			req, _ := http.NewRequest("POST", "/location", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			// Execute
			r.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tc.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}
