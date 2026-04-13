package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"nitrous-backend/config"
	"nitrous-backend/database"
	"nitrous-backend/models"
	"sort"
	"strings"
	"time"
)

var externalHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
}

// StartExternalDataIntegration starts background sync loops for external providers.
func StartExternalDataIntegration() {
	externalHTTPClient.Timeout = parseDurationOrDefault(config.AppConfig.ExternalRequestTimeout, 10*time.Second)

	go runJolpicaSyncLoop()
	go runSportsDBSyncLoop()
	go runOpenF1PollLoop()
}

func runJolpicaSyncLoop() {
	interval := parseDurationOrDefault(config.AppConfig.JolpicaSyncInterval, 24*time.Hour)
	for {
		if err := syncJolpicaCalendarAndResults(); err != nil {
			log.Printf("jolpica sync failed: %v", err)
		}
		time.Sleep(interval)
	}
}

func runSportsDBSyncLoop() {
	interval := parseDurationOrDefault(config.AppConfig.SportsDBSyncInterval, 7*24*time.Hour)
	for {
		if err := syncSportsDBTeams(); err != nil {
			log.Printf("sportsdb sync failed: %v", err)
		}
		time.Sleep(interval)
	}
}

func runOpenF1PollLoop() {
	activeInterval := parseDurationOrDefault(config.AppConfig.OpenF1ActiveInterval, 5*time.Second)
	idleInterval := parseDurationOrDefault(config.AppConfig.OpenF1IdleInterval, 1*time.Hour)

	for {
		active, err := syncOpenF1LiveData()
		if err != nil {
			log.Printf("openf1 sync failed: %v", err)
		}

		if active {
			time.Sleep(activeInterval)
		} else {
			time.Sleep(idleInterval)
		}
	}
}

func syncJolpicaCalendarAndResults() error {
	base := strings.TrimRight(config.AppConfig.JolpicaBaseURL, "/")
	calendarURL := fmt.Sprintf("%s/current.json", base)
	resultsURL := fmt.Sprintf("%s/current/last/results.json", base)

	var calendar jolpicaCalendarResponse
	if err := fetchJSON(calendarURL, &calendar); err != nil {
		return fmt.Errorf("calendar fetch: %w", err)
	}

	winnerByRound := map[string]string{}
	var results jolpicaResultsResponse
	if err := fetchJSON(resultsURL, &results); err == nil {
		for _, race := range results.MRData.RaceTable.Races {
			if len(race.Results) == 0 {
				continue
			}
			winner := strings.TrimSpace(race.Results[0].Driver.GivenName + " " + race.Results[0].Driver.FamilyName)
			winnerByRound[race.Round] = winner
		}
	}

	now := time.Now().UTC()
	fresh := make([]models.Event, 0, len(calendar.MRData.RaceTable.Races))
	for _, race := range calendar.MRData.RaceTable.Races {
		raceAt := parseRaceDateTime(race.Date, race.Time)
		location := strings.TrimSpace(fmt.Sprintf("%s - %s", race.Circuit.Location.Locality, race.Circuit.Location.Country))
		title := race.RaceName
		if winner := winnerByRound[race.Round]; winner != "" {
			title = fmt.Sprintf("%s (Winner: %s)", race.RaceName, winner)
		}

		event := models.Event{
			ID:        "jolpica-race-" + race.Round,
			Title:     title,
			Location:  location,
			Date:      raceAt,
			Time:      raceAt.UTC().Format("15:04 UTC"),
			IsLive:    now.After(raceAt.Add(-2*time.Hour)) && now.Before(raceAt.Add(4*time.Hour)),
			Category:  "motorsport",
			CreatedAt: time.Now(),
		}
		fresh = append(fresh, event)
	}

	database.Mu.Lock()
	preserved := make([]models.Event, 0, len(database.Events))
	for _, ev := range database.Events {
		if !strings.HasPrefix(ev.ID, "jolpica-race-") {
			preserved = append(preserved, ev)
		}
	}
	database.Events = append(preserved, fresh...)
	database.Mu.Unlock()

	log.Printf("jolpica sync complete: merged %d events", len(fresh))
	return nil
}

func syncSportsDBTeams() error {
	base := strings.TrimRight(config.AppConfig.SportsDBBaseURL, "/")
	key := strings.TrimSpace(config.AppConfig.SportsDBAPIKey)
	if key == "" {
		key = "123"
	}

	leagues := []string{
		"Formula 1",
		"NASCAR Cup Series",
		"World Rally Championship",
		"MotoGP",
	}

	collected := map[string]models.Team{}
	for _, league := range leagues {
		endpoint := fmt.Sprintf("%s/%s/search_all_teams.php?l=%s", base, key, url.QueryEscape(league))
		var resp sportsDBTeamsResponse
		if err := fetchJSON(endpoint, &resp); err != nil {
			log.Printf("sportsdb league fetch failed (%s): %v", league, err)
			continue
		}

		for _, t := range resp.Teams {
			if strings.TrimSpace(t.IDTeam) == "" || strings.TrimSpace(t.StrTeam) == "" {
				continue
			}
			id := "sportsdb-" + strings.TrimSpace(t.IDTeam)
			collected[id] = models.Team{
				ID:             id,
				Name:           strings.TrimSpace(t.StrTeam),
				Country:        strings.TrimSpace(t.StrCountry),
				Followers:      []string{},
				FollowersCount: 0,
				CreatedAt:      time.Now(),
			}
		}
	}

	if len(collected) == 0 {
		return nil
	}

	database.Mu.Lock()
	existingByName := map[string]models.Team{}
	preserved := make([]models.Team, 0, len(database.Teams))
	for _, team := range database.Teams {
		existingByName[strings.ToLower(team.Name)] = team
		if !strings.HasPrefix(team.ID, "sportsdb-") {
			preserved = append(preserved, team)
		}
	}

	merged := make([]models.Team, 0, len(collected))
	for _, team := range collected {
		if old, ok := existingByName[strings.ToLower(team.Name)]; ok {
			team.Followers = old.Followers
			team.FollowersCount = old.FollowersCount
			if !old.CreatedAt.IsZero() {
				team.CreatedAt = old.CreatedAt
			}
		}
		merged = append(merged, team)
	}
	sort.Slice(merged, func(i, j int) bool { return merged[i].Name < merged[j].Name })
	database.Teams = append(preserved, merged...)
	database.Mu.Unlock()

	log.Printf("sportsdb sync complete: merged %d teams", len(merged))
	return nil
}

func syncOpenF1LiveData() (bool, error) {
	session, ok, err := fetchOpenF1Session()
	if err != nil {
		return false, err
	}
	if !ok {
		updateOpenF1Stream(openF1Session{}, false, "No active race session", "0 km/h", "Standby", 0)
		return false, nil
	}

	active := isSessionActive(session)
	leader, speed, rpm, gear := fetchOpenF1TelemetrySummary(session.SessionKey)
	currentSpeed := fmt.Sprintf("%d km/h", speed)
	subtitle := fmt.Sprintf("%s - %s", session.SessionName, strings.TrimSpace(session.CircuitShortName))
	if subtitle == " - " {
		subtitle = "Live timing via OpenF1"
	}
	viewers := 0
	if active {
		viewers = 18000 + (speed % 4000)
	}

	streamID := updateOpenF1Stream(session, active, subtitle, currentSpeed, leader, viewers)
	if active {
		gForce := math.Min(4.5, float64(speed)/120.0)
		BroadcastTelemetry(streamID, speed, rpm, gear, gForce)
	}

	return active, nil
}

func fetchOpenF1Session() (openF1Session, bool, error) {
	base := strings.TrimRight(config.AppConfig.OpenF1BaseURL, "/")
	year := time.Now().Year()
	endpoint := fmt.Sprintf("%s/sessions?session_name=Race&year=%d", base, year)

	var sessions []openF1Session
	if err := fetchJSON(endpoint, &sessions); err != nil {
		return openF1Session{}, false, fmt.Errorf("session fetch: %w", err)
	}
	if len(sessions) == 0 {
		return openF1Session{}, false, nil
	}

	sort.Slice(sessions, func(i, j int) bool {
		return parseRFC3339OrZero(sessions[i].DateStart).After(parseRFC3339OrZero(sessions[j].DateStart))
	})

	now := time.Now().UTC()
	for _, s := range sessions {
		if isSessionActive(s) {
			return s, true, nil
		}
	}

	for _, s := range sessions {
		if parseRFC3339OrZero(s.DateStart).After(now.Add(-14 * 24 * time.Hour)) {
			return s, true, nil
		}
	}

	return sessions[0], true, nil
}

// GetOpenF1RecentRaceSessions fetches recent race sessions for a given year.
func GetOpenF1RecentRaceSessions(limit int, year int) ([]openF1Session, error) {
	base := strings.TrimRight(config.AppConfig.OpenF1BaseURL, "/")
	if year <= 0 {
		year = time.Now().Year()
	}
	if limit <= 0 {
		limit = 8
	}

	endpoint := fmt.Sprintf("%s/sessions?session_name=Race&year=%d", base, year)
	var sessions []openF1Session
	if err := fetchJSON(endpoint, &sessions); err != nil {
		return nil, fmt.Errorf("recent sessions fetch: %w", err)
	}

	sort.Slice(sessions, func(i, j int) bool {
		return parseRFC3339OrZero(sessions[i].DateStart).After(parseRFC3339OrZero(sessions[j].DateStart))
	})

	now := time.Now().UTC()
	past := make([]openF1Session, 0, len(sessions))
	for _, s := range sessions {
		start := parseRFC3339OrZero(s.DateStart)
		if !start.IsZero() && start.Before(now) {
			past = append(past, s)
		}
	}

	if len(past) == 0 {
		past = sessions
	}

	if len(past) > limit {
		past = past[:limit]
	}

	return past, nil
}

// OpenF1SessionTelemetry is a snapshot payload for a specific OpenF1 session.
type OpenF1SessionTelemetry struct {
	SessionKey    int     `json:"session_key"`
	CurrentLeader string  `json:"current_leader"`
	SpeedKPH      int     `json:"speed_kph"`
	RPM           int     `json:"rpm"`
	Gear          int     `json:"gear"`
	GForce        float64 `json:"g_force"`
	CapturedAt    string  `json:"captured_at"`
}

// FetchOpenF1SessionTelemetry fetches a telemetry snapshot for a specific OpenF1 session key.
func FetchOpenF1SessionTelemetry(sessionKey int) (OpenF1SessionTelemetry, error) {
	if sessionKey <= 0 {
		return OpenF1SessionTelemetry{}, fmt.Errorf("invalid session key")
	}

	var windowStart, windowEnd time.Time
	if session, ok, err := fetchOpenF1SessionByKey(sessionKey); err == nil && ok {
		start := parseRFC3339OrZero(session.DateStart)
		end := parseRFC3339OrZero(session.DateEnd)
		if !start.IsZero() {
			windowStart = start
		}
		if !end.IsZero() {
			windowEnd = end
		}
	}

	leader, speed, rpm, gear := fetchOpenF1TelemetrySummaryWithWindow(sessionKey, windowStart, windowEnd)
	gForce := math.Min(4.5, float64(speed)/120.0)

	return OpenF1SessionTelemetry{
		SessionKey:    sessionKey,
		CurrentLeader: leader,
		SpeedKPH:      speed,
		RPM:           rpm,
		Gear:          gear,
		GForce:        gForce,
		CapturedAt:    time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func fetchOpenF1TelemetrySummary(sessionKey int) (leader string, speed int, rpm int, gear int) {
	return fetchOpenF1TelemetrySummaryWithWindow(sessionKey, time.Time{}, time.Time{})
}

func fetchOpenF1TelemetrySummaryWithWindow(sessionKey int, windowStart, windowEnd time.Time) (leader string, speed int, rpm int, gear int) {
	leader = "Standby"

	base := strings.TrimRight(config.AppConfig.OpenF1BaseURL, "/")
	driverNames := fetchOpenF1DriverNames(sessionKey)

	queryWindowStart := time.Time{}
	queryWindowEnd := time.Time{}
	if !windowStart.IsZero() {
		queryWindowStart = windowStart
	}
	if !windowEnd.IsZero() {
		queryWindowEnd = windowEnd
	}
	if queryWindowStart.IsZero() && !queryWindowEnd.IsZero() {
		queryWindowStart = queryWindowEnd.Add(-20 * time.Minute)
	}
	if queryWindowEnd.IsZero() && !queryWindowStart.IsZero() {
		queryWindowEnd = queryWindowStart.Add(20 * time.Minute)
	}
	if queryWindowStart.IsZero() && queryWindowEnd.IsZero() {
		queryWindowEnd = time.Now().UTC()
		queryWindowStart = queryWindowEnd.Add(-20 * time.Minute)
	}

	windowParams := ""
	if !queryWindowStart.IsZero() {
		windowParams += "&date>%3D" + url.QueryEscape(queryWindowStart.UTC().Format(time.RFC3339))
	}
	if !queryWindowEnd.IsZero() {
		windowParams += "&date<%3D" + url.QueryEscape(queryWindowEnd.UTC().Format(time.RFC3339))
	}

	var positions []openF1Position
	posURL := fmt.Sprintf("%s/position?session_key=%d%s", base, sessionKey, windowParams)
	if err := fetchJSON(posURL, &positions); err == nil {
		latestLeadDate := time.Time{}
		latestLeaderNumber := 0
		for _, p := range positions {
			if p.Position != 1 {
				continue
			}
			ts := parseRFC3339OrZero(p.Date)
			if ts.After(latestLeadDate) {
				latestLeadDate = ts
				latestLeaderNumber = p.DriverNumber
			}
		}
		if latestLeaderNumber > 0 {
			if name, ok := driverNames[latestLeaderNumber]; ok && strings.TrimSpace(name) != "" {
				leader = name
			} else {
				leader = fmt.Sprintf("Driver #%d", latestLeaderNumber)
			}
		}
	}

	var cars []openF1CarData
	carURL := fmt.Sprintf("%s/car_data?session_key=%d%s", base, sessionKey, windowParams)
	if err := fetchJSON(carURL, &cars); err == nil {
		latestMoving := time.Time{}
		latestAny := time.Time{}
		latestMovingSpeed := 0
		latestMovingRPM := 0
		latestMovingGear := 0
		latestAnySpeed := 0
		latestAnyRPM := 0
		latestAnyGear := 0
		for _, c := range cars {
			ts := parseRFC3339OrZero(c.Date)
			if ts.After(latestAny) {
				latestAny = ts
				latestAnySpeed = c.Speed
				latestAnyRPM = c.RPM
				latestAnyGear = c.NGear
			}
			if c.Speed > 0 && ts.After(latestMoving) {
				latestMoving = ts
				latestMovingSpeed = c.Speed
				latestMovingRPM = c.RPM
				latestMovingGear = c.NGear
			}
		}

		if !latestMoving.IsZero() {
			speed = latestMovingSpeed
			rpm = latestMovingRPM
			gear = latestMovingGear
		} else if !latestAny.IsZero() {
			speed = latestAnySpeed
			rpm = latestAnyRPM
			gear = latestAnyGear
		}
	}

	if speed <= 0 {
		speed = 0
	}
	return leader, speed, rpm, gear
}

func fetchOpenF1SessionByKey(sessionKey int) (openF1Session, bool, error) {
	base := strings.TrimRight(config.AppConfig.OpenF1BaseURL, "/")
	endpoint := fmt.Sprintf("%s/sessions?session_key=%d", base, sessionKey)

	var sessions []openF1Session
	if err := fetchJSON(endpoint, &sessions); err != nil {
		return openF1Session{}, false, err
	}
	if len(sessions) == 0 {
		return openF1Session{}, false, nil
	}

	return sessions[0], true, nil
}

func fetchOpenF1DriverNames(sessionKey int) map[int]string {
	base := strings.TrimRight(config.AppConfig.OpenF1BaseURL, "/")
	endpoint := fmt.Sprintf("%s/drivers?session_key=%d", base, sessionKey)

	var drivers []openF1Driver
	if err := fetchJSON(endpoint, &drivers); err != nil {
		return map[int]string{}
	}

	names := make(map[int]string, len(drivers))
	for _, d := range drivers {
		if d.DriverNumber <= 0 {
			continue
		}
		name := strings.TrimSpace(d.BroadcastName)
		if name == "" {
			name = strings.TrimSpace(d.FullName)
		}
		if name == "" {
			name = strings.TrimSpace(d.NameAcronym)
		}
		if name != "" {
			names[d.DriverNumber] = name
		}
	}

	return names
}

func updateOpenF1Stream(session openF1Session, isLive bool, subtitle, currentSpeed, leader string, viewers int) string {
	streamID := "openf1-live"
	watchOptions := buildOpenF1WatchOptions()

	streamsMu.Lock()
	defer streamsMu.Unlock()

	for i := range streams {
		if streams[i].ID == streamID {
			streams[i].IsLive = isLive
			streams[i].Subtitle = subtitle
			streams[i].CurrentSpeed = currentSpeed
			streams[i].CurrentLeader = leader
			streams[i].Viewers = viewers
			streams[i].ExternalWatch = watchOptions
			streams[i].DateStart = strings.TrimSpace(session.DateStart)
			streams[i].DateEnd = strings.TrimSpace(session.DateEnd)
			streams[i].CountryName = strings.TrimSpace(session.CountryName)
			streams[i].CircuitShort = strings.TrimSpace(session.CircuitShortName)
			streams[i].CreatedAt = time.Now().UTC().Format(time.RFC3339)
			return streamID
		}
	}

	streams = append(streams, Stream{
		ID:            streamID,
		EventID:       "jolpica-next",
		Title:         "Formula 1 Live Timing",
		Subtitle:      subtitle,
		PlaybackURL:   "https://cdn.pixabay.com/video/2015/11/09/1295-145209438_large.mp4",
		ExternalWatch: watchOptions,
		DateStart:     strings.TrimSpace(session.DateStart),
		DateEnd:       strings.TrimSpace(session.DateEnd),
		CountryName:   strings.TrimSpace(session.CountryName),
		CircuitShort:  strings.TrimSpace(session.CircuitShortName),
		Category:      "motorsport",
		Location:      "OpenF1 Feed",
		Quality:       "HD",
		Viewers:       viewers,
		IsLive:        isLive,
		CurrentLeader: leader,
		CurrentSpeed:  currentSpeed,
		Color:         "cyan",
		CreatedAt:     time.Now().UTC().Format(time.RFC3339),
	})

	return streamID
}

func buildOpenF1WatchOptions() []ExternalWatchOption {
	options := []ExternalWatchOption{}
	if u := strings.TrimSpace(config.AppConfig.F1YouTubeLiveURL); u != "" {
		options = append(options, ExternalWatchOption{
			Platform: "youtube",
			Label:    "Watch races on YouTube",
			URL:      u,
		})
	}
	if u := strings.TrimSpace(config.AppConfig.F1TwitchLiveURL); u != "" {
		options = append(options, ExternalWatchOption{
			Platform: "twitch",
			Label:    "Watch races on Twitch",
			URL:      u,
		})
	}
	return options
}

func fetchJSON(endpoint string, target interface{}) error {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := externalHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return err
	}

	return nil
}

func parseDurationOrDefault(value string, fallback time.Duration) time.Duration {
	d, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil || d <= 0 {
		return fallback
	}
	return d
}

func parseRaceDateTime(dateStr, timeStr string) time.Time {
	dateStr = strings.TrimSpace(dateStr)
	timeStr = strings.TrimSpace(strings.TrimSuffix(timeStr, "Z"))
	if dateStr == "" {
		return time.Now().UTC()
	}
	if timeStr == "" {
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return time.Now().UTC()
		}
		return t.UTC()
	}
	combined := fmt.Sprintf("%sT%sZ", dateStr, timeStr)
	t, err := time.Parse(time.RFC3339, combined)
	if err != nil {
		return time.Now().UTC()
	}
	return t.UTC()
}

func parseRFC3339OrZero(value string) time.Time {
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}
	}
	return t
}

func isSessionActive(session openF1Session) bool {
	start := parseRFC3339OrZero(session.DateStart)
	end := parseRFC3339OrZero(session.DateEnd)
	if start.IsZero() {
		return false
	}
	if end.IsZero() {
		end = start.Add(3 * time.Hour)
	}
	now := time.Now().UTC()
	return now.After(start.Add(-10*time.Minute)) && now.Before(end.Add(20*time.Minute))
}

type jolpicaCalendarResponse struct {
	MRData struct {
		RaceTable struct {
			Races []struct {
				Round    string `json:"round"`
				RaceName string `json:"raceName"`
				Date     string `json:"date"`
				Time     string `json:"time"`
				Circuit  struct {
					Location struct {
						Locality string `json:"locality"`
						Country  string `json:"country"`
					} `json:"Location"`
				} `json:"Circuit"`
			} `json:"Races"`
		} `json:"RaceTable"`
	} `json:"MRData"`
}

type jolpicaResultsResponse struct {
	MRData struct {
		RaceTable struct {
			Races []struct {
				Round   string `json:"round"`
				Results []struct {
					Driver struct {
						GivenName  string `json:"givenName"`
						FamilyName string `json:"familyName"`
					} `json:"Driver"`
				} `json:"Results"`
			} `json:"Races"`
		} `json:"RaceTable"`
	} `json:"MRData"`
}

type sportsDBTeamsResponse struct {
	Teams []struct {
		IDTeam     string `json:"idTeam"`
		StrTeam    string `json:"strTeam"`
		StrCountry string `json:"strCountry"`
	} `json:"teams"`
}

type openF1Session struct {
	SessionKey       int    `json:"session_key"`
	SessionName      string `json:"session_name"`
	DateStart        string `json:"date_start"`
	DateEnd          string `json:"date_end"`
	CountryName      string `json:"country_name"`
	CircuitShortName string `json:"circuit_short_name"`
}

type openF1Position struct {
	Date         string `json:"date"`
	DriverNumber int    `json:"driver_number"`
	Position     int    `json:"position"`
}

type openF1CarData struct {
	Date         string `json:"date"`
	DriverNumber int    `json:"driver_number"`
	Speed        int    `json:"speed"`
	RPM          int    `json:"rpm"`
	NGear        int    `json:"n_gear"`
}

type openF1Driver struct {
	DriverNumber int    `json:"driver_number"`
	BroadcastName string `json:"broadcast_name"`
	FullName      string `json:"full_name"`
	NameAcronym   string `json:"name_acronym"`
}
