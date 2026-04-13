package handlers

import (
	"net/http"
	"nitrous-backend/database"
	"nitrous-backend/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetTeams returns all teams
func GetTeams(c *gin.Context) {
	database.Mu.RLock()
	defer database.Mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"teams": database.Teams,
		"count": len(database.Teams),
	})
}

// GetTeamByID returns a single team by ID
func GetTeamByID(c *gin.Context) {
	id := c.Param("id")

	database.Mu.RLock()
	defer database.Mu.RUnlock()

	for _, team := range database.Teams {
		if team.ID == id {
			c.JSON(http.StatusOK, team)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
}

// CreateTeam creates a new team (admin only)
func CreateTeam(c *gin.Context) {
	var team models.Team

	if err := c.ShouldBindJSON(&team); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	team.ID = uuid.New().String()
	team.CreatedAt = time.Now()
	team.Followers = []string{}
	team.FollowersCount = 0

	database.Mu.Lock()
	defer database.Mu.Unlock()

	database.Teams = append(database.Teams, team)
	c.JSON(http.StatusCreated, team)
}

// UpdateTeam updates an existing team (admin only)
func UpdateTeam(c *gin.Context) {
	id := c.Param("id")

	var updated models.Team
	if err := c.ShouldBindJSON(&updated); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	database.Mu.Lock()
	defer database.Mu.Unlock()

	for i, team := range database.Teams {
		if team.ID == id {
			updated.ID = id
			updated.CreatedAt = team.CreatedAt
			updated.Followers = team.Followers
			updated.FollowersCount = team.FollowersCount
			database.Teams[i] = updated
			c.JSON(http.StatusOK, updated)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
}

// DeleteTeam deletes a team (admin only)
func DeleteTeam(c *gin.Context) {
	id := c.Param("id")

	database.Mu.Lock()
	defer database.Mu.Unlock()

	for i, team := range database.Teams {
		if team.ID == id {
			database.Teams = append(database.Teams[:i], database.Teams[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"message": "Team deleted"})
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

	database.Mu.Lock()
	defer database.Mu.Unlock()

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

	database.Mu.Lock()
	defer database.Mu.Unlock()

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
