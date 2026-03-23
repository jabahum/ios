package handlers

import (
	"database/sql"
	"log/slog"
	"strconv"
	"time"

	"case/internal/middleware"
	"case/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func rmCurrentUID(c *fiber.Ctx) (int, bool) {
	return middleware.GetCurrentUserID(c)
}

func parseDateFlexible(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, sql.ErrNoRows
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02", s)
}

func parseNullDateFlexible(s string) sql.NullTime {
	if s == "" {
		return sql.NullTime{}
	}
	t, err := parseDateFlexible(s)
	if err != nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t, Valid: true}
}

// --- RRT teams ---

func HandlerResourceManagementRRTTeamGetAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	if _, ok := rmCurrentUID(c); !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	id := c.Params("id")
	h := NewResourceManagementHandler(db, store)
	team, err := h.getRRTTeamByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "Not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"team": team})
}

func HandlerResourceManagementRRTTeamCreateAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	uid, ok := rmCurrentUID(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	var team models.RRTTeam
	if err := c.BodyParser(&team); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON body"})
	}
	if team.TeamName == "" || team.TeamCode == "" {
		return c.Status(400).JSON(fiber.Map{"error": "team_name and team_code are required"})
	}
	team.CreatedBy = sql.NullInt64{Int64: int64(uid), Valid: true}
	h := NewResourceManagementHandler(db, store)
	if err := h.createRRTTeam(&team); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"message": "created", "id": team.ID, "team": team})
}

func HandlerResourceManagementRRTTeamUpdateAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	if _, ok := rmCurrentUID(c); !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid id"})
	}
	var team models.RRTTeam
	if err := c.BodyParser(&team); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON body"})
	}
	team.ID = id
	h := NewResourceManagementHandler(db, store)
	if err := h.updateRRTTeam(&team); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "updated", "team": team})
}

func HandlerResourceManagementRRTTeamDeleteAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	if _, ok := rmCurrentUID(c); !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid id"})
	}
	h := NewResourceManagementHandler(db, store)
	if err := h.deleteRRTTeam(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "deleted"})
}

// --- RRT deployments ---

func HandlerResourceManagementRRTDeploymentGetAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	if _, ok := rmCurrentUID(c); !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	h := NewResourceManagementHandler(db, store)
	d, err := h.getRRTDeploymentByID(c.Params("id"))
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "Not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"deployment": d})
}

func HandlerResourceManagementRRTDeploymentCreateAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	if _, ok := rmCurrentUID(c); !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	var body struct {
		TeamID             int64  `json:"team_id"`
		OutbreakID         int64  `json:"outbreak_id"`
		DeploymentDate     string `json:"deployment_date"`
		ExpectedReturnDate string `json:"expected_return_date"`
		ActualReturnDate   string `json:"actual_return_date"`
		DeploymentStatus   string `json:"deployment_status"`
		DeploymentPurpose  string `json:"deployment_purpose"`
		AssignedVehicle    string `json:"assigned_vehicle"`
		AssignedDriver     string `json:"assigned_driver"`
		DeploymentNotes    string `json:"deployment_notes"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON body"})
	}
	d := models.RRTDeployment{
		TeamID:            body.TeamID,
		OutbreakID:        body.OutbreakID,
		DeploymentStatus:  body.DeploymentStatus,
		DeploymentPurpose: sql.NullString{String: body.DeploymentPurpose, Valid: body.DeploymentPurpose != ""},
		AssignedVehicle:   sql.NullString{String: body.AssignedVehicle, Valid: body.AssignedVehicle != ""},
		AssignedDriver:    sql.NullString{String: body.AssignedDriver, Valid: body.AssignedDriver != ""},
		DeploymentNotes:   sql.NullString{String: body.DeploymentNotes, Valid: body.DeploymentNotes != ""},
	}
	var err error
	d.DeploymentDate, err = parseDateFlexible(body.DeploymentDate)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "deployment_date required (YYYY-MM-DD or RFC3339)"})
	}
	if body.ExpectedReturnDate != "" {
		d.ExpectedReturnDate = parseNullDateFlexible(body.ExpectedReturnDate)
	}
	if body.ActualReturnDate != "" {
		d.ActualReturnDate = parseNullDateFlexible(body.ActualReturnDate)
	}
	if d.TeamID == 0 || d.OutbreakID == 0 || d.DeploymentStatus == "" {
		return c.Status(400).JSON(fiber.Map{"error": "team_id, outbreak_id, deployment_status are required"})
	}
	h := NewResourceManagementHandler(db, store)
	if err := h.createRRTDeployment(&d); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"message": "created", "id": d.ID, "deployment": d})
}

func HandlerResourceManagementRRTDeploymentUpdateAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	if _, ok := rmCurrentUID(c); !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid id"})
	}
	h := NewResourceManagementHandler(db, store)
	existing, err := h.getRRTDeploymentByID(c.Params("id"))
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "Not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	var body struct {
		TeamID             int64  `json:"team_id"`
		OutbreakID         int64  `json:"outbreak_id"`
		DeploymentDate     string `json:"deployment_date"`
		ExpectedReturnDate string `json:"expected_return_date"`
		ActualReturnDate   string `json:"actual_return_date"`
		DeploymentStatus   string `json:"deployment_status"`
		DeploymentPurpose  string `json:"deployment_purpose"`
		AssignedVehicle    string `json:"assigned_vehicle"`
		AssignedDriver     string `json:"assigned_driver"`
		DeploymentNotes    string `json:"deployment_notes"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON body"})
	}
	d := *existing
	d.ID = id
	if body.TeamID != 0 {
		d.TeamID = body.TeamID
	}
	if body.OutbreakID != 0 {
		d.OutbreakID = body.OutbreakID
	}
	if body.DeploymentStatus != "" {
		d.DeploymentStatus = body.DeploymentStatus
	}
	d.DeploymentDate = existing.DeploymentDate
	if body.DeploymentDate != "" {
		d.DeploymentDate, err = parseDateFlexible(body.DeploymentDate)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid deployment_date"})
		}
	}
	d.ExpectedReturnDate = existing.ExpectedReturnDate
	if body.ExpectedReturnDate != "" {
		d.ExpectedReturnDate = parseNullDateFlexible(body.ExpectedReturnDate)
	}
	d.ActualReturnDate = existing.ActualReturnDate
	if body.ActualReturnDate != "" {
		d.ActualReturnDate = parseNullDateFlexible(body.ActualReturnDate)
	}
	if body.DeploymentPurpose != "" {
		d.DeploymentPurpose = sql.NullString{String: body.DeploymentPurpose, Valid: true}
	}
	if body.AssignedVehicle != "" {
		d.AssignedVehicle = sql.NullString{String: body.AssignedVehicle, Valid: true}
	}
	if body.AssignedDriver != "" {
		d.AssignedDriver = sql.NullString{String: body.AssignedDriver, Valid: true}
	}
	if body.DeploymentNotes != "" {
		d.DeploymentNotes = sql.NullString{String: body.DeploymentNotes, Valid: true}
	}
	if err := h.updateRRTDeployment(&d); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "updated", "deployment": d})
}

func HandlerResourceManagementRRTDeploymentDeleteAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	if _, ok := rmCurrentUID(c); !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid id"})
	}
	h := NewResourceManagementHandler(db, store)
	if err := h.deleteRRTDeployment(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "deleted"})
}

// --- Resources (catalog items) ---

func HandlerResourceManagementResourceGetAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	if _, ok := rmCurrentUID(c); !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	h := NewResourceManagementHandler(db, store)
	r, err := h.getResourceByID(c.Params("id"))
	if err != nil || r.ID == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Not found"})
	}
	return c.JSON(fiber.Map{"resource": r})
}

func HandlerResourceManagementResourceCreateAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	if _, ok := rmCurrentUID(c); !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	var r models.Resource
	if err := c.BodyParser(&r); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON body"})
	}
	if r.Name == "" || r.CategoryID == 0 || r.UnitOfMeasure == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name, category_id, unit_of_measure are required"})
	}
	h := NewResourceManagementHandler(db, store)
	if err := h.createResource(&r); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"message": "created", "id": r.ID, "resource": r})
}

func HandlerResourceManagementResourceUpdateAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	if _, ok := rmCurrentUID(c); !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid id"})
	}
	var r models.Resource
	if err := c.BodyParser(&r); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON body"})
	}
	r.ID = id
	h := NewResourceManagementHandler(db, store)
	if err := h.updateResource(&r); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "updated", "resource": r})
}

func HandlerResourceManagementResourceDeleteAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	if _, ok := rmCurrentUID(c); !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid id"})
	}
	h := NewResourceManagementHandler(db, store)
	if err := h.deleteResource(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "deleted"})
}

// --- Requisitions ---

func HandlerResourceManagementRequisitionGetAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	if _, ok := rmCurrentUID(c); !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	h := NewResourceManagementHandler(db, store)
	r, err := h.getRequisitionByID(c.Params("id"))
	if err != nil || r.ID == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Not found"})
	}
	return c.JSON(fiber.Map{"requisition": r})
}

func HandlerResourceManagementRequisitionCreateAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	uid, ok := rmCurrentUID(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	var body struct {
		RequisitionNumber string `json:"requisition_number"`
		OutbreakID        int64  `json:"outbreak_id"`
		DeploymentID      int64  `json:"deployment_id"`
		RequestedBy       int64  `json:"requested_by"`
		RequiredDate      string `json:"required_date"`
		Priority          string `json:"priority"`
		Status            string `json:"status"`
		Notes             string `json:"notes"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON body"})
	}
	q := models.Requisition{
		RequisitionNumber: body.RequisitionNumber,
		OutbreakID:        body.OutbreakID,
		Notes:             sql.NullString{String: body.Notes, Valid: body.Notes != ""},
	}
	if body.DeploymentID != 0 {
		q.DeploymentID = sql.NullInt64{Int64: body.DeploymentID, Valid: true}
	}
	if q.RequisitionNumber == "" || q.OutbreakID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "requisition_number and outbreak_id are required"})
	}
	if body.RequestedBy != 0 {
		q.RequestedBy = body.RequestedBy
	} else {
		q.RequestedBy = int64(uid)
	}
	q.RequestedDate = time.Now()
	if body.RequiredDate != "" {
		if t, err := parseDateFlexible(body.RequiredDate); err == nil {
			q.RequiredDate = sql.NullTime{Time: t, Valid: true}
		}
	}
	if body.Priority != "" {
		q.Priority = body.Priority
	} else {
		q.Priority = "normal"
	}
	if body.Status != "" {
		q.Status = body.Status
	} else {
		q.Status = "pending"
	}
	h := NewResourceManagementHandler(db, store)
	if err := h.createRequisition(&q); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"message": "created", "id": q.ID, "requisition": q})
}

func HandlerResourceManagementRequisitionUpdateAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	if _, ok := rmCurrentUID(c); !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid id"})
	}
	h := NewResourceManagementHandler(db, store)
	existing, err := h.getRequisitionByID(c.Params("id"))
	if err != nil || existing.ID == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Not found"})
	}
	var body struct {
		RequisitionNumber string `json:"requisition_number"`
		OutbreakID        int64  `json:"outbreak_id"`
		DeploymentID      int64  `json:"deployment_id"`
		RequestedBy       int64  `json:"requested_by"`
		RequiredDate      string `json:"required_date"`
		Priority          string `json:"priority"`
		Status            string `json:"status"`
		Notes             string `json:"notes"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON body"})
	}
	q := *existing
	q.ID = id
	if body.RequisitionNumber != "" {
		q.RequisitionNumber = body.RequisitionNumber
	}
	if body.OutbreakID != 0 {
		q.OutbreakID = body.OutbreakID
	}
	if body.DeploymentID != 0 {
		q.DeploymentID = sql.NullInt64{Int64: body.DeploymentID, Valid: true}
	}
	if body.RequestedBy != 0 {
		q.RequestedBy = body.RequestedBy
	}
	if body.RequiredDate != "" {
		if t, err := parseDateFlexible(body.RequiredDate); err == nil {
			q.RequiredDate = sql.NullTime{Time: t, Valid: true}
		}
	}
	if body.Priority != "" {
		q.Priority = body.Priority
	}
	if body.Status != "" {
		q.Status = body.Status
	}
	if body.Notes != "" {
		q.Notes = sql.NullString{String: body.Notes, Valid: true}
	}
	if err := h.updateRequisition(&q); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "updated", "requisition": q})
}

func HandlerResourceManagementRequisitionDeleteAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	if _, ok := rmCurrentUID(c); !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid id"})
	}
	h := NewResourceManagementHandler(db, store)
	if err := h.deleteRequisition(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "deleted"})
}

// --- Activity logs ---

func HandlerResourceManagementActivityLogGetAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	if _, ok := rmCurrentUID(c); !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	h := NewResourceManagementHandler(db, store)
	a, err := h.getActivityLogByID(c.Params("id"))
	if err != nil || a.ID == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Not found"})
	}
	return c.JSON(fiber.Map{"activity_log": a})
}

func HandlerResourceManagementActivityLogCreateAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	if _, ok := rmCurrentUID(c); !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	var body struct {
		models.ActivityLog
		ActivityDate string `json:"activity_date"`
		StartTime    string `json:"start_time"`
		EndTime      string `json:"end_time"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON body"})
	}
	a := body.ActivityLog
	var err error
	a.ActivityDate, err = parseDateFlexible(body.ActivityDate)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "activity_date required"})
	}
	if body.StartTime != "" {
		t, e := time.Parse("15:04", body.StartTime)
		if e == nil {
			a.StartTime = sql.NullTime{Time: t, Valid: true}
		}
	}
	if body.EndTime != "" {
		t, e := time.Parse("15:04", body.EndTime)
		if e == nil {
			a.EndTime = sql.NullTime{Time: t, Valid: true}
		}
	}
	if a.DeploymentID == 0 || a.ActivityType == "" {
		return c.Status(400).JSON(fiber.Map{"error": "deployment_id and activity_type are required"})
	}
	h := NewResourceManagementHandler(db, store)
	if err := h.createActivityLog(&a); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"message": "created", "id": a.ID, "activity_log": a})
}

func HandlerResourceManagementActivityLogUpdateAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	if _, ok := rmCurrentUID(c); !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid id"})
	}
	h := NewResourceManagementHandler(db, store)
	existing, err := h.getActivityLogByID(c.Params("id"))
	if err != nil || existing.ID == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Not found"})
	}
	var body struct {
		DeploymentID        int64  `json:"deployment_id"`
		ActivityType        string `json:"activity_type"`
		ActivityDate        string `json:"activity_date"`
		StartTime           string `json:"start_time"`
		EndTime             string `json:"end_time"`
		Location            string `json:"location"`
		ParticipantsCount   int64  `json:"participants_count"`
		ActivityDescription string `json:"activity_description"`
		Outcomes            string `json:"outcomes"`
		Challenges          string `json:"challenges"`
		Recommendations     string `json:"recommendations"`
		ResourcesUsed       string `json:"resources_used"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON body"})
	}
	a := *existing
	a.ID = id
	if body.DeploymentID != 0 {
		a.DeploymentID = body.DeploymentID
	}
	if body.ActivityType != "" {
		a.ActivityType = body.ActivityType
	}
	if body.ActivityDate != "" {
		a.ActivityDate, err = parseDateFlexible(body.ActivityDate)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid activity_date"})
		}
	}
	if body.StartTime != "" {
		t, e := time.Parse("15:04", body.StartTime)
		if e == nil {
			a.StartTime = sql.NullTime{Time: t, Valid: true}
		}
	}
	if body.EndTime != "" {
		t, e := time.Parse("15:04", body.EndTime)
		if e == nil {
			a.EndTime = sql.NullTime{Time: t, Valid: true}
		}
	}
	if body.Location != "" {
		a.Location = sql.NullString{String: body.Location, Valid: true}
	}
	if body.ParticipantsCount > 0 {
		a.ParticipantsCount = sql.NullInt64{Int64: body.ParticipantsCount, Valid: true}
	}
	if body.ActivityDescription != "" {
		a.ActivityDescription = sql.NullString{String: body.ActivityDescription, Valid: true}
	}
	if body.Outcomes != "" {
		a.Outcomes = sql.NullString{String: body.Outcomes, Valid: true}
	}
	if body.Challenges != "" {
		a.Challenges = sql.NullString{String: body.Challenges, Valid: true}
	}
	if body.Recommendations != "" {
		a.Recommendations = sql.NullString{String: body.Recommendations, Valid: true}
	}
	if body.ResourcesUsed != "" {
		a.ResourcesUsed = sql.NullString{String: body.ResourcesUsed, Valid: true}
	}
	if err := h.updateActivityLog(&a); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "updated", "activity_log": a})
}

func HandlerResourceManagementActivityLogDeleteAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	if _, ok := rmCurrentUID(c); !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid id"})
	}
	h := NewResourceManagementHandler(db, store)
	if err := h.deleteActivityLog(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "deleted"})
}

// --- Pillars ---

func HandlerResourceManagementPillarGetAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	if _, ok := rmCurrentUID(c); !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	ph := NewPillarsHandler(db, store)
	p, err := ph.getPillarByID(c.Params("id"))
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "Not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"pillar": p})
}

func HandlerResourceManagementPillarCreateAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	uid, ok := rmCurrentUID(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	var p models.Pillar
	if err := c.BodyParser(&p); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON body"})
	}
	if p.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name is required"})
	}
	p.CreatedBy = sql.NullInt64{Int64: int64(uid), Valid: true}
	ph := NewPillarsHandler(db, store)
	if err := ph.createPillar(&p); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"message": "created", "id": p.ID, "pillar": p})
}

func HandlerResourceManagementPillarUpdateAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	uid, ok := rmCurrentUID(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid id"})
	}
	var p models.Pillar
	if err := c.BodyParser(&p); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON body"})
	}
	p.ID = id
	p.UpdatedBy = sql.NullInt64{Int64: int64(uid), Valid: true}
	ph := NewPillarsHandler(db, store)
	if err := ph.updatePillar(&p); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "updated", "pillar": p})
}

func HandlerResourceManagementPillarDeleteAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	if _, ok := rmCurrentUID(c); !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid id"})
	}
	ph := NewPillarsHandler(db, store)
	if err := ph.deletePillarByID(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "deleted"})
}

// HandlerOutbreakUsersAPI lists users assigned to an outbreak (active user_outbreaks).
func HandlerOutbreakUsersAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store) error {
	if _, ok := rmCurrentUID(c); !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	oid, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || oid <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid outbreak id"})
	}
	svc := models.NewUserOutbreakService(db)
	users, err := svc.ListAssignedUsersForOutbreak(oid)
	if err != nil {
		sl.Error("list outbreak users", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"outbreak_id": oid, "users": users})
}
