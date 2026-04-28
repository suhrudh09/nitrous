package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"nitrous-backend/config"
	"nitrous-backend/database"
	"nitrous-backend/models"
	"nitrous-backend/utils"

	"github.com/gin-gonic/gin"
)

func setupHandlersTestEnv() {
	gin.SetMode(gin.TestMode)
	config.AppConfig.JWTSecret = "test-secret"
	database.Events = nil
	database.Categories = nil
	database.Journeys = nil
	database.MerchItems = nil
	database.Users = nil
	database.Teams = nil
	database.Reminders = nil
	database.Orders = nil
	database.Passes = nil
	database.PassPurchases = nil
}

func seedHandlersCatalogData() {
	database.Events = []models.Event{
		sampleEvent("event-1"),
		{
			ID:       "event-2",
			Title:    "Sample Event 2",
			Location: "Test Track 2",
			Date:     time.Now().Add(48 * time.Hour).UTC(),
			IsLive:   true,
			Category: "air",
			Time:     "14:00 UTC",
		},
	}
	database.Categories = []models.Category{
		{ID: "cat-1", Name: "MOTORSPORT", Slug: "motorsport", Icon: "R", LiveCount: 1, Description: "desc", Color: "cyan"},
		{ID: "cat-2", Name: "AIR", Slug: "air", Icon: "A", LiveCount: 1, Description: "desc", Color: "blue"},
	}
	database.Journeys = []models.Journey{
		{ID: "journey-1", Title: "Journey 1", Category: "MOTORSPORT", Description: "Desc", Badge: "EXCLUSIVE", SlotsLeft: 2, Date: time.Now().Add(24 * time.Hour).UTC(), Price: 100},
	}
	database.MerchItems = []models.MerchItem{
		{ID: "merch-1", Name: "Merch 1", Icon: "M", Price: 25, Category: "apparel"},
	}
	database.Teams = []models.Team{
		{ID: "team-1", Name: "Team 1", Country: "USA", Drivers: []string{"D1"}, Followers: []string{}, FollowersCount: 0, CreatedAt: time.Now().UTC()},
	}
	streams = []Stream{
		{ID: "stream-1", Title: "Main", Category: "motorsport", IsLive: true, Viewers: 100, Color: "cyan"},
	}
}

func makeToken(t *testing.T, userID string) string {
	t.Helper()
	token, err := utils.GenerateJWT(userID)
	if err != nil {
		t.Fatalf("failed generating token: %v", err)
	}
	return token
}

func performJSONRequest(r http.Handler, method string, path string, body any, token string) *httptest.ResponseRecorder {
	var payload []byte
	if body != nil {
		payload, _ = json.Marshal(body)
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func performRawRequest(r http.Handler, method string, path string, rawBody string, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(rawBody))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func sampleEvent(id string) models.Event {
	return models.Event{
		ID:       id,
		Title:    "Sample Event",
		Location: "Test Track",
		Date:     time.Now().Add(24 * time.Hour).UTC(),
		IsLive:   false,
		Category: "motorsport",
		Time:     "12:00 UTC",
	}
}
