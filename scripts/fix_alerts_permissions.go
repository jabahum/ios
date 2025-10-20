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
	dbHost := "localhost"
	dbPort := 5432
	dbUser := "postgres"
	dbPassword := "pwaiswa"
	dbName := "ios"

	// Create connection string
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		log.Fatal("Error pinging database:", err)
	}

	fmt.Println("✅ Connected to database successfully!")

	// Check if alerts permissions exist
	fmt.Println("\n📋 Checking alerts permissions...")

	var alertsPermID int
	err = db.QueryRowContext(ctx, `
		SELECT id FROM permissions 
		WHERE resource = 'alerts' AND action = 'read'
	`).Scan(&alertsPermID)

	if err == sql.ErrNoRows {
		// Create alerts permission
		fmt.Println("Creating alerts:read permission...")
		err = db.QueryRowContext(ctx, `
			INSERT INTO permissions (resource, action, description, is_active, created_at)
			VALUES ('alerts', 'read', 'Read access to alerts', true, NOW())
			RETURNING id
		`).Scan(&alertsPermID)
		if err != nil {
			log.Fatal("Error creating alerts permission:", err)
		}
		fmt.Printf("✅ Created alerts:read permission with ID: %d\n", alertsPermID)
	} else if err != nil {
		log.Fatal("Error checking alerts permission:", err)
	} else {
		fmt.Printf("✅ Alerts:read permission already exists with ID: %d\n", alertsPermID)
	}

	// Get admin role ID
	var adminRoleID int
	err = db.QueryRowContext(ctx, "SELECT id FROM roles WHERE name = 'admin'").Scan(&adminRoleID)
	if err != nil {
		log.Fatal("Error getting admin role ID:", err)
	}
	fmt.Printf("Admin role ID: %d\n", adminRoleID)

	// Check if admin role has alerts permission
	var hasAlertsPerm bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM role_permissions 
			WHERE role_id = $1 AND permission_id = $2
		)
	`, adminRoleID, alertsPermID).Scan(&hasAlertsPerm)

	if err != nil {
		log.Fatal("Error checking admin alerts permission:", err)
	}

	if !hasAlertsPerm {
		// Add alerts permission to admin role
		fmt.Println("Adding alerts:read permission to admin role...")
		_, err = db.ExecContext(ctx, `
			INSERT INTO role_permissions (role_id, permission_id, created_at)
			VALUES ($1, $2, NOW())
			ON CONFLICT (role_id, permission_id) DO NOTHING
		`, adminRoleID, alertsPermID)
		if err != nil {
			log.Fatal("Error adding alerts permission to admin role:", err)
		}
		fmt.Println("✅ Added alerts:read permission to admin role")
	} else {
		fmt.Println("✅ Admin role already has alerts:read permission")
	}

	// Check all users with admin role
	fmt.Println("\n👥 Users with admin role:")
	rows, err := db.QueryContext(ctx, `
		SELECT u.user_id, u.username, u.email
		FROM users u
		JOIN user_roles ur ON u.user_id = ur.user_id
		JOIN roles r ON ur.role_id = r.id
		WHERE r.name = 'admin' AND u.is_active = true
		ORDER BY u.username
	`)
	if err != nil {
		log.Fatal("Error querying admin users:", err)
	}
	defer rows.Close()

	adminUsers := []struct {
		ID       int
		Username string
		Email    string
	}{}

	for rows.Next() {
		var user struct {
			ID       int
			Username string
			Email    string
		}
		err := rows.Scan(&user.ID, &user.Username, &user.Email)
		if err != nil {
			continue
		}
		adminUsers = append(adminUsers, user)
		fmt.Printf("  - %s (%s) - ID: %d\n", user.Username, user.Email, user.ID)
	}

	if len(adminUsers) == 0 {
		fmt.Println("⚠️  No admin users found. You may need to create an admin user first.")
	} else {
		fmt.Printf("\n✅ Found %d admin user(s) who should now have access to alerts\n", len(adminUsers))
	}

	// Verify the permission is working
	fmt.Println("\n🔍 Verifying alerts permission...")
	for _, user := range adminUsers {
		var hasPermission bool
		err = db.QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM permissions p
				JOIN role_permissions rp ON p.id = rp.permission_id
				JOIN user_roles ur ON rp.role_id = ur.role_id
				WHERE ur.user_id = $1 AND p.resource = 'alerts' AND p.action = 'read'
			)
		`, user.ID).Scan(&hasPermission)

		if err != nil {
			fmt.Printf("❌ Error checking permission for user %s: %v\n", user.Username, err)
		} else if hasPermission {
			fmt.Printf("✅ User %s has alerts:read permission\n", user.Username)
		} else {
			fmt.Printf("❌ User %s does NOT have alerts:read permission\n", user.Username)
		}
	}

	fmt.Println("\n🎯 Summary:")
	fmt.Println("- Alerts:read permission has been created/verified")
	fmt.Println("- Admin role has been granted alerts:read permission")
	fmt.Println("- All admin users should now be able to access the alerts page")
	fmt.Println("- If you're still getting authorization errors, try logging out and back in")
}
