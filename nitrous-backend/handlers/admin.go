package handlers

import (
	"net/http"
	"nitrous-backend/database"

	"github.com/gin-gonic/gin"
)

// AdminTriggerSync triggers a one-off sync for external providers (admin only)
func AdminTriggerSync(c *gin.Context) {
	results := map[string]string{}

	if err := syncJolpicaCalendarAndResults(); err != nil {
		results["jolpica"] = err.Error()
	} else {
		results["jolpica"] = "ok"
	}

	if err := syncSportsDBTeams(); err != nil {
		results["sportsdb"] = err.Error()
	} else {
		results["sportsdb"] = "ok"
	}

	if _, err := syncOpenF1LiveData(); err != nil {
		results["openf1"] = err.Error()
	} else {
		results["openf1"] = "ok"
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}

// AdminSetUserRole changes a user's role (admin only)
func AdminSetUserRole(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if database.DB != nil {
		res, err := database.DB.Exec(`UPDATE users SET role=$1 WHERE id=$2`, body.Role, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if ra, _ := res.RowsAffected(); ra == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "role updated"})
		return
	}

	database.Mu.Lock()
	defer database.Mu.Unlock()
	for i := range database.Users {
		if database.Users[i].ID == id {
			database.Users[i].Role = body.Role
			c.JSON(http.StatusOK, gin.H{"message": "role updated"})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
}
