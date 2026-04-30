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
	"golang.org/x/crypto/bcrypt"
)

// Register creates a new user account
func Register(c *gin.Context) {
	var req models.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// New accounts always start as viewer.
	// Higher roles are granted only after membership payment flow completes.
	_ = req.Role

	// Create user struct
	newUser := models.User{
		ID:           uuid.New().String(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         "viewer",
		Name:         req.Name,
		Plan:         "FREE",
		CreatedAt:    time.Now(),
	}

	if database.DB != nil {
		// Check existing
		var exists bool
		err := database.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)`, req.Email).Scan(&exists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
			return
		}

		_, err = database.DB.Exec(`INSERT INTO users (id, email, password_hash, role, name, plan, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`, newUser.ID, newUser.Email, newUser.PasswordHash, newUser.Role, newUser.Name, newUser.Plan, newUser.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		token, err := utils.GenerateJWT(newUser.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"user": newUser, "token": token})
		return
	}

	database.Mu.Lock()
	defer database.Mu.Unlock()

	// Check if user already exists (in-memory fallback)
	for _, user := range database.Users {
		if user.Email == req.Email {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
			return
		}
	}

	database.Users = append(database.Users, newUser)

	// Generate JWT token
	token, err := utils.GenerateJWT(newUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user": newUser, "token": token})
}

// Login authenticates a user
func Login(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if database.DB != nil {
		var u models.User
		row := database.DB.QueryRow(`SELECT id, email, password_hash, role, name, COALESCE(plan, 'FREE'), created_at FROM users WHERE email = $1`, req.Email)
		if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.Name, &u.Plan, &u.CreatedAt); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		token, err := utils.GenerateJWT(u.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"user": u, "token": token})
		return
	}

	database.Mu.RLock()
	defer database.Mu.RUnlock()

	// Find user
	var foundUser *models.User
	for _, user := range database.Users {
		if user.Email == req.Email {
			foundUser = &user
			break
		}
	}

	if foundUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check password
	err := bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(foundUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": foundUser, "token": token})
}

// GetCurrentUser returns the authenticated user's info
func GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	if database.DB != nil {
		var u models.User
		row := database.DB.QueryRow(`SELECT id, email, role, name, COALESCE(plan, 'FREE'), created_at FROM users WHERE id = $1`, userID.(string))
		if err := row.Scan(&u.ID, &u.Email, &u.Role, &u.Name, &u.Plan, &u.CreatedAt); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, u)
		return
	}

	database.Mu.RLock()
	defer database.Mu.RUnlock()

	// Find user
	for _, user := range database.Users {
		if user.ID == userID.(string) {
			c.JSON(http.StatusOK, user)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
}

func planRank(plan string) int {
	switch strings.ToUpper(plan) {
	case "VIP":
		return 1
	case "PLATINUM":
		return 2
	default:
		return 0
	}
}

func roleAllowedForPlan(role, plan string) bool {
	r := strings.ToLower(strings.TrimSpace(role))
	pRank := planRank(plan)
	switch r {
	case "viewer":
		return true
	case "participant", "manager":
		return pRank >= 1
	case "sponsor":
		return pRank >= 2
	default:
		return false
	}
}

// UpdateCurrentUserPlan upgrades the authenticated user's membership plan.
func UpdateCurrentUserPlan(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		Plan string `json:"plan" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	targetPlan := strings.ToUpper(strings.TrimSpace(req.Plan))
	if targetPlan != "VIP" && targetPlan != "PLATINUM" && targetPlan != "FREE" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan"})
		return
	}

	if database.DB != nil {
		var currentPlan string
		if err := database.DB.QueryRow(`SELECT COALESCE(plan, 'FREE') FROM users WHERE id = $1`, userID.(string)).Scan(&currentPlan); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if planRank(targetPlan) < planRank(currentPlan) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Downgrades are not allowed"})
			return
		}

		var u models.User
		row := database.DB.QueryRow(`
			UPDATE users SET plan = $1
			WHERE id = $2
			RETURNING id, email, role, name, COALESCE(plan, 'FREE'), created_at
		`, targetPlan, userID.(string))
		if err := row.Scan(&u.ID, &u.Email, &u.Role, &u.Name, &u.Plan, &u.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"user": u, "message": "Plan updated"})
		return
	}

	database.Mu.Lock()
	defer database.Mu.Unlock()

	for i := range database.Users {
		if database.Users[i].ID == userID.(string) {
			if planRank(targetPlan) < planRank(database.Users[i].Plan) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Downgrades are not allowed"})
				return
			}
			database.Users[i].Plan = targetPlan
			c.JSON(http.StatusOK, gin.H{"user": database.Users[i], "message": "Plan updated"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
}

// UpdateCurrentUserRole updates the authenticated user's role after plan validation.
func UpdateCurrentUserRole(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	targetRole := strings.ToLower(strings.TrimSpace(req.Role))
	if targetRole != "viewer" && targetRole != "participant" && targetRole != "manager" && targetRole != "sponsor" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

	if database.DB != nil {
		var currentPlan string
		if err := database.DB.QueryRow(`SELECT COALESCE(plan, 'FREE') FROM users WHERE id = $1`, userID.(string)).Scan(&currentPlan); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if !roleAllowedForPlan(targetRole, currentPlan) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Current plan does not allow requested role"})
			return
		}

		var u models.User
		row := database.DB.QueryRow(`
			UPDATE users
			SET role = $1
			WHERE id = $2
			RETURNING id, email, role, name, COALESCE(plan, 'FREE'), created_at
		`, targetRole, userID.(string))
		if err := row.Scan(&u.ID, &u.Email, &u.Role, &u.Name, &u.Plan, &u.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"user": u, "message": "Role updated"})
		return
	}

	database.Mu.Lock()
	defer database.Mu.Unlock()

	for i := range database.Users {
		if database.Users[i].ID == userID.(string) {
			if !roleAllowedForPlan(targetRole, database.Users[i].Plan) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Current plan does not allow requested role"})
				return
			}
			database.Users[i].Role = targetRole
			c.JSON(http.StatusOK, gin.H{"user": database.Users[i], "message": "Role updated"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
}
