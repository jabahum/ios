package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection parameters
	// Using the same credentials as your application
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "pwaiswa")
	dbName := getEnv("DB_NAME", "ios")

	// Construct connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Connected to database successfully!")

	// Read SQL file
	sqlFile := "scripts/check_and_seed.sql"
	if _, err := os.Stat(sqlFile); os.IsNotExist(err) {
		// Try alternative path
		sqlFile = "check_and_seed.sql"
		if _, err := os.Stat(sqlFile); os.IsNotExist(err) {
			log.Fatalf("SQL file not found: %s", sqlFile)
		}
	}

	sqlContent, err := ioutil.ReadFile(sqlFile)
	if err != nil {
		log.Fatalf("Failed to read SQL file: %v", err)
	}

	// Execute SQL
	fmt.Println("Executing seed data SQL...")
	_, err = db.Exec(string(sqlContent))
	if err != nil {
		log.Fatalf("Failed to execute SQL: %v", err)
	}

	fmt.Println("Seed data inserted successfully!")

	// Verify data was inserted
	fmt.Println("\nVerifying inserted data...")

	queries := []struct {
		name  string
		query string
	}{
		{"Inventory Categories", "SELECT COUNT(*) FROM inventory_categories"},
		{"Inventory Suppliers", "SELECT COUNT(*) FROM inventory_suppliers"},
		{"Treatment Sites", "SELECT COUNT(*) FROM treatment_sites"},
		{"Departments", "SELECT COUNT(*) FROM departments"},
		{"Roles", "SELECT COUNT(*) FROM roles"},
		{"Permissions", "SELECT COUNT(*) FROM permissions"},
		{"Role-Permission Assignments", "SELECT COUNT(*) FROM role_permissions"},
		{"Inventory Items", "SELECT COUNT(*) FROM inventory_items"},
		{"Inventory Settings", "SELECT COUNT(*) FROM inventory_settings"},
	}

	for _, q := range queries {
		var count int
		err := db.QueryRow(q.query).Scan(&count)
		if err != nil {
			fmt.Printf("❌ %s: Error - %v\n", q.name, err)
		} else {
			fmt.Printf("✅ %s: %d records\n", q.name, count)
		}
	}

	fmt.Println("\nSeed data setup complete!")
	fmt.Println("You can now:")
	fmt.Println("- Create users and assign roles")
	fmt.Println("- Create inventory items and manage categories")
	fmt.Println("- Set up permissions for different user types")
	fmt.Println("- All dropdown lists should now be populated")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
