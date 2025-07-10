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

	// Create a VHF test user
	fmt.Println("\nCreating VHF Test User...")
	userID, err := createUser(db, "vhf_test", "VHF Test User", "vhf_test@example.com", "test123")
	if err != nil {
		log.Printf("Error creating VHF test user: %v", err)
		return
	}
	fmt.Printf("✓ Created VHF Test User (ID: %d)\n", userID)

	// Assign VHF lab technician role
	if err := assignRoleToUser(db, userID, "vhf_lab_technician"); err != nil {
		log.Printf("Error assigning role to VHF test user: %v", err)
		return
	}
	fmt.Println("✓ Assigned vhf_lab_technician role")

	fmt.Println("\n✓ VHF Test User created successfully!")
	fmt.Println("Username: vhf_test")
	fmt.Println("Password: test123")
	fmt.Println("This user will be redirected to /vhf-list when they log in.")
}

func hashPasswordSHA1(password string) string {
	h := sha1.New()
	h.Write([]byte(password))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func createUser(db *sql.DB, username, fullName, email, password string) (int, error) {
	hashedPassword := hashPasswordSHA1(password)

	query := `
		INSERT INTO users (user_name, user_pass, email, first_name, last_name, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, true, NOW())
		RETURNING user_id
	`

	var userID int
	err := db.QueryRow(query, username, hashedPassword, email, fullName, fullName).Scan(&userID)
	return userID, err
}

func assignRoleToUser(db *sql.DB, userID int, roleName string) error {
	query := `
		INSERT INTO user_roles (user_id, role_id, created_at)
		SELECT $1, r.id, NOW()
		FROM roles r
		WHERE r.name = $2 AND r.is_active = true
		ON CONFLICT (user_id, role_id) DO NOTHING
	`

	_, err := db.Exec(query, userID, roleName)
	return err
}
