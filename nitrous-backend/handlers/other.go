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

// GetCategories returns all categories
func GetCategories(c *gin.Context) {
	if database.DB != nil {
		rows, err := database.DB.Query(`SELECT id, name, slug, icon, live_count, description, color FROM categories ORDER BY name`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		cats := make([]models.Category, 0)
		for rows.Next() {
			var cat models.Category
			if err := rows.Scan(&cat.ID, &cat.Name, &cat.Slug, &cat.Icon, &cat.LiveCount, &cat.Description, &cat.Color); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			cats = append(cats, cat)
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"categories": cats, "count": len(cats)})
		return
	}

	database.Mu.RLock()
	defer database.Mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{"categories": database.Categories, "count": len(database.Categories)})
}

// GetCategoryBySlug returns a single category by slug
func GetCategoryBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if database.DB != nil {
		var cat models.Category
		row := database.DB.QueryRow(`SELECT id, name, slug, icon, live_count, description, color FROM categories WHERE slug = $1`, slug)
		if err := row.Scan(&cat.ID, &cat.Name, &cat.Slug, &cat.Icon, &cat.LiveCount, &cat.Description, &cat.Color); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, cat)
		return
	}

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
	if database.DB != nil {
		_, err := database.DB.Exec(`INSERT INTO categories (id, name, slug, icon, live_count, description, color) VALUES ($1,$2,$3,$4,$5,$6,$7)`, category.ID, category.Name, category.Slug, category.Icon, category.LiveCount, category.Description, category.Color)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, category)
		return
	}

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
	if database.DB != nil {
		res, err := database.DB.Exec(`UPDATE categories SET name=$1, icon=$2, live_count=$3, description=$4, color=$5 WHERE slug=$6`, updated.Name, updated.Icon, updated.LiveCount, updated.Description, updated.Color, slug)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
		}
		// Return updated row
		var cat models.Category
		row := database.DB.QueryRow(`SELECT id, name, slug, icon, live_count, description, color FROM categories WHERE slug = $1`, slug)
		if err := row.Scan(&cat.ID, &cat.Name, &cat.Slug, &cat.Icon, &cat.LiveCount, &cat.Description, &cat.Color); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, cat)
		return
	}

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
	if database.DB != nil {
		res, err := database.DB.Exec(`DELETE FROM categories WHERE slug = $1`, slug)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Category deleted"})
		return
	}

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
	if database.DB != nil {
		rows, err := database.DB.Query(`SELECT id, title, category, description, COALESCE(badge, ''), slots_left, date, price::float8, COALESCE(thumbnail_url, '') FROM journeys ORDER BY date`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		js := make([]models.Journey, 0)
		for rows.Next() {
			var j models.Journey
			if err := rows.Scan(&j.ID, &j.Title, &j.Category, &j.Description, &j.Badge, &j.SlotsLeft, &j.Date, &j.Price, &j.ThumbnailURL); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			js = append(js, j)
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"journeys": js, "count": len(js)})
		return
	}

	database.Mu.RLock()
	defer database.Mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{"journeys": database.Journeys, "count": len(database.Journeys)})
}

// GetJourneyByID returns a single journey
func GetJourneyByID(c *gin.Context) {
	id := c.Param("id")
	if database.DB != nil {
		var j models.Journey
		row := database.DB.QueryRow(`SELECT id, title, category, description, COALESCE(badge, ''), slots_left, date, price::float8, COALESCE(thumbnail_url, '') FROM journeys WHERE id = $1`, id)
		if err := row.Scan(&j.ID, &j.Title, &j.Category, &j.Description, &j.Badge, &j.SlotsLeft, &j.Date, &j.Price, &j.ThumbnailURL); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Journey not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, j)
		return
	}

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

	if database.DB != nil {
		_, err := database.DB.Exec(`INSERT INTO journeys (id, title, category, description, badge, slots_left, date, price, thumbnail_url) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, journey.ID, journey.Title, journey.Category, journey.Description, journey.Badge, journey.SlotsLeft, journey.Date, journey.Price, journey.ThumbnailURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, journey)
		return
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
	if database.DB != nil {
		res, err := database.DB.Exec(`UPDATE journeys SET title=$1, category=$2, description=$3, badge=$4, slots_left=$5, date=$6, price=$7, thumbnail_url=$8 WHERE id=$9`, updated.Title, updated.Category, updated.Description, updated.Badge, updated.SlotsLeft, updated.Date, updated.Price, updated.ThumbnailURL, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Journey not found"})
			return
		}
		var j models.Journey
		row := database.DB.QueryRow(`SELECT id, title, category, description, COALESCE(badge, ''), slots_left, date, price::float8, COALESCE(thumbnail_url, '') FROM journeys WHERE id = $1`, id)
		if err := row.Scan(&j.ID, &j.Title, &j.Category, &j.Description, &j.Badge, &j.SlotsLeft, &j.Date, &j.Price, &j.ThumbnailURL); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, j)
		return
	}

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
	if database.DB != nil {
		res, err := database.DB.Exec(`DELETE FROM journeys WHERE id = $1`, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Journey not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Journey deleted"})
		return
	}

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
	if database.DB != nil {
		// Attempt to decrement slots atomically
		res, err := database.DB.Exec(`UPDATE journeys SET slots_left = slots_left - 1 WHERE id = $1 AND slots_left > 0`, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			// check if journey exists
			var exists bool
			err := database.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM journeys WHERE id=$1)`, id).Scan(&exists)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "Journey not found"})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": "No slots available"})
			return
		}
		var j models.Journey
		row := database.DB.QueryRow(`SELECT id, title, category, description, COALESCE(badge, ''), slots_left, date, price::float8, COALESCE(thumbnail_url, '') FROM journeys WHERE id = $1`, id)
		if err := row.Scan(&j.ID, &j.Title, &j.Category, &j.Description, &j.Badge, &j.SlotsLeft, &j.Date, &j.Price, &j.ThumbnailURL); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Journey booked successfully", "journey": j})
		return
	}

	database.Mu.Lock()
	defer database.Mu.Unlock()

	for i, journey := range database.Journeys {
		if journey.ID == id {
			if journey.SlotsLeft <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "No slots available"})
				return
			}

			database.Journeys[i].SlotsLeft--

			c.JSON(http.StatusOK, gin.H{"message": "Journey booked successfully", "journey": database.Journeys[i]})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Journey not found"})
}

// GetMerchItems returns all merch items
func GetMerchItems(c *gin.Context) {
	if database.DB != nil {
		rows, err := database.DB.Query(`SELECT id, name, icon, price::float8, category FROM merch_items ORDER BY name`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		items := make([]models.MerchItem, 0)
		for rows.Next() {
			var it models.MerchItem
			if err := rows.Scan(&it.ID, &it.Name, &it.Icon, &it.Price, &it.Category); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			items = append(items, it)
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": items, "count": len(items)})
		return
	}

	database.Mu.RLock()
	defer database.Mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{"items": database.MerchItems, "count": len(database.MerchItems)})
}

// GetMerchItemByID returns a single merch item
func GetMerchItemByID(c *gin.Context) {
	id := c.Param("id")
	if database.DB != nil {
		var it models.MerchItem
		row := database.DB.QueryRow(`SELECT id, name, icon, price::float8, category FROM merch_items WHERE id = $1`, id)
		if err := row.Scan(&it.ID, &it.Name, &it.Icon, &it.Price, &it.Category); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Merch item not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, it)
		return
	}

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
