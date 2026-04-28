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

// GarageConfigRequest represents a garage config save request
type GarageConfigRequest struct {
	Name   string `json:"name" binding:"required"`
	Make   string `json:"make" binding:"required"`
	Model  string `json:"model" binding:"required"`
	Year   int    `json:"year" binding:"required"`
	Engine string `json:"engine" binding:"required"`
	Tuning string `json:"tuning"`
}

// SaveGarageConfig saves a user's garage configuration
func SaveGarageConfig(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var req GarageConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tuning := req.Tuning
	if tuning == "" {
		tuning = "stock"
	}

	newConfig := models.GarageConfig{
		ID:        uuid.New().String(),
		UserID:    userID,
		Name:      req.Name,
		Make:      req.Make,
		Model:     req.Model,
		Year:      req.Year,
		Engine:    req.Engine,
		Tuning:    tuning,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if database.DB != nil {
		err = database.DB.QueryRow(
			`INSERT INTO garage_configs (id, user_id, name, make, model, year, engine, tuning, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (user_id, make, model, year, engine, tuning)
			DO UPDATE SET name = EXCLUDED.name, updated_at = EXCLUDED.updated_at
			RETURNING id, created_at, updated_at`,
			newConfig.ID, newConfig.UserID, newConfig.Name, newConfig.Make, newConfig.Model,
			newConfig.Year, newConfig.Engine, newConfig.Tuning, newConfig.CreatedAt, newConfig.UpdatedAt,
		).Scan(&newConfig.ID, &newConfig.CreatedAt, &newConfig.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save configuration"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"config": newConfig, "message": "Configuration saved"})
		return
	}

	// In-memory fallback
	database.Mu.Lock()
	defer database.Mu.Unlock()

	// Store in memory (simple map for demo)
	if database.GarageConfigs == nil {
		database.GarageConfigs = make(map[string][]models.GarageConfig)
	}
	for i, cfg := range database.GarageConfigs[userID] {
		if cfg.Make == newConfig.Make &&
			cfg.Model == newConfig.Model &&
			cfg.Year == newConfig.Year &&
			cfg.Engine == newConfig.Engine &&
			cfg.Tuning == newConfig.Tuning {
			newConfig.ID = cfg.ID
			newConfig.CreatedAt = cfg.CreatedAt
			database.GarageConfigs[userID][i] = newConfig
			c.JSON(http.StatusCreated, gin.H{"config": newConfig, "message": "Configuration saved"})
			return
		}
	}
	database.GarageConfigs[userID] = append(database.GarageConfigs[userID], newConfig)

	c.JSON(http.StatusCreated, gin.H{"config": newConfig, "message": "Configuration saved"})
}

// GetGarageConfigs gets all configurations for a user
func GetGarageConfigs(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	if database.DB != nil {
		rows, err := database.DB.Query(
			`SELECT id, user_id, name, make, model, year, engine, tuning, created_at, updated_at 
			FROM garage_configs WHERE user_id = $1 ORDER BY created_at DESC`,
			userID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch configurations"})
			return
		}
		defer rows.Close()

		var configs []models.GarageConfig
		for rows.Next() {
			var config models.GarageConfig
			var createdAt, updatedAt time.Time
			err := rows.Scan(&config.ID, &config.UserID, &config.Name, &config.Make, &config.Model,
				&config.Year, &config.Engine, &config.Tuning, &createdAt, &updatedAt)
			if err != nil {
				continue
			}
			config.CreatedAt = createdAt
			config.UpdatedAt = updatedAt
			configs = append(configs, config)
		}

		if configs == nil {
			configs = []models.GarageConfig{}
		}
		c.JSON(http.StatusOK, gin.H{"configs": configs})
		return
	}

	// In-memory fallback
	database.Mu.Lock()
	defer database.Mu.Unlock()

	configs := database.GarageConfigs[userID]
	if configs == nil {
		configs = []models.GarageConfig{}
	}
	c.JSON(http.StatusOK, gin.H{"configs": configs})
}

// DeleteGarageConfig deletes a user's garage configuration
func DeleteGarageConfig(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	configID := c.Param("id")

	if database.DB != nil {
		result, err := database.DB.Exec(
			`DELETE FROM garage_configs WHERE id = $1 AND user_id = $2`,
			configID, userID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete configuration"})
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Configuration deleted"})
		return
	}

	// In-memory fallback
	database.Mu.Lock()
	defer database.Mu.Unlock()

	configs := database.GarageConfigs[userID]
	found := false
	var newConfigs []models.GarageConfig
	for _, config := range configs {
		if config.ID == configID {
			found = true
			continue
		}
		newConfigs = append(newConfigs, config)
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
		return
	}

	database.GarageConfigs[userID] = newConfigs
	c.JSON(http.StatusOK, gin.H{"message": "Configuration deleted"})
}

// getUserIDFromContext extracts user ID from JWT context
func getUserIDFromContext(c *gin.Context) (string, error) {
	userID, exists := c.Get("userID")
	if !exists {
		return "", sql.ErrNoRows
	}
	return userID.(string), nil
}
