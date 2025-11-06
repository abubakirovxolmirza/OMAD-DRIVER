package handlers

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"taxi-service/internal/database"
	"taxi-service/internal/middleware"
	"taxi-service/internal/models"
)

// RatingHandler handles rating-related endpoints
type RatingHandler struct{}

// NewRatingHandler creates a new rating handler
func NewRatingHandler() *RatingHandler {
	return &RatingHandler{}
}

// CreateRatingRequest represents rating creation request
type CreateRatingRequest struct {
	OrderID int64  `json:"order_id" validate:"required"`
	Rating  int    `json:"rating" validate:"required,min=1,max=5"`
	Comment string `json:"comment"`
}

// CreateRating godoc
// @Summary Rate a driver
// @Description Rate a driver after order completion
// @Tags Rating
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateRatingRequest true "Rating details"
// @Success 201 {object} models.Rating
// @Router /ratings [post]

func (h *RatingHandler) CreateRating(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)

	var req CreateRatingRequest
	if err := parseAndValidateJSON(c, &req); err != nil {
		return err
	}

	// Verify order belongs to user and is completed
	var order models.Order
	err := database.DB.QueryRow(`
		SELECT id, user_id, driver_id, status
		FROM orders WHERE id = $1 AND user_id = $2
	`, req.OrderID, userID).Scan(&order.ID, &order.UserID, &order.DriverID, &order.Status)
	
	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order not found"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	if order.Status != models.OrderStatusCompleted {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Can only rate completed orders"})
	}

	if order.DriverID == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Order has no driver assigned"})
	}

	// Check if already rated
	var existingID int64
	_ = database.DB.QueryRow("SELECT id FROM ratings WHERE order_id = $1", req.OrderID).Scan(&existingID)
	if existingID > 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Order already rated"})
	}

	// Begin transaction
	tx, err := database.DB.Begin()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}
	defer tx.Rollback()

	// Create rating
	var rating models.Rating
	var comment *string
	if req.Comment != "" {
		comment = &req.Comment
	}
	
	err = tx.QueryRow(`
		INSERT INTO ratings (order_id, user_id, driver_id, rating, comment)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, order_id, user_id, driver_id, rating, comment, created_at
	`, req.OrderID, userID, *order.DriverID, req.Rating, comment).Scan(
		&rating.ID, &rating.OrderID, &rating.UserID, &rating.DriverID, 
		&rating.Rating, &rating.Comment, &rating.CreatedAt,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create rating"})
	}

	// Update driver rating
	var avgRating float64
	var totalRatings int
	err = tx.QueryRow(`
		SELECT AVG(rating), COUNT(*) FROM ratings WHERE driver_id = $1
	`, *order.DriverID).Scan(&avgRating, &totalRatings)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to calculate rating"})
	}

	_, err = tx.Exec(`
		UPDATE drivers SET rating = $1, total_ratings = $2 WHERE id = $3
	`, avgRating, totalRatings, *order.DriverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update driver rating"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to commit transaction"})
	}

	return c.Status(fiber.StatusCreated).JSON(rating)
}

// GetDriverRatings godoc
// @Summary Get driver ratings
// @Description Get all ratings for a specific driver
// @Tags Rating
// @Produce json
// @Param driver_id path int true "Driver ID"
// @Success 200 {array} models.Rating
// @Router /ratings/driver/{driver_id} [get]

func (h *RatingHandler) GetDriverRatings(c *fiber.Ctx) error {
	driverID := c.Param("driver_id")

	rows, err := database.DB.Query(`
		SELECT id, order_id, user_id, driver_id, rating, comment, created_at
		FROM ratings WHERE driver_id = $1 ORDER BY created_at DESC
	`, driverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch ratings"})
	}
	defer rows.Close()

	ratings := []models.Rating{}
	for rows.Next() {
		var rating models.Rating
		err := rows.Scan(
			&rating.ID, &rating.OrderID, &rating.UserID, &rating.DriverID,
			&rating.Rating, &rating.Comment, &rating.CreatedAt,
		)
		if err != nil {
			continue
		}
		ratings = append(ratings, rating)
	}

	return c.Status(fiber.StatusOK).JSON(ratings)
}

// NotificationHandler handles notification endpoints
type NotificationHandler struct{}

// NewNotificationHandler creates new notification handler
func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{}
}

// GetMyNotifications godoc
// @Summary Get user notifications
// @Description Get all notifications for the current user
// @Tags Notifications
// @Security BearerAuth
// @Produce json
// @Param unread query bool false "Filter unread only"
// @Success 200 {array} models.Notification
// @Router /notifications [get]

func (h *NotificationHandler) GetMyNotifications(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)
	unreadOnly := c.Query("unread") == "true"

	query := "SELECT * FROM notifications WHERE user_id = $1"
	if unreadOnly {
		query += " AND is_read = false"
	}
	query += " ORDER BY created_at DESC"

	rows, err := database.DB.Query(query, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch notifications"})
	}
	defer rows.Close()

	notifications := []models.Notification{}
	for rows.Next() {
		var notif models.Notification
		err := rows.Scan(
			&notif.ID, &notif.UserID, &notif.Title, &notif.Message,
			&notif.Type, &notif.RelatedID, &notif.IsRead, &notif.CreatedAt,
		)
		if err != nil {
			continue
		}
		notifications = append(notifications, notif)
	}

	return c.Status(fiber.StatusOK).JSON(notifications)
}

// MarkNotificationRead godoc
// @Summary Mark notification as read
// @Description Mark a notification as read
// @Tags Notifications
// @Security BearerAuth
// @Produce json
// @Param id path int true "Notification ID"
// @Success 200 {object} map[string]string
// @Router /notifications/{id}/read [post]

func (h *NotificationHandler) MarkNotificationRead(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)
	notifID := c.Param("id")

	_, err := database.DB.Exec(`
		UPDATE notifications SET is_read = true WHERE id = $1 AND user_id = $2
	`, notifID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update notification"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Notification marked as read"})
}

// RegionHandler handles region and district endpoints
type RegionHandler struct{}

// NewRegionHandler creates new region handler
func NewRegionHandler() *RegionHandler {
	return &RegionHandler{}
}

// CreateRegionRequest represents region creation request
type CreateRegionRequest struct {
	NameUzLat string `json:"name_uz_lat" validate:"required"`
	NameUzCyr string `json:"name_uz_cyr" validate:"required"`
	NameRu    string `json:"name_ru" validate:"required"`
}

// UpdateRegionRequest represents region update request
type UpdateRegionRequest struct {
	NameUzLat string `json:"name_uz_lat"`
	NameUzCyr string `json:"name_uz_cyr"`
	NameRu    string `json:"name_ru"`
}

// GetRegions godoc
// @Summary Get all regions
// @Description Get list of all regions
// @Tags Regions
// @Produce json
// @Success 200 {array} models.Region
// @Router /regions [get]

func (h *RegionHandler) GetRegions(c *fiber.Ctx) error {
	rows, err := database.DB.Query("SELECT * FROM regions ORDER BY name_uz_lat")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch regions"})
	}
	defer rows.Close()

	regions := []models.Region{}
	for rows.Next() {
		var region models.Region
		err := rows.Scan(&region.ID, &region.NameUzLat, &region.NameUzCyr, &region.NameRu, &region.CreatedAt)
		if err != nil {
			continue
		}
		regions = append(regions, region)
	}

	return c.Status(fiber.StatusOK).JSON(regions)
}

// GetRegion godoc
// @Summary Get region by ID
// @Description Get a specific region by ID
// @Tags Regions
// @Produce json
// @Param id path int true "Region ID"
// @Success 200 {object} models.Region
// @Router /regions/{id} [get]

func (h *RegionHandler) GetRegion(c *fiber.Ctx) error {
	regionID := c.Param("id")

	var region models.Region
	err := database.DB.QueryRow("SELECT * FROM regions WHERE id = $1", regionID).Scan(
		&region.ID, &region.NameUzLat, &region.NameUzCyr, &region.NameRu, &region.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Region not found"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	return c.Status(fiber.StatusOK).JSON(region)
}

// CreateRegion godoc
// @Summary Create a new region
// @Description Create a new region (Admin only)
// @Tags Regions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateRegionRequest true "Region details"
// @Success 201 {object} models.Region
// @Router /regions [post]

func (h *RegionHandler) CreateRegion(c *fiber.Ctx) error {
	var req CreateRegionRequest
	if err := parseAndValidateJSON(c, &req); err != nil {
		return err
	}

	var region models.Region
	err := database.DB.QueryRow(`
		INSERT INTO regions (name_uz_lat, name_uz_cyr, name_ru)
		VALUES ($1, $2, $3)
		RETURNING id, name_uz_lat, name_uz_cyr, name_ru, created_at
	`, req.NameUzLat, req.NameUzCyr, req.NameRu).Scan(
		&region.ID, &region.NameUzLat, &region.NameUzCyr, &region.NameRu, &region.CreatedAt,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create region"})
	}

	return c.Status(fiber.StatusCreated).JSON(region)
}

// UpdateRegion godoc
// @Summary Update a region
// @Description Update an existing region (Admin only)
// @Tags Regions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Region ID"
// @Param request body UpdateRegionRequest true "Region details"
// @Success 200 {object} models.Region
// @Router /regions/{id} [put]
func (h *RegionHandler) UpdateRegion(c *fiber.Ctx) error {
	regionID := c.Param("id")

	var req UpdateRegionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	updates := make([]string, 0, 3)
	args := make([]interface{}, 0, 4)

	if req.NameUzLat != "" {
		args = append(args, req.NameUzLat)
		updates = append(updates, fmt.Sprintf("name_uz_lat = $%d", len(args)))
	}
	if req.NameUzCyr != "" {
		args = append(args, req.NameUzCyr)
		updates = append(updates, fmt.Sprintf("name_uz_cyr = $%d", len(args)))
	}
	if req.NameRu != "" {
		args = append(args, req.NameRu)
		updates = append(updates, fmt.Sprintf("name_ru = $%d", len(args)))
	}

	if len(updates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No fields to update"})
	}

	args = append(args, regionID)
	query := fmt.Sprintf("UPDATE regions SET %s WHERE id = $%d", strings.Join(updates, ", "), len(args))
	result, err := database.DB.Exec(query, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update region"})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Region not found"})
	}

	var region models.Region
	if err := database.DB.QueryRow("SELECT * FROM regions WHERE id = $1", regionID).Scan(
		&region.ID, &region.NameUzLat, &region.NameUzCyr, &region.NameRu, &region.CreatedAt,
	); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch updated region"})
	}

	return c.Status(fiber.StatusOK).JSON(region)
}

// DeleteRegion godoc
// @Summary Delete a region
// @Description Delete a region (Admin only)
// @Tags Regions
// @Security BearerAuth
// @Produce json
// @Param id path int true "Region ID"
// @Success 200 {object} map[string]string
// @Router /regions/{id} [delete]

func (h *RegionHandler) DeleteRegion(c *fiber.Ctx) error {
	regionID := c.Param("id")

	result, err := database.DB.Exec("DELETE FROM regions WHERE id = $1", regionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete region"})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Region not found"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Region deleted successfully"})
}

// GetDistricts godoc
// @Summary Get districts by region
// @Description Get all districts for a specific region
// @Tags Regions
// @Produce json
// @Param id path int true "Region ID"
// @Success 200 {array} models.District
// @Router /regions/{id}/districts [get]

func (h *RegionHandler) GetDistricts(c *fiber.Ctx) error {
	regionID := c.Param("id")

	rows, err := database.DB.Query("SELECT * FROM districts WHERE region_id = $1 ORDER BY name_uz_lat", regionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch districts"})
	}
	defer rows.Close()

	districts := []models.District{}
	for rows.Next() {
		var district models.District
		err := rows.Scan(&district.ID, &district.RegionID, &district.NameUzLat, &district.NameUzCyr, &district.NameRu, &district.CreatedAt)
		if err != nil {
			continue
		}
		districts = append(districts, district)
	}

	return c.Status(fiber.StatusOK).JSON(districts)
}

// CreateDistrictRequest represents district creation request
type CreateDistrictRequest struct {
	RegionID  int64  `json:"region_id" validate:"required"`
	NameUzLat string `json:"name_uz_lat" validate:"required"`
	NameUzCyr string `json:"name_uz_cyr" validate:"required"`
	NameRu    string `json:"name_ru" validate:"required"`
}

// UpdateDistrictRequest represents district update request
type UpdateDistrictRequest struct {
	NameUzLat string `json:"name_uz_lat"`
	NameUzCyr string `json:"name_uz_cyr"`
	NameRu    string `json:"name_ru"`
}

// GetDistrict godoc
// @Summary Get district by ID
// @Description Get a specific district by ID
// @Tags Regions
// @Produce json
// @Param id path int true "District ID"
// @Success 200 {object} models.District
// @Router /districts/{id} [get]

func (h *RegionHandler) GetDistrict(c *fiber.Ctx) error {
	districtID := c.Param("id")

	var district models.District
	err := database.DB.QueryRow("SELECT * FROM districts WHERE id = $1", districtID).Scan(
		&district.ID, &district.RegionID, &district.NameUzLat, &district.NameUzCyr, &district.NameRu, &district.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "District not found"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	return c.Status(fiber.StatusOK).JSON(district)
}

// CreateDistrict godoc
// @Summary Create a new district
// @Description Create a new district (Admin only)
// @Tags Regions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateDistrictRequest true "District details"
// @Success 201 {object} models.District
// @Router /districts [post]

func (h *RegionHandler) CreateDistrict(c *fiber.Ctx) error {
	var req CreateDistrictRequest
	if err := parseAndValidateJSON(c, &req); err != nil {
		return err
	}

	var district models.District
	err := database.DB.QueryRow(`
		INSERT INTO districts (region_id, name_uz_lat, name_uz_cyr, name_ru)
		VALUES ($1, $2, $3, $4)
		RETURNING id, region_id, name_uz_lat, name_uz_cyr, name_ru, created_at
	`, req.RegionID, req.NameUzLat, req.NameUzCyr, req.NameRu).Scan(
		&district.ID, &district.RegionID, &district.NameUzLat, &district.NameUzCyr, &district.NameRu, &district.CreatedAt,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create district"})
	}

	return c.Status(fiber.StatusCreated).JSON(district)
}

// UpdateDistrict godoc
// @Summary Update a district
// @Description Update an existing district (Admin only)
// @Tags Regions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "District ID"
// @Param request body UpdateDistrictRequest true "District details"
// @Success 200 {object} models.District
// @Router /districts/{id} [put]
func (h *RegionHandler) UpdateDistrict(c *fiber.Ctx) error {
	districtID := c.Param("id")

	var req UpdateDistrictRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	updates := make([]string, 0, 3)
	args := make([]interface{}, 0, 4)

	if req.NameUzLat != "" {
		args = append(args, req.NameUzLat)
		updates = append(updates, fmt.Sprintf("name_uz_lat = $%d", len(args)))
	}
	if req.NameUzCyr != "" {
		args = append(args, req.NameUzCyr)
		updates = append(updates, fmt.Sprintf("name_uz_cyr = $%d", len(args)))
	}
	if req.NameRu != "" {
		args = append(args, req.NameRu)
		updates = append(updates, fmt.Sprintf("name_ru = $%d", len(args)))
	}

	if len(updates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No fields to update"})
	}

	args = append(args, districtID)
	query := fmt.Sprintf("UPDATE districts SET %s WHERE id = $%d", strings.Join(updates, ", "), len(args))
	result, err := database.DB.Exec(query, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update district"})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "District not found"})
	}

	var district models.District
	if err := database.DB.QueryRow("SELECT * FROM districts WHERE id = $1", districtID).Scan(
		&district.ID, &district.RegionID, &district.NameUzLat, &district.NameUzCyr, &district.NameRu, &district.CreatedAt,
	); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch updated district"})
	}

	return c.Status(fiber.StatusOK).JSON(district)
}

// DeleteDistrict godoc
// @Summary Delete a district
// @Description Delete a district (Admin only)
// @Tags Regions
// @Security BearerAuth
// @Produce json
// @Param id path int true "District ID"
// @Success 200 {object} map[string]string
// @Router /districts/{id} [delete]

func (h *RegionHandler) DeleteDistrict(c *fiber.Ctx) error {
	districtID := c.Param("id")

	result, err := database.DB.Exec("DELETE FROM districts WHERE id = $1", districtID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete district"})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "District not found"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "District deleted successfully"})
}

// FeedbackHandler handles feedback endpoints
type FeedbackHandler struct{}

// NewFeedbackHandler creates new feedback handler
func NewFeedbackHandler() *FeedbackHandler {
	return &FeedbackHandler{}
}

// SubmitFeedbackRequest represents feedback submission
type SubmitFeedbackRequest struct {
	Message string `json:"message" validate:"required"`
}

// SubmitFeedback godoc
// @Summary Submit feedback
// @Description Submit feedback or suggestion
// @Tags Feedback
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body SubmitFeedbackRequest true "Feedback message"
// @Success 201 {object} models.Feedback
// @Router /feedback [post]

func (h *FeedbackHandler) SubmitFeedback(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)

	var req SubmitFeedbackRequest
	if err := parseAndValidateJSON(c, &req); err != nil {
		return err
	}

	var feedback models.Feedback
	err := database.DB.QueryRow(`
		INSERT INTO feedback (user_id, message)
		VALUES ($1, $2)
		RETURNING id, user_id, message, created_at
	`, userID, req.Message).Scan(&feedback.ID, &feedback.UserID, &feedback.Message, &feedback.CreatedAt)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to submit feedback"})
	}

	// TODO: Send to Telegram admin group

	return c.Status(fiber.StatusCreated).JSON(feedback)
}
