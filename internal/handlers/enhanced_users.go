package handlers

import (
	"case/internal/middleware"
	"case/internal/models"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/lib/pq"
)

// EnhancedUserHandler handles user management with RBAC
type EnhancedUserHandler struct {
	db     *sql.DB
	logger *slog.Logger
	store  *session.Store
	config Config
}

// NewEnhancedUserHandler creates a new user handler
func NewEnhancedUserHandler(db *sql.DB, logger *slog.Logger, store *session.Store, config Config) *EnhancedUserHandler {
	return &EnhancedUserHandler{
		db:     db,
		logger: logger,
		store:  store,
		config: config,
	}
}

// ListUsers handles listing users with pagination and filtering
func (h *EnhancedUserHandler) ListUsers(c *fiber.Ctx) error {
	// Check permission
	if !middleware.UserHasPermission(c, models.ResourceUsers, models.ActionRead) {
		return c.Status(403).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	// Get query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search", "")
	departmentID := c.Query("department_id", "")
	roleID := c.Query("role_id", "")
	isActive := c.Query("is_active", "")

	// Build query
	query := `
		SELECT DISTINCT u.user_id, u.user_name, u.email, u.first_name, u.last_name, 
		       u.is_active, u.is_locked, u.last_login_at, u.created_at,
		       d.name as department_name,
		       COALESCE(array_agg(r.name) FILTER (WHERE r.id IS NOT NULL), '{}') AS roles
		FROM users u
		LEFT JOIN departments d ON u.department_id = d.id
		LEFT JOIN user_roles ur ON u.user_id = ur.user_id
		LEFT JOIN roles r ON ur.role_id = r.id
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 0

	// Add search filter
	if search != "" {
		argCount++
		query += fmt.Sprintf(" AND (u.user_name ILIKE $%d OR u.email ILIKE $%d OR u.first_name ILIKE $%d OR u.last_name ILIKE $%d)",
			argCount, argCount, argCount, argCount)
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm)
	}

	// Add department filter
	if departmentID != "" {
		argCount++
		query += fmt.Sprintf(" AND u.department_id = $%d", argCount)
		args = append(args, departmentID)
	}

	// Add role filter
	if roleID != "" {
		argCount++
		query += fmt.Sprintf(" AND ur.role_id = $%d", argCount)
		args = append(args, roleID)
	}

	// Add active filter
	if isActive != "" {
		argCount++
		query += fmt.Sprintf(" AND u.is_active = $%d", argCount)
		args = append(args, isActive == "true")
	}

	// Add grouping and ordering
	query += " GROUP BY u.user_id, d.name ORDER BY u.created_at DESC"

	// Add pagination
	offset := (page - 1) * limit
	argCount++
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	// Execute query
	rows, err := h.db.QueryContext(c.Context(), query, args...)
	if err != nil {
		h.logger.Error("Error querying users", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	defer rows.Close()

	users := make([]map[string]interface{}, 0)
	for rows.Next() {
		var user struct {
			ID             int
			Username       string
			Email          sql.NullString
			FirstName      sql.NullString
			LastName       sql.NullString
			IsActive       bool
			IsLocked       bool
			LastLoginAt    sql.NullTime
			CreatedAt      time.Time
			DepartmentName sql.NullString
		}
		var roles []string

		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName,
			&user.IsActive, &user.IsLocked, &user.LastLoginAt, &user.CreatedAt,
			&user.DepartmentName, pq.Array(&roles),
		)
		if err != nil {
			h.logger.Error("Error scanning user row", "error", err)
			continue
		}

		users = append(users, map[string]interface{}{
			"id":              user.ID,
			"username":        user.Username,
			"email":           user.Email.String,
			"first_name":      user.FirstName.String,
			"last_name":       user.LastName.String,
			"is_active":       user.IsActive,
			"is_locked":       user.IsLocked,
			"last_login_at":   user.LastLoginAt.Time,
			"created_at":      user.CreatedAt,
			"department_name": user.DepartmentName.String,
			"roles":           roles,
		})
	}

	// Get total count for pagination
	countQuery := `
		SELECT COUNT(DISTINCT u.user_id)
		FROM users u
		LEFT JOIN user_roles ur ON u.user_id = ur.user_id
		WHERE 1=1
	`

	countArgs := []interface{}{}
	argCount = 0

	if search != "" {
		argCount++
		countQuery += fmt.Sprintf(" AND (u.user_name ILIKE $%d OR u.email ILIKE $%d OR u.first_name ILIKE $%d OR u.last_name ILIKE $%d)",
			argCount, argCount, argCount, argCount)
		searchTerm := "%" + search + "%"
		countArgs = append(countArgs, searchTerm)
	}

	if departmentID != "" {
		argCount++
		countQuery += fmt.Sprintf(" AND u.department_id = $%d", argCount)
		countArgs = append(countArgs, departmentID)
	}

	if roleID != "" {
		argCount++
		countQuery += fmt.Sprintf(" AND ur.role_id = $%d", argCount)
		countArgs = append(countArgs, roleID)
	}

	if isActive != "" {
		argCount++
		countQuery += fmt.Sprintf(" AND u.is_active = $%d", argCount)
		countArgs = append(countArgs, isActive == "true")
	}

	var totalCount int
	err = h.db.QueryRowContext(c.Context(), countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		h.logger.Error("Error counting users", "error", err)
	}

	// Calculate pagination info
	totalPages := 0
	if limit > 0 {
		totalPages = (totalCount + limit - 1) / limit
	}

	return c.JSON(fiber.Map{
		"users": users,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       totalCount,
			"total_pages": totalPages,
		},
	})
}

// CreateUser handles user creation
func (h *EnhancedUserHandler) CreateUser(c *fiber.Ctx) error {
	// Check permission
	if !middleware.UserHasPermission(c, models.ResourceUsers, models.ActionCreate) {
		return c.Status(403).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	var userData struct {
		Username     string `json:"username"`
		Email        string `json:"email"`
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		Password     string `json:"password"`
		DepartmentID int    `json:"department_id"`
		RoleIDs      []int  `json:"role_ids"`
		IsActive     bool   `json:"is_active"`
	}

	if err := c.BodyParser(&userData); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate required fields
	if userData.Username == "" || userData.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Username and password are required"})
	}

	// Check if username already exists
	var exists bool
	err := h.db.QueryRowContext(c.Context(), "SELECT EXISTS(SELECT 1 FROM users WHERE user_name = $1)", userData.Username).Scan(&exists)
	if err != nil {
		h.logger.Error("Error checking username existence", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	if exists {
		return c.Status(400).JSON(fiber.Map{"error": "Username already exists"})
	}

	// Generate password salt and hash
	salt := generateSalt()
	passwordHash := hashPassword(userData.Password, salt)

	// Get current user ID for audit
	currentUserID, _ := middleware.GetCurrentUserID(c)

	// Create user
	user := &models.EnhancedUser{
		UserName:     sql.NullString{String: userData.Username, Valid: true},
		Email:        sql.NullString{String: userData.Email, Valid: userData.Email != ""},
		FirstName:    sql.NullString{String: userData.FirstName, Valid: userData.FirstName != ""},
		LastName:     sql.NullString{String: userData.LastName, Valid: userData.LastName != ""},
		PasswordHash: sql.NullString{String: passwordHash, Valid: true},
		PasswordSalt: sql.NullString{String: salt, Valid: true},
		DepartmentID: sql.NullInt64{Int64: int64(userData.DepartmentID), Valid: userData.DepartmentID > 0},
		IsActive:     userData.IsActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		CreatedBy:    sql.NullInt64{Int64: int64(currentUserID), Valid: currentUserID > 0},
	}

	// Insert user
	query := `
		INSERT INTO users (user_name, email, first_name, last_name, password_hash, password_salt, 
		                  department_id, is_active, created_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING user_id
	`

	err = h.db.QueryRowContext(c.Context(), query,
		user.UserName.String, user.Email.String, user.FirstName.String, user.LastName.String,
		user.PasswordHash.String, user.PasswordSalt.String, user.DepartmentID,
		user.IsActive, user.CreatedAt, user.CreatedBy).Scan(&user.UserID)

	if err != nil {
		h.logger.Error("Error creating user", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create user"})
	}

	// Assign roles
	if len(userData.RoleIDs) > 0 {
		for _, roleID := range userData.RoleIDs {
			_, err := h.db.ExecContext(c.Context(),
				"INSERT INTO user_roles (user_id, role_id, created_at) VALUES ($1, $2, $3)",
				user.UserID, roleID, time.Now())
			if err != nil {
				h.logger.Error("Error assigning role to user", "error", err, "user_id", user.UserID, "role_id", roleID)
			}
		}
	}

	// Log audit event
	auditLog := &models.AuditLog{
		UserID:     sql.NullInt64{Int64: int64(currentUserID), Valid: currentUserID > 0},
		Action:     "create",
		Resource:   models.ResourceUsers,
		ResourceID: sql.NullInt64{Int64: int64(user.UserID), Valid: true},
		Details:    fmt.Sprintf("Created user: %s", user.UserName.String),
		IPAddress:  c.IP(),
		UserAgent:  c.Get("User-Agent"),
		CreatedAt:  time.Now(),
	}
	models.LogAuditEvent(c.Context(), h.db, auditLog)

	return c.Status(201).JSON(fiber.Map{
		"message": "User created successfully",
		"user_id": user.UserID,
	})
}

// UpdateUser handles user updates
func (h *EnhancedUserHandler) UpdateUser(c *fiber.Ctx) error {
	// Check permission
	if !middleware.UserHasPermission(c, models.ResourceUsers, models.ActionUpdate) {
		return c.Status(403).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var userData struct {
		Email        string `json:"email"`
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		DepartmentID int    `json:"department_id"`
		IsActive     bool   `json:"is_active"`
		RoleIDs      []int  `json:"role_ids"`
	}

	if err := c.BodyParser(&userData); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Get current user ID for audit
	currentUserID, _ := middleware.GetCurrentUserID(c)

	// Update user
	query := `
		UPDATE users 
		SET email = $1, first_name = $2, last_name = $3, department_id = $4, 
		    is_active = $5, updated_at = $6, updated_by = $7
		WHERE user_id = $8
	`

	result, err := h.db.ExecContext(c.Context(), query,
		userData.Email, userData.FirstName, userData.LastName,
		sql.NullInt64{Int64: int64(userData.DepartmentID), Valid: userData.DepartmentID > 0},
		userData.IsActive, time.Now(),
		sql.NullInt64{Int64: int64(currentUserID), Valid: currentUserID > 0},
		userID)

	if err != nil {
		h.logger.Error("Error updating user", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update user"})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	// Update roles if provided
	if len(userData.RoleIDs) > 0 {
		// Remove existing roles
		_, err = h.db.ExecContext(c.Context(), "DELETE FROM user_roles WHERE user_id = $1", userID)
		if err != nil {
			h.logger.Error("Error removing user roles", "error", err)
		}

		// Add new roles
		for _, roleID := range userData.RoleIDs {
			_, err := h.db.ExecContext(c.Context(),
				"INSERT INTO user_roles (user_id, role_id, created_at) VALUES ($1, $2, $3)",
				userID, roleID, time.Now())
			if err != nil {
				h.logger.Error("Error assigning role to user", "error", err, "user_id", userID, "role_id", roleID)
			}
		}
	}

	// Log audit event
	auditLog := &models.AuditLog{
		UserID:     sql.NullInt64{Int64: int64(currentUserID), Valid: currentUserID > 0},
		Action:     "update",
		Resource:   models.ResourceUsers,
		ResourceID: sql.NullInt64{Int64: int64(userID), Valid: true},
		Details:    "Updated user profile",
		IPAddress:  c.IP(),
		UserAgent:  c.Get("User-Agent"),
		CreatedAt:  time.Now(),
	}
	models.LogAuditEvent(c.Context(), h.db, auditLog)

	return c.JSON(fiber.Map{"message": "User updated successfully"})
}

// DeleteUser handles user deletion
func (h *EnhancedUserHandler) DeleteUser(c *fiber.Ctx) error {
	// Check permission
	if !middleware.UserHasPermission(c, models.ResourceUsers, models.ActionDelete) {
		return c.Status(403).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	// Get current user ID for audit
	currentUserID, _ := middleware.GetCurrentUserID(c)

	// Check if user is trying to delete themselves
	if userID == currentUserID {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot delete your own account"})
	}

	// Delete user (cascade will handle related records)
	result, err := h.db.ExecContext(c.Context(), "DELETE FROM users WHERE user_id = $1", userID)
	if err != nil {
		h.logger.Error("Error deleting user", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete user"})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	// Log audit event
	auditLog := &models.AuditLog{
		UserID:     sql.NullInt64{Int64: int64(currentUserID), Valid: currentUserID > 0},
		Action:     "delete",
		Resource:   models.ResourceUsers,
		ResourceID: sql.NullInt64{Int64: int64(userID), Valid: true},
		Details:    "Deleted user account",
		IPAddress:  c.IP(),
		UserAgent:  c.Get("User-Agent"),
		CreatedAt:  time.Now(),
	}
	models.LogAuditEvent(c.Context(), h.db, auditLog)

	return c.JSON(fiber.Map{"message": "User deleted successfully"})
}

// GetUserDetails returns detailed user information
func (h *EnhancedUserHandler) GetUserDetails(c *fiber.Ctx) error {
	// Check permission
	if !middleware.UserHasPermission(c, models.ResourceUsers, models.ActionRead) {
		return c.Status(403).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	// Get user details with roles and department
	query := `
		SELECT u.user_id, u.user_name, u.email, u.first_name, u.last_name, 
		       u.is_active, u.is_locked, u.last_login_at, u.created_at, u.updated_at,
		       d.id as department_id, d.name as department_name,
		       COALESCE(array_agg(r.id) FILTER (WHERE r.id IS NOT NULL), '{}') AS role_ids,
		       COALESCE(array_agg(r.name) FILTER (WHERE r.id IS NOT NULL), '{}') AS role_names
		FROM users u
		LEFT JOIN departments d ON u.department_id = d.id
		LEFT JOIN user_roles ur ON u.user_id = ur.user_id
		LEFT JOIN roles r ON ur.role_id = r.id
		WHERE u.user_id = $1
		GROUP BY u.user_id, d.id, d.name
	`

	var user struct {
		ID             int
		Username       string
		Email          sql.NullString
		FirstName      sql.NullString
		LastName       sql.NullString
		IsActive       bool
		IsLocked       bool
		LastLoginAt    sql.NullTime
		CreatedAt      time.Time
		UpdatedAt      time.Time
		DepartmentID   sql.NullInt64
		DepartmentName sql.NullString
	}
	var roleIDs []int64
	var roleNames []string

	err = h.db.QueryRowContext(c.Context(), query, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName,
		&user.IsActive, &user.IsLocked, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
		&user.DepartmentID, &user.DepartmentName, pq.Array(&roleIDs), pq.Array(&roleNames),
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "User not found"})
		}
		h.logger.Error("Error getting user details", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	roleIDInts := make([]int, len(roleIDs))
	for i, id := range roleIDs {
		roleIDInts[i] = int(id)
	}

	response := map[string]interface{}{
		"id":              user.ID,
		"username":        user.Username,
		"email":           user.Email.String,
		"first_name":      user.FirstName.String,
		"last_name":       user.LastName.String,
		"is_active":       user.IsActive,
		"is_locked":       user.IsLocked,
		"last_login_at":   user.LastLoginAt.Time,
		"created_at":      user.CreatedAt,
		"updated_at":      user.UpdatedAt,
		"department_id":   user.DepartmentID.Int64,
		"department_name": user.DepartmentName.String,
		"role_ids":        roleIDInts,
		"roles":           roleNames,
	}

	return c.JSON(response)
}

// Helper functions
func generateSalt() string {
	salt := make([]byte, 16)
	rand.Read(salt)
	return hex.EncodeToString(salt)
}

func hashPassword(password, salt string) string {
	hash := sha256.New()
	hash.Write([]byte(password + salt))
	return hex.EncodeToString(hash.Sum(nil))
}
