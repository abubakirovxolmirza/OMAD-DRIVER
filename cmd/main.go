package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "taxi-service/docs"
	"taxi-service/internal/config"
	"taxi-service/internal/database"
	"taxi-service/internal/handlers"
	"taxi-service/internal/middleware"
	"taxi-service/internal/models"
)

// @title Taxi Service API
// @version 1.0
// @description Professional taxi and delivery service backend API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@taxiservice.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host 64.225.107.130
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	if err := database.Connect(&cfg.Database); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Initialize database schema
	if err := database.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	// Seed initial data (for development)
	if cfg.Server.Env == "development" {
		if err := database.SeedInitialData(); err != nil {
			log.Printf("Warning: Failed to seed initial data: %v", err)
		}

		// Create default superadmin if not exists (only in development)
		createDefaultSuperAdmin()
	}

	// Create upload directory
	if err := os.MkdirAll(cfg.Upload.Directory, 0755); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	// Setup router
	router := setupRouter(cfg)

	// Start server
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Starting server on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRouter(cfg *config.Config) *gin.Engine {
	// Set Gin mode
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// CORS middleware
	router.Use(middleware.CORSMiddleware(cfg.CORS.AllowedOrigins))

	// Serve static files (uploads)
	router.Static("/uploads", cfg.Upload.Directory)

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Initialize handlers
		authHandler := handlers.NewAuthHandler(cfg)
		orderHandler := handlers.NewOrderHandler(cfg)
		driverHandler := handlers.NewDriverHandler(cfg)
		adminHandler := handlers.NewAdminHandler(cfg)
		ratingHandler := handlers.NewRatingHandler()
		notificationHandler := handlers.NewNotificationHandler()
		regionHandler := handlers.NewRegionHandler()
		feedbackHandler := handlers.NewFeedbackHandler()

		// Public routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Region routes (public)
		regions := v1.Group("/regions")
		{
			regions.GET("", regionHandler.GetRegions)
			regions.GET("/:id", regionHandler.GetRegion)
			regions.GET("/:id/districts", regionHandler.GetDistricts)
		}

		// District routes (public)
		districts := v1.Group("/districts")
		{
			districts.GET("/:id", regionHandler.GetDistrict)
		}

		// Protected routes (require authentication)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		{
			// Auth/Profile routes
			profile := protected.Group("/auth")
			{
				profile.GET("/profile", authHandler.GetProfile)
				profile.PUT("/profile", authHandler.UpdateProfile)
				profile.POST("/change-password", authHandler.ChangePassword)
				profile.POST("/avatar", authHandler.UploadAvatar)
			}

			// Order routes (users)
			orders := protected.Group("/orders")
			{
				orders.POST("/taxi", orderHandler.CreateTaxiOrder)
				orders.POST("/delivery", orderHandler.CreateDeliveryOrder)
				orders.GET("/my", orderHandler.GetMyOrders)
				orders.GET("/:id", orderHandler.GetOrderByID)
				orders.POST("/:id/cancel", orderHandler.CancelOrder)
			}

			// Rating routes
			ratings := protected.Group("/ratings")
			{
				ratings.POST("", ratingHandler.CreateRating)
				ratings.GET("/driver/:driver_id", ratingHandler.GetDriverRatings)
			}

			// Notification routes
			notifications := protected.Group("/notifications")
			{
				notifications.GET("", notificationHandler.GetMyNotifications)
				notifications.POST("/:id/read", notificationHandler.MarkNotificationRead)
			}

			// Feedback routes
			feedback := protected.Group("/feedback")
			{
				feedback.POST("", feedbackHandler.SubmitFeedback)
			}

			// Driver routes (require driver role)
			driver := protected.Group("/driver")
			{
				// Application (any user can apply)
				driver.POST("/apply", driverHandler.ApplyAsDriver)

				// Driver-only routes
				driverOnly := driver.Group("")
				driverOnly.Use(middleware.RoleMiddleware(models.RoleDriver, models.RoleAdmin, models.RoleSuperAdmin))
				{
					driverOnly.GET("/profile", driverHandler.GetDriverProfile)
					driverOnly.PUT("/profile", driverHandler.UpdateDriverProfile)
					driverOnly.GET("/orders/new", driverHandler.GetNewOrders)
					driverOnly.POST("/orders/:id/accept", driverHandler.AcceptOrder)
					driverOnly.POST("/orders/:id/complete", driverHandler.CompleteOrder)
					driverOnly.GET("/orders", driverHandler.GetDriverOrders)
					driverOnly.GET("/statistics", driverHandler.GetDriverStatistics)
				}
			}

			// Admin routes (require admin or superadmin role)
			admin := protected.Group("/admin")
			admin.Use(middleware.RoleMiddleware(models.RoleAdmin, models.RoleSuperAdmin))
			{
				admin.GET("/driver-applications", adminHandler.GetDriverApplications)
				admin.POST("/driver-applications/:id/review", adminHandler.ReviewDriverApplication)
				admin.GET("/drivers", adminHandler.GetDrivers)
				admin.POST("/drivers/:id/add-balance", adminHandler.AddDriverBalance)
				admin.POST("/users/:id/block", adminHandler.BlockUnblockUser)
				admin.POST("/pricing", adminHandler.SetPricing)
				admin.GET("/pricing", adminHandler.GetAllPricing)
				admin.GET("/orders", adminHandler.GetAllOrders)
				admin.GET("/statistics", adminHandler.GetStatistics)
				admin.GET("/feedback", adminHandler.GetFeedback)

				// Region & District management (Admin only)
				admin.POST("/regions", regionHandler.CreateRegion)
				admin.PUT("/regions/:id", regionHandler.UpdateRegion)
				admin.DELETE("/regions/:id", regionHandler.DeleteRegion)
				admin.POST("/districts", regionHandler.CreateDistrict)
				admin.PUT("/districts/:id", regionHandler.UpdateDistrict)
				admin.DELETE("/districts/:id", regionHandler.DeleteDistrict)

				// Superadmin only routes
				superadmin := admin.Group("")
				superadmin.Use(middleware.RoleMiddleware(models.RoleSuperAdmin))
				{
					superadmin.POST("/create-admin", adminHandler.CreateAdmin)
					superadmin.POST("/users/:id/reset-password", adminHandler.ResetUserPassword)
				}
			}
		}
	}

	return router
}

func createDefaultSuperAdmin() {
	// Check if superadmin exists
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = $1", models.RoleSuperAdmin).Scan(&count)
	if count > 0 {
		return
	}

	// Create default superadmin
	// Password: admin123
	hashedPassword := "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi" // bcrypt hash of "admin123"
	
	_, err := database.DB.Exec(`
		INSERT INTO users (phone_number, name, password, role)
		VALUES ($1, $2, $3, $4)
	`, "+998901234567", "Super Admin", hashedPassword, models.RoleSuperAdmin)
	
	if err != nil {
		log.Printf("Warning: Failed to create default superadmin: %v", err)
	} else {
		log.Println("Default superadmin created - Phone: +998901234567, Password: admin123")
	}
}
