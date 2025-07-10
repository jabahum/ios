package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// Config represents the application configuration
type Config struct {
	Port         string `json:"Port"`
	ReadTimeout  int    `json:"ReadTimeout"`
	WriteTimeout int    `json:"WriteTimeout"`
	Static       string `json:"Static"`
	Ux           string `json:"Ux"` // Username
	Px           string `json:"Px"` // Password
	Dx           string `json:"Dx"` // Database
	LogFile      string `json:"LogFile"`
	LogData      string `json:"LogData"`
	Facility     string `json:"Facility"`
	SMSUser      string `json:"SMSUser"`
	SMSPassword  string `json:"SMSPassword"`
	SMSURL       string `json:"SMSURL"`
}

// UserRightData represents the structure of user_right table
type UserRightData struct {
	UserRightsID        int
	UserID              int
	FunctionScope       int
	UserRightsFunction  int
	UserRightsCanCreate int
	UserRightsCanView   int
	UserRightsCanEdit   int
	UserRightsCanRemove int
}

// MetaData represents the meta table structure
type MetaData struct {
	MetaID          int
	MetaCategory    int
	MetaName        string
	MetaOrder       int
	MetaDescription sql.NullString
	MetaLink        sql.NullString
}

// MigrationResult tracks migration statistics
type MigrationResult struct {
	TotalUserRights      int
	UniquePermissionSets int
	RolesCreated         int
	UsersMigrated        int
	Errors               []string
}

func main() {
	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatal("Error loading config:", err)
	}

	// Build database connection string
	connStr := fmt.Sprintf("postgres://%s:%s@localhost:5432/%s?sslmode=disable",
		config.Ux, config.Px, config.Dx)

	// Database connection
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

	fmt.Println("Connected to database successfully")
	fmt.Printf("Using database: %s\n", config.Dx)
	fmt.Println("Starting migration from user_right to RBAC system...")

	// Run migration
	result, err := migrateUserRightsToRBAC(db)
	if err != nil {
		log.Fatal("Migration failed:", err)
	}

	// Print results
	fmt.Println("\n=== Migration Results ===")
	fmt.Printf("Total user_right entries processed: %d\n", result.TotalUserRights)
	fmt.Printf("Unique permission sets identified: %d\n", result.UniquePermissionSets)
	fmt.Printf("Roles created: %d\n", result.RolesCreated)
	fmt.Printf("Users migrated: %d\n", result.UsersMigrated)

	if len(result.Errors) > 0 {
		fmt.Printf("\nErrors encountered: %d\n", len(result.Errors))
		for i, err := range result.Errors {
			fmt.Printf("  %d. %s\n", i+1, err)
		}
	}

	fmt.Println("\nMigration completed successfully!")
}

func loadConfig() (*Config, error) {
	// Read config file
	configData, err := os.ReadFile("cmd/web/config.json")
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var config Config
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	return &config, nil
}

func migrateUserRightsToRBAC(db *sql.DB) (*MigrationResult, error) {
	result := &MigrationResult{}

	// 1. Get all user_right data
	fmt.Println("Step 1: Analyzing user_right data...")
	userRights, err := getAllUserRights(db)
	if err != nil {
		return nil, fmt.Errorf("error getting user_right data: %v", err)
	}
	result.TotalUserRights = len(userRights)

	// 2. Get meta data for function mapping
	fmt.Println("Step 2: Loading meta data...")
	metaData, err := getAllMetaData(db)
	if err != nil {
		return nil, fmt.Errorf("error getting meta data: %v", err)
	}

	// 3. Analyze permission patterns and create roles
	fmt.Println("Step 3: Analyzing permission patterns...")
	permissionSets := analyzePermissionPatterns(userRights, metaData)
	result.UniquePermissionSets = len(permissionSets)

	// 4. Create roles based on permission patterns
	fmt.Println("Step 4: Creating roles...")
	roleMap, err := createRolesFromPatterns(db, permissionSets)
	if err != nil {
		return nil, fmt.Errorf("error creating roles: %v", err)
	}
	result.RolesCreated = len(roleMap)

	// 5. Assign users to roles
	fmt.Println("Step 5: Assigning users to roles...")
	migratedUsers, err := assignUsersToRoles(db, userRights, roleMap)
	if err != nil {
		return nil, fmt.Errorf("error assigning users to roles: %v", err)
	}
	result.UsersMigrated = migratedUsers

	// 6. Create migration audit log
	fmt.Println("Step 6: Creating migration audit log...")
	err = createMigrationAuditLog(db, result)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to create audit log: %v", err))
	}

	return result, nil
}

func getAllUserRights(db *sql.DB) ([]UserRightData, error) {
	query := `
		SELECT user_rights_id, user_id, function_scope, user_rights_function,
		       user_rights_can_create, user_rights_can_view, user_rights_can_edit, user_rights_can_remove
		FROM user_right
		WHERE user_rights_can_create + user_rights_can_view + user_rights_can_edit + user_rights_can_remove > 0
		ORDER BY user_id, user_rights_function
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userRights []UserRightData
	for rows.Next() {
		var ur UserRightData
		err := rows.Scan(
			&ur.UserRightsID, &ur.UserID, &ur.FunctionScope, &ur.UserRightsFunction,
			&ur.UserRightsCanCreate, &ur.UserRightsCanView, &ur.UserRightsCanEdit, &ur.UserRightsCanRemove,
		)
		if err != nil {
			return nil, err
		}
		userRights = append(userRights, ur)
	}

	return userRights, nil
}

func getAllMetaData(db *sql.DB) (map[int]MetaData, error) {
	query := `
		SELECT meta_id, meta_category, meta_name, meta_order, meta_description, meta_link
		FROM meta
		ORDER BY meta_id
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metaMap := make(map[int]MetaData)
	for rows.Next() {
		var meta MetaData
		err := rows.Scan(
			&meta.MetaID, &meta.MetaCategory, &meta.MetaName,
			&meta.MetaOrder, &meta.MetaDescription, &meta.MetaLink,
		)
		if err != nil {
			return nil, err
		}
		metaMap[meta.MetaID] = meta
	}

	return metaMap, nil
}

// PermissionPattern represents a unique set of permissions
type PermissionPattern struct {
	FunctionID   int
	FunctionName string
	CanCreate    bool
	CanView      bool
	CanEdit      bool
	CanRemove    bool
	UserCount    int
	UserIDs      []int
}

func analyzePermissionPatterns(userRights []UserRightData, metaData map[int]MetaData) map[string]PermissionPattern {
	patterns := make(map[string]PermissionPattern)

	for _, ur := range userRights {
		meta, exists := metaData[ur.UserRightsFunction]
		if !exists {
			continue
		}

		// Create a unique key for this permission pattern
		key := fmt.Sprintf("%d_%d_%d_%d_%d",
			ur.UserRightsFunction, ur.UserRightsCanCreate, ur.UserRightsCanView,
			ur.UserRightsCanEdit, ur.UserRightsCanRemove)

		pattern, exists := patterns[key]
		if !exists {
			pattern = PermissionPattern{
				FunctionID:   ur.UserRightsFunction,
				FunctionName: meta.MetaName,
				CanCreate:    ur.UserRightsCanCreate == 1,
				CanView:      ur.UserRightsCanView == 1,
				CanEdit:      ur.UserRightsCanEdit == 1,
				CanRemove:    ur.UserRightsCanRemove == 1,
				UserIDs:      []int{},
			}
		}

		pattern.UserCount++
		pattern.UserIDs = append(pattern.UserIDs, ur.UserID)
		patterns[key] = pattern
	}

	return patterns
}

func createRolesFromPatterns(db *sql.DB, patterns map[string]PermissionPattern) (map[string]int, error) {
	roleMap := make(map[string]int)

	for key, pattern := range patterns {
		// Create role name based on function and permissions
		roleName := generateRoleName(pattern)

		// Check if role already exists
		var roleID int
		err := db.QueryRow("SELECT id FROM roles WHERE name = $1", roleName).Scan(&roleID)
		if err == sql.ErrNoRows {
			// Create new role
			description := generateRoleDescription(pattern)
			err = db.QueryRow(`
				INSERT INTO roles (name, description, is_active, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5)
				RETURNING id
			`, roleName, description, true, time.Now(), time.Now()).Scan(&roleID)
			if err != nil {
				return nil, fmt.Errorf("error creating role %s: %v", roleName, err)
			}
			fmt.Printf("  Created role: %s (ID: %d)\n", roleName, roleID)
		} else if err != nil {
			return nil, fmt.Errorf("error checking role %s: %v", roleName, err)
		}

		roleMap[key] = roleID

		// Assign permissions to role
		err = assignPermissionsToRole(db, roleID, pattern)
		if err != nil {
			return nil, fmt.Errorf("error assigning permissions to role %s: %v", roleName, err)
		}
	}

	return roleMap, nil
}

func generateRoleName(pattern PermissionPattern) string {
	baseName := strings.ReplaceAll(pattern.FunctionName, " ", "_")
	baseName = strings.ToLower(baseName)

	var permissions []string
	if pattern.CanCreate {
		permissions = append(permissions, "create")
	}
	if pattern.CanView {
		permissions = append(permissions, "view")
	}
	if pattern.CanEdit {
		permissions = append(permissions, "edit")
	}
	if pattern.CanRemove {
		permissions = append(permissions, "delete")
	}

	if len(permissions) == 4 {
		return fmt.Sprintf("%s_full_access", baseName)
	} else if len(permissions) == 1 && pattern.CanView {
		return fmt.Sprintf("%s_viewer", baseName)
	} else {
		return fmt.Sprintf("%s_%s", baseName, strings.Join(permissions, "_"))
	}
}

func generateRoleDescription(pattern PermissionPattern) string {
	var actions []string
	if pattern.CanCreate {
		actions = append(actions, "create")
	}
	if pattern.CanView {
		actions = append(actions, "view")
	}
	if pattern.CanEdit {
		actions = append(actions, "edit")
	}
	if pattern.CanRemove {
		actions = append(actions, "delete")
	}

	return fmt.Sprintf("Access to %s with permissions: %s",
		pattern.FunctionName, strings.Join(actions, ", "))
}

func assignPermissionsToRole(db *sql.DB, roleID int, pattern PermissionPattern) error {
	// Map function to resource name
	resourceName := mapFunctionToResource(pattern.FunctionName)

	// Get or create permissions for this role
	permissions := []struct {
		action string
		has    bool
	}{
		{"create", pattern.CanCreate},
		{"read", pattern.CanView},
		{"update", pattern.CanEdit},
		{"delete", pattern.CanRemove},
	}

	for _, perm := range permissions {
		if !perm.has {
			continue
		}

		// Get permission ID
		var permissionID int
		err := db.QueryRow(`
			SELECT id FROM permissions 
			WHERE resource = $1 AND action = $2
		`, resourceName, perm.action).Scan(&permissionID)

		if err == sql.ErrNoRows {
			// Create permission if it doesn't exist
			permName := fmt.Sprintf("%s %s", strings.Title(perm.action), resourceName)
			err = db.QueryRow(`
				INSERT INTO permissions (name, description, resource, action, is_active, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				RETURNING id
			`, permName, fmt.Sprintf("Can %s %s", perm.action, resourceName),
				resourceName, perm.action, true, time.Now(), time.Now()).Scan(&permissionID)
			if err != nil {
				return fmt.Errorf("error creating permission %s %s: %v", resourceName, perm.action, err)
			}
		} else if err != nil {
			return fmt.Errorf("error getting permission %s %s: %v", resourceName, perm.action, err)
		}

		// Assign permission to role
		_, err = db.Exec(`
			INSERT INTO role_permissions (role_id, permission_id, created_at)
			VALUES ($1, $2, $3)
			ON CONFLICT (role_id, permission_id) DO NOTHING
		`, roleID, permissionID, time.Now())
		if err != nil {
			return fmt.Errorf("error assigning permission to role: %v", err)
		}
	}

	return nil
}

func mapFunctionToResource(functionName string) string {
	// Map function names to resource names
	functionName = strings.ToLower(functionName)

	switch {
	case strings.Contains(functionName, "patient"):
		return "vhf_patients"
	case strings.Contains(functionName, "user"):
		return "users"
	case strings.Contains(functionName, "report"):
		return "reports"
	case strings.Contains(functionName, "outbreak"):
		return "outbreaks"
	case strings.Contains(functionName, "employee"):
		return "employees"
	case strings.Contains(functionName, "facility"):
		return "facilities"
	case strings.Contains(functionName, "lab"):
		return "laboratory"
	case strings.Contains(functionName, "surveillance"):
		return "surveillance"
	default:
		return strings.ReplaceAll(functionName, " ", "_")
	}
}

func assignUsersToRoles(db *sql.DB, userRights []UserRightData, roleMap map[string]int) (int, error) {
	// Group user rights by user
	userPermissions := make(map[int][]UserRightData)
	for _, ur := range userRights {
		userPermissions[ur.UserID] = append(userPermissions[ur.UserID], ur)
	}

	migratedUsers := 0
	for userID, permissions := range userPermissions {
		// Get or create user in users table
		var userExists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)", userID).Scan(&userExists)
		if err != nil {
			return migratedUsers, fmt.Errorf("error checking user %d: %v", userID, err)
		}

		if !userExists {
			// Skip users that don't exist in users table
			continue
		}

		// Remove existing role assignments for this user
		_, err = db.Exec("DELETE FROM user_roles WHERE user_id = $1", userID)
		if err != nil {
			return migratedUsers, fmt.Errorf("error removing existing roles for user %d: %v", userID, err)
		}

		// Assign roles based on user's permissions
		assignedRoles := make(map[int]bool)
		for _, permission := range permissions {
			// Create pattern key
			key := fmt.Sprintf("%d_%d_%d_%d_%d",
				permission.UserRightsFunction, permission.UserRightsCanCreate,
				permission.UserRightsCanView, permission.UserRightsCanEdit,
				permission.UserRightsCanRemove)

			roleID, exists := roleMap[key]
			if !exists {
				continue
			}

			// Assign role to user (avoid duplicates)
			if !assignedRoles[roleID] {
				_, err = db.Exec(`
					INSERT INTO user_roles (user_id, role_id, created_at)
					VALUES ($1, $2, $3)
					ON CONFLICT (user_id, role_id) DO NOTHING
				`, userID, roleID, time.Now())
				if err != nil {
					return migratedUsers, fmt.Errorf("error assigning role %d to user %d: %v", roleID, userID, err)
				}
				assignedRoles[roleID] = true
			}
		}

		if len(assignedRoles) > 0 {
			migratedUsers++
		}
	}

	return migratedUsers, nil
}

func createMigrationAuditLog(db *sql.DB, result *MigrationResult) error {
	// Create audit log entry
	_, err := db.Exec(`
		INSERT INTO audit_logs (user_id, action, resource, details, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, nil, "migration", "user_right_to_rbac",
		fmt.Sprintf(`{"total_user_rights": %d, "unique_patterns": %d, "roles_created": %d, "users_migrated": %d, "errors": %d}`,
			result.TotalUserRights, result.UniquePermissionSets, result.RolesCreated, result.UsersMigrated, len(result.Errors)),
		"migration_script", "migration_script", time.Now())

	return err
}
