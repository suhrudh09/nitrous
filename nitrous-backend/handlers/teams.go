package handlers

import (
	"net/http"
	"nitrous-backend/database"
	"nitrous-backend/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetTeams(c *gin.Context) {
	ts := database.GetTeams()
	c.JSON(http.StatusOK, gin.H{"teams": ts, "count": len(ts)})
}

func GetTeamByID(c *gin.Context) {
	team, found := database.FindTeamByID(c.Param("id"))
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}
	c.JSON(http.StatusOK, team)
}

func FollowTeam(c *gin.Context) {
	teamID := c.Param("id")
	userID := c.GetString("userID")

	if _, found := database.FindTeamByID(teamID); !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	if database.FollowExists(userID, teamID) {
		c.JSON(http.StatusConflict, gin.H{"error": "Already following this team"})
		return
	}

	database.AppendFollow(models.TeamFollow{
		ID:        uuid.New().String(),
		UserID:    userID,
		TeamID:    teamID,
		CreatedAt: time.Now(),
	})

	following, _ := database.IncrementFollowing(teamID)
	c.JSON(http.StatusOK, gin.H{"message": "Now following team", "following": following})
}

func UnfollowTeam(c *gin.Context) {
	teamID := c.Param("id")
	userID := c.GetString("userID")

	if _, found := database.FindTeamByID(teamID); !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	if !database.DeleteFollow(userID, teamID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Follow record not found"})
		return
	}

	following, _ := database.DecrementFollowing(teamID)
	c.JSON(http.StatusOK, gin.H{"message": "Unfollowed team", "following": following})
}
