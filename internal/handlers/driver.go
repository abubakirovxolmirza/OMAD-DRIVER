package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"taxi-service/internal/config"
	"taxi-service/internal/database"
	"taxi-service/internal/middleware"
	"taxi-service/internal/models"
	"taxi-service/internal/utils"
)

// DriverHandler handles driver-related endpoints
type DriverHandler struct {
	cfg *config.Config
}

// NewDriverHandler creates a new driver handler
func NewDriverHandler(cfg *config.Config) *DriverHandler {
	return &DriverHandler{cfg: cfg}
}

// ApplyAsDriverRequest represents driver application request
type ApplyAsDriverRequest struct {
	FullName   string `json:"full_name" binding:"required"`
	CarModel   string `json:"car_model" binding:"required"`
	CarNumber  string `json:"car_number" binding:"required"`
}

// ApplyAsDriver godoc
// @Summary Apply to become a driver
// @Description Submit an application to become a driver
// @Tags Driver
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param full_name formData string true "Full Name"
// @Param car_model formData string true "Car Model"
// @Param car_number formData string true "Car Number"
// @Param license_image formData file true "License Image"
// @Success 201 {object} models.DriverApplication
// @Router /driver/apply [post]
func (h *DriverHandler) ApplyAsDriver(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	// Check if user already has driver role
	var role string
	database.DB.QueryRow("SELECT role FROM users WHERE id = $1", userID).Scan(&role)
	if role == string(models.RoleDriver) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You are already a driver"})
		return
	}

	// Check if application already exists
	var existingID int64
	err := database.DB.QueryRow(`
		SELECT id FROM driver_applications 
		WHERE user_id = $1 AND status = 'pending'
	`, userID).Scan(&existingID)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Application already submitted and pending review"})
		return
	}

	// Get form data
	fullName := c.PostForm("full_name")
	carModel := c.PostForm("car_model")
	carNumber := c.PostForm("car_number")

	if fullName == "" || carModel == "" || carNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
		return
	}

	// Get phone number
	var phoneNumber string
	database.DB.QueryRow("SELECT phone_number FROM users WHERE id = $1", userID).Scan(&phoneNumber)

	// Handle license image upload
	file, err := c.FormFile("license_image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "License image is required"})
		return
	}

	// Save license image
	licenseImage, err := utils.SaveUploadedFile(file, h.cfg.Upload.Directory, "licenses")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create application
	var application models.DriverApplication
	err = database.DB.QueryRow(`
		INSERT INTO driver_applications (user_id, full_name, phone_number, car_model, car_number, license_image, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, user_id, full_name, phone_number, car_model, car_number, license_image, status, created_at, updated_at
	`, userID, fullName, phoneNumber, carModel, carNumber, licenseImage, "pending").Scan(
		&application.ID, &application.UserID, &application.FullName, &application.PhoneNumber,
		&application.CarModel, &application.CarNumber, &application.LicenseImage, &application.Status,
		&application.CreatedAt, &application.UpdatedAt,
	)
	if err != nil {
		utils.DeleteFile(h.cfg.Upload.Directory, licenseImage)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create application"})
		return
	}

	// TODO: Send notification to telegram admin group

	c.JSON(http.StatusCreated, application)
}

// GetDriverProfile godoc
// @Summary Get driver profile
// @Description Get driver's profile information including balance and rating
// @Tags Driver
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.Driver
// @Router /driver/profile [get]
func (h *DriverHandler) GetDriverProfile(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var driver models.Driver
	err := database.DB.QueryRow(`
		SELECT id, user_id, full_name, car_model, car_number, license_image, 
		       balance, rating, total_ratings, status, is_active, created_at, updated_at
		FROM drivers WHERE user_id = $1
	`, userID).Scan(
		&driver.ID, &driver.UserID, &driver.FullName, &driver.CarModel, &driver.CarNumber,
		&driver.LicenseImage, &driver.Balance, &driver.Rating, &driver.TotalRatings,
		&driver.Status, &driver.IsActive, &driver.CreatedAt, &driver.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver profile not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, driver)
}

// UpdateDriverProfileRequest represents driver profile update
type UpdateDriverProfileRequest struct {
	FullName  string `json:"full_name"`
	CarModel  string `json:"car_model"`
	CarNumber string `json:"car_number"`
}

// UpdateDriverProfile godoc
// @Summary Update driver profile
// @Description Update driver's profile information
// @Tags Driver
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body UpdateDriverProfileRequest true "Profile update"
// @Success 200 {object} models.Driver
// @Router /driver/profile [put]
func (h *DriverHandler) UpdateDriverProfile(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req UpdateDriverProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var driver models.Driver
	err := database.DB.QueryRow(`
		UPDATE drivers SET full_name = $1, car_model = $2, car_number = $3, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $4
		RETURNING id, user_id, full_name, car_model, car_number, license_image, 
		          balance, rating, total_ratings, status, is_active, created_at, updated_at
	`, req.FullName, req.CarModel, req.CarNumber, userID).Scan(
		&driver.ID, &driver.UserID, &driver.FullName, &driver.CarModel, &driver.CarNumber,
		&driver.LicenseImage, &driver.Balance, &driver.Rating, &driver.TotalRatings,
		&driver.Status, &driver.IsActive, &driver.CreatedAt, &driver.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, driver)
}

// GetNewOrders godoc
// @Summary Get new available orders
// @Description Get list of orders available for drivers to accept
// @Tags Driver
// @Security BearerAuth
// @Produce json
// @Param type query string false "Filter by type (taxi/delivery)"
// @Param from_region query int false "Filter by from region"
// @Param to_region query int false "Filter by to region"
// @Success 200 {array} models.Order
// @Router /driver/orders/new [get]
func (h *DriverHandler) GetNewOrders(c *gin.Context) {
	orderType := c.Query("type")
	fromRegion := c.Query("from_region")
	toRegion := c.Query("to_region")

	query := `SELECT * FROM orders WHERE status = $1 AND (accept_deadline IS NULL OR accept_deadline > CURRENT_TIMESTAMP)`
	args := []interface{}{models.OrderStatusPending}
	argCount := 1

	if orderType != "" {
		argCount++
		query += fmt.Sprintf(" AND order_type = $%d", argCount)
		args = append(args, orderType)
	}
	if fromRegion != "" {
		argCount++
		query += fmt.Sprintf(" AND from_region_id = $%d", argCount)
		args = append(args, fromRegion)
	}
	if toRegion != "" {
		argCount++
		query += fmt.Sprintf(" AND to_region_id = $%d", argCount)
		args = append(args, toRegion)
	}

	query += " ORDER BY created_at DESC"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
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

	c.JSON(http.StatusOK, orders)
}

// AcceptOrder godoc
// @Summary Accept an order
// @Description Driver accepts an order if they have sufficient balance
// @Tags Driver
// @Security BearerAuth
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} models.Order
// @Router /driver/orders/{id}/accept [post]
func (h *DriverHandler) AcceptOrder(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	orderID := c.Param("id")

	// Get driver info
	var driver models.Driver
	err := database.DB.QueryRow(`
		SELECT id, balance, is_active FROM drivers WHERE user_id = $1
	`, userID).Scan(&driver.ID, &driver.Balance, &driver.IsActive)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver profile not found"})
		return
	}

	// Check if driver is active
	if !driver.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "Driver account is not active"})
		return
	}

	// Get order
	var order models.Order
	err = database.DB.QueryRow(`
		SELECT id, status, service_fee, accept_deadline
		FROM orders WHERE id = $1
	`, orderID).Scan(&order.ID, &order.Status, &order.ServiceFee, &order.AcceptDeadline)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check order status
	if order.Status != models.OrderStatusPending {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order is no longer available"})
		return
	}

	// Check accept deadline
	if order.AcceptDeadline != nil && order.AcceptDeadline.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order acceptance deadline has passed"})
		return
	}

	// Check driver balance
	if driver.Balance < order.ServiceFee {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance to accept order"})
		return
	}

	// Begin transaction
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	// Update order
	_, err = tx.Exec(`
		UPDATE orders SET driver_id = $1, status = $2, accepted_at = CURRENT_TIMESTAMP, 
		                  accept_deadline = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3 AND status = $4
	`, driver.ID, models.OrderStatusAccepted, orderID, models.OrderStatusPending)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept order"})
		return
	}

	// Deduct service fee from driver balance
	_, err = tx.Exec(`
		UPDATE drivers SET balance = balance - $1 WHERE id = $2
	`, order.ServiceFee, driver.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update balance"})
		return
	}

	// Create transaction record
	_, err = tx.Exec(`
		INSERT INTO transactions (driver_id, order_id, amount, type, description)
		VALUES ($1, $2, $3, $4, $5)
	`, driver.ID, order.ID, -order.ServiceFee, "debit", "Service fee for accepting order")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Get updated order
	err = database.DB.QueryRow(`SELECT * FROM orders WHERE id = $1`, orderID).Scan(
		&order.ID, &order.UserID, &order.DriverID, &order.OrderType, &order.Status,
		&order.CustomerName, &order.CustomerPhone, &order.RecipientPhone,
		&order.FromRegionID, &order.FromDistrictID, &order.ToRegionID, &order.ToDistrictID,
		&order.PassengerCount, &order.DeliveryType, &order.ScheduledDate,
		&order.TimeRangeStart, &order.TimeRangeEnd, &order.Price, &order.ServiceFee,
		&order.DiscountPercentage, &order.FinalPrice, &order.Notes, &order.CancellationReason,
		&order.AcceptedAt, &order.AcceptDeadline, &order.CompletedAt, &order.CancelledAt,
		&order.CreatedAt, &order.UpdatedAt,
	)

	// TODO: Send notification to user

	c.JSON(http.StatusOK, order)
}

// CompleteOrder godoc
// @Summary Complete an order
// @Description Mark an order as completed
// @Tags Driver
// @Security BearerAuth
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} map[string]string
// @Router /driver/orders/{id}/complete [post]
func (h *DriverHandler) CompleteOrder(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	orderID := c.Param("id")

	// Get driver ID
	var driverID int64
	err := database.DB.QueryRow("SELECT id FROM drivers WHERE user_id = $1", userID).Scan(&driverID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver profile not found"})
		return
	}

	// Update order
	result, err := database.DB.Exec(`
		UPDATE orders SET status = $1, completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND driver_id = $3 AND status = $4
	`, models.OrderStatusCompleted, orderID, driverID, models.OrderStatusAccepted)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete order"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order not found or not assigned to you"})
		return
	}

	// TODO: Send notification to user for rating

	c.JSON(http.StatusOK, gin.H{"message": "Order completed successfully"})
}

// GetDriverOrders godoc
// @Summary Get driver's orders
// @Description Get all orders assigned to the driver
// @Tags Driver
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by status"
// @Success 200 {array} models.Order
// @Router /driver/orders [get]
func (h *DriverHandler) GetDriverOrders(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	status := c.Query("status")

	// Get driver ID
	var driverID int64
	err := database.DB.QueryRow("SELECT id FROM drivers WHERE user_id = $1", userID).Scan(&driverID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver profile not found"})
		return
	}

	query := "SELECT * FROM orders WHERE driver_id = $1"
	args := []interface{}{driverID}

	if status != "" {
		query += " AND status = $2"
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
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

	c.JSON(http.StatusOK, orders)
}

// DriverStatistics represents driver statistics
type DriverStatistics struct {
	TotalOrders      int     `json:"total_orders"`
	CompletedOrders  int     `json:"completed_orders"`
	TotalEarnings    float64 `json:"total_earnings"`
	CurrentBalance   float64 `json:"current_balance"`
	AverageRating    float64 `json:"average_rating"`
	TotalRatings     int     `json:"total_ratings"`
}

// GetDriverStatistics godoc
// @Summary Get driver statistics
// @Description Get driver's performance statistics
// @Tags Driver
// @Security BearerAuth
// @Produce json
// @Param period query string false "Period (daily/monthly/yearly)"
// @Success 200 {object} DriverStatistics
// @Router /driver/statistics [get]
func (h *DriverHandler) GetDriverStatistics(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	period := c.Query("period")

	// Get driver info
	var driver models.Driver
	err := database.DB.QueryRow(`
		SELECT id, balance, rating, total_ratings FROM drivers WHERE user_id = $1
	`, userID).Scan(&driver.ID, &driver.Balance, &driver.Rating, &driver.TotalRatings)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver profile not found"})
		return
	}

	// Build query based on period
	var timeFilter string
	switch period {
	case "daily":
		timeFilter = "AND DATE(o.created_at) = CURRENT_DATE"
	case "monthly":
		timeFilter = "AND DATE_TRUNC('month', o.created_at) = DATE_TRUNC('month', CURRENT_DATE)"
	case "yearly":
		timeFilter = "AND DATE_TRUNC('year', o.created_at) = DATE_TRUNC('year', CURRENT_DATE)"
	default:
		timeFilter = ""
	}

	var stats DriverStatistics
	err = database.DB.QueryRow(fmt.Sprintf(`
		SELECT 
			COUNT(o.id) as total_orders,
			COUNT(CASE WHEN o.status = 'completed' THEN 1 END) as completed_orders,
			COALESCE(SUM(CASE WHEN o.status = 'completed' THEN o.service_fee ELSE 0 END), 0) as total_earnings
		FROM orders o
		WHERE o.driver_id = $1 %s
	`, timeFilter), driver.ID).Scan(&stats.TotalOrders, &stats.CompletedOrders, &stats.TotalEarnings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch statistics"})
		return
	}

	stats.CurrentBalance = driver.Balance
	stats.AverageRating = driver.Rating
	stats.TotalRatings = driver.TotalRatings

	c.JSON(http.StatusOK, stats)
}
