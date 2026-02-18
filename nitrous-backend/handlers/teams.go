package handlers

import (
	"net/http"
	"nitrous-backend/database"

	"github.com/gin-gonic/gin"
)

// GetTeams returns all teams
func GetTeams(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"teams": database.Teams,
		"count": len(database.Teams),
	})
}

// GetTeamByID returns a single team by ID
func GetTeamByID(c *gin.Context) {
	id := c.Param("id")

	for _, team := range database.Teams {
		if team.ID == id {
			c.JSON(http.StatusOK, team)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
}

// FollowTeam adds the authenticated user to the team's followers
func FollowTeam(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id := c.Param("id")

	for i, team := range database.Teams {
		if team.ID == id {
			// check if already following
			uid := userID.(string)
			for _, f := range team.Followers {
				if f == uid {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Already following"})
					return
				}
			}

			database.Teams[i].Followers = append(database.Teams[i].Followers, uid)
			database.Teams[i].FollowersCount = len(database.Teams[i].Followers)

			c.JSON(http.StatusOK, gin.H{"message": "Team followed", "team": database.Teams[i]})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
}

// UnfollowTeam removes the authenticated user from the team's followers
func UnfollowTeam(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id := c.Param("id")
	uid := userID.(string)

	for i, team := range database.Teams {
		if team.ID == id {
			// find follower index
			idx := -1
			for j, f := range team.Followers {
				if f == uid {
					idx = j
					break
				}
			}

			if idx == -1 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Not following"})
				return
			}

			// remove follower
			database.Teams[i].Followers = append(database.Teams[i].Followers[:idx], database.Teams[i].Followers[idx+1:]...)
			database.Teams[i].FollowersCount = len(database.Teams[i].Followers)

			c.JSON(http.StatusOK, gin.H{"message": "Team unfollowed", "team": database.Teams[i]})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
}
