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

func TestBookJourneyEndpoint(t *testing.T) {
	setupHandlersTestEnv()
	database.Journeys = []models.Journey{
		{ID: "j1", Title: "A", SlotsLeft: 1, Date: time.Now().Add(24 * time.Hour)},
		{ID: "j2", Title: "B", SlotsLeft: 0, Date: time.Now().Add(24 * time.Hour)},
	}

	r := gin.New()
	r.POST("/journeys/:id/book", middleware.AuthMiddleware(), BookJourney)

	w := performJSONRequest(r, http.MethodPost, "/journeys/j1/book", nil, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 unauthorized, got %d", w.Code)
	}

	token := makeToken(t, "user-1")
	w = performJSONRequest(r, http.MethodPost, "/journeys/nope/book", nil, token)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for not found, got %d", w.Code)
	}

	w = performJSONRequest(r, http.MethodPost, "/journeys/j2/book", nil, token)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for no slots left, got %d", w.Code)
	}

	w = performJSONRequest(r, http.MethodPost, "/journeys/j1/book", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for successful booking, got %d", w.Code)
	}
	if database.Journeys[0].SlotsLeft != 0 {
		t.Fatalf("expected slots to decrement to 0, got %d", database.Journeys[0].SlotsLeft)
	}
}

func TestFollowTeamEndpoint(t *testing.T) {
	setupHandlersTestEnv()
	database.Teams = []models.Team{{ID: "t1", Name: "Team A", Followers: []string{}}}

	r := gin.New()
	r.POST("/teams/:id/follow", middleware.AuthMiddleware(), FollowTeam)

	w := performJSONRequest(r, http.MethodPost, "/teams/t1/follow", nil, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 unauthorized, got %d", w.Code)
	}

	token := makeToken(t, "user-1")
	w = performJSONRequest(r, http.MethodPost, "/teams/nope/follow", nil, token)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 team not found, got %d", w.Code)
	}

	w = performJSONRequest(r, http.MethodPost, "/teams/t1/follow", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for successful follow, got %d", w.Code)
	}
	if database.Teams[0].FollowersCount != 1 {
		t.Fatalf("expected follower count 1, got %d", database.Teams[0].FollowersCount)
	}

	w = performJSONRequest(r, http.MethodPost, "/teams/t1/follow", nil, token)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for duplicate follow, got %d", w.Code)
	}
}

func TestUnfollowTeamEndpoint(t *testing.T) {
	setupHandlersTestEnv()
	database.Teams = []models.Team{{ID: "t1", Name: "Team A", Followers: []string{"user-1"}, FollowersCount: 1}}

	r := gin.New()
	r.POST("/teams/:id/unfollow", middleware.AuthMiddleware(), UnfollowTeam)

	w := performJSONRequest(r, http.MethodPost, "/teams/t1/unfollow", nil, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 unauthorized, got %d", w.Code)
	}

	token := makeToken(t, "user-1")
	w = performJSONRequest(r, http.MethodPost, "/teams/nope/unfollow", nil, token)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 team not found, got %d", w.Code)
	}

	w = performJSONRequest(r, http.MethodPost, "/teams/t1/unfollow", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for successful unfollow, got %d", w.Code)
	}
	if database.Teams[0].FollowersCount != 0 {
		t.Fatalf("expected follower count 0, got %d", database.Teams[0].FollowersCount)
	}

	w = performJSONRequest(r, http.MethodPost, "/teams/t1/unfollow", nil, token)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unfollow-when-not-following, got %d", w.Code)
	}
}

func TestUnfollowTeamFrontendDeleteRoute(t *testing.T) {
	setupHandlersTestEnv()
	database.Teams = []models.Team{{ID: "t1", Name: "Team A", Followers: []string{"user-1"}, FollowersCount: 1}}

	r := gin.New()
	r.DELETE("/teams/:id/follow", middleware.AuthMiddleware(), UnfollowTeam)

	token := makeToken(t, "user-1")
	w := performJSONRequest(r, http.MethodDelete, "/teams/t1/follow", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for delete follow route, got %d", w.Code)
	}
	if database.Teams[0].FollowersCount != 0 {
		t.Fatalf("expected follower count 0, got %d", database.Teams[0].FollowersCount)
	}
}
