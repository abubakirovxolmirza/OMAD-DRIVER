package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"taxi-service/internal/config"
)

// DB is the global database connection
var DB *sql.DB

// Connect establishes database connection
func Connect(cfg *config.DatabaseConfig) error {
	var err error
	dsn := cfg.GetDSN()

	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)

	log.Println("Database connected successfully")
	return nil
}

// Close closes database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// InitSchema creates all database tables
func InitSchema() error {
	schema := `
	-- Users table
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		phone_number VARCHAR(20) UNIQUE NOT NULL,
		name VARCHAR(100) NOT NULL,
		password VARCHAR(255) NOT NULL,
		avatar VARCHAR(255),
		role VARCHAR(20) NOT NULL DEFAULT 'user',
		language VARCHAR(20) NOT NULL DEFAULT 'uz_latin',
		is_blocked BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Drivers table
	CREATE TABLE IF NOT EXISTS drivers (
		id SERIAL PRIMARY KEY,
		user_id INTEGER UNIQUE REFERENCES users(id) ON DELETE CASCADE,
		full_name VARCHAR(200) NOT NULL,
		car_model VARCHAR(100) NOT NULL,
		car_number VARCHAR(20) NOT NULL,
		license_image VARCHAR(255),
		balance DECIMAL(12,2) DEFAULT 0,
		rating DECIMAL(3,2) DEFAULT 0,
		total_ratings INTEGER DEFAULT 0,
		status VARCHAR(20) DEFAULT 'pending',
		is_active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Regions table
	CREATE TABLE IF NOT EXISTS regions (
		id SERIAL PRIMARY KEY,
		name_uz_lat VARCHAR(100) NOT NULL,
		name_uz_cyr VARCHAR(100) NOT NULL,
		name_ru VARCHAR(100) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Districts table
	CREATE TABLE IF NOT EXISTS districts (
		id SERIAL PRIMARY KEY,
		region_id INTEGER REFERENCES regions(id) ON DELETE CASCADE,
		name_uz_lat VARCHAR(100) NOT NULL,
		name_uz_cyr VARCHAR(100) NOT NULL,
		name_ru VARCHAR(100) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Orders table
	CREATE TABLE IF NOT EXISTS orders (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		driver_id INTEGER REFERENCES drivers(id) ON DELETE SET NULL,
		order_type VARCHAR(20) NOT NULL,
		status VARCHAR(20) NOT NULL DEFAULT 'pending',
		customer_name VARCHAR(100) NOT NULL,
		customer_phone VARCHAR(20) NOT NULL,
		recipient_phone VARCHAR(20),
		from_region_id INTEGER REFERENCES regions(id),
		from_district_id INTEGER REFERENCES districts(id),
		to_region_id INTEGER REFERENCES regions(id),
		to_district_id INTEGER REFERENCES districts(id),
		passenger_count INTEGER,
		delivery_type VARCHAR(20),
		scheduled_date DATE NOT NULL,
		time_range_start VARCHAR(10) NOT NULL,
		time_range_end VARCHAR(10) NOT NULL,
		price DECIMAL(12,2) NOT NULL,
		service_fee DECIMAL(12,2) NOT NULL,
		discount_percentage DECIMAL(5,2) DEFAULT 0,
		final_price DECIMAL(12,2) NOT NULL,
		notes TEXT,
		cancellation_reason TEXT,
		accepted_at TIMESTAMP,
		accept_deadline TIMESTAMP,
		completed_at TIMESTAMP,
		cancelled_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Pricing table
	CREATE TABLE IF NOT EXISTS pricing (
		id SERIAL PRIMARY KEY,
		from_region_id INTEGER REFERENCES regions(id),
		to_region_id INTEGER REFERENCES regions(id),
		base_price DECIMAL(12,2) NOT NULL,
		price_per_person DECIMAL(12,2) NOT NULL,
		service_fee DECIMAL(5,2) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(from_region_id, to_region_id)
	);

	-- Discounts table
	CREATE TABLE IF NOT EXISTS discounts (
		id SERIAL PRIMARY KEY,
		passenger_count INTEGER UNIQUE NOT NULL,
		discount_percentage DECIMAL(5,2) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Ratings table
	CREATE TABLE IF NOT EXISTS ratings (
		id SERIAL PRIMARY KEY,
		order_id INTEGER UNIQUE REFERENCES orders(id) ON DELETE CASCADE,
		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		driver_id INTEGER REFERENCES drivers(id) ON DELETE CASCADE,
		rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
		comment TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Notifications table
	CREATE TABLE IF NOT EXISTS notifications (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		title VARCHAR(200) NOT NULL,
		message TEXT NOT NULL,
		type VARCHAR(50) NOT NULL,
		related_id INTEGER,
		is_read BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Driver applications table
	CREATE TABLE IF NOT EXISTS driver_applications (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		full_name VARCHAR(200) NOT NULL,
		phone_number VARCHAR(20) NOT NULL,
		car_model VARCHAR(100) NOT NULL,
		car_number VARCHAR(20) NOT NULL,
		license_image VARCHAR(255) NOT NULL,
		status VARCHAR(20) DEFAULT 'pending',
		rejection_reason TEXT,
		reviewed_by INTEGER REFERENCES users(id),
		reviewed_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Transactions table
	CREATE TABLE IF NOT EXISTS transactions (
		id SERIAL PRIMARY KEY,
		driver_id INTEGER REFERENCES drivers(id) ON DELETE CASCADE,
		order_id INTEGER REFERENCES orders(id),
		amount DECIMAL(12,2) NOT NULL,
		type VARCHAR(20) NOT NULL,
		description TEXT NOT NULL,
		created_by INTEGER REFERENCES users(id),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Feedback table
	CREATE TABLE IF NOT EXISTS feedback (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		message TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Create indexes
	CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone_number);
	CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
	CREATE INDEX IF NOT EXISTS idx_drivers_user_id ON drivers(user_id);
	CREATE INDEX IF NOT EXISTS idx_drivers_status ON drivers(status);
	CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
	CREATE INDEX IF NOT EXISTS idx_orders_driver_id ON orders(driver_id);
	CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
	CREATE INDEX IF NOT EXISTS idx_orders_type ON orders(order_type);
	CREATE INDEX IF NOT EXISTS idx_orders_scheduled_date ON orders(scheduled_date);
	CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
	CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read);
	CREATE INDEX IF NOT EXISTS idx_districts_region_id ON districts(region_id);
	`

	_, err := DB.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	log.Println("Database schema initialized successfully")
	return nil
}

// SeedInitialData inserts initial data for development
func SeedInitialData() error {
	// Insert default discounts
	_, err := DB.Exec(`
		INSERT INTO discounts (passenger_count, discount_percentage) VALUES
		(1, 0), (2, 10), (3, 15), (4, 20)
		ON CONFLICT (passenger_count) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to seed discounts: %w", err)
	}

	// Insert sample regions (Uzbekistan regions)
	_, err = DB.Exec(`
		INSERT INTO regions (name_uz_lat, name_uz_cyr, name_ru) VALUES
		('Toshkent', 'Тошкент', 'Ташкент'),
		('Samarqand', 'Самарқанд', 'Самарканд'),
		('Buxoro', 'Бухоро', 'Бухара'),
		('Andijon', 'Андижон', 'Андижан'),
		('Farg''ona', 'Фарғона', 'Фергана'),
		('Namangan', 'Наманган', 'Наманган'),
		('Qashqadaryo', 'Қашқадарё', 'Кашкадарья'),
		('Surxondaryo', 'Сурхондарё', 'Сурхандарья'),
		('Sirdaryo', 'Сирдарё', 'Сырдарья'),
		('Jizzax', 'Жиззах', 'Джизак'),
		('Navoiy', 'Навоий', 'Навои'),
		('Xorazm', 'Хоразм', 'Хорезм'),
		('Qoraqalpog''iston', 'Қорақалпоғистон', 'Каракалпакстан')
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to seed regions: %w", err)
	}

	log.Println("Initial data seeded successfully")
	return nil
}
