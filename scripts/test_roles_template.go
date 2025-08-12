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

	// Test the exact query used in the users handler
	fmt.Println("\n🔍 Testing the exact query from users handler:")
	rolesQuery := `SELECT id, name FROM roles ORDER BY name`
	rows, err := db.QueryContext(context.Background(), rolesQuery)
	if err != nil {
		log.Fatal("Error querying roles:", err)
	}
	defer rows.Close()

	var roles []struct {
		ID   int
		Name string
	}

	for rows.Next() {
		var role struct {
			ID   int
			Name string
		}
		err := rows.Scan(&role.ID, &role.Name)
		if err != nil {
			log.Printf("Error scanning role: %v", err)
			continue
		}
		roles = append(roles, role)
		fmt.Printf("Role: ID=%d, Name=%s\n", role.ID, role.Name)
	}

	fmt.Printf("\n📊 Total roles found: %d\n", len(roles))

	// Test the query used in the user form handler
	fmt.Println("\n🔍 Testing the exact query from user form handler:")
	formRolesQuery := `SELECT id, name, COALESCE(description, '') as description FROM roles ORDER BY name`
	formRows, err := db.QueryContext(context.Background(), formRolesQuery)
	if err != nil {
		log.Fatal("Error querying form roles:", err)
	}
	defer formRows.Close()

	var formRoles []struct {
		ID          int
		Name        string
		Description string
	}

	for formRows.Next() {
		var role struct {
			ID          int
			Name        string
			Description string
		}
		err := formRows.Scan(&role.ID, &role.Name, &role.Description)
		if err != nil {
			log.Printf("Error scanning form role: %v", err)
			continue
		}
		formRoles = append(formRoles, role)
		fmt.Printf("Form Role: ID=%d, Name=%s, Description=%s\n", role.ID, role.Name, role.Description)
	}

	fmt.Printf("\n📊 Total form roles found: %d\n", len(formRoles))

	// Check if there are any differences
	if len(roles) != len(formRoles) {
		fmt.Printf("\n⚠️  WARNING: Different number of roles found!\n")
		fmt.Printf("List handler: %d roles\n", len(roles))
		fmt.Printf("Form handler: %d roles\n", len(formRoles))
	} else {
		fmt.Printf("\n✅ Both queries return the same number of roles\n")
	}

	// Test if the roles table has any inactive roles
	fmt.Println("\n🔍 Checking for inactive roles:")
	inactiveQuery := `SELECT id, name, is_active FROM roles WHERE is_active = false ORDER BY name`
	inactiveRows, err := db.QueryContext(context.Background(), inactiveQuery)
	if err != nil {
		log.Printf("Error querying inactive roles: %v", err)
	} else {
		defer inactiveRows.Close()

		inactiveCount := 0
		for inactiveRows.Next() {
			var id int
			var name string
			var isActive bool
			err := inactiveRows.Scan(&id, &name, &isActive)
			if err != nil {
				log.Printf("Error scanning inactive role: %v", err)
				continue
			}
			inactiveCount++
			fmt.Printf("Inactive Role: ID=%d, Name=%s, Active=%t\n", id, name, isActive)
		}

		if inactiveCount == 0 {
			fmt.Println("✅ No inactive roles found")
		} else {
			fmt.Printf("⚠️  Found %d inactive roles\n", inactiveCount)
		}
	}

	fmt.Println("\n🎯 Summary:")
	fmt.Printf("- Total roles in database: %d\n", len(roles))
	fmt.Printf("- Roles available for dropdown: %d\n", len(roles))
	fmt.Printf("- Form roles with descriptions: %d\n", len(formRoles))

	if len(roles) > 0 {
		fmt.Println("✅ Roles should appear in dropdown")
	} else {
		fmt.Println("❌ No roles found - dropdown will be empty")
	}
}
