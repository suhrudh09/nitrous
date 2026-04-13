package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Stream struct {
	ID            string                `json:"id"`
	EventID       string                `json:"eventId"`
	Title         string                `json:"title"`
	Subtitle      string                `json:"subtitle"`
	PlaybackURL   string                `json:"playbackUrl,omitempty"`
	ExternalWatch []ExternalWatchOption `json:"externalWatch,omitempty"`
	DateStart     string                `json:"date_start,omitempty"`
	DateEnd       string                `json:"date_end,omitempty"`
	CountryName   string                `json:"country_name,omitempty"`
	CircuitShort  string                `json:"circuit_short_name,omitempty"`
	Category      string                `json:"category"`
	Location      string                `json:"location"`
	Quality       string                `json:"quality"`
	Viewers       int                   `json:"viewers"`
	IsLive        bool                  `json:"isLive"`
	CurrentLeader string                `json:"currentLeader"`
	CurrentSpeed  string                `json:"currentSpeed"`
	Color         string                `json:"color"`
	CreatedAt     string                `json:"createdAt"`
}

type ExternalWatchOption struct {
	Platform string `json:"platform"`
	Label    string `json:"label"`
	URL      string `json:"url"`
}

type TelemetryPayload struct {
	Type      string    `json:"type"`
	StreamID  string    `json:"streamId"`
	SpeedKPH  int       `json:"speedKph"`
	RPM       int       `json:"rpm"`
	Gear      int       `json:"gear"`
	GForce    float64   `json:"gForce"`
	Timestamp time.Time `json:"timestamp"`
}

type Hub struct {
	clients    map[*websocket.Conn]bool
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	broadcast  chan []byte
}

var (
	streamsMu sync.RWMutex
	streams   = []Stream{}

	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	hubOnce   sync.Once
	streamHub = &Hub{
		clients:    make(map[*websocket.Conn]bool),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		broadcast:  make(chan []byte, 128),
	}
)

func ensureHubRunning() {
	hubOnce.Do(func() {
		go RunHub()
	})
}

// GetStreams returns all stream feeds
func GetStreams(c *gin.Context) {
	streamsMu.RLock()
	defer streamsMu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"streams": streams,
		"count":   len(streams),
	})
}

// GetStreamByID returns one stream feed
func GetStreamByID(c *gin.Context) {
	id := c.Param("id")

	streamsMu.RLock()
	defer streamsMu.RUnlock()

	for _, stream := range streams {
		if stream.ID == id {
			c.JSON(http.StatusOK, stream)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Stream not found"})
}

// CreateStream creates a new stream feed (admin only)
func CreateStream(c *gin.Context) {
	var stream Stream

	if err := c.ShouldBindJSON(&stream); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	streamsMu.Lock()
	defer streamsMu.Unlock()

	streams = append(streams, stream)
	c.JSON(http.StatusCreated, stream)
}

// UpdateStream updates an existing stream feed (admin only)
func UpdateStream(c *gin.Context) {
	id := c.Param("id")

	var updated Stream
	if err := c.ShouldBindJSON(&updated); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	streamsMu.Lock()
	defer streamsMu.Unlock()

	for i, stream := range streams {
		if stream.ID == id {
			updated.ID = id
			streams[i] = updated
			c.JSON(http.StatusOK, updated)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Stream not found"})
}

// DeleteStream deletes a stream feed (admin only)
func DeleteStream(c *gin.Context) {
	id := c.Param("id")

	streamsMu.Lock()
	defer streamsMu.Unlock()

	for i, stream := range streams {
		if stream.ID == id {
			streams = append(streams[:i], streams[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"message": "Stream deleted"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Stream not found"})
}

// StreamsWS upgrades the request to websocket and registers the client to telemetry updates
func StreamsWS(c *gin.Context) {
	ensureHubRunning()

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to establish websocket connection"})
		return
	}

	streamHub.register <- conn

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			streamHub.unregister <- conn
			break
		}
	}
}

// RunHub runs the websocket client registration, unregistration, and broadcast loop
func RunHub() {
	for {
		select {
		case conn := <-streamHub.register:
			streamHub.clients[conn] = true

		case conn := <-streamHub.unregister:
			if _, ok := streamHub.clients[conn]; ok {
				delete(streamHub.clients, conn)
				_ = conn.Close()
			}

		case message := <-streamHub.broadcast:
			for conn := range streamHub.clients {
				if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
					delete(streamHub.clients, conn)
					_ = conn.Close()
				}
			}
		}
	}
}

// BroadcastTelemetry publishes telemetry updates to all connected websocket clients
func BroadcastTelemetry(streamID string, speedKPH int, rpm int, gear int, gForce float64) {
	ensureHubRunning()

	payload := TelemetryPayload{
		Type:      "telemetry",
		StreamID:  streamID,
		SpeedKPH:  speedKPH,
		RPM:       rpm,
		Gear:      gear,
		GForce:    gForce,
		Timestamp: time.Now().UTC(),
	}

	message, err := json.Marshal(payload)
	if err != nil {
		log.Printf("failed to marshal telemetry payload: %v", err)
		return
	}

	streamHub.broadcast <- message
}
