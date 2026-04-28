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
	IsPrivate      bool      `json:"isPrivate,omitempty"`
	Managers       []string  `json:"managers,omitempty"`
	Sponsors       []string  `json:"sponsors,omitempty"`
	Members        []string  `json:"members,omitempty"`
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
	ID           string    `json:"id"`
	UserID       string    `json:"userId"`
	MerchItemIDs []string  `json:"merchItemIds"`
	Quantities   []int     `json:"quantities"`
	UnitPrices   []float64 `json:"unitPrices"`
	Total        float64   `json:"total"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
}

// User represents a platform user
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never send password hash to frontend
	Role         string    `json:"role"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"createdAt"`
}

// GarageConfig represents a user's saved vehicle configuration
type GarageConfig struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Name      string    `json:"name"`
	Make      string    `json:"make"`
	Model     string    `json:"model"`
	Year      int       `json:"year"`
	Engine    string    `json:"engine"`
	Tuning    string    `json:"tuning"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Payment represents a payment transaction
type Payment struct {
	ID                    string                 `json:"id"`
	UserID                string                 `json:"userId"`
	Amount                float64                `json:"amount"`
	Currency              string                 `json:"currency"`
	Status                string                 `json:"status"`
	PaymentMethod         string                 `json:"paymentMethod"`
	StripePaymentIntentID string                 `json:"stripePaymentIntentId,omitempty"`
	StripeCustomerID      string                 `json:"stripeCustomerId,omitempty"`
	Description           string                 `json:"description"`
	ReferenceType         string                 `json:"referenceType"`
	ReferenceID           string                 `json:"referenceId"`
	Metadata              map[string]interface{} `json:"metadata"`
	CreatedAt             time.Time              `json:"createdAt"`
	UpdatedAt             time.Time              `json:"updatedAt"`
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
	Role     string `json:"role"`
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
	MerchItemIDs []string  `json:"merchItemIds" binding:"required,min=1,dive,required"`
	Quantities   []int     `json:"quantities" binding:"required,min=1,dive,min=1"`
	UnitPrices   []float64 `json:"unitPrices" binding:"required,min=1,dive,gt=0"`
}
