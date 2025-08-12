package security

import (
	"context"
	"database/sql"
	"strings"
)

// GetRoles returns the role ID for a user and specific role name
func GetRoles(user int, roleName string) int {
	if user <= 0 || roleName == "" {
		return 0
	}

	// This function is used in handlers that expect a role ID
	// For now, return a placeholder that will be handled properly
	return 0
}

// GetRoleID returns the role ID for a user and specific role name
func GetRoleID(db *sql.DB, userID int, roleName string) int {
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

// HasRole checks if a user has a specific role by name
func HasRole(db *sql.DB, userID int, roleName string) bool {
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

// GetUserRoles returns a slice of role names for a user
func GetUserRoles(db *sql.DB, userID int) []string {
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

// HasAnyRole checks if a user has any of the specified roles
func HasAnyRole(db *sql.DB, userID int, roleNames []string) bool {
	if userID <= 0 || len(roleNames) == 0 {
		return false
	}

	// Build the query with placeholders
	placeholders := make([]string, len(roleNames))
	args := make([]interface{}, len(roleNames)+1)
	args[0] = userID

	for i, roleName := range roleNames {
		placeholders[i] = "$" + string(rune('2'+i))
		args[i+1] = roleName
	}

	query := `
		SELECT EXISTS(
			SELECT 1 
			FROM user_roles ur
			JOIN roles r ON ur.role_id = r.id
			WHERE ur.user_id = $1 AND r.name IN (` + strings.Join(placeholders, ",") + `) AND r.is_active = true
		)
	`

	var exists bool
	err := db.QueryRowContext(context.Background(), query, args...).Scan(&exists)
	if err != nil {
		return false
	}

	return exists
}
