package handlers

import (
	"net/http"
	"testing"
	"time"

	"nitrous-backend/database"
	"nitrous-backend/middleware"
	"nitrous-backend/models"

	"github.com/gin-gonic/gin"
)

func seedAdminAndUser() {
	database.Users = []models.User{
		{ID: "admin-1", Email: "admin@example.com", Role: "admin"},
		{ID: "user-1", Email: "user@example.com", Role: "user"},
	}
}

func TestCategoryManagementAdminRoutes(t *testing.T) {
	setupHandlersTestEnv()
	seedAdminAndUser()
	database.Categories = []models.Category{{ID: "cat-1", Name: "MOTORSPORT", Slug: "motorsport", Icon: "x", LiveCount: 1, Description: "desc", Color: "cyan"}}

	r := gin.New()
	r.POST("/categories", middleware.AuthMiddleware(), middleware.AdminMiddleware(), CreateCategory)
	r.PUT("/categories/:slug", middleware.AuthMiddleware(), middleware.AdminMiddleware(), UpdateCategory)
	r.DELETE("/categories/:slug", middleware.AuthMiddleware(), middleware.AdminMiddleware(), DeleteCategory)

	payload := map[string]any{"name": "WATER", "slug": "water", "icon": "i", "liveCount": 1, "description": "d", "color": "blue"}

	w := performJSONRequest(r, http.MethodPost, "/categories", payload, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 unauthorized, got %d", w.Code)
	}

	userToken := makeToken(t, "user-1")
	w = performJSONRequest(r, http.MethodPost, "/categories", payload, userToken)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 forbidden for non-admin, got %d", w.Code)
	}

	adminToken := makeToken(t, "admin-1")
	w = performJSONRequest(r, http.MethodPost, "/categories", payload, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 for admin create, got %d", w.Code)
	}

	updatePayload := map[string]any{"name": "MOTORSPORT UPDATED", "slug": "motorsport", "icon": "x", "liveCount": 2, "description": "new", "color": "orange"}
	w = performJSONRequest(r, http.MethodPut, "/categories/motorsport", updatePayload, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin update, got %d", w.Code)
	}

	w = performJSONRequest(r, http.MethodDelete, "/categories/motorsport", nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin delete, got %d", w.Code)
	}
}

func TestJourneyCatalogManagementAdminRoutes(t *testing.T) {
	setupHandlersTestEnv()
	seedAdminAndUser()
	database.Journeys = []models.Journey{{ID: "j-1", Title: "A", Category: "MOTORSPORT", Description: "D", Badge: "EXCLUSIVE", SlotsLeft: 2, Date: time.Now().Add(24 * time.Hour), Price: 100}}

	r := gin.New()
	r.POST("/journeys", middleware.AuthMiddleware(), middleware.AdminMiddleware(), CreateJourney)
	r.PUT("/journeys/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), UpdateJourney)
	r.DELETE("/journeys/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), DeleteJourney)

	journeyPayload := map[string]any{"title": "NEW", "category": "AIR", "description": "D", "badge": "LIMITED", "slotsLeft": 1, "price": 250}

	w := performJSONRequest(r, http.MethodPost, "/journeys", journeyPayload, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 unauthorized, got %d", w.Code)
	}

	userToken := makeToken(t, "user-1")
	w = performJSONRequest(r, http.MethodPost, "/journeys", journeyPayload, userToken)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 forbidden for non-admin, got %d", w.Code)
	}

	adminToken := makeToken(t, "admin-1")
	w = performJSONRequest(r, http.MethodPost, "/journeys", journeyPayload, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 for admin create, got %d", w.Code)
	}

	updatePayload := map[string]any{"title": "UPDATED", "category": "MOTORSPORT", "description": "D2", "badge": "EXCLUSIVE", "slotsLeft": 3, "price": 300}
	w = performJSONRequest(r, http.MethodPut, "/journeys/j-1", updatePayload, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin update, got %d", w.Code)
	}

	w = performJSONRequest(r, http.MethodDelete, "/journeys/j-1", nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin delete, got %d", w.Code)
	}
}

func TestTeamManagementAdminRoutes(t *testing.T) {
	setupHandlersTestEnv()
	seedAdminAndUser()
	database.Teams = []models.Team{{ID: "t-1", Name: "Team 1", Country: "USA", Drivers: []string{"D1"}, Followers: []string{}, FollowersCount: 0, CreatedAt: time.Now()}}

	r := gin.New()
	r.POST("/teams", middleware.AuthMiddleware(), middleware.AdminMiddleware(), CreateTeam)
	r.PUT("/teams/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), UpdateTeam)
	r.DELETE("/teams/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), DeleteTeam)

	teamPayload := map[string]any{"name": "Team New", "country": "UK", "drivers": []string{"A", "B"}}

	w := performJSONRequest(r, http.MethodPost, "/teams", teamPayload, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 unauthorized, got %d", w.Code)
	}

	userToken := makeToken(t, "user-1")
	w = performJSONRequest(r, http.MethodPost, "/teams", teamPayload, userToken)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 forbidden for non-admin, got %d", w.Code)
	}

	adminToken := makeToken(t, "admin-1")
	w = performJSONRequest(r, http.MethodPost, "/teams", teamPayload, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 for admin create, got %d", w.Code)
	}

	updatePayload := map[string]any{"name": "Team Updated", "country": "FR", "drivers": []string{"X"}}
	w = performJSONRequest(r, http.MethodPut, "/teams/t-1", updatePayload, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin update, got %d", w.Code)
	}

	w = performJSONRequest(r, http.MethodDelete, "/teams/t-1", nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin delete, got %d", w.Code)
	}
}

func TestStreamManagementAdminRoutes(t *testing.T) {
	setupHandlersTestEnv()
	seedAdminAndUser()

	originalStreams := streams
	streams = []Stream{{ID: "s-1", Title: "Main", Category: "motorsport", IsLive: true, Viewers: 100, Color: "cyan"}}
	defer func() { streams = originalStreams }()

	r := gin.New()
	r.POST("/streams", middleware.AuthMiddleware(), middleware.AdminMiddleware(), CreateStream)
	r.PUT("/streams/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), UpdateStream)
	r.DELETE("/streams/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), DeleteStream)

	createPayload := map[string]any{"id": "s-2", "title": "Alt", "category": "air", "isLive": false, "viewers": 0, "color": "blue"}

	w := performJSONRequest(r, http.MethodPost, "/streams", createPayload, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 unauthorized, got %d", w.Code)
	}

	userToken := makeToken(t, "user-1")
	w = performJSONRequest(r, http.MethodPost, "/streams", createPayload, userToken)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 forbidden for non-admin, got %d", w.Code)
	}

	adminToken := makeToken(t, "admin-1")
	w = performJSONRequest(r, http.MethodPost, "/streams", createPayload, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 for admin create, got %d", w.Code)
	}

	updatePayload := map[string]any{"title": "Main Updated", "category": "motorsport", "isLive": true, "viewers": 200, "color": "cyan"}
	w = performJSONRequest(r, http.MethodPut, "/streams/s-1", updatePayload, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin update, got %d", w.Code)
	}

	w = performJSONRequest(r, http.MethodDelete, "/streams/s-1", nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin delete, got %d", w.Code)
	}
}
