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

	var merchItem *models.MerchItem
	for _, item := range database.MerchItems {
		if item.ID == req.MerchItemID {
			itemCopy := item
			merchItem = &itemCopy
			break
		}
	}

	if merchItem == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Merch item not found"})
		return
	}

	order := models.Order{
		ID:          uuid.New().String(),
		UserID:      userID.(string),
		MerchItemID: req.MerchItemID,
		Quantity:    req.Quantity,
		UnitPrice:   merchItem.Price,
		TotalPrice:  merchItem.Price * float64(req.Quantity),
		Status:      "created",
		CreatedAt:   time.Now(),
	}

	database.Orders = append(database.Orders, order)

	c.JSON(http.StatusCreated, order)
}

// GetMyOrders returns all orders for the authenticated user.
func GetMyOrders(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

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

	for _, order := range database.Orders {
		if order.ID == orderID {
			if order.UserID != userID.(string) {
				c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
				return
			}

			c.JSON(http.StatusOK, order)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
}
