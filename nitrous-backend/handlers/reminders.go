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

// SetReminder creates a reminder for the authenticated user.
func SetReminder(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.SetReminderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.RemindAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reminder time must be in the future"})
		return
	}

	reminder := models.Reminder{
		ID:        uuid.New().String(),
		UserID:    userID.(string),
		EventID:   req.EventID,
		Message:   req.Message,
		RemindAt:  req.RemindAt,
		CreatedAt: time.Now(),
	}

	if database.DB != nil {
		// Check if event exists
		var exists bool
		err := database.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM events WHERE id=$1)`, req.EventID).Scan(&exists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}

		_, err = database.DB.Exec(`INSERT INTO reminders (id, user_id, event_id, message, remind_at, created_at) VALUES ($1,$2,$3,$4,$5,$6)`, reminder.ID, reminder.UserID, reminder.EventID, reminder.Message, reminder.RemindAt, reminder.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, reminder)
		return
	}

	database.Mu.RLock()
	eventExists := false
	for _, event := range database.Events {
		if event.ID == req.EventID {
			eventExists = true
			break
		}
	}
	database.Mu.RUnlock()

	if !eventExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	database.Mu.Lock()
	database.Reminders = append(database.Reminders, reminder)
	database.Mu.Unlock()

	c.JSON(http.StatusCreated, reminder)
}

// DeleteReminder deletes one reminder owned by the authenticated user.
func DeleteReminder(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	reminderID := c.Param("id")

	if database.DB != nil {
		// Check ownership first
		var ownerID string
		row := database.DB.QueryRow(`SELECT user_id::text FROM reminders WHERE id=$1`, reminderID)
		if err := row.Scan(&ownerID); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Reminder not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if ownerID != userID.(string) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		_, err := database.DB.Exec(`DELETE FROM reminders WHERE id=$1`, reminderID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Reminder deleted"})
		return
	}

	database.Mu.Lock()
	defer database.Mu.Unlock()

	for i, reminder := range database.Reminders {
		if reminder.ID == reminderID {
			if reminder.UserID != userID.(string) {
				c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
				return
			}

			database.Reminders = append(database.Reminders[:i], database.Reminders[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"message": "Reminder deleted"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Reminder not found"})
}

// GetMyReminders returns all reminders for the authenticated user.
func GetMyReminders(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if database.DB != nil {
		rows, err := database.DB.Query(`SELECT id, user_id::text, event_id::text, COALESCE(message, ''), remind_at, created_at FROM reminders WHERE user_id=$1 ORDER BY remind_at`, userID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		reminders := make([]models.Reminder, 0)
		for rows.Next() {
			var r models.Reminder
			if err := rows.Scan(&r.ID, &r.UserID, &r.EventID, &r.Message, &r.RemindAt, &r.CreatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			reminders = append(reminders, r)
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"reminders": reminders, "count": len(reminders)})
		return
	}

	database.Mu.RLock()
	defer database.Mu.RUnlock()

	var reminders []models.Reminder
	for _, reminder := range database.Reminders {
		if reminder.UserID == userID.(string) {
			reminders = append(reminders, reminder)
		}
	}

	c.JSON(http.StatusOK, gin.H{"reminders": reminders, "count": len(reminders)})
}
