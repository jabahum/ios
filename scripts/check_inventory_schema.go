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

	// Check schema of key inventory tables
	tables := []string{
		"inventory_items",
		"inventory_stock_levels",
		"inventory_categories",
		"inventory_suppliers",
	}

	fmt.Println("\n📋 Checking Inventory Table Schemas:")
	for _, table := range tables {
		query := fmt.Sprintf(`
			SELECT column_name, data_type, is_nullable
			FROM information_schema.columns 
			WHERE table_schema = 'public' 
			AND table_name = '%s'
			ORDER BY ordinal_position
		`, table)

		rows, err := db.Query(query)
		if err != nil {
			fmt.Printf("  ❌ Error checking %s schema: %v\n", table, err)
			continue
		}
		defer rows.Close()

		fmt.Printf("  📋 %s columns:\n", table)
		for rows.Next() {
			var columnName, dataType, isNullable string
			err := rows.Scan(&columnName, &dataType, &isNullable)
			if err != nil {
				continue
			}
			fmt.Printf("    %s (%s, nullable: %s)\n", columnName, dataType, isNullable)
		}
		fmt.Println()
	}

	// Check if treatment_sites table exists (might be named differently)
	fmt.Println("\n🔍 Checking for treatment sites table:")
	sitesQuery := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name LIKE '%site%'
	`

	rows, err := db.Query(sitesQuery)
	if err != nil {
		fmt.Printf("  ❌ Error checking for sites tables: %v\n", err)
	} else {
		defer rows.Close()

		for rows.Next() {
			var tableName string
			err := rows.Scan(&tableName)
			if err != nil {
				continue
			}
			fmt.Printf("  ✅ Found table: %s\n", tableName)
		}
	}

	fmt.Println("\n✅ Schema check completed!")
}
