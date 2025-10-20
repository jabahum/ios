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
	if err := db.Ping(); err != nil {
		log.Fatal("Error pinging database:", err)
	}

	fmt.Println("✅ Connected to database successfully!")

	// Check users table schema
	fmt.Println("\n📋 Users table schema:")
	rows, err := db.Query(`
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_name = 'users'
		ORDER BY ordinal_position
	`)
	if err != nil {
		log.Fatal("Error querying users schema:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var columnName, dataType, isNullable string
		err := rows.Scan(&columnName, &dataType, &isNullable)
		if err != nil {
			continue
		}
		fmt.Printf("  - %s: %s (nullable: %s)\n", columnName, dataType, isNullable)
	}

	// Check roles table schema
	fmt.Println("\n📋 Roles table schema:")
	rows, err = db.Query(`
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_name = 'roles'
		ORDER BY ordinal_position
	`)
	if err != nil {
		log.Fatal("Error querying roles schema:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var columnName, dataType, isNullable string
		err := rows.Scan(&columnName, &dataType, &isNullable)
		if err != nil {
			continue
		}
		fmt.Printf("  - %s: %s (nullable: %s)\n", columnName, dataType, isNullable)
	}

	// Check user_roles table schema
	fmt.Println("\n📋 User_roles table schema:")
	rows, err = db.Query(`
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_name = 'user_roles'
		ORDER BY ordinal_position
	`)
	if err != nil {
		log.Fatal("Error querying user_roles schema:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var columnName, dataType, isNullable string
		err := rows.Scan(&columnName, &dataType, &isNullable)
		if err != nil {
			continue
		}
		fmt.Printf("  - %s: %s (nullable: %s)\n", columnName, dataType, isNullable)
	}
}
