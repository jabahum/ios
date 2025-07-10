package main

import (
	"crypto/sha1"
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

	// Create a case manager test user
	fmt.Println("\nCreating Case Manager Test User...")
	userID, err := createUser(db, "case_manager", "Case Manager User", "case_manager@example.com", "test123")
	if err != nil {
		log.Printf("Error creating case manager user: %v", err)
		return
	}

	fmt.Printf("✓ Created case manager user with ID: %d\n", userID)

	// Assign case_manager role
	fmt.Println("\nAssigning case_manager role...")
	err = assignRole(db, userID, "case_manager")
	if err != nil {
		log.Printf("Error assigning case_manager role: %v", err)
		return
	}

	fmt.Println("✓ Successfully assigned case_manager role")
	fmt.Println("\n=== Test User Created Successfully ===")
	fmt.Println("Username: case_manager")
	fmt.Println("Password: test123")
	fmt.Println("Role: case_manager")
	fmt.Println("Expected redirect: /cases (with default outbreak)")
	fmt.Println("==========================================")
}

func createUser(db *sql.DB, username, fullName, email, password string) (int, error) {
	// Hash password with SHA-1
	hash := sha1.New()
	hash.Write([]byte(password))
	hashedPassword := fmt.Sprintf("%x", hash.Sum(nil))

	// Insert user
	query := `INSERT INTO users (user_name, user_pass, is_active, is_locked) 
	          VALUES ($1, $2, true, false) RETURNING user_id`

	var userID int
	err := db.QueryRow(query, username, hashedPassword).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("error creating user: %v", err)
	}

	// Create enhanced user record
	enhancedQuery := `INSERT INTO enhanced_users 
	                  (user_id, first_name, last_name, email, is_active, created_at, updated_at)
	                  VALUES ($1, $2, $3, $4, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	_, err = db.Exec(enhancedQuery, userID, fullName, "", email)
	if err != nil {
		fmt.Printf("Warning: Could not create enhanced user record: %v\n", err)
	}

	return userID, nil
}

func assignRole(db *sql.DB, userID int, roleName string) error {
	// Get role ID
	var roleID int
	query := `SELECT id FROM roles WHERE name = $1`
	err := db.QueryRow(query, roleName).Scan(&roleID)
	if err != nil {
		return fmt.Errorf("error getting role ID for %s: %v", roleName, err)
	}

	// Assign role to user
	assignQuery := `INSERT INTO user_roles (user_id, role_id, created_at) VALUES ($1, $2, CURRENT_TIMESTAMP)`
	_, err = db.Exec(assignQuery, userID, roleID)
	if err != nil {
		return fmt.Errorf("error assigning role %s to user %d: %v", roleName, userID, err)
	}

	return nil
}
