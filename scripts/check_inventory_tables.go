package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection
	dbHost := "localhost"
	dbPort := 5432
	dbUser := "postgres"
	dbPassword := "pwaiswa"
	dbName := "ios"

	// Create connection string
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Error pinging database:", err)
	}

	fmt.Println("✅ Connected to database successfully!")

	// Check if inventory tables exist
	tables := []string{
		"inventory_items",
		"inventory_categories",
		"inventory_suppliers",
		"inventory_treatment_sites",
		"inventory_stock_levels",
		"inventory_transactions",
		"inventory_purchase_orders",
		"inventory_requisitions",
		"inventory_alerts",
		"inventory_donors",
		"inventory_donation_types",
		"inventory_donations",
		"inventory_donation_items",
	}

	fmt.Println("\n📋 Checking Inventory Tables:")
	for _, table := range tables {
		query := fmt.Sprintf(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = '%s'
			)
		`, table)

		var exists bool
		err := db.QueryRow(query).Scan(&exists)
		if err != nil {
			fmt.Printf("  ❌ Error checking %s: %v\n", table, err)
		} else if exists {
			fmt.Printf("  ✅ %s exists\n", table)
		} else {
			fmt.Printf("  ❌ %s does not exist\n", table)
		}
	}

	// Check if we can query inventory stats
	fmt.Println("\n🔍 Testing Inventory Stats Query:")
	statsQuery := `
		SELECT 
			COUNT(DISTINCT ii.id) as total_items,
			COUNT(DISTINCT isl.id) as total_stock_levels,
			COUNT(CASE WHEN isl.quantity <= ii.min_stock THEN 1 END) as low_stock_count,
			COALESCE(SUM(isl.quantity * ii.unit_cost), 0) as total_value
		FROM inventory_items ii
		LEFT JOIN inventory_stock_levels isl ON ii.id = isl.item_id
		WHERE ii.status = 'active'
	`

	rows, err := db.Query(statsQuery)
	if err != nil {
		fmt.Printf("  ❌ Error querying inventory stats: %v\n", err)
	} else {
		defer rows.Close()

		if rows.Next() {
			var totalItems, totalStockLevels, lowStockCount int
			var totalValue float64
			err := rows.Scan(&totalItems, &totalStockLevels, &lowStockCount, &totalValue)
			if err != nil {
				fmt.Printf("  ❌ Error scanning inventory stats: %v\n", err)
			} else {
				fmt.Printf("  ✅ Inventory stats query successful:\n")
				fmt.Printf("    Total Items: %d\n", totalItems)
				fmt.Printf("    Total Stock Levels: %d\n", totalStockLevels)
				fmt.Printf("    Low Stock Count: %d\n", lowStockCount)
				fmt.Printf("    Total Value: $%.2f\n", totalValue)
			}
		}
	}

	fmt.Println("\n✅ Inventory table check completed!")
}
