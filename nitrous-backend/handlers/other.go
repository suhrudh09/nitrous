package handlers

import (
	"net/http"
	"nitrous-backend/database"
	"nitrous-backend/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetCategories returns all categories
func GetCategories(c *gin.Context) {
	database.Mu.RLock()
	defer database.Mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"categories": database.Categories,
		"count":      len(database.Categories),
	})
}

// GetCategoryBySlug returns a single category by slug
func GetCategoryBySlug(c *gin.Context) {
	slug := c.Param("slug")
	
	database.Mu.RLock()
	defer database.Mu.RUnlock()

	for _, category := range database.Categories {
		if category.Slug == slug {
			c.JSON(http.StatusOK, category)
			return
		}
	}
	
	c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
}

// CreateCategory creates a new category (admin only)
func CreateCategory(c *gin.Context) {
	var category models.Category

	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category.ID = uuid.New().String()
	database.Mu.Lock()
	defer database.Mu.Unlock()

	database.Categories = append(database.Categories, category)

	c.JSON(http.StatusCreated, category)
}

// UpdateCategory updates an existing category by slug (admin only)
func UpdateCategory(c *gin.Context) {
	slug := c.Param("slug")

	var updated models.Category
	if err := c.ShouldBindJSON(&updated); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	database.Mu.Lock()
	defer database.Mu.Unlock()

	for i, category := range database.Categories {
		if category.Slug == slug {
			updated.ID = category.ID
			updated.Slug = slug
			database.Categories[i] = updated
			c.JSON(http.StatusOK, updated)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
}

// DeleteCategory deletes a category by slug (admin only)
func DeleteCategory(c *gin.Context) {
	slug := c.Param("slug")

	database.Mu.Lock()
	defer database.Mu.Unlock()

	for i, category := range database.Categories {
		if category.Slug == slug {
			database.Categories = append(database.Categories[:i], database.Categories[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"message": "Category deleted"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
}

// GetJourneys returns all journeys
func GetJourneys(c *gin.Context) {
	database.Mu.RLock()
	defer database.Mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"journeys": database.Journeys,
		"count":    len(database.Journeys),
	})
}

// GetJourneyByID returns a single journey
func GetJourneyByID(c *gin.Context) {
	id := c.Param("id")
	
	database.Mu.RLock()
	defer database.Mu.RUnlock()

	for _, journey := range database.Journeys {
		if journey.ID == id {
			c.JSON(http.StatusOK, journey)
			return
		}
	}
	
	c.JSON(http.StatusNotFound, gin.H{"error": "Journey not found"})
}

// CreateJourney creates a new journey in the catalog (admin only)
func CreateJourney(c *gin.Context) {
	var journey models.Journey

	if err := c.ShouldBindJSON(&journey); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	journey.ID = uuid.New().String()
	if journey.Date.IsZero() {
		journey.Date = time.Now().Add(24 * time.Hour)
	}

	database.Mu.Lock()
	defer database.Mu.Unlock()

	database.Journeys = append(database.Journeys, journey)
	c.JSON(http.StatusCreated, journey)
}

// UpdateJourney updates a journey in the catalog (admin only)
func UpdateJourney(c *gin.Context) {
	id := c.Param("id")

	var updated models.Journey
	if err := c.ShouldBindJSON(&updated); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	database.Mu.Lock()
	defer database.Mu.Unlock()

	for i, journey := range database.Journeys {
		if journey.ID == id {
			updated.ID = id
			database.Journeys[i] = updated
			c.JSON(http.StatusOK, updated)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Journey not found"})
}

// DeleteJourney deletes a journey from the catalog (admin only)
func DeleteJourney(c *gin.Context) {
	id := c.Param("id")

	database.Mu.Lock()
	defer database.Mu.Unlock()

	for i, journey := range database.Journeys {
		if journey.ID == id {
			database.Journeys = append(database.Journeys[:i], database.Journeys[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"message": "Journey deleted"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Journey not found"})
}

// BookJourney handles journey booking
func BookJourney(c *gin.Context) {
	id := c.Param("id")
	
	database.Mu.Lock()
	defer database.Mu.Unlock()

	for i, journey := range database.Journeys {
		if journey.ID == id {
			if journey.SlotsLeft <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "No slots available"})
				return
			}
			
			database.Journeys[i].SlotsLeft--
			
			c.JSON(http.StatusOK, gin.H{
				"message": "Journey booked successfully",
				"journey": database.Journeys[i],
			})
			return
		}
	}
	
	c.JSON(http.StatusNotFound, gin.H{"error": "Journey not found"})
}

// GetMerchItems returns all merch items
func GetMerchItems(c *gin.Context) {
	database.Mu.RLock()
	defer database.Mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"items": database.MerchItems,
		"count": len(database.MerchItems),
	})
}

// GetMerchItemByID returns a single merch item
func GetMerchItemByID(c *gin.Context) {
	id := c.Param("id")
	
	database.Mu.RLock()
	defer database.Mu.RUnlock()

	for _, item := range database.MerchItems {
		if item.ID == id {
			c.JSON(http.StatusOK, item)
			return
		}
	}
	
	c.JSON(http.StatusNotFound, gin.H{"error": "Merch item not found"})
}
