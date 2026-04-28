package handlers

import (
	"database/sql"
	"net/http"
	"nitrous-backend/database"
	"nitrous-backend/models"
	"nitrous-backend/utils"
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

	// Create user struct
	// Use provided role or default to "viewer"
	role := req.Role
	if role == "" {
		role = "viewer"
	}
	newUser := models.User{
		ID:           uuid.New().String(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         role,
		Name:         req.Name,
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

		_, err = database.DB.Exec(`INSERT INTO users (id, email, password_hash, role, name, created_at) VALUES ($1,$2,$3,$4,$5,$6)`, newUser.ID, newUser.Email, newUser.PasswordHash, newUser.Role, newUser.Name, newUser.CreatedAt)
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
		row := database.DB.QueryRow(`SELECT id, email, password_hash, role, name, created_at FROM users WHERE email = $1`, req.Email)
		if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.Name, &u.CreatedAt); err != nil {
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
		row := database.DB.QueryRow(`SELECT id, email, role, name, created_at FROM users WHERE id = $1`, userID.(string))
		if err := row.Scan(&u.ID, &u.Email, &u.Role, &u.Name, &u.CreatedAt); err != nil {
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
