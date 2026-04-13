package handlers

import (
	"net/http"
	"nitrous-backend/database"
	"nitrous-backend/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateOrder(c *gin.Context) {
	userID := c.GetString("userID")

	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var total float64
	for i, item := range req.Items {
		m, found := database.FindMerchByID(item.MerchID)
		if !found {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Merch item not found: " + item.MerchID})
			return
		}
		req.Items[i].Name = m.Name
		req.Items[i].Price = m.Price
		total += m.Price * float64(item.Quantity)
	}

	order := models.Order{
		ID:        uuid.New().String(),
		UserID:    userID,
		Items:     req.Items,
		Total:     total,
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	database.AppendOrder(order)

	c.JSON(http.StatusCreated, gin.H{"message": "Order placed successfully", "order": order})
}

func GetMyOrders(c *gin.Context) {
	orders := database.GetOrdersByUser(c.GetString("userID"))
	c.JSON(http.StatusOK, gin.H{"orders": orders, "count": len(orders)})
}

func GetOrderByID(c *gin.Context) {
	order, found, owned := database.FindOrderByID(c.Param("id"), c.GetString("userID"))
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	if !owned {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}
	c.JSON(http.StatusOK, order)
}
