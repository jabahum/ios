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

	// Check actual column names in tables
	tables := []string{"outbreaks", "districts"}

	for _, table := range tables {
		fmt.Printf("\n📋 Table: %s\n", table)

		query := `
			SELECT column_name, data_type 
			FROM information_schema.columns 
			WHERE table_schema = 'public' 
			AND table_name = $1 
			ORDER BY ordinal_position
		`

		rows, err := db.QueryContext(context.Background(), query, table)
		if err != nil {
			fmt.Printf("❌ Error getting columns: %v\n", err)
			continue
		}
		defer rows.Close()

		for rows.Next() {
			var columnName, dataType string
			if err := rows.Scan(&columnName, &dataType); err == nil {
				fmt.Printf("   📝 %s (%s)\n", columnName, dataType)
			}
		}
	}

	// Check if facilities table exists with different name
	fmt.Printf("\n🔍 Checking for facilities table...\n")

	// List all tables that might be facilities
	tableQuery := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name LIKE '%facility%' OR table_name LIKE '%site%'
	`

	rows, err := db.QueryContext(context.Background(), tableQuery)
	if err != nil {
		fmt.Printf("❌ Error checking for facility tables: %v\n", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err == nil {
				fmt.Printf("   📋 Found table: %s\n", tableName)
			}
		}
	}

	fmt.Println("\n✅ Column check completed!")
}
