package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"case/internal/models"
	"case/internal/utils"

	"github.com/gin-gonic/gin"
)

// EmployeeHandler handles employee management operations
type EmployeeHandler struct {
	employeeService *models.EmployeeService
	userService     *models.UserService
}

// NewEmployeeHandler creates a new employee handler
func NewEmployeeHandler(employeeService *models.EmployeeService, userService *models.UserService) *EmployeeHandler {
	return &EmployeeHandler{
		employeeService: employeeService,
		userService:     userService,
	}
}

// ListEmployees shows the employee listing page
func (h *EmployeeHandler) ListEmployees(c *gin.Context) {
	// Get query parameters for filtering
	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if department := c.Query("department"); department != "" {
		filters["department"] = department
	}
	if facility := c.Query("facility"); facility != "" {
		if facilityID, err := strconv.ParseInt(facility, 10, 64); err == nil {
			filters["facility"] = facilityID
		}
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}

	// Get employees with filters
	employees, err := h.employeeService.GetEmployees(filters)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	// Get statistics
	stats, err := h.employeeService.GetEmployeeStatistics()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	// Get departments and titles for filters
	departments, err := h.employeeService.GetEmployeeDepartments()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	titles, err := h.employeeService.GetEmployeeTitles()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "list_employees.html", gin.H{
		"Employees":   employees,
		"Stats":       stats,
		"Departments": departments,
		"Titles":      titles,
		"Filters":     filters,
	})
}

// ShowEmployeeForm shows the employee creation/editing form
func (h *EmployeeHandler) ShowEmployeeForm(c *gin.Context) {
	employeeID := c.Param("id")
	var employee *models.ExtendedEmployee
	var err error

	if employeeID != "0" {
		// Editing existing employee
		id, err := strconv.ParseInt(employeeID, 10, 64)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid employee ID"})
			return
		}
		employee, err = h.employeeService.GetEmployee(id)
		if err != nil {
			c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Employee not found"})
			return
		}
	} else {
		// Creating new employee
		employee = &models.ExtendedEmployee{}
	}

	// Get departments and titles for dropdowns
	departments, err := h.employeeService.GetEmployeeDepartments()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	titles, err := h.employeeService.GetEmployeeTitles()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	// Get all employees for supervisor selection
	allEmployees, err := h.employeeService.GetEmployees(nil)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "form_employee.html", gin.H{
		"Employee":    employee,
		"Departments": departments,
		"Titles":      titles,
		"Employees":   allEmployees,
		"IsNew":       employeeID == "0",
	})
}

// SaveEmployee saves an employee (create or update)
func (h *EmployeeHandler) SaveEmployee(c *gin.Context) {
	var employee models.Employee

	// Parse form data
	if err := c.ShouldBind(&employee); err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": err.Error()})
		return
	}

	// Get current user ID from session
	currentUserID := utils.GetUserIDFromSession(c)
	if currentUserID == 0 {
		c.HTML(http.StatusUnauthorized, "error.html", gin.H{"error": "Unauthorized"})
		return
	}

	var err error
	if employee.EmployeeID == 0 {
		// Creating new employee
		err = h.employeeService.CreateEmployee(&employee, currentUserID)
	} else {
		// Updating existing employee
		err = h.employeeService.UpdateEmployee(&employee, currentUserID)
	}

	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	// Redirect based on form action
	from := c.PostForm("from")
	if from == "close" {
		c.Redirect(http.StatusSeeOther, "/employees")
	} else {
		c.Redirect(http.StatusSeeOther, "/employees/new/"+strconv.FormatInt(int64(employee.EmployeeID), 10))
	}
}

// DeleteEmployee deletes an employee
func (h *EmployeeHandler) DeleteEmployee(c *gin.Context) {
	id := c.Param("id")
	employeeID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}

	err = h.employeeService.DeleteEmployee(employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Employee deleted successfully"})
}

// GetEmployeeDetails gets employee details for AJAX requests
func (h *EmployeeHandler) GetEmployeeDetails(c *gin.Context) {
	id := c.Param("id")
	employeeID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}

	employee, err := h.employeeService.GetEmployee(employeeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"employee": employee})
}

// AssignEmployee assigns an employee to a facility
func (h *EmployeeHandler) AssignEmployee(c *gin.Context) {
	var req struct {
		EmployeeID     int64  `json:"employee_id" binding:"required"`
		AssignmentType string `json:"assignment_type" binding:"required"`
		AssignmentID   int64  `json:"assignment_id" binding:"required"`
		AssignmentName string `json:"assignment_name" binding:"required"`
		StartDate      string `json:"start_date" binding:"required"`
		EndDate        string `json:"end_date"`
		IsPrimary      bool   `json:"is_primary"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse start date
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
		return
	}

	// Parse end date if provided
	var endDate sql.NullTime
	if req.EndDate != "" {
		parsedEndDate, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
			return
		}
		endDate = sql.NullTime{Time: parsedEndDate, Valid: true}
	}

	// Get current user ID from session
	currentUserID := utils.GetUserIDFromSession(c)
	if currentUserID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	assignment := &models.EmployeeAssignment{
		EmployeeID:     req.EmployeeID,
		AssignmentType: req.AssignmentType,
		AssignmentID:   req.AssignmentID,
		AssignmentName: req.AssignmentName,
		StartDate:      startDate,
		EndDate:        endDate,
		IsPrimary:      req.IsPrimary,
		IsActive:       true,
	}

	err = h.employeeService.AssignEmployee(assignment, currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Employee assigned successfully"})
}

// RemoveEmployeeAssignment removes an employee assignment
func (h *EmployeeHandler) RemoveEmployeeAssignment(c *gin.Context) {
	assignmentID := c.Param("assignment_id")
	id, err := strconv.ParseInt(assignmentID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment ID"})
		return
	}

	err = h.employeeService.RemoveEmployeeAssignment(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Assignment removed successfully"})
}

// GetEmployeeAssignments gets assignments for an employee
func (h *EmployeeHandler) GetEmployeeAssignments(c *gin.Context) {
	id := c.Param("id")
	employeeID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}

	assignments, err := h.employeeService.GetEmployeeAssignments(employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"assignments": assignments})
}

// ExportEmployees exports employee data
func (h *EmployeeHandler) ExportEmployees(c *gin.Context) {
	// Get all employees
	employees, err := h.employeeService.GetEmployees(nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set response headers for CSV download
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=employees.csv")

	// Write CSV header
	c.Writer.WriteString("Employee ID,First Name,Last Name,Sex,Email,Phone,Cadre,Facility\n")

	// Write employee data
	for _, emp := range employees {
		facilityName := "Unknown"
		if emp.FacilityInfo != nil && emp.FacilityInfo.FacilityName.Valid {
			facilityName = emp.FacilityInfo.FacilityName.String
		}

		line := fmt.Sprintf("%d,%s,%s,%s,%s,%s,%s,%s\n",
			emp.EmployeeID,
			emp.EmployeeFname.String,
			emp.EmployeeLname.String,
			emp.EmployeeSex.String,
			emp.EmployeeEmail.String,
			emp.EmployeePhone.String,
			emp.EmployeeCadre.String,
			facilityName,
		)
		c.Writer.WriteString(line)
	}
}

// GetEmployeeStatistics gets employee statistics for dashboard
func (h *EmployeeHandler) GetEmployeeStatistics(c *gin.Context) {
	stats, err := h.employeeService.GetEmployeeStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"statistics": stats})
}

// GetEmployees returns all employees
func (h *EmployeeHandler) GetEmployees(c *gin.Context) {
	employees, err := h.employeeService.GetEmployees(nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"employees": employees})
}

// ShowEmployeeDetails shows employee details page
func (h *EmployeeHandler) ShowEmployeeDetails(c *gin.Context) {
	id := c.Param("id")
	employeeID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid employee ID"})
		return
	}

	employee, err := h.employeeService.GetEmployee(employeeID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Employee not found"})
		return
	}

	// Get employee assignments
	assignments, err := h.employeeService.GetEmployeeAssignments(employeeID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to load assignments"})
		return
	}

	c.HTML(http.StatusOK, "employee_details.html", gin.H{
		"employee":    employee,
		"assignments": assignments,
	})
}
