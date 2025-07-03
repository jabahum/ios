package routes

import (
	"case/internal/handlers"
	"case/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupOutbreakAssignmentRoutes sets up routes for outbreak assignment
func SetupOutbreakAssignmentRoutes(router *gin.Engine, outbreakHandler *handlers.OutbreakAssignmentHandler) {
	// Outbreak assignment routes (require authentication)
	outbreakGroup := router.Group("/outbreaks")
	outbreakGroup.Use(middleware.GinAuthRequired())
	{
		// Outbreak assignment management
		outbreakGroup.GET("/:id/assign", outbreakHandler.ShowOutbreakAssignmentForm)
		outbreakGroup.POST("/:id/assign", outbreakHandler.AssignUserToOutbreak)
		outbreakGroup.DELETE("/:outbreak_id/users/:user_id", outbreakHandler.RemoveUserFromOutbreak)
		outbreakGroup.GET("/assignments", outbreakHandler.ListOutbreakAssignments)

		// Get user's assigned outbreaks (AJAX)
		outbreakGroup.GET("/my-outbreaks", outbreakHandler.GetUserOutbreaks)
	}

	// Patient role management routes
	patientGroup := router.Group("/patient-roles")
	patientGroup.Use(middleware.GinAuthRequired())
	{
		patientGroup.GET("/assign", outbreakHandler.ShowPatientRoleAssignmentForm)
		patientGroup.POST("/assign", outbreakHandler.AssignPatientRole)
		patientGroup.DELETE("/remove", outbreakHandler.RemovePatientRole)
		patientGroup.GET("/user/:user_id", outbreakHandler.GetUserPatientRoles)
		patientGroup.GET("/check-permission", outbreakHandler.CheckPatientPermission)
	}
}
