package handlers

import (
	"case/internal/models"
	"case/internal/security"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type FormData struct {
	FidID   []string `form:"input[fid_id][]"`
	MetaID  []string `form:"input[meta_id][]"`
	Scope   []string `form:"input[scope][]"`
	View    []string `form:"input[view][]"`
	Add     []string `form:"input[add][]"`
	Edit    []string `form:"input[edit][]"`
	Execute []string `form:"input[execute][]"`
}

func HandlerUserForm(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userName := GetUser(c, sl, store)
	role := security.GetRoles(userID, "admin")

	id, err := strconv.Atoi(c.Params("i"))
	if err != nil {
		log.Println(err.Error())
	}

	// Create a simplified user form struct
	type UserForm struct {
		UserID            int
		UserName          sql.NullString
		UserPass          sql.NullString
		UserEmployee      sql.NullInt64
		IsActive          sql.NullBool
		IsLocked          sql.NullBool
		PasswordExpiresAt sql.NullTime
	}

	var uzer UserForm
	uzer.UserEmployee.Valid = true
	uzer.UserEmployee.Int64 = 0
	uzer.IsActive.Valid = true
	uzer.IsActive.Bool = true
	uzer.IsLocked.Valid = true
	uzer.IsLocked.Bool = false

	// Load existing user if editing
	if id > 0 {
		query := `SELECT user_id, user_name, user_pass, user_employee, 
		          COALESCE(is_active, true) as is_active,
		          COALESCE(is_locked, false) as is_locked,
		          password_expires_at
		          FROM public.users WHERE user_id = $1`
		err := db.QueryRowContext(c.Context(), query, id).Scan(
			&uzer.UserID, &uzer.UserName, &uzer.UserPass, &uzer.UserEmployee,
			&uzer.IsActive, &uzer.IsLocked, &uzer.PasswordExpiresAt,
		)
		if err != nil {
			sl.Error("Error loading user", "error", err)
			// Continue with empty user for new form
		}
	}

	// Get functions for permissions (simplified)
	// Define funclist struct
	type funclist struct {
		FID      sql.NullInt64 `json:"fid"`
		MetaID   sql.NullInt64 `json:"meta_id"`
		MetaName string        `json:"meta_name"`
		FScope   sql.NullInt64 `json:"f_scope"`
		FView    sql.NullInt64 `json:"f_view"`
		FCreate  sql.NullInt64 `json:"f_create"`
		FEdit    sql.NullInt64 `json:"f_edit"`
		FRemove  sql.NullInt64 `json:"f_remove"`
	}

	var functions []funclist
	functionsQuery := `SELECT fid_id, meta_id, meta_name, f_scope, f_view, f_create, f_edit, f_remove 
	                   FROM functions ORDER BY meta_name`
	rows, err := db.QueryContext(c.Context(), functionsQuery)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var f funclist
			err := rows.Scan(
				&f.FID, &f.MetaID, &f.MetaName, &f.FScope,
				&f.FView, &f.FCreate, &f.FEdit, &f.FRemove,
			)
			if err != nil {
				log.Println("Row Scan Error: ", err.Error())
				continue
			}
			functions = append(functions, f)
		}
	}

	// Add roles data
	data := NewTemplateData(c, store)
	data.Roles = []Role{}
	rolesQuery := `SELECT id, name, COALESCE(description, '') as description FROM roles ORDER BY name`
	rolesRows, err := db.QueryContext(c.Context(), rolesQuery)
	if err == nil {
		defer rolesRows.Close()
		for rolesRows.Next() {
			var role Role
			if err := rolesRows.Scan(&role.ID, &role.Name, &role.Description); err == nil {
				// Check if this role is assigned to the current user
				if id > 0 {
					userRoleQuery := `SELECT COUNT(*) FROM user_roles WHERE user_id = $1 AND role_id = $2`
					var count int
					if err := db.QueryRowContext(c.Context(), userRoleQuery, id, role.ID).Scan(&count); err == nil {
						role.Selected = count > 0
					}
				}
				data.Roles = append(data.Roles, role)
			}
		}
	}

	// Simplified permissions data
	data.Permissions = []Permission{
		{
			Resource: "users",
			Actions: []Action{
				{Action: "create", Granted: false},
				{Action: "read", Granted: false},
				{Action: "update", Granted: false},
				{Action: "delete", Granted: false},
			},
		},
		{
			Resource: "vhf_patients",
			Actions: []Action{
				{Action: "create", Granted: false},
				{Action: "read", Granted: false},
				{Action: "update", Granted: false},
				{Action: "delete", Granted: false},
			},
		},
		{
			Resource: "outbreaks",
			Actions: []Action{
				{Action: "create", Granted: false},
				{Action: "read", Granted: false},
				{Action: "update", Granted: false},
				{Action: "delete", Granted: false},
			},
		},
		{
			Resource: "reports",
			Actions: []Action{
				{Action: "create", Granted: false},
				{Action: "read", Granted: false},
				{Action: "update", Granted: false},
				{Action: "delete", Granted: false},
			},
		},
	}

	data.User = userName
	data.Role = role
	data.Form = uzer
	data.FormChild1 = functions
	data.From = c.Query("from", "") // Get the 'from' parameter from the URL

	return GenerateHTML(c, db, data, "form_user")
}

func HandlerUserSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	// Debug: Log all form values
	sl.Info("Form submission received", "content_type", c.Get("Content-Type"))

	id, er := strconv.Atoi(c.FormValue("id"))
	if er != nil {
		id = 0
	}

	// Debug: Log the user ID
	sl.Info("Processing user submission", "user_id", id)

	// Get employee details if employee is selected
	employeeID, err := strconv.Atoi(c.FormValue("user_employee"))
	if err != nil {
		employeeID = 0
	}

	var firstName, lastName, email sql.NullString
	var departmentID sql.NullInt64

	if employeeID > 0 {
		// Get employee details
		empQuery := `SELECT employee_fname, employee_lname, employee_email, facility 
		             FROM employee WHERE employee_id = $1`
		err := db.QueryRowContext(c.Context(), empQuery, employeeID).Scan(
			&firstName, &lastName, &email, &departmentID)
		if err != nil {
			sl.Error("Error getting employee details", "error", err, "employee_id", employeeID)
		}
	}

	user := models.User{
		UserID:       id,
		UserName:     ParseNullString(c.FormValue("user_name")),
		UserEmployee: ParseNullInt(c.FormValue("user_employee")),
	}

	// Set password for new users
	if user.UserID == 0 {
		password := c.FormValue("user_pass")
		if password == "" {
			password = "123456" // Default password
		}
		user.UserPass = sql.NullString{String: models.Encrypt(password), Valid: true}
		err := user.Insert(c.Context(), db)
		if err != nil {
			log.Println(err.Error())
		}
		sl.Info("Created new user", "user_id", user.UserID)
	} else {
		user.SetAsExists()
		err := user.Update_NoPass(c.Context(), db)
		if err != nil {
			log.Println(err.Error())
		}
		sl.Info("Updated existing user", "user_id", user.UserID)
	}

	// Update enhanced user data if employee is selected
	if employeeID > 0 && user.UserID > 0 {
		// Check if enhanced user record exists
		var exists bool
		err := db.QueryRowContext(c.Context(),
			"SELECT EXISTS(SELECT 1 FROM enhanced_users WHERE user_id = $1)", user.UserID).Scan(&exists)
		if err != nil {
			sl.Error("Error checking enhanced user existence", "error", err)
		}

		if exists {
			// Update existing enhanced user
			updateQuery := `UPDATE enhanced_users SET 
				first_name = $1, last_name = $2, email = $3, department_id = $4, updated_at = $5
				WHERE user_id = $6`
			_, err := db.ExecContext(c.Context(), updateQuery,
				firstName, lastName, email, departmentID, time.Now(), user.UserID)
			if err != nil {
				sl.Error("Error updating enhanced user", "error", err)
			}
		} else {
			// Create new enhanced user record
			insertQuery := `INSERT INTO enhanced_users 
				(user_id, first_name, last_name, email, department_id, is_active, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, true, $6, $7)`
			_, err := db.ExecContext(c.Context(), insertQuery,
				user.UserID, firstName, lastName, email, departmentID, time.Now(), time.Now())
			if err != nil {
				sl.Error("Error creating enhanced user", "error", err)
			}
		}
	}

	// Handle RBAC roles and permissions
	// First, remove existing user roles
	if user.UserID > 0 {
		_, err := db.ExecContext(c.Context(), "DELETE FROM user_roles WHERE user_id = $1", user.UserID)
		if err != nil {
			sl.Error("Error removing existing user roles", "error", err)
		}
	}

	// Add selected roles
	// For multiple select, we need to get all form values with the same name
	form, err := c.MultipartForm()
	if err != nil {
		sl.Error("Error parsing form", "error", err)
		// Try alternative approach for non-multipart forms
		sl.Info("Trying alternative form parsing approach")

		// Try to get roles from regular form values
		rolesValue := c.FormValue("roles")
		sl.Info("Roles from FormValue", "roles", rolesValue)

		if rolesValue != "" {
			// Split by comma if multiple roles are sent as a single value
			roleIDs := strings.Split(rolesValue, ",")
			for _, roleIDStr := range roleIDs {
				roleID, err := strconv.Atoi(strings.TrimSpace(roleIDStr))
				if err != nil {
					sl.Error("Invalid role ID", "role_id", roleIDStr, "error", err)
					continue
				}

				// Insert user role
				_, err = db.ExecContext(c.Context(),
					"INSERT INTO user_roles (user_id, role_id, created_at) VALUES ($1, $2, $3)",
					user.UserID, roleID, time.Now())
				if err != nil {
					sl.Error("Error assigning role to user", "user_id", user.UserID, "role_id", roleID, "error", err)
				} else {
					sl.Info("Successfully assigned role to user", "user_id", user.UserID, "role_id", roleID)
				}
			}
		} else {
			sl.Info("No roles selected for user", "user_id", user.UserID)
		}
	} else {
		selectedRoles := form.Value["roles"]
		sl.Info("Roles from MultipartForm", "roles", selectedRoles)
		if len(selectedRoles) > 0 {
			for _, roleIDStr := range selectedRoles {
				roleID, err := strconv.Atoi(strings.TrimSpace(roleIDStr))
				if err != nil {
					sl.Error("Invalid role ID", "role_id", roleIDStr, "error", err)
					continue
				}

				// Insert user role
				_, err = db.ExecContext(c.Context(),
					"INSERT INTO user_roles (user_id, role_id, created_at) VALUES ($1, $2, $3)",
					user.UserID, roleID, time.Now())
				if err != nil {
					sl.Error("Error assigning role to user", "user_id", user.UserID, "role_id", roleID, "error", err)
				} else {
					sl.Info("Successfully assigned role to user", "user_id", user.UserID, "role_id", roleID)
				}
			}
		} else {
			sl.Info("No roles selected for user", "user_id", user.UserID)
		}
	}

	// Handle individual permissions (if any are set)
	// This is for granular permissions beyond role-based permissions
	permissionResources := []string{"users", "vhf_patients", "outbreaks", "reports"}
	for _, resource := range permissionResources {
		for _, action := range []string{"create", "read", "update", "delete"} {
			permissionKey := fmt.Sprintf("permissions[%s][%s]", resource, action)
			if c.FormValue(permissionKey) == "1" {
				// This user has this specific permission
				// You might want to store this in a user_permissions table
				// For now, we'll just log it
				sl.Info("User has specific permission", "user_id", user.UserID, "resource", resource, "action", action)
			}
		}
	}

	urlx := "/users/new/" + strconv.Itoa(user.UserID)
	return c.Redirect(urlx)
}

func HandlerUserList(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	fmt.Println("starting user list")

	userID, userName := GetUser(c, sl, store)
	role := security.GetRoles(userID, "admin")

	data := NewTemplateDataWithDB(c, store, db)
	data.User = userName
	data.Role = role

	// Initialize stats with default values
	data.Stats = &Stats{
		TotalUsers:      0,
		ActiveUsers:     0,
		LockedUsers:     0,
		TotalRoles:      0,
		TotalOutbreaks:  0,
		TotalCases:      0,
		TotalFacilities: 0,
	}

	// Initialize empty arrays for departments and roles
	data.Departments = []Department{}
	data.Roles = []Role{}

	// Get user statistics - using actual table structure
	statsQuery := `SELECT COUNT(*) as total_users FROM public.users`

	var totalUsers int
	err := db.QueryRowContext(c.Context(), statsQuery).Scan(&totalUsers)
	if err != nil {
		sl.Error("Failed to get user statistics", "error", err)
	} else {
		data.Stats.TotalUsers = totalUsers
		// Since we don't have is_active/is_locked columns, assume all users are active
		data.Stats.ActiveUsers = totalUsers
		data.Stats.LockedUsers = 0
	}

	// Get total roles count - check if roles table exists
	rolesQuery := `SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'roles'`
	var rolesTableExists int
	err = db.QueryRowContext(c.Context(), rolesQuery).Scan(&rolesTableExists)
	if err != nil || rolesTableExists == 0 {
		sl.Error("Roles table does not exist", "error", err)
	} else {
		// If roles table exists, count roles and get role data
		rolesCountQuery := `SELECT COUNT(*) FROM roles`
		var totalRoles int
		err = db.QueryRowContext(c.Context(), rolesCountQuery).Scan(&totalRoles)
		if err != nil {
			sl.Error("Failed to get roles count", "error", err)
		} else {
			data.Stats.TotalRoles = totalRoles
		}

		// Get roles data for dropdown
		rolesDataQuery := `SELECT id, name FROM roles ORDER BY name`
		roleRows, err := db.QueryContext(c.Context(), rolesDataQuery)
		if err != nil {
			sl.Error("Failed to get roles data", "error", err)
		} else {
			defer roleRows.Close()
			for roleRows.Next() {
				var role Role
				if err := roleRows.Scan(&role.ID, &role.Name); err != nil {
					sl.Error("Failed to scan role data", "error", err)
					continue
				}
				data.Roles = append(data.Roles, role)
			}
		}
	}

	// Get departments data - check if departments table exists
	deptQuery := `SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'departments'`
	var deptTableExists int
	err = db.QueryRowContext(c.Context(), deptQuery).Scan(&deptTableExists)
	if err == nil && deptTableExists > 0 {
		// If departments table exists, get department data
		deptDataQuery := `SELECT id, name FROM departments ORDER BY name`
		deptRows, err := db.QueryContext(c.Context(), deptDataQuery)
		if err != nil {
			sl.Error("Failed to get departments data", "error", err)
		} else {
			defer deptRows.Close()
			for deptRows.Next() {
				var dept Department
				if err := deptRows.Scan(&dept.ID, &dept.Name); err != nil {
					sl.Error("Failed to scan department data", "error", err)
					continue
				}
				data.Departments = append(data.Departments, dept)
			}
		}
	}

	// Custom user structure for the template
	type UserWithDetails struct {
		UserID         int
		UserName       sql.NullString
		UserPass       sql.NullString
		UserEmployee   sql.NullInt64
		FirstName      sql.NullString
		LastName       sql.NullString
		Email          sql.NullString
		IsActive       sql.NullBool
		IsLocked       sql.NullBool
		LastLoginAt    sql.NullTime
		CreatedAt      sql.NullTime
		DepartmentName sql.NullString
		Roles          []Role
	}

	// Query to get user information with all available fields
	mysql := `
		SELECT user_id, user_name, user_pass, user_employee,
		       COALESCE(first_name, '') as first_name,
		       COALESCE(last_name, '') as last_name,
		       COALESCE(email, '') as email,
		       COALESCE(is_active, true) as is_active,
		       COALESCE(is_locked, false) as is_locked,
		       last_login_at,
		       COALESCE(created_at, CURRENT_TIMESTAMP) as created_at
		FROM public.users
		ORDER BY user_name
	`

	// Execute query
	rows, err := db.QueryContext(c.Context(), mysql)
	if err != nil {
		sl.Error("Database error in user list", "error", err.Error())
		// Return empty list instead of failing
		data.Form = []UserWithDetails{}
		return GenerateHTML(c, db, data, "list_users")
	}
	defer rows.Close()

	// Slice to hold users
	var users []UserWithDetails

	// Iterate through rows
	for rows.Next() {
		var u UserWithDetails
		if err := rows.Scan(&u.UserID, &u.UserName, &u.UserPass, &u.UserEmployee,
			&u.FirstName, &u.LastName, &u.Email, &u.IsActive, &u.IsLocked,
			&u.LastLoginAt, &u.CreatedAt); err != nil {
			sl.Error("Row scan error in user list", "error", err.Error())
			continue
		}

		// Initialize empty arrays and default values for missing fields
		u.Roles = []Role{}
		u.DepartmentName = sql.NullString{String: "", Valid: false}

		users = append(users, u)
	}

	// Check for iteration errors
	if err = rows.Err(); err != nil {
		sl.Error("Rows iteration error in user list", "error", err.Error())
	}

	data.Form = users
	return GenerateHTML(c, db, data, "list_users")
}
