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

// CreatePaymentIntentRequest represents a payment intent creation request
type CreatePaymentIntentRequest struct {
	Amount        float64 `json:"amount" binding:"required"`
	Currency      string  `json:"currency"`
	Description   string  `json:"description"`
	ReferenceType string  `json:"referenceType" binding:"required"`
	ReferenceID   string  `json:"referenceId" binding:"required"`
}

// PaymentIntentResponse represents the response for creating a payment intent
type PaymentIntentResponse struct {
	ClientSecret string  `json:"clientSecret"`
	PaymentID    string  `json:"paymentId"`
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
}

// CreatePaymentIntent creates a Stripe payment intent
func CreatePaymentIntent(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var req CreatePaymentIntentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currency := req.Currency
	if currency == "" {
		currency = "usd"
	}

	// Create payment record
	payment := models.Payment{
		ID:                    uuid.New().String(),
		UserID:                userID,
		Amount:                req.Amount,
		Currency:              currency,
		Status:                "pending",
		PaymentMethod:         "card",
		StripePaymentIntentID: "",
		Description:           req.Description,
		ReferenceType:         req.ReferenceType,
		ReferenceID:           req.ReferenceID,
		Metadata:              map[string]interface{}{},
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	// In production, this would create a real Stripe PaymentIntent
	// For now, we'll simulate the flow
	if database.DB != nil {
		_, err = database.DB.Exec(
			`INSERT INTO payments (id, user_id, amount, currency, status, payment_method, description, reference_type, reference_id, metadata, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			payment.ID, payment.UserID, payment.Amount, payment.Currency, payment.Status, payment.PaymentMethod,
			payment.Description, payment.ReferenceType, payment.ReferenceID, "{}", payment.CreatedAt, payment.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment"})
			return
		}
	} else {
		database.Mu.Lock()
		defer database.Mu.Unlock()
		database.Payments = append(database.Payments, payment)
	}

	// In production, you would call Stripe API here:
	// stripePaymentIntent, _ := stripe.PaymentIntent.New(&stripe.PaymentIntentParams{
	//   Amount: stripe.Int64(int64(req.Amount * 100)),
	//   Currency: stripe.String(currency),
	//   Customer: stripe.String(customerID),
	// })

	// For demo, we generate a mock client secret
	clientSecret := payment.ID + "_secret_" + time.Now().Format("20060102150405")

	c.JSON(http.StatusOK, PaymentIntentResponse{
		ClientSecret: clientSecret,
		PaymentID:    payment.ID,
		Amount:       payment.Amount,
		Currency:     currency,
	})
}

// ConfirmPayment confirms a payment after successful processing
func ConfirmPayment(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	paymentID := c.Param("id")

	if database.DB != nil {
		var referenceType, referenceID string
		row := database.DB.QueryRow(`SELECT reference_type, reference_id FROM payments WHERE id = $1 AND user_id = $2`, paymentID, userID)
		if err := row.Scan(&referenceType, &referenceID); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to confirm payment"})
			return
		}

		result, err := database.DB.Exec(
			`UPDATE payments SET status = $1, updated_at = $2 WHERE id = $3 AND user_id = $4`,
			"completed", time.Now(), paymentID, userID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to confirm payment"})
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
			return
		}

		if referenceType == "order" && referenceID != "" {
			_, _ = database.DB.Exec(`UPDATE orders SET status='confirmed' WHERE id=$1 AND user_id=$2`, referenceID, userID)
		}

		c.JSON(http.StatusOK, gin.H{"message": "Payment confirmed"})
		return
	}

	// In-memory fallback
	database.Mu.Lock()
	defer database.Mu.Unlock()

	found := false
	var referenceType, referenceID string
	for i := range database.Payments {
		if database.Payments[i].ID == paymentID && database.Payments[i].UserID == userID {
			database.Payments[i].Status = "completed"
			database.Payments[i].UpdatedAt = time.Now()
			referenceType = database.Payments[i].ReferenceType
			referenceID = database.Payments[i].ReferenceID
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	if referenceType == "order" && referenceID != "" {
		for i := range database.Orders {
			if database.Orders[i].ID == referenceID && database.Orders[i].UserID == userID {
				database.Orders[i].Status = "confirmed"
				break
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment confirmed"})
}

// GetPaymentStatus gets the status of a payment
func GetPaymentStatus(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	paymentID := c.Param("id")

	if database.DB != nil {
		var payment models.Payment
		row := database.DB.QueryRow(
			`SELECT id, user_id, amount, currency, status, payment_method, description, reference_type, reference_id, created_at, updated_at 
			FROM payments WHERE id = $1 AND user_id = $2`,
			paymentID, userID,
		)
		err := row.Scan(&payment.ID, &payment.UserID, &payment.Amount, &payment.Currency, &payment.Status,
			&payment.PaymentMethod, &payment.Description, &payment.ReferenceType, &payment.ReferenceID,
			&payment.CreatedAt, &payment.UpdatedAt)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payment"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"payment": payment})
		return
	}

	// In-memory fallback
	database.Mu.Lock()
	defer database.Mu.Unlock()

	for _, payment := range database.Payments {
		if payment.ID == paymentID && payment.UserID == userID {
			c.JSON(http.StatusOK, gin.H{"payment": payment})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
}

// GetUserPayments gets all payments for a user
func GetUserPayments(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	if database.DB != nil {
		rows, err := database.DB.Query(
			`SELECT id, user_id, amount, currency, status, payment_method, description, reference_type, reference_id, created_at, updated_at 
			FROM payments WHERE user_id = $1 ORDER BY created_at DESC`,
			userID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payments"})
			return
		}
		defer rows.Close()

		var payments []models.Payment
		for rows.Next() {
			var payment models.Payment
			var createdAt, updatedAt time.Time
			err := rows.Scan(&payment.ID, &payment.UserID, &payment.Amount, &payment.Currency, &payment.Status,
				&payment.PaymentMethod, &payment.Description, &payment.ReferenceType, &payment.ReferenceID,
				&createdAt, &updatedAt)
			if err != nil {
				continue
			}
			payment.CreatedAt = createdAt
			payment.UpdatedAt = updatedAt
			payments = append(payments, payment)
		}

		if payments == nil {
			payments = []models.Payment{}
		}
		c.JSON(http.StatusOK, gin.H{"payments": payments})
		return
	}

	// In-memory fallback
	database.Mu.Lock()
	defer database.Mu.Unlock()

	var payments []models.Payment
	for _, payment := range database.Payments {
		if payment.UserID == userID {
			payments = append(payments, payment)
		}
	}

	if payments == nil {
		payments = []models.Payment{}
	}
	c.JSON(http.StatusOK, gin.H{"payments": payments})
}
