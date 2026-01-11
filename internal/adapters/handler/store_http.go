package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mbaizhakyp/order-tracker/internal/core/service"
)

type StoreHandler struct {
	svc *service.StoreService
}

func NewStoreHandler(svc *service.StoreService) *StoreHandler {
	return &StoreHandler{svc: svc}
}

func (h *StoreHandler) GetStores(c *gin.Context) {
	stores, err := h.svc.GetStores(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stores)
}
