package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection
	dsn := "host=localhost port=5432 user=postgres password=pwaiswa dbname=ios sslmode=disable"

	// Connect to database
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Error pinging database:", err)
	}
	fmt.Println("✓ Connected to database successfully")

	// Get all roles
	fmt.Println("\n=== ALL ROLES ===")
	rolesQuery := `
		SELECT id, name, description, is_active 
		FROM roles 
		ORDER BY id
	`

	roles, err := db.Query(rolesQuery)
	if err != nil {
		log.Fatal("Error querying roles:", err)
	}
	defer roles.Close()

	for roles.Next() {
		var id int
		var name, description sql.NullString
		var isActive bool

		err := roles.Scan(&id, &name, &description, &isActive)
		if err != nil {
			fmt.Printf("Error scanning role: %v\n", err)
			continue
		}

		status := "Active"
		if !isActive {
			status = "Inactive"
		}

		desc := "No description"
		if description.Valid {
			desc = description.String
		}

		fmt.Printf("Role ID: %d | Name: %s | Status: %s | Description: %s\n",
			id, name.String, status, desc)
	}

	// Get all permissions
	fmt.Println("\n=== ALL PERMISSIONS ===")
	permissionsQuery := `
		SELECT id, name, description, resource, action, is_active 
		FROM permissions 
		ORDER BY resource, action
	`

	permissions, err := db.Query(permissionsQuery)
	if err != nil {
		log.Fatal("Error querying permissions:", err)
	}
	defer permissions.Close()

	for permissions.Next() {
		var id int
		var name, description, resource, action sql.NullString
		var isActive bool

		err := permissions.Scan(&id, &name, &description, &resource, &action, &isActive)
		if err != nil {
			fmt.Printf("Error scanning permission: %v\n", err)
			continue
		}

		status := "Active"
		if !isActive {
			status = "Inactive"
		}

		desc := "No description"
		if description.Valid {
			desc = description.String
		}

		fmt.Printf("Permission ID: %d | Name: %s | Resource: %s | Action: %s | Status: %s | Description: %s\n",
			id, name.String, resource.String, action.String, status, desc)
	}

	// Get role-permission mappings
	fmt.Println("\n=== ROLE-PERMISSION MAPPINGS ===")
	mappingsQuery := `
		SELECT r.name as role_name, p.name as permission_name, p.resource, p.action
		FROM role_permissions rp
		JOIN roles r ON rp.role_id = r.id
		JOIN permissions p ON rp.permission_id = p.id
		WHERE r.is_active = true AND p.is_active = true
		ORDER BY r.name, p.resource, p.action
	`

	mappings, err := db.Query(mappingsQuery)
	if err != nil {
		log.Fatal("Error querying role-permission mappings:", err)
	}
	defer mappings.Close()

	currentRole := ""
	for mappings.Next() {
		var roleName, permissionName, resource, action sql.NullString

		err := mappings.Scan(&roleName, &permissionName, &resource, &action)
		if err != nil {
			fmt.Printf("Error scanning mapping: %v\n", err)
			continue
		}

		if roleName.String != currentRole {
			fmt.Printf("\n--- Role: %s ---\n", roleName.String)
			currentRole = roleName.String
		}

		fmt.Printf("  • %s (%s:%s)\n", permissionName.String, resource.String, action.String)
	}

	// Get user-role assignments
	fmt.Println("\n=== USER-ROLE ASSIGNMENTS ===")
	userRolesQuery := `
		SELECT u.user_name, u.first_name, u.last_name, r.name as role_name
		FROM user_roles ur
		JOIN users u ON ur.user_id = u.user_id
		JOIN roles r ON ur.role_id = r.id
		WHERE u.is_active = true AND r.is_active = true
		ORDER BY u.user_name, r.name
	`

	userRoles, err := db.Query(userRolesQuery)
	if err != nil {
		log.Fatal("Error querying user-role assignments:", err)
	}
	defer userRoles.Close()

	for userRoles.Next() {
		var username, firstName, lastName, roleName sql.NullString

		err := userRoles.Scan(&username, &firstName, &lastName, &roleName)
		if err != nil {
			fmt.Printf("Error scanning user-role: %v\n", err)
			continue
		}

		fullName := "Unknown"
		if firstName.Valid && lastName.Valid {
			fullName = firstName.String + " " + lastName.String
		}

		fmt.Printf("User: %s (%s) | Role: %s\n",
			username.String, fullName, roleName.String)
	}

	fmt.Println("\n✓ Role and permission analysis completed!")
}
