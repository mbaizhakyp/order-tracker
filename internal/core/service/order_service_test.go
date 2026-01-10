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

type MockOrderRepo struct {
	mock.Mock
}

func (m *MockOrderRepo) Create(ctx context.Context, order *entity.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Order), args.Error(1)
}

func (m *MockOrderRepo) UpdateSTATUS(ctx context.Context, id uuid.UUID, status entity.OrderStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockOrderRepo) ClaimOrder(ctx context.Context, orderID uuid.UUID, shopperID uuid.UUID) error {
	args := m.Called(ctx, orderID, shopperID)
	return args.Error(0)
}

type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) Publish(ctx context.Context, topic string, key string, payload interface{}) error {
	args := m.Called(ctx, topic, key, payload)
	return args.Error(0)
}

// --- Tests ---

func TestCreateOrder(t *testing.T) {
	// Common dummy data
	custID := uuid.New()
	storeID := uuid.New()
	items := []entity.OrderItem{{Name: "Item1", Quantity: 1}}

	tests := []struct {
		name          string
		req           service.CreateOrderRequest
		setupMocks    func(*MockOrderRepo, *MockEventPublisher)
		expectedError string
	}{
		{
			name: "Success",
			req: service.CreateOrderRequest{
				CustomerID: custID,
				StoreID:    storeID,
				TotalCents: 1000,
				Items:      items,
			},
			setupMocks: func(r *MockOrderRepo, p *MockEventPublisher) {
				r.On("Create", mock.Anything, mock.MatchedBy(func(o *entity.Order) bool {
					return o.TotalAmount == 1000 && o.Status == entity.OrderStatusCreated
				})).Return(nil)
				p.On("Publish", mock.Anything, "orders.lifecycle", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: "",
		},
		{
			name: "Invalid Total Amount",
			req: service.CreateOrderRequest{
				CustomerID: custID,
				StoreID:    storeID,
				TotalCents: -50,
				Items:      items,
			},
			setupMocks: func(r *MockOrderRepo, p *MockEventPublisher) {
				// No calls expected
			},
			expectedError: "total amount must be positive",
		},
		{
			name: "Empty Items",
			req: service.CreateOrderRequest{
				CustomerID: custID,
				StoreID:    storeID,
				TotalCents: 1000,
				Items:      nil,
			},
			setupMocks: func(r *MockOrderRepo, p *MockEventPublisher) {
				// No calls expected
			},
			expectedError: "order must have items",
		},
		{
			name: "Repository Error",
			req: service.CreateOrderRequest{
				CustomerID: custID,
				StoreID:    storeID,
				TotalCents: 1000,
				Items:      items,
			},
			setupMocks: func(r *MockOrderRepo, p *MockEventPublisher) {
				r.On("Create", mock.Anything, mock.Anything).Return(errors.New("db disconnect"))
			},
			expectedError: "failed to save order: db disconnect",
		},
		{
			name: "Publisher Error (Should Not Fail Request)",
			req: service.CreateOrderRequest{
				CustomerID: custID,
				StoreID:    storeID,
				TotalCents: 1000,
				Items:      items,
			},
			setupMocks: func(r *MockOrderRepo, p *MockEventPublisher) {
				r.On("Create", mock.Anything, mock.Anything).Return(nil)
				// Test that business logic swallows the error (as currently implemented)
				p.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("kafka down"))
			},
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Init Mocks
			mockRepo := new(MockOrderRepo)
			mockPub := new(MockEventPublisher)
			tc.setupMocks(mockRepo, mockPub)

			// Init Service
			svc := service.NewOrderService(mockRepo, mockPub)

			// Execute
			_, err := svc.CreateOrder(context.Background(), tc.req)

			// Assert
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// Verify Mocks
			mockRepo.AssertExpectations(t)
			mockPub.AssertExpectations(t)
		})
	}
}
