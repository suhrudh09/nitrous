package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"nitrous-backend/database"
	"nitrous-backend/models"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// ── REST ──────────────────────────────────────────────────────────────────────

func GetStreams(c *gin.Context) {
	all := database.GetStreams()
	var live []models.Stream
	for _, s := range all {
		if s.IsLive {
			live = append(live, s)
		}
	}
	c.JSON(http.StatusOK, gin.H{"streams": live, "count": len(live)})
}

func GetStreamByID(c *gin.Context) {
	s, found := database.FindStreamByID(c.Param("id"))
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stream not found"})
		return
	}
	c.JSON(http.StatusOK, s)
}

// ── WebSocket hub ─────────────────────────────────────────────────────────────

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "http://localhost:3000" || origin == "https://nitrous.vercel.app"
	},
}

type Hub struct {
	clients   map[*websocket.Conn]bool
	broadcast chan []byte
	mu        sync.Mutex
}

var streamHub = &Hub{
	clients:   make(map[*websocket.Conn]bool),
	broadcast: make(chan []byte, 256),
}

func RunHub() {
	for msg := range streamHub.broadcast {
		streamHub.mu.Lock()
		for conn := range streamHub.clients {
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				conn.Close()
				delete(streamHub.clients, conn)
			}
		}
		streamHub.mu.Unlock()
	}
}

func StreamsWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	streamHub.mu.Lock()
	streamHub.clients[conn] = true
	streamHub.mu.Unlock()

	// Send full stream snapshot on connect
	snapshot, _ := json.Marshal(database.GetStreams())
	conn.WriteMessage(websocket.TextMessage, snapshot)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			streamHub.mu.Lock()
			delete(streamHub.clients, conn)
			streamHub.mu.Unlock()
			break
		}
	}
}

// BroadcastTelemetry updates the store and fans out to all WS clients.
func BroadcastTelemetry(t models.StreamTelemetry) {
	database.UpdateStreamTelemetry(t)

	payload, err := json.Marshal(t)
	if err != nil {
		return
	}
	streamHub.broadcast <- payload
}

// SimulateTelemetry is a dev fallback — replaced by PollOpenF1 when live.
func SimulateTelemetry() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	lap := 87
	for range ticker.C {
		s, ok := database.FirstStream()
		if !ok {
			continue
		}
		lap++
		BroadcastTelemetry(models.StreamTelemetry{
			StreamID:      s.ID,
			Viewers:       s.Viewers + 100,
			CurrentLeader: s.CurrentLeader,
			CurrentSpeed:  s.CurrentSpeed,
			Subtitle:      "Lap " + strconv.Itoa(lap) + " / 200",
		})
	}
}
