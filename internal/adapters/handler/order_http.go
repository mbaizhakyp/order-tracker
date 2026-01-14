package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mbaizhakyp/order-tracker/internal/core/service"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req service.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extract UserID from context (set by Middleware)
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Handle both string and UUID types for safety (middleware usually sets string from standard claims)
	if idStr, ok := userIDVal.(string); ok {
		if uid, err := uuid.Parse(idStr); err == nil {
			req.CustomerID = uid
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id in token"})
			return
		}
	} else if idUUID, ok := userIDVal.(uuid.UUID); ok {
		req.CustomerID = idUUID
	} else {
		// Try float64 (JWT numeric claim edge case) or map
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id type"})
		return
	}

	order, err := h.svc.CreateOrder(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid"})
		return
	}

	order, err := h.svc.GetOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

type ClaimOrderRequest struct {
	ShopperID uuid.UUID `json:"shopper_id" binding:"required"`
}

func (h *OrderHandler) ClaimOrder(c *gin.Context) {
	idParam := c.Param("id")
	orderID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	var req ClaimOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.ClaimOrder(c.Request.Context(), orderID, req.ShopperID); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()}) // 409 Conflict is appropriate for race condition failure
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "CLAIMED"})
}

func (h *OrderHandler) ArriveAtStore(c *gin.Context) {
	idParam := c.Param("id")
	orderID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	if err := h.svc.ArriveAtStore(c.Request.Context(), orderID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ARRIVED_AT_STORE"})
}

func (h *OrderHandler) PickUpOrder(c *gin.Context) {
	idParam := c.Param("id")
	orderID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	if err := h.svc.PickUpOrder(c.Request.Context(), orderID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "PICKED_UP"})
}

func (h *OrderHandler) ArriveAtCustomer(c *gin.Context) {
	idParam := c.Param("id")
	orderID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	if err := h.svc.ArriveAtCustomer(c.Request.Context(), orderID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ARRIVED_AT_CUSTOMER"})
}

func (h *OrderHandler) DeliverOrder(c *gin.Context) {
	idParam := c.Param("id")
	orderID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	if err := h.svc.DeliverOrder(c.Request.Context(), orderID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "DELIVERED"})
}

func (h *OrderHandler) GetOrderHistory(c *gin.Context) {
	idParam := c.Param("id")
	orderID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	events, err := h.svc.GetOrderHistory(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, events)
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	idParam := c.Param("id")
	orderID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	if err := h.svc.CancelOrder(c.Request.Context(), orderID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "CANCELLED"})
}

func (h *OrderHandler) GetUserOrders(c *gin.Context) {
	var userID uuid.UUID
	var err error

	// 1. Try to get from Auth Middleware
	userIDVal, exists := c.Get("userID")
	if exists {
		if idStr, ok := userIDVal.(string); ok {
			userID, err = uuid.Parse(idStr)
		} else if idUUID, ok := userIDVal.(uuid.UUID); ok {
			userID = idUUID
		} else {
			err = fmt.Errorf("invalid user id type in context")
		}
	} else {
		// 2. Fallback to Demo ID if not authenticated (should be behind middleware ideally)
		// For consistent demo experience if user hits this without token (though frontend sends it)
		userID, _ = uuid.Parse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	orders, err := h.svc.GetUserOrders(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) GetShopperActiveOrder(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	shopperID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	orderData, err := h.svc.GetShopperActiveOrder(c.Request.Context(), shopperID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if orderData == nil {
		c.JSON(http.StatusOK, nil)
		return
	}

	c.JSON(http.StatusOK, orderData)
}
