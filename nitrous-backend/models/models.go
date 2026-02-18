package models

import "time"

// Event represents a racing event
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

// Category represents an event category
type Category struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Icon        string `json:"icon"`
	LiveCount   int    `json:"liveCount"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

// Journey represents an exclusive experience
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

// MerchItem represents a merchandise item
type MerchItem struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Icon     string  `json:"icon"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
}

// Team represents a racing team or organization
type Team struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Country        string    `json:"country,omitempty"`
	Drivers        []string  `json:"drivers,omitempty"`
	Followers      []string  `json:"followers,omitempty"`
	FollowersCount int       `json:"followersCount"`
	CreatedAt      time.Time `json:"createdAt"`
}

// Reminder represents a user's reminder for an event.
type Reminder struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	EventID   string    `json:"eventId"`
	Message   string    `json:"message,omitempty"`
	RemindAt  time.Time `json:"remindAt"`
	CreatedAt time.Time `json:"createdAt"`
}

// Order represents a merch order placed by a user.
type Order struct {
	ID          string    `json:"id"`
	UserID      string    `json:"userId"`
	MerchItemID string    `json:"merchItemId"`
	Quantity    int       `json:"quantity"`
	UnitPrice   float64   `json:"unitPrice"`
	TotalPrice  float64   `json:"totalPrice"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

// User represents a platform user
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never send password hash to frontend
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"createdAt"`
}

// LoginRequest for authentication
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest for new users
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
}

// BookingRequest for journey bookings
type BookingRequest struct {
	JourneyID string `json:"journeyId" binding:"required"`
	UserID    string `json:"userId" binding:"required"`
}

// SetReminderRequest for creating a reminder.
type SetReminderRequest struct {
	EventID  string    `json:"eventId" binding:"required"`
	Message  string    `json:"message"`
	RemindAt time.Time `json:"remindAt" binding:"required"`
}

// CreateOrderRequest for creating merch orders.
type CreateOrderRequest struct {
	MerchItemID string `json:"merchItemId" binding:"required"`
	Quantity    int    `json:"quantity" binding:"required,min=1"`
}
