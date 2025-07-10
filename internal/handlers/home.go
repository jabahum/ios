package handlers

import (
	"case/internal/models"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	_ "github.com/lib/pq"
)

// HandlerHome handles the home page
func HandlerHome(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	// Check if an outbreak is selected
	selectedOutbreakID := GetSelectedOutbreak(c, store)
	if selectedOutbreakID == 0 {
		// If no outbreak is selected, redirect to outbreak selection
		return c.Redirect("/outbreaks")
	}

	// Get the selected outbreak
	outbreak, err := models.OutbreakByID(c.Context(), db, selectedOutbreakID)
	if err != nil {
		sl.Error("Failed to get selected outbreak: " + err.Error())
		return c.Redirect("/outbreaks")
	}

	// Get the current user
	userID, username := GetUser(c, sl, store)

	data := NewTemplateData(c, store)
	data.User = username
	data.Form = outbreak

	// Get user permissions for access control
	permissions, err := getUserPermissions(c, db, userID)
	if err != nil {
		sl.Error("Failed to get user permissions: " + err.Error())
		// Continue without permissions - user will see limited access
		permissions = make(map[string][]string)
	}

	data.UserPermissions = permissions

	// Instead of redirecting, render the home page with the selected outbreak
	return GenerateHTML(c, db, data, "home")
}

// getUserPermissions gets the permissions for a specific user
func getUserPermissions(c *fiber.Ctx, db *sql.DB, userID int) (map[string][]string, error) {
	query := `
		SELECT DISTINCT p.resource, p.action
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1 AND p.is_active = true
		ORDER BY p.resource, p.action
	`

	rows, err := db.QueryContext(c.Context(), query, userID)
	if err != nil {
		return nil, err
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

	return permissions, nil
}

// getUserPrimaryRole gets the primary role for a user to determine redirect destination
func getUserPrimaryRole(c *fiber.Ctx, db *sql.DB, userID int) (string, error) {
	query := `
		SELECT r.name
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND r.is_active = true
		ORDER BY 
			CASE r.name 
				WHEN 'super_admin' THEN 1
				WHEN 'admin' THEN 2
				WHEN 'vhf_lab_technician' THEN 3
				WHEN 'vhf_data_entry' THEN 4
				WHEN 'mpox_case_manager' THEN 5
				WHEN 'mpox_data_entry' THEN 6
				WHEN 'case_manager' THEN 7
				WHEN 'outbreak_manager' THEN 8
				WHEN 'outbreak_viewer' THEN 9
				ELSE 10
			END
		LIMIT 1
	`

	var roleName string
	err := db.QueryRowContext(c.Context(), query, userID).Scan(&roleName)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("DEBUG: No role found for user ID %d\n", userID)
			return "", nil // No role assigned
		}
		fmt.Printf("DEBUG: Error getting role for user ID %d: %v\n", userID, err)
		return "", err
	}

	fmt.Printf("DEBUG: Found role '%s' for user ID %d\n", roleName, userID)
	return roleName, nil
}

// getDefaultOutbreakID gets the default outbreak ID for case managers
func getDefaultOutbreakID(c *fiber.Ctx, db *sql.DB, userID int) (int, error) {
	// For now, return a default outbreak ID (you can customize this logic)
	// You might want to get this from user preferences or a specific assignment
	query := `SELECT id FROM outbreaks WHERE outbreak_type = 'general' ORDER BY id LIMIT 1`

	var outbreakID int
	err := db.QueryRowContext(c.Context(), query).Scan(&outbreakID)
	if err != nil {
		if err == sql.ErrNoRows {
			// If no outbreaks exist, return 0
			return 0, nil
		}
		return 0, err
	}

	return outbreakID, nil
}

func HandlerLoginForm(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {

	sess, err := store.Get(c)

	if err == nil {
		userID := sess.Get("user")
		if userID != nil {
			sl.Info("Session error, No ID set")
			return c.Redirect("/", 302)
		}
	}

	// load page
	data := map[string]string{"Title": "Login Page"}
	return GenerateHTML(c, db, data, "login")
}

func HandlerLoginSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {

	sess, err := store.Get(c)
	if err == nil {
		userID := sess.Get("user")
		if userID != nil {
			fmt.Println("Already logged in")
			return c.Redirect("/outbreaks", 302)
		}
	}

	// Extract form values
	username := c.FormValue("username")
	password := c.FormValue("password")

	if username == "" || password == "" {
		fmt.Println("No username and or password provided")
		c.Status(fiber.StatusBadRequest)      // Set HTTP 400 status
		return c.Redirect("/login?error=400") // Redirect to login page
	}

	id, er := models.Authenticate(c.Context(), db, username, password)
	if er != nil {
		fmt.Println("Failed Authentication: ", er.Error())
		return c.Redirect("/login?error=afail")
	}

	if id > 0 {
		// Get session
		sess, err := store.Get(c)
		if err != nil {
			sl.Info("Session error")
			return c.Redirect("/login?serror")
		}

		// Set session variables
		sess.Set("user", id)    // Keep for backward compatibility
		sess.Set("user_id", id) // Set for RBAC compatibility
		sess.Set("username", username)
		sess.Set("isAuthenticated", true)
		sess.Set("authenticated", true) // Set for RBAC compatibility

		// Save session
		if err := sess.Save(); err != nil {
			sl.Info("Failed to save session")
			return c.Redirect("/login?sfail")
		}

		// Get user's primary role to determine redirect destination
		primaryRole, err := getUserPrimaryRole(c, db, id)
		if err != nil {
			sl.Error("Failed to get user role for redirect", "error", err, "user_id", id)
			// Default to outbreaks page if role lookup fails
			return c.Redirect("/outbreaks")
		}

		fmt.Printf("DEBUG: User ID %d has primary role: '%s'\n", id, primaryRole)

		// Redirect based on user's primary role
		switch primaryRole {
		case "vhf_lab_technician", "vhf_data_entry":
			// VHF users go directly to VHF list
			fmt.Printf("DEBUG: Redirecting VHF user (ID: %d, role: %s) to /vhf-list\n", id, primaryRole)
			sl.Info("Redirecting VHF user to VHF list", "user_id", id, "role", primaryRole)
			return c.Redirect("/vhf-list")
		case "mpox_case_manager", "mpox_data_entry":
			// MPOX users go to outbreaks page (they can select MPOX outbreaks)
			fmt.Printf("DEBUG: Redirecting MPOX user (ID: %d, role: %s) to /outbreaks\n", id, primaryRole)
			sl.Info("Redirecting MPOX user to outbreaks", "user_id", id, "role", primaryRole)
			return c.Redirect("/outbreaks")
		case "case_manager":
			// Case managers go to /cases with default outbreak
			defaultOutbreakID, err := getDefaultOutbreakID(c, db, id)
			if err != nil {
				sl.Error("Failed to get default outbreak ID for case manager", "error", err, "user_id", id)
				return c.Redirect("/outbreaks") // Fallback to outbreaks if default ID fails
			}
			if defaultOutbreakID > 0 {
				fmt.Printf("DEBUG: Redirecting case manager (ID: %d, role: %s) to /cases/%d\n", id, primaryRole, defaultOutbreakID)
				sl.Info("Redirecting case manager to cases with default outbreak", "user_id", id, "role", primaryRole, "outbreak_id", defaultOutbreakID)
				return c.Redirect(fmt.Sprintf("/cases/%d", defaultOutbreakID))
			} else {
				fmt.Printf("DEBUG: Redirecting case manager (ID: %d, role: %s) to /outbreaks (no default outbreak)\n", id, primaryRole)
				sl.Info("Redirecting case manager to outbreaks (no default outbreak)", "user_id", id, "role", primaryRole)
				return c.Redirect("/outbreaks")
			}
		case "outbreak_viewer", "outbreak_manager":
			// Outbreak users go to outbreaks page
			fmt.Printf("DEBUG: Redirecting outbreak user (ID: %d, role: %s) to /outbreaks\n", id, primaryRole)
			sl.Info("Redirecting outbreak user to outbreaks", "user_id", id, "role", primaryRole)
			return c.Redirect("/outbreaks")
		case "super_admin", "admin":
			// Admin users go to outbreaks page (they can see all outbreaks)
			fmt.Printf("DEBUG: Redirecting admin user (ID: %d, role: %s) to /outbreaks\n", id, primaryRole)
			sl.Info("Redirecting admin user to outbreaks", "user_id", id, "role", primaryRole)
			return c.Redirect("/outbreaks")
		default:
			// Default redirect for users without specific roles or other roles
			fmt.Printf("DEBUG: Redirecting user (ID: %d, role: '%s') to /outbreaks (default)\n", id, primaryRole)
			sl.Info("Redirecting user to outbreaks (default)", "user_id", id, "role", primaryRole)
			return c.Redirect("/outbreaks")
		}
	}

	return nil
}

func HandlerLoginOut(c *fiber.Ctx, sl *slog.Logger, store *session.Store, config Config) error {

	sess, err := store.Get(c)
	if err != nil {
		//return c.Status(fiber.StatusInternalServerError).SendString("Session error")
		sl.Info("Session error")
		return c.Redirect("/login")
	}

	// Destroy session
	sess.Destroy()
	return c.Redirect("/login")
}

func HandlerLoginForgot(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {

	sess, err := store.Get(c)
	if err == nil {
		userID := sess.Get("user")
		if userID != nil {
			return c.Redirect("/", 302)
		}
	}

	// load page
	data := map[string]string{"Title": "Forgot Password and/or username"}
	return GenerateHTML(c, db, data, "forgot")
}

func HandlerHelp(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	data := map[string]string{"Title": "Help Page"}
	return GenerateHTML(c, db, data, "help")
}
