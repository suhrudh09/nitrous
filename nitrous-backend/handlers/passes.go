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

// UserPassPurchase represents a pass purchased by a user
type UserPassPurchase struct {
	Pass       Pass   `json:"pass"`
	PurchaseID string `json:"purchaseId"`
	CreatedAt  string `json:"createdAt"`
}

// GetMyPasses returns all passes purchased by the authenticated user
func GetMyPasses(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if database.DB != nil {
		rows, err := database.DB.Query(`
			SELECT p.id, p.tier, p.event, p.location, p.date, p.category, p.price, 
			       p.spots_left, p.total_spots, p.badge, p.tier_color,
			       pp.id as purchase_id, pp.created_at
			FROM pass_purchases pp
			JOIN passes p ON pp.pass_id = p.id
			WHERE pp.user_id = $1
			ORDER BY pp.created_at DESC
		`, userID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		purchases := make([]UserPassPurchase, 0)
		for rows.Next() {
			var p Pass
			var purchaseID string
			var createdAt time.Time
			var badge sql.NullString

			if err := rows.Scan(
				&p.ID, &p.Tier, &p.Event, &p.Location, &p.Date, &p.Category, &p.Price,
				&p.SpotsLeft, &p.TotalSpots, &badge, &p.TierColor,
				&purchaseID, &createdAt,
			); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if badge.Valid {
				p.Badge = &badge.String
			}

			purchases = append(purchases, UserPassPurchase{
				Pass:       p,
				PurchaseID: purchaseID,
				CreatedAt:  createdAt.Format(time.RFC3339),
			})
		}

		c.JSON(http.StatusOK, gin.H{"purchases": purchases, "count": len(purchases)})
		return
	}

	// In-memory fallback
	database.Mu.RLock()
	defer database.Mu.RUnlock()

	purchases := make([]UserPassPurchase, 0)
	for _, purchase := range database.PassPurchases {
		if purchase.UserID == userID.(string) {
			for _, pass := range database.Passes {
				if pass.ID == purchase.PassID {
					purchases = append(purchases, UserPassPurchase{
						Pass: Pass{
							ID:         pass.ID,
							Tier:       pass.Tier,
							Event:      pass.Event,
							Location:   pass.Location,
							Date:       pass.Date,
							Category:   pass.Category,
							Price:      pass.Price,
							Perks:      pass.Perks,
							SpotsLeft:  pass.SpotsLeft,
							TotalSpots: pass.TotalSpots,
							Badge:      pass.Badge,
							TierColor:  pass.TierColor,
						},
						PurchaseID: purchase.PassID,
						CreatedAt:  purchase.CreatedAt.Format(time.RFC3339),
					})
					break
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"purchases": purchases, "count": len(purchases)})
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
