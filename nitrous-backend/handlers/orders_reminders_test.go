package handlers

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"nitrous-backend/database"
	"nitrous-backend/middleware"
	"nitrous-backend/models"

	"github.com/gin-gonic/gin"
)

func TestCreateOrderEndpoint(t *testing.T) {
	setupHandlersTestEnv()
	database.MerchItems = []models.MerchItem{{ID: "m1", Name: "Cap", Price: 50}}

	r := gin.New()
	r.POST("/orders", middleware.AuthMiddleware(), CreateOrder)

	validReq := map[string]any{"merchItemIds": []string{"m1"}, "quantities": []int{2}, "unitPrices": []float64{50}}

	w := performJSONRequest(r, http.MethodPost, "/orders", validReq, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 unauthorized, got %d", w.Code)
	}

	token := makeToken(t, "user-1")
	invalidReq := map[string]any{"merchItemIds": []string{"m1"}, "quantities": []int{0}, "unitPrices": []float64{50}}
	w = performJSONRequest(r, http.MethodPost, "/orders", invalidReq, token)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 invalid payload, got %d", w.Code)
	}

	notFoundReq := map[string]any{"merchItemIds": []string{"unknown"}, "quantities": []int{1}, "unitPrices": []float64{50}}
	w = performJSONRequest(r, http.MethodPost, "/orders", notFoundReq, token)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 merch not found, got %d", w.Code)
	}

	w = performJSONRequest(r, http.MethodPost, "/orders", validReq, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 successful create, got %d", w.Code)
	}

	var order models.Order
	if err := json.Unmarshal(w.Body.Bytes(), &order); err != nil {
		t.Fatalf("failed to decode order: %v", err)
	}
	if order.Total != 100 {
		t.Fatalf("expected total price 100, got %.2f", order.Total)
	}
}

func TestGetMyOrdersEndpoint(t *testing.T) {
	setupHandlersTestEnv()
	database.Orders = []models.Order{
		{ID: "o1", UserID: "user-1", MerchItemIDs: []string{"m1"}, Quantities: []int{1}, UnitPrices: []float64{10}, Total: 10},
		{ID: "o2", UserID: "user-2", MerchItemIDs: []string{"m2"}, Quantities: []int{1}, UnitPrices: []float64{20}, Total: 20},
	}

	r := gin.New()
	r.GET("/orders", middleware.AuthMiddleware(), GetMyOrders)

	w := performJSONRequest(r, http.MethodGet, "/orders", nil, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 unauthorized, got %d", w.Code)
	}

	token := makeToken(t, "user-1")
	w = performJSONRequest(r, http.MethodGet, "/orders", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Orders []models.Order `json:"orders"`
		Count  int            `json:"count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Count != 1 || len(resp.Orders) != 1 || resp.Orders[0].UserID != "user-1" {
		t.Fatalf("expected only current user's orders")
	}
}

func TestGetOrderByIDEndpoint(t *testing.T) {
	setupHandlersTestEnv()
	database.Orders = []models.Order{
		{ID: "o1", UserID: "user-1", MerchItemIDs: []string{"m1"}, Quantities: []int{1}, UnitPrices: []float64{10}, Total: 10},
		{ID: "o2", UserID: "user-2", MerchItemIDs: []string{"m2"}, Quantities: []int{1}, UnitPrices: []float64{20}, Total: 20},
	}

	r := gin.New()
	r.GET("/orders/:id", middleware.AuthMiddleware(), GetOrderByID)

	w := performJSONRequest(r, http.MethodGet, "/orders/o1", nil, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 unauthorized, got %d", w.Code)
	}

	tokenUser1 := makeToken(t, "user-1")
	w = performJSONRequest(r, http.MethodGet, "/orders/o1", nil, tokenUser1)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for own order, got %d", w.Code)
	}

	w = performJSONRequest(r, http.MethodGet, "/orders/o2", nil, tokenUser1)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for other user's order, got %d", w.Code)
	}

	w = performJSONRequest(r, http.MethodGet, "/orders/nope", nil, tokenUser1)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing order, got %d", w.Code)
	}
}

func TestCancelOrderEndpoint(t *testing.T) {
	setupHandlersTestEnv()
	database.Orders = []models.Order{
		{ID: "o1", UserID: "user-1", MerchItemIDs: []string{"m1"}, Quantities: []int{1}, UnitPrices: []float64{10}, Total: 10, Status: "created"},
		{ID: "o2", UserID: "user-2", MerchItemIDs: []string{"m2"}, Quantities: []int{1}, UnitPrices: []float64{20}, Total: 20, Status: "created"},
	}

	r := gin.New()
	r.DELETE("/orders/:id", middleware.AuthMiddleware(), CancelOrder)

	w := performJSONRequest(r, http.MethodDelete, "/orders/o1", nil, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 unauthorized, got %d", w.Code)
	}

	token := makeToken(t, "user-1")
	w = performJSONRequest(r, http.MethodDelete, "/orders/o1", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for own order cancel, got %d", w.Code)
	}

	var order models.Order
	if err := json.Unmarshal(w.Body.Bytes(), &order); err != nil {
		t.Fatalf("failed to decode order: %v", err)
	}
	if order.Status != "cancelled" {
		t.Fatalf("expected order status cancelled, got %s", order.Status)
	}

	w = performJSONRequest(r, http.MethodDelete, "/orders/o2", nil, token)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for other user's order, got %d", w.Code)
	}

	w = performJSONRequest(r, http.MethodDelete, "/orders/nope", nil, token)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing order, got %d", w.Code)
	}
}

func TestSetReminderEndpoint(t *testing.T) {
	setupHandlersTestEnv()
	database.Events = []models.Event{sampleEvent("event-1")}

	r := gin.New()
	r.POST("/reminders", middleware.AuthMiddleware(), SetReminder)

	validReq := map[string]any{
		"eventId":  "event-1",
		"message":  "Test reminder",
		"remindAt": time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339),
	}

	w := performJSONRequest(r, http.MethodPost, "/reminders", validReq, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 unauthorized, got %d", w.Code)
	}

	token := makeToken(t, "user-1")
	invalidReq := map[string]any{"message": "missing required fields"}
	w = performJSONRequest(r, http.MethodPost, "/reminders", invalidReq, token)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 invalid payload, got %d", w.Code)
	}

	pastReq := map[string]any{
		"eventId":  "event-1",
		"message":  "Past",
		"remindAt": time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339),
	}
	w = performJSONRequest(r, http.MethodPost, "/reminders", pastReq, token)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for past remind time, got %d", w.Code)
	}

	notFoundReq := map[string]any{
		"eventId":  "missing-event",
		"message":  "Missing",
		"remindAt": time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339),
	}
	w = performJSONRequest(r, http.MethodPost, "/reminders", notFoundReq, token)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 event not found, got %d", w.Code)
	}

	w = performJSONRequest(r, http.MethodPost, "/reminders", validReq, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 successful reminder create, got %d", w.Code)
	}
}

func TestGetMyRemindersEndpoint(t *testing.T) {
	setupHandlersTestEnv()
	database.Reminders = []models.Reminder{
		{ID: "r1", UserID: "user-1", EventID: "e1", RemindAt: time.Now().Add(time.Hour)},
		{ID: "r2", UserID: "user-2", EventID: "e2", RemindAt: time.Now().Add(time.Hour)},
	}

	r := gin.New()
	r.GET("/reminders", middleware.AuthMiddleware(), GetMyReminders)

	w := performJSONRequest(r, http.MethodGet, "/reminders", nil, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 unauthorized, got %d", w.Code)
	}

	token := makeToken(t, "user-1")
	w = performJSONRequest(r, http.MethodGet, "/reminders", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Reminders []models.Reminder `json:"reminders"`
		Count     int               `json:"count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Count != 1 || len(resp.Reminders) != 1 || resp.Reminders[0].UserID != "user-1" {
		t.Fatalf("expected only current user's reminders")
	}
}

func TestDeleteReminderEndpoint(t *testing.T) {
	setupHandlersTestEnv()
	database.Reminders = []models.Reminder{
		{ID: "r1", UserID: "user-1", EventID: "e1", RemindAt: time.Now().Add(time.Hour)},
		{ID: "r2", UserID: "user-2", EventID: "e2", RemindAt: time.Now().Add(time.Hour)},
	}

	r := gin.New()
	r.DELETE("/reminders/:id", middleware.AuthMiddleware(), DeleteReminder)

	w := performJSONRequest(r, http.MethodDelete, "/reminders/r1", nil, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 unauthorized, got %d", w.Code)
	}

	token := makeToken(t, "user-1")
	w = performJSONRequest(r, http.MethodDelete, "/reminders/r1", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for own reminder delete, got %d", w.Code)
	}

	// database.Reminders = []models.Reminder{
	// 	{ID: "r2", UserID: "user-2", EventID: "e2", RemindAt: time.Now().Add(time.Hour)},
	// }

	w = performJSONRequest(r, http.MethodDelete, "/reminders/r2", nil, token)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for other user's reminder, got %d", w.Code)
	}

	w = performJSONRequest(r, http.MethodDelete, "/reminders/r1", nil, token)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing reminder, got %d", w.Code)
	}
}
