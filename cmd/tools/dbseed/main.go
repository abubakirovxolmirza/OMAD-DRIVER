package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"taxi-service/internal/config"
	"taxi-service/internal/database"
)

var baseFares = map[string]float64{
	"Toshkent shahri":      35000,
	"Toshkent viloyati":    45000,
	"Andijon":              180000,
	"Buxoro":               210000,
	"Farg'ona":             170000,
	"Jizzax":               120000,
	"Xorazm":               220000,
	"Namangan":             175000,
	"Navoiy":               195000,
	"Qashqadaryo":          185000,
	"Qoraqalpog'iston":     320000,
	"Samarqand":            160000,
	"Sirdaryo":             110000,
	"Surxondaryo":          240000,
}

func main() {
	force := flag.Bool("force", false, "Skip confirmation prompt")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := database.Connect(&cfg.Database); err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	if !*force {
		if !confirmExecute() {
			fmt.Println("Operation cancelled.")
			return
		}
	}

	if err := cleanupAndSeed(&cfg); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Database cleanup and seed completed successfully.")
}

func confirmExecute() bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("WARNING: This will truncate core tables and repopulate reference data.")
	fmt.Print("Type 'yes' to continue: ")
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(strings.ToLower(input)) == "yes"
}

func cleanupAndSeed(cfg *config.Config) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	tables := []string{
		"notifications",
		"ratings",
		"transactions",
		"orders",
		"driver_applications",
		"drivers",
		"feedback",
		"pricing",
		"districts",
		"regions",
		"discounts",
	}

	for _, table := range tables {
		stmt := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("failed to truncate %s: %w", table, err)
		}
	}

	sqlPath := filepath.Join("database", "migrations", "001_add_locations_and_uzbekistan_data.sql")
	sqlBytes, err := os.ReadFile(sqlPath)
	if err != nil {
		return fmt.Errorf("failed to read seed sql: %w", err)
	}

	if _, err := tx.Exec(string(sqlBytes)); err != nil {
		return fmt.Errorf("failed to execute seed sql: %w", err)
	}

	if err := seedDiscounts(tx, cfg); err != nil {
		return err
	}

	regionIDs, err := loadRegionIDs(tx)
	if err != nil {
		return err
	}

	if err := seedPricing(tx, cfg, regionIDs); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return reportCounts()
}

func seedDiscounts(tx *sql.Tx, cfg *config.Config) error {
	_, err := tx.Exec(`
		INSERT INTO discounts (passenger_count, discount_percentage)
		VALUES 
			(1, $1),
			(2, $2),
			(3, $3),
			(4, $4)
		ON CONFLICT (passenger_count) DO UPDATE SET
			discount_percentage = EXCLUDED.discount_percentage,
			updated_at = CURRENT_TIMESTAMP
	`,
		cfg.Pricing.Discount1Person,
		cfg.Pricing.Discount2Person,
		cfg.Pricing.Discount3Person,
		cfg.Pricing.DiscountFullCar,
	)
	if err != nil {
		return fmt.Errorf("failed to seed discounts: %w", err)
	}
	return nil
}

func loadRegionIDs(tx *sql.Tx) (map[string]int64, error) {
	rows, err := tx.Query("SELECT id, name_uz_lat FROM regions")
	if err != nil {
		return nil, fmt.Errorf("failed to load regions: %w", err)
	}
	defer rows.Close()

	regionIDs := make(map[string]int64)
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, fmt.Errorf("failed to scan region: %w", err)
		}
		regionIDs[name] = id
	}

	return regionIDs, rows.Err()
}

func seedPricing(tx *sql.Tx, cfg *config.Config, regions map[string]int64) error {
	serviceFee := cfg.Pricing.ServiceFeePercentage
	entries := 0

	for fromName, fromID := range regions {
		for toName, toID := range regions {
			if fromID == toID {
				continue
			}

			fromBase := baseFares[fromName]
			toBase := baseFares[toName]
			if fromBase == 0 {
				fromBase = 150000
			}
			if toBase == 0 {
				toBase = 150000
			}

			basePrice := roundTo((fromBase+toBase)/2, 5000)
			pricePerPerson := roundTo(basePrice*0.18, 500)

			if _, err := tx.Exec(`
				INSERT INTO pricing (from_region_id, to_region_id, base_price, price_per_person, service_fee)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (from_region_id, to_region_id) DO UPDATE SET
					base_price = EXCLUDED.base_price,
					price_per_person = EXCLUDED.price_per_person,
					service_fee = EXCLUDED.service_fee,
					updated_at = CURRENT_TIMESTAMP
			`, fromID, toID, basePrice, pricePerPerson, serviceFee); err != nil {
				return fmt.Errorf("failed to insert pricing for %s -> %s: %w", fromName, toName, err)
			}
			entries++
		}
	}

	fmt.Printf("Seeded %d pricing routes.\n", entries)
	return nil
}

func reportCounts() error {
	var regionCount, districtCount, pricingCount int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM regions").Scan(&regionCount); err != nil {
		return err
	}
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM districts").Scan(&districtCount); err != nil {
		return err
	}
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM pricing").Scan(&pricingCount); err != nil {
		return err
	}

	fmt.Printf("Regions: %d, Districts: %d, Pricing routes: %d\n", regionCount, districtCount, pricingCount)
	return nil
}

func roundTo(value, unit float64) float64 {
	if unit <= 0 {
		return value
	}
	return math.Round(value/unit) * unit
}

