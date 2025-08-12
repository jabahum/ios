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

	// Check facility table columns
	fmt.Printf("\n📋 Table: facility\n")

	query := `
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_schema = 'public' 
		AND table_name = 'facility' 
		ORDER BY ordinal_position
	`

	rows, err := db.QueryContext(context.Background(), query)
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

	// Check treatment_sites table columns
	fmt.Printf("\n📋 Table: treatment_sites\n")

	query = `
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_schema = 'public' 
		AND table_name = 'treatment_sites' 
		ORDER BY ordinal_position
	`

	rows, err = db.QueryContext(context.Background(), query)
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

	fmt.Println("\n✅ Column check completed!")
}
