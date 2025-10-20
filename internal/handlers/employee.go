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
	role := security.GetRoleID(db, userID, "admin")

	id, err := strconv.Atoi(c.Params("i"))
	if err != nil {
		log.Println(err.Error())
	}

	// Create an employee form struct that matches the template expectations
	var employee EmployeeForm
	if id > 0 {
		// Load existing employee
		query := `SELECT employee_id, employee_fname, employee_lname, employee_sex, 
		          employee_email, employee_phone, employee_cadre, facility, afi_facility, afi_region, afi_district
		          FROM public.employee WHERE employee_id = $1`
		err := db.QueryRowContext(c.Context(), query, id).Scan(
			&employee.EmployeeID, &employee.EmployeeFname, &employee.EmployeeLname, &employee.EmployeeSex,
			&employee.EmployeeEmail, &employee.EmployeePhone, &employee.EmployeeCadre, &employee.Facility,
			&employee.AFIFacility, &employee.AFIRegion, &employee.AFIDistrict,
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

	// Load AFI facilities for afi_facility selection
	type AFIFacility struct {
		ID   int
		Name string
	}
	var afiFacilities []AFIFacility
	if rows, err := db.QueryContext(c.Context(), `SELECT id, facility_name FROM afi_facilities ORDER BY facility_name`); err == nil {
		defer rows.Close()
		for rows.Next() {
			var f AFIFacility
			if err := rows.Scan(&f.ID, &f.Name); err == nil {
				afiFacilities = append(afiFacilities, f)
			}
		}
	}

	// Load unique regions from afi_facilities
	var regions []string
	if rows, err := db.QueryContext(c.Context(), `SELECT DISTINCT region FROM afi_facilities WHERE region IS NOT NULL AND region != '' ORDER BY region`); err == nil {
		defer rows.Close()
		for rows.Next() {
			var region string
			if err := rows.Scan(&region); err == nil {
				regions = append(regions, region)
			}
		}
	}

	// Load unique districts from afi_facilities
	var districts []string
	if rows, err := db.QueryContext(c.Context(), `SELECT DISTINCT district FROM afi_facilities WHERE district IS NOT NULL AND district != '' ORDER BY district`); err == nil {
		defer rows.Close()
		for rows.Next() {
			var district string
			if err := rows.Scan(&district); err == nil {
				districts = append(districts, district)
			}
		}
	}

	if data.Optionz == nil {
		data.Optionz = make(map[string]map[string]string)
	}
	data.Optionz["afi_facilities"] = map[string]string{"": "Select AFI Facility"}
	for _, f := range afiFacilities {
		data.Optionz["afi_facilities"][f.Name] = f.Name
	}

	data.Optionz["afi_regions"] = map[string]string{"": "Select Region"}
	for _, r := range regions {
		data.Optionz["afi_regions"][r] = r
	}

	data.Optionz["afi_districts"] = map[string]string{"": "Select District"}
	for _, d := range districts {
		data.Optionz["afi_districts"][d] = d
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

// HandlerGetEmployee handles getting a single employee by ID (API endpoint)
func HandlerGetEmployee(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	employeeID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid employee ID"})
	}

	// Get employee details
	query := `SELECT employee_id, employee_fname, employee_lname, employee_sex, 
	          employee_email, employee_phone, employee_cadre, facility, afi_facility, afi_region, afi_district,
	          employee_start_date, employee_end_date, employee_status, employee_department, employee_title
	          FROM public.employee WHERE employee_id = $1`

	var employee struct {
		EmployeeID         int            `json:"employee_id"`
		EmployeeFname      sql.NullString `json:"employee_fname"`
		EmployeeLname      sql.NullString `json:"employee_lname"`
		EmployeeSex        sql.NullString `json:"employee_sex"`
		EmployeeEmail      sql.NullString `json:"employee_email"`
		EmployeePhone      sql.NullString `json:"employee_phone"`
		EmployeeCadre      sql.NullString `json:"employee_cadre"`
		Facility           sql.NullInt64  `json:"facility"`
		AFIFacility        sql.NullString `json:"afi_facility"`
		AFIRegion          sql.NullString `json:"afi_region"`
		AFIDistrict        sql.NullString `json:"afi_district"`
		EmployeeStartDate  sql.NullTime   `json:"employee_start_date"`
		EmployeeEndDate    sql.NullTime   `json:"employee_end_date"`
		EmployeeStatus     sql.NullString `json:"employee_status"`
		EmployeeDepartment sql.NullString `json:"employee_department"`
		EmployeeTitle      sql.NullString `json:"employee_title"`
	}

	err = db.QueryRowContext(c.Context(), query, employeeID).Scan(
		&employee.EmployeeID, &employee.EmployeeFname, &employee.EmployeeLname, &employee.EmployeeSex,
		&employee.EmployeeEmail, &employee.EmployeePhone, &employee.EmployeeCadre, &employee.Facility,
		&employee.AFIFacility, &employee.AFIRegion, &employee.AFIDistrict,
		&employee.EmployeeStartDate, &employee.EmployeeEndDate, &employee.EmployeeStatus,
		&employee.EmployeeDepartment, &employee.EmployeeTitle,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "Employee not found"})
		}
		sl.Error("Error getting employee", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.JSON(employee)
}

// HandlerDeleteEmployee handles deleting an employee (API endpoint)
func HandlerDeleteEmployee(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	employeeID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid employee ID"})
	}

	// Check if employee exists
	var exists bool
	err = db.QueryRowContext(c.Context(), "SELECT EXISTS(SELECT 1 FROM employee WHERE employee_id = $1)", employeeID).Scan(&exists)
	if err != nil {
		sl.Error("Error checking employee existence", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	if !exists {
		return c.Status(404).JSON(fiber.Map{"error": "Employee not found"})
	}

	// Delete employee
	_, err = db.ExecContext(c.Context(), "DELETE FROM employee WHERE employee_id = $1", employeeID)
	if err != nil {
		sl.Error("Error deleting employee", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete employee"})
	}

	sl.Info("Employee deleted successfully", "employee_id", employeeID)
	return c.JSON(fiber.Map{"message": "Employee deleted successfully"})
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
		AFIFacility:   ParseNullString(c.FormValue("afi_facility")),
		AFIRegion:     ParseNullString(c.FormValue("afi_region")),
		AFIDistrict:   ParseNullString(c.FormValue("afi_district")),
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
	role := security.GetRoleID(db, userID, "admin")

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

	// Get employees with basic information and facility details
	query := `SELECT e.employee_id, e.employee_fname, e.employee_lname, e.employee_sex, 
	          e.employee_email, e.employee_phone, e.employee_cadre, e.facility,
	          f.facility_name
	          FROM public.employee e
	          LEFT JOIN public.facility f ON e.facility = f.facility_id
	          ORDER BY e.employee_fname, e.employee_lname`

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
			&emp.EmployeeEmail, &emp.EmployeePhone, &emp.EmployeeCadre, &emp.Facility, &emp.FacilityInfo.Name); err != nil {
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
