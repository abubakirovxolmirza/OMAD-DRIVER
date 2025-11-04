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
)

// OrderHandler handles order-related endpoints
type OrderHandler struct {
	cfg *config.Config
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(cfg *config.Config) *OrderHandler {
	return &OrderHandler{cfg: cfg}
}

// CreateTaxiOrderRequest represents taxi order creation request
type CreateTaxiOrderRequest struct {
	CustomerName    string    `json:"customer_name" binding:"required"`
	CustomerPhone   string    `json:"customer_phone" binding:"required"`
	FromRegionID    int64     `json:"from_region_id" binding:"required"`
	FromDistrictID  int64     `json:"from_district_id" binding:"required"`
	FromLatitude    *float64  `json:"from_latitude"`
	FromLongitude   *float64  `json:"from_longitude"`
	FromAddress     *string   `json:"from_address"`
	ToRegionID      int64     `json:"to_region_id" binding:"required"`
	ToDistrictID    int64     `json:"to_district_id" binding:"required"`
	ToLatitude      *float64  `json:"to_latitude"`
	ToLongitude     *float64  `json:"to_longitude"`
	ToAddress       *string   `json:"to_address"`
	PassengerCount  int       `json:"passenger_count" binding:"required,min=1,max=4"`
	ScheduledDate   string    `json:"scheduled_date" binding:"required"` // DD.MM.YYYY
	TimeRangeStart  string    `json:"time_range_start" binding:"required"`
	TimeRangeEnd    string    `json:"time_range_end" binding:"required"`
	Notes           string    `json:"notes"`
}

// CreateDeliveryOrderRequest represents delivery order creation request
type CreateDeliveryOrderRequest struct {
	CustomerName    string   `json:"customer_name" binding:"required"`
	CustomerPhone   string   `json:"customer_phone" binding:"required"`
	RecipientPhone  string   `json:"recipient_phone" binding:"required"`
	FromRegionID    int64    `json:"from_region_id" binding:"required"`
	FromDistrictID  int64    `json:"from_district_id" binding:"required"`
	FromLatitude    *float64 `json:"from_latitude"`
	FromLongitude   *float64 `json:"from_longitude"`
	FromAddress     *string  `json:"from_address"`
	ToRegionID      int64    `json:"to_region_id" binding:"required"`
	ToDistrictID    int64    `json:"to_district_id" binding:"required"`
	ToLatitude      *float64 `json:"to_latitude"`
	ToLongitude     *float64 `json:"to_longitude"`
	ToAddress       *string  `json:"to_address"`
	DeliveryType    string   `json:"delivery_type" binding:"required"`
	ScheduledDate   string   `json:"scheduled_date" binding:"required"` // DD.MM.YYYY
	TimeRangeStart  string   `json:"time_range_start" binding:"required"`
	TimeRangeEnd    string   `json:"time_range_end" binding:"required"`
	Notes           string   `json:"notes"`
}

// CreateTaxiOrder godoc
// @Summary Create a taxi order
// @Description Create a new taxi order with automatic price calculation
// @Tags Orders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateTaxiOrderRequest true "Taxi order details"
// @Success 201 {object} models.Order
// @Failure 400 {object} map[string]string
// @Router /orders/taxi [post]
func (h *OrderHandler) CreateTaxiOrder(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req CreateTaxiOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse date
	scheduledDate, err := time.Parse("02.01.2006", req.ScheduledDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, use DD.MM.YYYY"})
		return
	}

	// Validate regions are different
	if req.FromRegionID == req.ToRegionID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "From and To regions must be different"})
		return
	}

	// Calculate price
	price, serviceFee, discount, finalPriceCalc, err := h.calculateTaxiPrice(req.FromRegionID, req.ToRegionID, req.PassengerCount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create order
	var order models.Order
	acceptDeadline := time.Now().Add(5 * time.Minute)
	
	var notes *string
	if req.Notes != "" {
		notes = &req.Notes
	}
	
	err = database.DB.QueryRow(`
		INSERT INTO orders (
			user_id, order_type, status, customer_name, customer_phone,
			from_region_id, from_district_id, from_latitude, from_longitude, from_address,
			to_region_id, to_district_id, to_latitude, to_longitude, to_address,
			passenger_count, scheduled_date, time_range_start, time_range_end,
			price, service_fee, discount_percentage, final_price, notes, accept_deadline
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25)
		RETURNING id, created_at, updated_at
	`, userID, models.OrderTypeTaxi, models.OrderStatusPending,
		req.CustomerName, req.CustomerPhone, 
		req.FromRegionID, req.FromDistrictID, req.FromLatitude, req.FromLongitude, req.FromAddress,
		req.ToRegionID, req.ToDistrictID, req.ToLatitude, req.ToLongitude, req.ToAddress,
		req.PassengerCount, scheduledDate,
		req.TimeRangeStart, req.TimeRangeEnd, price, serviceFee, discount, finalPriceCalc,
		notes, acceptDeadline,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	// Populate order details for response
	order.UserID = userID
	order.OrderType = models.OrderTypeTaxi
	order.Status = models.OrderStatusPending
	order.CustomerName = req.CustomerName
	order.CustomerPhone = req.CustomerPhone
	order.FromRegionID = req.FromRegionID
	order.FromDistrictID = req.FromDistrictID
	order.FromLatitude = req.FromLatitude
	order.FromLongitude = req.FromLongitude
	order.FromAddress = req.FromAddress
	order.ToRegionID = req.ToRegionID
	order.ToDistrictID = req.ToDistrictID
	order.ToLatitude = req.ToLatitude
	order.ToLongitude = req.ToLongitude
	order.ToAddress = req.ToAddress
	passengerCount := int64(req.PassengerCount)
	order.PassengerCount = &passengerCount
	order.ScheduledDate = scheduledDate
	order.TimeRangeStart = req.TimeRangeStart
	order.TimeRangeEnd = req.TimeRangeEnd
	order.Price = price
	order.ServiceFee = serviceFee
	order.DiscountPercentage = discount
	order.FinalPrice = finalPriceCalc
	if req.Notes != "" {
		order.Notes = &req.Notes
	}

	// TODO: Send notification to all drivers
	go h.notifyDriversNewOrder(order.ID, models.OrderTypeTaxi)

	c.JSON(http.StatusCreated, order)
}

// CreateDeliveryOrder godoc
// @Summary Create a delivery order
// @Description Create a new delivery order with automatic price calculation
// @Tags Orders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateDeliveryOrderRequest true "Delivery order details"
// @Success 201 {object} models.Order
// @Failure 400 {object} map[string]string
// @Router /orders/delivery [post]
func (h *OrderHandler) CreateDeliveryOrder(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req CreateDeliveryOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse date
	scheduledDate, err := time.Parse("02.01.2006", req.ScheduledDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, use DD.MM.YYYY"})
		return
	}

	// Validate regions are different
	if req.FromRegionID == req.ToRegionID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "From and To regions must be different"})
		return
	}

	// Calculate price (same as taxi base price)
	price, serviceFee, _, finalPrice, err := h.calculateTaxiPrice(req.FromRegionID, req.ToRegionID, 1)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create order
	var order models.Order
	acceptDeadline := time.Now().Add(5 * time.Minute)
	
	var notes *string
	if req.Notes != "" {
		notes = &req.Notes
	}
	
	err = database.DB.QueryRow(`
		INSERT INTO orders (
			user_id, order_type, status, customer_name, customer_phone, recipient_phone,
			from_region_id, from_district_id, from_latitude, from_longitude, from_address,
			to_region_id, to_district_id, to_latitude, to_longitude, to_address,
			delivery_type, scheduled_date, time_range_start, time_range_end,
			price, service_fee, discount_percentage, final_price, notes, accept_deadline
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26)
		RETURNING id, created_at, updated_at
	`, userID, models.OrderTypeDelivery, models.OrderStatusPending,
		req.CustomerName, req.CustomerPhone, req.RecipientPhone,
		req.FromRegionID, req.FromDistrictID, req.FromLatitude, req.FromLongitude, req.FromAddress,
		req.ToRegionID, req.ToDistrictID, req.ToLatitude, req.ToLongitude, req.ToAddress,
		req.DeliveryType, scheduledDate, req.TimeRangeStart, req.TimeRangeEnd,
		price, serviceFee, 0, finalPrice,
		notes, acceptDeadline,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	// TODO: Send notification to all drivers
	go h.notifyDriversNewOrder(order.ID, models.OrderTypeDelivery)

	c.JSON(http.StatusCreated, order)
}

// GetMyOrders godoc
// @Summary Get user's orders
// @Description Get all orders created by the current user
// @Tags Orders
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by status"
// @Param type query string false "Filter by type (taxi/delivery)"
// @Success 200 {array} models.Order
// @Router /orders/my [get]
func (h *OrderHandler) GetMyOrders(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	status := c.Query("status")
	orderType := c.Query("type")

	query := "SELECT * FROM orders WHERE user_id = $1"
	args := []interface{}{userID}

	if status != "" {
		query += " AND status = $2"
		args = append(args, status)
	}
	if orderType != "" {
		if status != "" {
			query += " AND order_type = $3"
		} else {
			query += " AND order_type = $2"
		}
		args = append(args, orderType)
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

// GetOrderByID godoc
// @Summary Get order details
// @Description Get detailed information about a specific order
// @Tags Orders
// @Security BearerAuth
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} models.Order
// @Router /orders/{id} [get]
func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	userRole, _ := middleware.GetUserRole(c)
	orderID := c.Param("id")

	var order models.Order
	query := "SELECT * FROM orders WHERE id = $1"
	
	// Regular users can only see their own orders
	if userRole == models.RoleUser {
		query += " AND user_id = $2"
		err := database.DB.QueryRow(query, orderID, userID).Scan(
			&order.ID, &order.UserID, &order.DriverID, &order.OrderType, &order.Status,
			&order.CustomerName, &order.CustomerPhone, &order.RecipientPhone,
			&order.FromRegionID, &order.FromDistrictID, &order.ToRegionID, &order.ToDistrictID,
			&order.PassengerCount, &order.DeliveryType, &order.ScheduledDate,
			&order.TimeRangeStart, &order.TimeRangeEnd, &order.Price, &order.ServiceFee,
			&order.DiscountPercentage, &order.FinalPrice, &order.Notes, &order.CancellationReason,
			&order.AcceptedAt, &order.AcceptDeadline, &order.CompletedAt, &order.CancelledAt,
			&order.CreatedAt, &order.UpdatedAt,
		)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
	} else {
		// Drivers and admins can see any order
		err := database.DB.QueryRow(query, orderID).Scan(
			&order.ID, &order.UserID, &order.DriverID, &order.OrderType, &order.Status,
			&order.CustomerName, &order.CustomerPhone, &order.RecipientPhone,
			&order.FromRegionID, &order.FromDistrictID, &order.ToRegionID, &order.ToDistrictID,
			&order.PassengerCount, &order.DeliveryType, &order.ScheduledDate,
			&order.TimeRangeStart, &order.TimeRangeEnd, &order.Price, &order.ServiceFee,
			&order.DiscountPercentage, &order.FinalPrice, &order.Notes, &order.CancellationReason,
			&order.AcceptedAt, &order.AcceptDeadline, &order.CompletedAt, &order.CancelledAt,
			&order.CreatedAt, &order.UpdatedAt,
		)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
	}

	c.JSON(http.StatusOK, order)
}

// CancelOrder godoc
// @Summary Cancel an order
// @Description Cancel a pending or accepted order
// @Tags Orders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Param request body map[string]string true "Cancellation reason"
// @Success 200 {object} map[string]string
// @Router /orders/{id}/cancel [post]
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	orderID := c.Param("id")

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get order details
	var order models.Order
	err := database.DB.QueryRow(`
		SELECT id, user_id, driver_id, status, service_fee
		FROM orders WHERE id = $1 AND user_id = $2
	`, orderID, userID).Scan(&order.ID, &order.UserID, &order.DriverID, &order.Status, &order.ServiceFee)
	
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Can only cancel pending or accepted orders
	if order.Status != models.OrderStatusPending && order.Status != models.OrderStatusAccepted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot cancel order in current status"})
		return
	}

	// Update order status
	_, err = database.DB.Exec(`
		UPDATE orders SET status = $1, cancellation_reason = $2, cancelled_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`, models.OrderStatusCancelled, req.Reason, orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel order"})
		return
	}

	// Refund driver if order was accepted
	if order.DriverID != nil {
		_, err = database.DB.Exec(`
			UPDATE drivers SET balance = balance + $1 WHERE id = $2
		`, order.ServiceFee, *order.DriverID)
		if err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to refund driver: %v\n", err)
		}

		// Create transaction record
		database.DB.Exec(`
			INSERT INTO transactions (driver_id, order_id, amount, type, description)
			VALUES ($1, $2, $3, $4, $5)
		`, *order.DriverID, order.ID, order.ServiceFee, "credit", "Refund for cancelled order")
	}

	// TODO: Send notification to telegram group
	// TODO: Notify driver if assigned

	c.JSON(http.StatusOK, gin.H{"message": "Order cancelled successfully"})
}

// Helper functions

func (h *OrderHandler) calculateTaxiPrice(fromRegionID, toRegionID int64, passengerCount int) (float64, float64, float64, float64, error) {
	// Get pricing for route
	var pricing models.Pricing
	err := database.DB.QueryRow(`
		SELECT base_price, price_per_person, service_fee
		FROM pricing WHERE from_region_id = $1 AND to_region_id = $2
	`, fromRegionID, toRegionID).Scan(&pricing.BasePrice, &pricing.PricePerPerson, &pricing.ServiceFee)
	
	if err == sql.ErrNoRows {
		return 0, 0, 0, 0, fmt.Errorf("pricing not configured for this route")
	}
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("database error")
	}

	// Calculate base price
	basePrice := pricing.BasePrice + (pricing.PricePerPerson * float64(passengerCount))

	// Get discount
	var discount float64
	err = database.DB.QueryRow(`
		SELECT discount_percentage FROM discounts WHERE passenger_count = $1
	`, passengerCount).Scan(&discount)
	if err != nil {
		discount = 0
	}

	// Apply discount
	discountAmount := basePrice * (discount / 100)
	priceAfterDiscount := basePrice - discountAmount

	// Calculate service fee
	serviceFee := priceAfterDiscount * (pricing.ServiceFee / 100)

	// Final price
	finalPrice := priceAfterDiscount + serviceFee

	return basePrice, serviceFee, discount, finalPrice, nil
}

func (h *OrderHandler) notifyDriversNewOrder(orderID int64, orderType models.OrderType) {
	// Get all active drivers
	rows, err := database.DB.Query(`
		SELECT u.id FROM users u
		INNER JOIN drivers d ON u.id = d.user_id
		WHERE u.role = $1 AND d.status = 'approved' AND d.is_active = true AND u.is_blocked = false
	`, models.RoleDriver)
	if err != nil {
		return
	}
	defer rows.Close()

	title := "New Order Available"
	message := fmt.Sprintf("A new %s order is available. Check your orders page.", orderType)

	for rows.Next() {
		var driverUserID int64
		if rows.Scan(&driverUserID) == nil {
			// Create notification
			database.DB.Exec(`
				INSERT INTO notifications (user_id, title, message, type, related_id)
				VALUES ($1, $2, $3, $4, $5)
			`, driverUserID, title, message, "new_order", orderID)
		}
	}
}
