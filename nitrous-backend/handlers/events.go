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

// GetEvents returns all events
func GetEvents(c *gin.Context) {
	category := c.Query("category") // Optional filter by category
	// If Postgres is available, query it; otherwise use in-memory slice
	if database.DB != nil {
		var rows *sql.Rows
		var err error
		if category != "" {
			rows, err = database.DB.Query(`SELECT id, title, location, date, COALESCE(time, ''), is_live, category, COALESCE(thumbnail_url, ''), created_at FROM events WHERE category = $1 ORDER BY date`, category)
		} else {
			rows, err = database.DB.Query(`SELECT id, title, location, date, COALESCE(time, ''), is_live, category, COALESCE(thumbnail_url, ''), created_at FROM events ORDER BY date`)
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		events := make([]models.Event, 0)
		for rows.Next() {
			var e models.Event
			if err := rows.Scan(&e.ID, &e.Title, &e.Location, &e.Date, &e.Time, &e.IsLive, &e.Category, &e.ThumbnailURL, &e.CreatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			events = append(events, e)
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"events": events, "count": len(events)})
		return
	}

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

	c.JSON(http.StatusOK, gin.H{"events": filteredEvents, "count": len(filteredEvents)})
}

// GetLiveEvents returns only live events
func GetLiveEvents(c *gin.Context) {
	if database.DB != nil {
		rows, err := database.DB.Query(`SELECT id, title, location, date, COALESCE(time, ''), is_live, category, COALESCE(thumbnail_url, ''), created_at FROM events WHERE is_live = true ORDER BY date`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		events := make([]models.Event, 0)
		for rows.Next() {
			var e models.Event
			if err := rows.Scan(&e.ID, &e.Title, &e.Location, &e.Date, &e.Time, &e.IsLive, &e.Category, &e.ThumbnailURL, &e.CreatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			events = append(events, e)
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"events": events, "count": len(events)})
		return
	}

	database.Mu.RLock()
	defer database.Mu.RUnlock()

	var liveEvents []models.Event

	for _, event := range database.Events {
		if event.IsLive {
			liveEvents = append(liveEvents, event)
		}
	}

	c.JSON(http.StatusOK, gin.H{"events": liveEvents, "count": len(liveEvents)})
}

// GetEventByID returns a single event
func GetEventByID(c *gin.Context) {
	id := c.Param("id")
	if database.DB != nil {
		var e models.Event
		row := database.DB.QueryRow(`SELECT id, title, location, date, COALESCE(time, ''), is_live, category, COALESCE(thumbnail_url, ''), created_at FROM events WHERE id = $1`, id)
		if err := row.Scan(&e.ID, &e.Title, &e.Location, &e.Date, &e.Time, &e.IsLive, &e.Category, &e.ThumbnailURL, &e.CreatedAt); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, e)
		return
	}

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

	if database.DB != nil {
		_, err := database.DB.Exec(`INSERT INTO events (id, title, location, date, time, is_live, category, thumbnail_url, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, newEvent.ID, newEvent.Title, newEvent.Location, newEvent.Date, newEvent.Time, newEvent.IsLive, newEvent.Category, newEvent.ThumbnailURL, newEvent.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, newEvent)
		return
	}

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

	if database.DB != nil {
		// Ensure event exists
		res, err := database.DB.Exec(`UPDATE events SET title=$1, location=$2, date=$3, time=$4, is_live=$5, category=$6, thumbnail_url=$7 WHERE id=$8`, updatedEvent.Title, updatedEvent.Location, updatedEvent.Date, updatedEvent.Time, updatedEvent.IsLive, updatedEvent.Category, updatedEvent.ThumbnailURL, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}
		// Preserve created_at by re-querying
		var e models.Event
		row := database.DB.QueryRow(`SELECT id, title, location, date, COALESCE(time, ''), is_live, category, COALESCE(thumbnail_url, ''), created_at FROM events WHERE id=$1`, id)
		if err := row.Scan(&e.ID, &e.Title, &e.Location, &e.Date, &e.Time, &e.IsLive, &e.Category, &e.ThumbnailURL, &e.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, e)
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

	if database.DB != nil {
		res, err := database.DB.Exec(`DELETE FROM events WHERE id=$1`, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Event deleted"})
		return
	}

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
