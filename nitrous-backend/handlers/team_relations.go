package handlers

import (
	"net/http"
	"nitrous-backend/database"
	"nitrous-backend/models"

	"github.com/gin-gonic/gin"
)

// ListTeamMembers returns the list of member user IDs for a team.
func ListTeamMembers(c *gin.Context) {
	teamID := c.Param("id")
	if database.DB != nil {
		rows, err := database.DB.Query(`SELECT user_id::text FROM team_members WHERE team_id = $1 ORDER BY created_at`, teamID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		members := []string{}
		for rows.Next() {
			var u string
			rows.Scan(&u)
			members = append(members, u)
		}
		c.JSON(http.StatusOK, gin.H{"members": members})
		return
	}

	database.Mu.RLock()
	defer database.Mu.RUnlock()
	for _, t := range database.Teams {
		if t.ID == teamID {
			c.JSON(http.StatusOK, gin.H{"members": t.Members})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
}

// AddTeamMember assigns a member to a team. Allowed for admin or existing team managers.
func AddTeamMember(c *gin.Context) {
	teamID := c.Param("id")
	var body struct {
		UserID string `json:"userId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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
				c.JSON(http.StatusForbidden, gin.H{"error": "only admins or team managers can add members"})
				return
			}
		}

		if _, err := database.DB.Exec(`INSERT INTO team_members (team_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, teamID, body.UserID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "member added"})
		return
	}

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
			c.JSON(http.StatusForbidden, gin.H{"error": "only admins or team managers can add members"})
			return
		}
	}

	for _, m := range found.Members {
		if m == body.UserID {
			c.JSON(http.StatusOK, gin.H{"message": "member already present"})
			return
		}
	}
	found.Members = append(found.Members, body.UserID)
	c.JSON(http.StatusOK, gin.H{"message": "member added"})
}

// RemoveTeamMember removes a member from a team. Allowed for admin or existing team managers.
func RemoveTeamMember(c *gin.Context) {
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
				c.JSON(http.StatusForbidden, gin.H{"error": "only admins or team managers can remove members"})
				return
			}
		}

		if _, err := database.DB.Exec(`DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, teamID, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "member removed"})
		return
	}

	database.Mu.Lock()
	defer database.Mu.Unlock()

	for i := range database.Teams {
		if database.Teams[i].ID == teamID {
			if curRole != "admin" {
				ok := false
				for _, m := range database.Teams[i].Managers {
					if m == curUserID {
						ok = true
						break
					}
				}
				if !ok {
					c.JSON(http.StatusForbidden, gin.H{"error": "only admins or team managers can remove members"})
					return
				}
			}
			new := make([]string, 0, len(database.Teams[i].Members))
			for _, m := range database.Teams[i].Members {
				if m != userID {
					new = append(new, m)
				}
			}
			database.Teams[i].Members = new
			c.JSON(http.StatusOK, gin.H{"message": "member removed"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
}

// ListTeamSponsors returns sponsors
func ListTeamSponsors(c *gin.Context) {
	teamID := c.Param("id")
	if database.DB != nil {
		rows, err := database.DB.Query(`SELECT user_id::text FROM team_sponsors WHERE team_id = $1 ORDER BY created_at`, teamID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		sponsors := []string{}
		for rows.Next() {
			var u string
			rows.Scan(&u)
			sponsors = append(sponsors, u)
		}
		c.JSON(http.StatusOK, gin.H{"sponsors": sponsors})
		return
	}

	database.Mu.RLock()
	defer database.Mu.RUnlock()
	for _, t := range database.Teams {
		if t.ID == teamID {
			c.JSON(http.StatusOK, gin.H{"sponsors": t.Sponsors})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
}

// AddTeamSponsor assigns a sponsor to a team. Allowed for admin or existing team managers.
func AddTeamSponsor(c *gin.Context) {
	teamID := c.Param("id")
	var body struct {
		UserID string `json:"userId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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
				c.JSON(http.StatusForbidden, gin.H{"error": "only admins or team managers can add sponsors"})
				return
			}
		}

		if _, err := database.DB.Exec(`INSERT INTO team_sponsors (team_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, teamID, body.UserID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "sponsor added"})
		return
	}

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
			c.JSON(http.StatusForbidden, gin.H{"error": "only admins or team managers can add sponsors"})
			return
		}
	}

	for _, s := range found.Sponsors {
		if s == body.UserID {
			c.JSON(http.StatusOK, gin.H{"message": "sponsor already present"})
			return
		}
	}
	found.Sponsors = append(found.Sponsors, body.UserID)
	c.JSON(http.StatusOK, gin.H{"message": "sponsor added"})
}

// RemoveTeamSponsor removes a sponsor from a team. Allowed for admin or existing team managers.
func RemoveTeamSponsor(c *gin.Context) {
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
				c.JSON(http.StatusForbidden, gin.H{"error": "only admins or team managers can remove sponsors"})
				return
			}
		}

		if _, err := database.DB.Exec(`DELETE FROM team_sponsors WHERE team_id=$1 AND user_id=$2`, teamID, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "sponsor removed"})
		return
	}

	database.Mu.Lock()
	defer database.Mu.Unlock()

	for i := range database.Teams {
		if database.Teams[i].ID == teamID {
			if curRole != "admin" {
				ok := false
				for _, m := range database.Teams[i].Managers {
					if m == curUserID {
						ok = true
						break
					}
				}
				if !ok {
					c.JSON(http.StatusForbidden, gin.H{"error": "only admins or team managers can remove sponsors"})
					return
				}
			}
			new := make([]string, 0, len(database.Teams[i].Sponsors))
			for _, s := range database.Teams[i].Sponsors {
				if s != userID {
					new = append(new, s)
				}
			}
			database.Teams[i].Sponsors = new
			c.JSON(http.StatusOK, gin.H{"message": "sponsor removed"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
}
