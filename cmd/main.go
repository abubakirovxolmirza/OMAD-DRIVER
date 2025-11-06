package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	fiberSwagger "github.com/gofiber/swagger"

	docs "taxi-service/docs"
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

// @host api.omad-driver.uz
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

	// Setup application
	app := setupApp(cfg)

	// Start server
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Starting server on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupApp(cfg *config.Config) *fiber.App {
	app := fiber.New(fiber.Config{})

	docs.SwaggerInfo.Host = cfg.Server.Domain
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"https", "http"}

	// Core middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(middleware.CORSMiddleware(cfg.CORS.AllowedOrigins))

	// Serve static files (uploads)
	app.Static("/uploads", cfg.Upload.Directory)

	// Swagger documentation
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// API v1 routes
	v1 := app.Group("/api").Group("/v1")

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
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// Region & district routes (public)
	v1.Get("/regions", regionHandler.GetRegions)
	v1.Get("/regions/:id", regionHandler.GetRegion)
	v1.Get("/regions/:id/districts", regionHandler.GetDistricts)
	v1.Get("/districts/:id", regionHandler.GetDistrict)

	// Protected routes (require authentication)
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(cfg.JWT.Secret))

	// Auth/Profile routes
	profile := protected.Group("/auth")
	profile.Get("/profile", authHandler.GetProfile)
	profile.Put("/profile", authHandler.UpdateProfile)
	profile.Post("/change-password", authHandler.ChangePassword)
	profile.Post("/avatar", authHandler.UploadAvatar)

	// Order routes (users)
	orders := protected.Group("/orders")
	orders.Post("/taxi", orderHandler.CreateTaxiOrder)
	orders.Post("/delivery", orderHandler.CreateDeliveryOrder)
	orders.Get("/my", orderHandler.GetMyOrders)
	orders.Get("/:id", orderHandler.GetOrderByID)
	orders.Post("/:id/cancel", orderHandler.CancelOrder)

	// Rating routes
	protected.Post("/ratings", ratingHandler.CreateRating)
	protected.Get("/ratings/driver/:driver_id", ratingHandler.GetDriverRatings)

	// Notification routes
	protected.Get("/notifications", notificationHandler.GetMyNotifications)
	protected.Post("/notifications/:id/read", notificationHandler.MarkNotificationRead)

	// Feedback routes
	protected.Post("/feedback", feedbackHandler.SubmitFeedback)

	// Driver routes
	driver := protected.Group("/driver")
	driver.Post("/apply", driverHandler.ApplyAsDriver)

	driverOnly := driver.Group("")
	driverOnly.Use(middleware.RoleMiddleware(models.RoleDriver, models.RoleAdmin, models.RoleSuperAdmin))
	driverOnly.Get("/profile", driverHandler.GetDriverProfile)
	driverOnly.Put("/profile", driverHandler.UpdateDriverProfile)
	driverOnly.Get("/orders/new", driverHandler.GetNewOrders)
	driverOnly.Post("/orders/:id/accept", driverHandler.AcceptOrder)
	driverOnly.Post("/orders/:id/complete", driverHandler.CompleteOrder)
	driverOnly.Get("/orders", driverHandler.GetDriverOrders)
	driverOnly.Get("/statistics", driverHandler.GetDriverStatistics)

	// Admin routes
	admin := protected.Group("/admin")
	admin.Use(middleware.RoleMiddleware(models.RoleAdmin, models.RoleSuperAdmin))
	admin.Get("/driver-applications", adminHandler.GetDriverApplications)
	admin.Post("/driver-applications/:id/review", adminHandler.ReviewDriverApplication)
	admin.Get("/drivers", adminHandler.GetDrivers)
	admin.Post("/drivers/:id/add-balance", adminHandler.AddDriverBalance)
	admin.Post("/users/:id/block", adminHandler.BlockUnblockUser)
	admin.Post("/pricing", adminHandler.SetPricing)
	admin.Get("/pricing", adminHandler.GetAllPricing)
	admin.Get("/orders", adminHandler.GetAllOrders)
	admin.Get("/statistics", adminHandler.GetStatistics)
	admin.Get("/feedback", adminHandler.GetFeedback)

	// Region & District management
	admin.Post("/regions", regionHandler.CreateRegion)
	admin.Put("/regions/:id", regionHandler.UpdateRegion)
	admin.Delete("/regions/:id", regionHandler.DeleteRegion)
	admin.Post("/districts", regionHandler.CreateDistrict)
	admin.Put("/districts/:id", regionHandler.UpdateDistrict)
	admin.Delete("/districts/:id", regionHandler.DeleteDistrict)

	// Superadmin routes
	superadmin := admin.Group("")
	superadmin.Use(middleware.RoleMiddleware(models.RoleSuperAdmin))
	superadmin.Post("/create-admin", adminHandler.CreateAdmin)
	superadmin.Post("/users/:id/reset-password", adminHandler.ResetUserPassword)

	return app
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
