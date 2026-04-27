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

// GetTeams returns all teams
func GetTeams(c *gin.Context) {
	if database.DB != nil {
		rows, err := database.DB.Query(`SELECT id, name, country, followers_count, created_at FROM teams ORDER BY name`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		teams := make([]models.Team, 0)
		for rows.Next() {
			var t models.Team
			if err := rows.Scan(&t.ID, &t.Name, &t.Country, &t.FollowersCount, &t.CreatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			// load drivers
			drvRows, err := database.DB.Query(`SELECT driver_name FROM team_drivers WHERE team_id = $1 ORDER BY driver_name`, t.ID)
			if err == nil {
				for drvRows.Next() {
					var d string
					drvRows.Scan(&d)
					t.Drivers = append(t.Drivers, d)
				}
				drvRows.Close()
			}
			// load followers (user ids)
			folRows, err := database.DB.Query(`SELECT user_id::text FROM team_followers WHERE team_id = $1 ORDER BY created_at`, t.ID)
			if err == nil {
				for folRows.Next() {
					var u string
					folRows.Scan(&u)
					t.Followers = append(t.Followers, u)
				}
				folRows.Close()
			}
			teams = append(teams, t)
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"teams": teams, "count": len(teams)})
		return
	}

	database.Mu.RLock()
	defer database.Mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{"teams": database.Teams, "count": len(database.Teams)})
}

// GetTeamByID returns a single team by ID
func GetTeamByID(c *gin.Context) {
	id := c.Param("id")
	if database.DB != nil {
		var t models.Team
		row := database.DB.QueryRow(`SELECT id, name, country, followers_count, created_at FROM teams WHERE id = $1`, id)
		if err := row.Scan(&t.ID, &t.Name, &t.Country, &t.FollowersCount, &t.CreatedAt); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// drivers
		drvRows, err := database.DB.Query(`SELECT driver_name FROM team_drivers WHERE team_id = $1 ORDER BY driver_name`, t.ID)
		if err == nil {
			for drvRows.Next() {
				var d string
				drvRows.Scan(&d)
				t.Drivers = append(t.Drivers, d)
			}
			drvRows.Close()
		}
		// followers
		folRows, err := database.DB.Query(`SELECT user_id::text FROM team_followers WHERE team_id = $1 ORDER BY created_at`, t.ID)
		if err == nil {
			for folRows.Next() {
				var u string
				folRows.Scan(&u)
				t.Followers = append(t.Followers, u)
			}
			folRows.Close()
		}
		c.JSON(http.StatusOK, t)
		return
	}

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
	if database.DB != nil {
		_, err := database.DB.Exec(`INSERT INTO teams (id, name, country, followers_count, created_at) VALUES ($1,$2,$3,$4,$5)`, team.ID, team.Name, team.Country, team.FollowersCount, team.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// insert drivers
		for _, d := range team.Drivers {
			_, _ = database.DB.Exec(`INSERT INTO team_drivers (team_id, driver_name) VALUES ($1,$2) ON CONFLICT DO NOTHING`, team.ID, d)
		}
		c.JSON(http.StatusCreated, team)
		return
	}

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

	if database.DB != nil {
		// update core fields
		res, err := database.DB.Exec(`UPDATE teams SET name=$1, country=$2 WHERE id=$3`, updated.Name, updated.Country, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
			return
		}
		// replace drivers: simple approach - delete existing, insert new
		_, _ = database.DB.Exec(`DELETE FROM team_drivers WHERE team_id = $1`, id)
		for _, d := range updated.Drivers {
			_, _ = database.DB.Exec(`INSERT INTO team_drivers (team_id, driver_name) VALUES ($1,$2) ON CONFLICT DO NOTHING`, id, d)
		}
		// return updated team
		var t models.Team
		row := database.DB.QueryRow(`SELECT id, name, country, followers_count, created_at FROM teams WHERE id = $1`, id)
		if err := row.Scan(&t.ID, &t.Name, &t.Country, &t.FollowersCount, &t.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		drvRows, _ := database.DB.Query(`SELECT driver_name FROM team_drivers WHERE team_id = $1 ORDER BY driver_name`, id)
		if drvRows != nil {
			for drvRows.Next() {
				var d string
				drvRows.Scan(&d)
				t.Drivers = append(t.Drivers, d)
			}
			drvRows.Close()
		}
		folRows, _ := database.DB.Query(`SELECT user_id::text FROM team_followers WHERE team_id = $1 ORDER BY created_at`, id)
		if folRows != nil {
			for folRows.Next() {
				var u string
				folRows.Scan(&u)
				t.Followers = append(t.Followers, u)
			}
			folRows.Close()
		}
		c.JSON(http.StatusOK, t)
		return
	}

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
	if database.DB != nil {
		res, err := database.DB.Exec(`DELETE FROM teams WHERE id = $1`, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Team deleted"})
		return
	}

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
	uid := userID.(string)
	if database.DB != nil {
		// insert follower (ignore if already exists)
		_, err := database.DB.Exec(`INSERT INTO team_followers (team_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, id, uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// recompute followers_count
		_, _ = database.DB.Exec(`UPDATE teams SET followers_count = (SELECT COUNT(*) FROM team_followers WHERE team_id = $1) WHERE id = $1`, id)

		// return updated team
		var t models.Team
		row := database.DB.QueryRow(`SELECT id, name, country, followers_count, created_at FROM teams WHERE id = $1`, id)
		if err := row.Scan(&t.ID, &t.Name, &t.Country, &t.FollowersCount, &t.CreatedAt); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		folRows, _ := database.DB.Query(`SELECT user_id::text FROM team_followers WHERE team_id = $1 ORDER BY created_at`, id)
		if folRows != nil {
			for folRows.Next() {
				var u string
				folRows.Scan(&u)
				t.Followers = append(t.Followers, u)
			}
			folRows.Close()
		}
		drvRows, _ := database.DB.Query(`SELECT driver_name FROM team_drivers WHERE team_id = $1 ORDER BY driver_name`, id)
		if drvRows != nil {
			for drvRows.Next() {
				var d string
				drvRows.Scan(&d)
				t.Drivers = append(t.Drivers, d)
			}
			drvRows.Close()
		}

		c.JSON(http.StatusOK, gin.H{"message": "Team followed", "team": t})
		return
	}

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
	if database.DB != nil {
		res, err := database.DB.Exec(`DELETE FROM team_followers WHERE team_id = $1 AND user_id = $2`, id, uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Not following"})
			return
		}
		// recompute followers_count
		_, _ = database.DB.Exec(`UPDATE teams SET followers_count = (SELECT COUNT(*) FROM team_followers WHERE team_id = $1) WHERE id = $1`, id)

		// return updated team
		var t models.Team
		row := database.DB.QueryRow(`SELECT id, name, country, followers_count, created_at FROM teams WHERE id = $1`, id)
		if err := row.Scan(&t.ID, &t.Name, &t.Country, &t.FollowersCount, &t.CreatedAt); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		folRows, _ := database.DB.Query(`SELECT user_id::text FROM team_followers WHERE team_id = $1 ORDER BY created_at`, id)
		if folRows != nil {
			for folRows.Next() {
				var u string
				folRows.Scan(&u)
				t.Followers = append(t.Followers, u)
			}
			folRows.Close()
		}
		drvRows, _ := database.DB.Query(`SELECT driver_name FROM team_drivers WHERE team_id = $1 ORDER BY driver_name`, id)
		if drvRows != nil {
			for drvRows.Next() {
				var d string
				drvRows.Scan(&d)
				t.Drivers = append(t.Drivers, d)
			}
			drvRows.Close()
		}

		c.JSON(http.StatusOK, gin.H{"message": "Team unfollowed", "team": t})
		return
	}

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
