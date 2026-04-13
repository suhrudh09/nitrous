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
	config.LoadConfig()

	database.InitDB()
	defer database.CloseDB()

	go handlers.RunHub()
	go handlers.SimulateTelemetry() // swap for handlers.PollOpenF1() once live data is wired

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://nitrous.vercel.app"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "message": "Nitrous API is running"})
	})

	r.GET("/ws/streams", handlers.StreamsWS)

	api := r.Group("/api")
	{
		events := api.Group("/events")
		{
			events.GET("", handlers.GetEvents)
			events.GET("/live", handlers.GetLiveEvents)
			events.GET("/:id", handlers.GetEventByID)
			events.POST("/:id/remind", middleware.AuthMiddleware(), handlers.SetReminder)
			events.DELETE("/:id/remind", middleware.AuthMiddleware(), handlers.DeleteReminder)
			events.POST("", middleware.AuthMiddleware(), handlers.CreateEvent)
			events.PUT("/:id", middleware.AuthMiddleware(), handlers.UpdateEvent)
			events.DELETE("/:id", middleware.AuthMiddleware(), handlers.DeleteEvent)
		}

		streams := api.Group("/streams")
		{
			streams.GET("", handlers.GetStreams)
			streams.GET("/:id", handlers.GetStreamByID)
		}

		categories := api.Group("/categories")
		{
			categories.GET("", handlers.GetCategories)
			categories.GET("/:slug", handlers.GetCategoryBySlug)
		}

		journeys := api.Group("/journeys")
		{
			journeys.GET("", handlers.GetJourneys)
			journeys.GET("/:id", handlers.GetJourneyByID)
			journeys.POST("/:id/book", middleware.AuthMiddleware(), handlers.BookJourney)
		}

		merch := api.Group("/merch")
		{
			merch.GET("", handlers.GetMerchItems)
			merch.GET("/:id", handlers.GetMerchItemByID)
		}

		orders := api.Group("/orders")
		orders.Use(middleware.AuthMiddleware())
		{
			orders.POST("", handlers.CreateOrder)
			orders.GET("", handlers.GetMyOrders)
			orders.GET("/:id", handlers.GetOrderByID)
		}

		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
			auth.GET("/me", middleware.AuthMiddleware(), handlers.GetCurrentUser)
			auth.GET("/reminders", middleware.AuthMiddleware(), handlers.GetMyReminders)
		}

		teams := api.Group("/teams")
		{
			teams.GET("", handlers.GetTeams)
			teams.GET("/:id", handlers.GetTeamByID)
			teams.POST("/:id/follow", middleware.AuthMiddleware(), handlers.FollowTeam)
			teams.DELETE("/:id/follow", middleware.AuthMiddleware(), handlers.UnfollowTeam)
		}
	}

	log.Println("🚀 Nitrous API server starting on :" + config.AppConfig.Port)
	r.Run(":" + config.AppConfig.Port)
}
