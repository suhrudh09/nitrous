package handlers

import (
	"net/http"
	"nitrous-backend/database"
	"nitrous-backend/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetEvents(c *gin.Context) {
	all := database.GetEvents()
	category := c.Query("category")

	if category == "" {
		c.JSON(http.StatusOK, gin.H{"events": all, "count": len(all)})
		return
	}

	var filtered []models.Event
	for _, e := range all {
		if e.Category == category {
			filtered = append(filtered, e)
		}
	}
	c.JSON(http.StatusOK, gin.H{"events": filtered, "count": len(filtered)})
}

func GetLiveEvents(c *gin.Context) {
	all := database.GetEvents()
	var live []models.Event
	for _, e := range all {
		if e.IsLive {
			live = append(live, e)
		}
	}
	c.JSON(http.StatusOK, gin.H{"events": live, "count": len(live)})
}

func GetEventByID(c *gin.Context) {
	event, found := database.FindEventByID(c.Param("id"))
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}
	c.JSON(http.StatusOK, event)
}

func CreateEvent(c *gin.Context) {
	var e models.Event
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	e.ID = uuid.New().String()
	e.CreatedAt = time.Now()
	database.AppendEvent(e)
	c.JSON(http.StatusCreated, e)
}

func UpdateEvent(c *gin.Context) {
	var updated models.Event
	if err := c.ShouldBindJSON(&updated); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !database.UpdateEvent(c.Param("id"), updated) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func DeleteEvent(c *gin.Context) {
	if !database.DeleteEvent(c.Param("id")) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Event deleted"})
}
