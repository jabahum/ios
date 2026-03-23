package handlers

import (
	"case/internal/models"
	"database/sql"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/lib/pq"
)

// ResourceManagementHandler handles all resource management operations
type ResourceManagementHandler struct {
	db    *sql.DB
	store *session.Store
}

// NewResourceManagementHandler creates a new resource management handler
func NewResourceManagementHandler(db *sql.DB, store *session.Store) *ResourceManagementHandler {
	return &ResourceManagementHandler{db: db, store: store}
}

// HandlerResourceDashboard displays the main resource management dashboard
func (h *ResourceManagementHandler) HandlerResourceDashboard(c *fiber.Ctx) error {
	data := NewTemplateDataWithDB(c, h.store, h.db)

	// Get summary statistics
	stats, err := h.getResourceManagementStats()
	if err != nil {
		log.Printf("Error getting resource management stats: %v", err)
		stats = make(map[string]interface{})
	}

	data.ResourceStats = stats

	return GenerateHTML(c, h.db, data, "resource_dashboard")
}

// HandlerRRTTeamsList displays all RRT teams
func (h *ResourceManagementHandler) HandlerRRTTeamsList(c *fiber.Ctx) error {
	data := NewTemplateDataWithDB(c, h.store, h.db)

	teams, err := h.getAllRRTTeams()
	if err != nil {
		log.Printf("Error getting RRT teams: %v", err)
		teams = []*models.RRTTeam{}
	}

	data.RRTTeams = teams
	return GenerateHTML(c, h.db, data, "rrt_teams_list")
}

// HandlerRRTTeamForm displays the form for adding/editing RRT teams
func (h *ResourceManagementHandler) HandlerRRTTeamForm(c *fiber.Ctx) error {
	data := NewTemplateDataWithDB(c, h.store, h.db)

	// Get all team members for selection
	teamMembers, err := h.getAllRRTTeamMembers()
	if err != nil {
		log.Printf("Error getting team members: %v", err)
		teamMembers = []*models.RRTTeamMember{}
	}
	data.RRTTeamMembers = teamMembers

	// If editing, get the team
	teamID := c.Params("id")
	if teamID != "" {
		team, err := h.getRRTTeamByID(teamID)
		if err != nil {
			log.Printf("Error getting RRT team: %v", err)
		} else {
			data.RRTTeam = team
		}
	}

	return GenerateHTML(c, h.db, data, "rrt_team_form")
}

// HandlerRRTTeamSave saves an RRT team
func (h *ResourceManagementHandler) HandlerRRTTeamSave(c *fiber.Ctx) error {
	teamID := c.FormValue("id")

	team := &models.RRTTeam{
		TeamName:      c.FormValue("team_name"),
		TeamCode:      c.FormValue("team_code"),
		TeamType:      c.FormValue("team_type"),
		TeamLeadName:  c.FormValue("team_lead_name"),
		TeamLeadPhone: parseResourceNullString(c.FormValue("team_lead_phone")),
		TeamLeadEmail: parseResourceNullString(c.FormValue("team_lead_email")),
		TeamSize:      int(parseResourceInt(c.FormValue("team_size"))),
		BaseLocation:  parseResourceNullString(c.FormValue("base_location")),
		IsActive:      c.FormValue("is_active") == "true",
	}

	// Parse specializations from JSON
	if specializationsJSON := c.FormValue("specializations"); specializationsJSON != "" {
		if err := json.Unmarshal([]byte(specializationsJSON), &team.Specializations); err != nil {
			log.Printf("Error parsing specializations: %v", err)
		}
	}

	var err error
	if teamID == "" {
		// Create new team
		err = h.createRRTTeam(team)
	} else {
		// Update existing team
		team.ID = parseResourceInt(teamID)
		err = h.updateRRTTeam(team)
	}

	if err != nil {
		log.Printf("Error saving RRT team: %v", err)
		return c.Status(500).SendString("Error saving RRT team")
	}

	return c.Redirect("/resource-management/rrt-teams")
}

// HandlerRRTDeploymentsList displays all RRT deployments
func (h *ResourceManagementHandler) HandlerRRTDeploymentsList(c *fiber.Ctx) error {
	data := NewTemplateDataWithDB(c, h.store, h.db)

	// Get outbreak filter
	outbreakID := c.Query("outbreak_id")
	var deployments []*models.RRTDeployment
	var err error

	if outbreakID != "" {
		deployments, err = h.getDeploymentsByOutbreak(outbreakID)
	} else {
		deployments, err = h.getAllRRTDeployments()
	}

	if err != nil {
		log.Printf("Error getting RRT deployments: %v", err)
		deployments = []*models.RRTDeployment{}
	}

	data.RRTDeployments = deployments
	return GenerateHTML(c, h.db, data, "rrt_deployments_list")
}

// HandlerRRTDeploymentForm displays the form for adding/editing RRT deployments
func (h *ResourceManagementHandler) HandlerRRTDeploymentForm(c *fiber.Ctx) error {
	data := NewTemplateDataWithDB(c, h.store, h.db)

	// Get teams for dropdown
	teams, err := h.getAllRRTTeams()
	if err != nil {
		log.Printf("Error getting RRT teams: %v", err)
		teams = []*models.RRTTeam{}
	}

	// Get outbreaks for dropdown
	outbreaks, err := h.getAllOutbreaks()
	if err != nil {
		log.Printf("Error getting outbreaks: %v", err)
		outbreaks = []*models.Outbreak{}
	}

	data.RRTTeams = teams
	data.Outbreaks = outbreaks

	// If editing, get the deployment
	deploymentID := c.Params("id")
	if deploymentID != "" {
		deployment, err := h.getRRTDeploymentByID(deploymentID)
		if err != nil {
			log.Printf("Error getting RRT deployment: %v", err)
		} else {
			data.RRTDeployment = deployment
		}
	}

	return GenerateHTML(c, h.db, data, "rrt_deployment_form")
}

// HandlerRRTDeploymentSave saves an RRT deployment
func (h *ResourceManagementHandler) HandlerRRTDeploymentSave(c *fiber.Ctx) error {
	deploymentID := c.FormValue("id")

	deployment := &models.RRTDeployment{
		TeamID:             parseResourceInt(c.FormValue("team_id")),
		OutbreakID:         parseResourceInt(c.FormValue("outbreak_id")),
		DeploymentDate:     parseResourceDate(c.FormValue("deployment_date")),
		ExpectedReturnDate: parseResourceNullDate(c.FormValue("expected_return_date")),
		DeploymentStatus:   c.FormValue("deployment_status"),
		DeploymentPurpose:  parseResourceNullString(c.FormValue("deployment_purpose")),
		AssignedVehicle:    parseResourceNullString(c.FormValue("assigned_vehicle")),
		AssignedDriver:     parseResourceNullString(c.FormValue("assigned_driver")),
		DeploymentNotes:    parseResourceNullString(c.FormValue("deployment_notes")),
	}

	var err error
	if deploymentID == "" {
		// Create new deployment
		err = h.createRRTDeployment(deployment)
	} else {
		// Update existing deployment
		deployment.ID = parseResourceInt(deploymentID)
		err = h.updateRRTDeployment(deployment)
	}

	if err != nil {
		log.Printf("Error saving RRT deployment: %v", err)
		return c.Status(500).SendString("Error saving RRT deployment")
	}

	return c.Redirect("/resource-management/rrt-deployments")
}

// HandlerResourcesList displays all resources
func (h *ResourceManagementHandler) HandlerResourcesList(c *fiber.Ctx) error {
	data := NewTemplateDataWithDB(c, h.store, h.db)

	resources, err := h.getAllResources()
	if err != nil {
		log.Printf("Error getting resources: %v", err)
		resources = []*models.Resource{}
	}

	data.Resources = resources
	return GenerateHTML(c, h.db, data, "resources_list")
}

// HandlerResourceForm displays the form for adding/editing resources
func (h *ResourceManagementHandler) HandlerResourceForm(c *fiber.Ctx) error {
	data := NewTemplateDataWithDB(c, h.store, h.db)

	// Get categories for dropdown
	categories, err := h.getAllResourceCategories()
	if err != nil {
		log.Printf("Error getting resource categories: %v", err)
		categories = []*models.ResourceCategory{}
	}

	data.ResourceCategories = categories

	// If editing, get the resource
	resourceID := c.Params("id")
	if resourceID != "" {
		resource, err := h.getResourceByID(resourceID)
		if err != nil {
			log.Printf("Error getting resource: %v", err)
		} else {
			data.Resource = resource
		}
	}

	return GenerateHTML(c, h.db, data, "resource_form")
}

// HandlerResourceSave saves a resource
func (h *ResourceManagementHandler) HandlerResourceSave(c *fiber.Ctx) error {
	resourceID := c.FormValue("id")

	resource := &models.Resource{
		Name:          c.FormValue("name"),
		Description:   parseResourceNullString(c.FormValue("description")),
		ResourceCode:  parseResourceNullString(c.FormValue("resource_code")),
		CategoryID:    parseResourceInt(c.FormValue("category_id")),
		UnitOfMeasure: c.FormValue("unit_of_measure"),
		IsConsumable:  c.FormValue("is_consumable") == "true",
		HasExpiry:     c.FormValue("has_expiry") == "true",
		ShelfLifeDays: parseResourceNullInt64(c.FormValue("shelf_life_days")),
		IsCritical:    c.FormValue("is_critical") == "true",
		IsActive:      c.FormValue("is_active") == "true",
	}

	var err error
	if resourceID == "" {
		// Create new resource
		err = h.createResource(resource)
	} else {
		// Update existing resource
		resource.ID = parseResourceInt(resourceID)
		err = h.updateResource(resource)
	}

	if err != nil {
		log.Printf("Error saving resource: %v", err)
		return c.Status(500).SendString("Error saving resource")
	}

	return c.Redirect("/resource-management/resources")
}

// HandlerRequisitionsList displays all requisitions
func (h *ResourceManagementHandler) HandlerRequisitionsList(c *fiber.Ctx) error {
	data := NewTemplateDataWithDB(c, h.store, h.db)

	// Get outbreak filter
	outbreakID := c.Query("outbreak_id")
	var requisitions []*models.Requisition
	var err error

	if outbreakID != "" {
		requisitions, err = h.getRequisitionsByOutbreak(outbreakID)
	} else {
		requisitions, err = h.getAllRequisitions()
	}

	if err != nil {
		log.Printf("Error getting requisitions: %v", err)
		requisitions = []*models.Requisition{}
	}

	data.Requisitions = requisitions
	return GenerateHTML(c, h.db, data, "requisitions_list")
}

// HandlerRequisitionForm displays the form for adding/editing requisitions
func (h *ResourceManagementHandler) HandlerRequisitionForm(c *fiber.Ctx) error {
	data := NewTemplateDataWithDB(c, h.store, h.db)

	// Get outbreaks for dropdown
	outbreaks, err := h.getAllOutbreaks()
	if err != nil {
		log.Printf("Error getting outbreaks: %v", err)
		outbreaks = []*models.Outbreak{}
	}

	// Get deployments for dropdown
	deployments, err := h.getAllRRTDeployments()
	if err != nil {
		log.Printf("Error getting deployments: %v", err)
		deployments = []*models.RRTDeployment{}
	}

	// Get resources for dropdown
	resources, err := h.getAllResources()
	if err != nil {
		log.Printf("Error getting resources: %v", err)
		resources = []*models.Resource{}
	}

	data.Outbreaks = outbreaks
	data.RRTDeployments = deployments
	data.Resources = resources

	// If editing, get the requisition
	requisitionID := c.Params("id")
	if requisitionID != "" {
		requisition, err := h.getRequisitionByID(requisitionID)
		if err != nil {
			log.Printf("Error getting requisition: %v", err)
		} else {
			data.Requisition = requisition
		}
	}

	return GenerateHTML(c, h.db, data, "requisition_form")
}

// HandlerRequisitionSave saves a requisition
func (h *ResourceManagementHandler) HandlerRequisitionSave(c *fiber.Ctx) error {
	requisitionID := c.FormValue("id")

	// Get current user
	userID := GetCurrentUser(c, h.store)
	if userID == 0 {
		return c.Status(401).SendString("Unauthorized")
	}

	requisition := &models.Requisition{
		RequisitionNumber: c.FormValue("requisition_number"),
		OutbreakID:        parseResourceInt(c.FormValue("outbreak_id")),
		DeploymentID:      parseResourceNullInt64(c.FormValue("deployment_id")),
		RequestedBy:       int64(userID),
		RequiredDate:      parseResourceNullDate(c.FormValue("required_date")),
		Priority:          c.FormValue("priority"),
		Status:            c.FormValue("status"),
		Notes:             parseResourceNullString(c.FormValue("notes")),
	}

	var err error
	if requisitionID == "" {
		// Create new requisition
		err = h.createRequisition(requisition)
	} else {
		// Update existing requisition
		requisition.ID = parseResourceInt(requisitionID)
		err = h.updateRequisition(requisition)
	}

	if err != nil {
		log.Printf("Error saving requisition: %v", err)
		return c.Status(500).SendString("Error saving requisition")
	}

	return c.Redirect("/resource-management/requisitions")
}

// HandlerActivityLogsList displays all activity logs
func (h *ResourceManagementHandler) HandlerActivityLogsList(c *fiber.Ctx) error {
	data := NewTemplateDataWithDB(c, h.store, h.db)

	// Get deployment filter
	deploymentID := c.Query("deployment_id")
	var activities []*models.ActivityLog
	var err error

	if deploymentID != "" {
		activities, err = h.getActivitiesByDeployment(deploymentID)
	} else {
		activities, err = h.getAllActivityLogs()
	}

	if err != nil {
		log.Printf("Error getting activity logs: %v", err)
		activities = []*models.ActivityLog{}
	}

	data.ActivityLogs = activities
	return GenerateHTML(c, h.db, data, "activity_logs_list")
}

// HandlerActivityLogForm displays the form for adding/editing activity logs
func (h *ResourceManagementHandler) HandlerActivityLogForm(c *fiber.Ctx) error {
	data := NewTemplateDataWithDB(c, h.store, h.db)

	// Get deployments for dropdown
	deployments, err := h.getAllRRTDeployments()
	if err != nil {
		log.Printf("Error getting deployments: %v", err)
		deployments = []*models.RRTDeployment{}
	}

	data.RRTDeployments = deployments

	// If editing, get the activity log
	activityID := c.Params("id")
	if activityID != "" {
		activity, err := h.getActivityLogByID(activityID)
		if err != nil {
			log.Printf("Error getting activity log: %v", err)
		} else {
			data.ActivityLog = activity
		}
	}

	return GenerateHTML(c, h.db, data, "activity_log_form")
}

// HandlerActivityLogSave saves an activity log
func (h *ResourceManagementHandler) HandlerActivityLogSave(c *fiber.Ctx) error {
	activityID := c.FormValue("id")

	activity := &models.ActivityLog{
		DeploymentID:        parseResourceInt(c.FormValue("deployment_id")),
		ActivityType:        c.FormValue("activity_type"),
		ActivityDate:        parseResourceDate(c.FormValue("activity_date")),
		StartTime:           parseResourceNullTime(c.FormValue("start_time")),
		EndTime:             parseResourceNullTime(c.FormValue("end_time")),
		Location:            parseResourceNullString(c.FormValue("location")),
		ParticipantsCount:   parseResourceNullInt64(c.FormValue("participants_count")),
		ActivityDescription: parseResourceNullString(c.FormValue("activity_description")),
		Outcomes:            parseResourceNullString(c.FormValue("outcomes")),
		Challenges:          parseResourceNullString(c.FormValue("challenges")),
		Recommendations:     parseResourceNullString(c.FormValue("recommendations")),
		ResourcesUsed:       parseResourceNullString(c.FormValue("resources_used")),
	}

	var err error
	if activityID == "" {
		// Create new activity log
		err = h.createActivityLog(activity)
	} else {
		// Update existing activity log
		activity.ID = parseResourceInt(activityID)
		err = h.updateActivityLog(activity)
	}

	if err != nil {
		log.Printf("Error saving activity log: %v", err)
		return c.Status(500).SendString("Error saving activity log")
	}

	return c.Redirect("/resource-management/activity-logs")
}

// HandlerSitRepGeneration displays the SitRep generation interface
func (h *ResourceManagementHandler) HandlerSitRepGeneration(c *fiber.Ctx) error {
	data := NewTemplateDataWithDB(c, h.store, h.db)

	// Get outbreaks for dropdown
	outbreaks, err := h.getAllOutbreaks()
	if err != nil {
		log.Printf("Error getting outbreaks: %v", err)
		outbreaks = []*models.Outbreak{}
	}

	data.Outbreaks = outbreaks
	return GenerateHTML(c, h.db, data, "sitrep_generation")
}

// HandlerGenerateSitRep generates a SitRep for an outbreak
func (h *ResourceManagementHandler) HandlerGenerateSitRep(c *fiber.Ctx) error {
	outbreakID := parseResourceInt(c.FormValue("outbreak_id"))
	reportType := c.FormValue("report_type")
	periodStart := parseResourceDate(c.FormValue("period_start"))
	periodEnd := parseResourceDate(c.FormValue("period_end"))

	// Generate SitRep content
	sitrep, err := h.generateSitRep(outbreakID, reportType, periodStart, periodEnd)
	if err != nil {
		log.Printf("Error generating SitRep: %v", err)
		return c.Status(500).SendString("Error generating SitRep")
	}

	// Save generated SitRep
	err = h.saveGeneratedSitRep(sitrep)
	if err != nil {
		log.Printf("Error saving SitRep: %v", err)
		return c.Status(500).SendString("Error saving SitRep")
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"message":   "SitRep generated successfully",
		"sitrep_id": sitrep.ID,
	})
}

// Database helper methods

func (h *ResourceManagementHandler) getResourceManagementStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get total RRT teams
	var totalTeams int
	err := h.db.QueryRow("SELECT COUNT(*) FROM rrt_teams WHERE is_active = true").Scan(&totalTeams)
	if err != nil {
		return nil, err
	}
	stats["total_teams"] = totalTeams

	// Get active deployments
	var activeDeployments int
	err = h.db.QueryRow("SELECT COUNT(*) FROM rrt_deployments WHERE deployment_status IN ('deployed', 'extended')").Scan(&activeDeployments)
	if err != nil {
		return nil, err
	}
	stats["active_deployments"] = activeDeployments

	// Get total resources
	var totalResources int
	err = h.db.QueryRow("SELECT COUNT(*) FROM resources WHERE is_active = true").Scan(&totalResources)
	if err != nil {
		return nil, err
	}
	stats["total_resources"] = totalResources

	// Get pending requisitions
	var pendingRequisitions int
	err = h.db.QueryRow("SELECT COUNT(*) FROM requisitions WHERE status = 'pending'").Scan(&pendingRequisitions)
	if err != nil {
		return nil, err
	}
	stats["pending_requisitions"] = pendingRequisitions

	return stats, nil
}

func (h *ResourceManagementHandler) getAllRRTTeams() ([]*models.RRTTeam, error) {
	query := `SELECT id, team_name, team_code, team_type, team_lead_name, team_lead_phone, 
	          team_lead_email, team_size, specializations, base_location, is_active, 
	          created_at, updated_at, created_by 
	          FROM rrt_teams ORDER BY team_name`

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

func (h *ResourceManagementHandler) getAllRRTTeamMembers() ([]*models.RRTTeamMember, error) {
	query := `
		SELECT rtm.id, rtm.first_name, rtm.last_name, rtm.email, rtm.phone, 
		       rtm.national_id, rtm.employee_id, rtm.organization, rtm.position,
		       rtm.qualifications, rtm.specializations, rtm.certifications,
		       rtm.experience_years, rtm.is_driver, rtm.driver_license, 
		       rtm.driver_license_expiry, rtm.is_active, rtm.created_at, 
		       rtm.updated_at, rtm.created_by, rtm.pillar_id
		FROM rrt_team_members rtm
		WHERE rtm.is_active = true
		ORDER BY rtm.first_name, rtm.last_name
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*models.RRTTeamMember
	for rows.Next() {
		member := &models.RRTTeamMember{}
		var specializations pq.StringArray

		err := rows.Scan(
			&member.ID, &member.FirstName, &member.LastName, &member.Email, &member.Phone,
			&member.NationalID, &member.EmployeeID, &member.Organization, &member.Position,
			&member.Qualifications, &specializations, &member.Certifications,
			&member.ExperienceYears, &member.IsDriver, &member.DriverLicense,
			&member.DriverLicenseExpiry, &member.IsActive, &member.CreatedAt,
			&member.UpdatedAt, &member.CreatedBy, &member.PillarID,
		)
		if err != nil {
			return nil, err
		}

		// Convert pq.StringArray to []string
		member.Specializations = []string(specializations)
		members = append(members, member)
	}

	return members, nil
}

func (h *ResourceManagementHandler) getRRTTeamByID(id string) (*models.RRTTeam, error) {
	query := `SELECT id, team_name, team_code, team_type, team_lead_name, team_lead_phone, 
	          team_lead_email, team_size, specializations, base_location, is_active, 
	          created_at, updated_at, created_by 
	          FROM rrt_teams WHERE id = $1`

	var team models.RRTTeam
	var specializations pq.StringArray
	err := h.db.QueryRow(query, id).Scan(
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
	return &team, nil
}

func (h *ResourceManagementHandler) createRRTTeam(team *models.RRTTeam) error {
	query := `INSERT INTO rrt_teams (team_name, team_code, team_type, team_lead_name, team_lead_phone, 
	          team_lead_email, team_size, specializations, base_location, is_active, created_by) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`

	return h.db.QueryRow(query,
		team.TeamName, team.TeamCode, team.TeamType, team.TeamLeadName,
		team.TeamLeadPhone, team.TeamLeadEmail, team.TeamSize, pq.Array(team.Specializations),
		team.BaseLocation, team.IsActive, team.CreatedBy,
	).Scan(&team.ID)
}

func (h *ResourceManagementHandler) updateRRTTeam(team *models.RRTTeam) error {
	query := `UPDATE rrt_teams SET team_name = $1, team_code = $2, team_type = $3, team_lead_name = $4, 
	          team_lead_phone = $5, team_lead_email = $6, team_size = $7, specializations = $8, 
	          base_location = $9, is_active = $10, updated_at = CURRENT_TIMESTAMP 
	          WHERE id = $11`

	_, err := h.db.Exec(query,
		team.TeamName, team.TeamCode, team.TeamType, team.TeamLeadName,
		team.TeamLeadPhone, team.TeamLeadEmail, team.TeamSize, pq.Array(team.Specializations),
		team.BaseLocation, team.IsActive, team.ID,
	)
	return err
}

// Helper functions for parsing form data (resource management specific)
func parseResourceInt(s string) int64 {
	if s == "" {
		return 0
	}
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

func parseResourceNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func parseResourceNullInt64(s string) sql.NullInt64 {
	if s == "" {
		return sql.NullInt64{Valid: false}
	}
	i, _ := strconv.ParseInt(s, 10, 64)
	return sql.NullInt64{Int64: i, Valid: true}
}

func parseResourceDate(s string) time.Time {
	if s == "" {
		return time.Now()
	}
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func parseResourceNullDate(s string) sql.NullTime {
	if s == "" {
		return sql.NullTime{Valid: false}
	}
	t, _ := time.Parse("2006-01-02", s)
	return sql.NullTime{Time: t, Valid: true}
}

func parseResourceNullTime(s string) sql.NullTime {
	if s == "" {
		return sql.NullTime{Valid: false}
	}
	t, _ := time.Parse("15:04", s)
	return sql.NullTime{Time: t, Valid: true}
}

// RRT Deployment database methods
func (h *ResourceManagementHandler) getAllRRTDeployments() ([]*models.RRTDeployment, error) {
	query := `
		SELECT d.id, d.team_id, d.outbreak_id, d.deployment_date, d.expected_return_date,
		       d.actual_return_date, d.deployment_status, d.deployment_purpose, d.assigned_vehicle,
		       d.assigned_driver, d.deployment_notes, d.created_at, d.updated_at,
		       t.team_name, t.team_code,
		       o.name as outbreak_name
		FROM rrt_deployments d
		LEFT JOIN rrt_teams t ON d.team_id = t.id
		LEFT JOIN outbreaks o ON d.outbreak_id = o.id
		ORDER BY d.created_at DESC
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
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
			return nil, err
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

func (h *ResourceManagementHandler) getDeploymentsByOutbreak(outbreakID string) ([]*models.RRTDeployment, error) {
	query := `
		SELECT d.id, d.team_id, d.outbreak_id, d.deployment_date, d.expected_return_date,
		       d.actual_return_date, d.deployment_status, d.deployment_purpose, d.assigned_vehicle,
		       d.assigned_driver, d.deployment_notes, d.created_at, d.updated_at,
		       t.team_name, t.team_code
		FROM rrt_deployments d
		LEFT JOIN rrt_teams t ON d.team_id = t.id
		WHERE d.outbreak_id = $1
		ORDER BY d.created_at DESC
	`

	rows, err := h.db.Query(query, outbreakID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []*models.RRTDeployment
	for rows.Next() {
		deployment := &models.RRTDeployment{}
		var teamName, teamCode sql.NullString

		err := rows.Scan(
			&deployment.ID, &deployment.TeamID, &deployment.OutbreakID, &deployment.DeploymentDate,
			&deployment.ExpectedReturnDate, &deployment.ActualReturnDate, &deployment.DeploymentStatus,
			&deployment.DeploymentPurpose, &deployment.AssignedVehicle, &deployment.AssignedDriver,
			&deployment.DeploymentNotes, &deployment.CreatedAt, &deployment.UpdatedAt,
			&teamName, &teamCode,
		)
		if err != nil {
			return nil, err
		}

		if teamName.Valid {
			deployment.Team = &models.RRTTeam{
				ID:       deployment.TeamID,
				TeamName: teamName.String,
				TeamCode: teamCode.String,
			}
		}

		deployments = append(deployments, deployment)
	}

	return deployments, nil
}

func (h *ResourceManagementHandler) getRRTDeploymentByID(id string) (*models.RRTDeployment, error) {
	query := `
		SELECT d.id, d.team_id, d.outbreak_id, d.deployment_date, d.expected_return_date,
		       d.actual_return_date, d.deployment_status, d.deployment_purpose, d.assigned_vehicle,
		       d.assigned_driver, d.deployment_notes, d.created_at, d.updated_at,
		       t.team_name, t.team_code,
		       o.name as outbreak_name
		FROM rrt_deployments d
		LEFT JOIN rrt_teams t ON d.team_id = t.id
		LEFT JOIN outbreaks o ON d.outbreak_id = o.id
		WHERE d.id = $1
	`

	deployment := &models.RRTDeployment{}
	var teamName, teamCode, outbreakName sql.NullString

	err := h.db.QueryRow(query, id).Scan(
		&deployment.ID, &deployment.TeamID, &deployment.OutbreakID, &deployment.DeploymentDate,
		&deployment.ExpectedReturnDate, &deployment.ActualReturnDate, &deployment.DeploymentStatus,
		&deployment.DeploymentPurpose, &deployment.AssignedVehicle, &deployment.AssignedDriver,
		&deployment.DeploymentNotes, &deployment.CreatedAt, &deployment.UpdatedAt,
		&teamName, &teamCode, &outbreakName,
	)
	if err != nil {
		return nil, err
	}

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

	return deployment, nil
}

func (h *ResourceManagementHandler) createRRTDeployment(deployment *models.RRTDeployment) error {
	query := `
		INSERT INTO rrt_deployments (
			team_id, outbreak_id, deployment_date, expected_return_date,
			deployment_status, deployment_purpose, assigned_vehicle, assigned_driver,
			deployment_notes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`

	return h.db.QueryRow(query,
		deployment.TeamID, deployment.OutbreakID, deployment.DeploymentDate,
		deployment.ExpectedReturnDate, deployment.DeploymentStatus, deployment.DeploymentPurpose,
		deployment.AssignedVehicle, deployment.AssignedDriver, deployment.DeploymentNotes,
	).Scan(&deployment.ID)
}

func (h *ResourceManagementHandler) updateRRTDeployment(deployment *models.RRTDeployment) error {
	query := `
		UPDATE rrt_deployments SET
			team_id = $1, outbreak_id = $2, deployment_date = $3, expected_return_date = $4,
			deployment_status = $5, deployment_purpose = $6, assigned_vehicle = $7,
			assigned_driver = $8, deployment_notes = $9, actual_return_date = $10,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $11
	`

	_, err := h.db.Exec(query,
		deployment.TeamID, deployment.OutbreakID, deployment.DeploymentDate,
		deployment.ExpectedReturnDate, deployment.DeploymentStatus, deployment.DeploymentPurpose,
		deployment.AssignedVehicle, deployment.AssignedDriver, deployment.DeploymentNotes,
		deployment.ActualReturnDate, deployment.ID,
	)

	return err
}

// Resource database methods
func (h *ResourceManagementHandler) getAllResources() ([]*models.Resource, error) {
	query := `
		SELECT r.id, r.name, r.description, r.resource_code, r.category_id,
		       r.unit_of_measure, r.is_consumable, r.has_expiry, r.shelf_life_days,
		       r.is_critical, r.is_active, r.created_at, r.updated_at,
		       rc.name as category_name
		FROM resources r
		LEFT JOIN resource_categories rc ON r.category_id = rc.id
		WHERE r.is_active = true
		ORDER BY r.name
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return []*models.Resource{}, nil
	}
	defer rows.Close()

	var resources []*models.Resource
	for rows.Next() {
		resource := &models.Resource{}
		var categoryName sql.NullString

		err := rows.Scan(
			&resource.ID, &resource.Name, &resource.Description, &resource.ResourceCode,
			&resource.CategoryID, &resource.UnitOfMeasure, &resource.IsConsumable,
			&resource.HasExpiry, &resource.ShelfLifeDays, &resource.IsCritical,
			&resource.IsActive, &resource.CreatedAt, &resource.UpdatedAt,
			&categoryName,
		)
		if err != nil {
			continue
		}

		if categoryName.Valid {
			resource.Category = &models.ResourceCategory{
				ID:   resource.CategoryID,
				Name: categoryName.String,
			}
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

func (h *ResourceManagementHandler) getAllResourceCategories() ([]*models.ResourceCategory, error) {
	query := `SELECT id, name, description FROM resource_categories ORDER BY name`

	rows, err := h.db.Query(query)
	if err != nil {
		return []*models.ResourceCategory{}, nil
	}
	defer rows.Close()

	var categories []*models.ResourceCategory
	for rows.Next() {
		category := &models.ResourceCategory{}
		err := rows.Scan(&category.ID, &category.Name, &category.Description)
		if err != nil {
			continue
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (h *ResourceManagementHandler) getResourceByID(id string) (*models.Resource, error) {
	query := `
		SELECT r.id, r.name, r.description, r.resource_code, r.category_id,
		       r.unit_of_measure, r.is_consumable, r.has_expiry, r.shelf_life_days,
		       r.is_critical, r.is_active, r.created_at, r.updated_at
		FROM resources r
		WHERE r.id = $1
	`

	resource := &models.Resource{}
	err := h.db.QueryRow(query, id).Scan(
		&resource.ID, &resource.Name, &resource.Description, &resource.ResourceCode,
		&resource.CategoryID, &resource.UnitOfMeasure, &resource.IsConsumable,
		&resource.HasExpiry, &resource.ShelfLifeDays, &resource.IsCritical,
		&resource.IsActive, &resource.CreatedAt, &resource.UpdatedAt,
	)
	if err != nil {
		return &models.Resource{}, nil
	}

	return resource, nil
}

func (h *ResourceManagementHandler) createResource(resource *models.Resource) error {
	query := `
		INSERT INTO resources (
			name, description, resource_code, category_id, unit_of_measure,
			is_consumable, has_expiry, shelf_life_days, is_critical, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`

	return h.db.QueryRow(query,
		resource.Name, resource.Description, resource.ResourceCode, resource.CategoryID,
		resource.UnitOfMeasure, resource.IsConsumable, resource.HasExpiry,
		resource.ShelfLifeDays, resource.IsCritical, resource.IsActive,
	).Scan(&resource.ID)
}

func (h *ResourceManagementHandler) updateResource(resource *models.Resource) error {
	query := `
		UPDATE resources SET
			name = $1, description = $2, resource_code = $3, category_id = $4,
			unit_of_measure = $5, is_consumable = $6, has_expiry = $7,
			shelf_life_days = $8, is_critical = $9, is_active = $10,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $11
	`

	_, err := h.db.Exec(query,
		resource.Name, resource.Description, resource.ResourceCode, resource.CategoryID,
		resource.UnitOfMeasure, resource.IsConsumable, resource.HasExpiry,
		resource.ShelfLifeDays, resource.IsCritical, resource.IsActive,
		resource.ID,
	)

	return err
}

// Requisition database methods
func (h *ResourceManagementHandler) getAllRequisitions() ([]*models.Requisition, error) {
	query := `
		SELECT r.id, r.requisition_number, r.outbreak_id, r.deployment_id,
		       r.requested_by, r.required_date, r.priority, r.status, r.notes,
		       r.created_at, r.updated_at
		FROM requisitions r
		ORDER BY r.created_at DESC
		LIMIT 100
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return []*models.Requisition{}, nil
	}
	defer rows.Close()

	var requisitions []*models.Requisition
	for rows.Next() {
		requisition := &models.Requisition{}
		err := rows.Scan(
			&requisition.ID, &requisition.RequisitionNumber, &requisition.OutbreakID,
			&requisition.DeploymentID, &requisition.RequestedBy, &requisition.RequiredDate,
			&requisition.Priority, &requisition.Status, &requisition.Notes,
			&requisition.CreatedAt, &requisition.UpdatedAt,
		)
		if err != nil {
			continue
		}
		requisitions = append(requisitions, requisition)
	}

	return requisitions, nil
}

func (h *ResourceManagementHandler) getRequisitionsByOutbreak(outbreakID string) ([]*models.Requisition, error) {
	query := `
		SELECT r.id, r.requisition_number, r.outbreak_id, r.deployment_id,
		       r.requested_by, r.required_date, r.priority, r.status, r.notes,
		       r.created_at, r.updated_at
		FROM requisitions r
		WHERE r.outbreak_id = $1
		ORDER BY r.created_at DESC
	`

	rows, err := h.db.Query(query, outbreakID)
	if err != nil {
		return []*models.Requisition{}, nil
	}
	defer rows.Close()

	var requisitions []*models.Requisition
	for rows.Next() {
		requisition := &models.Requisition{}
		err := rows.Scan(
			&requisition.ID, &requisition.RequisitionNumber, &requisition.OutbreakID,
			&requisition.DeploymentID, &requisition.RequestedBy, &requisition.RequiredDate,
			&requisition.Priority, &requisition.Status, &requisition.Notes,
			&requisition.CreatedAt, &requisition.UpdatedAt,
		)
		if err != nil {
			continue
		}
		requisitions = append(requisitions, requisition)
	}

	return requisitions, nil
}

func (h *ResourceManagementHandler) getRequisitionByID(id string) (*models.Requisition, error) {
	query := `
		SELECT r.id, r.requisition_number, r.outbreak_id, r.deployment_id,
		       r.requested_by, r.required_date, r.priority, r.status, r.notes,
		       r.created_at, r.updated_at
		FROM requisitions r
		WHERE r.id = $1
	`

	requisition := &models.Requisition{}
	err := h.db.QueryRow(query, id).Scan(
		&requisition.ID, &requisition.RequisitionNumber, &requisition.OutbreakID,
		&requisition.DeploymentID, &requisition.RequestedBy, &requisition.RequiredDate,
		&requisition.Priority, &requisition.Status, &requisition.Notes,
		&requisition.CreatedAt, &requisition.UpdatedAt,
	)
	if err != nil {
		return &models.Requisition{}, nil
	}

	return requisition, nil
}

func (h *ResourceManagementHandler) createRequisition(requisition *models.Requisition) error {
	query := `
		INSERT INTO requisitions (
			requisition_number, outbreak_id, deployment_id, requested_by,
			required_date, priority, status, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	return h.db.QueryRow(query,
		requisition.RequisitionNumber, requisition.OutbreakID, requisition.DeploymentID,
		requisition.RequestedBy, requisition.RequiredDate, requisition.Priority,
		requisition.Status, requisition.Notes,
	).Scan(&requisition.ID)
}

func (h *ResourceManagementHandler) updateRequisition(requisition *models.Requisition) error {
	query := `
		UPDATE requisitions SET
			requisition_number = $1, outbreak_id = $2, deployment_id = $3,
			requested_by = $4, required_date = $5, priority = $6,
			status = $7, notes = $8, updated_at = CURRENT_TIMESTAMP
		WHERE id = $9
	`

	_, err := h.db.Exec(query,
		requisition.RequisitionNumber, requisition.OutbreakID, requisition.DeploymentID,
		requisition.RequestedBy, requisition.RequiredDate, requisition.Priority,
		requisition.Status, requisition.Notes, requisition.ID,
	)

	return err
}

// Activity Log database methods
func (h *ResourceManagementHandler) getAllActivityLogs() ([]*models.ActivityLog, error) {
	query := `
		SELECT a.id, a.deployment_id, a.activity_type, a.activity_date,
		       a.start_time, a.end_time, a.location, a.participants_count,
		       a.activity_description, a.outcomes, a.challenges, a.recommendations,
		       a.resources_used, a.created_at, a.created_by
		FROM activity_logs a
		ORDER BY a.activity_date DESC
		LIMIT 100
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return []*models.ActivityLog{}, nil
	}
	defer rows.Close()

	var activities []*models.ActivityLog
	for rows.Next() {
		activity := &models.ActivityLog{}
		err := rows.Scan(
			&activity.ID, &activity.DeploymentID, &activity.ActivityType,
			&activity.ActivityDate, &activity.StartTime, &activity.EndTime,
			&activity.Location, &activity.ParticipantsCount, &activity.ActivityDescription,
			&activity.Outcomes, &activity.Challenges, &activity.Recommendations,
			&activity.ResourcesUsed, &activity.CreatedAt, &activity.CreatedBy,
		)
		if err != nil {
			continue
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

func (h *ResourceManagementHandler) getActivitiesByDeployment(deploymentID string) ([]*models.ActivityLog, error) {
	query := `
		SELECT a.id, a.deployment_id, a.activity_type, a.activity_date,
		       a.start_time, a.end_time, a.location, a.participants_count,
		       a.activity_description, a.outcomes, a.challenges, a.recommendations,
		       a.resources_used, a.created_at, a.created_by
		FROM activity_logs a
		WHERE a.deployment_id = $1
		ORDER BY a.activity_date DESC
	`

	rows, err := h.db.Query(query, deploymentID)
	if err != nil {
		return []*models.ActivityLog{}, nil
	}
	defer rows.Close()

	var activities []*models.ActivityLog
	for rows.Next() {
		activity := &models.ActivityLog{}
		err := rows.Scan(
			&activity.ID, &activity.DeploymentID, &activity.ActivityType,
			&activity.ActivityDate, &activity.StartTime, &activity.EndTime,
			&activity.Location, &activity.ParticipantsCount, &activity.ActivityDescription,
			&activity.Outcomes, &activity.Challenges, &activity.Recommendations,
			&activity.ResourcesUsed, &activity.CreatedAt, &activity.CreatedBy,
		)
		if err != nil {
			continue
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

func (h *ResourceManagementHandler) getActivityLogByID(id string) (*models.ActivityLog, error) {
	query := `
		SELECT a.id, a.deployment_id, a.activity_type, a.activity_date,
		       a.start_time, a.end_time, a.location, a.participants_count,
		       a.activity_description, a.outcomes, a.challenges, a.recommendations,
		       a.resources_used, a.created_at, a.created_by
		FROM activity_logs a
		WHERE a.id = $1
	`

	activity := &models.ActivityLog{}
	err := h.db.QueryRow(query, id).Scan(
		&activity.ID, &activity.DeploymentID, &activity.ActivityType,
		&activity.ActivityDate, &activity.StartTime, &activity.EndTime,
		&activity.Location, &activity.ParticipantsCount, &activity.ActivityDescription,
		&activity.Outcomes, &activity.Challenges, &activity.Recommendations,
		&activity.ResourcesUsed, &activity.CreatedAt, &activity.CreatedBy,
	)
	if err != nil {
		return &models.ActivityLog{}, nil
	}

	return activity, nil
}

func (h *ResourceManagementHandler) createActivityLog(activity *models.ActivityLog) error {
	query := `
		INSERT INTO activity_logs (
			deployment_id, activity_type, activity_date, start_time, end_time,
			location, participants_count, activity_description, outcomes,
			challenges, recommendations, resources_used
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`

	return h.db.QueryRow(query,
		activity.DeploymentID, activity.ActivityType, activity.ActivityDate,
		activity.StartTime, activity.EndTime, activity.Location,
		activity.ParticipantsCount, activity.ActivityDescription, activity.Outcomes,
		activity.Challenges, activity.Recommendations, activity.ResourcesUsed,
	).Scan(&activity.ID)
}

func (h *ResourceManagementHandler) updateActivityLog(activity *models.ActivityLog) error {
	query := `
		UPDATE activity_logs SET
			deployment_id = $1, activity_type = $2, activity_date = $3,
			start_time = $4, end_time = $5, location = $6,
			participants_count = $7, activity_description = $8, outcomes = $9,
			challenges = $10, recommendations = $11, resources_used = $12
		WHERE id = $13
	`

	_, err := h.db.Exec(query,
		activity.DeploymentID, activity.ActivityType, activity.ActivityDate,
		activity.StartTime, activity.EndTime, activity.Location,
		activity.ParticipantsCount, activity.ActivityDescription, activity.Outcomes,
		activity.Challenges, activity.Recommendations, activity.ResourcesUsed,
		activity.ID,
	)

	return err
}

func (h *ResourceManagementHandler) getAllOutbreaks() ([]*models.Outbreak, error) {
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

func (h *ResourceManagementHandler) deleteRRTTeam(id int64) error {
	_, err := h.db.Exec(`DELETE FROM rrt_teams WHERE id = $1`, id)
	return err
}

func (h *ResourceManagementHandler) deleteRRTDeployment(id int64) error {
	_, err := h.db.Exec(`DELETE FROM rrt_deployments WHERE id = $1`, id)
	return err
}

func (h *ResourceManagementHandler) deleteResource(id int64) error {
	_, err := h.db.Exec(`DELETE FROM resources WHERE id = $1`, id)
	return err
}

func (h *ResourceManagementHandler) deleteRequisition(id int64) error {
	_, err := h.db.Exec(`DELETE FROM requisitions WHERE id = $1`, id)
	return err
}

func (h *ResourceManagementHandler) deleteActivityLog(id int64) error {
	_, err := h.db.Exec(`DELETE FROM activity_logs WHERE id = $1`, id)
	return err
}

func (h *ResourceManagementHandler) generateSitRep(outbreakID int64, reportType string, periodStart, periodEnd time.Time) (*models.GeneratedSitRep, error) {
	return &models.GeneratedSitRep{}, nil
}

func (h *ResourceManagementHandler) saveGeneratedSitRep(sitrep *models.GeneratedSitRep) error {
	return nil
}
