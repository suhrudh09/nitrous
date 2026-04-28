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

	// Start external provider integrations (Jolpica, OpenF1, TheSportsDB)
	handlers.StartExternalDataIntegration()

	// Create Gin router
	r := gin.Default()

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://nitrous-weld.vercel.app"},
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
		// Garage
		garage := api.Group("/garage")
		{
			garage.GET("/makes", handlers.GetGarageMakes)
			garage.GET("/models", handlers.GetGarageModels)
			garage.GET("/years", handlers.GetGarageYears)
			garage.GET("/trims", handlers.GetGarageTrims)
			garage.GET("/vehicle", handlers.GetGarageVehicle)
			garage.GET("/tuning-configs", handlers.GetGarageTuningConfigs)
			garage.POST("/tune", middleware.AuthMiddleware(), middleware.RequireRoles("admin", "manager"), handlers.PostGarageTune)
			garage.GET("/search", handlers.GetGarageSearch)
		}

		// Events
		events := api.Group("/events")
		{
			events.GET("", handlers.GetEvents)
			events.GET("/live", handlers.GetLiveEvents)
			events.GET("/:id", handlers.GetEventByID)
			events.POST("", middleware.AuthMiddleware(), middleware.AdminMiddleware(), handlers.CreateEvent)
			events.PUT("/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), handlers.UpdateEvent)
			events.DELETE("/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), handlers.DeleteEvent)
		}

		// Categories
		categories := api.Group("/categories")
		{
			categories.GET("", handlers.GetCategories)
			categories.GET("/:slug", handlers.GetCategoryBySlug)
			categories.POST("", middleware.AuthMiddleware(), middleware.AdminMiddleware(), handlers.CreateCategory)
			categories.PUT("/:slug", middleware.AuthMiddleware(), middleware.AdminMiddleware(), handlers.UpdateCategory)
			categories.DELETE("/:slug", middleware.AuthMiddleware(), middleware.AdminMiddleware(), handlers.DeleteCategory)
		}

		// Journeys
		journeys := api.Group("/journeys")
		{
			journeys.GET("", handlers.GetJourneys)
			journeys.GET("/:id", handlers.GetJourneyByID)
			journeys.POST("", middleware.AuthMiddleware(), middleware.AdminMiddleware(), handlers.CreateJourney)
			journeys.PUT("/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), handlers.UpdateJourney)
			journeys.DELETE("/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), handlers.DeleteJourney)
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
			teams.POST("", middleware.AuthMiddleware(), middleware.RequireRoles("admin", "manager"), handlers.CreateTeam)
			teams.PUT("/:id", middleware.AuthMiddleware(), middleware.RequireRoles("admin", "manager"), handlers.UpdateTeam)
			teams.DELETE("/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), handlers.DeleteTeam)
			teams.POST("/:id/managers", middleware.AuthMiddleware(), middleware.RequireRoles("admin", "manager"), handlers.AddTeamManager)
			teams.DELETE("/:id/managers/:userId", middleware.AuthMiddleware(), middleware.RequireRoles("admin", "manager"), handlers.RemoveTeamManager)
			teams.GET("/:id/members", handlers.ListTeamMembers)
			teams.POST("/:id/members", middleware.AuthMiddleware(), middleware.RequireRoles("admin", "manager"), handlers.AddTeamMember)
			teams.DELETE("/:id/members/:userId", middleware.AuthMiddleware(), middleware.RequireRoles("admin", "manager"), handlers.RemoveTeamMember)
			teams.GET("/:id/sponsors", handlers.ListTeamSponsors)
			teams.POST("/:id/sponsors", middleware.AuthMiddleware(), middleware.RequireRoles("admin", "manager"), handlers.AddTeamSponsor)
			teams.DELETE("/:id/sponsors/:userId", middleware.AuthMiddleware(), middleware.RequireRoles("admin", "manager"), handlers.RemoveTeamSponsor)
			teams.POST("/:id/follow", middleware.AuthMiddleware(), handlers.FollowTeam)
			teams.POST("/:id/unfollow", middleware.AuthMiddleware(), handlers.UnfollowTeam)
		}

		// Admin-only routes for syncs and role management
		admin := api.Group("/admin")
		{
			admin.POST("/sync", middleware.AuthMiddleware(), middleware.AdminMiddleware(), handlers.AdminTriggerSync)
			admin.POST("/users/:id/role", middleware.AuthMiddleware(), middleware.AdminMiddleware(), handlers.AdminSetUserRole)
		}

		// Streams
		streams := api.Group("/streams")
		{
			streams.GET("", handlers.GetStreams)
			streams.GET("/openf1/sessions", handlers.GetOpenF1RecentSessions)
			streams.GET("/openf1/sessions/:sessionKey/telemetry", handlers.GetOpenF1SessionTelemetry)
			streams.GET("/:id", handlers.GetStreamByID)
			streams.POST("", middleware.AuthMiddleware(), middleware.AdminMiddleware(), handlers.CreateStream)
			streams.PUT("/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), handlers.UpdateStream)
			streams.DELETE("/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), handlers.DeleteStream)
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
			orders.DELETE("/:id", middleware.AuthMiddleware(), handlers.CancelOrder)
		}

		// Passes
		passes := api.Group("/passes")
		{
			passes.POST("/:id/purchase", middleware.AuthMiddleware(), handlers.PurchasePass)
		}

		// Auth
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
			auth.GET("/me", middleware.AuthMiddleware(), handlers.GetCurrentUser)
		}

		// Garage Configs
		garageConfigs := api.Group("/garage")
		{
			garageConfigs.GET("/configs", middleware.AuthMiddleware(), handlers.GetGarageConfigs)
			garageConfigs.POST("/configs", middleware.AuthMiddleware(), handlers.SaveGarageConfig)
			garageConfigs.DELETE("/configs/:id", middleware.AuthMiddleware(), handlers.DeleteGarageConfig)
		}

		// Payments
		payments := api.Group("/payments")
		{
			payments.POST("/create-intent", middleware.AuthMiddleware(), handlers.CreatePaymentIntent)
			payments.POST("/:id/confirm", middleware.AuthMiddleware(), handlers.ConfirmPayment)
			payments.GET("/:id/status", middleware.AuthMiddleware(), handlers.GetPaymentStatus)
			payments.GET("", middleware.AuthMiddleware(), handlers.GetUserPayments)
		}
	}

	log.Println("🚀 Nitrous API server starting on :8080")
	r.Run(":8080")
}
