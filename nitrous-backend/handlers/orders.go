package handlers

import (
	"net/http"
	"nitrous-backend/database"
	"nitrous-backend/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateOrder creates a merch order for the authenticated user.
func CreateOrder(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.MerchItemIDs) != len(req.Quantities) || len(req.MerchItemIDs) != len(req.UnitPrices) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "merchItemIds, quantities, and unitPrices must have matching lengths"})
		return
	}

	if len(req.MerchItemIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order must include at least one item"})
		return
	}

	database.Mu.RLock()
	merchByID := make(map[string]models.MerchItem, len(database.MerchItems))
	for _, item := range database.MerchItems {
		merchByID[item.ID] = item
	}
	database.Mu.RUnlock()

	total := 0.0
	for i := range req.MerchItemIDs {
		if _, ok := merchByID[req.MerchItemIDs[i]]; !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "Merch item not found"})
			return
		}

		if req.Quantities[i] <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "quantity must be greater than 0 for each item"})
			return
		}

		if req.UnitPrices[i] <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unitPrice must be greater than 0 for each item"})
			return
		}

		total += req.UnitPrices[i] * float64(req.Quantities[i])
	}

	order := models.Order{
		ID:           uuid.New().String(),
		UserID:       userID.(string),
		MerchItemIDs: req.MerchItemIDs,
		Quantities:   req.Quantities,
		UnitPrices:   req.UnitPrices,
		Total:        total,
		Status:       "pending",
		CreatedAt:    time.Now(),
	}

	database.Mu.Lock()
	database.Orders = append(database.Orders, order)
	database.Mu.Unlock()

	c.JSON(http.StatusCreated, order)
}

// GetMyOrders returns all orders for the authenticated user.
func GetMyOrders(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	database.Mu.RLock()
	defer database.Mu.RUnlock()

	var orders []models.Order
	for _, order := range database.Orders {
		if order.UserID == userID.(string) {
			orders = append(orders, order)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"orders": orders,
		"count":  len(orders),
	})
}

// GetOrderByID returns one order if it belongs to the authenticated user.
func GetOrderByID(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	orderID := c.Param("id")
	database.Mu.RLock()

	for _, order := range database.Orders {
		if order.ID == orderID {
			if order.UserID != userID.(string) {
				database.Mu.RUnlock()
				c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
				return
			}

			database.Mu.RUnlock()
			c.JSON(http.StatusOK, order)
			return
		}
	}
	database.Mu.RUnlock()

	c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
}

// CancelOrder cancels an order owned by the authenticated user.
func CancelOrder(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	orderID := c.Param("id")
	database.Mu.Lock()

	for i, order := range database.Orders {
		if order.ID == orderID {
			if order.UserID != userID.(string) {
				database.Mu.Unlock()
				c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
				return
			}

			if order.Status == "cancelled" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Order already cancelled"})
				return
			}

			database.Orders[i].Status = "cancelled"
			database.Mu.Unlock()
			c.JSON(http.StatusOK, database.Orders[i])
			return
		}
	}
	database.Mu.Unlock()

	c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
}
