package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"case/internal/models"
	"case/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
)

// OutbreakAssignmentHandler handles outbreak assignment operations
type OutbreakAssignmentHandler struct {
	userOutbreakService *models.UserOutbreakService
	patientRoleService  *models.PatientManagementRoleService
	userService         *models.UserService
	outbreakService     *models.OutbreakService
	facilityService     *models.FacilityService
}

// NewOutbreakAssignmentHandler creates a new outbreak assignment handler
func NewOutbreakAssignmentHandler(
	userOutbreakService *models.UserOutbreakService,
	patientRoleService *models.PatientManagementRoleService,
	userService *models.UserService,
	outbreakService *models.OutbreakService,
	facilityService *models.FacilityService,
) *OutbreakAssignmentHandler {
	return &OutbreakAssignmentHandler{
		userOutbreakService: userOutbreakService,
		patientRoleService:  patientRoleService,
		userService:         userService,
		outbreakService:     outbreakService,
		facilityService:     facilityService,
	}
}

// ShowOutbreakAssignmentForm shows the outbreak assignment form
func (h *OutbreakAssignmentHandler) ShowOutbreakAssignmentForm(c *gin.Context) {
	c.HTML(http.StatusOK, "assign_outbreak.html", gin.H{})
}

// AssignUserToOutbreak assigns a user to an outbreak
func (h *OutbreakAssignmentHandler) AssignUserToOutbreak(c *gin.Context) {
	outbreakID := c.Param("id")
	id, err := strconv.ParseInt(outbreakID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid outbreak ID"})
		return
	}

	var req struct {
		UserID int64 `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user ID from session
	currentUserID := utils.GetUserIDFromSession(c)
	if currentUserID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	err = h.userOutbreakService.AssignUserToOutbreak(req.UserID, id, currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User assigned to outbreak successfully"})
}

// RemoveUserFromOutbreak removes a user from an outbreak
func (h *OutbreakAssignmentHandler) RemoveUserFromOutbreak(c *gin.Context) {
	outbreakID := c.Param("outbreak_id")
	userID := c.Param("user_id")

	oid, err := strconv.ParseInt(outbreakID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid outbreak ID"})
		return
	}

	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.userOutbreakService.RemoveUserFromOutbreak(uid, oid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User removed from outbreak successfully"})
}

// GetUserOutbreaks returns outbreaks assigned to the current user
func (h *OutbreakAssignmentHandler) GetUserOutbreaks(c *gin.Context) {
	userID := utils.GetUserIDFromSession(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	outbreaks, err := h.userOutbreakService.GetUserOutbreaks(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"outbreaks": outbreaks})
}

// ShowPatientRoleAssignmentForm shows the patient role assignment form
func (h *OutbreakAssignmentHandler) ShowPatientRoleAssignmentForm(c *gin.Context) {
	c.HTML(http.StatusOK, "assign_patient_role.html", gin.H{})
}

// AssignPatientRole assigns a patient management role to a user
func (h *OutbreakAssignmentHandler) AssignPatientRole(c *gin.Context) {
	var req struct {
		UserID     int64  `json:"user_id" binding:"required"`
		RoleType   string `json:"role_type" binding:"required"`
		OutbreakID *int64 `json:"outbreak_id"`
		FacilityID *int64 `json:"facility_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user ID from session
	currentUserID := utils.GetUserIDFromSession(c)
	if currentUserID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	err := h.patientRoleService.AssignPatientRole(req.UserID, req.RoleType, req.OutbreakID, req.FacilityID, currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Patient role assigned successfully"})
}

// RemovePatientRole removes a patient management role from a user
func (h *OutbreakAssignmentHandler) RemovePatientRole(c *gin.Context) {
	var req struct {
		UserID     int64  `json:"user_id" binding:"required"`
		RoleType   string `json:"role_type" binding:"required"`
		OutbreakID *int64 `json:"outbreak_id"`
		FacilityID *int64 `json:"facility_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.patientRoleService.RemovePatientRole(req.UserID, req.RoleType, req.OutbreakID, req.FacilityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Patient role removed successfully"})
}

// GetUserPatientRoles returns patient roles for a user
func (h *OutbreakAssignmentHandler) GetUserPatientRoles(c *gin.Context) {
	userID := c.Param("user_id")
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	roles, err := h.patientRoleService.GetUserPatientRoles(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

// CheckPatientPermission checks if a user has a specific patient permission
func (h *OutbreakAssignmentHandler) CheckPatientPermission(c *gin.Context) {
	userID := c.Query("user_id")
	roleType := c.Query("role_type")

	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var outbreakID, facilityID *int64
	if outbreakIDStr := c.Query("outbreak_id"); outbreakIDStr != "" {
		if id, err := strconv.ParseInt(outbreakIDStr, 10, 64); err == nil {
			outbreakID = &id
		}
	}
	if facilityIDStr := c.Query("facility_id"); facilityIDStr != "" {
		if id, err := strconv.ParseInt(facilityIDStr, 10, 64); err == nil {
			facilityID = &id
		}
	}

	hasPermission := h.patientRoleService.CheckPatientPermission(uid, roleType, outbreakID, facilityID)
	c.JSON(http.StatusOK, gin.H{"has_permission": hasPermission})
}

// ListOutbreakAssignments lists all outbreak assignments
func (h *OutbreakAssignmentHandler) ListOutbreakAssignments(c *gin.Context) {
	c.HTML(http.StatusOK, "outbreak_assignments.html", gin.H{})
}

// ShowOutbreakAssignments shows the outbreak assignments page
func (h *OutbreakAssignmentHandler) ShowOutbreakAssignments(c *gin.Context) {
	c.HTML(http.StatusOK, "outbreak_assignments.html", gin.H{})
}

// ShowAssignOutbreakForm shows the form to assign users to outbreaks
func (h *OutbreakAssignmentHandler) ShowAssignOutbreakForm(c *gin.Context) {
	c.HTML(http.StatusOK, "assign_outbreak.html", gin.H{})
}

// ShowPatientRoles shows the patient roles page
func (h *OutbreakAssignmentHandler) ShowPatientRoles(c *gin.Context) {
	c.HTML(http.StatusOK, "patient_roles.html", gin.H{})
}

// ShowAssignPatientRoleForm shows the form to assign patient roles
func (h *OutbreakAssignmentHandler) ShowAssignPatientRoleForm(c *gin.Context) {
	c.HTML(http.StatusOK, "assign_patient_role.html", gin.H{})
}

// ShowAssignForm shows the form to assign users to outbreaks
func (h *OutbreakAssignmentHandler) ShowAssignForm(c *gin.Context) {
	// Get outbreaks
	outbreaks, err := h.outbreakService.GetAllOutbreaks()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error", gin.H{
			"error": "Failed to load outbreaks: " + err.Error(),
		})
		return
	}

	// Get users
	users, err := h.userService.GetAllUsers()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error", gin.H{
			"error": "Failed to load users: " + err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "assign_outbreak", gin.H{
		"Outbreaks": outbreaks,
		"Users":     users,
	})
}

// ShowAssignFormFiber shows the form to assign users to outbreaks (Fiber version)
func (h *OutbreakAssignmentHandler) ShowAssignFormFiber(c *fiber.Ctx) error {
	// Get outbreaks
	outbreaks, err := h.outbreakService.GetAllOutbreaks()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Failed to load outbreaks: %v", err))
	}

	// Get users
	users, err := h.userService.GetAllUsers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Failed to load users: %v", err))
	}

	// Create template data
	data := &TemplateData{
		Outbreaks: outbreaks,
		Users:     users,
	}

	// Use GenerateHTML instead of c.Render
	return GenerateHTML(c, nil, data, "assign_outbreak")
}

// GetOutbreakService returns the outbreak service
func (h *OutbreakAssignmentHandler) GetOutbreakService() *models.OutbreakService {
	return h.outbreakService
}

// GetUserService returns the user service
func (h *OutbreakAssignmentHandler) GetUserService() *models.UserService {
	return h.userService
}
