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

	// Read migration file
	migrationSQL, err := os.ReadFile("migrations/045_create_lab_sample_types.sql")
	if err != nil {
		log.Fatal("Failed to read migration file:", err)
	}

	// Execute migration
	fmt.Println("🔄 Running lab sample types migration...")
	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		log.Fatal("Failed to execute migration:", err)
	}

	fmt.Println("✅ Lab sample types migration completed successfully!")

	// Verify tables were created
	fmt.Println("\n🔍 Verifying tables...")
	tables := []string{"lab_blood_types", "lab_swab_types", "lab_urine_types", "lab_sample_selections"}

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
		} else {
			fmt.Printf("  ❌ %s does not exist\n", table)
		}
	}

	fmt.Println("\n✅ Migration verification completed!")
}
