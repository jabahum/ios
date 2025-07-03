package handlers

import (
	"case/internal/models"
	"database/sql"
)

// Handlers contains all handler instances
type Handlers struct {
	AuthHandler               *AuthHandler
	PasswordHandler           *PasswordHandler
	DashboardHandler          *DashboardHandler
	UserHandler               *UserHandler
	EmployeeHandler           *EmployeeHandler
	OutbreakAssignmentHandler *OutbreakAssignmentHandler
	PatientHandler            *PatientHandler
	VHFHandler                *VHFHandler
	RoleHandler               *RoleHandler
	PermissionHandler         *PermissionHandler
}

// NewHandlers creates a new Handlers instance with all handler dependencies
func NewHandlers(db *sql.DB) *Handlers {
	// Create service instances using correct constructors
	userService := models.NewUserService(db)
	employeeService := models.NewEmployeeService(db)
	userOutbreakService := models.NewUserOutbreakService(db)
	patientRoleService := models.NewPatientManagementRoleService(db)
	outbreakService := models.NewOutbreakService(db)
	facilityService := models.NewFacilityService(db)

	// Create handler instances
	authHandler := &AuthHandler{}
	passwordHandler := NewPasswordHandler(userService)
	dashboardHandler := &DashboardHandler{}
	userHandler := &UserHandler{}
	employeeHandler := NewEmployeeHandler(employeeService, userService)
	outbreakAssignmentHandler := NewOutbreakAssignmentHandler(
		userOutbreakService, patientRoleService, userService, outbreakService, facilityService,
	)
	patientHandler := &PatientHandler{}
	vhfHandler := &VHFHandler{}
	roleHandler := &RoleHandler{}
	permissionHandler := &PermissionHandler{}

	return &Handlers{
		AuthHandler:               authHandler,
		PasswordHandler:           passwordHandler,
		DashboardHandler:          dashboardHandler,
		UserHandler:               userHandler,
		EmployeeHandler:           employeeHandler,
		OutbreakAssignmentHandler: outbreakAssignmentHandler,
		PatientHandler:            patientHandler,
		VHFHandler:                vhfHandler,
		RoleHandler:               roleHandler,
		PermissionHandler:         permissionHandler,
	}
}
