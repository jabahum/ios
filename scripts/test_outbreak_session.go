package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "password"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "case"
	}

	// Connect to database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test the outbreak session logic
	fmt.Println("Testing outbreak session handling...")

	// Test user ID 1 (assuming this user exists)
	userID := 1

	// Test getDefaultOutbreakID function logic
	fmt.Printf("\nTesting getDefaultOutbreakID for user %d:\n", userID)

	// Check if user has multiple active outbreaks
	query := `
		SELECT COUNT(*) 
		FROM user_outbreaks uo
		JOIN outbreaks o ON uo.outbreak_id = o.id
		WHERE uo.user_id = $1 AND uo.is_active = true AND o.status = 'active'
	`

	var outbreakCount int
	err = db.QueryRow(query, userID).Scan(&outbreakCount)
	if err != nil {
		log.Printf("Error checking outbreak count: %v", err)
		return
	}

	fmt.Printf("User has %d active outbreaks\n", outbreakCount)

	if outbreakCount == 0 {
		fmt.Println("User has no active outbreaks - should redirect to /outbreaks")
	} else if outbreakCount == 1 {
		// Get the single outbreak
		query = `
			SELECT uo.outbreak_id 
			FROM user_outbreaks uo
			JOIN outbreaks o ON uo.outbreak_id = o.id
			WHERE uo.user_id = $1 AND uo.is_active = true AND o.status = 'active'
			ORDER BY uo.assigned_at DESC
			LIMIT 1
		`

		var outbreakID int
		err := db.QueryRow(query, userID).Scan(&outbreakID)
		if err != nil {
			log.Printf("Error getting outbreak ID: %v", err)
			return
		}

		fmt.Printf("User has exactly one active outbreak (ID: %d) - should set outbreak_id in session\n", outbreakID)
	} else {
		fmt.Printf("User has multiple active outbreaks (%d) - should redirect to /outbreaks for selection\n", outbreakCount)

		// List the outbreaks
		query = `
			SELECT uo.outbreak_id, o.name, o.status
			FROM user_outbreaks uo
			JOIN outbreaks o ON uo.outbreak_id = o.id
			WHERE uo.user_id = $1 AND uo.is_active = true AND o.status = 'active'
			ORDER BY uo.assigned_at DESC
		`

		rows, err := db.Query(query, userID)
		if err != nil {
			log.Printf("Error getting outbreaks: %v", err)
			return
		}
		defer rows.Close()

		fmt.Println("Available outbreaks:")
		for rows.Next() {
			var outbreakID int
			var name, status string
			err := rows.Scan(&outbreakID, &name, &status)
			if err != nil {
				log.Printf("Error scanning outbreak: %v", err)
				continue
			}
			fmt.Printf("  - ID: %d, Name: %s, Status: %s\n", outbreakID, name, status)
		}
	}

	// Test case role check
	fmt.Printf("\nTesting case role check for user %d:\n", userID)

	query = `
		SELECT COUNT(*) 
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND r.is_active = true 
		AND LOWER(r.name) LIKE '%case%'
	`

	var caseRoleCount int
	err = db.QueryRow(query, userID).Scan(&caseRoleCount)
	if err != nil {
		log.Printf("Error checking case roles: %v", err)
		return
	}

	fmt.Printf("User has %d case-related roles\n", caseRoleCount)
	if caseRoleCount > 0 {
		fmt.Println("User has case roles - should be able to access case management features")
	} else {
		fmt.Println("User has no case roles - limited access to case management features")
	}

	fmt.Println("\nTest completed successfully!")
}
