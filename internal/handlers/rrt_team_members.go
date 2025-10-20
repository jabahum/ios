package handlers

import (
	"case/internal/models"
	"database/sql"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/lib/pq"
)

// RRTTeamMembersHandler handles all RRT team member operations
type RRTTeamMembersHandler struct {
	db    *sql.DB
	store *session.Store
}

// NewRRTTeamMembersHandler creates a new team members handler
func NewRRTTeamMembersHandler(db *sql.DB, store *session.Store) *RRTTeamMembersHandler {
	return &RRTTeamMembersHandler{db: db, store: store}
}

// HandlerRRTTeamMembersList displays all RRT team members
func (h *RRTTeamMembersHandler) HandlerRRTTeamMembersList(c *fiber.Ctx) error {
	data := NewTemplateDataWithDB(c, h.store, h.db)

	members, err := h.getAllRRTTeamMembers()
	if err != nil {
		log.Printf("Error getting RRT team members: %v", err)
		members = []*models.RRTTeamMember{}
	}

	data.RRTTeamMembers = members
	return GenerateHTML(c, h.db, data, "rrt_team_members_list")
}

// HandlerRRTTeamMemberForm displays the form for adding/editing team members
func (h *RRTTeamMembersHandler) HandlerRRTTeamMemberForm(c *fiber.Ctx) error {
	data := NewTemplateDataWithDB(c, h.store, h.db)

	// If editing, get the member
	memberID := c.Params("id")
	if memberID != "" {
		member, err := h.getRRTTeamMemberByID(memberID)
		if err != nil {
			log.Printf("Error getting RRT team member: %v", err)
		} else {
			data.RRTTeamMember = member
		}
	}

	// Get pillars for dropdown
	pillars, err := h.getAllPillars()
	if err != nil {
		log.Printf("Error getting pillars: %v", err)
		pillars = []*models.Pillar{}
	}
	data.Pillars = pillars

	return GenerateHTML(c, h.db, data, "rrt_team_member_form")
}

// HandlerRRTTeamMemberSave handles saving team member data
func (h *RRTTeamMembersHandler) HandlerRRTTeamMemberSave(c *fiber.Ctx) error {
	memberID := c.FormValue("id")

	// Parse specializations from JSON string
	specializations := []string{}
	if specStr := c.FormValue("specializations"); specStr != "" {
		// Handle comma-separated values or JSON array
		if specStr[0] == '[' {
			// JSON array format
			json.Unmarshal([]byte(specStr), &specializations)
		} else {
			// Comma-separated format
			specializations = strings.Split(specStr, ",")
			for i, s := range specializations {
				specializations[i] = strings.TrimSpace(s)
			}
		}
	}

	// Parse experience years
	experienceYears, _ := strconv.Atoi(c.FormValue("experience_years"))

	member := &models.RRTTeamMember{
		FirstName:           c.FormValue("first_name"),
		LastName:            c.FormValue("last_name"),
		Organization:        parseRRTNullString(c.FormValue("organization")),
		Position:            parseRRTNullString(c.FormValue("position")),
		Phone:               parseRRTNullString(c.FormValue("phone")),
		Email:               parseRRTNullString(c.FormValue("email")),
		NationalID:          parseRRTNullString(c.FormValue("national_id")),
		EmployeeID:          parseRRTNullString(c.FormValue("employee_id")),
		Qualifications:      parseRRTNullString(c.FormValue("qualifications")),
		Specializations:     specializations,
		Certifications:      parseRRTNullString(c.FormValue("certifications")),
		ExperienceYears:     experienceYears,
		IsDriver:            c.FormValue("is_driver") == "true",
		DriverLicense:       parseRRTNullString(c.FormValue("driver_license")),
		DriverLicenseExpiry: parseRRTNullDate(c.FormValue("driver_license_expiry")),
		IsActive:            c.FormValue("is_active") == "true",
		PillarID:            parseRRTNullInt64(c.FormValue("pillar_id")),
	}

	var err error
	if memberID != "" {
		member.ID, _ = strconv.ParseInt(memberID, 10, 64)
		err = h.updateRRTTeamMember(member)
	} else {
		err = h.createRRTTeamMember(member)
	}

	if err != nil {
		log.Printf("Error saving RRT team member: %v", err)
		return c.Status(500).SendString("Error saving RRT team member")
	}

	return c.Redirect("/resource-management/rrt-team-members")
}

// HandlerRRTTeamMemberAssignmentsList displays team member assignments
func (h *RRTTeamMembersHandler) HandlerRRTTeamMemberAssignmentsList(c *fiber.Ctx) error {
	data := NewTemplateDataWithDB(c, h.store, h.db)

	// Get team filter
	teamID := c.Query("team_id")
	var assignments []*models.RRTTeamMemberAssignment
	var err error

	if teamID != "" {
		assignments, err = h.getAssignmentsByTeam(teamID)
	} else {
		assignments, err = h.getAllRRTTeamMemberAssignments()
	}

	if err != nil {
		log.Printf("Error getting team member assignments: %v", err)
		assignments = []*models.RRTTeamMemberAssignment{}
	}

	data.RRTTeamMemberAssignments = assignments
	return GenerateHTML(c, h.db, data, "rrt_team_member_assignments_list")
}

// HandlerRRTTeamMemberAssignmentForm displays the form for adding/editing assignments
func (h *RRTTeamMembersHandler) HandlerRRTTeamMemberAssignmentForm(c *fiber.Ctx) error {
	data := NewTemplateDataWithDB(c, h.store, h.db)

	// Get teams for dropdown
	teams, err := h.getAllRRTTeams()
	if err != nil {
		log.Printf("Error getting RRT teams: %v", err)
		teams = []*models.RRTTeam{}
	}

	// Get members for dropdown
	members, err := h.getAllRRTTeamMembers()
	if err != nil {
		log.Printf("Error getting RRT team members: %v", err)
		members = []*models.RRTTeamMember{}
	}

	data.RRTTeams = teams
	data.RRTTeamMembers = members

	// If editing, get the assignment
	assignmentID := c.Params("id")
	if assignmentID != "" {
		assignment, err := h.getRRTTeamMemberAssignmentByID(assignmentID)
		if err != nil {
			log.Printf("Error getting RRT team member assignment: %v", err)
		} else {
			data.RRTTeamMemberAssignment = assignment
		}
	}

	return GenerateHTML(c, h.db, data, "rrt_team_member_assignment_form")
}

// HandlerRRTTeamMemberAssignmentSave handles saving team member assignment data
func (h *RRTTeamMembersHandler) HandlerRRTTeamMemberAssignmentSave(c *fiber.Ctx) error {
	assignmentID := c.FormValue("id")

	// Convert string IDs to int64
	teamID, _ := strconv.ParseInt(c.FormValue("team_id"), 10, 64)
	memberID, _ := strconv.ParseInt(c.FormValue("member_id"), 10, 64)

	assignment := &models.RRTTeamMemberAssignment{
		TeamID:          teamID,
		MemberID:        memberID,
		Role:            c.FormValue("role"),
		StartDate:       parseRRTDate(c.FormValue("start_date")),
		EndDate:         parseRRTNullDate(c.FormValue("end_date")),
		IsActive:        c.FormValue("is_active") == "true",
		AssignmentNotes: parseRRTNullString(c.FormValue("assignment_notes")),
	}

	var err error
	if assignmentID != "" {
		assignment.ID, _ = strconv.ParseInt(assignmentID, 10, 64)
		err = h.updateRRTTeamMemberAssignment(assignment)
	} else {
		err = h.createRRTTeamMemberAssignment(assignment)
	}

	if err != nil {
		log.Printf("Error saving RRT team member assignment: %v", err)
		return c.Status(500).SendString("Error saving RRT team member assignment")
	}

	return c.Redirect("/resource-management/rrt-team-member-assignments")
}

// RRT Deployment Proposal Handlers
func (h *RRTTeamMembersHandler) HandlerRRTDeploymentProposalsList(c *fiber.Ctx) error {
	data := NewTemplateDataWithDB(c, h.store, h.db)
	proposals, err := h.getAllRRTDeploymentProposals()
	if err != nil {
		log.Printf("Error getting RRT deployment proposals: %v", err)
		proposals = []*models.RRTDeploymentProposal{}
	}
	data.RRTDeploymentProposals = proposals
	return GenerateHTML(c, h.db, data, "rrt_deployment_proposals_list")
}

func (h *RRTTeamMembersHandler) HandlerRRTDeploymentProposalView(c *fiber.Ctx) error {
	data := NewTemplateDataWithDB(c, h.store, h.db)

	proposalID := c.Params("id")
	if proposalID == "" {
		return c.Status(400).SendString("Proposal ID is required")
	}

	// Get proposal by ID
	proposal, err := h.getRRTDeploymentProposalByID(proposalID)
	if err != nil {
		log.Printf("Error getting RRT deployment proposal: %v", err)
		return c.Status(404).SendString("Proposal not found")
	}

	data.RRTDeploymentProposal = proposal

	return GenerateHTML(c, h.db, data, "rrt_deployment_proposal_view")
}

func (h *RRTTeamMembersHandler) HandlerRRTDeploymentProposalForm(c *fiber.Ctx) error {
	data := NewTemplateDataWithDB(c, h.store, h.db)

	proposalID := c.Params("id")
	if proposalID != "" {
		// Get proposal by ID (placeholder)
		data.RRTDeploymentProposal = &models.RRTDeploymentProposal{}
	}

	// Get teams, outbreaks, and team members for dropdowns
	teams, err := h.getAllRRTTeams()
	if err != nil {
		log.Printf("Error getting RRT teams: %v", err)
		teams = []*models.RRTTeam{}
	}

	outbreaks, err := h.getAllOutbreaks()
	if err != nil {
		log.Printf("Error getting outbreaks: %v", err)
		outbreaks = []*models.Outbreak{}
	}

	members, err := h.getAllRRTTeamMembers()
	if err != nil {
		log.Printf("Error getting RRT team members: %v", err)
		members = []*models.RRTTeamMember{}
	}

	data.RRTTeams = teams
	data.Outbreaks = outbreaks
	data.RRTTeamMembers = members

	return GenerateHTML(c, h.db, data, "rrt_deployment_proposal_form")
}

func (h *RRTTeamMembersHandler) HandlerRRTDeploymentProposalSave(c *fiber.Ctx) error {
	// Get current user ID from session
	userID := GetCurrentUser(c, h.store)
	if userID == 0 {
		return c.Redirect("/login")
	}

	// Parse required skills
	requiredSkills := []string{}
	if skillsStr := c.FormValue("required_skills"); skillsStr != "" {
		if skillsStr[0] == '[' {
			json.Unmarshal([]byte(skillsStr), &requiredSkills)
		} else {
			requiredSkills = strings.Split(skillsStr, ",")
			for i, s := range requiredSkills {
				requiredSkills[i] = strings.TrimSpace(s)
			}
		}
	}

	// Parse IDs
	outbreakID, _ := strconv.ParseInt(c.FormValue("outbreak_id"), 10, 64)

	proposal := &models.RRTDeploymentProposal{
		OutbreakID:              outbreakID,
		ProposedBy:              int64(userID),
		ProposedDate:            time.Now(),
		DeploymentPurpose:       c.FormValue("deployment_purpose"),
		ProposedTeamComposition: parseRRTNullString(c.FormValue("proposed_team_composition")),
		RequiredSkills:          requiredSkills,
		DeploymentDurationDays:  parseRRTNullInt64(c.FormValue("deployment_duration_days")),
		ExpectedStartDate:       parseRRTNullDate(c.FormValue("expected_start_date")),
		ExpectedEndDate:         parseRRTNullDate(c.FormValue("expected_end_date")),
		SpecialRequirements:     parseRRTNullString(c.FormValue("special_requirements")),
		Justification:           c.FormValue("justification"),
		Status:                  "pending",
	}

	err := h.createRRTDeploymentProposal(proposal)
	if err != nil {
		log.Printf("Error saving RRT deployment proposal: %v", err)
		return c.Status(500).SendString("Error saving deployment proposal: " + err.Error())
	}

	return c.Redirect("/resource-management/deployment-proposals")
}

func (h *RRTTeamMembersHandler) HandlerRRTDeploymentProposalApprove(c *fiber.Ctx) error {
	proposalID := c.Params("id")
	reviewNotes := c.FormValue("review_notes")

	err := h.reviewRRTDeploymentProposal(proposalID, "approved", reviewNotes, "")
	if err != nil {
		log.Printf("Error approving RRT deployment proposal: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to approve proposal"})
	}

	return c.Redirect("/resource-management/deployment-proposals")
}

func (h *RRTTeamMembersHandler) HandlerRRTDeploymentProposalReject(c *fiber.Ctx) error {
	proposalID := c.Params("id")
	rejectionReason := c.FormValue("rejection_reason")

	err := h.reviewRRTDeploymentProposal(proposalID, "rejected", "", rejectionReason)
	if err != nil {
		log.Printf("Error rejecting RRT deployment proposal: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to reject proposal"})
	}

	return c.Redirect("/resource-management/deployment-proposals")
}

// RRT Deployment Extension Handlers
func (h *RRTTeamMembersHandler) HandlerRRTDeploymentExtensionsList(c *fiber.Ctx) error {
	data := NewTemplateDataWithDB(c, h.store, h.db)
	extensions, err := h.getAllRRTDeploymentExtensions()
	if err != nil {
		log.Printf("Error getting RRT deployment extensions: %v", err)
		extensions = []*models.RRTDeploymentExtension{}
	}
	data.RRTDeploymentExtensions = extensions
	return GenerateHTML(c, h.db, data, "rrt_deployment_extensions_list")
}

func (h *RRTTeamMembersHandler) HandlerRRTDeploymentExtensionForm(c *fiber.Ctx) error {
	data := NewTemplateDataWithDB(c, h.store, h.db)

	extensionID := c.Params("id")
	if extensionID != "" {
		// Get extension by ID (placeholder)
		data.RRTDeploymentExtension = &models.RRTDeploymentExtension{}
	}

	// Get deployments for dropdown
	deployments, _ := h.getAllRRTDeployments()
	data.RRTDeployments = deployments

	return GenerateHTML(c, h.db, data, "rrt_deployment_extension_form")
}

func (h *RRTTeamMembersHandler) HandlerRRTDeploymentExtensionSave(c *fiber.Ctx) error {
	// Parse IDs
	deploymentID, _ := strconv.ParseInt(c.FormValue("deployment_id"), 10, 64)
	requestedBy, _ := strconv.ParseInt(c.FormValue("requested_by"), 10, 64)

	extension := &models.RRTDeploymentExtension{
		DeploymentID:    deploymentID,
		ExtensionReason: c.FormValue("extension_reason"),
		OriginalEndDate: parseRRTDate(c.FormValue("original_end_date")),
		NewEndDate:      parseRRTDate(c.FormValue("new_end_date")),
		RequestedBy:     requestedBy,
		RequestedDate:   time.Now(),
		Status:          "pending",
		ApprovalNotes:   parseRRTNullString(c.FormValue("approval_notes")),
	}

	err := h.createRRTDeploymentExtension(extension)
	if err != nil {
		log.Printf("Error saving RRT deployment extension: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save extension"})
	}

	return c.Redirect("/resource-management/deployment-extensions")
}

func (h *RRTTeamMembersHandler) HandlerRRTDeploymentExtensionApprove(c *fiber.Ctx) error {
	extensionID := c.Params("id")
	approvalNotes := c.FormValue("approval_notes")

	err := h.reviewRRTDeploymentExtension(extensionID, "approved", approvalNotes, "")
	if err != nil {
		log.Printf("Error approving RRT deployment extension: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to approve extension"})
	}

	return c.Redirect("/resource-management/deployment-extensions")
}

func (h *RRTTeamMembersHandler) HandlerRRTDeploymentExtensionReject(c *fiber.Ctx) error {
	extensionID := c.Params("id")
	rejectionReason := c.FormValue("rejection_reason")

	err := h.reviewRRTDeploymentExtension(extensionID, "rejected", "", rejectionReason)
	if err != nil {
		log.Printf("Error rejecting RRT deployment extension: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to reject extension"})
	}

	return c.Redirect("/resource-management/deployment-extensions")
}

// RRT Field Role Assignment Handlers
func (h *RRTTeamMembersHandler) HandlerRRTFieldRoleAssignmentsList(c *fiber.Ctx) error {
	data := NewTemplateDataWithDB(c, h.store, h.db)
	assignments, err := h.getAllRRTFieldRoleAssignments()
	if err != nil {
		log.Printf("Error getting RRT field role assignments: %v", err)
		assignments = []*models.RRTFieldRoleAssignment{}
	}
	data.RRTFieldRoleAssignments = assignments
	return GenerateHTML(c, h.db, data, "rrt_field_role_assignments_list")
}

func (h *RRTTeamMembersHandler) HandlerRRTFieldRoleAssignmentForm(c *fiber.Ctx) error {
	data := NewTemplateDataWithDB(c, h.store, h.db)

	assignmentID := c.Params("id")
	if assignmentID != "" {
		// Get assignment by ID (placeholder)
		data.RRTFieldRoleAssignment = &models.RRTFieldRoleAssignment{}
	}

	// Get deployments and members for dropdowns
	deployments, _ := h.getAllRRTDeployments()
	members, _ := h.getAllRRTTeamMembers()
	data.RRTDeployments = deployments
	data.RRTTeamMembers = members

	return GenerateHTML(c, h.db, data, "rrt_field_role_assignment_form")
}

func (h *RRTTeamMembersHandler) HandlerRRTFieldRoleAssignmentSave(c *fiber.Ctx) error {
	// Parse IDs
	deploymentID, _ := strconv.ParseInt(c.FormValue("deployment_id"), 10, 64)
	memberID, _ := strconv.ParseInt(c.FormValue("member_id"), 10, 64)
	assignedBy, _ := strconv.ParseInt(c.FormValue("assigned_by"), 10, 64)

	assignment := &models.RRTFieldRoleAssignment{
		DeploymentID:     deploymentID,
		MemberID:         memberID,
		AdditionalRole:   c.FormValue("additional_role"),
		AssignmentDate:   time.Now(),
		EndDate:          parseRRTNullDate(c.FormValue("end_date")),
		AssignedBy:       assignedBy,
		AssignmentReason: parseRRTNullString(c.FormValue("assignment_reason")),
		IsActive:         c.FormValue("is_active") == "true",
	}

	err := h.createRRTFieldRoleAssignment(assignment)
	if err != nil {
		log.Printf("Error saving RRT field role assignment: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save assignment"})
	}

	return c.Redirect("/resource-management/field-role-assignments")
}

// Helper methods for database operations
func (h *RRTTeamMembersHandler) getAllRRTTeamMembers() ([]*models.RRTTeamMember, error) {
	query := `
		SELECT rtm.id, rtm.first_name, rtm.last_name, rtm.email, rtm.phone, 
		       rtm.national_id, rtm.employee_id, rtm.organization, rtm.position,
		       rtm.qualifications, rtm.specializations, rtm.certifications,
		       rtm.experience_years, rtm.is_driver, rtm.driver_license, 
		       rtm.driver_license_expiry, rtm.is_active, rtm.created_at, 
		       rtm.updated_at, rtm.created_by, rtm.pillar_id,
		       p.name as pillar_name
		FROM rrt_team_members rtm
		LEFT JOIN pillars p ON rtm.pillar_id = p.id
		ORDER BY rtm.created_at DESC
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*models.RRTTeamMember
	for rows.Next() {
		member := &models.RRTTeamMember{}
		var pillarName sql.NullString

		var specializations pq.StringArray
		err := rows.Scan(
			&member.ID, &member.FirstName, &member.LastName, &member.Email, &member.Phone,
			&member.NationalID, &member.EmployeeID, &member.Organization, &member.Position,
			&member.Qualifications, &specializations, &member.Certifications,
			&member.ExperienceYears, &member.IsDriver, &member.DriverLicense,
			&member.DriverLicenseExpiry, &member.IsActive, &member.CreatedAt,
			&member.UpdatedAt, &member.CreatedBy, &member.PillarID,
			&pillarName,
		)
		if err != nil {
			return nil, err
		}

		// Convert pq.StringArray to []string
		member.Specializations = []string(specializations)
		if err != nil {
			return nil, err
		}

		// Set pillar name if available
		if pillarName.Valid {
			member.Pillar = &models.Pillar{Name: pillarName.String}
		}

		members = append(members, member)
	}

	return members, nil
}

func (h *RRTTeamMembersHandler) getRRTTeamMemberByID(id string) (*models.RRTTeamMember, error) {
	query := `
		SELECT rtm.id, rtm.first_name, rtm.last_name, rtm.email, rtm.phone, 
		       rtm.national_id, rtm.employee_id, rtm.organization, rtm.position,
		       rtm.qualifications, rtm.specializations, rtm.certifications,
		       rtm.experience_years, rtm.is_driver, rtm.driver_license, 
		       rtm.driver_license_expiry, rtm.is_active, rtm.created_at, 
		       rtm.updated_at, rtm.created_by, rtm.pillar_id,
		       p.name as pillar_name
		FROM rrt_team_members rtm
		LEFT JOIN pillars p ON rtm.pillar_id = p.id
		WHERE rtm.id = $1
	`

	member := &models.RRTTeamMember{}
	var pillarName sql.NullString

	err := h.db.QueryRow(query, id).Scan(
		&member.ID, &member.FirstName, &member.LastName, &member.Email, &member.Phone,
		&member.NationalID, &member.EmployeeID, &member.Organization, &member.Position,
		&member.Qualifications, &member.Specializations, &member.Certifications,
		&member.ExperienceYears, &member.IsDriver, &member.DriverLicense,
		&member.DriverLicenseExpiry, &member.IsActive, &member.CreatedAt,
		&member.UpdatedAt, &member.CreatedBy, &member.PillarID,
		&pillarName,
	)
	if err != nil {
		return nil, err
	}

	// Set pillar name if available
	if pillarName.Valid {
		member.Pillar = &models.Pillar{Name: pillarName.String}
	}

	return member, nil
}

func (h *RRTTeamMembersHandler) createRRTTeamMember(member *models.RRTTeamMember) error {
	query := `
		INSERT INTO rrt_team_members (
			first_name, last_name, email, phone, national_id, employee_id,
			organization, position, qualifications, specializations, certifications,
			experience_years, is_driver, driver_license, driver_license_expiry,
			is_active, created_by, pillar_id
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		) RETURNING id
	`

	var id int64
	err := h.db.QueryRow(query,
		member.FirstName, member.LastName, member.Email, member.Phone,
		member.NationalID, member.EmployeeID, member.Organization, member.Position,
		member.Qualifications, pq.Array(member.Specializations), member.Certifications,
		member.ExperienceYears, member.IsDriver, member.DriverLicense,
		member.DriverLicenseExpiry, member.IsActive, member.CreatedBy, member.PillarID,
	).Scan(&id)

	if err != nil {
		return err
	}

	member.ID = id
	return nil
}

func (h *RRTTeamMembersHandler) updateRRTTeamMember(member *models.RRTTeamMember) error {
	query := `
		UPDATE rrt_team_members SET
			first_name = $1, last_name = $2, email = $3, phone = $4, national_id = $5,
			employee_id = $6, organization = $7, position = $8, qualifications = $9,
			specializations = $10, certifications = $11, experience_years = $12,
			is_driver = $13, driver_license = $14, driver_license_expiry = $15,
			is_active = $16, updated_at = CURRENT_TIMESTAMP,
			pillar_id = $17
		WHERE id = $18
	`

	_, err := h.db.Exec(query,
		member.FirstName, member.LastName, member.Email, member.Phone,
		member.NationalID, member.EmployeeID, member.Organization, member.Position,
		member.Qualifications, pq.Array(member.Specializations), member.Certifications,
		member.ExperienceYears, member.IsDriver, member.DriverLicense,
		member.DriverLicenseExpiry, member.IsActive, member.PillarID,
		member.ID,
	)

	return err
}

func (h *RRTTeamMembersHandler) getAllRRTTeams() ([]*models.RRTTeam, error) {
	query := `
		SELECT id, team_name, team_code, team_type, team_lead_name, team_lead_phone, 
		       team_lead_email, team_size, specializations, base_location, is_active, 
		       created_at, updated_at, created_by 
		FROM rrt_teams 
		WHERE is_active = true
		ORDER BY team_name
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*models.RRTTeam
	for rows.Next() {
		var team models.RRTTeam
		var specializations pq.StringArray
		err := rows.Scan(
			&team.ID, &team.TeamName, &team.TeamCode, &team.TeamType,
			&team.TeamLeadName, &team.TeamLeadPhone, &team.TeamLeadEmail,
			&team.TeamSize, &specializations, &team.BaseLocation,
			&team.IsActive, &team.CreatedAt, &team.UpdatedAt, &team.CreatedBy,
		)
		if err != nil {
			return nil, err
		}

		// Convert pq.StringArray to []string
		team.Specializations = []string(specializations)
		teams = append(teams, &team)
	}

	return teams, nil
}

func (h *RRTTeamMembersHandler) getAllOutbreaks() ([]*models.Outbreak, error) {
	query := `
		SELECT id, name, description, start_date, end_date, status, 
		       outbreak_type, outbreak_category, enter_on, enter_by, edit_on, edit_by
		FROM outbreaks 
		ORDER BY name
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var outbreaks []*models.Outbreak
	for rows.Next() {
		outbreak := &models.Outbreak{}
		err := rows.Scan(
			&outbreak.ID, &outbreak.Name, &outbreak.Description, &outbreak.StartDate,
			&outbreak.EndDate, &outbreak.Status, &outbreak.OutbreakType, &outbreak.OutbreakCategory,
			&outbreak.EnterOn, &outbreak.EnterBy, &outbreak.EditOn, &outbreak.EditBy,
		)
		if err != nil {
			return nil, err
		}
		outbreaks = append(outbreaks, outbreak)
	}

	return outbreaks, nil
}

func (h *RRTTeamMembersHandler) getAllRRTTeamMemberAssignments() ([]*models.RRTTeamMemberAssignment, error) {
	return []*models.RRTTeamMemberAssignment{}, nil
}

func (h *RRTTeamMembersHandler) getAssignmentsByTeam(teamID string) ([]*models.RRTTeamMemberAssignment, error) {
	return []*models.RRTTeamMemberAssignment{}, nil
}

func (h *RRTTeamMembersHandler) getRRTTeamMemberAssignmentByID(id string) (*models.RRTTeamMemberAssignment, error) {
	return &models.RRTTeamMemberAssignment{}, nil
}

func (h *RRTTeamMembersHandler) createRRTTeamMemberAssignment(assignment *models.RRTTeamMemberAssignment) error {
	return nil
}

func (h *RRTTeamMembersHandler) updateRRTTeamMemberAssignment(assignment *models.RRTTeamMemberAssignment) error {
	return nil
}

func (h *RRTTeamMembersHandler) getRRTDeploymentProposalByID(id string) (*models.RRTDeploymentProposal, error) {
	query := `
		SELECT p.id, p.proposal_number, p.outbreak_id, p.proposed_by, p.proposed_date,
		       p.deployment_purpose, p.proposed_team_composition, p.required_skills,
		       p.deployment_duration_days, p.expected_start_date, p.expected_end_date,
		       p.special_requirements, p.justification, p.status, p.reviewed_by,
		       p.reviewed_at, p.review_notes, p.rejection_reason, p.created_at, p.updated_at,
		       o.name as outbreak_name, o.description as outbreak_description,
		       u.user_name as proposed_by_name,
		       r.user_name as reviewed_by_name
		FROM rrt_deployment_proposals p
		LEFT JOIN outbreaks o ON p.outbreak_id = o.id
		LEFT JOIN users u ON p.proposed_by = u.user_id
		LEFT JOIN users r ON p.reviewed_by = r.user_id
		WHERE p.id = $1
	`

	proposal := &models.RRTDeploymentProposal{}
	var outbreakName, outbreakDesc, proposedByName, reviewedByName sql.NullString
	var requiredSkills pq.StringArray

	err := h.db.QueryRow(query, id).Scan(
		&proposal.ID, &proposal.ProposalNumber, &proposal.OutbreakID, &proposal.ProposedBy,
		&proposal.ProposedDate, &proposal.DeploymentPurpose, &proposal.ProposedTeamComposition,
		&requiredSkills, &proposal.DeploymentDurationDays, &proposal.ExpectedStartDate,
		&proposal.ExpectedEndDate, &proposal.SpecialRequirements, &proposal.Justification,
		&proposal.Status, &proposal.ReviewedBy, &proposal.ReviewedAt, &proposal.ReviewNotes,
		&proposal.RejectionReason, &proposal.CreatedAt, &proposal.UpdatedAt,
		&outbreakName, &outbreakDesc, &proposedByName, &reviewedByName,
	)
	if err != nil {
		return nil, err
	}

	// Convert pq.StringArray to []string
	proposal.RequiredSkills = []string(requiredSkills)

	// Set related data
	if outbreakName.Valid {
		proposal.Outbreak = &models.Outbreak{
			ID:          int(proposal.OutbreakID),
			Name:        outbreakName,
			Description: outbreakDesc,
		}
	}

	return proposal, nil
}

func (h *RRTTeamMembersHandler) getAllRRTDeploymentProposals() ([]*models.RRTDeploymentProposal, error) {
	query := `
		SELECT p.id, p.proposal_number, p.outbreak_id, p.proposed_by, p.proposed_date,
		       p.deployment_purpose, p.proposed_team_composition, p.required_skills,
		       p.deployment_duration_days, p.expected_start_date, p.expected_end_date,
		       p.special_requirements, p.justification, p.status, p.reviewed_by,
		       p.reviewed_at, p.review_notes, p.rejection_reason, p.created_at, p.updated_at,
		       o.name as outbreak_name,
		       u.user_name as proposed_by_name
		FROM rrt_deployment_proposals p
		LEFT JOIN outbreaks o ON p.outbreak_id = o.id
		LEFT JOIN users u ON p.proposed_by = u.user_id
		ORDER BY p.created_at DESC
		LIMIT 100
	`

	rows, err := h.db.Query(query)
	if err != nil {
		log.Printf("Error querying deployment proposals: %v", err)
		return []*models.RRTDeploymentProposal{}, nil
	}
	defer rows.Close()

	var proposals []*models.RRTDeploymentProposal
	for rows.Next() {
		proposal := &models.RRTDeploymentProposal{}
		var outbreakName, proposedByName sql.NullString
		var requiredSkills pq.StringArray

		err := rows.Scan(
			&proposal.ID, &proposal.ProposalNumber, &proposal.OutbreakID, &proposal.ProposedBy,
			&proposal.ProposedDate, &proposal.DeploymentPurpose, &proposal.ProposedTeamComposition,
			&requiredSkills, &proposal.DeploymentDurationDays, &proposal.ExpectedStartDate,
			&proposal.ExpectedEndDate, &proposal.SpecialRequirements, &proposal.Justification,
			&proposal.Status, &proposal.ReviewedBy, &proposal.ReviewedAt, &proposal.ReviewNotes,
			&proposal.RejectionReason, &proposal.CreatedAt, &proposal.UpdatedAt,
			&outbreakName, &proposedByName,
		)
		if err != nil {
			log.Printf("Error scanning proposal: %v", err)
			continue
		}

		// Convert pq.StringArray to []string
		proposal.RequiredSkills = []string(requiredSkills)

		// Set related data
		if outbreakName.Valid {
			proposal.Outbreak = &models.Outbreak{
				ID:   int(proposal.OutbreakID),
				Name: outbreakName,
			}
		}

		proposals = append(proposals, proposal)
	}

	return proposals, nil
}

func (h *RRTTeamMembersHandler) createRRTDeploymentProposal(proposal *models.RRTDeploymentProposal) error {
	query := `
		INSERT INTO rrt_deployment_proposals (
			outbreak_id, proposed_by, proposed_date, deployment_purpose,
			proposed_team_composition, required_skills, deployment_duration_days,
			expected_start_date, expected_end_date, special_requirements,
			justification, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, proposal_number
	`

	err := h.db.QueryRow(query,
		proposal.OutbreakID,
		proposal.ProposedBy,
		proposal.ProposedDate,
		proposal.DeploymentPurpose,
		proposal.ProposedTeamComposition,
		pq.Array(proposal.RequiredSkills),
		proposal.DeploymentDurationDays,
		proposal.ExpectedStartDate,
		proposal.ExpectedEndDate,
		proposal.SpecialRequirements,
		proposal.Justification,
		proposal.Status,
	).Scan(&proposal.ID, &proposal.ProposalNumber)

	return err
}

func (h *RRTTeamMembersHandler) reviewRRTDeploymentProposal(proposalID, action, reviewNotes, rejectionReason string) error {
	query := `
		UPDATE rrt_deployment_proposals 
		SET status = $1, review_notes = $2, rejection_reason = $3,
		    reviewed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
	`

	_, err := h.db.Exec(query, action, reviewNotes, rejectionReason, proposalID)
	return err
}

func (h *RRTTeamMembersHandler) getAllRRTDeployments() ([]*models.RRTDeployment, error) {
	query := `
		SELECT d.id, d.team_id, d.outbreak_id, d.deployment_date, d.expected_return_date,
		       d.actual_return_date, d.deployment_status, d.deployment_purpose, d.assigned_vehicle,
		       d.assigned_driver, d.deployment_notes, d.created_at, d.updated_at,
		       t.team_name, t.team_code,
		       o.name as outbreak_name
		FROM rrt_deployments d
		LEFT JOIN rrt_teams t ON d.team_id = t.id
		LEFT JOIN outbreaks o ON d.outbreak_id = o.id
		WHERE d.deployment_status IN ('deployed', 'extended')
		ORDER BY d.deployment_date DESC
	`

	rows, err := h.db.Query(query)
	if err != nil {
		log.Printf("Error querying deployments: %v", err)
		return []*models.RRTDeployment{}, nil
	}
	defer rows.Close()

	var deployments []*models.RRTDeployment
	for rows.Next() {
		deployment := &models.RRTDeployment{}
		var teamName, teamCode, outbreakName sql.NullString

		err := rows.Scan(
			&deployment.ID, &deployment.TeamID, &deployment.OutbreakID, &deployment.DeploymentDate,
			&deployment.ExpectedReturnDate, &deployment.ActualReturnDate, &deployment.DeploymentStatus,
			&deployment.DeploymentPurpose, &deployment.AssignedVehicle, &deployment.AssignedDriver,
			&deployment.DeploymentNotes, &deployment.CreatedAt, &deployment.UpdatedAt,
			&teamName, &teamCode, &outbreakName,
		)
		if err != nil {
			log.Printf("Error scanning deployment: %v", err)
			continue
		}

		// Set related data
		if teamName.Valid {
			deployment.Team = &models.RRTTeam{
				ID:       deployment.TeamID,
				TeamName: teamName.String,
				TeamCode: teamCode.String,
			}
		}
		if outbreakName.Valid {
			deployment.Outbreak = &models.Outbreak{
				ID:   int(deployment.OutbreakID),
				Name: outbreakName,
			}
		}

		deployments = append(deployments, deployment)
	}

	return deployments, nil
}

func (h *RRTTeamMembersHandler) getAllRRTDeploymentExtensions() ([]*models.RRTDeploymentExtension, error) {
	query := `
		SELECT e.id, e.deployment_id, e.extension_reason, e.original_end_date,
		       e.new_end_date, e.requested_by, e.requested_date, e.approved_by,
		       e.approved_date, e.status, e.approval_notes,
		       d.team_id, t.team_name
		FROM rrt_deployment_extensions e
		LEFT JOIN rrt_deployments d ON e.deployment_id = d.id
		LEFT JOIN rrt_teams t ON d.team_id = t.id
		ORDER BY e.requested_date DESC
		LIMIT 100
	`

	rows, err := h.db.Query(query)
	if err != nil {
		log.Printf("Error querying deployment extensions: %v", err)
		return []*models.RRTDeploymentExtension{}, nil
	}
	defer rows.Close()

	var extensions []*models.RRTDeploymentExtension
	for rows.Next() {
		extension := &models.RRTDeploymentExtension{}
		var teamID sql.NullInt64
		var teamName sql.NullString

		err := rows.Scan(
			&extension.ID, &extension.DeploymentID, &extension.ExtensionReason,
			&extension.OriginalEndDate, &extension.NewEndDate, &extension.RequestedBy,
			&extension.RequestedDate, &extension.ApprovedBy, &extension.ApprovedDate,
			&extension.Status, &extension.ApprovalNotes,
			&teamID, &teamName,
		)
		if err != nil {
			log.Printf("Error scanning extension: %v", err)
			continue
		}

		// Set related deployment info
		if teamName.Valid {
			extension.Deployment = &models.RRTDeployment{
				ID:     extension.DeploymentID,
				TeamID: teamID.Int64,
				Team: &models.RRTTeam{
					TeamName: teamName.String,
				},
			}
		}

		extensions = append(extensions, extension)
	}

	return extensions, nil
}

func (h *RRTTeamMembersHandler) createRRTDeploymentExtension(extension *models.RRTDeploymentExtension) error {
	return nil
}

func (h *RRTTeamMembersHandler) reviewRRTDeploymentExtension(extensionID, action, approvalNotes, rejectionReason string) error {
	return nil
}

func (h *RRTTeamMembersHandler) getAllRRTFieldRoleAssignments() ([]*models.RRTFieldRoleAssignment, error) {
	return []*models.RRTFieldRoleAssignment{}, nil
}

func (h *RRTTeamMembersHandler) createRRTFieldRoleAssignment(assignment *models.RRTFieldRoleAssignment) error {
	return nil
}

func (h *RRTTeamMembersHandler) getAllPillars() ([]*models.Pillar, error) {
	query := `
		SELECT id, name, description, pillar_head_id, pillar_head_name, 
		       pillar_head_email, pillar_head_phone, is_active, created_at, 
		       updated_at, created_by, updated_by
		FROM pillars
		WHERE is_active = true
		ORDER BY name
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pillars []*models.Pillar
	for rows.Next() {
		pillar := &models.Pillar{}
		err := rows.Scan(
			&pillar.ID, &pillar.Name, &pillar.Description, &pillar.PillarHeadID,
			&pillar.PillarHeadName, &pillar.PillarHeadEmail, &pillar.PillarHeadPhone,
			&pillar.IsActive, &pillar.CreatedAt, &pillar.UpdatedAt,
			&pillar.CreatedBy, &pillar.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		pillars = append(pillars, pillar)
	}

	return pillars, nil
}

// Helper functions for parsing form data (renamed to avoid conflicts)
func parseRRTDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}
	t, _ := time.Parse("2006-01-02", dateStr)
	return t
}

func parseRRTNullDate(dateStr string) sql.NullTime {
	if dateStr == "" {
		return sql.NullTime{Valid: false}
	}
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: t, Valid: true}
}

func parseRRTNullString(str string) sql.NullString {
	if str == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: str, Valid: true}
}

func parseRRTNullInt64(str string) sql.NullInt64 {
	if str == "" {
		return sql.NullInt64{Valid: false}
	}
	val, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: val, Valid: true}
}
