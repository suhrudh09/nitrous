package handlers

import (
	"database/sql"
	"net/http"
	"nitrous-backend/database"
	"nitrous-backend/models"
	"nitrous-backend/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetTeams returns all teams
func GetTeams(c *gin.Context) {
	// attempt to resolve optional current user (non-fatal)
	var currentUserID string
	var currentUserRole string
	if v, ok := c.Get("userID"); ok {
		currentUserID = v.(string)
	} else {
		// try optional token parse
		if auth := c.GetHeader("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			token := strings.TrimPrefix(auth, "Bearer ")
			if claims, err := utils.ValidateJWT(token); err == nil {
				currentUserID = claims.UserID
				// try to set role from in-memory users if DB nil
				if database.DB == nil {
					for _, u := range database.Users {
						if u.ID == currentUserID {
							currentUserRole = u.Role
							break
						}
					}
				}
			}
		}
	}

	if v, ok := c.Get("userRole"); ok {
		currentUserRole = v.(string)
	}

	if database.DB != nil {
		rows, err := database.DB.Query(`SELECT id, name, country, followers_count, created_at, is_private FROM teams ORDER BY name`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		teams := make([]models.Team, 0)
		for rows.Next() {
			var t models.Team
			var isPrivate bool
			if err := rows.Scan(&t.ID, &t.Name, &t.Country, &t.FollowersCount, &t.CreatedAt, &isPrivate); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			t.IsPrivate = isPrivate
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
			// load managers, members, sponsors
			mgrRows, _ := database.DB.Query(`SELECT user_id::text FROM team_managers WHERE team_id = $1`, t.ID)
			if mgrRows != nil {
				for mgrRows.Next() {
					var u string
					mgrRows.Scan(&u)
					t.Managers = append(t.Managers, u)
				}
				mgrRows.Close()
			}
			memRows, _ := database.DB.Query(`SELECT user_id::text FROM team_members WHERE team_id = $1`, t.ID)
			if memRows != nil {
				for memRows.Next() {
					var u string
					memRows.Scan(&u)
					t.Members = append(t.Members, u)
				}
				memRows.Close()
			}
			sponRows, _ := database.DB.Query(`SELECT user_id::text FROM team_sponsors WHERE team_id = $1`, t.ID)
			if sponRows != nil {
				for sponRows.Next() {
					var u string
					sponRows.Scan(&u)
					t.Sponsors = append(t.Sponsors, u)
				}
				sponRows.Close()
			}

			// enforce visibility: if private and current user cannot see, skip
			if t.IsPrivate {
				allowed := false
				if currentUserRole == "admin" {
					allowed = true
				}
				if currentUserID != "" {
					for _, u := range t.Members {
						if u == currentUserID {
							allowed = true
							break
						}
					}
					for _, u := range t.Managers {
						if u == currentUserID {
							allowed = true
							break
						}
					}
					for _, u := range t.Sponsors {
						if u == currentUserID {
							allowed = true
							break
						}
					}
				}
				if !allowed {
					continue
				}
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

	// optional current user
	var currentUserID string
	var currentUserRole string
	if v, ok := c.Get("userID"); ok {
		currentUserID = v.(string)
	} else {
		if auth := c.GetHeader("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			token := strings.TrimPrefix(auth, "Bearer ")
			if claims, err := utils.ValidateJWT(token); err == nil {
				currentUserID = claims.UserID
			}
		}
	}
	if v, ok := c.Get("userRole"); ok {
		currentUserRole = v.(string)
	}

	if database.DB != nil {
		var t models.Team
		row := database.DB.QueryRow(`SELECT id, name, country, followers_count, created_at, is_private FROM teams WHERE id = $1`, id)
		if err := row.Scan(&t.ID, &t.Name, &t.Country, &t.FollowersCount, &t.CreatedAt, &t.IsPrivate); err != nil {
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

		// managers, members, sponsors
		mgrRows, _ := database.DB.Query(`SELECT user_id::text FROM team_managers WHERE team_id = $1`, t.ID)
		if mgrRows != nil {
			for mgrRows.Next() {
				var u string
				mgrRows.Scan(&u)
				t.Managers = append(t.Managers, u)
			}
			mgrRows.Close()
		}
		memRows, _ := database.DB.Query(`SELECT user_id::text FROM team_members WHERE team_id = $1`, t.ID)
		if memRows != nil {
			for memRows.Next() {
				var u string
				memRows.Scan(&u)
				t.Members = append(t.Members, u)
			}
			memRows.Close()
		}
		sponRows, _ := database.DB.Query(`SELECT user_id::text FROM team_sponsors WHERE team_id = $1`, t.ID)
		if sponRows != nil {
			for sponRows.Next() {
				var u string
				sponRows.Scan(&u)
				t.Sponsors = append(t.Sponsors, u)
			}
			sponRows.Close()
		}

		// enforce visibility
		if t.IsPrivate {
			allowed := false
			if currentUserRole == "admin" {
				allowed = true
			}
			if currentUserID != "" {
				for _, u := range t.Members {
					if u == currentUserID {
						allowed = true
						break
					}
				}
				for _, u := range t.Managers {
					if u == currentUserID {
						allowed = true
						break
					}
				}
				for _, u := range t.Sponsors {
					if u == currentUserID {
						allowed = true
						break
					}
				}
			}
			if !allowed {
				c.JSON(http.StatusForbidden, gin.H{"error": "Team is private"})
				return
			}
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
		_, err := database.DB.Exec(`INSERT INTO teams (id, name, country, is_private, followers_count, created_at) VALUES ($1,$2,$3,$4,$5,$6)`, team.ID, team.Name, team.Country, team.IsPrivate, team.FollowersCount, team.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// insert drivers
		for _, d := range team.Drivers {
			_, _ = database.DB.Exec(`INSERT INTO team_drivers (team_id, driver_name) VALUES ($1,$2) ON CONFLICT DO NOTHING`, team.ID, d)
		}
		// insert managers, members, sponsors
		for _, m := range team.Managers {
			_, _ = database.DB.Exec(`INSERT INTO team_managers (team_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, team.ID, m)
		}
		for _, m := range team.Members {
			_, _ = database.DB.Exec(`INSERT INTO team_members (team_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, team.ID, m)
		}
		for _, s := range team.Sponsors {
			_, _ = database.DB.Exec(`INSERT INTO team_sponsors (team_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, team.ID, s)
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
		// update core fields including privacy
		res, err := database.DB.Exec(`UPDATE teams SET name=$1, country=$2, is_private=$3 WHERE id=$4`, updated.Name, updated.Country, updated.IsPrivate, id)
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
		// replace managers/members/sponsors
		_, _ = database.DB.Exec(`DELETE FROM team_managers WHERE team_id = $1`, id)
		for _, m := range updated.Managers {
			_, _ = database.DB.Exec(`INSERT INTO team_managers (team_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, id, m)
		}
		_, _ = database.DB.Exec(`DELETE FROM team_members WHERE team_id = $1`, id)
		for _, m := range updated.Members {
			_, _ = database.DB.Exec(`INSERT INTO team_members (team_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, id, m)
		}
		_, _ = database.DB.Exec(`DELETE FROM team_sponsors WHERE team_id = $1`, id)
		for _, s := range updated.Sponsors {
			_, _ = database.DB.Exec(`INSERT INTO team_sponsors (team_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, id, s)
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
