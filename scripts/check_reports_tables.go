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

	// Check if required tables exist
	tables := []string{"outbreaks", "facilities", "districts"}

	for _, table := range tables {
		var exists bool
		query := `SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)`

		err := db.QueryRowContext(context.Background(), query, table).Scan(&exists)
		if err != nil {
			fmt.Printf("❌ Error checking table %s: %v\n", table, err)
			continue
		}

		if exists {
			fmt.Printf("✅ Table '%s' exists\n", table)

			// Count rows in the table
			var count int
			countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
			err := db.QueryRowContext(context.Background(), countQuery).Scan(&count)
			if err != nil {
				fmt.Printf("   ⚠️  Error counting rows: %v\n", err)
			} else {
				fmt.Printf("   📊 Rows in table: %d\n", count)
			}

			// Check if table has the expected columns
			if table == "outbreaks" {
				checkTableColumns(db, table, []string{"outbreak_id", "outbreak_name", "active"})
			} else if table == "facilities" {
				checkTableColumns(db, table, []string{"facility_id", "facility_name", "active", "district_id"})
			} else if table == "districts" {
				checkTableColumns(db, table, []string{"district_id", "district_name", "active"})
			}
		} else {
			fmt.Printf("❌ Table '%s' does not exist\n", table)
		}
	}

	fmt.Println("\n✅ Table check completed!")
}

func checkTableColumns(db *sql.DB, tableName string, expectedColumns []string) {
	for _, column := range expectedColumns {
		var exists bool
		query := `SELECT EXISTS (
			SELECT FROM information_schema.columns 
			WHERE table_schema = 'public' 
			AND table_name = $1 
			AND column_name = $2
		)`

		err := db.QueryRowContext(context.Background(), query, tableName, column).Scan(&exists)
		if err != nil {
			fmt.Printf("   ⚠️  Error checking column %s: %v\n", column, err)
			continue
		}

		if exists {
			fmt.Printf("   ✅ Column '%s' exists\n", column)
		} else {
			fmt.Printf("   ❌ Column '%s' missing\n", column)
		}
	}
}
