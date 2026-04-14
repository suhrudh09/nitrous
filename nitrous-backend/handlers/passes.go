package handlers

import (
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

	c.JSON(http.StatusOK, gin.H{
		"message": "Pass purchased successfully",
		"passId":  passID,
	})
}
