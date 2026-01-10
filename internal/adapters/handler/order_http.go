package handler

import (
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
