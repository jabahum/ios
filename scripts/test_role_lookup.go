package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection
	dsn := "host=localhost port=5432 user=postgres password=pwaiswa dbname=ios sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("Error pinging database:", err)
	}

	fmt.Println("✓ Connected to database successfully")

	// Get user ID for vhf_test
	var userID int
	err = db.QueryRowContext(context.Background(), "SELECT user_id FROM users WHERE user_name = $1", "vhf_test").Scan(&userID)
	if err != nil {
		log.Fatal("Error getting user ID:", err)
	}

	fmt.Printf("✓ Found user vhf_test with ID: %d\n", userID)

	// Test the role lookup query
	query := `
		SELECT r.name
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND r.is_active = true
		ORDER BY 
			CASE r.name 
				WHEN 'super_admin' THEN 1
				WHEN 'admin' THEN 2
				WHEN 'vhf_lab_technician' THEN 3
				WHEN 'vhf_data_entry' THEN 4
				WHEN 'mpox_case_manager' THEN 5
				WHEN 'mpox_data_entry' THEN 6
				WHEN 'outbreak_manager' THEN 7
				WHEN 'outbreak_viewer' THEN 8
				ELSE 9
			END
		LIMIT 1
	`

	var roleName string
	err = db.QueryRowContext(context.Background(), query, userID).Scan(&roleName)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("❌ No role found for user")
			return
		}
		log.Fatal("Error getting role:", err)
	}

	fmt.Printf("✓ Found role: '%s'\n", roleName)

	// Test the redirect logic
	switch roleName {
	case "vhf_lab_technician", "vhf_data_entry":
		fmt.Println("✓ Should redirect to: /vhf-list")
	case "mpox_case_manager", "mpox_data_entry":
		fmt.Println("✓ Should redirect to: /outbreaks")
	case "outbreak_viewer", "outbreak_manager":
		fmt.Println("✓ Should redirect to: /outbreaks")
	case "super_admin", "admin":
		fmt.Println("✓ Should redirect to: /outbreaks")
	default:
		fmt.Printf("✓ Should redirect to: /outbreaks (default for role: '%s')\n", roleName)
	}
}
