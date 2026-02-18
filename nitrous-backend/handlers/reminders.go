package handlers

import (
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

	eventExists := false
	for _, event := range database.Events {
		if event.ID == req.EventID {
			eventExists = true
			break
		}
	}

	if !eventExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
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

	database.Reminders = append(database.Reminders, reminder)

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

	var reminders []models.Reminder
	for _, reminder := range database.Reminders {
		if reminder.UserID == userID.(string) {
			reminders = append(reminders, reminder)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"reminders": reminders,
		"count":     len(reminders),
	})
}
