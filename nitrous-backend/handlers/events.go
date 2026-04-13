package handlers

import (
	"net/http"
	"nitrous-backend/database"
	"nitrous-backend/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetEvents returns all events
func GetEvents(c *gin.Context) {
	category := c.Query("category") // Optional filter by category

	database.Mu.RLock()
	defer database.Mu.RUnlock()

	var filteredEvents []models.Event

	if category != "" {
		for _, event := range database.Events {
			if event.Category == category {
				filteredEvents = append(filteredEvents, event)
			}
		}
	} else {
		// Copy all events to avoid data race
		filteredEvents = make([]models.Event, len(database.Events))
		copy(filteredEvents, database.Events)
	}

	c.JSON(http.StatusOK, gin.H{
		"events": filteredEvents,
		"count":  len(filteredEvents),
	})
}

// GetLiveEvents returns only live events
func GetLiveEvents(c *gin.Context) {
	database.Mu.RLock()
	defer database.Mu.RUnlock()

	var liveEvents []models.Event

	for _, event := range database.Events {
		if event.IsLive {
			liveEvents = append(liveEvents, event)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"events": liveEvents,
		"count":  len(liveEvents),
	})
}

// GetEventByID returns a single event
func GetEventByID(c *gin.Context) {
	id := c.Param("id")

	database.Mu.RLock()
	defer database.Mu.RUnlock()

	for _, event := range database.Events {
		if event.ID == id {
			c.JSON(http.StatusOK, event)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
}

// CreateEvent creates a new event (admin only)
func CreateEvent(c *gin.Context) {
	var newEvent models.Event

	if err := c.ShouldBindJSON(&newEvent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newEvent.ID = uuid.New().String()
	newEvent.CreatedAt = time.Now()

	database.Mu.Lock()
	defer database.Mu.Unlock()

	database.Events = append(database.Events, newEvent)

	c.JSON(http.StatusCreated, newEvent)
}

// UpdateEvent updates an existing event
func UpdateEvent(c *gin.Context) {
	id := c.Param("id")

	var updatedEvent models.Event
	if err := c.ShouldBindJSON(&updatedEvent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	database.Mu.Lock()
	defer database.Mu.Unlock()

	for i, event := range database.Events {
		if event.ID == id {
			updatedEvent.ID = id
			updatedEvent.CreatedAt = event.CreatedAt
			database.Events[i] = updatedEvent
			c.JSON(http.StatusOK, updatedEvent)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
}

// DeleteEvent deletes an event
func DeleteEvent(c *gin.Context) {
	id := c.Param("id")

	database.Mu.Lock()
	defer database.Mu.Unlock()

	for i, event := range database.Events {
		if event.ID == id {
			database.Events = append(database.Events[:i], database.Events[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"message": "Event deleted"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
}
