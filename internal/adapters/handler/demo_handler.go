package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mbaizhakyp/order-tracker/internal/core/ports"
)

type DemoHandler struct {
	storeRepo ports.StoreRepository
}

func NewDemoHandler(storeRepo ports.StoreRepository) *DemoHandler {
	return &DemoHandler{storeRepo: storeRepo}
}

func (h *DemoHandler) SetLocation(c *gin.Context) {
	var req struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	// User requested "Target, Tuscaloosa" for the demo
	targetLat := 33.1956
	targetLng := -87.5268

	log.Printf("DEMO: Setting Store Location to Target, Tuscaloosa (%f, %f)", targetLat, targetLng)

	// Update the main store (fixed ID)
	if err := h.storeRepo.UpdateLocation(c.Request.Context(), "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22", targetLat, targetLng); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update store"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "msg": "Store moved to Target, Tuscaloosa"})
}
