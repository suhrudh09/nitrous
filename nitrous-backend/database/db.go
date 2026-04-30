package database

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"nitrous-backend/models"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

//go:embed schema.sql seed.sql
var dbFiles embed.FS

// In-memory storage for prototype
// Replace with actual DB connection for production
var (
	DB            *sql.DB
	Mu            sync.RWMutex
	Events        []models.Event
	Categories    []models.Category
	Journeys      []models.Journey
	Passes        []Pass
	PassPurchases []PassPurchase
	MerchItems    []models.MerchItem
	Users         []models.User
	Teams         []models.Team
	Reminders     []models.Reminder
	Orders        []models.Order
	CartItems     map[string][]models.CartItem
	GarageConfigs map[string][]models.GarageConfig
	Payments      []models.Payment
)

type Pass struct {
	ID         string
	Tier       string
	Event      string
	Location   string
	Date       string
	Category   string
	Price      float64
	Perks      []string
	SpotsLeft  int
	TotalSpots int
	Badge      *string
	TierColor  string
}

type PassPurchase struct {
	UserID    string
	PassID    string
	CreatedAt time.Time
}

func InitDB() {
	Mu.Lock()
	defer Mu.Unlock()

	if err := initPostgres(); err != nil {
		log.Fatalf("PostgreSQL unavailable: %v", err)
	}

	if err := migratePostgres(); err != nil {
		log.Fatalf("PostgreSQL migration failed: %v", err)
	}

	if err := loadSeedDataFromPostgres(); err != nil {
		log.Fatalf("PostgreSQL seed load failed: %v", err)
	}

	log.Println("✓ Connected to PostgreSQL and loaded seed data")
}

func initPostgres() error {
	dsn := postgresDSN()
	if dsn == "" {
		return fmt.Errorf("no PostgreSQL connection string configured")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}

	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxIdleConns(2)
	db.SetMaxOpenConns(10)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return err
	}

	DB = db
	return nil
}

func postgresDSN() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		// Add sslmode=disable if not already present
		if !strings.Contains(url, "sslmode=") {
			url += "?sslmode=disable"
		}
		return url
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if host == "" || user == "" || dbName == "" {
		return ""
	}
	if port == "" {
		port = "5432"
	}

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbName)
}

func migratePostgres() error {
	if DB == nil {
		return fmt.Errorf("postgres connection is not initialized")
	}

	schemaSQL, err := dbFiles.ReadFile("schema.sql")
	if err != nil {
		return err
	}
	if _, err := DB.Exec(string(schemaSQL)); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}

	if _, err := DB.Exec(`ALTER TABLE reminders ADD COLUMN IF NOT EXISTS notified_at TIMESTAMPTZ`); err != nil {
		return fmt.Errorf("ensure reminders.notified_at column: %w", err)
	}

	if _, err := DB.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS plan TEXT NOT NULL DEFAULT 'FREE'`); err != nil {
		return fmt.Errorf("ensure users.plan column: %w", err)
	}

	if _, err := DB.Exec(`ALTER TABLE pass_purchases ADD COLUMN IF NOT EXISTS quantity INT NOT NULL DEFAULT 1`); err != nil {
		return fmt.Errorf("ensure pass_purchases.quantity column: %w", err)
	}

	if _, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS journey_bookings (
			id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id     UUID NOT NULL,
			journey_id  UUID NOT NULL,
			quantity    INT  NOT NULL DEFAULT 1,
			booked_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT fk_jb_user    FOREIGN KEY (user_id)    REFERENCES users(id)    ON DELETE CASCADE,
			CONSTRAINT fk_jb_journey FOREIGN KEY (journey_id) REFERENCES journeys(id) ON DELETE CASCADE
		)
	`); err != nil {
		return fmt.Errorf("ensure journey_bookings table: %w", err)
	}

	if _, err := DB.Exec(`CREATE INDEX IF NOT EXISTS idx_journey_bookings_user_id ON journey_bookings(user_id)`); err != nil {
		return fmt.Errorf("ensure journey_bookings user index: %w", err)
	}

	seedSQL, err := dbFiles.ReadFile("seed.sql")
	if err != nil {
		return err
	}
	if _, err := DB.Exec(string(seedSQL)); err != nil {
		return fmt.Errorf("apply seed data: %w", err)
	}

	return nil
}

func loadSeedDataFromPostgres() error {
	if DB == nil {
		return fmt.Errorf("postgres connection is not initialized")
	}

	if err := loadUsersFromPostgres(); err != nil {
		return err
	}
	if err := loadTeamsFromPostgres(); err != nil {
		return err
	}
	if err := loadCategoriesFromPostgres(); err != nil {
		return err
	}
	if err := loadEventsFromPostgres(); err != nil {
		return err
	}
	if err := loadJourneysFromPostgres(); err != nil {
		return err
	}
	if err := loadMerchFromPostgres(); err != nil {
		return err
	}
	if err := loadPassesFromPostgres(); err != nil {
		return err
	}

	Reminders = []models.Reminder{}
	Orders = []models.Order{}
	CartItems = map[string][]models.CartItem{}
	PassPurchases = []PassPurchase{}

	return nil
}

func loadUsersFromPostgres() error {
	rows, err := DB.Query(`SELECT id, email, password_hash, role, name, COALESCE(plan, 'FREE'), created_at FROM users ORDER BY email`)
	if err != nil {
		return err
	}
	defer rows.Close()

	Users = make([]models.User, 0)
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.Name, &user.Plan, &user.CreatedAt); err != nil {
			return err
		}
		Users = append(Users, user)
	}

	return rows.Err()
}

func loadTeamsFromPostgres() error {
	rows, err := DB.Query(`SELECT id, name, COALESCE(country, ''), is_private, followers_count, created_at FROM teams ORDER BY name`)
	if err != nil {
		return err
	}
	defer rows.Close()

	Teams = make([]models.Team, 0)
	for rows.Next() {
		var team models.Team
		if err := rows.Scan(&team.ID, &team.Name, &team.Country, &team.IsPrivate, &team.FollowersCount, &team.CreatedAt); err != nil {
			return err
		}
		Teams = append(Teams, team)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	teamByID := make(map[string]*models.Team, len(Teams))
	for i := range Teams {
		teamByID[Teams[i].ID] = &Teams[i]
	}

	if err := loadTeamStringRelations(`SELECT team_id, user_id FROM team_managers ORDER BY team_id, user_id`, func(teamID, value string) {
		if team := teamByID[teamID]; team != nil {
			team.Managers = append(team.Managers, value)
		}
	}); err != nil {
		return err
	}
	if err := loadTeamStringRelations(`SELECT team_id, user_id FROM team_members ORDER BY team_id, user_id`, func(teamID, value string) {
		if team := teamByID[teamID]; team != nil {
			team.Members = append(team.Members, value)
		}
	}); err != nil {
		return err
	}
	if err := loadTeamStringRelations(`SELECT team_id, user_id FROM team_sponsors ORDER BY team_id, user_id`, func(teamID, value string) {
		if team := teamByID[teamID]; team != nil {
			team.Sponsors = append(team.Sponsors, value)
		}
	}); err != nil {
		return err
	}
	if err := loadTeamStringRelations(`SELECT team_id, driver_name FROM team_drivers ORDER BY team_id, driver_name`, func(teamID, value string) {
		if team := teamByID[teamID]; team != nil {
			team.Drivers = append(team.Drivers, value)
		}
	}); err != nil {
		return err
	}
	if err := loadTeamStringRelations(`SELECT team_id, user_id FROM team_followers ORDER BY team_id, user_id`, func(teamID, value string) {
		if team := teamByID[teamID]; team != nil {
			team.Followers = append(team.Followers, value)
		}
	}); err != nil {
		return err
	}

	return nil
}

func loadTeamStringRelations(query string, assign func(teamID, value string)) error {
	rows, err := DB.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var teamID string
		var value string
		if err := rows.Scan(&teamID, &value); err != nil {
			return err
		}
		assign(teamID, value)
	}

	return rows.Err()
}

func loadCategoriesFromPostgres() error {
	rows, err := DB.Query(`SELECT id, name, slug, icon, live_count, description, color FROM categories ORDER BY name`)
	if err != nil {
		return err
	}
	defer rows.Close()

	Categories = make([]models.Category, 0)
	for rows.Next() {
		var category models.Category
		if err := rows.Scan(&category.ID, &category.Name, &category.Slug, &category.Icon, &category.LiveCount, &category.Description, &category.Color); err != nil {
			return err
		}
		Categories = append(Categories, category)
	}

	return rows.Err()
}

func loadEventsFromPostgres() error {
	rows, err := DB.Query(`SELECT id, title, location, date, COALESCE(time, ''), is_live, category, COALESCE(thumbnail_url, ''), created_at FROM events ORDER BY date`)
	if err != nil {
		return err
	}
	defer rows.Close()

	Events = make([]models.Event, 0)
	for rows.Next() {
		var event models.Event
		if err := rows.Scan(&event.ID, &event.Title, &event.Location, &event.Date, &event.Time, &event.IsLive, &event.Category, &event.ThumbnailURL, &event.CreatedAt); err != nil {
			return err
		}
		Events = append(Events, event)
	}

	return rows.Err()
}

func loadJourneysFromPostgres() error {
	rows, err := DB.Query(`SELECT id, title, category, description, COALESCE(badge, ''), slots_left, date, price::float8, COALESCE(thumbnail_url, '') FROM journeys ORDER BY date`)
	if err != nil {
		return err
	}
	defer rows.Close()

	Journeys = make([]models.Journey, 0)
	for rows.Next() {
		var journey models.Journey
		if err := rows.Scan(&journey.ID, &journey.Title, &journey.Category, &journey.Description, &journey.Badge, &journey.SlotsLeft, &journey.Date, &journey.Price, &journey.ThumbnailURL); err != nil {
			return err
		}
		Journeys = append(Journeys, journey)
	}

	return rows.Err()
}

func loadMerchFromPostgres() error {
	rows, err := DB.Query(`SELECT id, name, icon, price, category FROM merch_items ORDER BY name`)
	if err != nil {
		return err
	}
	defer rows.Close()

	MerchItems = make([]models.MerchItem, 0)
	for rows.Next() {
		var item models.MerchItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Icon, &item.Price, &item.Category); err != nil {
			return err
		}
		MerchItems = append(MerchItems, item)
	}

	return rows.Err()
}

func loadPassesFromPostgres() error {
	rows, err := DB.Query(`SELECT id, tier, event_name, location, event_date, category, price::float8, perks, spots_left, total_spots, badge, tier_color FROM passes ORDER BY event_date`)
	if err != nil {
		return err
	}
	defer rows.Close()

	Passes = make([]Pass, 0)
	for rows.Next() {
		var pass Pass
		var perksJSON []byte
		var badge sql.NullString
		var eventDate time.Time
		if err := rows.Scan(&pass.ID, &pass.Tier, &pass.Event, &pass.Location, &eventDate, &pass.Category, &pass.Price, &perksJSON, &pass.SpotsLeft, &pass.TotalSpots, &badge, &pass.TierColor); err != nil {
			return err
		}
		pass.Date = eventDate.UTC().Format(time.RFC3339)
		if len(perksJSON) > 0 {
			if err := json.Unmarshal(perksJSON, &pass.Perks); err != nil {
				return err
			}
		}
		if badge.Valid {
			value := badge.String
			pass.Badge = &value
		}
		Passes = append(Passes, pass)
	}

	return rows.Err()
}

func seedInMemory() {
	seedUsers()
	seedEvents()
	seedCategories()
	seedJourneys()
	seedMerch()
	seedPasses()
	seedTeams()
	seedReminders()
	seedOrders()
}

func seedUsers() {
	pw := "password123"
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("failed to seed users password hash: %v", err)
		Users = []models.User{}
		return
	}

	Users = []models.User{
		{
			ID:           uuid.New().String(),
			Email:        "viewer@example.com",
			PasswordHash: string(passwordHash),
			Role:         "viewer",
			Name:         "Viewer User",
			Plan:         "FREE",
			CreatedAt:    time.Now(),
		},
		{
			ID:           uuid.New().String(),
			Email:        "participant@example.com",
			PasswordHash: string(passwordHash),
			Role:         "participant",
			Name:         "Participant User",
			Plan:         "VIP",
			CreatedAt:    time.Now(),
		},
		{
			ID:           uuid.New().String(),
			Email:        "manager@example.com",
			PasswordHash: string(passwordHash),
			Role:         "manager",
			Name:         "Manager User",
			Plan:         "VIP",
			CreatedAt:    time.Now(),
		},
		{
			ID:           uuid.New().String(),
			Email:        "sponsor@example.com",
			PasswordHash: string(passwordHash),
			Role:         "sponsor",
			Name:         "Sponsor User",
			Plan:         "PLATINUM",
			CreatedAt:    time.Now(),
		},
		{
			ID:           uuid.New().String(),
			Email:        "admin@example.com",
			PasswordHash: string(passwordHash),
			Role:         "admin",
			Name:         "Admin User",
			Plan:         "PLATINUM",
			CreatedAt:    time.Now(),
		},
	}
}

func seedReminders() {
	Reminders = []models.Reminder{}
}

func seedOrders() {
	Orders = []models.Order{}
}

func seedTeams() {
	Teams = []models.Team{}
}

func CloseDB() {
	Mu.Lock()
	defer Mu.Unlock()

	if DB != nil {
		if err := DB.Close(); err != nil {
			log.Printf("failed closing database connection: %v", err)
		}
		DB = nil
	}

	log.Println("Closing database connection...")
}

func seedEvents() {
	Events = []models.Event{
		{
			ID:       uuid.New().String(),
			Title:    "Speed Boat Cup - Finals",
			Location: "Lake Como - Italy",
			Date:     time.Now().Add(14 * 24 * time.Hour),
			IsLive:   false,
			Category: "water",
			Time:     "14:00 UTC",
		},
		{
			ID:       uuid.New().String(),
			Title:    "Red Bull Skydive Series - Rd. 3",
			Location: "Interlaken Drop Zone - Switzerland",
			Date:     time.Now().Add(20 * 24 * time.Hour),
			IsLive:   false,
			Category: "air",
			Time:     "11:30 UTC",
		},
		{
			ID:       uuid.New().String(),
			Title:    "Crop Duster Air Racing",
			Location: "Bakersfield Airfield - California",
			Date:     time.Now().Add(26 * 24 * time.Hour),
			IsLive:   false,
			Category: "air",
			Time:     "16:00 UTC",
		},
	}
}

func seedCategories() {
	Categories = []models.Category{
		{
			ID:          uuid.New().String(),
			Name:        "MOTORSPORT",
			Slug:        "motorsport",
			Icon:        "R",
			LiveCount:   24,
			Description: "NASCAR - F1 - Dirt - Rally",
			Color:       "cyan",
		},
		{
			ID:          uuid.New().String(),
			Name:        "WATER",
			Slug:        "water",
			Icon:        "W",
			LiveCount:   8,
			Description: "Speed Boats - Jet Ski - Surf",
			Color:       "blue",
		},
		{
			ID:          uuid.New().String(),
			Name:        "AIR & SKY",
			Slug:        "air",
			Icon:        "A",
			LiveCount:   5,
			Description: "Skydive - Air Race - Wing",
			Color:       "purple",
		},
		{
			ID:          uuid.New().String(),
			Name:        "OFF-ROAD",
			Slug:        "offroad",
			Icon:        "O",
			LiveCount:   12,
			Description: "Dakar - Baja - Enduro",
			Color:       "orange",
		},
	}
}

func seedJourneys() {
	Journeys = []models.Journey{
		{
			ID:          uuid.New().String(),
			Title:       "DAYTONA PIT CREW EXPERIENCE",
			Category:    "MOTORSPORT - BEHIND THE SCENES",
			Description: "Go behind the wall at Daytona 500. Watch pit stops up close, meet the crew chiefs, and ride the pace car on track.",
			Badge:       "EXCLUSIVE",
			SlotsLeft:   12,
			Date:        time.Now().Add(10 * 24 * time.Hour),
			Price:       2400,
		},
		{
			ID:          uuid.New().String(),
			Title:       "DAKAR DESERT CONVOY",
			Category:    "RALLY - DESERT EXPEDITION",
			Description: "Ride a support vehicle through the Dakar stages. Sleep under the stars, eat with the team, and feel the dust.",
			Badge:       "MEMBERS ONLY",
			SlotsLeft:   6,
			Date:        time.Now().Add(345 * 24 * time.Hour),
			Price:       5800,
		},
		{
			ID:          uuid.New().String(),
			Title:       "RED BULL TANDEM SKYDIVE",
			Category:    "AIR - EXTREME SPORT",
			Description: "Jump with a Red Bull certified instructor at 15,000ft. Camera-equipped, full debrief, and a story you'll never forget.",
			Badge:       "LIMITED",
			SlotsLeft:   3,
			Date:        time.Now().Add(20 * 24 * time.Hour),
			Price:       1200,
		},
	}
}

func seedMerch() {
	MerchItems = []models.MerchItem{
		{ID: "merch-team-hoodie", Name: "Team Hoodie", Icon: "H", Price: 89, Category: "apparel"},
		{ID: "merch-nitrous-cap", Name: "NITROUS Cap", Icon: "C", Price: 42, Category: "apparel"},
		{ID: "merch-racing-jacket", Name: "Racing Jacket", Icon: "J", Price: 189, Category: "apparel"},
		{ID: "merch-pit-watch", Name: "Pit Watch", Icon: "W", Price: 249, Category: "accessories"},
		{ID: "merch-gear-backpack", Name: "Gear Backpack", Icon: "B", Price: 120, Category: "accessories"},
		{ID: "merch-drop-keychain", Name: "Drop Keychain", Icon: "K", Price: 28, Category: "collectibles"},
	}
}

func seedPasses() {
	badge := "LIMITED"
	Passes = []Pass{
		{ID: "pass-daytona-grandstand", Tier: "GRANDSTAND", Event: "Daytona 500", Location: "Daytona Beach, FL", Date: time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339), Category: "motorsport", Price: 299, Perks: []string{"Track access", "Pit lane tour"}, SpotsLeft: 4, TotalSpots: 20, Badge: &badge, TierColor: "#ff4d4d"},
		{ID: "pass-pit-experience", Tier: "PIT ACCESS", Event: "F1 Grand Prix", Location: "Austin, TX", Date: time.Now().Add(45 * 24 * time.Hour).Format(time.RFC3339), Category: "motorsport", Price: 599, Perks: []string{"Pit walk", "Garage access"}, SpotsLeft: 12, TotalSpots: 50, Badge: nil, TierColor: "#60a5fa"},
	}
	PassPurchases = []PassPurchase{}
}
