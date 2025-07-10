package handlers

import (
	"fmt"
	"strconv"

	"case/internal/models"
	"case/internal/utils"

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
func (h *OutbreakAssignmentHandler) ShowOutbreakAssignmentForm(c *fiber.Ctx) error {
	return GenerateHTML(c, nil, nil, "assign_outbreak")
}

// AssignUserToOutbreak assigns a user to an outbreak
func (h *OutbreakAssignmentHandler) AssignUserToOutbreak(c *fiber.Ctx) error {
	outbreakID := c.Params("id")
	id, err := strconv.ParseInt(outbreakID, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid outbreak ID"})
	}

	var req struct {
		UserID int64 `json:"user_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Get current user ID from session
	currentUserID := utils.GetUserIDFromSession(c)
	if currentUserID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	err = h.userOutbreakService.AssignUserToOutbreak(req.UserID, id, currentUserID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "User assigned to outbreak successfully"})
}

// RemoveUserFromOutbreak removes a user from an outbreak
func (h *OutbreakAssignmentHandler) RemoveUserFromOutbreak(c *fiber.Ctx) error {
	outbreakID := c.Params("outbreak_id")
	userID := c.Params("user_id")

	oid, err := strconv.ParseInt(outbreakID, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid outbreak ID"})
	}

	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	err = h.userOutbreakService.RemoveUserFromOutbreak(uid, oid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "User removed from outbreak successfully"})
}

// GetUserOutbreaks returns outbreaks assigned to the current user
func (h *OutbreakAssignmentHandler) GetUserOutbreaks(c *fiber.Ctx) error {
	userID := utils.GetUserIDFromSession(c)
	if userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	outbreaks, err := h.userOutbreakService.GetUserOutbreaks(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"outbreaks": outbreaks})
}

// ShowPatientRoleAssignmentForm shows the patient role assignment form
func (h *OutbreakAssignmentHandler) ShowPatientRoleAssignmentForm(c *fiber.Ctx) error {
	return GenerateHTML(c, nil, nil, "assign_patient_role")
}

// AssignPatientRole assigns a patient management role to a user
func (h *OutbreakAssignmentHandler) AssignPatientRole(c *fiber.Ctx) error {
	var req struct {
		UserID     int64  `json:"user_id"`
		RoleType   string `json:"role_type"`
		OutbreakID *int64 `json:"outbreak_id"`
		FacilityID *int64 `json:"facility_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Get current user ID from session
	currentUserID := utils.GetUserIDFromSession(c)
	if currentUserID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	err := h.patientRoleService.AssignPatientRole(req.UserID, req.RoleType, req.OutbreakID, req.FacilityID, currentUserID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Patient role assigned successfully"})
}

// RemovePatientRole removes a patient management role from a user
func (h *OutbreakAssignmentHandler) RemovePatientRole(c *fiber.Ctx) error {
	var req struct {
		UserID     int64  `json:"user_id"`
		RoleType   string `json:"role_type"`
		OutbreakID *int64 `json:"outbreak_id"`
		FacilityID *int64 `json:"facility_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	err := h.patientRoleService.RemovePatientRole(req.UserID, req.RoleType, req.OutbreakID, req.FacilityID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Patient role removed successfully"})
}

// GetUserPatientRoles returns patient roles for a user
func (h *OutbreakAssignmentHandler) GetUserPatientRoles(c *fiber.Ctx) error {
	userID := c.Params("user_id")
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	roles, err := h.patientRoleService.GetUserPatientRoles(uid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"roles": roles})
}

// CheckPatientPermission checks if a user has a specific patient permission
func (h *OutbreakAssignmentHandler) CheckPatientPermission(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	roleType := c.Query("role_type")

	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	hasPermission := h.patientRoleService.CheckPatientPermission(uid, roleType, nil, nil)
	return c.JSON(fiber.Map{"has_permission": hasPermission})
}

// ListOutbreakAssignments lists all outbreak assignments
func (h *OutbreakAssignmentHandler) ListOutbreakAssignments(c *fiber.Ctx) error {
	// For now, return empty list since the method doesn't exist
	return c.JSON(fiber.Map{"assignments": []interface{}{}})
}

// ShowOutbreakAssignments shows the outbreak assignments page
func (h *OutbreakAssignmentHandler) ShowOutbreakAssignments(c *fiber.Ctx) error {
	return GenerateHTML(c, nil, nil, "outbreak_assignments")
}

// ShowAssignOutbreakForm shows the assign outbreak form
func (h *OutbreakAssignmentHandler) ShowAssignOutbreakForm(c *fiber.Ctx) error {
	return GenerateHTML(c, nil, nil, "assign_outbreak")
}

// ShowPatientRoles shows the patient roles page
func (h *OutbreakAssignmentHandler) ShowPatientRoles(c *fiber.Ctx) error {
	return GenerateHTML(c, nil, nil, "patient_roles")
}

// ShowAssignPatientRoleForm shows the assign patient role form
func (h *OutbreakAssignmentHandler) ShowAssignPatientRoleForm(c *fiber.Ctx) error {
	return GenerateHTML(c, nil, nil, "assign_patient_role")
}

// ShowAssignForm shows the assign form (legacy method)
func (h *OutbreakAssignmentHandler) ShowAssignForm(c *fiber.Ctx) error {
	return GenerateHTML(c, nil, nil, "assign_outbreak")
}

// ShowAssignFormFiber shows the assign form using Fiber
func (h *OutbreakAssignmentHandler) ShowAssignFormFiber(c *fiber.Ctx) error {
	// Get outbreak ID from URL parameter
	outbreakID := c.Params("i")
	if outbreakID == "" {
		return c.Status(400).SendString("Outbreak ID is required")
	}

	// Parse outbreak ID
	id, err := strconv.Atoi(outbreakID)
	if err != nil {
		return c.Status(400).SendString("Invalid outbreak ID")
	}

	// Get outbreak details
	outbreak, err := h.outbreakService.GetOutbreakByID(int64(id))
	if err != nil {
		return c.Status(500).SendString(fmt.Sprintf("Error getting outbreak: %v", err))
	}

	// Get all users for assignment
	users, err := h.userService.GetAllUsers()
	if err != nil {
		return c.Status(500).SendString(fmt.Sprintf("Error getting users: %v", err))
	}

	// Prepare data for template
	data := fiber.Map{
		"Outbreak": outbreak,
		"Users":    users,
	}

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
