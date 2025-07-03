package models

import (
	"context"
	"database/sql"
	"time"
)

// Role represents a user role in the system
type Role struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Permission represents a system permission
type Permission struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Resource    string    `json:"resource"` // e.g., "vhf_patients", "users", "reports"
	Action      string    `json:"action"`   // e.g., "create", "read", "update", "delete"
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RolePermission represents the many-to-many relationship between roles and permissions
type RolePermission struct {
	ID           int       `json:"id"`
	RoleID       int       `json:"role_id"`
	PermissionID int       `json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// UserRole represents the many-to-many relationship between users and roles
type UserRole struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	RoleID    int       `json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Department represents organizational departments
type Department struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserSession represents user login sessions
type UserSession struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	SessionToken string    `json:"session_token"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	IsActive     bool      `json:"is_active"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// AuditLog represents system audit trail
type AuditLog struct {
	ID         int           `json:"id"`
	UserID     sql.NullInt64 `json:"user_id"`
	Action     string        `json:"action"`   // e.g., "login", "create", "update", "delete"
	Resource   string        `json:"resource"` // e.g., "user", "vhf_patient"
	ResourceID sql.NullInt64 `json:"resource_id"`
	Details    string        `json:"details"` // JSON details of the action
	IPAddress  string        `json:"ip_address"`
	UserAgent  string        `json:"user_agent"`
	CreatedAt  time.Time     `json:"created_at"`
}

// Permission constants
const (
	// Resource types
	ResourceUsers        = "users"
	ResourceVHFPatients  = "vhf_patients"
	ResourceReports      = "reports"
	ResourceOutbreaks    = "outbreaks"
	ResourceEmployees    = "employees"
	ResourceFacilities   = "facilities"
	ResourceLaboratory   = "laboratory"
	ResourceSurveillance = "surveillance"

	// Action types
	ActionCreate = "create"
	ActionRead   = "read"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionExport = "export"
	ActionImport = "import"
)

// Predefined roles
const (
	RoleSuperAdmin   = "super_admin"
	RoleAdmin        = "admin"
	RoleManager      = "manager"
	RoleDataEntry    = "data_entry"
	RoleViewer       = "viewer"
	RoleLabTech      = "lab_technician"
	RoleSurveillance = "surveillance_officer"
)

// Database operations for RBAC
func CreateRole(ctx context.Context, db *sql.DB, role *Role) error {
	query := `
		INSERT INTO roles (name, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	now := time.Now()
	return db.QueryRowContext(ctx, query,
		role.Name, role.Description, role.IsActive, now, now).Scan(&role.ID)
}

func GetRoleByID(ctx context.Context, db *sql.DB, id int) (*Role, error) {
	query := `SELECT id, name, description, is_active, created_at, updated_at 
			  FROM roles WHERE id = $1`

	role := &Role{}
	err := db.QueryRowContext(ctx, query, id).Scan(
		&role.ID, &role.Name, &role.Description, &role.IsActive,
		&role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func GetUserRoles(ctx context.Context, db *sql.DB, userID int) ([]Role, error) {
	query := `
		SELECT r.id, r.name, r.description, r.is_active, r.created_at, r.updated_at
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND r.is_active = true
		ORDER BY r.name`

	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		err := rows.Scan(&role.ID, &role.Name, &role.Description,
			&role.IsActive, &role.CreatedAt, &role.UpdatedAt)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func GetUserPermissions(ctx context.Context, db *sql.DB, userID int) ([]Permission, error) {
	query := `
		SELECT DISTINCT p.id, p.name, p.description, p.resource, p.action, 
		       p.is_active, p.created_at, p.updated_at
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1 AND p.is_active = true
		ORDER BY p.resource, p.action`

	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []Permission
	for rows.Next() {
		var perm Permission
		err := rows.Scan(&perm.ID, &perm.Name, &perm.Description,
			&perm.Resource, &perm.Action, &perm.IsActive,
			&perm.CreatedAt, &perm.UpdatedAt)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}
	return permissions, nil
}

func HasPermission(ctx context.Context, db *sql.DB, userID int, resource, action string) (bool, error) {
	query := `
		SELECT COUNT(*) > 0
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1 AND p.resource = $2 AND p.action = $3 
		  AND p.is_active = true`

	var hasPermission bool
	err := db.QueryRowContext(ctx, query, userID, resource, action).Scan(&hasPermission)
	return hasPermission, err
}

func LogAuditEvent(ctx context.Context, db *sql.DB, log *AuditLog) error {
	query := `
		INSERT INTO audit_logs (user_id, action, resource, resource_id, details, 
		                       ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := db.ExecContext(ctx, query, log.UserID, log.Action, log.Resource,
		log.ResourceID, log.Details, log.IPAddress, log.UserAgent, log.CreatedAt)
	return err
}
