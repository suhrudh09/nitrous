package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"nitrous-backend/database"
	"nitrous-backend/middleware"
	"nitrous-backend/models"

	"github.com/gin-gonic/gin"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func mockNHTSATransport() http.RoundTripper {
	modelYearRe := regexp.MustCompile(`/modelyear/(\d+)`)

	return roundTripFunc(func(req *http.Request) (*http.Response, error) {
		path := req.URL.Path
		query := req.URL.RawQuery

		status := http.StatusOK
		body := `{"Count":0,"Results":[]}`

		switch {
		case strings.Contains(path, "/GetModelsForMakeYear/"):
			match := modelYearRe.FindStringSubmatch(path)
			if len(match) == 2 {
				year, _ := strconv.Atoi(match[1])
				if year == 2024 || year == 2022 {
					body = `{"Count":1,"Results":[{"Make_ID":448,"Make_Name":"TOYOTA","Model_ID":2469,"Model_Name":"Camry"}]}`
				}
			}
		case strings.Contains(path, "/GetModelsForMake/"):
			body = `{"Count":2,"Results":[{"Make_ID":448,"Make_Name":"TOYOTA","Model_ID":2469,"Model_Name":"Camry"},{"Make_ID":448,"Make_Name":"TOYOTA","Model_ID":2208,"Model_Name":"Corolla"}]}`
		case strings.Contains(path, "/GetMakesForVehicleType/"):
			if strings.Contains(query, "format=json") {
				body = `{"Count":2,"Results":[{"MakeId":448,"MakeName":"TOYOTA"},{"MakeId":460,"MakeName":"FORD"}]}`
			}
		default:
			status = http.StatusNotFound
			body = `{"error":"not found"}`
		}

		return &http.Response{
			StatusCode: status,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    req,
		}, nil
	})
}

func TestGetGarageYearsReturnsActualRange(t *testing.T) {
	setupHandlersTestEnv()

	oldClient := httpClient
	httpClient = &http.Client{Transport: mockNHTSATransport()}
	defer func() { httpClient = oldClient }()

	r := gin.New()
	r.GET("/garage/years", GetGarageYears)

	w := performJSONRequest(r, http.MethodGet, "/garage/years?make=Toyota&model=Camry", nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid garage years, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed decoding years response: %v", err)
	}

	if resp["maxYear"] != "2024" || resp["minYear"] != "2022" {
		t.Fatalf("expected min/max year 2022/2024, got min=%s max=%s", resp["minYear"], resp["maxYear"])
	}
}

func TestGetGarageVehicleReturnsSpec(t *testing.T) {
	setupHandlersTestEnv()

	oldClient := httpClient
	httpClient = &http.Client{Transport: mockNHTSATransport()}
	defer func() { httpClient = oldClient }()

	r := gin.New()
	r.GET("/garage/vehicle", GetGarageVehicle)

	w := performJSONRequest(r, http.MethodGet, "/garage/vehicle?make=Toyota&model=Camry&year=2024", nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid garage vehicle, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Vehicle VehicleSpec `json:"vehicle"`
		Source  string      `json:"source"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed decoding vehicle response: %v", err)
	}

	if resp.Source != "nhtsa" {
		t.Fatalf("expected source nhtsa, got %q", resp.Source)
	}
	if resp.Vehicle.HP <= 0 || resp.Vehicle.TopSpeed <= 0 || resp.Vehicle.Torque <= 0 {
		t.Fatalf("expected non-zero spec values, got hp=%.2f topSpeed=%.2f torque=%.2f", resp.Vehicle.HP, resp.Vehicle.TopSpeed, resp.Vehicle.Torque)
	}
}

func TestPostGarageTuneAppliesConfig(t *testing.T) {
	setupHandlersTestEnv()
	database.Users = []models.User{{ID: "manager-1", Email: "manager@example.com", Role: "manager"}}

	oldClient := httpClient
	httpClient = &http.Client{Transport: mockNHTSATransport()}
	defer func() { httpClient = oldClient }()

	r := gin.New()
	r.POST("/garage/tune", middleware.AuthMiddleware(), middleware.RequireRoles("admin", "manager"), PostGarageTune)
	token := makeToken(t, "manager-1")

	body := map[string]any{
		"make":   "Toyota",
		"model":  "Camry",
		"year":   2024,
		"tuning": "track",
	}
	w := performJSONRequest(r, http.MethodPost, "/garage/tune", body, token)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid garage tune, got %d: %s", w.Code, w.Body.String())
	}

	var resp TuneResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed decoding tune response: %v", err)
	}

	if resp.Tuned.HP <= resp.Base.HP {
		t.Fatalf("expected tuned HP greater than base HP, got tuned=%.2f base=%.2f", resp.Tuned.HP, resp.Base.HP)
	}
	if resp.Config.Label == "" {
		t.Fatalf("expected tune config label to be populated")
	}
}

func TestPurchasePassEndpoint(t *testing.T) {
	setupHandlersTestEnv()
	database.Passes = []database.Pass{
		{ID: "p1", SpotsLeft: 1},
		{ID: "p2", SpotsLeft: 0},
	}

	r := gin.New()
	r.POST("/passes/:id/purchase", middleware.AuthMiddleware(), PurchasePass)

	w := performJSONRequest(r, http.MethodPost, "/passes/p1/purchase", nil, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 unauthorized, got %d", w.Code)
	}

	token := makeToken(t, "user-1")
	w = performJSONRequest(r, http.MethodPost, "/passes/nope/purchase", nil, token)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 pass not found, got %d", w.Code)
	}

	w = performJSONRequest(r, http.MethodPost, "/passes/p2/purchase", nil, token)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 no spots remaining, got %d", w.Code)
	}

	w = performJSONRequest(r, http.MethodPost, "/passes/p1/purchase", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 successful purchase, got %d", w.Code)
	}
	if database.Passes[0].SpotsLeft != 0 {
		t.Fatalf("expected spotsLeft decremented to 0, got %d", database.Passes[0].SpotsLeft)
	}
	if len(database.PassPurchases) != 1 {
		t.Fatalf("expected one recorded pass purchase, got %d", len(database.PassPurchases))
	}

	w = performJSONRequest(r, http.MethodPost, "/passes/p1/purchase", nil, token)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 duplicate purchase, got %d", w.Code)
	}
}
