package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// Connect to database with correct credentials
	db, err := sql.Open("postgres", "postgres://postgres:pwaiswa@localhost:5432/ios?sslmode=disable")
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

	// Check total number of roles
	var totalRoles int
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM roles").Scan(&totalRoles)
	if err != nil {
		log.Fatal("Error counting roles:", err)
	}
	fmt.Printf("📊 Total roles in database: %d\n", totalRoles)

	// List all roles
	fmt.Println("\n📋 All roles in database:")
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

		fmt.Printf("ID: %2d | Name: %-20s | Description: %-30s | Active: %s\n",
			id, name, truncate(description, 30), status)
	}

	// Check if migration 044 has been applied
	fmt.Println("\n🔍 Checking for new roles from migration 044:")
	newRoles := []string{"reports", "vhf_lab_technician", "case_manager", "data_analyst"}

	missingRoles := []string{}
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
			missingRoles = append(missingRoles, roleName)
		}
	}

	if len(missingRoles) > 0 {
		fmt.Println("\n🚨 Migration 044 needs to be applied!")
		fmt.Println("Run this command to apply the migration:")
		fmt.Println("psql -d ios -f migrations/044_add_reports_role.sql")
	} else {
		fmt.Println("\n✅ Migration 044 appears to be applied")
	}

	// Test the exact query used in the dropdown
	fmt.Println("\n🔍 Testing the exact query used in roles dropdown:")
	dropdownQuery := `SELECT id, name FROM roles ORDER BY name`
	dropdownRows, err := db.QueryContext(context.Background(), dropdownQuery)
	if err != nil {
		log.Fatal("Error testing dropdown query:", err)
	}
	defer dropdownRows.Close()

	var dropdownRoles []string
	for dropdownRows.Next() {
		var id int
		var name string
		err := dropdownRows.Scan(&id, &name)
		if err != nil {
			log.Printf("Error scanning dropdown role: %v", err)
			continue
		}
		dropdownRoles = append(dropdownRoles, name)
		fmt.Printf("Dropdown role: ID=%d, Name=%s\n", id, name)
	}

	fmt.Printf("\n📊 Total roles in dropdown: %d\n", len(dropdownRoles))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
