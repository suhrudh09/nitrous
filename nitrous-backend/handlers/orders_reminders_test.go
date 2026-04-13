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

	validReq := map[string]any{"merchItemId": "m1", "quantity": 2}

	w := performJSONRequest(r, http.MethodPost, "/orders", validReq, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 unauthorized, got %d", w.Code)
	}

	token := makeToken(t, "user-1")
	invalidReq := map[string]any{"merchItemId": "m1", "quantity": 0}
	w = performJSONRequest(r, http.MethodPost, "/orders", invalidReq, token)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 invalid payload, got %d", w.Code)
	}

	notFoundReq := map[string]any{"merchItemId": "unknown", "quantity": 1}
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
	if order.TotalPrice != 100 {
		t.Fatalf("expected total price 100, got %.2f", order.TotalPrice)
	}
}

func TestGetMyOrdersEndpoint(t *testing.T) {
	setupHandlersTestEnv()
	database.Orders = []models.Order{
		{ID: "o1", UserID: "user-1", MerchItemID: "m1", Quantity: 1, TotalPrice: 10},
		{ID: "o2", UserID: "user-2", MerchItemID: "m2", Quantity: 1, TotalPrice: 20},
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
		{ID: "o1", UserID: "user-1", MerchItemID: "m1", Quantity: 1, TotalPrice: 10},
		{ID: "o2", UserID: "user-2", MerchItemID: "m2", Quantity: 1, TotalPrice: 20},
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
		{ID: "o1", UserID: "user-1", MerchItemID: "m1", Quantity: 1, TotalPrice: 10, Status: "created"},
		{ID: "o2", UserID: "user-2", MerchItemID: "m2", Quantity: 1, TotalPrice: 20, Status: "created"},
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

func TestCreateOrderFrontendPayload(t *testing.T) {
	setupHandlersTestEnv()
	database.MerchItems = []models.MerchItem{
		{ID: "m1", Name: "Cap", Price: 50},
		{ID: "m2", Name: "Jacket", Price: 100},
	}

	r := gin.New()
	r.POST("/orders", middleware.AuthMiddleware(), CreateOrder)

	token := makeToken(t, "user-1")
	req := map[string]any{
		"items": []map[string]any{
			{"merchId": "m1", "quantity": 2},
			{"merchId": "m2", "quantity": 1},
		},
	}

	w := performJSONRequest(r, http.MethodPost, "/orders", req, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 for cart payload, got %d", w.Code)
	}

	var resp struct {
		Message string        `json:"message"`
		Order   orderResponse `json:"order"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode compat order response: %v", err)
	}
	if resp.Order.Total != 200 {
		t.Fatalf("expected total 200, got %.2f", resp.Order.Total)
	}
	if len(resp.Order.Items) != 2 {
		t.Fatalf("expected 2 order items, got %d", len(resp.Order.Items))
	}
}

func TestEventReminderCompatEndpoints(t *testing.T) {
	setupHandlersTestEnv()
	database.Events = []models.Event{sampleEvent("event-1")}

	r := gin.New()
	r.POST("/events/:id/remind", middleware.AuthMiddleware(), SetEventReminderCompat)
	r.DELETE("/events/:id/remind", middleware.AuthMiddleware(), DeleteEventReminderCompat)

	token := makeToken(t, "user-1")

	w := performJSONRequest(r, http.MethodPost, "/events/event-1/remind", nil, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 for compat reminder create, got %d", w.Code)
	}
	if len(database.Reminders) != 1 {
		t.Fatalf("expected 1 reminder, got %d", len(database.Reminders))
	}

	w = performJSONRequest(r, http.MethodDelete, "/events/event-1/remind", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for compat reminder delete, got %d", w.Code)
	}
	if len(database.Reminders) != 0 {
		t.Fatalf("expected 0 reminders after delete, got %d", len(database.Reminders))
	}
}
