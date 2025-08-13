package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection string
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:pwaiswa@localhost:5432/ios?sslmode=disable"
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	fmt.Println("✅ Connected to database successfully!")

	// Check if lab tables exist
	fmt.Println("\n📋 Checking Lab Tables:")

	tables := []string{
		"lab_blood_types",
		"lab_swab_types",
		"lab_urine_types",
		"lab_sample_selections",
	}

	for _, table := range tables {
		var exists bool
		query := `SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)`

		err := db.QueryRow(query, table).Scan(&exists)
		if err != nil {
			fmt.Printf("  ❌ Error checking %s: %v\n", table, err)
			continue
		}

		if exists {
			fmt.Printf("  ✅ %s exists\n", table)

			// Count rows in the table
			var count int
			countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
			err := db.QueryRow(countQuery).Scan(&count)
			if err != nil {
				fmt.Printf("    ❌ Error counting rows: %v\n", err)
			} else {
				fmt.Printf("    📊 %d rows found\n", count)
			}
		} else {
			fmt.Printf("  ❌ %s does not exist\n", table)
		}
	}

	// Test lab blood types query
	fmt.Println("\n🔍 Testing Lab Blood Types Query:")
	rows, err := db.Query("SELECT id, name, description, category FROM lab_blood_types LIMIT 5")
	if err != nil {
		fmt.Printf("  ❌ Error querying lab_blood_types: %v\n", err)
	} else {
		defer rows.Close()
		fmt.Println("  ✅ lab_blood_types query successful")

		var count int
		for rows.Next() {
			var id int
			var name, description, category string
			err := rows.Scan(&id, &name, &description, &category)
			if err != nil {
				fmt.Printf("    ❌ Error scanning row: %v\n", err)
				continue
			}
			fmt.Printf("    📋 ID: %d, Name: %s, Category: %s\n", id, name, category)
			count++
		}
		fmt.Printf("    📊 Found %d blood types\n", count)
	}

	fmt.Println("\n✅ Lab table check completed!")
}
