package handlers

import (
	"case/internal/middleware"
	"case/internal/models"
	"database/sql"
	"encoding/json"
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/lib/pq"
)

// RBACManagementHandler handles role and permission management
type RBACManagementHandler struct {
	db     *sql.DB
	logger *slog.Logger
	store  *session.Store
	config Config
}

// NewRBACManagementHandler creates a new RBAC management handler
func NewRBACManagementHandler(db *sql.DB, logger *slog.Logger, store *session.Store, config Config) *RBACManagementHandler {
	return &RBACManagementHandler{
		db:     db,
		logger: logger,
		store:  store,
		config: config,
	}
}

// ==================== ROLE MANAGEMENT ====================

// ListRoles handles listing all roles with their permissions
func (h *RBACManagementHandler) ListRoles(c *fiber.Ctx) error {
	// TODO: Re-enable permission check when authentication is properly set up
	// if !middleware.UserHasPermission(c, models.ResourceUsers, models.ActionRead) {
	// 	return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	// }

	query := `
		SELECT r.id, r.name, r.description, r.is_active, r.created_at, r.updated_at,
		       COUNT(DISTINCT ur.user_id) as user_count,
		       COUNT(DISTINCT rp.permission_id) as permission_count
		FROM roles r
		LEFT JOIN user_roles ur ON r.id = ur.role_id
		LEFT JOIN role_permissions rp ON r.id = rp.role_id
		GROUP BY r.id, r.name, r.description, r.is_active, r.created_at, r.updated_at
		ORDER BY r.name
	`

	rows, err := h.db.QueryContext(c.Context(), query)
	if err != nil {
		h.logger.Error("Error querying roles", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	defer rows.Close()

	var roles []fiber.Map
	for rows.Next() {
		var role struct {
			ID              int
			Name            string
			Description     sql.NullString
			IsActive        bool
			CreatedAt       time.Time
			UpdatedAt       time.Time
			UserCount       int
			PermissionCount int
		}

		err := rows.Scan(
			&role.ID, &role.Name, &role.Description, &role.IsActive,
			&role.CreatedAt, &role.UpdatedAt, &role.UserCount, &role.PermissionCount,
		)
		if err != nil {
			h.logger.Error("Error scanning role", "error", err)
			continue
		}

		roles = append(roles, fiber.Map{
			"id":               role.ID,
			"name":             role.Name,
			"description":      role.Description.String,
			"is_active":        role.IsActive,
			"created_at":       role.CreatedAt,
			"updated_at":       role.UpdatedAt,
			"user_count":       role.UserCount,
			"permission_count": role.PermissionCount,
		})
	}

	return c.JSON(fiber.Map{"roles": roles})
}

// CreateRole handles creating a new role
func (h *RBACManagementHandler) CreateRole(c *fiber.Ctx) error {
	// Check permission
	if !middleware.UserHasPermission(c, models.ResourceUsers, models.ActionCreate) {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Role name is required"})
	}

	// Check if role already exists
	var exists bool
	err := h.db.QueryRowContext(c.Context(), `
		SELECT EXISTS(SELECT 1 FROM roles WHERE name = $1)
	`, req.Name).Scan(&exists)

	if err != nil {
		h.logger.Error("Error checking role existence", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	if exists {
		return c.Status(409).JSON(fiber.Map{"error": "Role already exists"})
	}

	// Insert new role
	var roleID int
	err = h.db.QueryRowContext(c.Context(), `
		INSERT INTO roles (name, description, is_active, created_at, updated_at)
		VALUES ($1, $2, true, NOW(), NOW())
		RETURNING id
	`, req.Name, req.Description).Scan(&roleID)

	if err != nil {
		h.logger.Error("Error creating role", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	// Log the action
	userID := GetCurrentUser(c, h.store)
	h.logger.Info("Role created", "user_id", userID, "role_id", roleID, "role_name", req.Name)

	return c.JSON(fiber.Map{
		"message": "Role created successfully",
		"role_id": roleID,
	})
}

// UpdateRole handles updating a role
func (h *RBACManagementHandler) UpdateRole(c *fiber.Ctx) error {
	// Check permission
	if !middleware.UserHasPermission(c, models.ResourceUsers, models.ActionUpdate) {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	roleID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid role ID"})
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		IsActive    bool   `json:"is_active"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Update role
	result, err := h.db.ExecContext(c.Context(), `
		UPDATE roles SET name = $1, description = $2, is_active = $3, updated_at = NOW()
		WHERE id = $4
	`, req.Name, req.Description, req.IsActive, roleID)

	if err != nil {
		h.logger.Error("Error updating role", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.Error("Error getting rows affected", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	if rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Role not found"})
	}

	// Log the action
	userID := GetCurrentUser(c, h.store)
	h.logger.Info("Role updated", "user_id", userID, "role_id", roleID)

	return c.JSON(fiber.Map{"message": "Role updated successfully"})
}

// DeleteRole handles deleting a role
func (h *RBACManagementHandler) DeleteRole(c *fiber.Ctx) error {
	// Check permission
	if !middleware.UserHasPermission(c, models.ResourceUsers, models.ActionDelete) {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	roleID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid role ID"})
	}

	// Check if role is assigned to any users
	var userCount int
	err = h.db.QueryRowContext(c.Context(), `
		SELECT COUNT(*) FROM user_roles WHERE role_id = $1
	`, roleID).Scan(&userCount)

	if err != nil {
		h.logger.Error("Error checking role usage", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	if userCount > 0 {
		return c.Status(409).JSON(fiber.Map{"error": "Cannot delete role that is assigned to users"})
	}

	// Delete role permissions first
	_, err = h.db.ExecContext(c.Context(), `
		DELETE FROM role_permissions WHERE role_id = $1
	`, roleID)

	if err != nil {
		h.logger.Error("Error deleting role permissions", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	// Delete role
	result, err := h.db.ExecContext(c.Context(), `
		DELETE FROM roles WHERE id = $1
	`, roleID)

	if err != nil {
		h.logger.Error("Error deleting role", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.Error("Error getting rows affected", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	if rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Role not found"})
	}

	// Log the action
	userID := GetCurrentUser(c, h.store)
	h.logger.Info("Role deleted", "user_id", userID, "role_id", roleID)

	return c.JSON(fiber.Map{"message": "Role deleted successfully"})
}

// ==================== PERMISSION MANAGEMENT ====================

// ListPermissions handles listing all permissions
func (h *RBACManagementHandler) ListPermissions(c *fiber.Ctx) error {
	// TODO: Re-enable permission check when authentication is properly set up
	// if !middleware.UserHasPermission(c, models.ResourceUsers, models.ActionRead) {
	// 	return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	// }

	query := `
		SELECT id, name, description, resource, action, is_active, created_at, updated_at
		FROM permissions
		ORDER BY resource, action
	`

	rows, err := h.db.QueryContext(c.Context(), query)
	if err != nil {
		h.logger.Error("Error querying permissions", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	defer rows.Close()

	var permissions []fiber.Map
	for rows.Next() {
		var perm struct {
			ID          int
			Name        string
			Description sql.NullString
			Resource    string
			Action      string
			IsActive    bool
			CreatedAt   time.Time
			UpdatedAt   time.Time
		}

		err := rows.Scan(
			&perm.ID, &perm.Name, &perm.Description, &perm.Resource, &perm.Action,
			&perm.IsActive, &perm.CreatedAt, &perm.UpdatedAt,
		)
		if err != nil {
			h.logger.Error("Error scanning permission", "error", err)
			continue
		}

		permissions = append(permissions, fiber.Map{
			"id":          perm.ID,
			"name":        perm.Name,
			"description": perm.Description.String,
			"resource":    perm.Resource,
			"action":      perm.Action,
			"is_active":   perm.IsActive,
			"created_at":  perm.CreatedAt,
			"updated_at":  perm.UpdatedAt,
		})
	}

	return c.JSON(fiber.Map{"permissions": permissions})
}

// CreatePermission handles creating a new permission
func (h *RBACManagementHandler) CreatePermission(c *fiber.Ctx) error {
	// Check permission
	if !middleware.UserHasPermission(c, models.ResourceUsers, models.ActionCreate) {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Resource    string `json:"resource"`
		Action      string `json:"action"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" || req.Resource == "" || req.Action == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Name, resource, and action are required"})
	}

	// Check if permission already exists
	var exists bool
	err := h.db.QueryRowContext(c.Context(), `
		SELECT EXISTS(SELECT 1 FROM permissions WHERE resource = $1 AND action = $2)
	`, req.Resource, req.Action).Scan(&exists)

	if err != nil {
		h.logger.Error("Error checking permission existence", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	if exists {
		return c.Status(409).JSON(fiber.Map{"error": "Permission already exists"})
	}

	// Insert new permission
	var permID int
	err = h.db.QueryRowContext(c.Context(), `
		INSERT INTO permissions (name, description, resource, action, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, true, NOW(), NOW())
		RETURNING id
	`, req.Name, req.Description, req.Resource, req.Action).Scan(&permID)

	if err != nil {
		h.logger.Error("Error creating permission", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	// Log the action
	userID := GetCurrentUser(c, h.store)
	h.logger.Info("Permission created", "user_id", userID, "permission_id", permID)

	return c.JSON(fiber.Map{
		"message":       "Permission created successfully",
		"permission_id": permID,
	})
}

// UpdatePermission handles updating a permission
func (h *RBACManagementHandler) UpdatePermission(c *fiber.Ctx) error {
	// Check permission
	if !middleware.UserHasPermission(c, models.ResourceUsers, models.ActionUpdate) {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	permID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid permission ID"})
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Resource    string `json:"resource"`
		Action      string `json:"action"`
		IsActive    bool   `json:"is_active"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Update permission
	result, err := h.db.ExecContext(c.Context(), `
		UPDATE permissions SET name = $1, description = $2, resource = $3, action = $4, is_active = $5, updated_at = NOW()
		WHERE id = $6
	`, req.Name, req.Description, req.Resource, req.Action, req.IsActive, permID)

	if err != nil {
		h.logger.Error("Error updating permission", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.Error("Error getting rows affected", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	if rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Permission not found"})
	}

	// Log the action
	userID := GetCurrentUser(c, h.store)
	h.logger.Info("Permission updated", "user_id", userID, "permission_id", permID)

	return c.JSON(fiber.Map{"message": "Permission updated successfully"})
}

// DeletePermission handles deleting a permission
func (h *RBACManagementHandler) DeletePermission(c *fiber.Ctx) error {
	// Check permission
	if !middleware.UserHasPermission(c, models.ResourceUsers, models.ActionDelete) {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	permID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid permission ID"})
	}

	// Check if permission is assigned to any roles
	var roleCount int
	err = h.db.QueryRowContext(c.Context(), `
		SELECT COUNT(*) FROM role_permissions WHERE permission_id = $1
	`, permID).Scan(&roleCount)

	if err != nil {
		h.logger.Error("Error checking permission usage", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	if roleCount > 0 {
		return c.Status(409).JSON(fiber.Map{"error": "Cannot delete permission that is assigned to roles"})
	}

	// Delete permission
	result, err := h.db.ExecContext(c.Context(), `
		DELETE FROM permissions WHERE id = $1
	`, permID)

	if err != nil {
		h.logger.Error("Error deleting permission", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.Error("Error getting rows affected", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	if rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Permission not found"})
	}

	// Log the action
	userID := GetCurrentUser(c, h.store)
	h.logger.Info("Permission deleted", "user_id", userID, "permission_id", permID)

	return c.JSON(fiber.Map{"message": "Permission deleted successfully"})
}

// ==================== USER ROLE MANAGEMENT ====================

// AssignUserRole handles assigning a role to a user
func (h *RBACManagementHandler) AssignUserRole(c *fiber.Ctx) error {
	// TODO: Re-enable permission check when authentication is properly set up
	// if !middleware.UserHasPermission(c, models.ResourceUsers, models.ActionCreate) {
	// 	return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	// }

	var req struct {
		UserID int `json:"user_id"`
		RoleID int `json:"role_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Check if assignment already exists
	var exists bool
	err := h.db.QueryRowContext(c.Context(), `
		SELECT EXISTS(SELECT 1 FROM user_roles WHERE user_id = $1 AND role_id = $2)
	`, req.UserID, req.RoleID).Scan(&exists)

	if err != nil {
		h.logger.Error("Error checking user role assignment", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	if exists {
		return c.Status(409).JSON(fiber.Map{"error": "User already has this role"})
	}

	// Insert the assignment
	_, err = h.db.ExecContext(c.Context(), `
		INSERT INTO user_roles (user_id, role_id, created_at)
		VALUES ($1, $2, NOW())
	`, req.UserID, req.RoleID)

	if err != nil {
		h.logger.Error("Error assigning role to user", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	// Log the action
	adminUserID := GetCurrentUser(c, h.store)
	h.logger.Info("Role assigned to user",
		"admin_user_id", adminUserID,
		"user_id", req.UserID,
		"role_id", req.RoleID)

	return c.JSON(fiber.Map{"message": "Role assigned to user successfully"})
}

// GetUserRoles handles getting all roles for a specific user
func (h *RBACManagementHandler) GetUserRoles(c *fiber.Ctx) error {
	// Check permission
	if !middleware.UserHasPermission(c, models.ResourceUsers, models.ActionRead) {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	userID, err := strconv.Atoi(c.Params("user_id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	query := `
		SELECT r.id, r.name, r.description, r.is_active, ur.created_at as assigned_at
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.name
	`

	rows, err := h.db.QueryContext(c.Context(), query, userID)
	if err != nil {
		h.logger.Error("Error getting user roles", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	defer rows.Close()

	var roles []fiber.Map
	for rows.Next() {
		var role struct {
			ID          int
			Name        string
			Description sql.NullString
			IsActive    bool
			AssignedAt  time.Time
		}

		err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.IsActive, &role.AssignedAt)
		if err != nil {
			h.logger.Error("Error scanning user role", "error", err)
			continue
		}

		roles = append(roles, fiber.Map{
			"id":          role.ID,
			"name":        role.Name,
			"description": role.Description.String,
			"is_active":   role.IsActive,
			"assigned_at": role.AssignedAt,
		})
	}

	return c.JSON(fiber.Map{"roles": roles})
}

// RemoveUserRole handles removing a role from a user
func (h *RBACManagementHandler) RemoveUserRole(c *fiber.Ctx) error {
	// Check permission
	if !middleware.UserHasPermission(c, models.ResourceUsers, models.ActionDelete) {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	userID, err := strconv.Atoi(c.Params("user_id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	roleID, err := strconv.Atoi(c.Params("role_id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid role ID"})
	}

	// Delete the assignment
	result, err := h.db.ExecContext(c.Context(), `
		DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2
	`, userID, roleID)

	if err != nil {
		h.logger.Error("Error removing user role", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.Error("Error getting rows affected", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	if rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "User role assignment not found"})
	}

	// Log the action
	adminUserID := GetCurrentUser(c, h.store)
	h.logger.Info("Role removed from user",
		"admin_user_id", adminUserID,
		"user_id", userID,
		"role_id", roleID)

	return c.JSON(fiber.Map{"message": "Role removed from user successfully"})
}

// ==================== ROLE PERMISSION MANAGEMENT ====================

// GetRolePermissions handles getting permissions for a specific role
func (h *RBACManagementHandler) GetRolePermissions(c *fiber.Ctx) error {
	// Check permission
	if !middleware.UserHasPermission(c, models.ResourceUsers, models.ActionRead) {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	roleID, err := strconv.Atoi(c.Params("role_id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid role ID"})
	}

	query := `
		SELECT p.id, p.name, p.description, p.resource, p.action, p.is_active
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.resource, p.action
	`

	rows, err := h.db.QueryContext(c.Context(), query, roleID)
	if err != nil {
		h.logger.Error("Error getting role permissions", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	defer rows.Close()

	var permissions []fiber.Map
	for rows.Next() {
		var perm struct {
			ID          int
			Name        string
			Description sql.NullString
			Resource    string
			Action      string
			IsActive    bool
		}

		err := rows.Scan(&perm.ID, &perm.Name, &perm.Description, &perm.Resource, &perm.Action, &perm.IsActive)
		if err != nil {
			h.logger.Error("Error scanning permission", "error", err)
			continue
		}

		permissions = append(permissions, fiber.Map{
			"id":          perm.ID,
			"name":        perm.Name,
			"description": perm.Description.String,
			"resource":    perm.Resource,
			"action":      perm.Action,
			"is_active":   perm.IsActive,
		})
	}

	return c.JSON(fiber.Map{"permissions": permissions})
}

// AssignRolePermission handles assigning a permission to a role
func (h *RBACManagementHandler) AssignRolePermission(c *fiber.Ctx) error {
	// Check permission
	if !middleware.UserHasPermission(c, models.ResourceUsers, models.ActionCreate) {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	var req struct {
		RoleID       int `json:"role_id"`
		PermissionID int `json:"permission_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Check if assignment already exists
	var exists bool
	err := h.db.QueryRowContext(c.Context(), `
		SELECT EXISTS(SELECT 1 FROM role_permissions WHERE role_id = $1 AND permission_id = $2)
	`, req.RoleID, req.PermissionID).Scan(&exists)

	if err != nil {
		h.logger.Error("Error checking role permission assignment", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	if exists {
		return c.Status(409).JSON(fiber.Map{"error": "Permission already assigned to role"})
	}

	// Insert the assignment
	_, err = h.db.ExecContext(c.Context(), `
		INSERT INTO role_permissions (role_id, permission_id, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
	`, req.RoleID, req.PermissionID)

	if err != nil {
		h.logger.Error("Error assigning permission to role", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	// Log the action
	userID := GetCurrentUser(c, h.store)
	h.logger.Info("Role permission assigned",
		"user_id", userID,
		"role_id", req.RoleID,
		"permission_id", req.PermissionID)

	return c.JSON(fiber.Map{"message": "Permission assigned to role successfully"})
}

// RemoveRolePermission handles removing a permission from a role
func (h *RBACManagementHandler) RemoveRolePermission(c *fiber.Ctx) error {
	// Check permission
	if !middleware.UserHasPermission(c, models.ResourceUsers, models.ActionDelete) {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	roleID, err := strconv.Atoi(c.Params("role_id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid role ID"})
	}

	permissionID, err := strconv.Atoi(c.Params("permission_id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid permission ID"})
	}

	// Delete the assignment
	result, err := h.db.ExecContext(c.Context(), `
		DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2
	`, roleID, permissionID)

	if err != nil {
		h.logger.Error("Error removing role permission", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		h.logger.Error("Error getting rows affected", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	if rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Role permission assignment not found"})
	}

	// Log the action
	userID := GetCurrentUser(c, h.store)
	h.logger.Info("Role permission removed",
		"user_id", userID,
		"role_id", roleID,
		"permission_id", permissionID)

	return c.JSON(fiber.Map{"message": "Permission removed from role successfully"})
}

// ==================== MIGRATION STATUS ====================

// GetMigrationStatus handles getting the status of user_right to RBAC migration
func (h *RBACManagementHandler) GetMigrationStatus(c *fiber.Ctx) error {
	// TODO: Re-enable permission check when authentication is properly set up
	// if !middleware.UserHasPermission(c, models.ResourceUsers, models.ActionRead) {
	// 	return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	// }

	// Get user_right statistics
	var userRightCount int
	err := h.db.QueryRowContext(c.Context(), "SELECT COUNT(*) FROM user_right").Scan(&userRightCount)
	if err != nil {
		h.logger.Error("Error getting user_right count", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	// Get RBAC statistics
	var roleCount, permissionCount, userRoleCount int
	err = h.db.QueryRowContext(c.Context(), "SELECT COUNT(*) FROM roles").Scan(&roleCount)
	if err != nil {
		h.logger.Error("Error getting role count", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	err = h.db.QueryRowContext(c.Context(), "SELECT COUNT(*) FROM permissions").Scan(&permissionCount)
	if err != nil {
		h.logger.Error("Error getting permission count", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	err = h.db.QueryRowContext(c.Context(), "SELECT COUNT(*) FROM user_roles").Scan(&userRoleCount)
	if err != nil {
		h.logger.Error("Error getting user_role count", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	// Get migration audit log
	var migrationLog string
	err = h.db.QueryRowContext(c.Context(), `
		SELECT details FROM audit_logs 
		WHERE action = 'migration' AND resource = 'user_right_to_rbac'
		ORDER BY created_at DESC LIMIT 1
	`).Scan(&migrationLog)

	migrationDetails := fiber.Map{}
	if err == nil {
		json.Unmarshal([]byte(migrationLog), &migrationDetails)
	}

	return c.JSON(fiber.Map{
		"user_right_count":  userRightCount,
		"role_count":        roleCount,
		"permission_count":  permissionCount,
		"user_role_count":   userRoleCount,
		"migration_details": migrationDetails,
	})
}

// MigrateUserRightsToRBAC handles migrating user rights to RBAC
func (h *RBACManagementHandler) MigrateUserRightsToRBAC(c *fiber.Ctx) error {
	// Check permission
	if !middleware.UserHasPermission(c, models.ResourceUsers, models.ActionCreate) {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	// This would call the migration script
	// For now, return a placeholder response
	return c.JSON(fiber.Map{
		"message": "Migration initiated",
		"status":  "pending",
		"note":    "Migration functionality needs to be implemented",
	})
}

// CreateDefaultAdminUser creates a default admin user for testing/development
func (h *RBACManagementHandler) CreateDefaultAdminUser(c *fiber.Ctx) error {
	// Check if admin user already exists
	var count int
	err := h.db.QueryRowContext(c.Context(), "SELECT COUNT(*) FROM users WHERE user_name = 'admin'").Scan(&count)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	if count > 0 {
		return c.JSON(fiber.Map{"message": "Admin user already exists"})
	}

	// Create admin user
	query := `
		INSERT INTO users (user_name, email, first_name, last_name, password_hash, password_salt, 
		                  is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING user_id
	`

	var userID int
	passwordHash := "hashed_admin_password" // In production, use proper hashing
	passwordSalt := "admin_salt"

	err = h.db.QueryRowContext(c.Context(), query,
		"admin", "admin@system.local", "System", "Administrator",
		passwordHash, passwordSalt, true, time.Now(), time.Now()).Scan(&userID)

	if err != nil {
		h.logger.Error("Error creating admin user", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create admin user"})
	}

	// Get super admin role
	var roleID int
	err = h.db.QueryRowContext(c.Context(), "SELECT id FROM roles WHERE name = 'super_admin'").Scan(&roleID)
	if err != nil {
		h.logger.Error("Error getting super admin role", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get super admin role"})
	}

	// Assign super admin role to admin user
	_, err = h.db.ExecContext(c.Context(),
		"INSERT INTO user_roles (user_id, role_id, created_at) VALUES ($1, $2, $3)",
		userID, roleID, time.Now())

	if err != nil {
		h.logger.Error("Error assigning role to admin user", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to assign role to admin user"})
	}

	return c.JSON(fiber.Map{
		"message":  "Default admin user created successfully",
		"username": "admin",
		"password": "admin123", // Default password
		"user_id":  userID,
	})
}

// ==================== STANDALONE HANDLER FUNCTIONS ====================

// HandlerGetRoles handles getting all roles
func HandlerGetRoles(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	handler := NewRBACManagementHandler(db, sl, nil, Config{})
	return handler.ListRoles(c)
}

// HandlerCreateRole handles creating a new role
func HandlerCreateRole(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	handler := NewRBACManagementHandler(db, sl, store, config)
	return handler.CreateRole(c)
}

// HandlerUpdateRole handles updating a role
func HandlerUpdateRole(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	handler := NewRBACManagementHandler(db, sl, store, config)
	return handler.UpdateRole(c)
}

// HandlerDeleteRole handles deleting a role
func HandlerDeleteRole(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	handler := NewRBACManagementHandler(db, sl, store, config)
	return handler.DeleteRole(c)
}

// HandlerGetPermissions handles getting all permissions
func HandlerGetPermissions(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	handler := NewRBACManagementHandler(db, sl, nil, Config{})
	return handler.ListPermissions(c)
}

// HandlerCreatePermission handles creating a new permission
func HandlerCreatePermission(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	handler := NewRBACManagementHandler(db, sl, store, config)
	return handler.CreatePermission(c)
}

// HandlerUpdatePermission handles updating a permission
func HandlerUpdatePermission(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	handler := NewRBACManagementHandler(db, sl, store, config)
	return handler.UpdatePermission(c)
}

// HandlerDeletePermission handles deleting a permission
func HandlerDeletePermission(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	handler := NewRBACManagementHandler(db, sl, store, config)
	return handler.DeletePermission(c)
}

// HandlerGetUserRoles handles getting user roles
func HandlerGetUserRoles(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	handler := NewRBACManagementHandler(db, sl, nil, Config{})
	return handler.GetUserRoles(c)
}

// HandlerAssignUserRole handles assigning a role to a user
func HandlerAssignUserRole(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	handler := NewRBACManagementHandler(db, sl, store, config)
	return handler.AssignUserRole(c)
}

// HandlerRemoveUserRole handles removing a role from a user
func HandlerRemoveUserRole(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	handler := NewRBACManagementHandler(db, sl, store, config)
	return handler.RemoveUserRole(c)
}

// HandlerGetRolePermissions handles getting role permissions
func HandlerGetRolePermissions(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	handler := NewRBACManagementHandler(db, sl, nil, Config{})
	return handler.GetRolePermissions(c)
}

// HandlerAssignRolePermission handles assigning a permission to a role
func HandlerAssignRolePermission(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	handler := NewRBACManagementHandler(db, sl, store, config)
	return handler.AssignRolePermission(c)
}

// HandlerRemoveRolePermission handles removing a permission from a role
func HandlerRemoveRolePermission(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	handler := NewRBACManagementHandler(db, sl, store, config)
	return handler.RemoveRolePermission(c)
}

// HandlerGetMigrationStatus handles getting migration status
func HandlerGetMigrationStatus(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	handler := NewRBACManagementHandler(db, sl, nil, Config{})
	return handler.GetMigrationStatus(c)
}

// HandlerGetUsers handles getting all users
func HandlerGetUsers(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	// Query to get users with their roles
	query := `
		SELECT DISTINCT u.user_id, u.user_name, u.email, u.first_name, u.last_name, 
		       u.is_active, u.is_locked, u.last_login_at, u.created_at,
		       COALESCE(array_agg(r.name) FILTER (WHERE r.name IS NOT NULL), ARRAY[]::text[]) as roles
		FROM users u
		LEFT JOIN user_roles ur ON u.user_id = ur.user_id
		LEFT JOIN roles r ON ur.role_id = r.id
		GROUP BY u.user_id, u.user_name, u.email, u.first_name, u.last_name, 
		         u.is_active, u.is_locked, u.last_login_at, u.created_at
		ORDER BY u.user_name
	`

	rows, err := db.QueryContext(c.Context(), query)
	if err != nil {
		sl.Error("Error querying users", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch users"})
	}
	defer rows.Close()

	var users []fiber.Map
	for rows.Next() {
		var user struct {
			ID          int
			UserName    string
			Email       sql.NullString
			FirstName   sql.NullString
			LastName    sql.NullString
			IsActive    bool
			IsLocked    bool
			LastLoginAt sql.NullTime
			CreatedAt   time.Time
			Roles       pq.StringArray
		}

		err := rows.Scan(
			&user.ID, &user.UserName, &user.Email, &user.FirstName, &user.LastName,
			&user.IsActive, &user.IsLocked, &user.LastLoginAt, &user.CreatedAt, &user.Roles,
		)
		if err != nil {
			sl.Error("Error scanning user row", "error", err)
			continue
		}

		users = append(users, fiber.Map{
			"user_id":       user.ID,
			"user_name":     user.UserName,
			"email":         user.Email.String,
			"first_name":    user.FirstName.String,
			"last_name":     user.LastName.String,
			"is_active":     user.IsActive,
			"is_locked":     user.IsLocked,
			"last_login_at": user.LastLoginAt.Time,
			"created_at":    user.CreatedAt,
			"roles":         user.Roles,
		})
	}

	return c.JSON(fiber.Map{"users": users})
}

// HandlerMigrateUserRightsToRBAC handles migrating user rights to RBAC
func HandlerMigrateUserRightsToRBAC(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	handler := NewRBACManagementHandler(db, sl, store, config)
	return handler.MigrateUserRightsToRBAC(c)
}

// GetRole handles fetching a single role by ID
func (h *RBACManagementHandler) GetRole(c *fiber.Ctx) error {
	if !middleware.UserHasPermission(c, models.ResourceUsers, models.ActionRead) {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}
	roleID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid role ID"})
	}
	var role struct {
		ID          int
		Name        string
		Description sql.NullString
		IsActive    bool
		CreatedAt   time.Time
		UpdatedAt   time.Time
	}
	err = h.db.QueryRowContext(c.Context(), `SELECT id, name, description, is_active, created_at, updated_at FROM roles WHERE id = $1`, roleID).Scan(
		&role.ID, &role.Name, &role.Description, &role.IsActive, &role.CreatedAt, &role.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return c.Status(404).JSON(fiber.Map{"error": "Role not found"})
	} else if err != nil {
		h.logger.Error("Error fetching role", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	return c.JSON(fiber.Map{
		"id":          role.ID,
		"name":        role.Name,
		"description": role.Description.String,
		"is_active":   role.IsActive,
		"created_at":  role.CreatedAt,
		"updated_at":  role.UpdatedAt,
	})
}

// GetPermission handles fetching a single permission by ID
func (h *RBACManagementHandler) GetPermission(c *fiber.Ctx) error {
	if !middleware.UserHasPermission(c, models.ResourceUsers, models.ActionRead) {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}
	permID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid permission ID"})
	}
	var perm struct {
		ID          int
		Name        string
		Description sql.NullString
		Resource    string
		Action      string
		IsActive    bool
		CreatedAt   time.Time
		UpdatedAt   time.Time
	}
	err = h.db.QueryRowContext(c.Context(), `SELECT id, name, description, resource, action, is_active, created_at, updated_at FROM permissions WHERE id = $1`, permID).Scan(
		&perm.ID, &perm.Name, &perm.Description, &perm.Resource, &perm.Action, &perm.IsActive, &perm.CreatedAt, &perm.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return c.Status(404).JSON(fiber.Map{"error": "Permission not found"})
	} else if err != nil {
		h.logger.Error("Error fetching permission", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	return c.JSON(fiber.Map{
		"id":          perm.ID,
		"name":        perm.Name,
		"description": perm.Description.String,
		"resource":    perm.Resource,
		"action":      perm.Action,
		"is_active":   perm.IsActive,
		"created_at":  perm.CreatedAt,
		"updated_at":  perm.UpdatedAt,
	})
}

// HandlerGetRole handles getting a single role by ID
func HandlerGetRole(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	handler := NewRBACManagementHandler(db, sl, nil, Config{})
	return handler.GetRole(c)
}

// HandlerGetPermission handles getting a single permission by ID
func HandlerGetPermission(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	handler := NewRBACManagementHandler(db, sl, nil, Config{})
	return handler.GetPermission(c)
}

// GetUserPermissions handles getting all permissions for a specific user
func (h *RBACManagementHandler) GetUserPermissions(c *fiber.Ctx) error {
	// TODO: Re-enable permission check when authentication is properly set up
	// if !middleware.UserHasPermission(c, models.ResourceUsers, models.ActionRead) {
	// 	return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	// }

	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	query := `
		SELECT DISTINCT p.id, p.name, p.description, p.resource, p.action, p.is_active
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1 AND p.is_active = true
		ORDER BY p.resource, p.action
	`

	rows, err := h.db.QueryContext(c.Context(), query, userID)
	if err != nil {
		h.logger.Error("Error getting user permissions", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	defer rows.Close()

	var permissions []fiber.Map
	for rows.Next() {
		var perm struct {
			ID          int
			Name        string
			Description sql.NullString
			Resource    string
			Action      string
			IsActive    bool
		}

		err := rows.Scan(&perm.ID, &perm.Name, &perm.Description, &perm.Resource, &perm.Action, &perm.IsActive)
		if err != nil {
			h.logger.Error("Error scanning permission", "error", err)
			continue
		}

		permissions = append(permissions, fiber.Map{
			"id":          perm.ID,
			"name":        perm.Name,
			"description": perm.Description.String,
			"resource":    perm.Resource,
			"action":      perm.Action,
			"is_active":   perm.IsActive,
		})
	}

	return c.JSON(fiber.Map{"permissions": permissions})
}

// HandlerGetUserPermissions handles getting all permissions for a specific user
func HandlerGetUserPermissions(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	handler := NewRBACManagementHandler(db, sl, nil, Config{})
	return handler.GetUserPermissions(c)
}
