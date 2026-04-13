package handlers

import (
	"net/http"
	"nitrous-backend/database"
	"nitrous-backend/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
)

type orderResponseItem struct {
	MerchID  string  `json:"merchId"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
	Size     string  `json:"size,omitempty"`
}

type orderResponse struct {
	ID          string              `json:"id"`
	UserID      string              `json:"userId"`
	MerchItemID string              `json:"merchItemId"`
	Quantity    int                 `json:"quantity"`
	UnitPrice   float64             `json:"unitPrice"`
	TotalPrice  float64             `json:"totalPrice"`
	Items       []orderResponseItem `json:"items"`
	Total       float64             `json:"total"`
	Status      string              `json:"status"`
	CreatedAt   time.Time           `json:"createdAt"`
}

func lookupMerchItem(merchID string) *models.MerchItem {
	for _, item := range database.MerchItems {
		if item.ID == merchID {
			itemCopy := item
			return &itemCopy
		}
	}
	return nil
}

func normalizeOrderStatus(status string) string {
	switch status {
	case "cancelled":
		return "pending"
	case "confirmed", "shipped", "pending":
		return status
	case "created":
		return "confirmed"
	default:
		return "pending"
	}
}

func buildOrderResponse(order models.Order) orderResponse {
	items := make([]orderResponseItem, 0, 1)
	if len(order.Items) > 0 {
		for _, item := range order.Items {
			items = append(items, orderResponseItem{
				MerchID:  item.MerchID,
				Name:     item.Name,
				Price:    item.Price,
				Quantity: item.Quantity,
				Size:     item.Size,
			})
		}
	} else {
		itemName := ""
		itemPrice := order.UnitPrice

		if merch := lookupMerchItem(order.MerchItemID); merch != nil {
			itemName = merch.Name
			if itemPrice == 0 {
				itemPrice = merch.Price
			}
		}

		items = append(items, orderResponseItem{
			MerchID:  order.MerchItemID,
			Name:     itemName,
			Price:    itemPrice,
			Quantity: order.Quantity,
		})
	}

	unitPrice := order.UnitPrice
	if unitPrice == 0 && len(items) > 0 {
		unitPrice = items[0].Price
	}

	return orderResponse{
		ID:          order.ID,
		UserID:      order.UserID,
		MerchItemID: order.MerchItemID,
		Quantity:    order.Quantity,
		UnitPrice:   unitPrice,
		TotalPrice:  order.TotalPrice,
		Items:       items,
		Total:       order.TotalPrice,
		Status:      normalizeOrderStatus(order.Status),
		CreatedAt:   order.CreatedAt,
	}
}

// CreateOrder creates a merch order for the authenticated user.
func CreateOrder(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var cartReq models.CreateOrderItemsRequest
	if err := c.ShouldBindBodyWith(&cartReq, binding.JSON); err == nil && len(cartReq.Items) > 0 {
		total := 0.0
		items := make([]models.OrderItem, 0, len(cartReq.Items))
		for _, incoming := range cartReq.Items {
			merchItem := lookupMerchItem(incoming.MerchID)
			if merchItem == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Merch item not found"})
				return
			}

			price := merchItem.Price
			total += price * float64(incoming.Quantity)
			items = append(items, models.OrderItem{
				MerchID:  incoming.MerchID,
				Name:     merchItem.Name,
				Price:    price,
				Quantity: incoming.Quantity,
				Size:     incoming.Size,
			})
		}

		order := models.Order{
			ID:          uuid.New().String(),
			UserID:      userID.(string),
			MerchItemID: items[0].MerchID,
			Quantity:    items[0].Quantity,
			UnitPrice:   items[0].Price,
			TotalPrice:  total,
			Items:       items,
			Status:      "confirmed",
			CreatedAt:   time.Now(),
		}

		database.Mu.Lock()
		database.Orders = append(database.Orders, order)
		database.Mu.Unlock()

		c.JSON(http.StatusCreated, gin.H{
			"message": "Order created",
			"order":   buildOrderResponse(order),
		})
		return
	}

	var req models.CreateOrderRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	merchItem := lookupMerchItem(req.MerchItemID)
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

	respOrders := make([]orderResponse, 0, len(orders))
	for _, order := range orders {
		respOrders = append(respOrders, buildOrderResponse(order))
	}

	c.JSON(http.StatusOK, gin.H{
		"orders": respOrders,
		"count":  len(respOrders),
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
			c.JSON(http.StatusOK, buildOrderResponse(order))
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
