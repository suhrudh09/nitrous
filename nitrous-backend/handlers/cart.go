package handlers

import (
	"net/http"
	"nitrous-backend/database"
	"nitrous-backend/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GetCart returns the authenticated user's saved cart.
func GetCart(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	if database.DB != nil {
		rows, err := database.DB.Query(
			`SELECT ci.merch_item_id, COALESCE(mi.name, ''), COALESCE(mi.icon, ''), COALESCE(mi.price::float8, 0), COALESCE(mi.category, ''), ci.quantity, ci.size
			 FROM cart_items ci
			 JOIN merch_items mi ON mi.id = ci.merch_item_id
			 WHERE ci.user_id = $1
			 ORDER BY ci.updated_at DESC, ci.id DESC`,
			userID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cart"})
			return
		}
		defer rows.Close()

		items := make([]models.CartItem, 0)
		for rows.Next() {
			var item models.CartItem
			if err := rows.Scan(&item.MerchID, &item.Name, &item.Icon, &item.Price, &item.Category, &item.Quantity, &item.Size); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read cart items"})
				return
			}
			if item.Size == "" {
				item.Size = ""
			}
			items = append(items, item)
		}

		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read cart items"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"items": items, "count": len(items)})
		return
	}

	database.Mu.RLock()
	items := append([]models.CartItem(nil), database.CartItems[userID]...)
	database.Mu.RUnlock()
	c.JSON(http.StatusOK, gin.H{"items": items, "count": len(items)})
}

// SaveCart replaces the authenticated user's cart contents.
func SaveCart(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var req models.UpsertCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items := make([]models.CartItem, 0, len(req.Items))
	seen := map[string]bool{}
	for _, item := range req.Items {
		item.MerchID = strings.TrimSpace(item.MerchID)
		item.Size = strings.TrimSpace(item.Size)
		if item.MerchID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "merchId is required for each cart item"})
			return
		}
		if item.Quantity <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "quantity must be greater than zero"})
			return
		}
		key := item.MerchID + "::" + strings.ToLower(item.Size)
		if seen[key] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "duplicate cart entries for same item and size"})
			return
		}
		seen[key] = true
		items = append(items, item)
	}

	if database.DB != nil {
		tx, err := database.DB.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save cart"})
			return
		}
		defer tx.Rollback()

		if _, err := tx.Exec(`DELETE FROM cart_items WHERE user_id = $1`, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save cart"})
			return
		}

		now := time.Now()
		for _, item := range items {
			if _, err := tx.Exec(
				`INSERT INTO cart_items (user_id, merch_item_id, size, quantity, created_at, updated_at)
				 VALUES ($1, $2, $3, $4, $5, $6)`,
				userID, item.MerchID, item.Size, item.Quantity, now, now,
			); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to save cart item"})
				return
			}
		}

		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save cart"})
			return
		}

		GetCart(c)
		return
	}

	database.Mu.Lock()
	database.CartItems[userID] = items
	database.Mu.Unlock()
	c.JSON(http.StatusOK, gin.H{"items": items, "count": len(items)})
}

// ClearCart deletes all cart items for the authenticated user.
func ClearCart(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	if database.DB != nil {
		if _, err := database.DB.Exec(`DELETE FROM cart_items WHERE user_id = $1`, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear cart"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Cart cleared"})
		return
	}

	database.Mu.Lock()
	delete(database.CartItems, userID)
	database.Mu.Unlock()
	c.JSON(http.StatusOK, gin.H{"message": "Cart cleared"})
}
