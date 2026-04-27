package handlers

import (
	"database/sql"
	"net/http"
	"nitrous-backend/database"
	"nitrous-backend/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type Pass struct {
	ID         string   `json:"id" db:"id"`
	Tier       string   `json:"tier" db:"tier"`
	Event      string   `json:"event" db:"event"`
	Location   string   `json:"location" db:"location"`
	Date       string   `json:"date" db:"date"`
	Category   string   `json:"category" db:"category"`
	Price      float64  `json:"price" db:"price"`
	Perks      []string `json:"perks" db:"-"`
	SpotsLeft  int      `json:"spotsLeft" db:"spots_left"`
	TotalSpots int      `json:"totalSpots" db:"total_spots"`
	Badge      *string  `json:"badge" db:"badge"`
	TierColor  string   `json:"tierColor" db:"tier_color"`
}

func PurchasePass(c *gin.Context) {
	passID := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	_ = utils.Claims{}

	if database.DB != nil {
		// Check if pass exists and has spots
		var spotsLeft int
		row := database.DB.QueryRow(`SELECT spots_left FROM passes WHERE id=$1`, passID)
		if err := row.Scan(&spotsLeft); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Pass not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if spotsLeft <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No spots remaining"})
			return
		}

		// Try to decrement spots and insert purchase in one atomic operation
		// The unique constraint on (user_id, pass_id) will prevent duplicate purchases
		tx, err := database.DB.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer tx.Rollback()

		// Check if user already has this pass
		var purchaseExists bool
		err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM pass_purchases WHERE user_id=$1 AND pass_id=$2)`, userID.(string), passID).Scan(&purchaseExists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if purchaseExists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You already have this pass"})
			return
		}

		// Decrement spots
		res, err := tx.Exec(`UPDATE passes SET spots_left = spots_left - 1 WHERE id=$1 AND spots_left > 0`, passID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No spots remaining"})
			return
		}

		// Insert purchase record
		_, err = tx.Exec(`INSERT INTO pass_purchases (user_id, pass_id, created_at) VALUES ($1,$2,$3)`, userID.(string), passID, time.Now())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Pass purchased successfully", "passId": passID})
		return
	}

	database.Mu.Lock()
	defer database.Mu.Unlock()

	var foundPass *database.Pass
	for i := range database.Passes {
		if database.Passes[i].ID == passID {
			foundPass = &database.Passes[i]
			break
		}
	}
	if foundPass == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pass not found"})
		return
	}

	if foundPass.SpotsLeft <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No spots remaining"})
		return
	}

	for _, purchase := range database.PassPurchases {
		if purchase.UserID == userID.(string) && purchase.PassID == passID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You already have this pass"})
			return
		}
	}

	foundPass.SpotsLeft--
	database.PassPurchases = append(database.PassPurchases, database.PassPurchase{
		UserID:    userID.(string),
		PassID:    passID,
		CreatedAt: time.Now(),
	})

	c.JSON(http.StatusOK, gin.H{"message": "Pass purchased successfully", "passId": passID})
}
