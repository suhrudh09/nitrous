package handlers

import (
	"net/http"
	"nitrous-backend/database"

	"github.com/gin-gonic/gin"
)

// ── Categories ────────────────────────────────────────────────────────────────

func GetCategories(c *gin.Context) {
	cats := database.GetCategories()
	c.JSON(http.StatusOK, gin.H{"categories": cats, "count": len(cats)})
}

func GetCategoryBySlug(c *gin.Context) {
	cat, found := database.FindCategoryBySlug(c.Param("slug"))
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}
	c.JSON(http.StatusOK, cat)
}

// ── Journeys ──────────────────────────────────────────────────────────────────

func GetJourneys(c *gin.Context) {
	js := database.GetJourneys()
	c.JSON(http.StatusOK, gin.H{"journeys": js, "count": len(js)})
}

func GetJourneyByID(c *gin.Context) {
	j, found := database.FindJourneyByID(c.Param("id"))
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Journey not found"})
		return
	}
	c.JSON(http.StatusOK, j)
}

func BookJourney(c *gin.Context) {
	j, ok := database.BookJourney(c.Param("id"))
	if !ok {
		// ok=false covers both "not found" and "sold out"
		if j.ID == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Journey not found"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No slots available"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Journey booked successfully", "journey": j})
}

// ── Merch ─────────────────────────────────────────────────────────────────────

func GetMerchItems(c *gin.Context) {
	items := database.GetMerchItems()
	c.JSON(http.StatusOK, gin.H{"items": items, "count": len(items)})
}

func GetMerchItemByID(c *gin.Context) {
	item, found := database.FindMerchByID(c.Param("id"))
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Merch item not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}
