package handlers

import (
	"net/http"
	"nitrous-backend/database"
	"nitrous-backend/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func SetReminder(c *gin.Context) {
	eventID := c.Param("id")
	userID := c.GetString("userID")

	if _, found := database.FindEventByID(eventID); !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	if database.ReminderExists(userID, eventID) {
		c.JSON(http.StatusConflict, gin.H{"error": "Reminder already set for this event"})
		return
	}

	r := models.Reminder{
		ID:        uuid.New().String(),
		UserID:    userID,
		EventID:   eventID,
		CreatedAt: time.Now(),
	}
	database.AppendReminder(r)

	c.JSON(http.StatusCreated, gin.H{"message": "Reminder set", "reminder": r})
}

func DeleteReminder(c *gin.Context) {
	if !database.DeleteReminder(c.GetString("userID"), c.Param("id")) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Reminder not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Reminder removed"})
}

func GetMyReminders(c *gin.Context) {
	rs := database.GetRemindersByUser(c.GetString("userID"))
	c.JSON(http.StatusOK, gin.H{"reminders": rs, "count": len(rs)})
}
