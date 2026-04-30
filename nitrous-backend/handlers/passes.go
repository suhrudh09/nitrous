package handlers

import (
	"database/sql"
	"encoding/json"
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
	Quantity   int    `json:"quantity"`
}

// GetMyPasses returns all passes purchased by the authenticated user
// GetAllPasses returns all available passes (public)
func GetAllPasses(c *gin.Context) {
	if database.DB != nil {
		rows, err := database.DB.Query(`
			SELECT id, tier, event_name, location, event_date, category, price::float8,
			       perks, spots_left, total_spots, badge, tier_color
			FROM passes ORDER BY event_date
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		passes := make([]Pass, 0)
		for rows.Next() {
			var p Pass
			var perksJSON []byte
			var badge sql.NullString
			var eventDate time.Time
			if err := rows.Scan(&p.ID, &p.Tier, &p.Event, &p.Location, &eventDate, &p.Category, &p.Price, &perksJSON, &p.SpotsLeft, &p.TotalSpots, &badge, &p.TierColor); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			p.Date = eventDate.UTC().Format(time.RFC3339)
			if len(perksJSON) > 0 {
				_ = json.Unmarshal(perksJSON, &p.Perks)
			}
			if badge.Valid {
				p.Badge = &badge.String
			}
			passes = append(passes, p)
		}
		c.JSON(http.StatusOK, gin.H{"passes": passes, "count": len(passes)})
		return
	}

	database.Mu.RLock()
	defer database.Mu.RUnlock()

	result := make([]Pass, 0, len(database.Passes))
	for _, p := range database.Passes {
		badge := (*string)(nil)
		if p.Badge != nil {
			v := *p.Badge
			badge = &v
		}
		result = append(result, Pass{
			ID: p.ID, Tier: p.Tier, Event: p.Event, Location: p.Location,
			Date: p.Date, Category: p.Category, Price: p.Price, Perks: p.Perks,
			SpotsLeft: p.SpotsLeft, TotalSpots: p.TotalSpots, Badge: badge, TierColor: p.TierColor,
		})
	}
	c.JSON(http.StatusOK, gin.H{"passes": result, "count": len(result)})
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
			       pp.id as purchase_id, pp.created_at, COALESCE(pp.quantity, 1)
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
			var qty int

			if err := rows.Scan(
				&p.ID, &p.Tier, &p.Event, &p.Location, &p.Date, &p.Category, &p.Price,
				&p.SpotsLeft, &p.TotalSpots, &badge, &p.TierColor,
				&purchaseID, &createdAt, &qty,
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
				Quantity:   qty,
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

	var body struct {
		Quantity int `json:"quantity"`
	}
	body.Quantity = 1
	if c.Request.ContentLength > 0 {
		_ = c.ShouldBindJSON(&body)
	}
	if body.Quantity < 1 {
		body.Quantity = 1
	}

	if database.DB != nil {
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

		if spotsLeft < body.Quantity {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Not enough spots remaining"})
			return
		}

		tx, err := database.DB.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer tx.Rollback()

		// Decrement spots by quantity
		res, err := tx.Exec(`UPDATE passes SET spots_left = spots_left - $1 WHERE id=$2 AND spots_left >= $1`, body.Quantity, passID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Not enough spots remaining"})
			return
		}

		// Upsert purchase record — update quantity if already exists
		_, err = tx.Exec(`
			INSERT INTO pass_purchases (user_id, pass_id, quantity, created_at)
			VALUES ($1,$2,$3,$4)
			ON CONFLICT (user_id, pass_id) DO UPDATE SET quantity = pass_purchases.quantity + EXCLUDED.quantity
		`, userID.(string), passID, body.Quantity, time.Now())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Pass purchased successfully", "passId": passID, "quantity": body.Quantity})
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

	if foundPass.SpotsLeft < body.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not enough spots remaining"})
		return
	}

	foundPass.SpotsLeft -= body.Quantity
	database.PassPurchases = append(database.PassPurchases, database.PassPurchase{
		UserID:    userID.(string),
		PassID:    passID,
		CreatedAt: time.Now(),
	})

	c.JSON(http.StatusOK, gin.H{"message": "Pass purchased successfully", "passId": passID, "quantity": body.Quantity})
}
