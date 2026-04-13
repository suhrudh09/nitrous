package models

import "time"

// ── Event ─────────────────────────────────────────────────────────────────────

type Event struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Location     string    `json:"location"`
	Date         time.Time `json:"date"`
	Time         string    `json:"time,omitempty"`
	IsLive       bool      `json:"isLive"`
	Category     string    `json:"category"`
	ThumbnailURL string    `json:"thumbnailUrl,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

// ── Category ──────────────────────────────────────────────────────────────────

type Category struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Icon        string `json:"icon"`
	LiveCount   int    `json:"liveCount"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

// ── Journey ───────────────────────────────────────────────────────────────────

type Journey struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Category     string    `json:"category"`
	Description  string    `json:"description"`
	Badge        string    `json:"badge"`
	SlotsLeft    int       `json:"slotsLeft"`
	Date         time.Time `json:"date"`
	Price        float64   `json:"price"`
	ThumbnailURL string    `json:"thumbnailUrl,omitempty"`
}

// ── MerchItem ─────────────────────────────────────────────────────────────────

type MerchItem struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Icon     string  `json:"icon"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
}

// ── User ──────────────────────────────────────────────────────────────────────

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // never serialized to frontend
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"createdAt"`
}

// ── Auth requests ─────────────────────────────────────────────────────────────

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name"     binding:"required"`
}

type BookingRequest struct {
	JourneyID string `json:"journeyId" binding:"required"`
	UserID    string `json:"userId"    binding:"required"`
}

// ── Team ──────────────────────────────────────────────────────────────────────

type Team struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Category  string    `json:"category"` // e.g. "MOTORSPORT · F1"
	Country   string    `json:"country"`
	Founded   int       `json:"founded"`
	Rank      int       `json:"rank"`
	Wins      int       `json:"wins"`
	Points    int       `json:"points"`
	Following int       `json:"following"`
	Drivers   []string  `json:"drivers"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"createdAt"`
}

type TeamFollow struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	TeamID    string    `json:"teamId"`
	CreatedAt time.Time `json:"createdAt"`
}

// ── Stream ────────────────────────────────────────────────────────────────────

type Stream struct {
	ID            string    `json:"id"`
	EventID       string    `json:"eventId"`
	Title         string    `json:"title"`
	Subtitle      string    `json:"subtitle"` // e.g. "Lap 87 / 200"
	Category      string    `json:"category"`
	Location      string    `json:"location"`
	Quality       string    `json:"quality"` // "4K" | "HD"
	Viewers       int       `json:"viewers"`
	IsLive        bool      `json:"isLive"`
	CurrentLeader string    `json:"currentLeader"`
	CurrentSpeed  string    `json:"currentSpeed"`
	Color         string    `json:"color"`
	CreatedAt     time.Time `json:"createdAt"`
}

// StreamTelemetry is broadcast over WebSocket to update live data
type StreamTelemetry struct {
	StreamID      string `json:"streamId"`
	Viewers       int    `json:"viewers"`
	CurrentLeader string `json:"currentLeader"`
	CurrentSpeed  string `json:"currentSpeed"`
	Subtitle      string `json:"subtitle"`
}

// ── Reminder ──────────────────────────────────────────────────────────────────

type Reminder struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	EventID   string    `json:"eventId"`
	CreatedAt time.Time `json:"createdAt"`
}

// ── Order ─────────────────────────────────────────────────────────────────────

type OrderItem struct {
	MerchID  string  `json:"merchId"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
	Size     string  `json:"size,omitempty"`
}

type Order struct {
	ID        string      `json:"id"`
	UserID    string      `json:"userId"`
	Items     []OrderItem `json:"items"`
	Total     float64     `json:"total"`
	Status    string      `json:"status"` // "pending" | "confirmed" | "shipped"
	CreatedAt time.Time   `json:"createdAt"`
}

type CreateOrderRequest struct {
	Items []OrderItem `json:"items" binding:"required,min=1"`
}
