package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection parameters
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbName := getEnv("DB_NAME", "ios_db")
	dbUser := getEnv("DB_USER", "ios_user")
	dbPassword := getEnv("DB_PASSWORD", "ios_password")

	// Create connection string
	connStr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		dbHost, dbPort, dbName, dbUser, dbPassword)

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

	fmt.Println("✅ Connected to database successfully")
	fmt.Println()

	// Check if roles table exists
	var tableExists bool
	err = db.QueryRowContext(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'roles')").Scan(&tableExists)
	if err != nil {
		log.Fatal("Error checking if roles table exists:", err)
	}

	if !tableExists {
		fmt.Println("❌ Roles table does not exist!")
		return
	}
	fmt.Println("✅ Roles table exists")

	// Check total number of roles
	var totalRoles int
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM roles").Scan(&totalRoles)
	if err != nil {
		log.Fatal("Error counting roles:", err)
	}
	fmt.Printf("📊 Total roles in database: %d\n", totalRoles)

	// List all roles
	fmt.Println("\n📋 All roles in database:")
	fmt.Println("ID | Name | Description | Active")
	fmt.Println("---|------|-------------|-------")

	rows, err := db.QueryContext(context.Background(),
		"SELECT id, name, COALESCE(description, '') as description, is_active FROM roles ORDER BY name")
	if err != nil {
		log.Fatal("Error querying roles:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name, description string
		var isActive bool

		err := rows.Scan(&id, &name, &description, &isActive)
		if err != nil {
			log.Printf("Error scanning role: %v", err)
			continue
		}

		status := "✅"
		if !isActive {
			status = "❌"
		}

		fmt.Printf("%2d | %-20s | %-30s | %s\n", id, name, truncate(description, 30), status)
	}

	// Check for specific new roles
	fmt.Println("\n🔍 Checking for new roles from migration 044:")
	newRoles := []string{"reports", "vhf_lab_technician", "case_manager", "data_analyst"}

	for _, roleName := range newRoles {
		var exists bool
		err := db.QueryRowContext(context.Background(),
			"SELECT EXISTS(SELECT 1 FROM roles WHERE name = $1)", roleName).Scan(&exists)
		if err != nil {
			log.Printf("Error checking role %s: %v", roleName, err)
			continue
		}

		if exists {
			fmt.Printf("✅ Role '%s' exists\n", roleName)
		} else {
			fmt.Printf("❌ Role '%s' does not exist\n", roleName)
		}
	}

	// Check permissions table
	fmt.Println("\n🔍 Checking permissions table:")
	var totalPermissions int
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM permissions").Scan(&totalPermissions)
	if err != nil {
		fmt.Println("❌ Error counting permissions:", err)
	} else {
		fmt.Printf("📊 Total permissions: %d\n", totalPermissions)
	}

	// Check role_permissions table
	fmt.Println("\n🔍 Checking role_permissions table:")
	var totalRolePermissions int
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM role_permissions").Scan(&totalRolePermissions)
	if err != nil {
		fmt.Println("❌ Error counting role_permissions:", err)
	} else {
		fmt.Printf("📊 Total role_permissions: %d\n", totalRolePermissions)
	}

	// Check if the new functions exist
	fmt.Println("\n🔍 Checking for new database functions:")
	functions := []string{"user_has_reports_access", "get_user_report_access_level"}

	for _, funcName := range functions {
		var exists bool
		err := db.QueryRowContext(context.Background(),
			"SELECT EXISTS(SELECT 1 FROM information_schema.routines WHERE routine_name = $1)", funcName).Scan(&exists)
		if err != nil {
			log.Printf("Error checking function %s: %v", funcName, err)
			continue
		}

		if exists {
			fmt.Printf("✅ Function '%s' exists\n", funcName)
		} else {
			fmt.Printf("❌ Function '%s' does not exist\n", funcName)
		}
	}

	fmt.Println("\n🎯 Migration Status Summary:")
	if totalRoles >= 10 { // Should have at least the default roles + new ones
		fmt.Println("✅ Migration 044 appears to be applied")
	} else {
		fmt.Println("❌ Migration 044 may not be applied - run the migration!")
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
