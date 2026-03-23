package handlers

import (
	"fmt"
	"strconv"

	"case/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// OutbreakAssignmentHandler handles outbreak assignment operations
type OutbreakAssignmentHandler struct {
	userOutbreakService *models.UserOutbreakService
	patientRoleService  *models.PatientManagementRoleService
	userService         *models.UserService
	outbreakService     *models.OutbreakService
	facilityService     *models.FacilityService
	store               *session.Store
}

// NewOutbreakAssignmentHandler creates a new outbreak assignment handler
func NewOutbreakAssignmentHandler(
	userOutbreakService *models.UserOutbreakService,
	patientRoleService *models.PatientManagementRoleService,
	userService *models.UserService,
	outbreakService *models.OutbreakService,
	facilityService *models.FacilityService,
	store *session.Store,
) *OutbreakAssignmentHandler {
	return &OutbreakAssignmentHandler{
		userOutbreakService: userOutbreakService,
		patientRoleService:  patientRoleService,
		userService:         userService,
		outbreakService:     outbreakService,
		facilityService:     facilityService,
		store:               store,
	}
}

// ShowOutbreakAssignmentForm shows the outbreak assignment form
func (h *OutbreakAssignmentHandler) ShowOutbreakAssignmentForm(c *fiber.Ctx) error {
	return GenerateHTML(c, nil, NewTemplateData(c, h.store), "assign_outbreak")
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
	sess, err := h.store.Get(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Session error"})
	}

	userIDFromSession := sess.Get("user")
	if userIDFromSession == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized - No user in session"})
	}

	var currentUserID int64
	switch v := userIDFromSession.(type) {
	case int:
		currentUserID = int64(v)
	case int64:
		currentUserID = v
	case float64:
		currentUserID = int64(v)
	default:
		return c.Status(401).JSON(fiber.Map{"error": "Invalid user session"})
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
	// Get current user ID from session
	sess, err := h.store.Get(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Session error"})
	}

	userIDFromSession := sess.Get("user")
	if userIDFromSession == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized - No user in session"})
	}

	var userID int64
	switch v := userIDFromSession.(type) {
	case int:
		userID = int64(v)
	case int64:
		userID = v
	case float64:
		userID = int64(v)
	default:
		return c.Status(401).JSON(fiber.Map{"error": "Invalid user session"})
	}

	outbreaks, err := h.userOutbreakService.GetUserOutbreaks(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"outbreaks": outbreaks})
}

// ShowPatientRoleAssignmentForm shows the patient role assignment form
func (h *OutbreakAssignmentHandler) ShowPatientRoleAssignmentForm(c *fiber.Ctx) error {
	return GenerateHTML(c, nil, NewTemplateData(c, h.store), "assign_patient_role")
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
	sess, err := h.store.Get(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Session error"})
	}

	userIDFromSession := sess.Get("user")
	if userIDFromSession == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized - No user in session"})
	}

	var currentUserID int64
	switch v := userIDFromSession.(type) {
	case int:
		currentUserID = int64(v)
	case int64:
		currentUserID = v
	case float64:
		currentUserID = int64(v)
	default:
		return c.Status(401).JSON(fiber.Map{"error": "Invalid user session"})
	}

	err = h.patientRoleService.AssignPatientRole(req.UserID, req.RoleType, req.OutbreakID, req.FacilityID, currentUserID)
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
	// Check if user is authenticated
	sess, err := h.store.Get(c)
	if err != nil {
		return c.Status(401).SendString("Session error")
	}

	userIDFromSession := sess.Get("user")
	if userIDFromSession == nil {
		return c.Status(401).SendString("Unauthorized - No user in session")
	}

	// Get all outbreak assignments (not just for current user)
	assignments, err := h.userOutbreakService.GetAllOutbreakAssignments()
	if err != nil {
		return c.Status(500).SendString("Failed to load outbreak assignments: " + err.Error())
	}

	// Debug logging
	fmt.Printf("DEBUG: Loaded %d outbreak assignments\n", len(assignments))
	for i, assignment := range assignments {
		fmt.Printf("DEBUG: Assignment %d - User: %s, Outbreak: %s\n", i+1, assignment.User.UserName, assignment.Outbreak.Name.String)
	}

	// Get all outbreaks for the dropdown
	outbreaks, err := h.outbreakService.GetAllOutbreaks()
	if err != nil {
		return c.Status(500).SendString("Failed to load outbreaks: " + err.Error())
	}

	// Get all users for the dropdown
	users, err := h.userService.GetAllEnhancedUsers()
	if err != nil {
		return c.Status(500).SendString("Failed to load users: " + err.Error())
	}

	// Convert assignments to interface slice for template
	assignmentItems := make([]interface{}, len(assignments))
	for i, assignment := range assignments {
		assignmentItems[i] = assignment
	}

	data := NewTemplateData(c, h.store)
	data.Items = assignmentItems
	data.Outbreaks = outbreaks
	data.Users = users

	// Debug logging for template data
	fmt.Printf("DEBUG: Template data - Items count: %d\n", len(data.Items))

	return GenerateHTML(c, nil, data, "outbreak_assignments")
}

// ShowOutbreakAssignmentsAPI returns outbreak assignments and related data as JSON
func (h *OutbreakAssignmentHandler) ShowOutbreakAssignmentsAPI(c *fiber.Ctx) error {
	// Ensure user is authenticated
	sess, err := h.store.Get(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Session error"})
	}
	if sess.Get("user") == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	assignments, err := h.userOutbreakService.GetAllOutbreakAssignments()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load assignments"})
	}

	outbreaks, err := h.outbreakService.GetAllOutbreaks()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load outbreaks"})
	}

	users, err := h.userService.GetAllEnhancedUsers()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load users"})
	}

	return c.JSON(fiber.Map{
		"assignments": assignments,
		"outbreaks":   outbreaks,
		"users":       users,
	})
}

// ShowAssignOutbreakForm shows the assign outbreak form
func (h *OutbreakAssignmentHandler) ShowAssignOutbreakForm(c *fiber.Ctx) error {
	return GenerateHTML(c, nil, NewTemplateData(c, h.store), "assign_outbreak")
}

// ShowPatientRoles shows the patient roles page
func (h *OutbreakAssignmentHandler) ShowPatientRoles(c *fiber.Ctx) error {
	return GenerateHTML(c, nil, NewTemplateData(c, h.store), "patient_roles")
}

// ShowAssignPatientRoleForm shows the assign patient role form
func (h *OutbreakAssignmentHandler) ShowAssignPatientRoleForm(c *fiber.Ctx) error {
	return GenerateHTML(c, nil, NewTemplateData(c, h.store), "assign_patient_role")
}

// ShowAssignForm shows the assign form (legacy method)
func (h *OutbreakAssignmentHandler) ShowAssignForm(c *fiber.Ctx) error {
	return GenerateHTML(c, nil, NewTemplateData(c, h.store), "assign_outbreak")
}

// ShowAssignFormFiber shows the assign form using Fiber
func (h *OutbreakAssignmentHandler) ShowAssignFormFiber(c *fiber.Ctx) error {
	// The template loads outbreaks and users via JavaScript from API endpoints
	// No outbreak ID parameter is required for this form
	return GenerateHTML(c, nil, NewTemplateData(c, h.store), "assign_outbreak")
}

// HandleAssignFormSubmission handles the form submission for assigning users to outbreaks
func (h *OutbreakAssignmentHandler) HandleAssignFormSubmission(c *fiber.Ctx) error {
	// Parse form data
	outbreakID := c.FormValue("outbreak_id")
	userID := c.FormValue("user_id")

	if outbreakID == "" || userID == "" {
		return c.Status(400).SendString("Outbreak ID and User ID are required")
	}

	// Parse IDs
	outbreakIDInt, err := strconv.ParseInt(outbreakID, 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid outbreak ID")
	}

	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid user ID")
	}

	// Get current user ID from session
	sess, err := h.store.Get(c)
	if err != nil {
		return c.Status(401).SendString("Session error")
	}

	userIDFromSession := sess.Get("user")
	if userIDFromSession == nil {
		return c.Status(401).SendString("Unauthorized - No user in session")
	}

	var currentUserID int64
	switch v := userIDFromSession.(type) {
	case int:
		currentUserID = int64(v)
	case int64:
		currentUserID = v
	case float64:
		currentUserID = int64(v)
	default:
		return c.Status(401).SendString("Invalid user session")
	}

	// Assign user to outbreak
	err = h.userOutbreakService.AssignUserToOutbreak(userIDInt, outbreakIDInt, currentUserID)
	if err != nil {
		return c.Status(500).SendString("Failed to assign user to outbreak: " + err.Error())
	}

	// Redirect to assignments page with success message
	return c.Redirect("/outbreaks/assignments?success=User assigned successfully")
}

// HandleAssignFormSubmissionAPI assigns a user to an outbreak from JSON body
func (h *OutbreakAssignmentHandler) HandleAssignFormSubmissionAPI(c *fiber.Ctx) error {
	var req struct {
		OutbreakID int64 `json:"outbreak_id"`
		UserID     int64 `json:"user_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if req.OutbreakID == 0 || req.UserID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "outbreak_id and user_id are required"})
	}

	currentUID := GetCurrentUser(c, h.store)
	if currentUID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	currentUserID := int64(currentUID)

	if err := h.userOutbreakService.AssignUserToOutbreak(req.UserID, req.OutbreakID, currentUserID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "User assigned to outbreak successfully"})
}

// RemoveUserFromOutbreakAPI removes a user from an outbreak using URL params
func (h *OutbreakAssignmentHandler) RemoveUserFromOutbreakAPI(c *fiber.Ctx) error {
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

	if err := h.userOutbreakService.RemoveUserFromOutbreak(uid, oid); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "User removed from outbreak successfully"})
}

// GetOutbreakService returns the outbreak service
func (h *OutbreakAssignmentHandler) GetOutbreakService() *models.OutbreakService {
	return h.outbreakService
}

// GetUserService returns the user service
func (h *OutbreakAssignmentHandler) GetUserService() *models.UserService {
	return h.userService
}
