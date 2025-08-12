package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// Connect to database
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

	// List all users with their roles
	fmt.Println("\n📋 Current users and their roles:")
	usersQuery := `
		SELECT u.user_id, u.user_name, u.email, 
		       COALESCE(array_agg(r.name) FILTER (WHERE r.name IS NOT NULL), ARRAY[]::text[]) as roles
		FROM users u
		LEFT JOIN user_roles ur ON u.user_id = ur.user_id
		LEFT JOIN roles r ON ur.role_id = r.id
		GROUP BY u.user_id, u.user_name, u.email
		ORDER BY u.user_name
	`

	rows, err := db.QueryContext(context.Background(), usersQuery)
	if err != nil {
		log.Fatal("Error querying users:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var userID int
		var userName, email sql.NullString
		var roles []string

		err := rows.Scan(&userID, &userName, &email, &roles)
		if err != nil {
			log.Printf("Error scanning user: %v", err)
			continue
		}

		fmt.Printf("User ID: %d | Name: %s | Email: %s | Roles: %v\n",
			userID, userName.String, email.String, roles)
	}

	// Check if admin role exists
	fmt.Println("\n🔍 Checking for admin role:")
	var adminRoleID int
	err = db.QueryRowContext(context.Background(), "SELECT id FROM roles WHERE name = 'admin'").Scan(&adminRoleID)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("❌ Admin role does not exist!")
			return
		}
		log.Fatal("Error checking admin role:", err)
	}
	fmt.Printf("✅ Admin role found with ID: %d\n", adminRoleID)

	// Check if super_admin role exists
	fmt.Println("\n🔍 Checking for super_admin role:")
	var superAdminRoleID int
	err = db.QueryRowContext(context.Background(), "SELECT id FROM roles WHERE name = 'super_admin'").Scan(&superAdminRoleID)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("❌ Super admin role does not exist!")
			return
		}
		log.Fatal("Error checking super admin role:", err)
	}
	fmt.Printf("✅ Super admin role found with ID: %d\n", superAdminRoleID)

	// Ask user which user ID to give admin permissions to
	fmt.Println("\n🔧 To fix role management permissions, you need admin or super_admin role.")
	fmt.Println("Enter the user ID you want to give admin permissions to (or press Enter to skip):")

	var userIDInput string
	fmt.Scanln(&userIDInput)

	if userIDInput == "" {
		fmt.Println("Skipping user permission update.")
		return
	}

	// Convert to int
	var userID int
	_, err = fmt.Sscanf(userIDInput, "%d", &userID)
	if err != nil {
		fmt.Println("❌ Invalid user ID format")
		return
	}

	// Check if user exists
	var userName string
	err = db.QueryRowContext(context.Background(), "SELECT user_name FROM users WHERE user_id = $1", userID).Scan(&userName)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("❌ User with ID %d does not exist\n", userID)
			return
		}
		log.Fatal("Error checking user:", err)
	}

	fmt.Printf("✅ User found: %s (ID: %d)\n", userName, userID)

	// Check if user already has admin role
	var hasAdmin bool
	err = db.QueryRowContext(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM user_roles WHERE user_id = $1 AND role_id = $2)",
		userID, adminRoleID).Scan(&hasAdmin)
	if err != nil {
		log.Fatal("Error checking admin role assignment:", err)
	}

	if hasAdmin {
		fmt.Println("✅ User already has admin role")
	} else {
		// Assign admin role
		_, err = db.ExecContext(context.Background(),
			"INSERT INTO user_roles (user_id, role_id, created_at) VALUES ($1, $2, NOW())",
			userID, adminRoleID)
		if err != nil {
			log.Fatal("Error assigning admin role:", err)
		}
		fmt.Println("✅ Admin role assigned successfully!")
	}

	// Check if user has super_admin role
	var hasSuperAdmin bool
	err = db.QueryRowContext(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM user_roles WHERE user_id = $1 AND role_id = $2)",
		userID, superAdminRoleID).Scan(&hasSuperAdmin)
	if err != nil {
		log.Fatal("Error checking super admin role assignment:", err)
	}

	if hasSuperAdmin {
		fmt.Println("✅ User already has super_admin role")
	} else {
		// Assign super_admin role
		_, err = db.ExecContext(context.Background(),
			"INSERT INTO user_roles (user_id, role_id, created_at) VALUES ($1, $2, NOW())",
			userID, superAdminRoleID)
		if err != nil {
			log.Fatal("Error assigning super admin role:", err)
		}
		fmt.Println("✅ Super admin role assigned successfully!")
	}

	fmt.Println("\n🎯 Summary:")
	fmt.Printf("- User %s (ID: %d) now has admin and super_admin roles\n", userName, userID)
	fmt.Println("- You should now be able to manage roles and role assignments")
	fmt.Println("- Log out and log back in to refresh your permissions")
}
