package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
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
	OrderID int64  `json:"order_id" binding:"required"`
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
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
func (h *RatingHandler) CreateRating(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req CreateRatingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify order belongs to user and is completed
	var order models.Order
	err := database.DB.QueryRow(`
		SELECT id, user_id, driver_id, status
		FROM orders WHERE id = $1 AND user_id = $2
	`, req.OrderID, userID).Scan(&order.ID, &order.UserID, &order.DriverID, &order.Status)
	
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if order.Status != models.OrderStatusCompleted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can only rate completed orders"})
		return
	}

	if !order.DriverID.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order has no driver assigned"})
		return
	}

	// Check if already rated
	var existingID int64
	database.DB.QueryRow("SELECT id FROM ratings WHERE order_id = $1", req.OrderID).Scan(&existingID)
	if existingID > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Order already rated"})
		return
	}

	// Begin transaction
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	// Create rating
	var rating models.Rating
	err = tx.QueryRow(`
		INSERT INTO ratings (order_id, user_id, driver_id, rating, comment)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, order_id, user_id, driver_id, rating, comment, created_at
	`, req.OrderID, userID, order.DriverID.Int64, req.Rating, 
		sql.NullString{String: req.Comment, Valid: req.Comment != ""}).Scan(
		&rating.ID, &rating.OrderID, &rating.UserID, &rating.DriverID, 
		&rating.Rating, &rating.Comment, &rating.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create rating"})
		return
	}

	// Update driver rating
	var avgRating float64
	var totalRatings int
	err = tx.QueryRow(`
		SELECT AVG(rating), COUNT(*) FROM ratings WHERE driver_id = $1
	`, order.DriverID.Int64).Scan(&avgRating, &totalRatings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate rating"})
		return
	}

	_, err = tx.Exec(`
		UPDATE drivers SET rating = $1, total_ratings = $2 WHERE id = $3
	`, avgRating, totalRatings, order.DriverID.Int64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update driver rating"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusCreated, rating)
}

// GetDriverRatings godoc
// @Summary Get driver ratings
// @Description Get all ratings for a specific driver
// @Tags Rating
// @Produce json
// @Param driver_id path int true "Driver ID"
// @Success 200 {array} models.Rating
// @Router /ratings/driver/{driver_id} [get]
func (h *RatingHandler) GetDriverRatings(c *gin.Context) {
	driverID := c.Param("driver_id")

	rows, err := database.DB.Query(`
		SELECT id, order_id, user_id, driver_id, rating, comment, created_at
		FROM ratings WHERE driver_id = $1 ORDER BY created_at DESC
	`, driverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch ratings"})
		return
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

	c.JSON(http.StatusOK, ratings)
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
func (h *NotificationHandler) GetMyNotifications(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	unreadOnly := c.Query("unread") == "true"

	query := "SELECT * FROM notifications WHERE user_id = $1"
	if unreadOnly {
		query += " AND is_read = false"
	}
	query += " ORDER BY created_at DESC"

	rows, err := database.DB.Query(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
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

	c.JSON(http.StatusOK, notifications)
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
func (h *NotificationHandler) MarkNotificationRead(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	notifID := c.Param("id")

	_, err := database.DB.Exec(`
		UPDATE notifications SET is_read = true WHERE id = $1 AND user_id = $2
	`, notifID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// RegionHandler handles region and district endpoints
type RegionHandler struct{}

// NewRegionHandler creates new region handler
func NewRegionHandler() *RegionHandler {
	return &RegionHandler{}
}

// GetRegions godoc
// @Summary Get all regions
// @Description Get list of all regions
// @Tags Regions
// @Produce json
// @Success 200 {array} models.Region
// @Router /regions [get]
func (h *RegionHandler) GetRegions(c *gin.Context) {
	rows, err := database.DB.Query("SELECT * FROM regions ORDER BY name_uz_lat")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch regions"})
		return
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

	c.JSON(http.StatusOK, regions)
}

// GetDistricts godoc
// @Summary Get districts by region
// @Description Get all districts for a specific region
// @Tags Regions
// @Produce json
// @Param region_id path int true "Region ID"
// @Success 200 {array} models.District
// @Router /regions/{region_id}/districts [get]
func (h *RegionHandler) GetDistricts(c *gin.Context) {
	regionID := c.Param("region_id")

	rows, err := database.DB.Query("SELECT * FROM districts WHERE region_id = $1 ORDER BY name_uz_lat", regionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch districts"})
		return
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

	c.JSON(http.StatusOK, districts)
}

// FeedbackHandler handles feedback endpoints
type FeedbackHandler struct{}

// NewFeedbackHandler creates new feedback handler
func NewFeedbackHandler() *FeedbackHandler {
	return &FeedbackHandler{}
}

// SubmitFeedbackRequest represents feedback submission
type SubmitFeedbackRequest struct {
	Message string `json:"message" binding:"required"`
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
func (h *FeedbackHandler) SubmitFeedback(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req SubmitFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var feedback models.Feedback
	err := database.DB.QueryRow(`
		INSERT INTO feedback (user_id, message)
		VALUES ($1, $2)
		RETURNING id, user_id, message, created_at
	`, userID, req.Message).Scan(&feedback.ID, &feedback.UserID, &feedback.Message, &feedback.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit feedback"})
		return
	}

	// TODO: Send to Telegram admin group

	c.JSON(http.StatusCreated, feedback)
}
