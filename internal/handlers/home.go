package handlers

import (
	"case/internal/models"
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	_ "github.com/lib/pq"
)

// HandlerHome handles the home page
func HandlerHome(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	// Get the current user
	userID, username := GetUser(c, sl, store)

	// Check if an outbreak is selected
	selectedOutbreakID := GetSelectedOutbreak(c, store)

	// Get user's primary role to check if they're admin
	primaryRole, err := getUserPrimaryRole(c, db, userID)
	if err != nil {
		sl.Error("Failed to get user role for home access", "error", err, "user_id", userID)
		primaryRole = ""
	}

	// If no outbreak is selected and user is not admin, try to auto-select outbreak
	if selectedOutbreakID == 0 && primaryRole != "super_admin" && primaryRole != "admin" {
		// Try to get user's default outbreak
		defaultOutbreakID, err := getDefaultOutbreakID(c, db, userID)
		if err != nil {
			sl.Error("Failed to get default outbreak ID for home access", "error", err, "user_id", userID)
			return c.Redirect("/outbreaks")
		}

		if defaultOutbreakID > 0 {
			// User has exactly one active outbreak - set it in session
			sess, err := store.Get(c)
			if err != nil {
				sl.Error("Failed to get session for outbreak selection", "error", err, "user_id", userID)
				return c.Redirect("/outbreaks")
			}

			sess.Set("outbreak_id", defaultOutbreakID)
			sess.Set("selected_outbreak", defaultOutbreakID)
			if err := sess.Save(); err != nil {
				sl.Error("Failed to save session with outbreak selection", "error", err, "user_id", userID)
				return c.Redirect("/outbreaks")
			}

			selectedOutbreakID = defaultOutbreakID
			sl.Info("Auto-selected outbreak for user", "user_id", userID, "outbreak_id", defaultOutbreakID)
		} else {
			// User has multiple outbreaks or no outbreaks - redirect to selection
			return c.Redirect("/outbreaks")
		}
	}

	data := NewTemplateData(c, store)
	data.User = username

	// Only try to get outbreak data if an outbreak is selected
	if selectedOutbreakID > 0 {
		// Get the selected outbreak
		outbreak, err := models.OutbreakByID(c.Context(), db, selectedOutbreakID)
		if err != nil {
			sl.Error("Failed to get selected outbreak: " + err.Error())
			// For admin users, continue without outbreak data
			if primaryRole != "super_admin" && primaryRole != "admin" {
				return c.Redirect("/outbreaks")
			}
		} else {
			data.Form = outbreak
		}
	}

	// Get user permissions for access control
	permissions, err := getUserPermissions(c, db, userID)
	if err != nil {
		sl.Error("Failed to get user permissions: " + err.Error())
		// Continue without permissions - user will see limited access
		permissions = make(map[string][]string)
	}

	// Check if user has case-related roles
	hasCaseRole, err := hasCaseRole(c, db, userID)
	if err != nil {
		sl.Error("Failed to check case role: " + err.Error())
		hasCaseRole = false
	}

	data.UserPermissions = permissions

	// Ensure Optionz is properly initialized
	if data.Optionz == nil {
		data.Optionz = make(map[string]map[string]string)
	}

	// Initialize has_case_role if it doesn't exist
	if data.Optionz["has_case_role"] == nil {
		data.Optionz["has_case_role"] = make(map[string]string)
	}

	data.Optionz["has_case_role"]["value"] = strconv.FormatBool(hasCaseRole)

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
	// Check if user has multiple active outbreaks
	query := `
		SELECT COUNT(*) 
		FROM user_outbreaks uo
		JOIN outbreaks o ON uo.outbreak_id = o.id
		WHERE uo.user_id = $1 AND uo.is_active = true AND o.status = 'active'
	`

	var outbreakCount int
	err := db.QueryRowContext(c.Context(), query, userID).Scan(&outbreakCount)
	if err != nil {
		fmt.Printf("DEBUG: Error counting outbreaks for user %d: %v\n", userID, err)
		return 0, err
	}

	fmt.Printf("DEBUG: User %d has %d active outbreaks\n", userID, outbreakCount)

	// If user has multiple active outbreaks, return 0 to redirect to selection page
	if outbreakCount > 1 {
		fmt.Printf("DEBUG: User %d has multiple outbreaks, returning 0\n", userID)
		return 0, nil
	}

	// If user has exactly one active outbreak, return it
	if outbreakCount == 1 {
		query = `
			SELECT uo.outbreak_id 
			FROM user_outbreaks uo
			JOIN outbreaks o ON uo.outbreak_id = o.id
			WHERE uo.user_id = $1 AND uo.is_active = true AND o.status = 'active'
			ORDER BY uo.assigned_at DESC
			LIMIT 1
		`

		var outbreakID int
		err := db.QueryRowContext(c.Context(), query, userID).Scan(&outbreakID)
		if err != nil {
			if err == sql.ErrNoRows {
				fmt.Printf("DEBUG: No outbreak found for user %d despite count being 1\n", userID)
				return 0, nil
			}
			fmt.Printf("DEBUG: Error getting outbreak ID for user %d: %v\n", userID, err)
			return 0, err
		}
		fmt.Printf("DEBUG: User %d has exactly one outbreak: %d\n", userID, outbreakID)
		return outbreakID, nil
	}

	// If user has no active outbreaks, return 0
	fmt.Printf("DEBUG: User %d has no active outbreaks\n", userID)
	return 0, nil
}

// hasCaseRole checks if a user has any role containing "case" in the name
func hasCaseRole(c *fiber.Ctx, db *sql.DB, userID int) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND r.is_active = true 
		AND LOWER(r.name) LIKE '%case%'
	`

	var count int
	err := db.QueryRowContext(c.Context(), query, userID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
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
	data := NewTemplateData(c, store)
	data.Form = map[string]string{"Title": "Login Page"}
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

		// Get user's default outbreak and set it in session
		defaultOutbreakID, err := getDefaultOutbreakID(c, db, id)
		if err != nil {
			sl.Error("Failed to get default outbreak ID for session", "error", err, "user_id", id)
			// Continue without outbreak_id if there's an error
		} else if defaultOutbreakID > 0 {
			// Only set outbreak_id if user has exactly one active outbreak
			sess.Set("outbreak_id", defaultOutbreakID)
			sess.Set("selected_outbreak", defaultOutbreakID) // Set both keys for consistency
			sl.Info("Set default outbreak in session", "user_id", id, "outbreak_id", defaultOutbreakID)
		} else {
			// User has multiple outbreaks or no outbreaks - don't set outbreak_id
			sl.Info("User has multiple outbreaks or no outbreaks - outbreak selection required", "user_id", id)
		}

		// Get user's facility and set it in session
		userFacility, err := GetUserFacility(c, db, id)
		if err != nil {
			sl.Error("Failed to get user facility for session", "error", err, "user_id", id)
			// Continue without facility_id if there's an error
		} else {
			sess.Set("facility_id", userFacility)
			sl.Info("Set user facility in session", "user_id", id, "facility_id", userFacility)
		}

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
			// Case managers go to /cases with default outbreak or /outbreaks for selection
			defaultOutbreakID, err := getDefaultOutbreakID(c, db, id)
			if err != nil {
				sl.Error("Failed to get default outbreak ID for case manager", "error", err, "user_id", id)
				return c.Redirect("/outbreaks") // Fallback to outbreaks if lookup fails
			}
			if defaultOutbreakID > 0 {
				// User has exactly one active outbreak - redirect to cases
				fmt.Printf("DEBUG: Redirecting case manager (ID: %d, role: %s) to /cases/%d\n", id, primaryRole, defaultOutbreakID)
				sl.Info("Redirecting case manager to cases with default outbreak", "user_id", id, "role", primaryRole, "outbreak_id", defaultOutbreakID)
				return c.Redirect(fmt.Sprintf("/cases/%d", defaultOutbreakID))
			} else {
				// User has multiple outbreaks or no outbreaks - redirect to outbreaks selection
				fmt.Printf("DEBUG: Redirecting case manager (ID: %d, role: %s) to /outbreaks (multiple outbreaks or no outbreaks)\n", id, primaryRole)
				sl.Info("Redirecting case manager to outbreaks selection", "user_id", id, "role", primaryRole)
				return c.Redirect("/outbreaks")
			}
		case "outbreak_viewer", "outbreak_manager":
			// Outbreak users go to outbreaks page
			fmt.Printf("DEBUG: Redirecting outbreak user (ID: %d, role: %s) to /outbreaks\n", id, primaryRole)
			sl.Info("Redirecting outbreak user to outbreaks", "user_id", id, "role", primaryRole)
			return c.Redirect("/outbreaks")
		case "super_admin", "admin":
			// Admin users go directly to home page (they can access everything)
			fmt.Printf("DEBUG: Redirecting admin user (ID: %d, role: %s) to /\n", id, primaryRole)
			sl.Info("Redirecting admin user to home", "user_id", id, "role", primaryRole)
			return c.Redirect("/")
		default:
			// Check if the role contains "case" (for any case-related roles)
			if strings.Contains(strings.ToLower(primaryRole), "case") {
				// Case users go directly to /cases page
				fmt.Printf("DEBUG: Redirecting case user (ID: %d, role: %s) to /cases\n", id, primaryRole)
				sl.Info("Redirecting case user to cases", "user_id", id, "role", primaryRole)
				return c.Redirect("/cases")
			}
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
	data := NewTemplateData(c, store)
	data.Form = map[string]string{"Title": "Forgot Password and/or username"}
	return GenerateHTML(c, db, data, "forgot")
}

func HandlerHelp(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	data := NewTemplateData(c, store)
	data.Form = map[string]string{"Title": "Help Page"}
	return GenerateHTML(c, db, data, "help")
}
