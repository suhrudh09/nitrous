// database/store.go
// Single source of truth for all in-memory collections.
// Replaces db.go + db_new.go — delete both of those files.
//
// Every exported function acquires the appropriate lock before
// reading or writing, making all handlers safe under concurrent requests.

package database

import (
	"log"
	"nitrous-backend/models"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ── Locks ─────────────────────────────────────────────────────────────────────
// One RWMutex per collection so reads don't block each other.

var (
	eventsMu     sync.RWMutex
	categoriesMu sync.RWMutex
	journeysMu   sync.RWMutex
	merchMu      sync.RWMutex
	usersMu      sync.RWMutex
	teamsMu      sync.RWMutex
	streamsMu    sync.RWMutex
	remindersMu  sync.RWMutex
	ordersMu     sync.RWMutex
	followsMu    sync.RWMutex
)

// ── Backing slices (unexported — access only via functions below) ──────────────

var (
	events     []models.Event
	categories []models.Category
	journeys   []models.Journey
	merchItems []models.MerchItem
	users      []models.User
	teams      []models.Team
	streams    []models.Stream
	reminders  []models.Reminder
	orders     []models.Order
	follows    []models.TeamFollow
)

// ── Init / Close ──────────────────────────────────────────────────────────────

func InitDB() {
	log.Println("Initializing in-memory database...")
	seedEvents()
	seedCategories()
	seedJourneys()
	seedMerch()
	seedTeams()
	seedStreams()
	log.Println("✓ Database initialized with seed data")
}

// InitNewCollections is kept for backward compatibility with main.go.
// It's a no-op because InitDB now seeds everything.
func InitNewCollections() {}

func CloseDB() {
	log.Println("Closing database connection...")
}

// ── Events ────────────────────────────────────────────────────────────────────

func GetEvents() []models.Event {
	eventsMu.RLock()
	defer eventsMu.RUnlock()
	cp := make([]models.Event, len(events))
	copy(cp, events)
	return cp
}

func AppendEvent(e models.Event) {
	eventsMu.Lock()
	defer eventsMu.Unlock()
	events = append(events, e)
}

func UpdateEvent(id string, updated models.Event) bool {
	eventsMu.Lock()
	defer eventsMu.Unlock()
	for i, e := range events {
		if e.ID == id {
			updated.ID = id
			updated.CreatedAt = e.CreatedAt
			events[i] = updated
			return true
		}
	}
	return false
}

func DeleteEvent(id string) bool {
	eventsMu.Lock()
	defer eventsMu.Unlock()
	for i, e := range events {
		if e.ID == id {
			events = append(events[:i], events[i+1:]...)
			return true
		}
	}
	return false
}

func FindEventByID(id string) (models.Event, bool) {
	eventsMu.RLock()
	defer eventsMu.RUnlock()
	for _, e := range events {
		if e.ID == id {
			return e, true
		}
	}
	return models.Event{}, false
}

// ── Categories ────────────────────────────────────────────────────────────────

func GetCategories() []models.Category {
	categoriesMu.RLock()
	defer categoriesMu.RUnlock()
	cp := make([]models.Category, len(categories))
	copy(cp, categories)
	return cp
}

func FindCategoryBySlug(slug string) (models.Category, bool) {
	categoriesMu.RLock()
	defer categoriesMu.RUnlock()
	for _, c := range categories {
		if c.Slug == slug {
			return c, true
		}
	}
	return models.Category{}, false
}

// ── Journeys ──────────────────────────────────────────────────────────────────

func GetJourneys() []models.Journey {
	journeysMu.RLock()
	defer journeysMu.RUnlock()
	cp := make([]models.Journey, len(journeys))
	copy(cp, journeys)
	return cp
}

func FindJourneyByID(id string) (models.Journey, bool) {
	journeysMu.RLock()
	defer journeysMu.RUnlock()
	for _, j := range journeys {
		if j.ID == id {
			return j, true
		}
	}
	return models.Journey{}, false
}

// BookJourney atomically decrements SlotsLeft.
// Returns the updated journey and true, or false if not found / sold out.
func BookJourney(id string) (models.Journey, bool) {
	journeysMu.Lock()
	defer journeysMu.Unlock()
	for i, j := range journeys {
		if j.ID == id {
			if j.SlotsLeft <= 0 {
				return j, false
			}
			journeys[i].SlotsLeft--
			return journeys[i], true
		}
	}
	return models.Journey{}, false
}

// ── Merch ─────────────────────────────────────────────────────────────────────

func GetMerchItems() []models.MerchItem {
	merchMu.RLock()
	defer merchMu.RUnlock()
	cp := make([]models.MerchItem, len(merchItems))
	copy(cp, merchItems)
	return cp
}

func FindMerchByID(id string) (models.MerchItem, bool) {
	merchMu.RLock()
	defer merchMu.RUnlock()
	for _, m := range merchItems {
		if m.ID == id {
			return m, true
		}
	}
	return models.MerchItem{}, false
}

// ── Users ─────────────────────────────────────────────────────────────────────

func FindUserByEmail(email string) (models.User, bool) {
	usersMu.RLock()
	defer usersMu.RUnlock()
	for _, u := range users {
		if u.Email == email {
			return u, true
		}
	}
	return models.User{}, false
}

func FindUserByID(id string) (models.User, bool) {
	usersMu.RLock()
	defer usersMu.RUnlock()
	for _, u := range users {
		if u.ID == id {
			return u, true
		}
	}
	return models.User{}, false
}

func AppendUser(u models.User) {
	usersMu.Lock()
	defer usersMu.Unlock()
	users = append(users, u)
}

// ── Teams ─────────────────────────────────────────────────────────────────────

func GetTeams() []models.Team {
	teamsMu.RLock()
	defer teamsMu.RUnlock()
	cp := make([]models.Team, len(teams))
	copy(cp, teams)
	return cp
}

func FindTeamByID(id string) (models.Team, bool) {
	teamsMu.RLock()
	defer teamsMu.RUnlock()
	for _, t := range teams {
		if t.ID == id {
			return t, true
		}
	}
	return models.Team{}, false
}

// IncrementFollowing bumps a team's follower count and returns the new value.
func IncrementFollowing(teamID string) (int, bool) {
	teamsMu.Lock()
	defer teamsMu.Unlock()
	for i, t := range teams {
		if t.ID == teamID {
			teams[i].Following++
			return teams[i].Following, true
		}
	}
	return 0, false
}

// DecrementFollowing decrements a team's follower count (floor 0).
func DecrementFollowing(teamID string) (int, bool) {
	teamsMu.Lock()
	defer teamsMu.Unlock()
	for i, t := range teams {
		if t.ID == teamID {
			if teams[i].Following > 0 {
				teams[i].Following--
			}
			return teams[i].Following, true
		}
	}
	return 0, false
}

// ── Streams ───────────────────────────────────────────────────────────────────

func GetStreams() []models.Stream {
	streamsMu.RLock()
	defer streamsMu.RUnlock()
	cp := make([]models.Stream, len(streams))
	copy(cp, streams)
	return cp
}

func FindStreamByID(id string) (models.Stream, bool) {
	streamsMu.RLock()
	defer streamsMu.RUnlock()
	for _, s := range streams {
		if s.ID == id {
			return s, true
		}
	}
	return models.Stream{}, false
}

// UpdateStreamTelemetry applies a telemetry update to the matching stream.
// Returns the first stream's ID for convenience (used by SimulateTelemetry).
func UpdateStreamTelemetry(t models.StreamTelemetry) {
	streamsMu.Lock()
	defer streamsMu.Unlock()
	for i, s := range streams {
		if s.ID == t.StreamID {
			streams[i].Viewers = t.Viewers
			streams[i].CurrentLeader = t.CurrentLeader
			streams[i].CurrentSpeed = t.CurrentSpeed
			streams[i].Subtitle = t.Subtitle
			break
		}
	}
}

// FirstStreamSnapshot returns a JSON-ready copy of all streams (for WS connect).
func FirstStream() (models.Stream, bool) {
	streamsMu.RLock()
	defer streamsMu.RUnlock()
	if len(streams) == 0 {
		return models.Stream{}, false
	}
	return streams[0], true
}

// ── Reminders ─────────────────────────────────────────────────────────────────

func GetRemindersByUser(userID string) []models.Reminder {
	remindersMu.RLock()
	defer remindersMu.RUnlock()
	var out []models.Reminder
	for _, r := range reminders {
		if r.UserID == userID {
			out = append(out, r)
		}
	}
	return out
}

func ReminderExists(userID, eventID string) bool {
	remindersMu.RLock()
	defer remindersMu.RUnlock()
	for _, r := range reminders {
		if r.UserID == userID && r.EventID == eventID {
			return true
		}
	}
	return false
}

func AppendReminder(r models.Reminder) {
	remindersMu.Lock()
	defer remindersMu.Unlock()
	reminders = append(reminders, r)
}

func DeleteReminder(userID, eventID string) bool {
	remindersMu.Lock()
	defer remindersMu.Unlock()
	for i, r := range reminders {
		if r.UserID == userID && r.EventID == eventID {
			reminders = append(reminders[:i], reminders[i+1:]...)
			return true
		}
	}
	return false
}

// ── Orders ────────────────────────────────────────────────────────────────────

func AppendOrder(o models.Order) {
	ordersMu.Lock()
	defer ordersMu.Unlock()
	orders = append(orders, o)
}

func GetOrdersByUser(userID string) []models.Order {
	ordersMu.RLock()
	defer ordersMu.RUnlock()
	var out []models.Order
	for _, o := range orders {
		if o.UserID == userID {
			out = append(out, o)
		}
	}
	return out
}

func FindOrderByID(id, userID string) (models.Order, bool, bool) {
	ordersMu.RLock()
	defer ordersMu.RUnlock()
	for _, o := range orders {
		if o.ID == id {
			return o, true, o.UserID == userID
		}
	}
	return models.Order{}, false, false
}

// ── Follows ───────────────────────────────────────────────────────────────────

func FollowExists(userID, teamID string) bool {
	followsMu.RLock()
	defer followsMu.RUnlock()
	for _, f := range follows {
		if f.UserID == userID && f.TeamID == teamID {
			return true
		}
	}
	return false
}

func AppendFollow(f models.TeamFollow) {
	followsMu.Lock()
	defer followsMu.Unlock()
	follows = append(follows, f)
}

func DeleteFollow(userID, teamID string) bool {
	followsMu.Lock()
	defer followsMu.Unlock()
	for i, f := range follows {
		if f.UserID == userID && f.TeamID == teamID {
			follows = append(follows[:i], follows[i+1:]...)
			return true
		}
	}
	return false
}

// ── Seed functions ────────────────────────────────────────────────────────────

func seedEvents() {
	events = []models.Event{
		{ID: uuid.New().String(), Title: "NASCAR Daytona 500", Location: "Daytona International Speedway · Florida", Date: time.Now().Add(10 * 24 * time.Hour), IsLive: true, Category: "motorsport", Time: "15:00 UTC"},
		{ID: uuid.New().String(), Title: "Dakar Rally — Stage 9", Location: "Al Ula → Ha'il · Saudi Arabia", Date: time.Now().Add(-2 * 24 * time.Hour), IsLive: false, Category: "offroad", Time: "09:00 UTC"},
		{ID: uuid.New().String(), Title: "World Dirt Track Championship", Location: "Knob Noster · Missouri, USA", Date: time.Now().Add(5 * 24 * time.Hour), IsLive: true, Category: "motorsport", Time: "18:00 UTC"},
		{ID: uuid.New().String(), Title: "Speed Boat Cup — Finals", Location: "Lake Como · Italy", Date: time.Now().Add(14 * 24 * time.Hour), IsLive: false, Category: "water", Time: "14:00 UTC"},
		{ID: uuid.New().String(), Title: "Red Bull Skydive Series — Rd. 3", Location: "Interlaken Drop Zone · Switzerland", Date: time.Now().Add(20 * 24 * time.Hour), IsLive: false, Category: "air", Time: "11:30 UTC"},
		{ID: uuid.New().String(), Title: "Crop Duster Air Racing", Location: "Bakersfield Airfield · California", Date: time.Now().Add(26 * 24 * time.Hour), IsLive: false, Category: "air", Time: "16:00 UTC"},
	}
}

func seedCategories() {
	categories = []models.Category{
		{ID: uuid.New().String(), Name: "MOTORSPORT", Slug: "motorsport", Icon: "🏎️", LiveCount: 24, Description: "NASCAR · F1 · Dirt · Rally", Color: "cyan"},
		{ID: uuid.New().String(), Name: "WATER", Slug: "water", Icon: "🌊", LiveCount: 8, Description: "Speed Boats · Jet Ski · Surf", Color: "blue"},
		{ID: uuid.New().String(), Name: "AIR & SKY", Slug: "air", Icon: "🪂", LiveCount: 5, Description: "Skydive · Air Race · Wing", Color: "purple"},
		{ID: uuid.New().String(), Name: "OFF-ROAD", Slug: "offroad", Icon: "🏔️", LiveCount: 12, Description: "Dakar · Baja · Enduro", Color: "orange"},
	}
}

func seedJourneys() {
	journeys = []models.Journey{
		{ID: uuid.New().String(), Title: "DAYTONA PIT CREW EXPERIENCE", Category: "MOTORSPORT · BEHIND THE SCENES", Description: "Go behind the wall at Daytona 500. Watch pit stops up close, meet the crew chiefs, and ride the pace car on track.", Badge: "EXCLUSIVE", SlotsLeft: 12, Date: time.Now().Add(10 * 24 * time.Hour), Price: 2400},
		{ID: uuid.New().String(), Title: "DAKAR DESERT CONVOY", Category: "RALLY · DESERT EXPEDITION", Description: "Ride a support vehicle through the Dakar stages. Sleep under the stars, eat with the team, and feel the dust.", Badge: "MEMBERS ONLY", SlotsLeft: 6, Date: time.Now().Add(345 * 24 * time.Hour), Price: 5800},
		{ID: uuid.New().String(), Title: "RED BULL TANDEM SKYDIVE", Category: "AIR · EXTREME SPORT", Description: "Jump with a Red Bull certified instructor at 15,000ft. Camera-equipped, full debrief, and a story you will never forget.", Badge: "LIMITED", SlotsLeft: 3, Date: time.Now().Add(20 * 24 * time.Hour), Price: 1200},
	}
}

func seedMerch() {
	merchItems = []models.MerchItem{
		{ID: uuid.New().String(), Name: "Team Hoodie", Icon: "👕", Price: 89, Category: "apparel"},
		{ID: uuid.New().String(), Name: "NITROUS Cap", Icon: "🧢", Price: 42, Category: "apparel"},
		{ID: uuid.New().String(), Name: "Racing Jacket", Icon: "🏎️", Price: 189, Category: "apparel"},
		{ID: uuid.New().String(), Name: "Pit Watch", Icon: "⌚", Price: 249, Category: "accessories"},
		{ID: uuid.New().String(), Name: "Gear Backpack", Icon: "🎒", Price: 120, Category: "accessories"},
		{ID: uuid.New().String(), Name: "Drop Keychain", Icon: "🏆", Price: 28, Category: "collectibles"},
	}
}

func seedTeams() {
	teams = []models.Team{
		{ID: uuid.New().String(), Name: "Red Bull Racing", Category: "MOTORSPORT · F1", Country: "Austria", Founded: 2005, Rank: 1, Wins: 21, Points: 860, Following: 8200000, Drivers: []string{"Max Verstappen", "Sergio Pérez"}, Color: "red", CreatedAt: time.Now()},
		{ID: uuid.New().String(), Name: "Hendrick Motorsports", Category: "MOTORSPORT · NASCAR", Country: "USA", Founded: 1984, Rank: 2, Wins: 14, Points: 2340, Following: 3100000, Drivers: []string{"Kyle Larson", "Chase Elliott", "William Byron", "Alex Bowman"}, Color: "cyan", CreatedAt: time.Now()},
		{ID: uuid.New().String(), Name: "Toyota Gazoo Racing", Category: "RALLY · WRC", Country: "Japan", Founded: 1957, Rank: 3, Wins: 9, Points: 564, Following: 1800000, Drivers: []string{"Sébastien Ogier", "Elfyn Evans", "Kalle Rovanperä"}, Color: "orange", CreatedAt: time.Now()},
		{ID: uuid.New().String(), Name: "Team Sea Force", Category: "WATER · SPEED BOAT", Country: "Italy", Founded: 2010, Rank: 4, Wins: 7, Points: 320, Following: 420000, Drivers: []string{"F. Bertrand", "L. Capelli"}, Color: "blue", CreatedAt: time.Now()},
		{ID: uuid.New().String(), Name: "Falcon Air Squadron", Category: "AIR · AIR RACING", Country: "France", Founded: 2015, Rank: 5, Wins: 5, Points: 198, Following: 280000, Drivers: []string{"A. Garnier", "B. Morin"}, Color: "purple", CreatedAt: time.Now()},
		{ID: uuid.New().String(), Name: "Baja Iron Squad", Category: "OFF-ROAD · TROPHY TRUCK", Country: "Mexico", Founded: 1998, Rank: 6, Wins: 12, Points: 415, Following: 640000, Drivers: []string{"C. Wedekin", "P. McMillin"}, Color: "gold", CreatedAt: time.Now()},
	}
}

func seedStreams() {
	streams = []models.Stream{
		{ID: uuid.New().String(), Title: "NASCAR Daytona 500", Subtitle: "Lap 87 / 200", Category: "MOTORSPORT", Location: "Daytona International Speedway · FL", Quality: "4K", Viewers: 1200000, IsLive: true, CurrentLeader: "Bubba Wallace #23", CurrentSpeed: "198 mph", Color: "red", CreatedAt: time.Now()},
		{ID: uuid.New().String(), Title: "World Dirt Track Championship", Subtitle: "Heat 3 — Semi Finals", Category: "MOTORSPORT", Location: "Knob Noster · Missouri, USA", Quality: "HD", Viewers: 340000, IsLive: true, CurrentLeader: "Kyle Larson #57", CurrentSpeed: "142 mph", Color: "orange", CreatedAt: time.Now()},
		{ID: uuid.New().String(), Title: "Lake Como Speed Boat Qualifier", Subtitle: "Qualifying Round 2", Category: "WATER", Location: "Lake Como · Italy", Quality: "HD", Viewers: 89000, IsLive: true, CurrentLeader: "F. Bertrand #9", CurrentSpeed: "87 knots", Color: "cyan", CreatedAt: time.Now()},
		{ID: uuid.New().String(), Title: "Red Bull Skydive Series", Subtitle: "Live Drop — 14,800ft", Category: "AIR", Location: "Interlaken Drop Zone · Switzerland", Quality: "HD", Viewers: 220000, IsLive: true, CurrentLeader: "A. Garnier", CurrentSpeed: "120 mph", Color: "purple", CreatedAt: time.Now()},
	}
}
