package handlers

import (
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
	
	database.Mu.Lock()
	defer database.Mu.Unlock()

	// Check if user already exists
	for _, user := range database.Users {
		if user.Email == req.Email {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
			return
		}
	}
	
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	
	// Create user
	newUser := models.User{
		ID:           uuid.New().String(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         "user",
		Name:         req.Name,
		CreatedAt:    time.Now(),
	}
	
	database.Users = append(database.Users, newUser)
	
	// Generate JWT token
	token, err := utils.GenerateJWT(newUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"user":  newUser,
		"token": token,
	})
}

// Login authenticates a user
func Login(c *gin.Context) {
	var req models.LoginRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	
	c.JSON(http.StatusOK, gin.H{
		"user":  foundUser,
		"token": token,
	})
}

// GetCurrentUser returns the authenticated user's info
func GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
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
