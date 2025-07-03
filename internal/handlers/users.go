package handlers

import (
	"case/internal/models"
	"case/internal/security"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"strconv"

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

	var uzer models.User
	uzer.UserEmployee.Valid = true
	uzer.UserEmployee.Int64 = 0

	data := NewTemplateData(c, store)

	if id > 0 {
		u, er := models.UserByUserID(c.Context(), db, id)
		if er == nil {
			uzer = *u
		}
	} else {
		id = 0
	}

	fmt.Println("Creating")
	// Correct struct definition with semicolons
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

	// Use parameterized query to prevent SQL injection
	mysql := `  
	SELECT 
		ur.user_rights_id, m.meta_id, m.meta_name, function_scope,
		COALESCE(ur.user_rights_can_view, 0), 
		COALESCE(ur.user_rights_can_create, 0), 
		COALESCE(ur.user_rights_can_edit, 0), 
		COALESCE(ur.user_rights_can_remove, 0)
	FROM meta m 
	LEFT JOIN public.user_right ur 
		ON m.meta_id = ur.user_rights_function AND ur.user_id = $1
	WHERE m.meta_category = 3`

	// Execute query safely with parameterized input
	rows, err := db.QueryContext(c.Context(), mysql, id)
	if err != nil {
		log.Println("Query Error:", err.Error())
	}
	defer rows.Close()

	// Slice to store results
	var functions []funclist

	// Iterate over query results
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

	// Check for errors after looping
	if err = rows.Err(); err != nil {
		log.Println("Rows Iteration Error:", err)
	}

	data.User = userName
	data.Role = role
	data.Form = uzer
	data.FormChild1 = functions

	return GenerateHTML(c, db, data, "form_user")
}

func HandlerUserSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	id, er := strconv.Atoi(c.FormValue("id"))
	if er != nil {
		id = 0
	}

	user := models.User{
		UserID:       id,
		UserName:     ParseNullString(c.FormValue("user_name")),
		UserEmployee: ParseNullInt(c.FormValue("user_employee")),
	}

	if user.UserID == 0 {
		user.UserPass = sql.NullString{String: models.Encrypt("123456"), Valid: true}
		err := user.Insert(c.Context(), db)
		if err != nil {
			log.Println(err.Error())
		}
	} else {
		user.SetAsExists()
		err := user.Update_NoPass(c.Context(), db)
		if err != nil {
			log.Println(err.Error())
		}
	}

	//================================================================

	mysql := `SELECT m.meta_id, m.meta_name FROM meta m WHERE m.meta_category = 3`

	// Execute query safely with parameterized input
	rows, err := db.QueryContext(c.Context(), mysql)
	if err != nil {
		log.Println("Query Error:", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var m_id int64
		var m_nm string
		err := rows.Scan(&m_id, &m_nm)
		if err != nil {
			log.Println("Row Scan Error:", err)
			continue
		}

		fid, err := strconv.ParseInt(c.FormValue("input_fid_id_"+m_nm), 10, 64)
		scope, err := strconv.ParseInt(c.FormValue("input_scope_"+m_nm), 10, 64)
		view, err := strconv.ParseInt(c.FormValue("input_view_"+m_nm), 10, 64)
		add, err := strconv.ParseInt(c.FormValue("input_add_"+m_nm), 10, 64)
		edit, err := strconv.ParseInt(c.FormValue("input_edit_"+m_nm), 10, 64)
		exec, err := strconv.ParseInt(c.FormValue("input_execute_"+m_nm), 10, 64)

		right := models.UserRight{}

		right.UserID.Valid = true
		right.UserRightsFunction.Valid = true
		right.FunctionScope.Valid = true
		right.UserRightsCanView.Valid = true
		right.UserRightsCanCreate.Valid = true
		right.UserRightsCanEdit.Valid = true
		right.UserRightsCanRemove.Valid = true

		right.UserID.Int64 = int64(user.UserID)
		right.UserRightsFunction.Int64 = m_id
		right.FunctionScope.Int64 = scope
		right.UserRightsCanView.Int64 = view
		right.UserRightsCanCreate.Int64 = add
		right.UserRightsCanEdit.Int64 = edit
		right.UserRightsCanRemove.Int64 = exec

		if fid > 0 {
			right.UserRightsID = int(fid)
			right.SetAsExists()
			er := right.Update(c.Context(), db)
			if er != nil {
				log.Println(err.Error())
			}
		} else {
			er := right.Insert(c.Context(), db)
			if er != nil {
				log.Println(err.Error())
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

	data := NewTemplateData(c, store)
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
		Email          sql.NullString
		FirstName      sql.NullString
		LastName       sql.NullString
		IsActive       sql.NullBool
		IsLocked       sql.NullBool
		LastLoginAt    sql.NullTime
		CreatedAt      sql.NullTime
		DepartmentName sql.NullString
		Roles          []Role
	}

	// Enhanced query to include RBAC fields
	mysql := `
		SELECT DISTINCT u.user_id, u.user_name, u.email, u.first_name, u.last_name, 
		       u.is_active, u.is_locked, u.last_login_at, u.created_at, d.name as department_name,
		       array_agg(r.name) as role_names
		FROM public.users u
		LEFT JOIN departments d ON u.department_id = d.id
		LEFT JOIN user_roles ur ON u.user_id = ur.user_id
		LEFT JOIN roles r ON ur.role_id = r.id
		GROUP BY u.user_id, d.name, u.last_login_at, u.created_at
		ORDER BY u.user_name
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
		var roleNames []string
		if err := rows.Scan(&u.UserID, &u.UserName, &u.Email, &u.FirstName, &u.LastName,
			&u.IsActive, &u.IsLocked, &u.LastLoginAt, &u.CreatedAt, &u.DepartmentName, &roleNames); err != nil {
			sl.Error("Row scan error in user list", "error", err.Error())
			continue
		}

		// Convert role names to Role structs
		for _, roleName := range roleNames {
			u.Roles = append(u.Roles, Role{Name: roleName})
		}

		users = append(users, u)
	}

	// Check for iteration errors
	if err = rows.Err(); err != nil {
		sl.Error("Rows iteration error in user list", "error", err.Error())
	}

	data.Form = users
	return GenerateHTML(c, db, data, "list_users")
}
