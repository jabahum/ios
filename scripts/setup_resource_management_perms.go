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
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "ios")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Connected to database successfully")

	// Create permissions
	permissions := []struct {
		action      string
		description string
	}{
		{"read", "View resource management dashboard and lists"},
		{"create", "Create new resources, teams, deployments, requisitions, etc."},
		{"update", "Update existing resources, approve/reject proposals, etc."},
		{"delete", "Delete resources, teams, deployments, etc."},
	}

	log.Println("Creating resource_management permissions...")
	for _, perm := range permissions {
		var exists bool
		err := db.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM permissions 
				WHERE resource = 'resource_management' AND action = $1
			)
		`, perm.action).Scan(&exists)

		if err != nil {
			log.Printf("Error checking permission existence: %v", err)
			continue
		}

		if !exists {
			name := "Resource management " + perm.action
			_, err := db.Exec(`
				INSERT INTO permissions (name, resource, action, description)
				VALUES ($1, 'resource_management', $2, $3)
			`, name, perm.action, perm.description)

			if err != nil {
				log.Printf("Error creating permission %s: %v", perm.action, err)
			} else {
				log.Printf("✓ Created permission: resource_management:%s", perm.action)
			}
		} else {
			log.Printf("  Permission already exists: resource_management:%s", perm.action)
		}
	}

	// Assign to roles
	type RolePermissions struct {
		roleName string
		actions  []string
	}

	roleAssignments := []RolePermissions{
		{"super_admin", []string{"read", "create", "update", "delete"}},
		{"admin", []string{"read", "create", "update", "delete"}},
		{"outbreak_coordinator", []string{"read", "create"}},
		{"case_manager", []string{"read"}},
	}

	log.Println("\nAssigning permissions to roles...")
	for _, ra := range roleAssignments {
		// Get role ID
		var roleID int
		err := db.QueryRow("SELECT id FROM roles WHERE name = $1", ra.roleName).Scan(&roleID)
		if err != nil {
			log.Printf("Warning: Role '%s' not found, skipping...", ra.roleName)
			continue
		}

		for _, action := range ra.actions {
			// Get permission ID
			var permID int
			err := db.QueryRow(`
				SELECT id FROM permissions 
				WHERE resource = 'resource_management' AND action = $1
			`, action).Scan(&permID)

			if err != nil {
				log.Printf("Error getting permission ID for %s: %v", action, err)
				continue
			}

			// Check if assignment exists
			var exists bool
			err = db.QueryRow(`
				SELECT EXISTS(
					SELECT 1 FROM role_permissions 
					WHERE role_id = $1 AND permission_id = $2
				)
			`, roleID, permID).Scan(&exists)

			if err != nil {
				log.Printf("Error checking role permission: %v", err)
				continue
			}

			if !exists {
				_, err := db.Exec(`
					INSERT INTO role_permissions (role_id, permission_id)
					VALUES ($1, $2)
				`, roleID, permID)

				if err != nil {
					log.Printf("Error assigning permission: %v", err)
				} else {
					log.Printf("✓ Assigned %s:%s to '%s'", "resource_management", action, ra.roleName)
				}
			} else {
				log.Printf("  Already assigned: %s:%s to '%s'", "resource_management", action, ra.roleName)
			}
		}
	}

	// Display results
	log.Println("\n=== Resource Management Permissions Summary ===")
	rows, err := db.Query(`
		SELECT 
			r.name AS role_name,
			p.resource,
			p.action,
			p.description
		FROM roles r
		JOIN role_permissions rp ON r.id = rp.role_id
		JOIN permissions p ON p.id = rp.permission_id
		WHERE p.resource = 'resource_management'
		ORDER BY r.name, p.action
	`)

	if err != nil {
		log.Fatalf("Error querying permissions: %v", err)
	}
	defer rows.Close()

	fmt.Println("\nRole | Resource | Action | Description")
	fmt.Println("-----|----------|--------|------------")

	for rows.Next() {
		var roleName, resource, action, description string
		if err := rows.Scan(&roleName, &resource, &action, &description); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		fmt.Printf("%s | %s | %s | %s\n", roleName, resource, action, description)
	}

	log.Println("\n✅ Resource management permissions setup completed successfully!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

