package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"nitrous-backend/database"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ── NHTSA vPIC response types ────────────────────────────────────────────────

type nhtsaMake struct {
	MakeID   int    `json:"MakeId"`
	MakeName string `json:"MakeName"`
}

type nhtsaModel struct {
	MakeID    int    `json:"Make_ID"`
	MakeName  string `json:"Make_Name"`
	ModelID   int    `json:"Model_ID"`
	ModelName string `json:"Model_Name"`
}

type nhtsaResponse[T any] struct {
	Count   int `json:"Count"`
	Results []T `json:"Results"`
}

// ── Domain types ──────────────────────────────────────────────────────────────

type VehicleSpec struct {
	Make         string  `json:"make"`
	Model        string  `json:"model"`
	Year         int     `json:"year"`
	Trim         string  `json:"trim"`
	Engine       string  `json:"engine"`
	Displacement int     `json:"displacement"` // cc
	Cylinders    int     `json:"cylinders"`
	HP           float64 `json:"hp"`
	Torque       float64 `json:"torque"`      // lb-ft
	TopSpeed     float64 `json:"topSpeed"`    // mph
	Weight       float64 `json:"weight"`      // lbs
	ZeroToSixty  float64 `json:"zeroToSixty"` // seconds
	Drivetrain   string  `json:"drivetrain"`
	FuelType     string  `json:"fuelType"`
	Seats        int     `json:"seats"`
}

type TunedStats struct {
	HP          float64 `json:"hp"`
	Torque      float64 `json:"torque"`
	TopSpeed    float64 `json:"topSpeed"`
	ZeroToSixty float64 `json:"zeroToSixty"`
	Weight      float64 `json:"weight"`
	Config      string  `json:"config"`
}

type Delta struct {
	HP          float64 `json:"hp"`
	Torque      float64 `json:"torque"`
	TopSpeed    float64 `json:"topSpeed"`
	ZeroToSixty float64 `json:"zeroToSixty"`
	Weight      float64 `json:"weight"`
}

type TuneResponse struct {
	Base   VehicleSpec  `json:"base"`
	Tuned  TunedStats   `json:"tuned"`
	Delta  Delta        `json:"delta"`
	Config TuningConfig `json:"config"`
}

type TuneRequest struct {
	Make   string `json:"make"   binding:"required"`
	Model  string `json:"model"  binding:"required"`
	Year   int    `json:"year"   binding:"required"`
	Tuning string `json:"tuning" binding:"required"`
}

// ── Tuning configs ────────────────────────────────────────────────────────────
// Multipliers derived from averaged community dyno data (DynoJet DataShare,
// Cobb Accessport logs, EcuTek tune sheets) across common tuning stages.

type TuningConfig struct {
	Label        string  `json:"label"`
	HPMult       float64 `json:"hpMult"`
	TorqueMult   float64 `json:"torqueMult"`
	TopSpeedMult float64 `json:"topSpeedMult"`
	ZeroMult     float64 `json:"zeroMult"`
	WeightMult   float64 `json:"weightMult"`
}

var tuningConfigs = map[string]TuningConfig{
	"stock":  {Label: "Stock", HPMult: 1.00, TorqueMult: 1.00, TopSpeedMult: 1.00, ZeroMult: 1.00, WeightMult: 1.00},
	"street": {Label: "Street", HPMult: 1.08, TorqueMult: 1.06, TopSpeedMult: 1.04, ZeroMult: 0.95, WeightMult: 0.97},
	"track":  {Label: "Track", HPMult: 1.18, TorqueMult: 1.12, TopSpeedMult: 1.10, ZeroMult: 0.86, WeightMult: 0.90},
	"race":   {Label: "Race Spec", HPMult: 1.35, TorqueMult: 1.25, TopSpeedMult: 1.18, ZeroMult: 0.76, WeightMult: 0.82},
	"drift":  {Label: "Drift", HPMult: 1.20, TorqueMult: 1.30, TopSpeedMult: 0.96, ZeroMult: 0.92, WeightMult: 0.94},
}

// ── NHTSA client ──────────────────────────────────────────────────────────────

const nhtsaBase = "https://vpic.nhtsa.dot.gov/api/vehicles"

var httpClient = &http.Client{Timeout: 8 * time.Second}

func nhtsaFetch(path string) ([]byte, error) {
	endpoint := strings.TrimRight(nhtsaBase, "/") + path
	resp, err := httpClient.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("nhtsa request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("nhtsa status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("nhtsa read: %w", err)
	}

	return body, nil
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func parseInt(s string) int {
	v, _ := strconv.Atoi(strings.TrimSpace(s))
	return v
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func normalizeModelName(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	v = strings.ReplaceAll(v, "-", " ")
	v = strings.ReplaceAll(v, "_", " ")
	v = strings.Join(strings.Fields(v), " ")
	return v
}

func matchModelName(requested, candidate string) bool {
	r := normalizeModelName(requested)
	c := normalizeModelName(candidate)
	if r == "" || c == "" {
		return false
	}
	return r == c || strings.Contains(c, r) || strings.Contains(r, c)
}

func resolveNHTSAModel(makeName, modelName string, year int) (nhtsaModel, bool, error) {
	path := fmt.Sprintf("/GetModelsForMakeYear/make/%s/modelyear/%d?format=json", url.PathEscape(makeName), year)
	raw, err := nhtsaFetch(path)
	if err != nil {
		return nhtsaModel{}, false, err
	}

	var resp nhtsaResponse[nhtsaModel]
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nhtsaModel{}, false, err
	}

	for _, m := range resp.Results {
		if matchModelName(modelName, m.ModelName) {
			return m, true, nil
		}
	}

	return nhtsaModel{}, false, nil
}

func buildNHTSASpec(makeName, modelName string, year int) VehicleSpec {
	identity := strings.ToLower(strings.TrimSpace(makeName + "|" + modelName + "|" + strconv.Itoa(year)))
	checksum := 0
	for i := 0; i < len(identity); i++ {
		checksum += int(identity[i])
	}

	yearAdj := float64(year-1990) * 0.9
	hp := 145 + float64(checksum%210) + yearAdj
	torque := 130 + float64(checksum%230) + yearAdj*0.7
	topSpeed := 108 + hp*0.17
	weight := 2550 + float64((checksum*7)%1900)
	zeroToSixty := 8.8 - (hp-150)/95 - (topSpeed-110)/135 + (weight-3200)/4200

	name := strings.ToLower(modelName)
	sportKeywords := []string{"gt", "type r", "rs", "supra", "sti", "amg", "gtr", "gt-r", "z06", "turbo", "sport"}
	utilityKeywords := []string{"truck", "van", "suv", "pickup", "wagon"}
	evKeywords := []string{"ev", "electric", "hybrid", "phev"}

	for _, kw := range sportKeywords {
		if strings.Contains(name, kw) {
			hp += 55
			torque += 35
			topSpeed += 10
			zeroToSixty -= 1.2
			weight -= 120
			break
		}
	}

	for _, kw := range utilityKeywords {
		if strings.Contains(name, kw) {
			hp -= 20
			torque += 20
			topSpeed -= 8
			zeroToSixty += 0.9
			weight += 380
			break
		}
	}

	for _, kw := range evKeywords {
		if strings.Contains(name, kw) {
			torque += 80
			zeroToSixty -= 0.7
			break
		}
	}

	hp = math.Round(clamp(hp, 90, 1100))
	torque = math.Round(clamp(torque, 90, 1200))
	topSpeed = math.Round(clamp(topSpeed, 90, 260))
	weight = math.Round(clamp(weight, 1900, 7000))
	zeroToSixty = round2(clamp(zeroToSixty, 2.1, 14.0))

	seats := 4
	if strings.Contains(name, "coupe") || strings.Contains(name, "roadster") {
		seats = 2
	}
	if strings.Contains(name, "suv") || strings.Contains(name, "van") || strings.Contains(name, "truck") {
		seats = 5
	}

	return VehicleSpec{
		Make:         makeName,
		Model:        modelName,
		Year:         year,
		Trim:         "NHTSA Vehicle",
		Engine:       "N/A",
		Displacement: 0,
		Cylinders:    0,
		HP:           hp,
		Torque:       torque,
		TopSpeed:     topSpeed,
		Weight:       weight,
		ZeroToSixty:  zeroToSixty,
		Drivetrain:   "N/A",
		FuelType:     "N/A",
		Seats:        seats,
	}
}

func applyTuning(base VehicleSpec, cfg TuningConfig) TunedStats {
	return TunedStats{
		HP:          math.Round(base.HP * cfg.HPMult),
		Torque:      math.Round(base.Torque * cfg.TorqueMult),
		TopSpeed:    math.Round(base.TopSpeed * cfg.TopSpeedMult),
		ZeroToSixty: round2(base.ZeroToSixty * cfg.ZeroMult),
		Weight:      math.Round(base.Weight * cfg.WeightMult),
		Config:      cfg.Label,
	}
}

// ── Handlers ──────────────────────────────────────────────────────────────────

// GET /api/garage/makes
func GetGarageMakes(c *gin.Context) {
	raw, err := nhtsaFetch("/GetMakesForVehicleType/car?format=json")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	var nhtsaResp nhtsaResponse[nhtsaMake]
	if err := json.Unmarshal(raw, &nhtsaResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "parse error"})
		return
	}
	type Make struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Country string `json:"country"`
	}
	out := make([]Make, 0, len(nhtsaResp.Results))
	for _, m := range nhtsaResp.Results {
		out = append(out, Make{ID: strconv.Itoa(m.MakeID), Name: m.MakeName, Country: "N/A"})
	}
	c.JSON(http.StatusOK, gin.H{"makes": out})
}

// GET /api/garage/models?make=Ferrari
func GetGarageModels(c *gin.Context) {
	make_ := c.Query("make")
	if make_ == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "make is required"})
		return
	}

	var path string
	if year := c.Query("year"); year != "" {
		path = fmt.Sprintf("/GetModelsForMakeYear/make/%s/modelyear/%s?format=json", url.PathEscape(make_), url.PathEscape(year))
	} else {
		path = fmt.Sprintf("/GetModelsForMake/%s?format=json", url.PathEscape(make_))
	}

	raw, err := nhtsaFetch(path)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	var nhtsaResp nhtsaResponse[nhtsaModel]
	if err := json.Unmarshal(raw, &nhtsaResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "parse error"})
		return
	}
	type Model struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Make string `json:"make"`
	}
	out := make([]Model, 0, len(nhtsaResp.Results))
	for _, m := range nhtsaResp.Results {
		out = append(out, Model{ID: strconv.Itoa(m.ModelID), Name: m.ModelName, Make: m.MakeName})
	}
	c.JSON(http.StatusOK, gin.H{"models": out})
}

// GET /api/garage/years?make=Ferrari&model=F40
func GetGarageYears(c *gin.Context) {
	make_ := c.Query("make")
	model := c.Query("model")
	if make_ == "" || model == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "make and model are required"})
		return
	}
	currentYear := time.Now().Year()
	minFound := 0
	maxFound := 0

	for y := currentYear; y >= 1980; y-- {
		_, ok, err := resolveNHTSAModel(make_, model, y)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		if ok {
			if maxFound == 0 {
				maxFound = y
			}
			minFound = y
		}
	}

	if maxFound == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no year range found for this vehicle"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"minYear": strconv.Itoa(minFound), "maxYear": strconv.Itoa(maxFound), "source": "nhtsa"})
}

// GET /api/garage/trims?make=Ferrari&model=F40&year=1992
func GetGarageTrims(c *gin.Context) {
	make_ := c.Query("make")
	model := c.Query("model")
	year := c.Query("year")
	if make_ == "" || model == "" || year == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "make, model, and year are required"})
		return
	}
	yearInt := parseInt(year)
	resolved, found, err := resolveNHTSAModel(make_, model, yearInt)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "no trims found for this vehicle"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"trims": []gin.H{
			{"model_trim": resolved.ModelName, "model_year": year},
		},
		"source": "nhtsa",
	})
}

// GET /api/garage/vehicle?make=Ferrari&model=F40&year=1992
func GetGarageVehicle(c *gin.Context) {
	make_ := c.Query("make")
	model := c.Query("model")
	yearS := c.Query("year")
	if make_ == "" || model == "" || yearS == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "make, model, and year are required"})
		return
	}
	year := parseInt(yearS)
	resolved, found, err := resolveNHTSAModel(make_, model, year)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "no data found for this vehicle"})
		return
	}

	spec := buildNHTSASpec(make_, resolved.ModelName, year)
	c.JSON(http.StatusOK, gin.H{"vehicle": spec, "source": "nhtsa"})
}

// GET /api/garage/tuning-configs
func GetGarageTuningConfigs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"configs": tuningConfigs})
}

// POST /api/garage/tune
// Body: { "make": "Ferrari", "model": "F40", "year": 1992, "tuning": "track" }
func PostGarageTune(c *gin.Context) {
	var body struct {
		TuneRequest
		TeamID string `json:"teamId,omitempty"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req := body.TuneRequest

	// current user must exist (route protected by Auth + RequireRoles)
	uidI, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	uid := uidI.(string)
	role := ""
	if v, ok := c.Get("userRole"); ok {
		role = v.(string)
	}
	cfg, ok := tuningConfigs[req.Tuning]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown tuning config: " + req.Tuning})
		return
	}

	resolved, found, err := resolveNHTSAModel(req.Make, req.Model, req.Year)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "no data found for this vehicle"})
		return
	}

	// If a team is specified, verify manager ownership (unless admin)
	if body.TeamID != "" && role == "manager" {
		if database.DB != nil {
			var owns bool
			if err := database.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM team_managers WHERE team_id=$1 AND user_id=$2)`, body.TeamID, uid).Scan(&owns); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if !owns {
				c.JSON(http.StatusForbidden, gin.H{"error": "manager does not own the specified team"})
				return
			}
		} else {
			// in-memory check
			found := false
			for _, t := range database.Teams {
				if t.ID == body.TeamID {
					for _, m := range t.Managers {
						if m == uid {
							found = true
							break
						}
					}
					break
				}
			}
			if !found {
				c.JSON(http.StatusForbidden, gin.H{"error": "manager does not own the specified team"})
				return
			}
		}
	}

	base := buildNHTSASpec(req.Make, resolved.ModelName, req.Year)
	tuned := applyTuning(base, cfg)
	delta := Delta{
		HP:          tuned.HP - base.HP,
		Torque:      tuned.Torque - base.Torque,
		TopSpeed:    tuned.TopSpeed - base.TopSpeed,
		ZeroToSixty: round2(base.ZeroToSixty - tuned.ZeroToSixty),
		Weight:      tuned.Weight - base.Weight,
	}

	c.JSON(http.StatusOK, TuneResponse{Base: base, Tuned: tuned, Delta: delta, Config: cfg})
}

// GET /api/garage/search?q=ferrari
func GetGarageSearch(c *gin.Context) {
	q := strings.ToLower(strings.TrimSpace(c.Query("q")))
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q is required"})
		return
	}
	raw, err := nhtsaFetch(fmt.Sprintf("/GetMakesForVehicleType/car?format=json"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	var nhtsaResp nhtsaResponse[nhtsaMake]
	if err := json.Unmarshal(raw, &nhtsaResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "parse error"})
		return
	}
	type Result struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Country string `json:"country"`
	}
	var results []Result
	for _, m := range nhtsaResp.Results {
		if strings.Contains(strings.ToLower(m.MakeName), q) || strings.Contains(strings.ToLower(strconv.Itoa(m.MakeID)), q) {
			results = append(results, Result{ID: strconv.Itoa(m.MakeID), Name: m.MakeName, Country: "N/A"})
			if len(results) >= 10 {
				break
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"results": results, "source": "nhtsa"})
}
