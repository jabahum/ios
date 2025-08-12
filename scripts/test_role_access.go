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

	fmt.Println("✅ Connected to database successfully")

	// Test user ID 1 (paul.mbaka@gmail.com)
	userID := 1

	// Test role checking functions
	fmt.Println("\n🔍 Testing role access for user ID:", userID)

	// Test HasRole function
	hasAdmin := hasRole(db, userID, "admin")
	hasReports := hasRole(db, userID, "reports")
	hasSuperAdmin := hasRole(db, userID, "super_admin")

	fmt.Printf("Has 'admin' role: %t\n", hasAdmin)
	fmt.Printf("Has 'reports' role: %t\n", hasReports)
	fmt.Printf("Has 'super_admin' role: %t\n", hasSuperAdmin)

	// Test HasAnyRole function
	allowedRoles := []string{"admin", "reports", "vhf_lab_technician", "case_manager", "data_analyst"}
	hasAnyAllowed := hasAnyRole(db, userID, allowedRoles)
	fmt.Printf("Has any of the allowed roles: %t\n", hasAnyAllowed)

	// Test GetUserRoles function
	userRoles := getUserRoles(db, userID)
	fmt.Printf("User roles: %v\n", userRoles)

	// Test GetRoleID function
	adminRoleID := getRoleID(db, userID, "admin")
	reportsRoleID := getRoleID(db, userID, "reports")
	fmt.Printf("Admin role ID: %d\n", adminRoleID)
	fmt.Printf("Reports role ID: %d\n", reportsRoleID)

	fmt.Println("\n✅ Role access testing completed!")
}

// Helper functions (copied from security package for testing)
func hasRole(db *sql.DB, userID int, roleName string) bool {
	if userID <= 0 || roleName == "" {
		return false
	}

	query := `
		SELECT EXISTS(
			SELECT 1 
			FROM user_roles ur
			JOIN roles r ON ur.role_id = r.id
			WHERE ur.user_id = $1 AND r.name = $2 AND r.is_active = true
		)
	`

	var exists bool
	err := db.QueryRowContext(context.Background(), query, userID, roleName).Scan(&exists)
	if err != nil {
		return false
	}

	return exists
}

func getUserRoles(db *sql.DB, userID int) []string {
	if userID <= 0 {
		return []string{}
	}

	query := `
		SELECT r.name
		FROM user_roles ur
		JOIN roles r ON ur.role_id = r.id
		WHERE ur.user_id = $1 AND r.is_active = true
		ORDER BY r.name
	`

	rows, err := db.QueryContext(context.Background(), query, userID)
	if err != nil {
		return []string{}
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var roleName string
		if err := rows.Scan(&roleName); err == nil {
			roles = append(roles, roleName)
		}
	}

	return roles
}

func hasAnyRole(db *sql.DB, userID int, roleNames []string) bool {
	if userID <= 0 || len(roleNames) == 0 {
		return false
	}

	// Build the query with placeholders
	placeholders := make([]string, len(roleNames))
	args := make([]interface{}, len(roleNames)+1)
	args[0] = userID

	for i, roleName := range roleNames {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = roleName
	}

	query := fmt.Sprintf(`
		SELECT EXISTS(
			SELECT 1 
			FROM user_roles ur
			JOIN roles r ON ur.role_id = r.id
			WHERE ur.user_id = $1 AND r.name IN (%s) AND r.is_active = true
		)
	`, fmt.Sprintf("$%d", 2), fmt.Sprintf("$%d", 3), fmt.Sprintf("$%d", 4), fmt.Sprintf("$%d", 5), fmt.Sprintf("$%d", 6))

	var exists bool
	err := db.QueryRowContext(context.Background(), query, args...).Scan(&exists)
	if err != nil {
		return false
	}

	return exists
}

func getRoleID(db *sql.DB, userID int, roleName string) int {
	if userID <= 0 || roleName == "" {
		return 0
	}

	query := `
		SELECT r.id
		FROM user_roles ur
		JOIN roles r ON ur.role_id = r.id
		WHERE ur.user_id = $1 AND r.name = $2 AND r.is_active = true
		LIMIT 1
	`

	var roleID int
	err := db.QueryRowContext(context.Background(), query, userID, roleName).Scan(&roleID)
	if err != nil {
		return 0
	}

	return roleID
}
