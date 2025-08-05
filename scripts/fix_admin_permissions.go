package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "host=localhost port=5432 user=postgres password=pwaiswa dbname=ios sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		log.Fatal("Error pinging database:", err)
	}

	fmt.Println("Connected to database successfully")

	// Add missing permissions for admin role
	fmt.Println("Adding missing permissions for admin role...")

	// First, let's check what permissions exist for users resource
	var permissions []struct {
		ID       int
		Resource string
		Action   string
	}

	rows, err := db.QueryContext(ctx, `
		SELECT id, resource, action 
		FROM permissions 
		WHERE resource = 'users' 
		ORDER BY action
	`)
	if err != nil {
		log.Fatal("Error querying permissions:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var perm struct {
			ID       int
			Resource string
			Action   string
		}
		if err := rows.Scan(&perm.ID, &perm.Resource, &perm.Action); err != nil {
			log.Fatal("Error scanning permission:", err)
		}
		permissions = append(permissions, perm)
	}

	fmt.Printf("Found %d user permissions:\n", len(permissions))
	for _, perm := range permissions {
		fmt.Printf("  - %s: %s (ID: %d)\n", perm.Resource, perm.Action, perm.ID)
	}

	// Get admin role ID
	var adminRoleID int
	err = db.QueryRowContext(ctx, "SELECT id FROM roles WHERE name = 'admin'").Scan(&adminRoleID)
	if err != nil {
		log.Fatal("Error getting admin role ID:", err)
	}
	fmt.Printf("Admin role ID: %d\n", adminRoleID)

	// Check current admin permissions
	fmt.Println("\nChecking current admin permissions...")
	rows, err = db.QueryContext(ctx, `
		SELECT p.resource, p.action 
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.resource, p.action
	`, adminRoleID)
	if err != nil {
		log.Fatal("Error querying admin permissions:", err)
	}
	defer rows.Close()

	var currentPerms []string
	for rows.Next() {
		var resource, action string
		if err := rows.Scan(&resource, &action); err != nil {
			log.Fatal("Error scanning admin permission:", err)
		}
		currentPerms = append(currentPerms, fmt.Sprintf("%s:%s", resource, action))
	}

	fmt.Printf("Current admin permissions (%d):\n", len(currentPerms))
	for _, perm := range currentPerms {
		fmt.Printf("  - %s\n", perm)
	}

	// Add missing user permissions to admin role
	fmt.Println("\nAdding missing user permissions to admin role...")
	for _, perm := range permissions {
		permKey := fmt.Sprintf("%s:%s", perm.Resource, perm.Action)

		// Check if permission already exists
		exists := false
		for _, currentPerm := range currentPerms {
			if currentPerm == permKey {
				exists = true
				break
			}
		}

		if !exists {
			fmt.Printf("Adding permission: %s\n", permKey)
			_, err := db.ExecContext(ctx, `
				INSERT INTO role_permissions (role_id, permission_id, created_at)
				VALUES ($1, $2, NOW())
				ON CONFLICT (role_id, permission_id) DO NOTHING
			`, adminRoleID, perm.ID)
			if err != nil {
				log.Printf("Error adding permission %s: %v\n", permKey, err)
			} else {
				fmt.Printf("  ✓ Added %s\n", permKey)
			}
		} else {
			fmt.Printf("  - Permission %s already exists\n", permKey)
		}
	}

	// Verify the changes
	fmt.Println("\nVerifying admin permissions after update...")
	rows, err = db.QueryContext(ctx, `
		SELECT p.resource, p.action 
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.resource, p.action
	`, adminRoleID)
	if err != nil {
		log.Fatal("Error querying updated admin permissions:", err)
	}
	defer rows.Close()

	var updatedPerms []string
	for rows.Next() {
		var resource, action string
		if err := rows.Scan(&resource, &action); err != nil {
			log.Fatal("Error scanning updated admin permission:", err)
		}
		updatedPerms = append(updatedPerms, fmt.Sprintf("%s:%s", resource, action))
	}

	fmt.Printf("Updated admin permissions (%d):\n", len(updatedPerms))
	for _, perm := range updatedPerms {
		fmt.Printf("  - %s\n", perm)
	}

	// Check if admin user exists and has admin role
	fmt.Println("\nChecking admin user...")
	var adminUserID int
	err = db.QueryRowContext(ctx, "SELECT user_id FROM users WHERE user_name = 'admin'").Scan(&adminUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Admin user not found. Creating admin user...")
			// Create admin user
			err = db.QueryRowContext(ctx, `
				INSERT INTO users (user_name, user_pass, email, first_name, last_name, is_active, password_salt)
				VALUES ('admin', 'hashed_password_here', 'admin@system.local', 'System', 'Administrator', true, 'salt_here')
				RETURNING user_id
			`).Scan(&adminUserID)
			if err != nil {
				log.Fatal("Error creating admin user:", err)
			}
			fmt.Printf("Created admin user with ID: %d\n", adminUserID)
		} else {
			log.Fatal("Error checking admin user:", err)
		}
	} else {
		fmt.Printf("Admin user found with ID: %d\n", adminUserID)
	}

	// Check if admin user has admin role
	var hasAdminRole bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM user_roles ur
			JOIN roles r ON ur.role_id = r.id
			WHERE ur.user_id = $1 AND r.name = 'admin'
		)
	`, adminUserID).Scan(&hasAdminRole)
	if err != nil {
		log.Fatal("Error checking admin user role:", err)
	}

	if !hasAdminRole {
		fmt.Println("Assigning admin role to admin user...")
		_, err := db.ExecContext(ctx, `
			INSERT INTO user_roles (user_id, role_id, created_at)
			VALUES ($1, $2, NOW())
			ON CONFLICT (user_id, role_id) DO NOTHING
		`, adminUserID, adminRoleID)
		if err != nil {
			log.Fatal("Error assigning admin role to user:", err)
		}
		fmt.Println("✓ Admin role assigned to admin user")
	} else {
		fmt.Println("Admin user already has admin role")
	}

	fmt.Println("\n✅ Admin permissions fix completed successfully!")
	fmt.Println("The admin user should now be able to access the RBAC dashboard and manage user assignments.")
}
