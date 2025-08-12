package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// Connect to database
	db, err := sql.Open("postgres", "postgres://postgres:pwaiswa@localhost:5432/ios?sslmode=disable")
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer db.Close()

	fmt.Println("✅ Connected to database successfully")

	// Check if afi_facilities table exists
	var exists bool
	query := `SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name = 'afi_facilities'
	)`

	err = db.QueryRowContext(context.Background(), query).Scan(&exists)
	if err != nil {
		fmt.Printf("❌ Error checking afi_facilities table: %v\n", err)
		return
	}

	if exists {
		fmt.Printf("✅ Table 'afi_facilities' exists\n")

		// Count rows
		var count int
		countQuery := "SELECT COUNT(*) FROM afi_facilities"
		err := db.QueryRowContext(context.Background(), countQuery).Scan(&count)
		if err != nil {
			fmt.Printf("   ⚠️  Error counting rows: %v\n", err)
		} else {
			fmt.Printf("   📊 Rows in table: %d\n", count)
		}

		// Get column structure
		fmt.Printf("\n📋 Table: afi_facilities\n")
		columnQuery := `
			SELECT column_name, data_type 
			FROM information_schema.columns 
			WHERE table_schema = 'public' 
			AND table_name = 'afi_facilities' 
			ORDER BY ordinal_position
		`

		rows, err := db.QueryContext(context.Background(), columnQuery)
		if err != nil {
			fmt.Printf("❌ Error getting columns: %v\n", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var columnName, dataType string
			if err := rows.Scan(&columnName, &dataType); err == nil {
				fmt.Printf("   📝 %s (%s)\n", columnName, dataType)
			}
		}
	} else {
		fmt.Printf("❌ Table 'afi_facilities' does not exist\n")

		// Check for similar tables
		fmt.Printf("\n🔍 Checking for similar tables...\n")
		similarQuery := `
			SELECT table_name 
			FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name LIKE '%facility%' OR table_name LIKE '%afi%'
		`

		rows, err := db.QueryContext(context.Background(), similarQuery)
		if err != nil {
			fmt.Printf("❌ Error checking for similar tables: %v\n", err)
		} else {
			defer rows.Close()
			for rows.Next() {
				var tableName string
				if err := rows.Scan(&tableName); err == nil {
					fmt.Printf("   📋 Found table: %s\n", tableName)
				}
			}
		}
	}

	fmt.Println("\n✅ afi_facilities check completed!")
}
