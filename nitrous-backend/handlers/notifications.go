package handlers

import (
	"database/sql"
	"net/http"
	"nitrous-backend/database"
	"nitrous-backend/models"
	"time"

	"github.com/gin-gonic/gin"
)

func cleanupReadNotificationsForUser(userID string) {
	if database.DB == nil {
		return
	}
	_, _ = database.DB.Exec(`DELETE FROM notifications WHERE user_id = $1::uuid AND read_at IS NOT NULL AND read_at <= NOW() - INTERVAL '1 minute'`, userID)
}

// GetMyNotifications returns notifications for the authenticated user.
func GetMyNotifications(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uid := userID.(string)
	cleanupReadNotificationsForUser(uid)

	if database.DB != nil {
		rows, err := database.DB.Query(`
			SELECT id::text, user_id::text, title, body, type, read_at, created_at
			FROM notifications
			WHERE user_id = $1::uuid
			ORDER BY created_at DESC
			LIMIT 50
		`, uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		notifications := make([]models.Notification, 0)
		unreadCount := 0
		for rows.Next() {
			var n models.Notification
			if err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Body, &n.Type, &n.ReadAt, &n.CreatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if n.ReadAt == nil {
				unreadCount++
			}
			notifications = append(notifications, n)
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"notifications": notifications, "count": len(notifications), "unread": unreadCount})
		return
	}

	c.JSON(http.StatusOK, gin.H{"notifications": []models.Notification{}, "count": 0, "unread": 0})
}

// MarkNotificationRead marks a notification as read for the authenticated user.
func MarkNotificationRead(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	notificationID := c.Param("id")
	uid := userID.(string)

	if database.DB != nil {
		res, err := database.DB.Exec(`
			UPDATE notifications
			SET read_at = COALESCE(read_at, NOW())
			WHERE id = $1::uuid AND user_id = $2::uuid
		`, notificationID, uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
			return
		}

		var readAt time.Time
		err = database.DB.QueryRow(`SELECT read_at FROM notifications WHERE id = $1::uuid`, notificationID).Scan(&readAt)
		if err != nil && err != sql.ErrNoRows {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read", "readAt": readAt})
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{"error": "Notifications require database mode"})
}
