package handlers

import (
	"net/http"
	"nitrous-backend/database"
	"nitrous-backend/models"

	"github.com/gin-gonic/gin"
)

// AddTeamManager assigns a manager to a team. Allowed for admin or existing team managers.
func AddTeamManager(c *gin.Context) {
	teamID := c.Param("id")
	var body struct {
		UserID string `json:"userId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// current user
	curID, _ := c.Get("userID")
	curRole := ""
	if v, ok := c.Get("userRole"); ok {
		curRole = v.(string)
	}
	curUserID := ""
	if curID != nil {
		curUserID = curID.(string)
	}

	if database.DB != nil {
		// ensure team exists
		var exists bool
		if err := database.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM teams WHERE id=$1)`, teamID).Scan(&exists); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
			return
		}

		// if not admin, ensure requester is a manager of the team
		if curRole != "admin" {
			var isMgr bool
			if err := database.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM team_managers WHERE team_id=$1 AND user_id=$2)`, teamID, curUserID).Scan(&isMgr); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if !isMgr {
				c.JSON(http.StatusForbidden, gin.H{"error": "only admins or existing team managers can assign managers"})
				return
			}
		}

		if _, err := database.DB.Exec(`INSERT INTO team_managers (team_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, teamID, body.UserID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "manager assigned"})
		return
	}

	// in-memory fallback
	database.Mu.Lock()
	defer database.Mu.Unlock()

	var found *models.Team
	for i := range database.Teams {
		if database.Teams[i].ID == teamID {
			found = &database.Teams[i]
			break
		}
	}
	if found == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
		return
	}

	if curRole != "admin" {
		ok := false
		for _, m := range found.Managers {
			if m == curUserID {
				ok = true
				break
			}
		}
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "only admins or existing team managers can assign managers"})
			return
		}
	}

	for _, m := range found.Managers {
		if m == body.UserID {
			c.JSON(http.StatusOK, gin.H{"message": "manager already assigned"})
			return
		}
	}
	found.Managers = append(found.Managers, body.UserID)
	c.JSON(http.StatusOK, gin.H{"message": "manager assigned"})
}

// RemoveTeamManager removes a manager from a team. Allowed for admin or existing team managers.
func RemoveTeamManager(c *gin.Context) {
	teamID := c.Param("id")
	userID := c.Param("userId")

	curID, _ := c.Get("userID")
	curRole := ""
	if v, ok := c.Get("userRole"); ok {
		curRole = v.(string)
	}
	curUserID := ""
	if curID != nil {
		curUserID = curID.(string)
	}

	if database.DB != nil {
		// ensure team exists
		var exists bool
		if err := database.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM teams WHERE id=$1)`, teamID).Scan(&exists); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
			return
		}

		if curRole != "admin" {
			var isMgr bool
			if err := database.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM team_managers WHERE team_id=$1 AND user_id=$2)`, teamID, curUserID).Scan(&isMgr); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if !isMgr {
				c.JSON(http.StatusForbidden, gin.H{"error": "only admins or existing team managers can remove managers"})
				return
			}
		}

		if _, err := database.DB.Exec(`DELETE FROM team_managers WHERE team_id=$1 AND user_id=$2`, teamID, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "manager removed"})
		return
	}

	database.Mu.Lock()
	defer database.Mu.Unlock()

	for i := range database.Teams {
		if database.Teams[i].ID == teamID {
			// permission check
			if curRole != "admin" {
				ok := false
				for _, m := range database.Teams[i].Managers {
					if m == curUserID {
						ok = true
						break
					}
				}
				if !ok {
					c.JSON(http.StatusForbidden, gin.H{"error": "only admins or existing team managers can remove managers"})
					return
				}
			}
			// remove
			newMgrs := make([]string, 0, len(database.Teams[i].Managers))
			for _, m := range database.Teams[i].Managers {
				if m != userID {
					newMgrs = append(newMgrs, m)
				}
			}
			database.Teams[i].Managers = newMgrs
			c.JSON(http.StatusOK, gin.H{"message": "manager removed"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
}
