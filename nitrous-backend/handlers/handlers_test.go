package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nitrous-backend/database"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGetEvents_ListAndCategoryFilter(t *testing.T) {
	database.InitDB()

	// no filter
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/", nil)
	c.Request = req

	GetEvents(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	count := int(resp["count"].(float64))
	if count != len(database.Events) {
		t.Fatalf("expected count %d, got %d", len(database.Events), count)
	}

	// apply category filter (pick one from seeded data)
	target := database.Events[0].Category

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	req = httptest.NewRequest("GET", "/?category="+target, nil)
	c.Request = req

	GetEvents(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for filtered, got %d", w.Code)
	}

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal filtered response: %v", err)
	}

	filteredCount := int(resp["count"].(float64))
	// compute expected
	expected := 0
	for _, e := range database.Events {
		if e.Category == target {
			expected++
		}
	}

	if filteredCount != expected {
		t.Fatalf("expected %d filtered events, got %d", expected, filteredCount)
	}
}

func TestGetLiveEvents_ReturnsOnlyLive(t *testing.T) {
	database.InitDB()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	GetLiveEvents(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	count := int(resp["count"].(float64))
	expected := 0
	for _, e := range database.Events {
		if e.IsLive {
			expected++
		}
	}

	if count != expected {
		t.Fatalf("expected %d live events, got %d", expected, count)
	}
}

func TestGetEventByID_FoundAndNotFound(t *testing.T) {
	database.InitDB()

	// found
	id := database.Events[0].ID
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: id}}
	c.Request = httptest.NewRequest("GET", "/", nil)
	GetEventByID(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for found, got %d", w.Code)
	}

	// not found
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "no-such-id"}}
	c.Request = httptest.NewRequest("GET", "/", nil)
	GetEventByID(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for not found, got %d", w.Code)
	}
}

func TestCategories_ListAndBySlug(t *testing.T) {
	database.InitDB()

	// list
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	GetCategories(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on categories list, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal categories list: %v", err)
	}
	count := int(resp["count"].(float64))
	if count != len(database.Categories) {
		t.Fatalf("expected %d categories, got %d", len(database.Categories), count)
	}

	// by slug found
	slug := database.Categories[0].Slug
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "slug", Value: slug}}
	c.Request = httptest.NewRequest("GET", "/", nil)
	GetCategoryBySlug(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for category found, got %d", w.Code)
	}

	// by slug not found
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "slug", Value: "no-such-slug"}}
	c.Request = httptest.NewRequest("GET", "/", nil)
	GetCategoryBySlug(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for category not found, got %d", w.Code)
	}
}

func TestJourneys_ListAndByID(t *testing.T) {
	database.InitDB()

	// list
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	GetJourneys(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on journeys list, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal journeys list: %v", err)
	}
	count := int(resp["count"].(float64))
	if count != len(database.Journeys) {
		t.Fatalf("expected %d journeys, got %d", len(database.Journeys), count)
	}

	// by id found
	id := database.Journeys[0].ID
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: id}}
	c.Request = httptest.NewRequest("GET", "/", nil)
	GetJourneyByID(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for journey found, got %d", w.Code)
	}

	// not found
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "no-id"}}
	c.Request = httptest.NewRequest("GET", "/", nil)
	GetJourneyByID(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for journey not found, got %d", w.Code)
	}
}

func TestMerch_ListAndByID(t *testing.T) {
	database.InitDB()

	// list
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	GetMerchItems(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on merch list, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal merch list: %v", err)
	}
	count := int(resp["count"].(float64))
	if count != len(database.MerchItems) {
		t.Fatalf("expected %d merch items, got %d", len(database.MerchItems), count)
	}

	// by id found
	id := database.MerchItems[0].ID
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: id}}
	c.Request = httptest.NewRequest("GET", "/", nil)
	GetMerchItemByID(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for merch found, got %d", w.Code)
	}

	// not found
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "no-id"}}
	c.Request = httptest.NewRequest("GET", "/", nil)
	GetMerchItemByID(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for merch not found, got %d", w.Code)
	}
}

func TestTeams_ListAndByID(t *testing.T) {
	database.InitDB()

	// list
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	GetTeams(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on teams list, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal teams list: %v", err)
	}
	count := int(resp["count"].(float64))
	if count != len(database.Teams) {
		t.Fatalf("expected %d teams, got %d", len(database.Teams), count)
	}
	if len(database.Teams) == 0 {
		t.Skip("teams seed is empty in API-first mode; skipping by-id happy path")
	}

	// by id found
	id := database.Teams[0].ID
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: id}}
	c.Request = httptest.NewRequest("GET", "/", nil)
	GetTeamByID(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for team found, got %d", w.Code)
	}

	// not found
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "no-id"}}
	c.Request = httptest.NewRequest("GET", "/", nil)
	GetTeamByID(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for team not found, got %d", w.Code)
	}
}

func TestStreams_ListAndByID(t *testing.T) {
	database.InitDB()

	// list
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	GetStreams(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on streams list, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal streams list: %v", err)
	}
	count := int(resp["count"].(float64))
	if count != len(streams) {
		t.Fatalf("expected %d streams, got %d", len(streams), count)
	}
	if len(streams) == 0 {
		t.Skip("streams seed is empty in provider-driven mode; skipping by-id happy path")
	}

	// by id found
	id := streams[0].ID
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: id}}
	c.Request = httptest.NewRequest("GET", "/", nil)
	GetStreamByID(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for stream found, got %d", w.Code)
	}

	// not found
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "no-id"}}
	c.Request = httptest.NewRequest("GET", "/", nil)
	GetStreamByID(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for stream not found, got %d", w.Code)
	}
}

func TestStreamsWS_UpgradeAndTelemetryBroadcast(t *testing.T) {
	// Start hub and server
	database.InitDB()

	r := gin.New()
	r.GET("/ws", StreamsWS)

	srv := httptest.NewServer(r)
	defer srv.Close()

	// build ws url
	wsURL := "ws://" + strings.TrimPrefix(srv.URL, "http://") + "/ws"

	// dial websocket
	dialer := websocket.DefaultDialer
	conn, resp, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("websocket dial failed (status %v): %v", resp.Status, err)
	}
	defer conn.Close()

	// broadcast telemetry and verify client receives it
	BroadcastTelemetry("stream-1", 123, 6000, 5, 0.98)

	// allow a small window for the message to be dispatched
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	mt, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read telemetry message: %v", err)
	}
	if mt != websocket.TextMessage {
		t.Fatalf("expected text message, got type %d", mt)
	}

	// basic check that message contains telemetry type and stream id
	var payload map[string]interface{}
	if err := json.Unmarshal(message, &payload); err != nil {
		t.Fatalf("failed to unmarshal telemetry payload: %v", err)
	}
	if payload["type"] != "telemetry" {
		t.Fatalf("expected telemetry type, got %v", payload["type"])
	}
	if payload["streamId"] != "stream-1" {
		t.Fatalf("expected streamId 'stream-1', got %v", payload["streamId"])
	}
}
