package main

import (
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

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Error pinging database:", err)
	}

	fmt.Println("✅ Connected to database successfully!")

	// Check available permissions
	fmt.Println("\n📋 Available Permissions:")
	query := `
		SELECT DISTINCT p.resource, p.action
		FROM permissions p
		WHERE p.is_active = true
		ORDER BY p.resource, p.action
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("Error querying permissions:", err)
	}
	defer rows.Close()

	permissions := make(map[string][]string)
	for rows.Next() {
		var resource, action string
		err := rows.Scan(&resource, &action)
		if err != nil {
			continue
		}

		if permissions[resource] == nil {
			permissions[resource] = []string{}
		}
		permissions[resource] = append(permissions[resource], action)
	}

	for resource, actions := range permissions {
		fmt.Printf("  %s: %v\n", resource, actions)
	}

	// Check available roles
	fmt.Println("\n👥 Available Roles:")
	roleQuery := `
		SELECT id, name, description
		FROM roles
		WHERE is_active = true
		ORDER BY name
	`

	roleRows, err := db.Query(roleQuery)
	if err != nil {
		log.Fatal("Error querying roles:", err)
	}
	defer roleRows.Close()

	for roleRows.Next() {
		var id int
		var name, description string
		err := roleRows.Scan(&id, &name, &description)
		if err != nil {
			continue
		}
		fmt.Printf("  %d: %s - %s\n", id, name, description)
	}

	// Check user-role assignments
	fmt.Println("\n🔗 User-Role Assignments:")
	userRoleQuery := `
		SELECT u.username, r.name as role_name
		FROM users u
		JOIN user_roles ur ON u.id = ur.user_id
		JOIN roles r ON ur.role_id = r.id
		WHERE u.is_active = true AND r.is_active = true
		ORDER BY u.username, r.name
	`

	userRoleRows, err := db.Query(userRoleQuery)
	if err != nil {
		log.Fatal("Error querying user roles:", err)
	}
	defer userRoleRows.Close()

	for userRoleRows.Next() {
		var username, roleName string
		err := userRoleRows.Scan(&username, &roleName)
		if err != nil {
			continue
		}
		fmt.Printf("  %s: %s\n", username, roleName)
	}

	// Test specific user permissions
	fmt.Println("\n🔍 Testing User Permissions:")
	testUserQuery := `
		SELECT DISTINCT p.resource, p.action
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		JOIN users u ON ur.user_id = u.id
		WHERE u.username = $1 AND p.is_active = true
		ORDER BY p.resource, p.action
	`

	// Test with a specific username (change this to test different users)
	testUsername := "admin"
	rows, err = db.Query(testUserQuery, testUsername)
	if err != nil {
		log.Printf("Error querying user permissions for %s: %v", testUsername, err)
	} else {
		defer rows.Close()

		userPermissions := make(map[string][]string)
		for rows.Next() {
			var resource, action string
			err := rows.Scan(&resource, &action)
			if err != nil {
				continue
			}

			if userPermissions[resource] == nil {
				userPermissions[resource] = []string{}
			}
			userPermissions[resource] = append(userPermissions[resource], action)
		}

		fmt.Printf("  Permissions for user '%s':\n", testUsername)
		if len(userPermissions) == 0 {
			fmt.Printf("    No permissions found for user '%s'\n", testUsername)
		} else {
			for resource, actions := range userPermissions {
				fmt.Printf("    %s: %v\n", resource, actions)
			}
		}
	}

	fmt.Println("\n✅ Permission test completed!")
}
