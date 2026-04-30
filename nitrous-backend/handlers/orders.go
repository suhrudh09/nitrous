package handlers

import (
	"database/sql"
	"net/http"
	"nitrous-backend/database"
	"nitrous-backend/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const pendingOrderExpiry = time.Minute

func expirePendingOrdersForUser(userID string) error {
	if database.DB != nil {
		_, err := database.DB.Exec(
			`UPDATE orders
			 SET status='failed'
			 WHERE user_id=$1
			   AND status='pending'
			   AND created_at < $2`,
			userID,
			time.Now().Add(-pendingOrderExpiry),
		)
		return err
	}

	database.Mu.Lock()
	defer database.Mu.Unlock()
	for i := range database.Orders {
		if database.Orders[i].UserID == userID && database.Orders[i].Status == "pending" {
			if time.Since(database.Orders[i].CreatedAt) > pendingOrderExpiry {
				database.Orders[i].Status = "failed"
			}
		}
	}

	return nil
}

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

	if database.DB != nil {
		// Insert order
		_, err := database.DB.Exec(`INSERT INTO orders (id, user_id, total, status, created_at) VALUES ($1,$2,$3,$4,$5)`, order.ID, order.UserID, order.Total, order.Status, order.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Insert items
		for i, itemID := range req.MerchItemIDs {
			_, err := database.DB.Exec(`INSERT INTO order_items (order_id, merch_item_id, quantity, unit_price) VALUES ($1,$2,$3,$4)`, order.ID, itemID, req.Quantities[i], req.UnitPrices[i])
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		c.JSON(http.StatusCreated, order)
		return
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

	if err := expirePendingOrdersForUser(userID.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if database.DB != nil {
		rows, err := database.DB.Query(`SELECT id, user_id::text, total::float8, status, created_at FROM orders WHERE user_id=$1 ORDER BY created_at DESC`, userID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		orders := make([]models.Order, 0)
		for rows.Next() {
			var o models.Order
			if err := rows.Scan(&o.ID, &o.UserID, &o.Total, &o.Status, &o.CreatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			// load items
			itemRows, err := database.DB.Query(`SELECT merch_item_id, quantity, unit_price::float8 FROM order_items WHERE order_id=$1 ORDER BY id`, o.ID)
			if err == nil {
				for itemRows.Next() {
					var itemID string
					var qty int
					var price float64
					itemRows.Scan(&itemID, &qty, &price)
					o.MerchItemIDs = append(o.MerchItemIDs, itemID)
					o.Quantities = append(o.Quantities, qty)
					o.UnitPrices = append(o.UnitPrices, price)
				}
				itemRows.Close()
			}
			orders = append(orders, o)
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"orders": orders, "count": len(orders)})
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

	c.JSON(http.StatusOK, gin.H{"orders": orders, "count": len(orders)})
}

// GetOrderByID returns one order if it belongs to the authenticated user.
func GetOrderByID(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := expirePendingOrdersForUser(userID.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	orderID := c.Param("id")

	if database.DB != nil {
		var o models.Order
		row := database.DB.QueryRow(`SELECT id, user_id::text, total::float8, status, created_at FROM orders WHERE id=$1`, orderID)
		if err := row.Scan(&o.ID, &o.UserID, &o.Total, &o.Status, &o.CreatedAt); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if o.UserID != userID.(string) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		// load items
		itemRows, err := database.DB.Query(`SELECT merch_item_id, quantity, unit_price::float8 FROM order_items WHERE order_id=$1 ORDER BY id`, o.ID)
		if err == nil {
			for itemRows.Next() {
				var itemID string
				var qty int
				var price float64
				itemRows.Scan(&itemID, &qty, &price)
				o.MerchItemIDs = append(o.MerchItemIDs, itemID)
				o.Quantities = append(o.Quantities, qty)
				o.UnitPrices = append(o.UnitPrices, price)
			}
			itemRows.Close()
		}
		c.JSON(http.StatusOK, o)
		return
	}

	database.Mu.RLock()
	defer database.Mu.RUnlock()

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

// CancelOrder cancels an order owned by the authenticated user.
func CancelOrder(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := expirePendingOrdersForUser(userID.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	orderID := c.Param("id")

	if database.DB != nil {
		// Check ownership and current status
		var ownerID, status string
		row := database.DB.QueryRow(`SELECT user_id::text, status FROM orders WHERE id=$1`, orderID)
		if err := row.Scan(&ownerID, &status); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if ownerID != userID.(string) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		if status != "pending" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending orders can be cancelled"})
			return
		}

		_, err := database.DB.Exec(`UPDATE orders SET status='cancelled' WHERE id=$1`, orderID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return updated order
		var o models.Order
		row = database.DB.QueryRow(`SELECT id, user_id::text, total::float8, status, created_at FROM orders WHERE id=$1`, orderID)
		if err := row.Scan(&o.ID, &o.UserID, &o.Total, &o.Status, &o.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		itemRows, _ := database.DB.Query(`SELECT merch_item_id, quantity, unit_price::float8 FROM order_items WHERE order_id=$1 ORDER BY id`, o.ID)
		if itemRows != nil {
			for itemRows.Next() {
				var itemID string
				var qty int
				var price float64
				itemRows.Scan(&itemID, &qty, &price)
				o.MerchItemIDs = append(o.MerchItemIDs, itemID)
				o.Quantities = append(o.Quantities, qty)
				o.UnitPrices = append(o.UnitPrices, price)
			}
			itemRows.Close()
		}
		c.JSON(http.StatusOK, o)
		return
	}

	database.Mu.Lock()
	defer database.Mu.Unlock()

	for i, order := range database.Orders {
		if order.ID == orderID {
			if order.UserID != userID.(string) {
				c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
				return
			}

			if order.Status != "pending" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending orders can be cancelled"})
				return
			}

			database.Orders[i].Status = "cancelled"
			c.JSON(http.StatusOK, database.Orders[i])
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
}
