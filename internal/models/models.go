package models

import (
	"database/sql"
	"time"
)

// UserRole represents user role in the system
type UserRole string

const (
	RoleUser       UserRole = "user"
	RoleDriver     UserRole = "driver"
	RoleAdmin      UserRole = "admin"
	RoleSuperAdmin UserRole = "superadmin"
)

// Language represents supported languages
type Language string

const (
	LangUzLatin    Language = "uz_latin"
	LangUzCyrillic Language = "uz_cyrillic"
	LangRussian    Language = "ru"
)

// User represents a user in the system
type User struct {
	ID           int64          `json:"id" db:"id"`
	PhoneNumber  string         `json:"phone_number" db:"phone_number"`
	Name         string         `json:"name" db:"name"`
	Password     string         `json:"-" db:"password"`
	Avatar       sql.NullString `json:"avatar" db:"avatar"`
	Role         UserRole       `json:"role" db:"role"`
	Language     Language       `json:"language" db:"language"`
	IsBlocked    bool           `json:"is_blocked" db:"is_blocked"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
}

// Driver represents additional driver information
type Driver struct {
	ID              int64          `json:"id" db:"id"`
	UserID          int64          `json:"user_id" db:"user_id"`
	FullName        string         `json:"full_name" db:"full_name"`
	CarModel        string         `json:"car_model" db:"car_model"`
	CarNumber       string         `json:"car_number" db:"car_number"`
	LicenseImage    sql.NullString `json:"license_image" db:"license_image"`
	Balance         float64        `json:"balance" db:"balance"`
	Rating          float64        `json:"rating" db:"rating"`
	TotalRatings    int            `json:"total_ratings" db:"total_ratings"`
	Status          string         `json:"status" db:"status"` // pending, approved, rejected
	IsActive        bool           `json:"is_active" db:"is_active"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
}

// Region represents a region/province
type Region struct {
	ID        int64     `json:"id" db:"id"`
	NameUzLat string    `json:"name_uz_lat" db:"name_uz_lat"`
	NameUzCyr string    `json:"name_uz_cyr" db:"name_uz_cyr"`
	NameRu    string    `json:"name_ru" db:"name_ru"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// District represents a district within a region
type District struct {
	ID        int64     `json:"id" db:"id"`
	RegionID  int64     `json:"region_id" db:"region_id"`
	NameUzLat string    `json:"name_uz_lat" db:"name_uz_lat"`
	NameUzCyr string    `json:"name_uz_cyr" db:"name_uz_cyr"`
	NameRu    string    `json:"name_ru" db:"name_ru"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// OrderType represents the type of order
type OrderType string

const (
	OrderTypeTaxi     OrderType = "taxi"
	OrderTypeDelivery OrderType = "delivery"
)

// OrderStatus represents order status
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusAccepted   OrderStatus = "accepted"
	OrderStatusInProgress OrderStatus = "in_progress"
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

// PassengerCount represents number of passengers
type PassengerCount int

const (
	OnePassenger    PassengerCount = 1
	TwoPassengers   PassengerCount = 2
	ThreePassengers PassengerCount = 3
	FullCar         PassengerCount = 4
)

// DeliveryType represents type of delivery item
type DeliveryType string

const (
	DeliveryDocument  DeliveryType = "document"
	DeliveryBox       DeliveryType = "box"
	DeliveryLuggage   DeliveryType = "luggage"
	DeliveryValuable  DeliveryType = "valuable"
	DeliveryOther     DeliveryType = "other"
)

// Order represents both taxi and delivery orders
type Order struct {
	ID                    int64           `json:"id" db:"id"`
	UserID                int64           `json:"user_id" db:"user_id"`
	DriverID              sql.NullInt64   `json:"driver_id" db:"driver_id"`
	OrderType             OrderType       `json:"order_type" db:"order_type"`
	Status                OrderStatus     `json:"status" db:"status"`
	
	// Customer info
	CustomerName          string          `json:"customer_name" db:"customer_name"`
	CustomerPhone         string          `json:"customer_phone" db:"customer_phone"`
	RecipientPhone        sql.NullString  `json:"recipient_phone" db:"recipient_phone"` // For delivery
	
	// Location info
	FromRegionID          int64           `json:"from_region_id" db:"from_region_id"`
	FromDistrictID        int64           `json:"from_district_id" db:"from_district_id"`
	ToRegionID            int64           `json:"to_region_id" db:"to_region_id"`
	ToDistrictID          int64           `json:"to_district_id" db:"to_district_id"`
	
	// Taxi specific
	PassengerCount        sql.NullInt64   `json:"passenger_count" db:"passenger_count"`
	
	// Delivery specific
	DeliveryType          sql.NullString  `json:"delivery_type" db:"delivery_type"`
	
	// Schedule
	ScheduledDate         time.Time       `json:"scheduled_date" db:"scheduled_date"`
	TimeRangeStart        string          `json:"time_range_start" db:"time_range_start"`
	TimeRangeEnd          string          `json:"time_range_end" db:"time_range_end"`
	
	// Pricing
	Price                 float64         `json:"price" db:"price"`
	ServiceFee            float64         `json:"service_fee" db:"service_fee"`
	DiscountPercentage    float64         `json:"discount_percentage" db:"discount_percentage"`
	FinalPrice            float64         `json:"final_price" db:"final_price"`
	
	// Additional info
	Notes                 sql.NullString  `json:"notes" db:"notes"`
	CancellationReason    sql.NullString  `json:"cancellation_reason" db:"cancellation_reason"`
	
	// Timing
	AcceptedAt            sql.NullTime    `json:"accepted_at" db:"accepted_at"`
	AcceptDeadline        sql.NullTime    `json:"accept_deadline" db:"accept_deadline"`
	CompletedAt           sql.NullTime    `json:"completed_at" db:"completed_at"`
	CancelledAt           sql.NullTime    `json:"cancelled_at" db:"cancelled_at"`
	CreatedAt             time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at" db:"updated_at"`
}

// Pricing represents pricing configuration between regions
type Pricing struct {
	ID             int64     `json:"id" db:"id"`
	FromRegionID   int64     `json:"from_region_id" db:"from_region_id"`
	ToRegionID     int64     `json:"to_region_id" db:"to_region_id"`
	BasePrice      float64   `json:"base_price" db:"base_price"`
	PricePerPerson float64   `json:"price_per_person" db:"price_per_person"`
	ServiceFee     float64   `json:"service_fee" db:"service_fee"` // Percentage
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Discount represents discount configuration
type Discount struct {
	ID                 int64     `json:"id" db:"id"`
	PassengerCount     int       `json:"passenger_count" db:"passenger_count"`
	DiscountPercentage float64   `json:"discount_percentage" db:"discount_percentage"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// Rating represents driver rating by user
type Rating struct {
	ID        int64          `json:"id" db:"id"`
	OrderID   int64          `json:"order_id" db:"order_id"`
	UserID    int64          `json:"user_id" db:"user_id"`
	DriverID  int64          `json:"driver_id" db:"driver_id"`
	Rating    int            `json:"rating" db:"rating"` // 1-5
	Comment   sql.NullString `json:"comment" db:"comment"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
}

// Notification represents a notification
type Notification struct {
	ID        int64          `json:"id" db:"id"`
	UserID    int64          `json:"user_id" db:"user_id"`
	Title     string         `json:"title" db:"title"`
	Message   string         `json:"message" db:"message"`
	Type      string         `json:"type" db:"type"` // order, rating, system, etc.
	RelatedID sql.NullInt64  `json:"related_id" db:"related_id"` // Order ID, etc.
	IsRead    bool           `json:"is_read" db:"is_read"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
}

// DriverApplication represents driver application request
type DriverApplication struct {
	ID           int64          `json:"id" db:"id"`
	UserID       int64          `json:"user_id" db:"user_id"`
	FullName     string         `json:"full_name" db:"full_name"`
	PhoneNumber  string         `json:"phone_number" db:"phone_number"`
	CarModel     string         `json:"car_model" db:"car_model"`
	CarNumber    string         `json:"car_number" db:"car_number"`
	LicenseImage string         `json:"license_image" db:"license_image"`
	Status       string         `json:"status" db:"status"` // pending, approved, rejected
	RejectionReason sql.NullString `json:"rejection_reason" db:"rejection_reason"`
	ReviewedBy   sql.NullInt64  `json:"reviewed_by" db:"reviewed_by"`
	ReviewedAt   sql.NullTime   `json:"reviewed_at" db:"reviewed_at"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
}

// Transaction represents balance transactions
type Transaction struct {
	ID          int64     `json:"id" db:"id"`
	DriverID    int64     `json:"driver_id" db:"driver_id"`
	OrderID     sql.NullInt64 `json:"order_id" db:"order_id"`
	Amount      float64   `json:"amount" db:"amount"`
	Type        string    `json:"type" db:"type"` // debit, credit, refund
	Description string    `json:"description" db:"description"`
	CreatedBy   sql.NullInt64 `json:"created_by" db:"created_by"` // Admin ID if manual
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Feedback represents user feedback/suggestions
type Feedback struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Message   string    `json:"message" db:"message"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
