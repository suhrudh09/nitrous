package handlers

import (
    "encoding/json"
    "net/http"
    "testing"
    "time"

    "nitrous-backend/config"
    "nitrous-backend/database"
    "nitrous-backend/middleware"
    "nitrous-backend/models"

    "github.com/gin-gonic/gin"
)

func TestUpdateCurrentUserPlanAndRoleFlow(t *testing.T) {
    setupHandlersTestEnv()
    database.Users = []models.User{{ID: "user-1", Email: "u@example.com", Role: "viewer", Plan: "FREE"}}

    r := gin.New()
    r.PUT("/auth/me/plan", middleware.AuthMiddleware(), UpdateCurrentUserPlan)
    r.PUT("/auth/me/role", middleware.AuthMiddleware(), UpdateCurrentUserRole)

    token := makeToken(t, "user-1")

    w := performJSONRequest(r, http.MethodPut, "/auth/me/role", map[string]any{"role": "sponsor"}, token)
    if w.Code != http.StatusBadRequest {
        t.Fatalf("expected 400 when FREE user tries sponsor role, got %d", w.Code)
    }

    w = performJSONRequest(r, http.MethodPut, "/auth/me/plan", map[string]any{"plan": "VIP"}, token)
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for VIP upgrade, got %d", w.Code)
    }

    w = performJSONRequest(r, http.MethodPut, "/auth/me/role", map[string]any{"role": "manager"}, token)
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for manager role after VIP, got %d", w.Code)
    }

    w = performJSONRequest(r, http.MethodPut, "/auth/me/plan", map[string]any{"plan": "FREE"}, token)
    if w.Code != http.StatusBadRequest {
        t.Fatalf("expected 400 for downgrade attempt, got %d", w.Code)
    }
}

func TestCartHandlersFlow(t *testing.T) {
    setupHandlersTestEnv()
    database.Users = []models.User{{ID: "user-1", Email: "u@example.com", Role: "viewer", Plan: "FREE"}}
    database.CartItems = map[string][]models.CartItem{}

    r := gin.New()
    r.GET("/cart", middleware.AuthMiddleware(), GetCart)
    r.PUT("/cart", middleware.AuthMiddleware(), SaveCart)
    r.DELETE("/cart", middleware.AuthMiddleware(), ClearCart)

    w := performJSONRequest(r, http.MethodGet, "/cart", nil, "")
    if w.Code != http.StatusUnauthorized {
        t.Fatalf("expected 401 without auth, got %d", w.Code)
    }

    token := makeToken(t, "user-1")
    dupBody := map[string]any{
        "items": []map[string]any{
            {"merchId": "m1", "quantity": 1, "size": "M"},
            {"merchId": "m1", "quantity": 2, "size": "m"},
        },
    }
    w = performJSONRequest(r, http.MethodPut, "/cart", dupBody, token)
    if w.Code != http.StatusBadRequest {
        t.Fatalf("expected 400 for duplicate cart entries, got %d", w.Code)
    }

    body := map[string]any{
        "items": []map[string]any{
            {"merchId": "m1", "name": "Cap", "price": 50, "quantity": 1, "size": "M"},
            {"merchId": "m2", "name": "Tee", "price": 30, "quantity": 2, "size": "L"},
        },
    }
    w = performJSONRequest(r, http.MethodPut, "/cart", body, token)
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for save cart, got %d", w.Code)
    }

    w = performJSONRequest(r, http.MethodGet, "/cart", nil, token)
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for get cart, got %d", w.Code)
    }

    var cartResp struct {
        Items []models.CartItem `json:"items"`
        Count int               `json:"count"`
    }
    if err := json.Unmarshal(w.Body.Bytes(), &cartResp); err != nil {
        t.Fatalf("failed to decode cart response: %v", err)
    }
    if cartResp.Count != 2 || len(cartResp.Items) != 2 {
        t.Fatalf("expected 2 cart items, got count=%d len=%d", cartResp.Count, len(cartResp.Items))
    }

    w = performJSONRequest(r, http.MethodDelete, "/cart", nil, token)
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for clear cart, got %d", w.Code)
    }

    w = performJSONRequest(r, http.MethodGet, "/cart", nil, token)
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 after clear cart, got %d", w.Code)
    }
    if err := json.Unmarshal(w.Body.Bytes(), &cartResp); err != nil {
        t.Fatalf("failed to decode cart response after clear: %v", err)
    }
    if cartResp.Count != 0 {
        t.Fatalf("expected empty cart after clear, got count=%d", cartResp.Count)
    }
}

func TestPaymentHandlersFlow(t *testing.T) {
    setupHandlersTestEnv()
    database.Users = []models.User{{ID: "user-1", Email: "u@example.com", Role: "viewer", Plan: "FREE"}}
    database.Orders = []models.Order{{ID: "order-1", UserID: "user-1", Status: "created"}}
    database.Payments = nil

    r := gin.New()
    r.POST("/payments/intents", middleware.AuthMiddleware(), CreatePaymentIntent)
    r.POST("/payments/:id/confirm", middleware.AuthMiddleware(), ConfirmPayment)
    r.GET("/payments/:id", middleware.AuthMiddleware(), GetPaymentStatus)
    r.GET("/payments", middleware.AuthMiddleware(), GetUserPayments)

    token := makeToken(t, "user-1")
    createReq := map[string]any{
        "amount":        99.99,
        "currency":      "usd",
        "description":   "Membership",
        "referenceType": "order",
        "referenceId":   "order-1",
    }

    w := performJSONRequest(r, http.MethodPost, "/payments/intents", createReq, token)
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for create payment intent, got %d", w.Code)
    }

    var intentResp PaymentIntentResponse
    if err := json.Unmarshal(w.Body.Bytes(), &intentResp); err != nil {
        t.Fatalf("failed to decode payment intent response: %v", err)
    }
    if intentResp.PaymentID == "" {
        t.Fatalf("expected payment id to be returned")
    }

    w = performJSONRequest(r, http.MethodGet, "/payments/"+intentResp.PaymentID, nil, token)
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for payment status, got %d", w.Code)
    }

    w = performJSONRequest(r, http.MethodPost, "/payments/"+intentResp.PaymentID+"/confirm", nil, token)
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for payment confirm, got %d", w.Code)
    }

    if database.Orders[0].Status != "confirmed" {
        t.Fatalf("expected order status to be confirmed after payment, got %s", database.Orders[0].Status)
    }

    w = performJSONRequest(r, http.MethodGet, "/payments", nil, token)
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for list user payments, got %d", w.Code)
    }
}

func TestNotificationHandlersFallback(t *testing.T) {
    setupHandlersTestEnv()
    database.Users = []models.User{{ID: "user-1", Email: "u@example.com", Role: "viewer", Plan: "FREE"}}

    r := gin.New()
    r.GET("/notifications", middleware.AuthMiddleware(), GetMyNotifications)
    r.PUT("/notifications/:id/read", middleware.AuthMiddleware(), MarkNotificationRead)

    token := makeToken(t, "user-1")

    w := performJSONRequest(r, http.MethodGet, "/notifications", nil, token)
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for notifications list in fallback mode, got %d", w.Code)
    }

    w = performJSONRequest(r, http.MethodPut, "/notifications/any/read", nil, token)
    if w.Code != http.StatusNotImplemented {
        t.Fatalf("expected 501 for mark-read without DB mode, got %d", w.Code)
    }
}

func TestGarageConfigHandlersFlow(t *testing.T) {
    setupHandlersTestEnv()
    database.Users = []models.User{{ID: "user-1", Email: "u@example.com", Role: "viewer", Plan: "FREE"}}
    database.GarageConfigs = map[string][]models.GarageConfig{}

    r := gin.New()
    r.POST("/garage/configs", middleware.AuthMiddleware(), SaveGarageConfig)
    r.GET("/garage/configs", middleware.AuthMiddleware(), GetGarageConfigs)
    r.DELETE("/garage/configs/:id", middleware.AuthMiddleware(), DeleteGarageConfig)

    token := makeToken(t, "user-1")
    req := map[string]any{
        "name":   "Track Build",
        "make":   "Toyota",
        "model":  "Camry",
        "year":   2024,
        "engine": "2.5L",
    }

    w := performJSONRequest(r, http.MethodPost, "/garage/configs", req, token)
    if w.Code != http.StatusCreated {
        t.Fatalf("expected 201 for save garage config, got %d", w.Code)
    }

    var saveResp struct {
        Config models.GarageConfig `json:"config"`
    }
    if err := json.Unmarshal(w.Body.Bytes(), &saveResp); err != nil {
        t.Fatalf("failed to decode save config response: %v", err)
    }
    if saveResp.Config.Tuning != "stock" {
        t.Fatalf("expected default tuning stock, got %s", saveResp.Config.Tuning)
    }

    w = performJSONRequest(r, http.MethodGet, "/garage/configs", nil, token)
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for get configs, got %d", w.Code)
    }

    w = performJSONRequest(r, http.MethodDelete, "/garage/configs/"+saveResp.Config.ID, nil, token)
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for delete config, got %d", w.Code)
    }

    w = performJSONRequest(r, http.MethodDelete, "/garage/configs/"+saveResp.Config.ID, nil, token)
    if w.Code != http.StatusNotFound {
        t.Fatalf("expected 404 when deleting config twice, got %d", w.Code)
    }
}

func TestTeamManagerHandlers(t *testing.T) {
    setupHandlersTestEnv()
    database.Users = []models.User{
        {ID: "mgr-1", Email: "mgr@example.com", Role: "manager", Plan: "VIP"},
        {ID: "viewer-1", Email: "v@example.com", Role: "viewer", Plan: "FREE"},
    }
    database.Teams = []models.Team{{ID: "team-1", Name: "Alpha", Managers: []string{"mgr-1"}, CreatedAt: time.Now()}}

    r := gin.New()
    r.POST("/teams/:id/managers", middleware.AuthMiddleware(), AddTeamManager)
    r.DELETE("/teams/:id/managers/:userId", middleware.AuthMiddleware(), RemoveTeamManager)

    viewerToken := makeToken(t, "viewer-1")
    w := performJSONRequest(r, http.MethodPost, "/teams/team-1/managers", map[string]any{"userId": "mgr-2"}, viewerToken)
    if w.Code != http.StatusForbidden {
        t.Fatalf("expected 403 for non-manager add manager, got %d", w.Code)
    }

    mgrToken := makeToken(t, "mgr-1")
    w = performJSONRequest(r, http.MethodPost, "/teams/team-1/managers", map[string]any{"userId": "mgr-2"}, mgrToken)
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for manager assigning manager, got %d", w.Code)
    }

    w = performJSONRequest(r, http.MethodDelete, "/teams/team-1/managers/mgr-2", nil, mgrToken)
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for manager removal, got %d", w.Code)
    }
}

func TestTeamRelationsMemberAndSponsorHandlers(t *testing.T) {
    setupHandlersTestEnv()
    database.Users = []models.User{
        {ID: "mgr-1", Email: "mgr@example.com", Role: "manager", Plan: "VIP"},
        {ID: "viewer-1", Email: "v@example.com", Role: "viewer", Plan: "FREE"},
    }
    database.Teams = []models.Team{{ID: "team-1", Name: "Alpha", Managers: []string{"mgr-1"}, CreatedAt: time.Now()}}

    r := gin.New()
    r.GET("/teams/:id/members", ListTeamMembers)
    r.POST("/teams/:id/members", middleware.AuthMiddleware(), AddTeamMember)
    r.DELETE("/teams/:id/members/:userId", middleware.AuthMiddleware(), RemoveTeamMember)
    r.GET("/teams/:id/sponsors", ListTeamSponsors)
    r.POST("/teams/:id/sponsors", middleware.AuthMiddleware(), AddTeamSponsor)
    r.DELETE("/teams/:id/sponsors/:userId", middleware.AuthMiddleware(), RemoveTeamSponsor)

    viewerToken := makeToken(t, "viewer-1")
    mgrToken := makeToken(t, "mgr-1")

    w := performJSONRequest(r, http.MethodGet, "/teams/team-1/members", nil, "")
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for list members, got %d", w.Code)
    }

    w = performJSONRequest(r, http.MethodPost, "/teams/team-1/members", map[string]any{"userId": "member-1"}, viewerToken)
    if w.Code != http.StatusForbidden {
        t.Fatalf("expected 403 for viewer adding member, got %d", w.Code)
    }

    w = performJSONRequest(r, http.MethodPost, "/teams/team-1/members", map[string]any{"userId": "member-1"}, mgrToken)
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for manager add member, got %d", w.Code)
    }

    w = performJSONRequest(r, http.MethodDelete, "/teams/team-1/members/member-1", nil, mgrToken)
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for manager remove member, got %d", w.Code)
    }

    w = performJSONRequest(r, http.MethodGet, "/teams/team-1/sponsors", nil, "")
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for list sponsors, got %d", w.Code)
    }

    w = performJSONRequest(r, http.MethodPost, "/teams/team-1/sponsors", map[string]any{"userId": "sponsor-1"}, mgrToken)
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for manager add sponsor, got %d", w.Code)
    }

    w = performJSONRequest(r, http.MethodDelete, "/teams/team-1/sponsors/sponsor-1", nil, mgrToken)
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for manager remove sponsor, got %d", w.Code)
    }
}

func TestPassesCatalogAndMyPassesAndJourneyBookings(t *testing.T) {
    setupHandlersTestEnv()
    database.Users = []models.User{{ID: "user-1", Email: "u@example.com", Role: "viewer", Plan: "FREE"}}
    database.Passes = []database.Pass{{ID: "p1", Tier: "VIP", Event: "Race A", SpotsLeft: 10, TotalSpots: 100, Date: time.Now().UTC().Format(time.RFC3339)}}
    database.PassPurchases = []database.PassPurchase{{UserID: "user-1", PassID: "p1", CreatedAt: time.Now().UTC()}}

    r := gin.New()
    r.GET("/passes/catalog", GetAllPasses)
    r.GET("/passes/my", middleware.AuthMiddleware(), GetMyPasses)
    r.GET("/journeys/my", middleware.AuthMiddleware(), GetMyJourneyBookings)

    w := performJSONRequest(r, http.MethodGet, "/passes/catalog", nil, "")
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for passes catalog, got %d", w.Code)
    }

    token := makeToken(t, "user-1")
    w = performJSONRequest(r, http.MethodGet, "/passes/my", nil, token)
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for my passes, got %d", w.Code)
    }

    var passesResp struct {
        Purchases []UserPassPurchase `json:"purchases"`
        Count     int                `json:"count"`
    }
    if err := json.Unmarshal(w.Body.Bytes(), &passesResp); err != nil {
        t.Fatalf("failed to decode my passes response: %v", err)
    }
    if passesResp.Count != 1 || len(passesResp.Purchases) != 1 {
        t.Fatalf("expected one pass purchase, got count=%d len=%d", passesResp.Count, len(passesResp.Purchases))
    }

    w = performJSONRequest(r, http.MethodGet, "/journeys/my", nil, token)
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for my journeys fallback, got %d", w.Code)
    }
}

func TestGarageSupplementalHandlers(t *testing.T) {
    setupHandlersTestEnv()

    oldClient := httpClient
    httpClient = &http.Client{Transport: mockNHTSATransport()}
    defer func() { httpClient = oldClient }()

    r := gin.New()
    r.GET("/garage/makes", GetGarageMakes)
    r.GET("/garage/models", GetGarageModels)
    r.GET("/garage/trims", GetGarageTrims)
    r.GET("/garage/tuning-configs", GetGarageTuningConfigs)
    r.GET("/garage/search", GetGarageSearch)

    w := performJSONRequest(r, http.MethodGet, "/garage/makes", nil, "")
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for garage makes, got %d", w.Code)
    }

    w = performJSONRequest(r, http.MethodGet, "/garage/models", nil, "")
    if w.Code != http.StatusBadRequest {
        t.Fatalf("expected 400 for missing make on garage models, got %d", w.Code)
    }

    w = performJSONRequest(r, http.MethodGet, "/garage/models?make=Toyota", nil, "")
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for garage models, got %d", w.Code)
    }

    w = performJSONRequest(r, http.MethodGet, "/garage/trims", nil, "")
    if w.Code != http.StatusBadRequest {
        t.Fatalf("expected 400 for missing trim params, got %d", w.Code)
    }

    w = performJSONRequest(r, http.MethodGet, "/garage/trims?make=Toyota&model=Camry&year=2024", nil, "")
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for garage trims, got %d", w.Code)
    }

    w = performJSONRequest(r, http.MethodGet, "/garage/tuning-configs", nil, "")
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for tuning configs, got %d", w.Code)
    }

    w = performJSONRequest(r, http.MethodGet, "/garage/search", nil, "")
    if w.Code != http.StatusBadRequest {
        t.Fatalf("expected 400 for missing search query, got %d", w.Code)
    }

    w = performJSONRequest(r, http.MethodGet, "/garage/search?q=toy", nil, "")
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for garage search, got %d", w.Code)
    }
}

func TestStreamYoutubeAndAdminHandlers(t *testing.T) {
    setupHandlersTestEnv()
    database.Users = []models.User{
        {ID: "admin-1", Email: "admin@example.com", Role: "admin", Plan: "PLATINUM"},
        {ID: "user-1", Email: "u@example.com", Role: "viewer", Plan: "FREE"},
    }

    origOpenF1 := config.AppConfig.OpenF1BaseURL
    origJolpica := config.AppConfig.JolpicaBaseURL
    origSportsDB := config.AppConfig.SportsDBBaseURL
    origSportsKey := config.AppConfig.SportsDBAPIKey
    origYT := config.AppConfig.YouTubeAPIKey
    config.AppConfig.OpenF1BaseURL = "://bad-openf1"
    config.AppConfig.JolpicaBaseURL = "://bad-jolpica"
    config.AppConfig.SportsDBBaseURL = "://bad-sportsdb"
    config.AppConfig.SportsDBAPIKey = "bad"
    config.AppConfig.YouTubeAPIKey = ""
    defer func() {
        config.AppConfig.OpenF1BaseURL = origOpenF1
        config.AppConfig.JolpicaBaseURL = origJolpica
        config.AppConfig.SportsDBBaseURL = origSportsDB
        config.AppConfig.SportsDBAPIKey = origSportsKey
        config.AppConfig.YouTubeAPIKey = origYT
    }()

    r := gin.New()
    r.GET("/streams/openf1/recent", GetOpenF1RecentSessions)
    r.GET("/streams/openf1/telemetry/:sessionKey", GetOpenF1SessionTelemetry)
    r.GET("/youtube/openf1/video", GetOpenF1SessionVideo)
    r.PUT("/admin/users/:id/role", AdminSetUserRole)
    r.POST("/admin/sync", AdminTriggerSync)

    w := performJSONRequest(r, http.MethodGet, "/streams/openf1/recent", nil, "")
    if w.Code != http.StatusBadGateway {
        t.Fatalf("expected 502 for openf1 recent sessions with bad base URL, got %d", w.Code)
    }

    w = performJSONRequest(r, http.MethodGet, "/streams/openf1/telemetry/not-a-number", nil, "")
    if w.Code != http.StatusBadRequest {
        t.Fatalf("expected 400 for invalid session key, got %d", w.Code)
    }

    w = performJSONRequest(r, http.MethodGet, "/youtube/openf1/video", nil, "")
    if w.Code != http.StatusServiceUnavailable {
        t.Fatalf("expected 503 when YouTube API key missing, got %d", w.Code)
    }

    w = performJSONRequest(r, http.MethodPut, "/admin/users/user-1/role", map[string]any{"role": "manager"}, "")
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for admin set user role, got %d", w.Code)
    }

    w = performJSONRequest(r, http.MethodPut, "/admin/users/nope/role", map[string]any{"role": "manager"}, "")
    if w.Code != http.StatusNotFound {
        t.Fatalf("expected 404 for admin set role on missing user, got %d", w.Code)
    }

    w = performJSONRequest(r, http.MethodPost, "/admin/sync", nil, "")
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 for admin trigger sync summary, got %d", w.Code)
    }
}