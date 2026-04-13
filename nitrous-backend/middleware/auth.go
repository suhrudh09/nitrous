package middleware

import (
	"net/http"
	"nitrous-backend/database"
	"nitrous-backend/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}
		
		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}
		
		token := parts[1]
		
		// Validate token
		claims, err := utils.ValidateJWT(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}
		
		// Set user ID in context
		c.Set("userID", claims.UserID)
		c.Next()
	}
}

// AdminMiddleware allows only authenticated users with admin role.
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		for _, user := range database.Users {
			if user.ID == userID.(string) {
				if user.Role != "admin" {
					c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
					c.Abort()
					return
				}

				c.Set("userRole", user.Role)
				c.Next()
				return
			}
		}

		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		c.Abort()
	}
}
