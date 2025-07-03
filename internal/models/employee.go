package models

import (
	"database/sql"
)

// EmployeeService provides methods for managing employees
type EmployeeService struct {
	db *sql.DB
}

// Extended Employee struct with related data
type ExtendedEmployee struct {
	Employee
	// Related data
	FacilityInfo  *Facility
	Supervisor    *Employee
	Assignments   []EmployeeAssignment
	CreatedByUser *User
	UpdatedByUser *User
}

// NewEmployeeService creates a new employee service
func NewEmployeeService(db *sql.DB) *EmployeeService {
	return &EmployeeService{db: db}
}

// CreateEmployee creates a new employee
func (s *EmployeeService) CreateEmployee(employee *Employee, createdBy int64) error {
	query := `
		INSERT INTO employee (employee_fname, employee_lname, employee_sex, employee_email, 
		                     employee_phone, employee_cadre, facility)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING employee_id
	`
	err := s.db.QueryRow(
		query,
		employee.EmployeeFname, employee.EmployeeLname, employee.EmployeeSex,
		employee.EmployeeEmail, employee.EmployeePhone, employee.EmployeeCadre,
		employee.Facility,
	).Scan(&employee.EmployeeID)
	return err
}

// UpdateEmployee updates an existing employee
func (s *EmployeeService) UpdateEmployee(employee *Employee, updatedBy int64) error {
	query := `
		UPDATE employee SET
			employee_fname = $1, employee_lname = $2, employee_sex = $3,
			employee_email = $4, employee_phone = $5, employee_cadre = $6,
			facility = $7
		WHERE employee_id = $8
	`
	_, err := s.db.Exec(
		query,
		employee.EmployeeFname, employee.EmployeeLname, employee.EmployeeSex,
		employee.EmployeeEmail, employee.EmployeePhone, employee.EmployeeCadre,
		employee.Facility, employee.EmployeeID,
	)
	return err
}

// GetEmployee gets an employee by ID
func (s *EmployeeService) GetEmployee(employeeID int64) (*ExtendedEmployee, error) {
	query := `
		SELECT e.employee_id, e.employee_fname, e.employee_lname, e.employee_sex,
		       e.employee_email, e.employee_phone, e.employee_cadre, e.facility,
		       f.facility_name
		FROM employee e
		LEFT JOIN facility f ON e.facility = f.facility_id
		WHERE e.employee_id = $1
	`
	employee := &ExtendedEmployee{}
	var facilityName sql.NullString

	err := s.db.QueryRow(query, employeeID).Scan(
		&employee.EmployeeID, &employee.EmployeeFname, &employee.EmployeeLname, &employee.EmployeeSex,
		&employee.EmployeeEmail, &employee.EmployeePhone, &employee.EmployeeCadre, &employee.Facility,
		&facilityName,
	)
	if err != nil {
		return nil, err
	}

	// Set related data
	if facilityName.Valid {
		employee.FacilityInfo = &Facility{FacilityName: facilityName}
	}

	return employee, nil
}

// GetEmployees gets all employees with optional filtering
func (s *EmployeeService) GetEmployees(filters map[string]interface{}) ([]ExtendedEmployee, error) {
	query := `
		SELECT e.employee_id, e.employee_fname, e.employee_lname, e.employee_sex,
		       e.employee_email, e.employee_phone, e.employee_cadre, e.facility,
		       f.facility_name
		FROM employee e
		LEFT JOIN facility f ON e.facility = f.facility_id
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	// Add filters
	if status, ok := filters["status"].(string); ok && status != "" {
		// Note: status field doesn't exist in current schema, so we'll skip this filter
	}
	if department, ok := filters["department"].(string); ok && department != "" {
		// Note: department field doesn't exist in current schema, so we'll skip this filter
	}
	if facility, ok := filters["facility"].(int64); ok && facility > 0 {
		query += " AND e.facility = $" + string(rune(argCount+'0'))
		args = append(args, facility)
		argCount++
	}
	if search, ok := filters["search"].(string); ok && search != "" {
		query += " AND (e.employee_fname ILIKE $" + string(rune(argCount+'0')) +
			" OR e.employee_lname ILIKE $" + string(rune(argCount+'0')) + ")"
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm, searchTerm)
		argCount += 2
	}

	query += " ORDER BY e.employee_fname, e.employee_lname"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []ExtendedEmployee
	for rows.Next() {
		var employee ExtendedEmployee
		var facilityName sql.NullString

		err := rows.Scan(
			&employee.EmployeeID, &employee.EmployeeFname, &employee.EmployeeLname, &employee.EmployeeSex,
			&employee.EmployeeEmail, &employee.EmployeePhone, &employee.EmployeeCadre, &employee.Facility,
			&facilityName,
		)
		if err != nil {
			return nil, err
		}

		// Set related data
		if facilityName.Valid {
			employee.FacilityInfo = &Facility{FacilityName: facilityName}
		}

		employees = append(employees, employee)
	}
	return employees, nil
}

// DeleteEmployee deletes an employee
func (s *EmployeeService) DeleteEmployee(employeeID int64) error {
	query := `DELETE FROM employee WHERE employee_id = $1`
	_, err := s.db.Exec(query, employeeID)
	return err
}

// GetEmployeeAssignments gets assignments for an employee
func (s *EmployeeService) GetEmployeeAssignments(employeeID int64) ([]EmployeeAssignment, error) {
	query := `
		SELECT id, employee_id, assignment_type, assignment_id, assignment_name,
		       start_date, end_date, is_primary, is_active, created_at, created_by
		FROM employee_assignments
		WHERE employee_id = $1 AND is_active = true
		ORDER BY is_primary DESC, start_date DESC
	`
	rows, err := s.db.Query(query, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []EmployeeAssignment
	for rows.Next() {
		var assignment EmployeeAssignment
		err := rows.Scan(
			&assignment.ID, &assignment.EmployeeID, &assignment.AssignmentType,
			&assignment.AssignmentID, &assignment.AssignmentName, &assignment.StartDate,
			&assignment.EndDate, &assignment.IsPrimary, &assignment.IsActive,
			&assignment.CreatedAt, &assignment.CreatedBy,
		)
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment)
	}
	return assignments, nil
}

// AssignEmployee assigns an employee to a role/position
func (s *EmployeeService) AssignEmployee(assignment *EmployeeAssignment, createdBy int64) error {
	query := `
		INSERT INTO employee_assignments (employee_id, assignment_type, assignment_id, 
		                                 assignment_name, start_date, end_date, is_primary, 
		                                 is_active, created_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, $9)
		RETURNING id
	`
	err := s.db.QueryRow(
		query,
		assignment.EmployeeID, assignment.AssignmentType, assignment.AssignmentID,
		assignment.AssignmentName, assignment.StartDate, assignment.EndDate,
		assignment.IsPrimary, assignment.IsActive, createdBy,
	).Scan(&assignment.ID)
	return err
}

// RemoveEmployeeAssignment removes an employee assignment
func (s *EmployeeService) RemoveEmployeeAssignment(assignmentID int64) error {
	query := `UPDATE employee_assignments SET is_active = false WHERE id = $1`
	_, err := s.db.Exec(query, assignmentID)
	return err
}

// GetEmployeeDepartments gets all employee departments
func (s *EmployeeService) GetEmployeeDepartments() ([]EmployeeDepartment, error) {
	query := `
		SELECT id, name, description, is_active, created_at, updated_at
		FROM employee_departments
		WHERE is_active = true
		ORDER BY name
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var departments []EmployeeDepartment
	for rows.Next() {
		var dept EmployeeDepartment
		err := rows.Scan(
			&dept.ID, &dept.Name, &dept.Description, &dept.IsActive,
			&dept.CreatedAt, &dept.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		departments = append(departments, dept)
	}
	return departments, nil
}

// GetEmployeeTitles gets all employee titles
func (s *EmployeeService) GetEmployeeTitles() ([]EmployeeTitle, error) {
	query := `
		SELECT id, name, description, is_active, created_at, updated_at
		FROM employee_titles
		WHERE is_active = true
		ORDER BY name
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var titles []EmployeeTitle
	for rows.Next() {
		var title EmployeeTitle
		err := rows.Scan(
			&title.ID, &title.Name, &title.Description, &title.IsActive,
			&title.CreatedAt, &title.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		titles = append(titles, title)
	}
	return titles, nil
}

// GetEmployeeStatistics gets employee statistics
func (s *EmployeeService) GetEmployeeStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total employees
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM employee").Scan(&total)
	if err != nil {
		return nil, err
	}
	stats["total"] = total

	// Active employees (assuming all are active for now since there's no status field)
	stats["active"] = total

	// Employees by facility
	facilityQuery := `
		SELECT f.facility_name, COUNT(e.employee_id) as count
		FROM employee e
		LEFT JOIN facility f ON e.facility = f.facility_id
		GROUP BY f.facility_name
		ORDER BY count DESC
	`
	rows, err := s.db.Query(facilityQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var facilityStats []map[string]interface{}
	for rows.Next() {
		var facilityName sql.NullString
		var count int
		err := rows.Scan(&facilityName, &count)
		if err != nil {
			return nil, err
		}
		name := "Unknown"
		if facilityName.Valid {
			name = facilityName.String
		}
		facilityStats = append(facilityStats, map[string]interface{}{
			"facility": name,
			"count":    count,
		})
	}
	stats["by_facility"] = facilityStats

	return stats, nil
}
