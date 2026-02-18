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
	ID          string `json:"id"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	IsLive      bool   `json:"isLive"`
	ViewerCount int    `json:"viewerCount"`
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
	streams = []Stream{
		{ID: "stream-1", Title: "Daytona 500 — Main Feed", Category: "motorsport", IsLive: true, ViewerCount: 12042},
		{ID: "stream-2", Title: "Dakar Rally — Stage Cam", Category: "offroad", IsLive: true, ViewerCount: 5421},
		{ID: "stream-3", Title: "Sky Racing Cockpit View", Category: "air", IsLive: false, ViewerCount: 0},
	}

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

// GetStreams returns all stream feeds.
func GetStreams(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"streams": streams,
		"count":   len(streams),
	})
}

// GetStreamByID returns one stream feed.
func GetStreamByID(c *gin.Context) {
	id := c.Param("id")

	for _, stream := range streams {
		if stream.ID == id {
			c.JSON(http.StatusOK, stream)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Stream not found"})
}

// StreamsWS upgrades the request to websocket and registers the client to telemetry updates.
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

// RunHub runs the websocket client registration, unregistration, and broadcast loop.
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

// BroadcastTelemetry publishes telemetry updates to all connected websocket clients.
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

	select {
	case streamHub.broadcast <- message:
	default:
		log.Printf("telemetry message dropped: broadcast channel is full")
	}
}
