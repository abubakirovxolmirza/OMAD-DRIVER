package handlers

import (
	"database/sql"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"taxi-service/internal/config"
	"taxi-service/internal/database"
	"taxi-service/internal/middleware"
	"taxi-service/internal/models"
	"taxi-service/internal/utils"
)

// AdminHandler handles admin-related endpoints
type AdminHandler struct {
	cfg *config.Config
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(cfg *config.Config) *AdminHandler {
	return &AdminHandler{cfg: cfg}
}

// ReviewDriverApplicationRequest represents application review
type ReviewDriverApplicationRequest struct {
	Status          string `json:"status" validate:"required,oneof=approved rejected"`
	RejectionReason string `json:"rejection_reason"`
}

// ReviewDriverApplication godoc
// @Summary Review driver application
// @Description Approve or reject a driver application
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Application ID"
// @Param request body ReviewDriverApplicationRequest true "Review details"
// @Success 200 {object} map[string]string
// @Router /admin/driver-applications/{id}/review [post]

func (h *AdminHandler) ReviewDriverApplication(c *fiber.Ctx) error {
	adminID, _ := middleware.GetUserID(c)
	applicationID := c.Param("id")

	var req ReviewDriverApplicationRequest
	if err := parseAndValidateJSON(c, &req); err != nil {
		return err
	}

	// Get application
	var application models.DriverApplication
	err := database.DB.QueryRow(`
		SELECT id, user_id, full_name, phone_number, car_model, car_number, license_image
		FROM driver_applications WHERE id = $1 AND status = 'pending'
	`, applicationID).Scan(
		&application.ID, &application.UserID, &application.FullName, &application.PhoneNumber,
		&application.CarModel, &application.CarNumber, &application.LicenseImage,
	)
	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Application not found or already reviewed"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	// Begin transaction
	tx, err := database.DB.Begin()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}
	defer tx.Rollback()

	// Update application status
	// Update application status
	var rejectionReason *string
	if req.Status == "rejected" && req.RejectionReason != "" {
    	rejectionReason = &req.RejectionReason
	}

	_, err = tx.Exec(`
    	UPDATE driver_applications 
    	SET status = $1, rejection_reason = $2, reviewed_by = $3, reviewed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
    	WHERE id = $4
	`, req.Status, rejectionReason, adminID, applicationID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update application"})
	}

	if req.Status == "approved" {
		// Update user role to driver (preserve admin/superadmin)
		_, err = tx.Exec(`
			UPDATE users SET role = CASE 
				WHEN role IN ($1, $2) THEN role
				ELSE $3
			END,
			updated_at = CURRENT_TIMESTAMP
			WHERE id = $4
		`, models.RoleAdmin, models.RoleSuperAdmin, models.RoleDriver, application.UserID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update user role"})
		}

		// Create or update driver record
		_, err = tx.Exec(`
			INSERT INTO drivers (user_id, full_name, car_model, car_number, license_image, status, balance, is_active)
			VALUES ($1, $2, $3, $4, $5, 'approved', 0, true)
			ON CONFLICT (user_id) DO UPDATE
			SET full_name = EXCLUDED.full_name,
				car_model = EXCLUDED.car_model,
				car_number = EXCLUDED.car_number,
				license_image = EXCLUDED.license_image,
				status = 'approved',
				is_active = true,
				updated_at = CURRENT_TIMESTAMP
		`, application.UserID, application.FullName, application.CarModel, application.CarNumber, application.LicenseImage)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to upsert driver profile"})
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to commit transaction"})
	}

	// Create notification for user
	notifMessage := "Your driver application has been approved!"
	if req.Status == "rejected" {
		notifMessage = "Your driver application has been rejected."
		if req.RejectionReason != "" {
			notifMessage += " Reason: " + req.RejectionReason
		}
	}
	database.DB.Exec(`
		INSERT INTO notifications (user_id, title, message, type)
		VALUES ($1, $2, $3, $4)
	`, application.UserID, "Driver Application Status", notifMessage, "application_review")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Application reviewed successfully"})
}

// GetDriverApplications godoc
// @Summary Get driver applications
// @Description Get list of driver applications
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by status"
// @Success 200 {array} models.DriverApplication
// @Router /admin/driver-applications [get]

func (h *AdminHandler) GetDriverApplications(c *fiber.Ctx) error {
	status := c.Query("status")

	query := "SELECT * FROM driver_applications"
	args := []interface{}{}

	if status != "" {
		query += " WHERE status = $1"
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC"

rows, err := database.DB.Query(query, args...)
if err != nil {
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch applications"})
}
	defer rows.Close()

	applications := []models.DriverApplication{}
	for rows.Next() {
		var app models.DriverApplication
		err := rows.Scan(
			&app.ID, &app.UserID, &app.FullName, &app.PhoneNumber, &app.CarModel,
			&app.CarNumber, &app.LicenseImage, &app.Status, &app.RejectionReason,
			&app.ReviewedBy, &app.ReviewedAt, &app.CreatedAt, &app.UpdatedAt,
		)
		if err != nil {
			continue
		}
		applications = append(applications, app)
	}

	return c.Status(fiber.StatusOK).JSON(applications)
}

// BlockUserRequest represents block/unblock request
type BlockUserRequest struct {
	IsBlocked bool `json:"is_blocked"`
}

// BlockUnblockUser godoc
// @Summary Block or unblock a user
// @Description Block or unblock a user or driver
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body BlockUserRequest true "Block status"
// @Success 200 {object} map[string]string
// @Router /admin/users/{id}/block [post]

func (h *AdminHandler) BlockUnblockUser(c *fiber.Ctx) error {
	userID := c.Param("id")

	var req BlockUserRequest
	if err := parseAndValidateJSON(c, &req); err != nil {
		return err
	}

	_, err := database.DB.Exec(`
		UPDATE users SET is_blocked = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2
	`, req.IsBlocked, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update user"})
	}

	action := "unblocked"
	if req.IsBlocked {
		action = "blocked"
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": fmt.Sprintf("User %s successfully", action)})
}

// GetDrivers godoc
// @Summary Get all drivers
// @Description Get list of all drivers with optional filters
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by status"
// @Success 200 {array} models.Driver
// @Router /admin/drivers [get]

func (h *AdminHandler) GetDrivers(c *fiber.Ctx) error {
	status := c.Query("status")

	query := "SELECT * FROM drivers"
	args := []interface{}{}

	if status != "" {
		query += " WHERE status = $1"
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch drivers"})
	}
	defer rows.Close()

	drivers := []models.Driver{}
	for rows.Next() {
		var driver models.Driver
		err := rows.Scan(
			&driver.ID, &driver.UserID, &driver.FullName, &driver.CarModel, &driver.CarNumber,
			&driver.LicenseImage, &driver.Balance, &driver.Rating, &driver.TotalRatings,
			&driver.Status, &driver.IsActive, &driver.CreatedAt, &driver.UpdatedAt,
		)
		if err != nil {
			continue
		}
		drivers = append(drivers, driver)
	}

	return c.Status(fiber.StatusOK).JSON(drivers)
}

// AddBalanceRequest represents balance addition request
type AddBalanceRequest struct {
	Amount float64 `json:"amount" validate:"required,gt=0"`
}

// AddDriverBalance godoc
// @Summary Add balance to driver
// @Description Add balance to a driver's account
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Driver ID"
// @Param request body AddBalanceRequest true "Amount to add"
// @Success 200 {object} map[string]string
// @Router /admin/drivers/{id}/add-balance [post]

func (h *AdminHandler) AddDriverBalance(c *fiber.Ctx) error {
	adminID, _ := middleware.GetUserID(c)
	driverID := c.Param("id")

	var req AddBalanceRequest
	if err := parseAndValidateJSON(c, &req); err != nil {
		return err
	}

	// Begin transaction
	tx, err := database.DB.Begin()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}
	defer tx.Rollback()

	// Update driver balance
	result, err := tx.Exec(`UPDATE drivers SET balance = balance + $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, req.Amount, driverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update balance"})
	}
	rowsChanged, _ := result.RowsAffected()
	if rowsChanged == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Driver not found"})
	}

	// Create transaction record
	_, err = tx.Exec(`
		INSERT INTO transactions (driver_id, amount, type, description, created_by)
		VALUES ($1, $2, $3, $4, $5)
	`, driverID, req.Amount, "credit", "Balance added by admin", adminID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create transaction"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to commit transaction"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Balance added successfully"})
}

// SetPricingRequest represents pricing configuration
type SetPricingRequest struct {
	FromRegionID   int64   `json:"from_region_id" validate:"required"`
	ToRegionID     int64   `json:"to_region_id" validate:"required"`
	BasePrice      float64 `json:"base_price" validate:"required,gt=0"`
	PricePerPerson float64 `json:"price_per_person" validate:"required,gte=0"`
	ServiceFee     float64 `json:"service_fee" validate:"required,gte=0,lte=100"`
}

// SetPricing godoc
// @Summary Set pricing for route
// @Description Set or update pricing between two regions
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body SetPricingRequest true "Pricing details"
// @Success 200 {object} models.Pricing
// @Router /admin/pricing [post]
func (h *AdminHandler) SetPricing(c *fiber.Ctx) error {
	var req SetPricingRequest
	if err := parseAndValidateJSON(c, &req); err != nil {
		return err
	}

	if req.FromRegionID == req.ToRegionID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "From and To regions must differ"})
	}

	var pricing models.Pricing
	err := database.DB.QueryRow(`
		INSERT INTO pricing (from_region_id, to_region_id, base_price, price_per_person, service_fee)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (from_region_id, to_region_id) 
		DO UPDATE SET base_price = $3, price_per_person = $4, service_fee = $5, updated_at = CURRENT_TIMESTAMP
		RETURNING id, from_region_id, to_region_id, base_price, price_per_person, service_fee, created_at, updated_at
	`, req.FromRegionID, req.ToRegionID, req.BasePrice, req.PricePerPerson, req.ServiceFee).Scan(
		&pricing.ID, &pricing.FromRegionID, &pricing.ToRegionID, &pricing.BasePrice,
		&pricing.PricePerPerson, &pricing.ServiceFee, &pricing.CreatedAt, &pricing.UpdatedAt,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to set pricing"})
	}

	return c.Status(fiber.StatusOK).JSON(pricing)
}

// GetAllPricing godoc
// @Summary Get all pricing
// @Description Get all configured pricing routes
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.Pricing
// @Router /admin/pricing [get]

func (h *AdminHandler) GetAllPricing(c *fiber.Ctx) error {
	rows, err := database.DB.Query("SELECT * FROM pricing ORDER BY created_at DESC")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch pricing"})
	}
	defer rows.Close()

	pricings := []models.Pricing{}
	for rows.Next() {
		var pricing models.Pricing
		err := rows.Scan(
			&pricing.ID, &pricing.FromRegionID, &pricing.ToRegionID, &pricing.BasePrice,
			&pricing.PricePerPerson, &pricing.ServiceFee, &pricing.CreatedAt, &pricing.UpdatedAt,
		)
		if err != nil {
			continue
		}
		pricings = append(pricings, pricing)
	}

	return c.Status(fiber.StatusOK).JSON(pricings)
}

// GetAllOrders godoc
// @Summary Get all orders
// @Description Get all orders with filters (admin only)
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by status"
// @Param type query string false "Filter by type"
// @Param from_date query string false "From date (YYYY-MM-DD)"
// @Param to_date query string false "To date (YYYY-MM-DD)"
// @Success 200 {array} models.Order
// @Router /admin/orders [get]

func (h *AdminHandler) GetAllOrders(c *fiber.Ctx) error {
	status := c.Query("status")
	orderType := c.Query("type")
	fromDate := c.Query("from_date")
	toDate := c.Query("to_date")

	query := "SELECT * FROM orders WHERE 1=1"
	args := []interface{}{}
	argCount := 0

	if status != "" {
		argCount++
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
	}
	if orderType != "" {
		argCount++
		query += fmt.Sprintf(" AND order_type = $%d", argCount)
		args = append(args, orderType)
	}
	if fromDate != "" {
		argCount++
		query += fmt.Sprintf(" AND DATE(created_at) >= $%d", argCount)
		args = append(args, fromDate)
	}
	if toDate != "" {
		argCount++
		query += fmt.Sprintf(" AND DATE(created_at) <= $%d", argCount)
		args = append(args, toDate)
	}

	query += " ORDER BY created_at DESC"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch orders"})
	}
	defer rows.Close()

	orders := []models.Order{}
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID, &order.UserID, &order.DriverID, &order.OrderType, &order.Status,
			&order.CustomerName, &order.CustomerPhone, &order.RecipientPhone,
			&order.FromRegionID, &order.FromDistrictID, &order.ToRegionID, &order.ToDistrictID,
			&order.PassengerCount, &order.DeliveryType, &order.ScheduledDate,
			&order.TimeRangeStart, &order.TimeRangeEnd, &order.Price, &order.ServiceFee,
			&order.DiscountPercentage, &order.FinalPrice, &order.Notes, &order.CancellationReason,
			&order.AcceptedAt, &order.AcceptDeadline, &order.CompletedAt, &order.CancelledAt,
			&order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			continue
		}
		orders = append(orders, order)
	}

	return c.Status(fiber.StatusOK).JSON(orders)
}

// AdminStatistics represents admin statistics
type AdminStatistics struct {
	TotalUsers          int     `json:"total_users"`
	TotalDrivers        int     `json:"total_drivers"`
	ActiveDrivers       int     `json:"active_drivers"`
	TotalOrders         int     `json:"total_orders"`
	CompletedOrders     int     `json:"completed_orders"`
	TotalRevenue        float64 `json:"total_revenue"`
	TodayOrders         int     `json:"today_orders"`
	TodayRevenue        float64 `json:"today_revenue"`
}

// GetStatistics godoc
// @Summary Get platform statistics
// @Description Get overall platform statistics
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} AdminStatistics
// @Router /admin/statistics [get]

func (h *AdminHandler) GetStatistics(c *fiber.Ctx) error {
	var stats AdminStatistics

	// Total users
	database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = $1", models.RoleUser).Scan(&stats.TotalUsers)

	// Drivers
	database.DB.QueryRow("SELECT COUNT(*) FROM drivers").Scan(&stats.TotalDrivers)
	database.DB.QueryRow("SELECT COUNT(*) FROM drivers WHERE is_active = true AND status = 'approved'").Scan(&stats.ActiveDrivers)

	// Orders
	database.DB.QueryRow("SELECT COUNT(*) FROM orders").Scan(&stats.TotalOrders)
	database.DB.QueryRow("SELECT COUNT(*) FROM orders WHERE status = $1", models.OrderStatusCompleted).Scan(&stats.CompletedOrders)
	database.DB.QueryRow("SELECT COALESCE(SUM(service_fee), 0) FROM orders WHERE status = $1", models.OrderStatusCompleted).Scan(&stats.TotalRevenue)

	// Today's stats
	database.DB.QueryRow("SELECT COUNT(*) FROM orders WHERE DATE(created_at) = CURRENT_DATE").Scan(&stats.TodayOrders)
	database.DB.QueryRow("SELECT COALESCE(SUM(service_fee), 0) FROM orders WHERE DATE(created_at) = CURRENT_DATE AND status = $1", models.OrderStatusCompleted).Scan(&stats.TodayRevenue)

	return c.Status(fiber.StatusOK).JSON(stats)
}

// ResetPasswordRequest represents password reset by admin
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

// ResetUserPassword godoc
// @Summary Reset user password (superadmin only)
// @Description Reset a user's password
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body ResetPasswordRequest true "New password"
// @Success 200 {object} map[string]string
// @Router /admin/users/{id}/reset-password [post]

func (h *AdminHandler) ResetUserPassword(c *fiber.Ctx) error {
	userID := c.Param("id")

	var req ResetPasswordRequest
	if err := parseAndValidateJSON(c, &req); err != nil {
		return err
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to process password"})
	}

	_, err = database.DB.Exec(`
		UPDATE users SET password = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2
	`, hashedPassword, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to reset password"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Password reset successfully"})
}

// CreateAdminRequest represents admin creation
type CreateAdminRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Password    string `json:"password" validate:"required,min=6"`
}

// CreateAdmin godoc
// @Summary Create admin user (superadmin only)
// @Description Create a new admin user
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateAdminRequest true "Admin details"
// @Success 201 {object} models.User
// @Router /admin/create-admin [post]

func (h *AdminHandler) CreateAdmin(c *fiber.Ctx) error {
	var req CreateAdminRequest
	if err := parseAndValidateJSON(c, &req); err != nil {
		return err
	}

	// Check if phone already exists
	var existingID int64
	if err := database.DB.QueryRow("SELECT id FROM users WHERE phone_number = $1", req.PhoneNumber).Scan(&existingID); err != nil && err != sql.ErrNoRows {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}
	if existingID > 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Phone number already registered"})
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to process password"})
	}

	// Create admin user
	var user models.User
	err = database.DB.QueryRow(`
		INSERT INTO users (phone_number, name, password, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, phone_number, name, role, language, created_at, updated_at
	`, req.PhoneNumber, req.Name, hashedPassword, models.RoleAdmin).Scan(
		&user.ID, &user.PhoneNumber, &user.Name, &user.Role,
		&user.Language, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create admin"})
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

// GetFeedback godoc
// @Summary Get all feedback
// @Description Get all user feedback/suggestions
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.Feedback
// @Router /admin/feedback [get]

func (h *AdminHandler) GetFeedback(c *fiber.Ctx) error {
	rows, err := database.DB.Query("SELECT * FROM feedback ORDER BY created_at DESC")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch feedback"})
	}
	defer rows.Close()

	feedbacks := []models.Feedback{}
	for rows.Next() {
		var feedback models.Feedback
		err := rows.Scan(&feedback.ID, &feedback.UserID, &feedback.Message, &feedback.CreatedAt)
		if err != nil {
			continue
		}
		feedbacks = append(feedbacks, feedback)
	}

	return c.Status(fiber.StatusOK).JSON(feedbacks)
}
