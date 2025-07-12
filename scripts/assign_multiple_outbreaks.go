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
	connStr := "host=localhost port=5432 user=postgres password=pwaiswa dbname=ios sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test parameters
	userID := 2      // The case manager user
	outbreakID1 := 1 // First outbreak
	outbreakID2 := 2 // Second outbreak

	// Check if user exists
	var userExists bool
	err = db.QueryRowContext(context.Background(), "SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)", userID).Scan(&userExists)
	if err != nil {
		log.Fatal("Error checking if user exists:", err)
	}
	if !userExists {
		log.Fatal("User with ID", userID, "does not exist")
	}

	// Check if outbreaks exist and are active
	for _, outbreakID := range []int{outbreakID1, outbreakID2} {
		var outbreakExists bool
		var status string
		err = db.QueryRowContext(context.Background(), "SELECT EXISTS(SELECT 1 FROM outbreaks WHERE id = $1), status FROM outbreaks WHERE id = $1", outbreakID).Scan(&outbreakExists, &status)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Fatal("Outbreak with ID", outbreakID, "does not exist")
			}
			log.Fatal("Error checking outbreak", outbreakID, ":", err)
		}
		if !outbreakExists {
			log.Fatal("Outbreak with ID", outbreakID, "does not exist")
		}
		if status != "active" {
			log.Fatal("Outbreak with ID", outbreakID, "is not active (status:", status, ")")
		}
		fmt.Printf("Outbreak %d exists and is active\n", outbreakID)
	}

	// Assign user to both outbreaks
	for _, outbreakID := range []int{outbreakID1, outbreakID2} {
		// Check if assignment already exists
		var assignmentExists bool
		err = db.QueryRowContext(context.Background(), "SELECT EXISTS(SELECT 1 FROM user_outbreaks WHERE user_id = $1 AND outbreak_id = $2)", userID, outbreakID).Scan(&assignmentExists)
		if err != nil {
			log.Fatal("Error checking if assignment exists:", err)
		}

		if assignmentExists {
			fmt.Printf("User %d is already assigned to outbreak %d\n", userID, outbreakID)
		} else {
			// Insert the assignment
			_, err = db.ExecContext(context.Background(),
				"INSERT INTO user_outbreaks (user_id, outbreak_id, assigned_by, is_active) VALUES ($1, $2, $1, true)",
				userID, outbreakID)
			if err != nil {
				log.Fatal("Error assigning user to outbreak", outbreakID, ":", err)
			}
			fmt.Printf("Successfully assigned user %d to outbreak %d\n", userID, outbreakID)
		}
	}

	// Verify the assignments
	rows, err := db.QueryContext(context.Background(),
		`SELECT uo.outbreak_id, o.name, o.status 
		 FROM user_outbreaks uo 
		 JOIN outbreaks o ON uo.outbreak_id = o.id 
		 WHERE uo.user_id = $1 AND uo.is_active = true 
		 ORDER BY uo.assigned_at DESC`, userID)
	if err != nil {
		log.Fatal("Error verifying assignments:", err)
	}
	defer rows.Close()

	fmt.Printf("\nUser %d is currently assigned to:\n", userID)
	count := 0
	for rows.Next() {
		var outbreakID int
		var name, status string
		err := rows.Scan(&outbreakID, &name, &status)
		if err != nil {
			log.Fatal("Error scanning assignment:", err)
		}
		fmt.Printf("- Outbreak %d: %s (status: %s)\n", outbreakID, name, status)
		count++
	}

	if count == 0 {
		fmt.Println("No active outbreak assignments found for user", userID)
	} else {
		fmt.Printf("\nTotal active assignments: %d\n", count)
	}
}
