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
	userID := 2     // The case manager user
	outbreakID := 1 // The outbreak to assign

	// Check if user exists
	var userExists bool
	err = db.QueryRowContext(context.Background(), "SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)", userID).Scan(&userExists)
	if err != nil {
		log.Fatal("Error checking if user exists:", err)
	}
	if !userExists {
		log.Fatal("User with ID", userID, "does not exist")
	}

	// Check if outbreak exists
	var outbreakExists bool
	err = db.QueryRowContext(context.Background(), "SELECT EXISTS(SELECT 1 FROM outbreaks WHERE id = $1)", outbreakID).Scan(&outbreakExists)
	if err != nil {
		log.Fatal("Error checking if outbreak exists:", err)
	}
	if !outbreakExists {
		log.Fatal("Outbreak with ID", outbreakID, "does not exist")
	}

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
			log.Fatal("Error assigning user to outbreak:", err)
		}
		fmt.Printf("Successfully assigned user %d to outbreak %d\n", userID, outbreakID)
	}

	// Verify the assignment
	var assignedOutbreakID int
	err = db.QueryRowContext(context.Background(),
		"SELECT outbreak_id FROM user_outbreaks WHERE user_id = $1 AND is_active = true ORDER BY assigned_at DESC LIMIT 1",
		userID).Scan(&assignedOutbreakID)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("No active outbreak assignments found for user", userID)
		} else {
			log.Fatal("Error verifying assignment:", err)
		}
	} else {
		fmt.Printf("User %d is currently assigned to outbreak %d\n", userID, assignedOutbreakID)
	}
}
