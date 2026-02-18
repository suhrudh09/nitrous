package main

import (
	"log"
	"nitrous-backend/config"
	"nitrous-backend/database"
	"nitrous-backend/handlers"
	"nitrous-backend/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load environment variables
	config.LoadConfig()

	// Initialize database
	database.InitDB()
	defer database.CloseDB()

	// Create Gin router
	r := gin.Default()

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://nitrous.vercel.app"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "message": "Nitrous API is running"})
	})

	// API routes
	api := r.Group("/api")
	{
		// Events
		events := api.Group("/events")
		{
			events.GET("", handlers.GetEvents)
			events.GET("/live", handlers.GetLiveEvents)
			events.GET("/:id", handlers.GetEventByID)
			events.POST("", middleware.AuthMiddleware(), handlers.CreateEvent)
			events.PUT("/:id", middleware.AuthMiddleware(), handlers.UpdateEvent)
			events.DELETE("/:id", middleware.AuthMiddleware(), handlers.DeleteEvent)
		}

		// Categories
		categories := api.Group("/categories")
		{
			categories.GET("", handlers.GetCategories)
			categories.GET("/:slug", handlers.GetCategoryBySlug)
		}

		// Journeys
		journeys := api.Group("/journeys")
		{
			journeys.GET("", handlers.GetJourneys)
			journeys.GET("/:id", handlers.GetJourneyByID)
			journeys.POST("/:id/book", middleware.AuthMiddleware(), handlers.BookJourney)
		}

		// Merch
		merch := api.Group("/merch")
		{
			merch.GET("", handlers.GetMerchItems)
			merch.GET("/:id", handlers.GetMerchItemByID)
		}

		// Teams
		teams := api.Group("/teams")
		{
			teams.GET("", handlers.GetTeams)
			teams.GET("/:id", handlers.GetTeamByID)
			teams.POST("/:id/follow", middleware.AuthMiddleware(), handlers.FollowTeam)
			teams.POST("/:id/unfollow", middleware.AuthMiddleware(), handlers.UnfollowTeam)
		}

		// Streams
		streams := api.Group("/streams")
		{
			streams.GET("", handlers.GetStreams)
			streams.GET("/:id", handlers.GetStreamByID)
			streams.GET("/ws", handlers.StreamsWS)
		}

		// Reminders
		reminders := api.Group("/reminders")
		{
			reminders.GET("", middleware.AuthMiddleware(), handlers.GetMyReminders)
			reminders.POST("", middleware.AuthMiddleware(), handlers.SetReminder)
			reminders.DELETE("/:id", middleware.AuthMiddleware(), handlers.DeleteReminder)
		}

		// Orders
		orders := api.Group("/orders")
		{
			orders.GET("", middleware.AuthMiddleware(), handlers.GetMyOrders)
			orders.POST("", middleware.AuthMiddleware(), handlers.CreateOrder)
			orders.GET("/:id", middleware.AuthMiddleware(), handlers.GetOrderByID)
		}

		// Auth
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
			auth.GET("/me", middleware.AuthMiddleware(), handlers.GetCurrentUser)
		}
	}

	log.Println("🚀 Nitrous API server starting on :8080")
	r.Run(":8080")
}
