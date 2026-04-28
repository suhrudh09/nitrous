package middleware

import (
	"database/sql"
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

		// Also set user role in context (from DB if available, else from in-memory users)
		userRole := "viewer"
		if database.DB != nil {
			var role string
			row := database.DB.QueryRow(`SELECT role FROM users WHERE id = $1`, claims.UserID)
			if err := row.Scan(&role); err == nil {
				userRole = role
			}
		} else {
			for _, u := range database.Users {
				if u.ID == claims.UserID {
					userRole = u.Role
					break
				}
			}
		}
		c.Set("userRole", userRole)
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

		if database.DB != nil {
			var role string
			row := database.DB.QueryRow(`SELECT role FROM users WHERE id = $1`, userID.(string))
			if err := row.Scan(&role); err != nil {
				if err == sql.ErrNoRows {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				}
				c.Abort()
				return
			}

			if role != "admin" {
				c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
				c.Abort()
				return
			}

			c.Set("userRole", role)
			c.Next()
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

// RequireRoles ensures the authenticated user has one of the allowed roles.
func RequireRoles(allowed ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}
		roleStr := userRole.(string)
		for _, a := range allowed {
			if a == roleStr {
				c.Next()
				return
			}
		}
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient role privileges"})
		c.Abort()
	}
}
