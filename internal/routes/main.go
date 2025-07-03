package routes

import (
	"case/internal/handlers"
	"case/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupAllRoutes sets up all application routes
func SetupAllRoutes(router *gin.Engine, handlers *handlers.Handlers) {
	// Public routes
	setupPublicRoutes(router, handlers)

	// Protected routes (require authentication)
	setupProtectedRoutes(router, handlers)
}

// setupPublicRoutes sets up public routes that don't require authentication
func setupPublicRoutes(router *gin.Engine, handlers *handlers.Handlers) {
	// Authentication routes
	auth := router.Group("/auth")
	{
		auth.GET("/login", handlers.AuthHandler.ShowLoginForm)
		auth.POST("/login", handlers.AuthHandler.Login)
		auth.GET("/logout", handlers.AuthHandler.Logout)
		auth.GET("/forgot-password", handlers.PasswordHandler.ShowForgotPasswordForm)
		auth.POST("/forgot-password", handlers.PasswordHandler.RequestPasswordReset)
		auth.GET("/reset-password/:token", handlers.PasswordHandler.ShowResetPasswordForm)
		auth.POST("/reset-password/:token", handlers.PasswordHandler.ResetPassword)
	}

	// Public API endpoints (if any)
	// api := router.Group("/api/public")
	// {
	// 	// Add any public API endpoints here
	// }
}

// setupProtectedRoutes sets up routes that require authentication
func setupProtectedRoutes(router *gin.Engine, handlers *handlers.Handlers) {
	// Apply authentication middleware to all protected routes
	protected := router.Group("/")
	protected.Use(middleware.GinAuthRequired())

	// Dashboard
	protected.GET("/dashboard", handlers.DashboardHandler.ShowDashboard)

	// User management routes
	setupUserRoutes(router, handlers.UserHandler)

	// Employee management routes
	setupEmployeeRoutes(router, handlers.EmployeeHandler)

	// Outbreak assignment routes
	setupOutbreakAssignmentRoutes(router, handlers.OutbreakAssignmentHandler)

	// Patient management routes
	setupPatientRoutes(router, handlers.PatientHandler)

	// VHF patient routes
	setupVHFRoutes(router, handlers.VHFHandler)

	// Role and permission management routes
	setupRolePermissionRoutes(router, handlers.RoleHandler, handlers.PermissionHandler)

	// Password change route
	protected.GET("/change-password", handlers.PasswordHandler.ShowChangePasswordForm)
	protected.POST("/change-password", handlers.PasswordHandler.ChangePassword)

	// API routes for AJAX requests
	api := router.Group("/api")
	api.Use(middleware.GinAuthRequired())
	{
		// User API
		api.GET("/users", handlers.UserHandler.GetUsers)
		api.GET("/users/:id", handlers.UserHandler.GetUser)
		api.POST("/users", handlers.UserHandler.CreateUser)
		api.PUT("/users/:id", handlers.UserHandler.UpdateUser)
		api.DELETE("/users/:id", handlers.UserHandler.DeleteUser)

		// Employee API
		api.GET("/employees", handlers.EmployeeHandler.GetEmployees)
		api.GET("/employees/:id", handlers.EmployeeHandler.GetEmployeeDetails)
		api.POST("/employees", handlers.EmployeeHandler.SaveEmployee)
		api.DELETE("/employees/:id", handlers.EmployeeHandler.DeleteEmployee)
		api.GET("/employees/:id/assignments", handlers.EmployeeHandler.GetEmployeeAssignments)
		api.POST("/employees/assign", handlers.EmployeeHandler.AssignEmployee)
		api.DELETE("/employees/assignments/:assignment_id", handlers.EmployeeHandler.RemoveEmployeeAssignment)

		// Outbreak API
		api.GET("/outbreaks/my-outbreaks", handlers.OutbreakAssignmentHandler.GetUserOutbreaks)
		api.POST("/outbreaks/:id/assign", handlers.OutbreakAssignmentHandler.AssignUserToOutbreak)
		api.DELETE("/outbreaks/:outbreak_id/users/:user_id", handlers.OutbreakAssignmentHandler.RemoveUserFromOutbreak)

		// Patient role API
		api.POST("/patient-roles/assign", handlers.OutbreakAssignmentHandler.AssignPatientRole)
		api.DELETE("/patient-roles/remove", handlers.OutbreakAssignmentHandler.RemovePatientRole)
		api.GET("/patient-roles/user/:user_id", handlers.OutbreakAssignmentHandler.GetUserPatientRoles)
		api.GET("/patient-roles/check-permission", handlers.OutbreakAssignmentHandler.CheckPatientPermission)

		// Role and permission API
		api.GET("/roles", handlers.RoleHandler.GetRoles)
		api.POST("/roles", handlers.RoleHandler.CreateRole)
		api.PUT("/roles/:id", handlers.RoleHandler.UpdateRole)
		api.DELETE("/roles/:id", handlers.RoleHandler.DeleteRole)
		api.GET("/permissions", handlers.PermissionHandler.GetPermissions)
		api.POST("/permissions", handlers.PermissionHandler.CreatePermission)
		api.PUT("/permissions/:id", handlers.PermissionHandler.UpdatePermission)
		api.DELETE("/permissions/:id", handlers.PermissionHandler.DeletePermission)
	}
}

// setupUserRoutes sets up user management routes
func setupUserRoutes(router *gin.Engine, userHandler *handlers.UserHandler) {
	users := router.Group("/users")
	users.Use(middleware.GinAuthRequired())
	{
		users.GET("", userHandler.ListUsers)
		users.GET("/new", userHandler.ShowUserForm)
		users.GET("/edit/:id", userHandler.ShowUserForm)
		users.POST("/save", userHandler.SaveUser)
		users.DELETE("/:id", userHandler.DeleteUser)
	}
}

// setupPatientRoutes sets up patient management routes
func setupPatientRoutes(router *gin.Engine, patientHandler *handlers.PatientHandler) {
	patients := router.Group("/patients")
	patients.Use(middleware.GinAuthRequired())
	{
		patients.GET("", patientHandler.ListPatients)
		patients.GET("/new", patientHandler.ShowPatientForm)
		patients.GET("/edit/:id", patientHandler.ShowPatientForm)
		patients.POST("/save", patientHandler.SavePatient)
		patients.DELETE("/:id", patientHandler.DeletePatient)
		patients.GET("/:id", patientHandler.ShowPatientDetails)
	}
}

// setupVHFRoutes sets up VHF patient routes
func setupVHFRoutes(router *gin.Engine, vhfHandler *handlers.VHFHandler) {
	vhf := router.Group("/vhf")
	vhf.Use(middleware.GinAuthRequired())
	{
		vhf.GET("", vhfHandler.ListVHFPatients)
		vhf.GET("/new", vhfHandler.ShowVHFForm)
		vhf.GET("/edit/:id", vhfHandler.ShowVHFForm)
		vhf.POST("/save", vhfHandler.SaveVHFPatient)
		vhf.DELETE("/:id", vhfHandler.DeleteVHFPatient)
		vhf.GET("/:id", vhfHandler.ShowVHFDetails)
	}
}

// setupRolePermissionRoutes sets up role and permission management routes
func setupRolePermissionRoutes(router *gin.Engine, roleHandler *handlers.RoleHandler, permissionHandler *handlers.PermissionHandler) {
	// Role management
	roles := router.Group("/roles")
	roles.Use(middleware.GinAuthRequired())
	{
		roles.GET("", roleHandler.ListRoles)
		roles.GET("/new", roleHandler.ShowRoleForm)
		roles.GET("/edit/:id", roleHandler.ShowRoleForm)
		roles.POST("/save", roleHandler.SaveRole)
		roles.DELETE("/:id", roleHandler.DeleteRole)
	}

	// Permission management
	permissions := router.Group("/permissions")
	permissions.Use(middleware.GinAuthRequired())
	{
		permissions.GET("", permissionHandler.ListPermissions)
		permissions.GET("/new", permissionHandler.ShowPermissionForm)
		permissions.GET("/edit/:id", permissionHandler.ShowPermissionForm)
		permissions.POST("/save", permissionHandler.SavePermission)
		permissions.DELETE("/:id", permissionHandler.DeletePermission)
	}
}

// setupEmployeeRoutes sets up employee management routes
func setupEmployeeRoutes(router *gin.Engine, employeeHandler *handlers.EmployeeHandler) {
	employees := router.Group("/employees")
	employees.Use(middleware.GinAuthRequired())
	{
		employees.GET("", employeeHandler.ListEmployees)
		employees.GET("/new", employeeHandler.ShowEmployeeForm)
		employees.GET("/edit/:id", employeeHandler.ShowEmployeeForm)
		employees.POST("/save", employeeHandler.SaveEmployee)
		employees.DELETE("/:id", employeeHandler.DeleteEmployee)
		employees.GET("/:id", employeeHandler.ShowEmployeeDetails)
	}
}

// setupOutbreakAssignmentRoutes sets up outbreak assignment routes
func setupOutbreakAssignmentRoutes(router *gin.Engine, outbreakAssignmentHandler *handlers.OutbreakAssignmentHandler) {
	outbreaks := router.Group("/outbreaks")
	outbreaks.Use(middleware.GinAuthRequired())
	{
		outbreaks.GET("/assignments", outbreakAssignmentHandler.ShowOutbreakAssignments)
		outbreaks.GET("/assign", outbreakAssignmentHandler.ShowAssignOutbreakForm)
		outbreaks.POST("/assign", outbreakAssignmentHandler.AssignUserToOutbreak)
		outbreaks.DELETE("/:outbreak_id/users/:user_id", outbreakAssignmentHandler.RemoveUserFromOutbreak)
	}

	patientRoles := router.Group("/patient-roles")
	patientRoles.Use(middleware.GinAuthRequired())
	{
		patientRoles.GET("", outbreakAssignmentHandler.ShowPatientRoles)
		patientRoles.GET("/assign", outbreakAssignmentHandler.ShowAssignPatientRoleForm)
		patientRoles.POST("/assign", outbreakAssignmentHandler.AssignPatientRole)
		patientRoles.DELETE("/remove", outbreakAssignmentHandler.RemovePatientRole)
	}
}
