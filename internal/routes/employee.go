package routes

import (
	"case/internal/handlers"
	"case/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupEmployeeRoutes sets up routes for employee management
func SetupEmployeeRoutes(router *gin.Engine, employeeHandler *handlers.EmployeeHandler) {
	// Employee management routes (require authentication)
	employeeGroup := router.Group("/employees")
	employeeGroup.Use(middleware.GinAuthRequired())
	{
		// Employee CRUD operations
		employeeGroup.GET("", employeeHandler.ListEmployees)
		employeeGroup.GET("/new/:id", employeeHandler.ShowEmployeeForm)
		employeeGroup.POST("/save", employeeHandler.SaveEmployee)
		employeeGroup.DELETE("/:id", employeeHandler.DeleteEmployee)

		// Employee details and assignments
		employeeGroup.GET("/:id/details", employeeHandler.GetEmployeeDetails)
		employeeGroup.GET("/:id/assignments", employeeHandler.GetEmployeeAssignments)
		employeeGroup.POST("/assign", employeeHandler.AssignEmployee)
		employeeGroup.DELETE("/assignments/:assignment_id", employeeHandler.RemoveEmployeeAssignment)

		// Export and statistics
		employeeGroup.GET("/export", employeeHandler.ExportEmployees)
		employeeGroup.GET("/statistics", employeeHandler.GetEmployeeStatistics)
	}
}
