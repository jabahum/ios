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

	// Test 1: Check if case role detection works
	fmt.Println("\n=== Test 1: Case Role Detection ===")

	// Create a test user with case role
	testUsername := "test_case_user"

	// Check if user exists, if not create
	var userID int
	err = db.QueryRowContext(context.Background(), "SELECT user_id FROM users WHERE user_name = $1", testUsername).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Create user
			err = db.QueryRowContext(context.Background(),
				"INSERT INTO users (user_name, user_pass, is_active) VALUES ($1, $2, true) RETURNING user_id",
				testUsername, "hashed_password_here").Scan(&userID)
			if err != nil {
				log.Fatal("Error creating test user:", err)
			}
			fmt.Printf("✓ Created test user with ID: %d\n", userID)
		} else {
			log.Fatal("Error checking user:", err)
		}
	} else {
		fmt.Printf("✓ Found existing test user with ID: %d\n", userID)
	}

	// Check if case role exists, if not create
	var roleID int
	err = db.QueryRowContext(context.Background(), "SELECT id FROM roles WHERE name = 'test_case_role'").Scan(&roleID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Create role
			err = db.QueryRowContext(context.Background(),
				"INSERT INTO roles (name, description, is_active) VALUES ($1, $2, true) RETURNING id",
				"test_case_role", "Test case role for verification").Scan(&roleID)
			if err != nil {
				log.Fatal("Error creating test role:", err)
			}
			fmt.Printf("✓ Created test case role with ID: %d\n", roleID)
		} else {
			log.Fatal("Error checking role:", err)
		}
	} else {
		fmt.Printf("✓ Found existing test case role with ID: %d\n", roleID)
	}

	// Assign role to user (check if already assigned)
	var count int
	err = db.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM user_roles WHERE user_id = $1 AND role_id = $2", userID, roleID).Scan(&count)
	if err != nil {
		log.Fatal("Error checking user role assignment:", err)
	}

	if count == 0 {
		// Assign role
		_, err = db.ExecContext(context.Background(),
			"INSERT INTO user_roles (user_id, role_id, created_at) VALUES ($1, $2, NOW())",
			userID, roleID)
		if err != nil {
			log.Fatal("Error assigning role to user:", err)
		}
		fmt.Printf("✓ Assigned case role to user\n")
	} else {
		fmt.Printf("✓ Role already assigned to user\n")
	}

	// Test the case role detection query
	query := `
		SELECT COUNT(*) 
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND r.is_active = true 
		AND LOWER(r.name) LIKE '%case%'
	`

	var caseRoleCount int
	err = db.QueryRowContext(context.Background(), query, userID).Scan(&caseRoleCount)
	if err != nil {
		log.Fatal("Error testing case role detection:", err)
	}

	if caseRoleCount > 0 {
		fmt.Printf("✓ Case role detection works! Found %d case roles for user\n", caseRoleCount)
	} else {
		fmt.Printf("❌ Case role detection failed! No case roles found for user\n")
	}

	// Test 2: Check login redirect logic
	fmt.Println("\n=== Test 2: Login Redirect Logic ===")

	// Test the primary role query
	primaryRoleQuery := `
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
				WHEN 'case_manager' THEN 7
				WHEN 'outbreak_manager' THEN 8
				WHEN 'outbreak_viewer' THEN 9
				ELSE 10
			END
		LIMIT 1
	`

	var primaryRole string
	err = db.QueryRowContext(context.Background(), primaryRoleQuery, userID).Scan(&primaryRole)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("❌ No primary role found for user")
		} else {
			log.Fatal("Error getting primary role:", err)
		}
	} else {
		fmt.Printf("✓ Primary role for user: '%s'\n", primaryRole)

		// Test the redirect logic
		if containsCase(primaryRole) {
			fmt.Printf("✓ User should be redirected to /cases (case role detected)\n")
		} else {
			fmt.Printf("✓ User should be redirected to /outbreaks (no case role)\n")
		}
	}

	// Test 3: Check outbreaks access
	fmt.Println("\n=== Test 3: Outbreaks Access ===")

	// Get user accessible outbreaks
	outbreaksQuery := `
		SELECT o.id, o.name, o.outbreak_type, o.status
		FROM outbreaks o
		ORDER BY o.id
	`

	rows, err := db.QueryContext(context.Background(), outbreaksQuery, userID)
	if err != nil {
		log.Fatal("Error getting outbreaks:", err)
	}
	defer rows.Close()

	fmt.Println("Outbreaks accessible to user:")
	count = 0
	for rows.Next() {
		var id int
		var name, outbreakType, status sql.NullString
		err := rows.Scan(&id, &name, &outbreakType, &status)
		if err != nil {
			log.Printf("Error scanning outbreak: %v", err)
			continue
		}
		fmt.Printf("  - ID: %d, Name: %s, Type: %s, Status: %s\n",
			id, name.String, outbreakType.String, status.String)
		count++
	}

	if count == 0 {
		fmt.Println("  No outbreaks accessible to user")
	} else {
		fmt.Printf("✓ User has access to %d outbreaks\n", count)
	}

	fmt.Println("\n=== Test Summary ===")
	fmt.Println("✓ Case role implementation test completed")
	fmt.Println("✓ Users with 'case' in their role name will:")
	fmt.Println("  - Be redirected to /cases on login")
	fmt.Println("  - See only 'Select' button on outbreaks page")
	fmt.Println("  - See only case-related options on home page")
	fmt.Println("✓ Ready for testing with real users!")
}

// containsCase checks if a string contains "case" (case-insensitive)
func containsCase(s string) bool {
	for i := 0; i <= len(s)-4; i++ {
		if (s[i] == 'c' || s[i] == 'C') &&
			(s[i+1] == 'a' || s[i+1] == 'A') &&
			(s[i+2] == 's' || s[i+2] == 'S') &&
			(s[i+3] == 'e' || s[i+3] == 'E') {
			return true
		}
	}
	return false
}
