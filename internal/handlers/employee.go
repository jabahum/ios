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

// HandlerEmployeeForm handles the employee form display
func HandlerEmployeeForm(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userName := GetUser(c, sl, store)
	role := security.GetRoles(userID, "admin")

	id, err := strconv.Atoi(c.Params("i"))
	if err != nil {
		log.Println(err.Error())
	}

	// Create an employee form struct that matches the template expectations
	var employee EmployeeForm
	if id > 0 {
		// Load existing employee
		query := `SELECT employee_id, employee_fname, employee_lname, employee_sex, 
		          employee_email, employee_phone, employee_cadre, facility
		          FROM public.employee WHERE employee_id = $1`
		err := db.QueryRowContext(c.Context(), query, id).Scan(
			&employee.EmployeeID, &employee.EmployeeFname, &employee.EmployeeLname, &employee.EmployeeSex,
			&employee.EmployeeEmail, &employee.EmployeePhone, &employee.EmployeeCadre, &employee.Facility,
		)
		if err != nil {
			sl.Error("Error loading employee", "error", err)
			// Continue with empty employee for new form
		}
	}

	data := NewTemplateData(c, store)
	data.User = userName
	data.Role = role
	data.Employee = &employee
	data.IsNew = id == 0

	// Add titles data
	data.Titles = []EmployeeTitle{
		{ID: 1, Name: "Doctor"},
		{ID: 2, Name: "Nurse"},
		{ID: 3, Name: "Clinical Officer"},
		{ID: 4, Name: "Laboratory Technician"},
		{ID: 5, Name: "Pharmacist"},
		{ID: 6, Name: "Administrator"},
		{ID: 7, Name: "Data Entry Clerk"},
		{ID: 8, Name: "Security Guard"},
		{ID: 9, Name: "Cleaner"},
		{ID: 10, Name: "Driver"},
	}

	// Add departments data
	data.Departments = []Department{
		{ID: 1, Name: "Clinical Services"},
		{ID: 2, Name: "Laboratory"},
		{ID: 3, Name: "Pharmacy"},
		{ID: 4, Name: "Administration"},
		{ID: 5, Name: "Support Services"},
		{ID: 6, Name: "Security"},
	}

	// Add facilities data
	data.Facilities = []Facility{}
	facilityQuery := `SELECT facility_id, facility_name FROM public.facility ORDER BY facility_name`
	facilityRows, err := db.QueryContext(c.Context(), facilityQuery)
	if err == nil {
		defer facilityRows.Close()
		for facilityRows.Next() {
			var facility Facility
			if err := facilityRows.Scan(&facility.ID, &facility.Name); err == nil {
				data.Facilities = append(data.Facilities, facility)
			}
		}
	}

	// Add employees data for supervisor selection
	data.Employees = []EmployeeForm{}
	employeeQuery := `SELECT employee_id, employee_fname, employee_lname FROM public.employee ORDER BY employee_fname, employee_lname`
	employeeRows, err := db.QueryContext(c.Context(), employeeQuery)
	if err == nil {
		defer employeeRows.Close()
		for employeeRows.Next() {
			var emp EmployeeForm
			if err := employeeRows.Scan(&emp.EmployeeID, &emp.EmployeeFname, &emp.EmployeeLname); err == nil {
				data.Employees = append(data.Employees, emp)
			}
		}
	}

	return GenerateHTML(c, db, data, "form_employee")
}

// HandlerEmployeeSubmit handles employee form submission
func HandlerEmployeeSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	_, _ = GetUser(c, sl, store) // Get user info but not used in this function

	// Parse form data
	employeeID, err := strconv.Atoi(c.FormValue("employee_id"))
	if err != nil {
		employeeID = 0
	}

	employee := models.Employee{
		EmployeeID:    employeeID,
		EmployeeFname: ParseNullString(c.FormValue("employee_fname")),
		EmployeeLname: ParseNullString(c.FormValue("employee_lname")),
		EmployeeSex:   ParseNullString(c.FormValue("employee_sex")),
		EmployeeEmail: ParseNullString(c.FormValue("employee_email")),
		EmployeePhone: ParseNullString(c.FormValue("employee_phone")),
		EmployeeCadre: ParseNullString(c.FormValue("employee_cadre")),
		Facility:      ParseNullInt(c.FormValue("facility")),
	}

	if employee.EmployeeID == 0 {
		// Creating new employee
		err := employee.Insert(c.Context(), db)
		if err != nil {
			sl.Error("Error creating employee", "error", err)
			return c.Status(500).SendString("Error creating employee")
		}
	} else {
		// Updating existing employee
		employee.SetAsExists()
		err := employee.Update(c.Context(), db)
		if err != nil {
			sl.Error("Error updating employee", "error", err)
			return c.Status(500).SendString("Error updating employee")
		}
	}

	// Redirect based on form action
	from := c.FormValue("from")
	if from == "close" {
		return c.Redirect("/employees")
	} else {
		return c.Redirect(fmt.Sprintf("/employees/new/%d", employee.EmployeeID))
	}
}

// HandlerEmployeeList handles the employee list display
func HandlerEmployeeList(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
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
		TotalEmployees:  0,
	}

	// Get employee statistics
	statsQuery := `SELECT COUNT(*) as total_employees FROM public.employee`
	var totalEmployees int
	err := db.QueryRowContext(c.Context(), statsQuery).Scan(&totalEmployees)
	if err != nil {
		sl.Error("Failed to get employee statistics", "error", err)
	} else {
		data.Stats.TotalEmployees = totalEmployees
	}

	// Get active employees count
	activeQuery := `SELECT COUNT(*) as active_employees FROM public.employee WHERE employee_status = 'active'`
	var activeEmployees int
	err = db.QueryRowContext(c.Context(), activeQuery).Scan(&activeEmployees)
	if err != nil {
		sl.Error("Failed to get active employee statistics", "error", err)
	} else {
		data.Stats.ActiveUsers = activeEmployees
	}

	// Get facilities count
	facilityQuery := `SELECT COUNT(*) as total_facilities FROM public.facility`
	var totalFacilities int
	err = db.QueryRowContext(c.Context(), facilityQuery).Scan(&totalFacilities)
	if err != nil {
		sl.Error("Failed to get facility statistics", "error", err)
	} else {
		data.Stats.TotalFacilities = totalFacilities
	}

	// Get employees with basic information
	query := `SELECT employee_id, employee_fname, employee_lname, employee_sex, 
	          employee_email, employee_phone, employee_cadre, facility
	          FROM public.employee ORDER BY employee_fname, employee_lname`

	rows, err := db.QueryContext(c.Context(), query)
	if err != nil {
		sl.Error("Database error in employee list", "error", err.Error())
		data.Employees = []EmployeeForm{}
		return GenerateHTML(c, db, data, "list_employees")
	}
	defer rows.Close()

	var employees []EmployeeForm
	for rows.Next() {
		var emp EmployeeForm
		if err := rows.Scan(&emp.EmployeeID, &emp.EmployeeFname, &emp.EmployeeLname, &emp.EmployeeSex,
			&emp.EmployeeEmail, &emp.EmployeePhone, &emp.EmployeeCadre, &emp.Facility); err != nil {
			sl.Error("Row scan error in employee list", "error", err.Error())
			continue
		}
		employees = append(employees, emp)
	}

	if err = rows.Err(); err != nil {
		sl.Error("Rows iteration error in employee list", "error", err.Error())
	}

	data.Employees = employees
	return GenerateHTML(c, db, data, "list_employees")
}
