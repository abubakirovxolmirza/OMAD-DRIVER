package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

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
// @schemes https http

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
	app := setupRouter(cfg)

	// Start server
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Starting Fiber server on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRouter(cfg *config.Config) *fiber.App {
	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Taxi Service API v1.0",
		ErrorHandler: errorHandler,
	})

	// CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowedOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Content-Type, Authorization, Accept, Origin, Cache-Control, X-Requested-With",
		AllowCredentials: true,
		ExposeHeaders:    "Content-Length, X-JSON-Response",
	}))

	// Serve static files (uploads)
	app.Static("/uploads", cfg.Upload.Directory)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.H{"status": "ok"})
	})

	// API v1 routes
	api := app.Group("/api/v1")

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
	auth := api.Group("/auth")
	{
		auth.Post("/register", authHandler.RegisterFiber)
		auth.Post("/login", authHandler.LoginFiber)
	}

	// Region routes (public)
	regions := api.Group("/regions")
	{
		regions.Get("", regionHandler.GetRegionsFiber)
		regions.Get("/:id", regionHandler.GetRegionFiber)
		regions.Get("/:id/districts", regionHandler.GetDistrictsFiber)
	}

	// District routes (public)
	districts := api.Group("/districts")
	{
		districts.Get("/:id", regionHandler.GetDistrictFiber)
	}

	// Protected routes (require authentication)
	protected := api.Group("")
	protected.Use(middleware.AuthMiddlewareFiber(cfg.JWT.Secret))

	// Auth/Profile routes
	profile := protected.Group("/auth")
	{
		profile.Get("/profile", authHandler.GetProfileFiber)
		profile.Put("/profile", authHandler.UpdateProfileFiber)
		profile.Post("/change-password", authHandler.ChangePasswordFiber)
		profile.Post("/avatar", authHandler.UploadAvatarFiber)
	}

	// Order routes (users)
	orders := protected.Group("/orders")
	{
		orders.Post("/taxi", orderHandler.CreateTaxiOrderFiber)
		orders.Post("/delivery", orderHandler.CreateDeliveryOrderFiber)
		orders.Get("/my", orderHandler.GetMyOrdersFiber)
		orders.Get("/:id", orderHandler.GetOrderByIDFiber)
		orders.Post("/:id/cancel", orderHandler.CancelOrderFiber)
	}

	// Rating routes
	ratings := protected.Group("/ratings")
	{
		ratings.Post("", ratingHandler.CreateRatingFiber)
		ratings.Get("/driver/:driver_id", ratingHandler.GetDriverRatingsFiber)
	}

	// Notification routes
	notifications := protected.Group("/notifications")
	{
		notifications.Get("", notificationHandler.GetMyNotificationsFiber)
		notifications.Post("/:id/read", notificationHandler.MarkNotificationReadFiber)
	}

	// Feedback routes
	feedback := protected.Group("/feedback")
	{
		feedback.Post("", feedbackHandler.SubmitFeedbackFiber)
	}

	// Driver routes
	driver := protected.Group("/driver")
	{
		driver.Post("/apply", driverHandler.ApplyAsDriverFiber)

		driverOnly := driver.Group("")
		driverOnly.Use(middleware.RoleMiddlewareFiber(models.RoleDriver, models.RoleAdmin, models.RoleSuperAdmin))
		{
			driverOnly.Get("/profile", driverHandler.GetDriverProfileFiber)
			driverOnly.Put("/profile", driverHandler.UpdateDriverProfileFiber)
			driverOnly.Get("/orders/new", driverHandler.GetNewOrdersFiber)
			driverOnly.Post("/orders/:id/accept", driverHandler.AcceptOrderFiber)
			driverOnly.Post("/orders/:id/complete", driverHandler.CompleteOrderFiber)
			driverOnly.Get("/orders", driverHandler.GetDriverOrdersFiber)
			driverOnly.Get("/statistics", driverHandler.GetDriverStatisticsFiber)
		}
	}

	// Admin routes
	admin := protected.Group("/admin")
	admin.Use(middleware.RoleMiddlewareFiber(models.RoleAdmin, models.RoleSuperAdmin))
	{
		admin.Get("/driver-applications", adminHandler.GetDriverApplicationsFiber)
		admin.Post("/driver-applications/:id/review", adminHandler.ReviewDriverApplicationFiber)
		admin.Get("/drivers", adminHandler.GetDriversFiber)
		admin.Post("/drivers/:id/add-balance", adminHandler.AddDriverBalanceFiber)
		admin.Post("/users/:id/block", adminHandler.BlockUnblockUserFiber)
		admin.Post("/pricing", adminHandler.SetPricingFiber)
		admin.Get("/pricing", adminHandler.GetAllPricingFiber)
		admin.Get("/orders", adminHandler.GetAllOrdersFiber)
		admin.Get("/statistics", adminHandler.GetStatisticsFiber)
		admin.Get("/feedback", adminHandler.GetFeedbackFiber)

		admin.Post("/regions", regionHandler.CreateRegionFiber)
		admin.Put("/regions/:id", regionHandler.UpdateRegionFiber)
		admin.Delete("/regions/:id", regionHandler.DeleteRegionFiber)
		admin.Post("/districts", regionHandler.CreateDistrictFiber)
		admin.Put("/districts/:id", regionHandler.UpdateDistrictFiber)
		admin.Delete("/districts/:id", regionHandler.DeleteDistrictFiber)

		superadmin := admin.Group("")
		superadmin.Use(middleware.RoleMiddlewareFiber(models.RoleSuperAdmin))
		{
			superadmin.Post("/create-admin", adminHandler.CreateAdminFiber)
			superadmin.Post("/users/:id/reset-password", adminHandler.ResetUserPasswordFiber)
		}
	}

	return app
}

// Error handler for Fiber
func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := err.Error()

	if fe, ok := err.(*fiber.Error); ok {
		code = fe.Code
		message = fe.Message
	}

	return c.Status(code).JSON(fiber.H{
		"error": message,
	})
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
