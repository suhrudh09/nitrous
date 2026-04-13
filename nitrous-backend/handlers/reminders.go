package handlers

import (
	"net/http"
	"nitrous-backend/database"
	"nitrous-backend/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func findEventByID(eventID string) (models.Event, bool) {
	for _, event := range database.Events {
		if event.ID == eventID {
			return event, true
		}
	}
	return models.Event{}, false
}

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

	database.Mu.RLock()
	_, eventExists := findEventByID(req.EventID)
	database.Mu.RUnlock()

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

	database.Mu.Lock()
	database.Reminders = append(database.Reminders, reminder)
	database.Mu.Unlock()

	c.JSON(http.StatusCreated, reminder)
}

// SetEventReminderCompat supports the frontend reminder shortcut route.
func SetEventReminderCompat(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	eventID := c.Param("id")

	database.Mu.RLock()
	event, found := findEventByID(eventID)
	database.Mu.RUnlock()
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	remindAt := event.Date.Add(-1 * time.Hour)
	if remindAt.Before(time.Now()) {
		remindAt = time.Now().Add(30 * time.Minute)
	}

	database.Mu.Lock()
	defer database.Mu.Unlock()

	for _, reminder := range database.Reminders {
		if reminder.UserID == userID.(string) && reminder.EventID == eventID {
			c.JSON(http.StatusOK, gin.H{"message": "Reminder already set"})
			return
		}
	}

	reminder := models.Reminder{
		ID:        uuid.New().String(),
		UserID:    userID.(string),
		EventID:   eventID,
		Message:   "Reminder set",
		RemindAt:  remindAt,
		CreatedAt: time.Now(),
	}
	database.Reminders = append(database.Reminders, reminder)

	c.JSON(http.StatusCreated, gin.H{"message": "Reminder set"})
}

// DeleteEventReminderCompat deletes a reminder by event ID for the current user.
func DeleteEventReminderCompat(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	eventID := c.Param("id")

	database.Mu.Lock()
	defer database.Mu.Unlock()

	for i, reminder := range database.Reminders {
		if reminder.UserID == userID.(string) && reminder.EventID == eventID {
			database.Reminders = append(database.Reminders[:i], database.Reminders[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"message": "Reminder deleted"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Reminder not found"})
}

// DeleteReminder deletes one reminder owned by the authenticated user.
func DeleteReminder(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	reminderID := c.Param("id")
	database.Mu.Lock()

	for i, reminder := range database.Reminders {
		if reminder.ID == reminderID {
			if reminder.UserID != userID.(string) {
				database.Mu.Unlock()
				c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
				return
			}

			database.Reminders = append(database.Reminders[:i], database.Reminders[i+1:]...)
			database.Mu.Unlock()
			c.JSON(http.StatusOK, gin.H{"message": "Reminder deleted"})
			return
		}
	}
	database.Mu.Unlock()

	c.JSON(http.StatusNotFound, gin.H{"error": "Reminder not found"})
}

// GetMyReminders returns all reminders for the authenticated user.
func GetMyReminders(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
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

	c.JSON(http.StatusOK, gin.H{
		"reminders": reminders,
		"count":     len(reminders),
	})
}
