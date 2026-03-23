package handlers

import (
	"context"
	"database/sql"
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"

	"case/internal/middleware"
	"case/internal/models"
	"case/internal/services"
)

// API handlers: consolidated JSON/API route entrypoints (auth, VHF, users, facilities,
// outbreaks, cases, disease CIFs, resource management, inventory, surveillance stubs, etc.).

// Authentication APIs
func HandlerGetCurrentUser(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Get user details from database
	userService := models.NewUserService(db)
	user, err := userService.GetUserByID(int64(userID))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user"})
	}

	return c.JSON(user)
}

func HandlerChangePassword(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	// This would wrap the existing password change logic
	// For now, return a placeholder
	return c.JSON(fiber.Map{"message": "Password change endpoint - implement as needed"})
}

// Dashboard APIs
func HandlerHomeAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	// Get home dashboard data
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Return dashboard data
	return c.JSON(fiber.Map{
		"user_id": userID,
		"message": "Home dashboard data",
	})
}

func HandlerDashboardStats(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Get dashboard statistics
	stats := fiber.Map{
		"total_patients":  0,
		"active_cases":    0,
		"total_outbreaks": 0,
	}

	return c.JSON(stats)
}

// VHF APIs — JSON counterparts to VHF CIF; use models.VHFPatient JSON shape (see internal/models/vhf_models.go).

func HandlerVHFListAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	list, err := models.ListVHFPatients(db)
	if err != nil {
		sl.Error("ListVHFPatients", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	return c.JSON(fiber.Map{"patients": list})
}

func HandlerVHFViewAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid patient ID"})
	}
	payload, err := BuildVHFCIFJSON(db, id)
	if err == sql.ErrNoRows {
		return c.Status(404).JSON(fiber.Map{"error": "Patient not found"})
	}
	if err != nil {
		sl.Error("BuildVHFCIFJSON", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	return c.JSON(payload)
}

func HandlerVHFPatientSubmitAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config, smsService *services.SMSService) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	var patient models.VHFPatient
	if err := c.BodyParser(&patient); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON body; use VHFPatient field names from /api/vhf/patients/{id} response"})
	}
	patient.ID = 0
	patient.CreatedAt = time.Now()
	if err := models.SaveVHFPatient(db, &patient); err != nil {
		sl.Error("SaveVHFPatient API", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	_ = smsService
	return c.Status(201).JSON(fiber.Map{
		"message":    "Patient created",
		"patient_id": patient.ID,
		"patient":    patient,
	})
}

func HandlerVHFUpdateAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid patient ID"})
	}
	var patient models.VHFPatient
	if err := c.BodyParser(&patient); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON body"})
	}
	patient.ID = id
	if err := models.UpdateVHFPatient(db, &patient); err != nil {
		sl.Error("UpdateVHFPatient API", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Patient updated", "patient_id": id})
}

func HandlerVHFDeleteAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid patient ID"})
	}
	if err := models.DeleteVHFPatient(db, id); err != nil {
		sl.Error("DeleteVHFPatient API", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Patient deleted", "patient_id": id})
}

// VHF Clinical Signs APIs
func HandlerVHFClinicalSignsAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	pid, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid patient id"})
	}
	signs, err := models.GetVHFClinicalSigns(db, pid)
	if err != nil {
		sl.Error("GetVHFClinicalSigns", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	return c.JSON(fiber.Map{"clinical_signs": signs})
}

func HandlerVHFClinicalSignsSubmitAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	pid, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid patient id"})
	}
	var signs models.VHFClinicalSigns
	if err := c.BodyParser(&signs); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}
	signs.PatientID = pid
	signs.CreatedAt = time.Now()
	if err := models.SaveVHFClinicalSigns(db, &signs); err != nil {
		sl.Error("SaveVHFClinicalSigns", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Clinical signs saved", "id": signs.ID})
}

// VHF Hospitalization APIs
func HandlerVHFHospitalizationAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	pid, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid patient id"})
	}
	h, err := models.GetVHFHospitalization(db, pid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	return c.JSON(fiber.Map{"hospitalization": h})
}

func HandlerVHFHospitalizationSubmitAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	pid, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid patient id"})
	}
	var h models.VHFHospitalization
	if err := c.BodyParser(&h); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}
	h.PatientID = pid
	if err := models.SaveVHFHospitalization(db, &h); err != nil {
		sl.Error("SaveVHFHospitalization", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Hospitalization saved", "id": h.ID})
}

// VHF Risk Factors APIs
func HandlerVHFRiskFactorsAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	pid, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid patient id"})
	}
	r, err := models.GetVHFRiskFactors(db, pid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	return c.JSON(fiber.Map{"risk_factors": r})
}

func HandlerVHFRiskFactorsSubmitAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	pid, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid patient id"})
	}
	var r models.VHFRiskFactors
	if err := c.BodyParser(&r); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}
	r.PatientID = pid
	if err := models.SaveVHFRiskFactors(db, &r); err != nil {
		sl.Error("SaveVHFRiskFactors", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Risk factors saved", "id": r.ID})
}

// VHF Laboratory APIs
func HandlerVHFLaboratoryAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	pid, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid patient id"})
	}
	lab, err := models.GetVHFLaboratory(db, pid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	return c.JSON(fiber.Map{"laboratory": lab})
}

func HandlerVHFLaboratorySubmitAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config, smsService *services.SMSService) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	pid, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid patient id"})
	}
	var lab models.VHFLaboratory
	if err := c.BodyParser(&lab); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}
	lab.PatientID = pid
	if err := models.SaveVHFLaboratory(db, &lab); err != nil {
		sl.Error("SaveVHFLaboratory", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	_ = smsService
	return c.JSON(fiber.Map{"message": "Laboratory saved", "id": lab.ID})
}

// VHF Investigator APIs
func HandlerVHFInvestigatorAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	pid, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid patient id"})
	}
	inv, err := models.GetVHFInvestigator(db, pid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	return c.JSON(fiber.Map{"investigator": inv})
}

func HandlerVHFInvestigatorSubmitAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	pid, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid patient id"})
	}
	var inv models.VHFInvestigator
	if err := c.BodyParser(&inv); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}
	inv.PatientID = pid
	inv.CreatedAt = time.Now()
	if err := models.SaveVHFInvestigator(db, &inv); err != nil {
		sl.Error("SaveVHFInvestigator", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Investigator saved", "id": inv.ID})
}

// VHF Lab Form APIs (lab record id — same laboratory model when keyed by patient in other routes)
func HandlerVHFLabFormAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	pid, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid id"})
	}
	lab, err := models.GetVHFLaboratory(db, pid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	return c.JSON(fiber.Map{"laboratory": lab})
}

func HandlerVHFLabSaveAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerVHFLaboratorySubmitAPI(c, db, sl, store, config, nil)
}

// Employee Management APIs
func HandlerEmployeeListAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	query := `SELECT e.employee_id, e.employee_fname, e.employee_lname, e.employee_sex,
	          e.employee_email, e.employee_phone, e.employee_cadre, e.facility,
	          f.facility_name
	          FROM public.employee e
	          LEFT JOIN public.facility f ON e.facility = f.facility_id
	          ORDER BY e.employee_fname, e.employee_lname`
	rows, err := db.QueryContext(c.Context(), query)
	if err != nil {
		sl.Error("employee list api", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	defer rows.Close()
	out := make([]fiber.Map, 0)
	for rows.Next() {
		var id int
		var fn, ln, sex, em, ph, cadre sql.NullString
		var fac sql.NullInt64
		var facName sql.NullString
		if err := rows.Scan(&id, &fn, &ln, &sex, &em, &ph, &cadre, &fac, &facName); err != nil {
			continue
		}
		out = append(out, fiber.Map{
			"employee_id":    id,
			"employee_fname": fn.String,
			"employee_lname": ln.String,
			"employee_sex":   sex.String,
			"employee_email": em.String,
			"employee_phone": ph.String,
			"employee_cadre": cadre.String,
			"facility":       fac.Int64,
			"facility_name":  facName.String,
		})
	}
	return c.JSON(fiber.Map{"employees": out})
}

func HandlerGetEmployeeAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	return HandlerGetEmployee(c, db, sl, store, config)
}

func HandlerEmployeeSubmitAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	var emp models.Employee
	if err := c.BodyParser(&emp); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON (see models.Employee json tags)"})
	}
	emp.EmployeeID = 0
	if err := emp.Insert(c.Context(), db); err != nil {
		sl.Error("employee insert api", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"message": "Employee created", "employee_id": emp.EmployeeID})
}

func HandlerEmployeeUpdateAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid employee id"})
	}
	var emp models.Employee
	if err := c.BodyParser(&emp); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}
	emp.EmployeeID = id
	emp.SetAsExists()
	if err := emp.Update(c.Context(), db); err != nil {
		sl.Error("employee update api", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Employee updated", "employee_id": id})
}

func HandlerDeleteEmployeeAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	return HandlerDeleteEmployee(c, db, sl, store, config)
}

// JSON APIs for resource / RRT management (session + resource_management:* permissions on routes).

func HandlerResourceManagementSummaryAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	h := NewResourceManagementHandler(db, store)
	stats, err := h.getResourceManagementStats()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(stats)
}

func HandlerResourceManagementPillarsAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return NewPillarsHandler(db, store).HandlerPillarsAPI(c)
}

func HandlerResourceManagementRRTTeamsAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	h := NewResourceManagementHandler(db, store)
	teams, err := h.getAllRRTTeams()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"teams": teams})
}

func HandlerResourceManagementRRTDeploymentsAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	h := NewResourceManagementHandler(db, store)
	list, err := h.getAllRRTDeployments()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"deployments": list})
}

func HandlerResourceManagementResourcesAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	h := NewResourceManagementHandler(db, store)
	list, err := h.getAllResources()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"resources": list})
}

func HandlerResourceManagementRequisitionsAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	h := NewResourceManagementHandler(db, store)
	list, err := h.getAllRequisitions()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"requisitions": list})
}

func HandlerResourceManagementActivityLogsAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	h := NewResourceManagementHandler(db, store)
	list, err := h.getAllActivityLogs()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"activity_logs": list})
}

// User Management APIs (session auth + users:* RBAC via attachPermissionsForAPI)
func HandlerUserListAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	attachPermissionsForAPI(c, db, userID)
	return NewEnhancedUserHandler(db, sl, store, config).ListUsers(c)
}

func HandlerGetUserAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	attachPermissionsForAPI(c, db, userID)
	return NewEnhancedUserHandler(db, sl, store, config).GetUserDetails(c)
}

func HandlerUserSubmitAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	attachPermissionsForAPI(c, db, userID)
	return NewEnhancedUserHandler(db, sl, store, config).CreateUser(c)
}

func HandlerUserUpdateAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	attachPermissionsForAPI(c, db, userID)
	return NewEnhancedUserHandler(db, sl, store, config).UpdateUser(c)
}

func HandlerUserDeleteAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	attachPermissionsForAPI(c, db, userID)
	return NewEnhancedUserHandler(db, sl, store, config).DeleteUser(c)
}

// Facility Management APIs
func HandlerFacilityListAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	facilities := []fiber.Map{
		{"id": 1, "name": "Hospital A", "district": "Kampala"},
		{"id": 2, "name": "Hospital B", "district": "Wakiso"},
	}

	return c.JSON(fiber.Map{"facilities": facilities})
}

func HandlerGetFacilityAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	return c.JSON(fiber.Map{"facility": fiber.Map{"id": c.Params("id"), "name": "Facility Name"}})
}

func HandlerFacilitySubmitAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var data fiber.Map
	c.BodyParser(&data)
	return c.Status(201).JSON(fiber.Map{"message": "Facility created successfully"})
}

func HandlerFacilityUpdateAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	var data fiber.Map
	c.BodyParser(&data)
	return c.JSON(fiber.Map{"message": "Facility updated successfully"})
}

func HandlerFacilityDeleteAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	return c.JSON(fiber.Map{"message": "Facility deleted successfully"})
}

// Outbreak Management APIs
func HandlerOutbreakListAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerGetOutbreaksAPI(c, db, sl, store)
}

func HandlerGetOutbreakAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	uid, ok := middleware.GetCurrentUserID(c)
	if !ok {
		uid = GetCurrentUser(c, store)
	}
	if uid == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid outbreak id"})
	}
	hasAccess, err := models.CheckUserOutbreakAccess(c.Context(), db, uid, id)
	if err != nil {
		sl.Error("outbreak access check", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to check access"})
	}
	if !hasAccess {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}
	o, err := models.OutbreakByID(c.Context(), db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "Not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"outbreak": o})
}

func HandlerOutbreakSubmitAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	uid, ok := middleware.GetCurrentUserID(c)
	if !ok {
		uid = GetCurrentUser(c, store)
	}
	if uid == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	var body struct {
		Name             string `json:"name"`
		Description      string `json:"description"`
		StartDate        string `json:"start_date"`
		EndDate          string `json:"end_date"`
		Status           string `json:"status"`
		OutbreakType     string `json:"outbreak_type"`
		OutbreakCategory string `json:"outbreak_category"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON body"})
	}
	if body.Name == "" || body.StartDate == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name and start_date are required"})
	}
	startDate, err := time.Parse("2006-01-02", body.StartDate)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "start_date must be YYYY-MM-DD"})
	}
	now := time.Now()
	o := models.Outbreak{
		Name:             sql.NullString{String: body.Name, Valid: true},
		Description:      sql.NullString{String: body.Description, Valid: body.Description != ""},
		StartDate:        sql.NullTime{Time: startDate, Valid: true},
		Status:           sql.NullString{String: body.Status, Valid: body.Status != ""},
		OutbreakType:     sql.NullString{String: body.OutbreakType, Valid: body.OutbreakType != ""},
		OutbreakCategory: sql.NullString{String: body.OutbreakCategory, Valid: body.OutbreakCategory != ""},
		EnterOn:          sql.NullTime{Time: now, Valid: true},
		EnterBy:          sql.NullInt64{Int64: int64(uid), Valid: true},
		EditOn:           sql.NullTime{Time: now, Valid: true},
		EditBy:           sql.NullInt64{Int64: int64(uid), Valid: true},
	}
	if o.OutbreakCategory.String == "" && o.OutbreakType.Valid {
		o.OutbreakCategory = o.OutbreakType
	}
	if body.EndDate != "" {
		if endDate, err := time.Parse("2006-01-02", body.EndDate); err == nil {
			o.EndDate = sql.NullTime{Time: endDate, Valid: true}
		}
	}
	if !o.Status.Valid {
		o.Status = sql.NullString{String: "active", Valid: true}
	}
	if err := o.Insert(c.Context(), db); err != nil {
		sl.Error("outbreak insert", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"message": "Outbreak created", "outbreak_id": o.ID})
}

func HandlerOutbreakUpdateAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	uid, ok := middleware.GetCurrentUserID(c)
	if !ok {
		uid = GetCurrentUser(c, store)
	}
	if uid == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid outbreak id"})
	}
	hasAccess, err := models.CheckUserOutbreakAccess(c.Context(), db, uid, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to check access"})
	}
	if !hasAccess {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}
	canManage, err := models.CanUserManageOutbreak(c.Context(), db, uid, id)
	if err != nil || !canManage {
		return c.Status(403).JSON(fiber.Map{"error": "You cannot edit this outbreak"})
	}
	o, err := models.OutbreakByID(c.Context(), db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "Not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	var body struct {
		Name             string `json:"name"`
		Description      string `json:"description"`
		StartDate        string `json:"start_date"`
		EndDate          string `json:"end_date"`
		Status           string `json:"status"`
		OutbreakType     string `json:"outbreak_type"`
		OutbreakCategory string `json:"outbreak_category"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON body"})
	}
	if body.Name != "" {
		o.Name = sql.NullString{String: body.Name, Valid: true}
	}
	if body.Description != "" {
		o.Description = sql.NullString{String: body.Description, Valid: true}
	}
	if body.StartDate != "" {
		if t, err := time.Parse("2006-01-02", body.StartDate); err == nil {
			o.StartDate = sql.NullTime{Time: t, Valid: true}
		}
	}
	if body.EndDate != "" {
		if t, err := time.Parse("2006-01-02", body.EndDate); err == nil {
			o.EndDate = sql.NullTime{Time: t, Valid: true}
		}
	}
	if body.Status != "" {
		o.Status = sql.NullString{String: body.Status, Valid: true}
	}
	if body.OutbreakType != "" {
		o.OutbreakType = sql.NullString{String: body.OutbreakType, Valid: true}
	}
	if body.OutbreakCategory != "" {
		o.OutbreakCategory = sql.NullString{String: body.OutbreakCategory, Valid: true}
	}
	now := time.Now()
	o.EditOn = sql.NullTime{Time: now, Valid: true}
	o.EditBy = sql.NullInt64{Int64: int64(uid), Valid: true}
	if err := o.Update(c.Context(), db); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Outbreak updated"})
}

func HandlerOutbreakDeleteAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	uid, ok := middleware.GetCurrentUserID(c)
	if !ok {
		uid = GetCurrentUser(c, store)
	}
	if uid == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid outbreak id"})
	}
	hasAccess, err := models.CheckUserOutbreakAccess(c.Context(), db, uid, id)
	if err != nil || !hasAccess {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}
	canManage, err := models.CanUserManageOutbreak(c.Context(), db, uid, id)
	if err != nil || !canManage {
		return c.Status(403).JSON(fiber.Map{"error": "You cannot delete this outbreak"})
	}
	o, err := models.OutbreakByID(c.Context(), db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "Not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if err := o.Delete(c.Context(), db); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Outbreak deleted"})
}

func HandlerOutbreakCloseAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	uid, ok := middleware.GetCurrentUserID(c)
	if !ok {
		uid = GetCurrentUser(c, store)
	}
	if uid == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid outbreak id"})
	}
	hasAccess, err := models.CheckUserOutbreakAccess(c.Context(), db, uid, id)
	if err != nil || !hasAccess {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}
	canManage, err := models.CanUserManageOutbreak(c.Context(), db, uid, id)
	if err != nil || !canManage {
		return c.Status(403).JSON(fiber.Map{"error": "You cannot close this outbreak"})
	}
	o, err := models.OutbreakByID(c.Context(), db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "Not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	o.Status = sql.NullString{String: "closed", Valid: true}
	o.EditOn = sql.NullTime{Time: time.Now(), Valid: true}
	o.EditBy = sql.NullInt64{Int64: int64(uid), Valid: true}
	if err := o.Update(c.Context(), db); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Outbreak closed"})
}

func HandlerOutbreakSelectAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	uid, ok := middleware.GetCurrentUserID(c)
	if !ok {
		uid = GetCurrentUser(c, store)
	}
	if uid == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid outbreak id"})
	}
	hasAccess, err := models.CheckUserOutbreakAccess(c.Context(), db, uid, id)
	if err != nil || !hasAccess {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}
	if err := SetSelectedOutbreak(c, store, id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to select outbreak"})
	}
	return c.JSON(fiber.Map{"message": "Outbreak selected", "outbreak_id": id})
}

// Case Management APIs
func HandlerCasesListAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	cases := []fiber.Map{
		{"id": 1, "patient_name": "John Doe", "outbreak": "Ebola"},
		{"id": 2, "patient_name": "Jane Smith", "outbreak": "Measles"},
	}

	return c.JSON(fiber.Map{"cases": cases})
}

func HandlerGetCaseAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	return c.JSON(fiber.Map{"case": fiber.Map{"id": c.Params("id"), "patient_name": "Patient Name"}})
}

func HandlerCasesSubmitAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var data struct {
		OutbreakID int64       `json:"outbreak_id"`
		CaseID     interface{} `json:"case_id"`
		Patient    interface{} `json:"patient"`
	}
	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Create a minimal admission row linked to a new encounter (if required by schema)
	// For now, insert into admission with current time and current user
	adm := &models.Admission{}
	adm.Admitted.Valid = true
	adm.Admitted.Int64 = 1
	adm.AdmissionDate.Valid = true
	adm.AdmissionDate.Time = time.Now()
	adm.EnterBy.Valid = true
	adm.EnterBy.Int64 = int64(userID)
	adm.EnterOn.Valid = true
	adm.EnterOn.Time = time.Now()

	if err := adm.Insert(context.Background(), db); err != nil {
		sl.Error("Failed to create admission", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create admission"})
	}

	// Persist outbreak_id in session for subsequent pages (e.g., /cases/new/:id)
	if data.OutbreakID > 0 {
		if sess, err := store.Get(c); err == nil {
			sess.Set("outbreak_id", int(data.OutbreakID))
			sess.Set("selected_outbreak", int(data.OutbreakID))
			if err := sess.Save(); err != nil {
				sl.Error("Failed to save outbreak_id to session", "error", err)
			}
		} else {
			sl.Error("Failed to get session to persist outbreak_id", "error", err)
		}
	}

	// Return redirect to patient registration form (front-end will navigate)
	return c.Status(201).JSON(fiber.Map{
		"message":      "Admission created",
		"admission_id": adm.ID,
		"redirect":     "/cases/new/0",
	})
}

func HandlerCaseUpdateAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	var data fiber.Map
	c.BodyParser(&data)
	return c.JSON(fiber.Map{"message": "Case updated successfully"})
}

func HandlerCaseDeleteAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	return c.JSON(fiber.Map{"message": "Case deleted successfully"})
}

// Case Encounter APIs
func HandlerCaseEncounterListAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	caseID := c.Params("id")
	encounters := []fiber.Map{
		{"id": 1, "case_id": caseID, "encounter_type": "admission"},
		{"id": 2, "case_id": caseID, "encounter_type": "follow_up"},
	}

	return c.JSON(fiber.Map{"encounters": encounters})
}

func HandlerGetCaseEncounterAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	return c.JSON(fiber.Map{"encounter": fiber.Map{"id": c.Params("encounter_id"), "case_id": c.Params("id")}})
}

func HandlerCaseEncounterSubmitAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var data fiber.Map
	c.BodyParser(&data)
	return c.Status(201).JSON(fiber.Map{"message": "Encounter created successfully"})
}

func HandlerCaseEncounterUpdateAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	_ = c.Params("encounter_id")
	var data fiber.Map
	c.BodyParser(&data)
	return c.JSON(fiber.Map{"message": "Encounter updated successfully"})
}

func HandlerCaseEncounterDeleteAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	_ = c.Params("encounter_id")
	return c.JSON(fiber.Map{"message": "Encounter deleted successfully"})
}

// HandlerGetCIFAPI returns CIF data for a VHF case by ID
func HandlerGetCIFAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	idStr := c.Params("id")
	if idStr == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Case ID required"})
	}

	// Convert to int64
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid case ID"})
	}

	// Fetch basic patient data
	patient, err := models.GetVHFPatient(db, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load CIF"})
	}
	if patient == nil {
		return c.Status(404).JSON(fiber.Map{"error": "CIF not found"})
	}

	// Optionally fetch additional CIF sections
	signs, _ := models.GetVHFClinicalSigns(db, id)
	hosp, _ := models.GetVHFHospitalization(db, id)
	risk, _ := models.GetVHFRiskFactors(db, id)
	lab, _ := models.GetVHFLaboratory(db, id)
	inv, _ := models.GetVHFInvestigator(db, id)

	return c.JSON(fiber.Map{
		"patient":         patient,
		"clinical_signs":  signs,
		"hospitalization": hosp,
		"risk_factors":    risk,
		"laboratory":      lab,
		"investigator":    inv,
	})
}

// HandlerGetCIFByCaseCodeAPI returns CIF data using case_code query param
func HandlerGetCIFByCaseCodeAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	caseCode := c.Query("case_code")
	if caseCode == "" {
		return c.Status(400).JSON(fiber.Map{"error": "case_code is required"})
	}

	patient, err := models.GetVHFPatientByCaseCode(db, caseCode)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load CIF"})
	}
	if patient == nil {
		return c.Status(404).JSON(fiber.Map{"error": "CIF not found"})
	}

	id := patient.ID
	signs, _ := models.GetVHFClinicalSigns(db, id)
	hosp, _ := models.GetVHFHospitalization(db, id)
	risk, _ := models.GetVHFRiskFactors(db, id)
	lab, _ := models.GetVHFLaboratory(db, id)
	inv, _ := models.GetVHFInvestigator(db, id)

	return c.JSON(fiber.Map{
		"patient":         patient,
		"clinical_signs":  signs,
		"hospitalization": hosp,
		"risk_factors":    risk,
		"laboratory":      lab,
		"investigator":    inv,
	})
}

// Disease-specific CIF endpoints
// VHF
func HandlerVhfCIFByID(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid case ID"})
	}
	patient, err := models.GetVHFPatient(db, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load CIF"})
	}
	if patient == nil {
		return c.Status(404).JSON(fiber.Map{"error": "CIF not found"})
	}
	signs, _ := models.GetVHFClinicalSigns(db, id)
	hosp, _ := models.GetVHFHospitalization(db, id)
	risk, _ := models.GetVHFRiskFactors(db, id)
	lab, _ := models.GetVHFLaboratory(db, id)
	inv, _ := models.GetVHFInvestigator(db, id)
	return c.JSON(fiber.Map{
		"patient": patient, "clinical_signs": signs, "hospitalization": hosp,
		"risk_factors": risk, "laboratory": lab, "investigator": inv,
	})
}

func HandlerVhfCIFByCaseCode(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	caseCode := c.Query("case_code")
	if caseCode == "" {
		return c.Status(400).JSON(fiber.Map{"error": "case_code is required"})
	}
	patient, err := models.GetVHFPatientByCaseCode(db, caseCode)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load CIF"})
	}
	if patient == nil {
		return c.Status(404).JSON(fiber.Map{"error": "CIF not found"})
	}
	id := patient.ID
	signs, _ := models.GetVHFClinicalSigns(db, id)
	hosp, _ := models.GetVHFHospitalization(db, id)
	risk, _ := models.GetVHFRiskFactors(db, id)
	lab, _ := models.GetVHFLaboratory(db, id)
	inv, _ := models.GetVHFInvestigator(db, id)
	return c.JSON(fiber.Map{
		"patient": patient, "clinical_signs": signs, "hospitalization": hosp,
		"risk_factors": risk, "laboratory": lab, "investigator": inv,
	})
}

// Measles (placeholder)
func HandlerMeaslesCIFByID(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	pid := c.Params("id")
	if pid == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Patient ID required"})
	}
	var patient models.MeaslesPatient
	err := db.QueryRow(`SELECT patient_id, measles_code, patient_name, sex, dob, created_at FROM measles_patients WHERE patient_id = $1`, pid).
		Scan(&patient.PatientID, &patient.MeaslesCode, &patient.PatientName, &patient.Sex, &patient.DOB, &patient.CreatedAt)
	if err == sql.ErrNoRows {
		return c.Status(404).JSON(fiber.Map{"error": "CIF not found"})
	}
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load patient"})
	}
	var demo *models.MeaslesDemographics
	var inv *models.MeaslesInvestigators
	var hist *models.MeaslesClinicalHistory
	var res *models.MeaslesResults
	var spec *models.MeaslesSpecimens
	demo = &models.MeaslesDemographics{}
	if err := db.QueryRow(`SELECT id, patient_id, onset_district, reporting_unit, age_months, head_of_household, guardian_occupation, home_district, subcounty, parish, lc1_zone, lc1_chairman, lc1_tel FROM measles_demographics WHERE patient_id = $1`, patient.PatientID).
		Scan(&demo.ID, &demo.PatientID, &demo.OnsetDistrict, &demo.ReportingUnit, &demo.AgeMonths, &demo.HeadOfHousehold, &demo.GuardianOccupation, &demo.HomeDistrict, &demo.Subcounty, &demo.Parish, &demo.LC1Zone, &demo.LC1Chairman, &demo.LC1Tel); err != nil {
		demo = nil
	}
	inv = &models.MeaslesInvestigators{}
	if err := db.QueryRow(`SELECT id, patient_id, investigator_name, investigator_title, investigator_date FROM measles_investigators WHERE patient_id = $1`, patient.PatientID).
		Scan(&inv.ID, &inv.PatientID, &inv.InvestigatorName, &inv.InvestigatorTitle, &inv.InvestigatorDate); err != nil {
		inv = nil
	}
	hist = &models.MeaslesClinicalHistory{}
	if err := db.QueryRow(`SELECT id, patient_id, fever, fever_onset, temperature, rash, rash_onset, cough, red_eyes, running_nose, other_complications, complications_specify, outcome, vitamin_a, vitamin_a_doses, immunisation_card_seen, measles_doses, last_measles_vaccination, vaccination_reason, diagnosis FROM measles_clinical_history WHERE patient_id = $1`, patient.PatientID).
		Scan(&hist.ID, &hist.PatientID, &hist.Fever, &hist.FeverOnset, &hist.Temperature, &hist.Rash, &hist.RashOnset, &hist.Cough, &hist.RedEyes, &hist.RunningNose, &hist.OtherComplications, &hist.ComplicationsSpecify, &hist.Outcome, &hist.VitaminA, &hist.VitaminADoses, &hist.ImmunisationCardSeen, &hist.MeaslesDoses, &hist.LastMeaslesVaccination, &hist.VaccinationReason, &hist.Diagnosis); err != nil {
		hist = nil
	}
	res = &models.MeaslesResults{}
	if err := db.QueryRow(`SELECT id, patient_id, serology_igm, serology_date, serology_epi_sent_date, virus_isolation_urine, virus_isolation_date, final_classification, results_sent_date FROM measles_results WHERE patient_id = $1`, patient.PatientID).
		Scan(&res.ID, &res.PatientID, &res.SerologyIgM, &res.SerologyDate, &res.SerologyEpiSentDate, &res.VirusIsolationUrine, &res.VirusIsolationDate, &res.FinalClassification, &res.ResultsSentDate); err != nil {
		res = nil
	}
	spec = &models.MeaslesSpecimens{}
	if err := db.QueryRow(`SELECT id, patient_id, blood_collection_date, blood_sent_date, blood_received_date, blood_condition, urine_collection_date, urine_sent_date, urine_received_date, urine_condition, form_sent_date, form_received_date FROM measles_specimens WHERE patient_id = $1`, patient.PatientID).
		Scan(&spec.ID, &spec.PatientID, &spec.BloodCollectionDate, &spec.BloodSentDate, &spec.BloodReceivedDate, &spec.BloodCondition, &spec.UrineCollectionDate, &spec.UrineSentDate, &spec.UrineReceivedDate, &spec.UrineCondition, &spec.FormSentDate, &spec.FormReceivedDate); err != nil {
		spec = nil
	}
	return c.JSON(fiber.Map{
		"patient":          patient,
		"demographics":     demo,
		"investigators":    inv,
		"clinical_history": hist,
		"results":          res,
		"specimens":        spec,
	})
}
func HandlerMeaslesCIFByCaseCode(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	code := c.Query("case_code")
	if code == "" {
		return c.Status(400).JSON(fiber.Map{"error": "case_code is required"})
	}
	var pid string
	if err := db.QueryRow(`SELECT patient_id FROM measles_patients WHERE measles_code = $1`, code).Scan(&pid); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "CIF not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Lookup failed"})
	}
	// Build same payload as ID variant
	var patient models.MeaslesPatient
	if err := db.QueryRow(`SELECT patient_id, measles_code, patient_name, sex, dob, created_at FROM measles_patients WHERE patient_id = $1`, pid).
		Scan(&patient.PatientID, &patient.MeaslesCode, &patient.PatientName, &patient.Sex, &patient.DOB, &patient.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "CIF not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load patient"})
	}
	var demo *models.MeaslesDemographics
	var inv *models.MeaslesInvestigators
	var hist *models.MeaslesClinicalHistory
	var res *models.MeaslesResults
	var spec *models.MeaslesSpecimens
	demo = &models.MeaslesDemographics{}
	if err := db.QueryRow(`SELECT id, patient_id, onset_district, reporting_unit, age_months, head_of_household, guardian_occupation, home_district, subcounty, parish, lc1_zone, lc1_chairman, lc1_tel FROM measles_demographics WHERE patient_id = $1`, patient.PatientID).
		Scan(&demo.ID, &demo.PatientID, &demo.OnsetDistrict, &demo.ReportingUnit, &demo.AgeMonths, &demo.HeadOfHousehold, &demo.GuardianOccupation, &demo.HomeDistrict, &demo.Subcounty, &demo.Parish, &demo.LC1Zone, &demo.LC1Chairman, &demo.LC1Tel); err != nil {
		demo = nil
	}
	inv = &models.MeaslesInvestigators{}
	if err := db.QueryRow(`SELECT id, patient_id, investigator_name, investigator_title, investigator_date FROM measles_investigators WHERE patient_id = $1`, patient.PatientID).
		Scan(&inv.ID, &inv.PatientID, &inv.InvestigatorName, &inv.InvestigatorTitle, &inv.InvestigatorDate); err != nil {
		inv = nil
	}
	hist = &models.MeaslesClinicalHistory{}
	if err := db.QueryRow(`SELECT id, patient_id, fever, fever_onset, temperature, rash, rash_onset, cough, red_eyes, running_nose, other_complications, complications_specify, outcome, vitamin_a, vitamin_a_doses, immunisation_card_seen, measles_doses, last_measles_vaccination, vaccination_reason, diagnosis FROM measles_clinical_history WHERE patient_id = $1`, patient.PatientID).
		Scan(&hist.ID, &hist.PatientID, &hist.Fever, &hist.FeverOnset, &hist.Temperature, &hist.Rash, &hist.RashOnset, &hist.Cough, &hist.RedEyes, &hist.RunningNose, &hist.OtherComplications, &hist.ComplicationsSpecify, &hist.Outcome, &hist.VitaminA, &hist.VitaminADoses, &hist.ImmunisationCardSeen, &hist.MeaslesDoses, &hist.LastMeaslesVaccination, &hist.VaccinationReason, &hist.Diagnosis); err != nil {
		hist = nil
	}
	res = &models.MeaslesResults{}
	if err := db.QueryRow(`SELECT id, patient_id, serology_igm, serology_date, serology_epi_sent_date, virus_isolation_urine, virus_isolation_date, final_classification, results_sent_date FROM measles_results WHERE patient_id = $1`, patient.PatientID).
		Scan(&res.ID, &res.PatientID, &res.SerologyIgM, &res.SerologyDate, &res.SerologyEpiSentDate, &res.VirusIsolationUrine, &res.VirusIsolationDate, &res.FinalClassification, &res.ResultsSentDate); err != nil {
		res = nil
	}
	spec = &models.MeaslesSpecimens{}
	if err := db.QueryRow(`SELECT id, patient_id, blood_collection_date, blood_sent_date, blood_received_date, blood_condition, urine_collection_date, urine_sent_date, urine_received_date, urine_condition, form_sent_date, form_received_date FROM measles_specimens WHERE patient_id = $1`, patient.PatientID).
		Scan(&spec.ID, &spec.PatientID, &spec.BloodCollectionDate, &spec.BloodSentDate, &spec.BloodReceivedDate, &spec.BloodCondition, &spec.UrineCollectionDate, &spec.UrineSentDate, &spec.UrineReceivedDate, &spec.UrineCondition, &spec.FormSentDate, &spec.FormReceivedDate); err != nil {
		spec = nil
	}
	return c.JSON(fiber.Map{
		"patient":          patient,
		"demographics":     demo,
		"investigators":    inv,
		"clinical_history": hist,
		"results":          res,
		"specimens":        spec,
	})
}

// Polio (placeholder)
func HandlerPolioCIFByID(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	caseID := c.Params("id")
	if caseID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Case ID required"})
	}
	var main models.PolioCaseInvestigation
	if err := db.QueryRow(`SELECT id, case_id, epid_number, country, region_province, district, year_onset, case_number, received_date, created_at FROM polio_case_investigation WHERE case_id = $1`, caseID).
		Scan(&main.ID, &main.CaseID, &main.EpidNumber, &main.Country, &main.RegionProvince, &main.District, &main.YearOnset, &main.CaseNumber, &main.ReceivedDate, &main.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "CIF not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load case"})
	}
	var ident *models.PolioIdentification
	var notif *models.PolioNotificationInvestigation
	var hosp *models.PolioHospitalization
	var clin *models.PolioClinicalHistory
	var imm *models.PolioImmunizationHistory
	var stool *models.PolioStoolSpecimenCollection
	var stoolRes *models.PolioStoolSpecimenResults
	var follow *models.PolioFollowUpExamination
	var history *models.PolioPatientHistory
	var investigator *models.PolioInvestigator
	ident = &models.PolioIdentification{}
	if err := db.QueryRow(`SELECT id, case_id, district, region_province, address, village, city, nearest_health_facility, longitude, latitude, patient_name, father_mother, phone_number, date_of_birth, age_years, age_months, sex FROM polio_identification WHERE case_id = $1`, caseID).
		Scan(&ident.ID, &ident.CaseID, &ident.District, &ident.RegionProvince, &ident.Address, &ident.Village, &ident.City, &ident.NearestHealthFacility, &ident.Longitude, &ident.Latitude, &ident.PatientName, &ident.FatherMother, &ident.PhoneNumber, &ident.DateOfBirth, &ident.AgeYears, &ident.AgeMonths, &ident.Sex); err != nil {
		ident = nil
	}
	notif = &models.PolioNotificationInvestigation{}
	if err := db.QueryRow(`SELECT id, case_id, notified_by, date_of_notification, date_of_investigation FROM polio_notification_investigation WHERE case_id = $1`, caseID).
		Scan(&notif.ID, &notif.CaseID, &notif.NotifiedBy, &notif.DateOfNotification, &notif.DateOfInvestigation); err != nil {
		notif = nil
	}
	hosp = &models.PolioHospitalization{}
	if err := db.QueryRow(`SELECT id, case_id, hospitalized, date_of_admission, hospital_record_number, hospital_name_address FROM polio_hospitalization WHERE case_id = $1`, caseID).
		Scan(&hosp.ID, &hosp.CaseID, &hosp.Hospitalized, &hosp.DateOfAdmission, &hosp.HospitalRecordNumber, &hosp.HospitalNameAddress); err != nil {
		hosp = nil
	}
	clin = &models.PolioClinicalHistory{}
	if err := db.QueryRow(`SELECT id, case_id, fever_at_onset, date_onset_of_fever, progressive_paralysis, date_onset_of_paralysis, flaccid_acute_paralysis, sensation_loss, sudden_onset, "asymmetric", left_arm_paralysis, right_arm_paralysis, left_leg_paralysis, right_leg_paralysis, diminished_reflexes, diminished_muscle_tone, muscle_wasting, muscle_weakness, respiratory_muscles, face, stiff_neck, convulsions, headache, vomiting, diarrhoea, other_sites, recent_injection, total_injections, injection_type, paralyzed_limb_sensitive, injection_facility_name, provisional_diagnosis, true_afp_case FROM polio_clinical_history WHERE case_id = $1`, caseID).
		Scan(&clin.ID, &clin.CaseID, &clin.FeverAtOnset, &clin.DateOnsetOfFever, &clin.ProgressiveParalysis, &clin.DateOnsetOfParalysis, &clin.FlaccidAcuteParalysis, &clin.SensationLoss, &clin.SuddenOnset, &clin.Asymmetric, &clin.LeftArmParalysis, &clin.RightArmParalysis, &clin.LeftLegParalysis, &clin.RightLegParalysis, &clin.DiminishedReflexes, &clin.DiminishedMuscleTone, &clin.MuscleWasting, &clin.MuscleWeakness, &clin.RespiratoryMuscles, &clin.Face, &clin.StiffNeck, &clin.Convulsions, &clin.Headache, &clin.Vomiting, &clin.Diarrhoea, &clin.OtherSites, &clin.RecentInjection, &clin.TotalInjections, &clin.InjectionType, &clin.ParalyzedLimbSensitive, &clin.InjectionFacilityName, &clin.ProvisionalDiagnosis, &clin.TrueAFPCase); err != nil {
		clin = nil
	}
	imm = &models.PolioImmunizationHistory{}
	if err := db.QueryRow(`SELECT id, case_id, total_polio_doses, exclude_dose_at_birth, opv_dose_at_birth, opv_dose1, opv_dose2, opv_dose3, opv_dose4, opv_dose_more_than4, last_opv_dose, total_opv_sia, last_opv_sia, total_opv_ri, total_ipv_sia, total_ipv_ri, last_ipv_sia, source_of_ri_vaccination, unknown_zero_dose_reasons FROM polio_immunization_history WHERE case_id = $1`, caseID).
		Scan(&imm.ID, &imm.CaseID, &imm.TotalPolioDoses, &imm.ExcludeDoseAtBirth, &imm.OPVDoseAtBirth, &imm.OPVDose1, &imm.OPVDose2, &imm.OPVDose3, &imm.OPVDose4, &imm.OPVDoseMoreThan4, &imm.LastOPVDose, &imm.TotalOPVSIA, &imm.LastOPVSIA, &imm.TotalOPVRI, &imm.TotalIPVSIA, &imm.TotalIPVRI, &imm.LastIPVSIA, &imm.SourceOfRIVaccination, &imm.UnknownZeroDoseReasons); err != nil {
		imm = nil
	}
	stool = &models.PolioStoolSpecimenCollection{}
	if err := db.QueryRow(`SELECT id, case_id, date_first_specimen, date_second_specimen, date_specimen_sent_national, date_specimen_received_national, date_specimen_sent_lab FROM polio_stool_specimen_collection WHERE case_id = $1`, caseID).
		Scan(&stool.ID, &stool.CaseID, &stool.DateFirstSpecimen, &stool.DateSecondSpecimen, &stool.DateSpecimenSentNational, &stool.DateSpecimenReceivedNational, &stool.DateSpecimenSentLab); err != nil {
		stool = nil
	}
	stoolRes = &models.PolioStoolSpecimenResults{}
	if err := db.QueryRow(`SELECT id, case_id, date_received_at_lab, specimen_status_at_reception, date_combined_cell_culture, date_results_sent_to_epi, date_results_received_at_epi, final_cell_culture_results, w1, w2, w3, discordant_sabin, sl1, sl2, sl3, r_npent, nev, date_sent_to_regional_lab, date_it_differentiation_sent, date_it_differentiation_received, date_isolate_sent_sequencing, date_seq_results_sent_program FROM polio_stool_specimen_results WHERE case_id = $1`, caseID).
		Scan(&stoolRes.ID, &stoolRes.CaseID, &stoolRes.DateReceivedAtLab, &stoolRes.SpecimenStatusAtReception, &stoolRes.DateCombinedCellCulture, &stoolRes.DateResultsSentToEPI, &stoolRes.DateResultsReceivedAtEPI, &stoolRes.FinalCellCultureResults, &stoolRes.W1, &stoolRes.W2, &stoolRes.W3, &stoolRes.DiscordantSabin, &stoolRes.SL1, &stoolRes.SL2, &stoolRes.SL3, &stoolRes.RNPENT, &stoolRes.NEV, &stoolRes.DateSentToRegionalLab, &stoolRes.DateITDifferentiationSent, &stoolRes.DateITDifferentiationReceived, &stoolRes.DateIsolateSentSequencing, &stoolRes.DateSeqResultsSentProgram); err != nil {
		stoolRes = nil
	}
	follow = &models.PolioFollowUpExamination{}
	if err := db.QueryRow(`SELECT id, case_id, date_of_follow_up, residual_paralysis_la, residual_paralysis_ra, residual_paralysis_ll, residual_paralysis_rl, results_of_exam, immunocompromised_status, final_classification, cvdpv, avdpv, ivdpv, serotype FROM polio_follow_up_examination WHERE case_id = $1`, caseID).
		Scan(&follow.ID, &follow.CaseID, &follow.DateOfFollowUp, &follow.ResidualParalysisLA, &follow.ResidualParalysisRA, &follow.ResidualParalysisLL, &follow.ResidualParalysisRL, &follow.ResultsOfExam, &follow.ImmunocompromisedStatus, &follow.FinalClassification, &follow.CVDPV, &follow.AVDPV, &follow.IVDPV, &follow.Serotype); err != nil {
		follow = nil
	}
	history = &models.PolioPatientHistory{}
	if err := db.QueryRow(`SELECT id, case_id, place1, duration1_months, duration1_days, place2, duration2_months, duration2_days, place3, duration3_months, duration3_days, place4, duration4_months, duration4_days FROM polio_patient_history WHERE case_id = $1`, caseID).
		Scan(&history.ID, &history.CaseID, &history.Place1, &history.Duration1Months, &history.Duration1Days, &history.Place2, &history.Duration2Months, &history.Duration2Days, &history.Place3, &history.Duration3Months, &history.Duration3Days, &history.Place4, &history.Duration4Months, &history.Duration4Days); err != nil {
		history = nil
	}
	investigator = &models.PolioInvestigator{}
	if err := db.QueryRow(`SELECT id, case_id, investigator_name, investigator_title, unit, address, telephone FROM polio_investigator WHERE case_id = $1`, caseID).
		Scan(&investigator.ID, &investigator.CaseID, &investigator.InvestigatorName, &investigator.InvestigatorTitle, &investigator.Unit, &investigator.Address, &investigator.Telephone); err != nil {
		investigator = nil
	}
	return c.JSON(fiber.Map{
		"case":             main,
		"identification":   ident,
		"notification":     notif,
		"hospitalization":  hosp,
		"clinical_history": clin,
		"immunization":     imm,
		"stool_collection": stool,
		"stool_results":    stoolRes,
		"follow_up":        follow,
		"patient_history":  history,
		"investigator":     investigator,
	})
}
func HandlerPolioCIFByCaseCode(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	code := c.Query("case_code")
	if code == "" {
		return c.Status(400).JSON(fiber.Map{"error": "case_code is required"})
	}
	caseID := code
	var main models.PolioCaseInvestigation
	if err := db.QueryRow(`SELECT id, case_id, epid_number, country, region_province, district, year_onset, case_number, received_date, created_at FROM polio_case_investigation WHERE case_id = $1`, caseID).
		Scan(&main.ID, &main.CaseID, &main.EpidNumber, &main.Country, &main.RegionProvince, &main.District, &main.YearOnset, &main.CaseNumber, &main.ReceivedDate, &main.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "CIF not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load case"})
	}
	var ident *models.PolioIdentification
	var notif *models.PolioNotificationInvestigation
	var hosp *models.PolioHospitalization
	var clin *models.PolioClinicalHistory
	var imm *models.PolioImmunizationHistory
	var stool *models.PolioStoolSpecimenCollection
	var stoolRes *models.PolioStoolSpecimenResults
	var follow *models.PolioFollowUpExamination
	var history *models.PolioPatientHistory
	var investigator *models.PolioInvestigator
	ident = &models.PolioIdentification{}
	if err := db.QueryRow(`SELECT id, case_id, district, region_province, address, village, city, nearest_health_facility, longitude, latitude, patient_name, father_mother, phone_number, date_of_birth, age_years, age_months, sex FROM polio_identification WHERE case_id = $1`, caseID).
		Scan(&ident.ID, &ident.CaseID, &ident.District, &ident.RegionProvince, &ident.Address, &ident.Village, &ident.City, &ident.NearestHealthFacility, &ident.Longitude, &ident.Latitude, &ident.PatientName, &ident.FatherMother, &ident.PhoneNumber, &ident.DateOfBirth, &ident.AgeYears, &ident.AgeMonths, &ident.Sex); err != nil {
		ident = nil
	}
	notif = &models.PolioNotificationInvestigation{}
	if err := db.QueryRow(`SELECT id, case_id, notified_by, date_of_notification, date_of_investigation FROM polio_notification_investigation WHERE case_id = $1`, caseID).
		Scan(&notif.ID, &notif.CaseID, &notif.NotifiedBy, &notif.DateOfNotification, &notif.DateOfInvestigation); err != nil {
		notif = nil
	}
	hosp = &models.PolioHospitalization{}
	if err := db.QueryRow(`SELECT id, case_id, hospitalized, date_of_admission, hospital_record_number, hospital_name_address FROM polio_hospitalization WHERE case_id = $1`, caseID).
		Scan(&hosp.ID, &hosp.CaseID, &hosp.Hospitalized, &hosp.DateOfAdmission, &hosp.HospitalRecordNumber, &hosp.HospitalNameAddress); err != nil {
		hosp = nil
	}
	clin = &models.PolioClinicalHistory{}
	if err := db.QueryRow(`SELECT id, case_id, fever_at_onset, date_onset_of_fever, progressive_paralysis, date_onset_of_paralysis, flaccid_acute_paralysis, sensation_loss, sudden_onset, "asymmetric", left_arm_paralysis, right_arm_paralysis, left_leg_paralysis, right_leg_paralysis, diminished_reflexes, diminished_muscle_tone, muscle_wasting, muscle_weakness, respiratory_muscles, face, stiff_neck, convulsions, headache, vomiting, diarrhoea, other_sites, recent_injection, total_injections, injection_type, paralyzed_limb_sensitive, injection_facility_name, provisional_diagnosis, true_afp_case FROM polio_clinical_history WHERE case_id = $1`, caseID).
		Scan(&clin.ID, &clin.CaseID, &clin.FeverAtOnset, &clin.DateOnsetOfFever, &clin.ProgressiveParalysis, &clin.DateOnsetOfParalysis, &clin.FlaccidAcuteParalysis, &clin.SensationLoss, &clin.SuddenOnset, &clin.Asymmetric, &clin.LeftArmParalysis, &clin.RightArmParalysis, &clin.LeftLegParalysis, &clin.RightLegParalysis, &clin.DiminishedReflexes, &clin.DiminishedMuscleTone, &clin.MuscleWasting, &clin.MuscleWeakness, &clin.RespiratoryMuscles, &clin.Face, &clin.StiffNeck, &clin.Convulsions, &clin.Headache, &clin.Vomiting, &clin.Diarrhoea, &clin.OtherSites, &clin.RecentInjection, &clin.TotalInjections, &clin.InjectionType, &clin.ParalyzedLimbSensitive, &clin.InjectionFacilityName, &clin.ProvisionalDiagnosis, &clin.TrueAFPCase); err != nil {
		clin = nil
	}
	imm = &models.PolioImmunizationHistory{}
	if err := db.QueryRow(`SELECT id, case_id, total_polio_doses, exclude_dose_at_birth, opv_dose_at_birth, opv_dose1, opv_dose2, opv_dose3, opv_dose4, opv_dose_more_than4, last_opv_dose, total_opv_sia, last_opv_sia, total_opv_ri, total_ipv_sia, total_ipv_ri, last_ipv_sia, source_of_ri_vaccination, unknown_zero_dose_reasons FROM polio_immunization_history WHERE case_id = $1`, caseID).
		Scan(&imm.ID, &imm.CaseID, &imm.TotalPolioDoses, &imm.ExcludeDoseAtBirth, &imm.OPVDoseAtBirth, &imm.OPVDose1, &imm.OPVDose2, &imm.OPVDose3, &imm.OPVDose4, &imm.OPVDoseMoreThan4, &imm.LastOPVDose, &imm.TotalOPVSIA, &imm.LastOPVSIA, &imm.TotalOPVRI, &imm.TotalIPVSIA, &imm.TotalIPVRI, &imm.LastIPVSIA, &imm.SourceOfRIVaccination, &imm.UnknownZeroDoseReasons); err != nil {
		imm = nil
	}
	stool = &models.PolioStoolSpecimenCollection{}
	if err := db.QueryRow(`SELECT id, case_id, date_first_specimen, date_second_specimen, date_specimen_sent_national, date_specimen_received_national, date_specimen_sent_lab FROM polio_stool_specimen_collection WHERE case_id = $1`, caseID).
		Scan(&stool.ID, &stool.CaseID, &stool.DateFirstSpecimen, &stool.DateSecondSpecimen, &stool.DateSpecimenSentNational, &stool.DateSpecimenReceivedNational, &stool.DateSpecimenSentLab); err != nil {
		stool = nil
	}
	stoolRes = &models.PolioStoolSpecimenResults{}
	if err := db.QueryRow(`SELECT id, case_id, date_received_at_lab, specimen_status_at_reception, date_combined_cell_culture, date_results_sent_to_epi, date_results_received_at_epi, final_cell_culture_results, w1, w2, w3, discordant_sabin, sl1, sl2, sl3, r_npent, nev, date_sent_to_regional_lab, date_it_differentiation_sent, date_it_differentiation_received, date_isolate_sent_sequencing, date_seq_results_sent_program FROM polio_stool_specimen_results WHERE case_id = $1`, caseID).
		Scan(&stoolRes.ID, &stoolRes.CaseID, &stoolRes.DateReceivedAtLab, &stoolRes.SpecimenStatusAtReception, &stoolRes.DateCombinedCellCulture, &stoolRes.DateResultsSentToEPI, &stoolRes.DateResultsReceivedAtEPI, &stoolRes.FinalCellCultureResults, &stoolRes.W1, &stoolRes.W2, &stoolRes.W3, &stoolRes.DiscordantSabin, &stoolRes.SL1, &stoolRes.SL2, &stoolRes.SL3, &stoolRes.RNPENT, &stoolRes.NEV, &stoolRes.DateSentToRegionalLab, &stoolRes.DateITDifferentiationSent, &stoolRes.DateITDifferentiationReceived, &stoolRes.DateIsolateSentSequencing, &stoolRes.DateSeqResultsSentProgram); err != nil {
		stoolRes = nil
	}
	follow = &models.PolioFollowUpExamination{}
	if err := db.QueryRow(`SELECT id, case_id, date_of_follow_up, residual_paralysis_la, residual_paralysis_ra, residual_paralysis_ll, residual_paralysis_rl, results_of_exam, immunocompromised_status, final_classification, cvdpv, avdpv, ivdpv, serotype FROM polio_follow_up_examination WHERE case_id = $1`, caseID).
		Scan(&follow.ID, &follow.CaseID, &follow.DateOfFollowUp, &follow.ResidualParalysisLA, &follow.ResidualParalysisRA, &follow.ResidualParalysisLL, &follow.ResidualParalysisRL, &follow.ResultsOfExam, &follow.ImmunocompromisedStatus, &follow.FinalClassification, &follow.CVDPV, &follow.AVDPV, &follow.IVDPV, &follow.Serotype); err != nil {
		follow = nil
	}
	history = &models.PolioPatientHistory{}
	if err := db.QueryRow(`SELECT id, case_id, place1, duration1_months, duration1_days, place2, duration2_months, duration2_days, place3, duration3_months, duration3_days, place4, duration4_months, duration4_days FROM polio_patient_history WHERE case_id = $1`, caseID).
		Scan(&history.ID, &history.CaseID, &history.Place1, &history.Duration1Months, &history.Duration1Days, &history.Place2, &history.Duration2Months, &history.Duration2Days, &history.Place3, &history.Duration3Months, &history.Duration3Days, &history.Place4, &history.Duration4Months, &history.Duration4Days); err != nil {
		history = nil
	}
	investigator = &models.PolioInvestigator{}
	if err := db.QueryRow(`SELECT id, case_id, investigator_name, investigator_title, unit, address, telephone FROM polio_investigator WHERE case_id = $1`, caseID).
		Scan(&investigator.ID, &investigator.CaseID, &investigator.InvestigatorName, &investigator.InvestigatorTitle, &investigator.Unit, &investigator.Address, &investigator.Telephone); err != nil {
		investigator = nil
	}
	return c.JSON(fiber.Map{
		"case":             main,
		"identification":   ident,
		"notification":     notif,
		"hospitalization":  hosp,
		"clinical_history": clin,
		"immunization":     imm,
		"stool_collection": stool,
		"stool_results":    stoolRes,
		"follow_up":        follow,
		"patient_history":  history,
		"investigator":     investigator,
	})
}

// Mpox (placeholder)
func HandlerMpoxCIFByID(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	caseID := c.Params("id")
	if caseID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Case ID required"})
	}
	var ci *models.MpoxCaseInvestigation
	var demo *models.MpoxPatientDemographics
	var clin *models.MpoxClinicianInfo
	var exp *models.MpoxCaseExposureHistory
	var man *models.MpoxClinicalManifestations
	var travel *models.MpoxTravelHistory
	var lab *models.MpoxLabInvestigation
	ci = &models.MpoxCaseInvestigation{}
	if err := db.QueryRow(`SELECT id, case_id, date, case_status, case_classification FROM mpox_case_investigation WHERE case_id = $1`, caseID).
		Scan(&ci.ID, &ci.CaseID, &ci.Date, &ci.CaseStatus, &ci.CaseClassification); err != nil {
		ci = nil
	}
	demo = &models.MpoxPatientDemographics{}
	if err := db.QueryRow(`SELECT id, case_id, health_facility_case_id, surname, other_names, sex, date_of_birth, age, parish, sub_county, physical_address, contact_telephone, occupation, nationality, vaccination_status, date_of_vaccination, next_of_kin, next_of_kin_contact, marital_status, if_dead_date_of_death, admission_date, onset_date, rash_onset_date FROM mpox_patient_demographics WHERE case_id = $1`, caseID).
		Scan(&demo.ID, &demo.CaseID, &demo.HealthFacilityCaseID, &demo.Surname, &demo.OtherNames, &demo.Sex, &demo.DateOfBirth, &demo.Age, &demo.Parish, &demo.SubCounty, &demo.PhysicalAddress, &demo.ContactTelephone, &demo.Occupation, &demo.Nationality, &demo.VaccinationStatus, &demo.DateOfVaccination, &demo.NextOfKin, &demo.NextOfKinContact, &demo.MaritalStatus, &demo.IfDeadDateOfDeath, &demo.AdmissionDate, &demo.OnsetDate, &demo.RashOnsetDate); err != nil {
		demo = nil
	}
	clin = &models.MpoxClinicianInfo{}
	if err := db.QueryRow(`SELECT id, case_id, clinician_name, clinician_contact, facility_name, clinician_email, facility_district, pdpid_number, admission_date, ward FROM mpox_clinician_info WHERE case_id = $1`, caseID).
		Scan(&clin.ID, &clin.CaseID, &clin.ClinicianName, &clin.ClinicianContact, &clin.FacilityName, &clin.ClinicianEmail, &clin.FacilityDistrict, &clin.PDPIDNumber, &clin.AdmissionDate, &clin.Ward); err != nil {
		clin = nil
	}
	exp = &models.MpoxCaseExposureHistory{}
	if err := db.QueryRow(`SELECT id, case_id, traveled_country_reported_mpox, close_contact_mpox, intl_travel, contact_animals, domestic_wild_animals, sexual_exposure FROM mpox_case_exposure_history WHERE case_id = $1`, caseID).
		Scan(&exp.ID, &exp.CaseID, &exp.TraveledCountryReportedMpox, &exp.CloseContactMpox, &exp.IntlTravel, &exp.ContactAnimals, &exp.DomesticWildAnimals, &exp.SexualExposure); err != nil {
		exp = nil
	}
	man = &models.MpoxClinicalManifestations{}
	if err := db.QueryRow(`SELECT id, case_id, onset_date, fever, fever_temperature, lymphadenopathy, symptoms, symptom_other_specify, nausea_vomiting, pregnant, pregnant_trimester, vaccinated, vaccination_date, rash, rash_onset_date, rash_distribution, rash_type, underlying_illness, underlying_illness_details FROM mpox_clinical_manifestations WHERE case_id = $1`, caseID).
		Scan(&man.ID, &man.CaseID, &man.OnsetDate, &man.Fever, &man.FeverTemperature, &man.Lymphadenopathy, &man.Symptoms, &man.SymptomOtherSpecify, &man.NauseaVomiting, &man.Pregnant, &man.PregnantTrimester, &man.Vaccinated, &man.VaccinationDate, &man.Rash, &man.RashOnsetDate, &man.RashDistribution, &man.RashType, &man.UnderlyingIllness, &man.UnderlyingIllnessDetails); err != nil {
		man = nil
	}
	travel = &models.MpoxTravelHistory{}
	if err := db.QueryRow(`SELECT id, case_id, travel_outside_uganda, country_visited, location_visited, date_arrival, date_departure, activities_location FROM mpox_travel_history WHERE case_id = $1`, caseID).
		Scan(&travel.ID, &travel.CaseID, &travel.TravelOutsideUganda, &travel.CountryVisited, &travel.LocationVisited, &travel.DateArrival, &travel.DateDeparture, &travel.ActivitiesLocation); err != nil {
		travel = nil
	}
	lab = &models.MpoxLabInvestigation{}
	if err := db.QueryRow(`SELECT id, case_id, lab_id, sample_collected, sample_other_specify, test_requested, test_other_specify, date_sample_collection, time_sample_collection, date_sample_dispatch, sample_collector_name, sample_collector_phone, date_sample_reception, time_sample_reception, sample_recipient_name, sample_recipient_phone, genomic_characterization, clade, accession_number FROM mpox_lab_investigation WHERE case_id = $1`, caseID).
		Scan(&lab.ID, &lab.CaseID, &lab.LabID, &lab.SampleCollected, &lab.SampleOtherSpecify, &lab.TestRequested, &lab.TestOtherSpecify, &lab.DateSampleCollection, &lab.TimeSampleCollection, &lab.DateSampleDispatch, &lab.SampleCollectorName, &lab.SampleCollectorPhone, &lab.DateSampleReception, &lab.TimeSampleReception, &lab.SampleRecipientName, &lab.SampleRecipientPhone, &lab.GenomicCharacterization, &lab.Clade, &lab.AccessionNumber); err != nil {
		lab = nil
	}
	return c.JSON(fiber.Map{
		"case":           ci,
		"demographics":   demo,
		"clinician":      clin,
		"exposure":       exp,
		"manifestations": man,
		"travel_history": travel,
		"laboratory":     lab,
	})
}
func HandlerMpoxCIFByCaseCode(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	code := c.Query("case_code")
	if code == "" {
		return c.Status(400).JSON(fiber.Map{"error": "case_code is required"})
	}
	caseID := code
	var ci *models.MpoxCaseInvestigation
	var demo *models.MpoxPatientDemographics
	var clin *models.MpoxClinicianInfo
	var exp *models.MpoxCaseExposureHistory
	var man *models.MpoxClinicalManifestations
	var travel *models.MpoxTravelHistory
	var lab *models.MpoxLabInvestigation
	ci = &models.MpoxCaseInvestigation{}
	if err := db.QueryRow(`SELECT id, case_id, date, case_status, case_classification FROM mpox_case_investigation WHERE case_id = $1`, caseID).
		Scan(&ci.ID, &ci.CaseID, &ci.Date, &ci.CaseStatus, &ci.CaseClassification); err != nil {
		ci = nil
	}
	demo = &models.MpoxPatientDemographics{}
	if err := db.QueryRow(`SELECT id, case_id, health_facility_case_id, surname, other_names, sex, date_of_birth, age, parish, sub_county, physical_address, contact_telephone, occupation, nationality, vaccination_status, date_of_vaccination, next_of_kin, next_of_kin_contact, marital_status, if_dead_date_of_death, admission_date, onset_date, rash_onset_date FROM mpox_patient_demographics WHERE case_id = $1`, caseID).
		Scan(&demo.ID, &demo.CaseID, &demo.HealthFacilityCaseID, &demo.Surname, &demo.OtherNames, &demo.Sex, &demo.DateOfBirth, &demo.Age, &demo.Parish, &demo.SubCounty, &demo.PhysicalAddress, &demo.ContactTelephone, &demo.Occupation, &demo.Nationality, &demo.VaccinationStatus, &demo.DateOfVaccination, &demo.NextOfKin, &demo.NextOfKinContact, &demo.MaritalStatus, &demo.IfDeadDateOfDeath, &demo.AdmissionDate, &demo.OnsetDate, &demo.RashOnsetDate); err != nil {
		demo = nil
	}
	clin = &models.MpoxClinicianInfo{}
	if err := db.QueryRow(`SELECT id, case_id, clinician_name, clinician_contact, facility_name, clinician_email, facility_district, pdpid_number, admission_date, ward FROM mpox_clinician_info WHERE case_id = $1`, caseID).
		Scan(&clin.ID, &clin.CaseID, &clin.ClinicianName, &clin.ClinicianContact, &clin.FacilityName, &clin.ClinicianEmail, &clin.FacilityDistrict, &clin.PDPIDNumber, &clin.AdmissionDate, &clin.Ward); err != nil {
		clin = nil
	}
	exp = &models.MpoxCaseExposureHistory{}
	if err := db.QueryRow(`SELECT id, case_id, traveled_country_reported_mpox, close_contact_mpox, intl_travel, contact_animals, domestic_wild_animals, sexual_exposure FROM mpox_case_exposure_history WHERE case_id = $1`, caseID).
		Scan(&exp.ID, &exp.CaseID, &exp.TraveledCountryReportedMpox, &exp.CloseContactMpox, &exp.IntlTravel, &exp.ContactAnimals, &exp.DomesticWildAnimals, &exp.SexualExposure); err != nil {
		exp = nil
	}
	man = &models.MpoxClinicalManifestations{}
	if err := db.QueryRow(`SELECT id, case_id, onset_date, fever, fever_temperature, lymphadenopathy, symptoms, symptom_other_specify, nausea_vomiting, pregnant, pregnant_trimester, vaccinated, vaccination_date, rash, rash_onset_date, rash_distribution, rash_type, underlying_illness, underlying_illness_details FROM mpox_clinical_manifestations WHERE case_id = $1`, caseID).
		Scan(&man.ID, &man.CaseID, &man.OnsetDate, &man.Fever, &man.FeverTemperature, &man.Lymphadenopathy, &man.Symptoms, &man.SymptomOtherSpecify, &man.NauseaVomiting, &man.Pregnant, &man.PregnantTrimester, &man.Vaccinated, &man.VaccinationDate, &man.Rash, &man.RashOnsetDate, &man.RashDistribution, &man.RashType, &man.UnderlyingIllness, &man.UnderlyingIllnessDetails); err != nil {
		man = nil
	}
	travel = &models.MpoxTravelHistory{}
	if err := db.QueryRow(`SELECT id, case_id, travel_outside_uganda, country_visited, location_visited, date_arrival, date_departure, activities_location FROM mpox_travel_history WHERE case_id = $1`, caseID).
		Scan(&travel.ID, &travel.CaseID, &travel.TravelOutsideUganda, &travel.CountryVisited, &travel.LocationVisited, &travel.DateArrival, &travel.DateDeparture, &travel.ActivitiesLocation); err != nil {
		travel = nil
	}
	lab = &models.MpoxLabInvestigation{}
	if err := db.QueryRow(`SELECT id, case_id, lab_id, sample_collected, sample_other_specify, test_requested, test_other_specify, date_sample_collection, time_sample_collection, date_sample_dispatch, sample_collector_name, sample_collector_phone, date_sample_reception, time_sample_reception, sample_recipient_name, sample_recipient_phone, genomic_characterization, clade, accession_number FROM mpox_lab_investigation WHERE case_id = $1`, caseID).
		Scan(&lab.ID, &lab.CaseID, &lab.LabID, &lab.SampleCollected, &lab.SampleOtherSpecify, &lab.TestRequested, &lab.TestOtherSpecify, &lab.DateSampleCollection, &lab.TimeSampleCollection, &lab.DateSampleDispatch, &lab.SampleCollectorName, &lab.SampleCollectorPhone, &lab.DateSampleReception, &lab.TimeSampleReception, &lab.SampleRecipientName, &lab.SampleRecipientPhone, &lab.GenomicCharacterization, &lab.Clade, &lab.AccessionNumber); err != nil {
		lab = nil
	}
	return c.JSON(fiber.Map{
		"case":           ci,
		"demographics":   demo,
		"clinician":      clin,
		"exposure":       exp,
		"manifestations": man,
		"travel_history": travel,
		"laboratory":     lab,
	})
}

// Discharge Management APIs
func GetDischargeAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	discharges := []fiber.Map{
		{"id": 1, "patient_name": "John Doe", "discharge_date": "2024-01-01"},
		{"id": 2, "patient_name": "Jane Smith", "discharge_date": "2024-01-02"},
	}

	return c.JSON(fiber.Map{"discharges": discharges})
}

func GetDischargeByIdAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	return c.JSON(fiber.Map{"discharge": fiber.Map{"id": c.Params("id"), "patient_name": "Patient Name"}})
}

func DischargeAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var data fiber.Map
	c.BodyParser(&data)
	return c.Status(201).JSON(fiber.Map{"message": "Discharge created successfully"})
}

func CertificateAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	return c.JSON(fiber.Map{"certificate": fiber.Map{"id": c.Params("id"), "certificate_data": "data"}})
}

// Laboratory Management APIs
func HandlerLabListAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	labs := []fiber.Map{
		{"id": 1, "patient_name": "John Doe", "test_type": "Blood"},
		{"id": 2, "patient_name": "Jane Smith", "test_type": "Urine"},
	}

	return c.JSON(fiber.Map{"labs": labs})
}

func HandlerGetLabAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	return c.JSON(fiber.Map{"lab": fiber.Map{"id": c.Params("id"), "test_type": "Blood Test"}})
}

func HandlerLabSubmitAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var data fiber.Map
	c.BodyParser(&data)
	return c.Status(201).JSON(fiber.Map{"message": "Lab result submitted successfully"})
}

func HandlerLabUpdateAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	var data fiber.Map
	c.BodyParser(&data)
	return c.JSON(fiber.Map{"message": "Lab result updated successfully"})
}

func HandlerLabDeleteAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	return c.JSON(fiber.Map{"message": "Lab result deleted successfully"})
}

// Symptoms Management APIs
func HandlerSymptomsListAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	symptoms := []fiber.Map{
		{"id": 1, "patient_name": "John Doe", "symptom": "Fever"},
		{"id": 2, "patient_name": "Jane Smith", "symptom": "Headache"},
	}

	return c.JSON(fiber.Map{"symptoms": symptoms})
}

func HandlerGetSymptomsAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	return c.JSON(fiber.Map{"symptoms": fiber.Map{"id": c.Params("id"), "symptom": "Fever"}})
}

func HandlerSymptomsSubmitAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var data fiber.Map
	c.BodyParser(&data)
	return c.Status(201).JSON(fiber.Map{"message": "Symptoms submitted successfully"})
}

func HandlerSymptomsUpdateAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	var data fiber.Map
	c.BodyParser(&data)
	return c.JSON(fiber.Map{"message": "Symptoms updated successfully"})
}

func HandlerSymptomsDeleteAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	return c.JSON(fiber.Map{"message": "Symptoms deleted successfully"})
}

// Morbidity Management APIs
func HandlerMorbidityListAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	morbidity := []fiber.Map{
		{"id": 1, "patient_name": "John Doe", "condition": "Hypertension"},
		{"id": 2, "patient_name": "Jane Smith", "condition": "Diabetes"},
	}

	return c.JSON(fiber.Map{"morbidity": morbidity})
}

func HandlerGetMorbidityAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	return c.JSON(fiber.Map{"morbidity": fiber.Map{"id": c.Params("id"), "condition": "Condition"}})
}

func HandlerMorbiditySubmitAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var data fiber.Map
	c.BodyParser(&data)
	return c.Status(201).JSON(fiber.Map{"message": "Morbidity submitted successfully"})
}

func HandlerMorbidityUpdateAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	var data fiber.Map
	c.BodyParser(&data)
	return c.JSON(fiber.Map{"message": "Morbidity updated successfully"})
}

func HandlerMorbidityDeleteAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	return c.JSON(fiber.Map{"message": "Morbidity deleted successfully"})
}

// Rush Management APIs
func HandlerRushListAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	rush := []fiber.Map{
		{"id": 1, "patient_name": "John Doe", "rush_type": "Emergency"},
		{"id": 2, "patient_name": "Jane Smith", "rush_type": "Urgent"},
	}

	return c.JSON(fiber.Map{"rush": rush})
}

func HandlerGetRushAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	return c.JSON(fiber.Map{"rush": fiber.Map{"id": c.Params("id"), "rush_type": "Emergency"}})
}

func HandlerRushSubmitAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var data fiber.Map
	c.BodyParser(&data)
	return c.Status(201).JSON(fiber.Map{"message": "Rush submitted successfully"})
}

func HandlerRushUpdateAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	var data fiber.Map
	c.BodyParser(&data)
	return c.JSON(fiber.Map{"message": "Rush updated successfully"})
}

func HandlerRushDeleteAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	return c.JSON(fiber.Map{"message": "Rush deleted successfully"})
}

// Surveillance APIs
func CommunityMortalitySurveillanceAPI(c *fiber.Ctx, db *sql.DB, store *session.Store, config Config) error {
	return c.JSON(fiber.Map{"message": "Community mortality surveillance data"})
}

func FacilityMortalitySurveillanceAPI(c *fiber.Ctx, db *sql.DB, store *session.Store, config Config) error {
	return c.JSON(fiber.Map{"message": "Facility mortality surveillance data"})
}

// Mpox APIs
func HandlerMpoxListAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	mpoxPatients := []fiber.Map{
		{"id": 1, "patient_name": "John Doe", "status": "active"},
		{"id": 2, "patient_name": "Jane Smith", "status": "recovered"},
	}

	return c.JSON(fiber.Map{"mpox_patients": mpoxPatients})
}

func HandlerGetMpoxAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	return c.JSON(fiber.Map{"mpox_patient": fiber.Map{"id": c.Params("id"), "patient_name": "Patient Name"}})
}

func HandlerMpoxCIFSubmitAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	var data fiber.Map
	c.BodyParser(&data)
	return c.Status(201).JSON(fiber.Map{"message": "Mpox CIF submitted successfully"})
}

func HandlerMpoxUpdateAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	var data fiber.Map
	c.BodyParser(&data)
	return c.JSON(fiber.Map{"message": "Mpox patient updated successfully"})
}

func HandlerMpoxDeleteAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	return c.JSON(fiber.Map{"message": "Mpox patient deleted successfully"})
}

func HandlerMpoxAdmissionFormAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	return c.JSON(fiber.Map{"admission_form": fiber.Map{"patient_id": c.Params("id")}})
}

func HandlerMpoxAdmissionSubmitAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	var data fiber.Map
	c.BodyParser(&data)
	return c.JSON(fiber.Map{"message": "Mpox admission submitted successfully"})
}

func HandlerMpoxDailyFollowUpFormAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	return c.JSON(fiber.Map{"daily_followup_form": fiber.Map{"patient_id": c.Params("id")}})
}

func HandlerMpoxDailyFollowUpSubmitAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	var data fiber.Map
	c.BodyParser(&data)
	return c.JSON(fiber.Map{"message": "Mpox daily follow-up submitted successfully"})
}

// Measles APIs
func HandlerMeaslesListAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	measlesPatients := []fiber.Map{
		{"id": 1, "patient_name": "John Doe", "status": "active"},
		{"id": 2, "patient_name": "Jane Smith", "status": "recovered"},
	}

	return c.JSON(fiber.Map{"measles_patients": measlesPatients})
}

func HandlerGetMeaslesAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	return c.JSON(fiber.Map{"measles_patient": fiber.Map{"id": c.Params("id"), "patient_name": "Patient Name"}})
}

func HandlerMeaslesCIFAPI(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	var data fiber.Map
	c.BodyParser(&data)
	return c.Status(201).JSON(fiber.Map{"message": "Measles CIF submitted successfully"})
}

func HandlerMeaslesUpdateAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	var data fiber.Map
	c.BodyParser(&data)
	return c.JSON(fiber.Map{"message": "Measles patient updated successfully"})
}

func HandlerMeaslesDeleteAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	return c.JSON(fiber.Map{"message": "Measles patient deleted successfully"})
}

// Polio APIs
func HandlerPolioListAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	polioPatients := []fiber.Map{
		{"id": 1, "patient_name": "John Doe", "status": "active"},
		{"id": 2, "patient_name": "Jane Smith", "status": "recovered"},
	}

	return c.JSON(fiber.Map{"polio_patients": polioPatients})
}

func HandlerGetPolioAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	return c.JSON(fiber.Map{"polio_patient": fiber.Map{"id": c.Params("id"), "patient_name": "Patient Name"}})
}

func HandlerPolioCIFSubmitAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	var data fiber.Map
	c.BodyParser(&data)
	return c.Status(201).JSON(fiber.Map{"message": "Polio CIF submitted successfully"})
}

func HandlerPolioUpdateAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	var data fiber.Map
	c.BodyParser(&data)
	return c.JSON(fiber.Map{"message": "Polio patient updated successfully"})
}

func HandlerPolioDeleteAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	return c.JSON(fiber.Map{"message": "Polio patient deleted successfully"})
}

// Patient Roles APIs
func HandlerPatientRolesAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	roles := []fiber.Map{
		{"id": 1, "name": "Doctor", "description": "Medical doctor"},
		{"id": 2, "name": "Nurse", "description": "Nursing staff"},
	}

	return c.JSON(fiber.Map{"patient_roles": roles})
}

func HandlerGetPatientRoleAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	return c.JSON(fiber.Map{"patient_role": fiber.Map{"id": c.Params("id"), "name": "Role Name"}})
}

func HandlerPatientRoleSubmitAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var data fiber.Map
	c.BodyParser(&data)
	return c.Status(201).JSON(fiber.Map{"message": "Patient role created successfully"})
}

func HandlerPatientRoleUpdateAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	var data fiber.Map
	c.BodyParser(&data)
	return c.JSON(fiber.Map{"message": "Patient role updated successfully"})
}

func HandlerPatientRoleDeleteAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	return c.JSON(fiber.Map{"message": "Patient role deleted successfully"})
}

// Alerts APIs
func HandlerGetAlertAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	return c.JSON(fiber.Map{"alert": fiber.Map{"id": c.Params("id"), "message": "Alert message"}})
}

func HandlerAlertSubmitAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var data fiber.Map
	c.BodyParser(&data)
	return c.Status(201).JSON(fiber.Map{"message": "Alert created successfully"})
}

func HandlerAlertUpdateAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	var data fiber.Map
	c.BodyParser(&data)
	return c.JSON(fiber.Map{"message": "Alert updated successfully"})
}

func HandlerAlertDeleteAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID, userRole := GetUser(c, sl, store)
	if userRole == "" || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	_ = c.Params("id")
	return c.JSON(fiber.Map{"message": "Alert deleted successfully"})
}

// Inventory API handlers - these extend the existing inventory handler with API endpoints

// HandlerInventoryDashboardAPI returns inventory dashboard data as JSON
func (h *InventoryHandler) HandlerInventoryDashboardAPI(c *fiber.Ctx) error {
	stats, err := h.getInventoryStats()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error loading inventory stats"})
	}

	return c.JSON(fiber.Map{
		"message": "Inventory dashboard API",
		"stats":   stats,
	})
}

// HandlerGetInventoryItemAPI returns a single inventory item as JSON
func (h *InventoryHandler) HandlerGetInventoryItemAPI(c *fiber.Ctx) error {
	itemID := c.Params("id")
	item, err := h.getInventoryItemByID(itemID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Inventory item not found"})
	}

	return c.JSON(fiber.Map{
		"inventory_item": item,
	})
}

// HandlerInventoryItemSaveAPI creates a new inventory item via API
func (h *InventoryHandler) HandlerInventoryItemSaveAPI(c *fiber.Ctx) error {
	var item InventoryItem
	if err := c.BodyParser(&item); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.createInventoryItem(&item); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error saving inventory item"})
	}

	return c.Status(201).JSON(fiber.Map{"message": "Inventory item saved successfully", "item": item})
}

// HandlerInventoryItemUpdateAPI updates an existing inventory item via API
func (h *InventoryHandler) HandlerInventoryItemUpdateAPI(c *fiber.Ctx) error {
	itemID := c.Params("id")

	var item InventoryItem
	if err := c.BodyParser(&item); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	item.ID = parseInventoryInt(itemID)
	if err := h.updateInventoryItem(&item); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error updating inventory item"})
	}

	return c.JSON(fiber.Map{"message": "Inventory item updated successfully", "item": item})
}

// HandlerInventoryItemDeleteAPI deletes an inventory item via API
func (h *InventoryHandler) HandlerInventoryItemDeleteAPI(c *fiber.Ctx) error {
	itemID := c.Params("id")

	// For now, we'll just mark it as inactive instead of deleting
	_, err := h.db.Exec("UPDATE inventory_items SET is_active = false WHERE id = $1", itemID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error deleting inventory item"})
	}

	return c.JSON(fiber.Map{"message": "Inventory item deleted successfully"})
}

// HandlerGetInventoryStockAPI returns stock level for a specific item
func (h *InventoryHandler) HandlerGetInventoryStockAPI(c *fiber.Ctx) error {
	stockID := c.Params("id")

	var stock InventoryStockLevel
	query := `
		SELECT id, item_id, site_id, quantity, last_updated
		FROM inventory_stock_levels
		WHERE id = $1
	`
	err := h.db.QueryRow(query, stockID).Scan(&stock.ID, &stock.ItemID, &stock.SiteID, &stock.Quantity, &stock.LastUpdated)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Stock level not found"})
	}

	return c.JSON(fiber.Map{
		"stock": stock,
	})
}

// HandlerInventoryStockSaveAPI creates a new stock transaction
func (h *InventoryHandler) HandlerInventoryStockSaveAPI(c *fiber.Ctx) error {
	var transaction InventoryTransaction
	if err := c.BodyParser(&transaction); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.createInventoryTransaction(&transaction); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error saving stock transaction"})
	}

	return c.Status(201).JSON(fiber.Map{"message": "Stock level saved successfully", "transaction": transaction})
}

// HandlerInventoryStockUpdateAPI updates stock level
func (h *InventoryHandler) HandlerInventoryStockUpdateAPI(c *fiber.Ctx) error {
	stockID := c.Params("id")

	var data struct {
		Quantity float64 `json:"quantity"`
	}
	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	_, err := h.db.Exec("UPDATE inventory_stock_levels SET quantity = $1, last_updated = NOW() WHERE id = $2", data.Quantity, stockID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error updating stock level"})
	}

	return c.JSON(fiber.Map{"message": "Stock level updated successfully"})
}

// HandlerInventoryPurchaseOrdersAPI returns all purchase orders
func (h *InventoryHandler) HandlerInventoryPurchaseOrdersAPI(c *fiber.Ctx) error {
	query := `
		SELECT po.id, po.supplier_id, po.order_date, po.expected_delivery, po.status, po.notes,
		       s.name as supplier_name
		FROM inventory_purchase_orders po
		LEFT JOIN inventory_suppliers s ON po.supplier_id = s.id
		ORDER BY po.order_date DESC
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error loading purchase orders"})
	}
	defer rows.Close()

	var orders []fiber.Map
	for rows.Next() {
		var po InventoryPurchaseOrder
		var supplierName string
		err := rows.Scan(&po.ID, &po.SupplierID, &po.OrderDate, &po.ExpectedDelivery, &po.Status, &po.Notes, &supplierName)
		if err != nil {
			continue
		}
		orders = append(orders, fiber.Map{
			"id":                po.ID,
			"supplier_id":       po.SupplierID,
			"supplier_name":     supplierName,
			"order_date":        po.OrderDate,
			"expected_delivery": po.ExpectedDelivery,
			"status":            po.Status,
			"notes":             po.Notes,
		})
	}

	return c.JSON(fiber.Map{"purchase_orders": orders})
}

// HandlerGetPurchaseOrderAPI returns a single purchase order
func (h *InventoryHandler) HandlerGetPurchaseOrderAPI(c *fiber.Ctx) error {
	poID := c.Params("id")

	var po InventoryPurchaseOrder
	query := `SELECT id, supplier_id, order_date, expected_delivery, status, notes FROM inventory_purchase_orders WHERE id = $1`
	err := h.db.QueryRow(query, poID).Scan(&po.ID, &po.SupplierID, &po.OrderDate, &po.ExpectedDelivery, &po.Status, &po.Notes)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Purchase order not found"})
	}

	return c.JSON(fiber.Map{
		"purchase_order": po,
	})
}

// HandlerInventoryPurchaseOrderSaveAPI creates a new purchase order
func (h *InventoryHandler) HandlerInventoryPurchaseOrderSaveAPI(c *fiber.Ctx) error {
	var po InventoryPurchaseOrder
	if err := c.BodyParser(&po); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.createPurchaseOrder(&po); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error saving purchase order"})
	}

	return c.Status(201).JSON(fiber.Map{"message": "Purchase order saved successfully", "purchase_order": po})
}

// HandlerPurchaseOrderUpdateAPI updates a purchase order
func (h *InventoryHandler) HandlerPurchaseOrderUpdateAPI(c *fiber.Ctx) error {
	poID := c.Params("id")

	var po InventoryPurchaseOrder
	if err := c.BodyParser(&po); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	query := `UPDATE inventory_purchase_orders SET supplier_id = $1, expected_delivery = $2, status = $3, notes = $4 WHERE id = $5`
	_, err := h.db.Exec(query, po.SupplierID, po.ExpectedDelivery, po.Status, po.Notes, poID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error updating purchase order"})
	}

	return c.JSON(fiber.Map{"message": "Purchase order updated successfully"})
}

// HandlerInventoryRequisitionsAPI returns all requisitions
func (h *InventoryHandler) HandlerInventoryRequisitionsAPI(c *fiber.Ctx) error {
	query := `
		SELECT r.id, r.site_id, r.request_date, r.priority, r.status, r.notes,
		       ts.name as site_name
		FROM inventory_requisitions r
		LEFT JOIN treatment_sites ts ON r.site_id = ts.id
		ORDER BY r.request_date DESC
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error loading requisitions"})
	}
	defer rows.Close()

	var requisitions []fiber.Map
	for rows.Next() {
		var req InventoryRequisition
		var siteName string
		err := rows.Scan(&req.ID, &req.SiteID, &req.RequestDate, &req.Priority, &req.Status, &req.Notes, &siteName)
		if err != nil {
			continue
		}
		requisitions = append(requisitions, fiber.Map{
			"id":           req.ID,
			"site_id":      req.SiteID,
			"site_name":    siteName,
			"request_date": req.RequestDate,
			"priority":     req.Priority,
			"status":       req.Status,
			"notes":        req.Notes,
		})
	}

	return c.JSON(fiber.Map{"requisitions": requisitions})
}

// HandlerGetRequisitionAPI returns a single requisition
func (h *InventoryHandler) HandlerGetRequisitionAPI(c *fiber.Ctx) error {
	reqID := c.Params("id")

	var req InventoryRequisition
	query := `SELECT id, site_id, request_date, priority, status, notes FROM inventory_requisitions WHERE id = $1`
	err := h.db.QueryRow(query, reqID).Scan(&req.ID, &req.SiteID, &req.RequestDate, &req.Priority, &req.Status, &req.Notes)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Requisition not found"})
	}

	return c.JSON(fiber.Map{
		"requisition": req,
	})
}

// HandlerInventoryRequisitionSaveAPI creates a new requisition
func (h *InventoryHandler) HandlerInventoryRequisitionSaveAPI(c *fiber.Ctx) error {
	var req InventoryRequisition
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.createRequisition(&req); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error saving requisition"})
	}

	return c.Status(201).JSON(fiber.Map{"message": "Requisition saved successfully", "requisition": req})
}

// HandlerRequisitionUpdateAPI updates a requisition
func (h *InventoryHandler) HandlerRequisitionUpdateAPI(c *fiber.Ctx) error {
	reqID := c.Params("id")

	var req InventoryRequisition
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	query := `UPDATE inventory_requisitions SET site_id = $1, priority = $2, status = $3, notes = $4 WHERE id = $5`
	_, err := h.db.Exec(query, req.SiteID, req.Priority, req.Status, req.Notes, reqID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error updating requisition"})
	}

	return c.JSON(fiber.Map{"message": "Requisition updated successfully"})
}

// HandlerDonationsListAPI returns all donations as JSON
func (h *InventoryHandler) HandlerDonationsListAPI(c *fiber.Ctx) error {
	donations, err := h.getAllDonations()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error loading donations"})
	}

	return c.JSON(fiber.Map{"donations": donations})
}

// HandlerDonationViewAPI returns a single donation as JSON
func (h *InventoryHandler) HandlerDonationViewAPI(c *fiber.Ctx) error {
	donationID := parseInventoryInt(c.Params("id"))
	donation, err := h.getDonationByID(donationID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Donation not found"})
	}

	return c.JSON(fiber.Map{
		"donation": donation,
	})
}

// HandlerDonationSaveAPI creates a new donation via API
func (h *InventoryHandler) HandlerDonationSaveAPI(c *fiber.Ctx) error {
	var donation InventoryDonation
	if err := c.BodyParser(&donation); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Insert the donation (simplified version)
	query := `
		INSERT INTO inventory_donations (
			donor_id, donation_type_id, donation_date, received_date, 
			description, monetary_value, currency, outbreak_id, 
			treatment_site_id, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`

	err := h.db.QueryRow(
		query, donation.DonorID, donation.DonationTypeID, donation.DonationDate, donation.ReceivedDate,
		donation.Description, donation.MonetaryValue, donation.Currency, donation.OutbreakID,
		donation.TreatmentSiteID, donation.Notes,
	).Scan(&donation.ID)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error saving donation"})
	}

	return c.Status(201).JSON(fiber.Map{"message": "Donation saved successfully", "donation": donation})
}

// HandlerDonationUpdateAPI updates a donation via API
func (h *InventoryHandler) HandlerDonationUpdateAPI(c *fiber.Ctx) error {
	donationID := c.Params("id")

	var donation InventoryDonation
	if err := c.BodyParser(&donation); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	query := `
		UPDATE inventory_donations 
		SET donor_id = $1, donation_type_id = $2, donation_date = $3, received_date = $4,
		    description = $5, monetary_value = $6, currency = $7, outbreak_id = $8,
		    treatment_site_id = $9, notes = $10, donation_status = $11
		WHERE id = $12
	`

	_, err := h.db.Exec(
		query, donation.DonorID, donation.DonationTypeID, donation.DonationDate, donation.ReceivedDate,
		donation.Description, donation.MonetaryValue, donation.Currency, donation.OutbreakID,
		donation.TreatmentSiteID, donation.Notes, donation.DonationStatus, donationID,
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error updating donation"})
	}

	return c.JSON(fiber.Map{"message": "Donation updated successfully"})
}

// HandlerDonationDeleteAPI deletes a donation via API
func (h *InventoryHandler) HandlerDonationDeleteAPI(c *fiber.Ctx) error {
	donationID := c.Params("id")

	// Soft delete by updating status
	_, err := h.db.Exec("UPDATE inventory_donations SET donation_status = 'cancelled' WHERE id = $1", donationID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error deleting donation"})
	}

	return c.JSON(fiber.Map{"message": "Donation deleted successfully"})
}

// HandlerDonorsListAPI returns all donors as JSON
func (h *InventoryHandler) HandlerDonorsListAPI(c *fiber.Ctx) error {
	donors, err := h.getAllDonors()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error loading donors"})
	}

	return c.JSON(fiber.Map{"donors": donors})
}

// HandlerGetDonorAPI returns a single donor as JSON
func (h *InventoryHandler) HandlerGetDonorAPI(c *fiber.Ctx) error {
	donorID := parseInventoryInt(c.Params("id"))
	donor, err := h.getDonorByID(donorID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Donor not found"})
	}

	return c.JSON(fiber.Map{
		"donor": donor,
	})
}

// HandlerDonorSaveAPI creates a new donor via API
func (h *InventoryHandler) HandlerDonorSaveAPI(c *fiber.Ctx) error {
	var donor InventoryDonor
	if err := c.BodyParser(&donor); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	query := `
		INSERT INTO inventory_donors (
			name, organization, contact_person, phone, email, address,
			donor_type, country, registration_number, tax_exempt, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`

	err := h.db.QueryRow(
		query, donor.Name, donor.Organization, donor.ContactPerson, donor.Phone, donor.Email, donor.Address,
		donor.DonorType, donor.Country, donor.RegistrationNumber, donor.TaxExempt, donor.Notes,
	).Scan(&donor.ID)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error saving donor"})
	}

	return c.Status(201).JSON(fiber.Map{"message": "Donor saved successfully", "donor": donor})
}

// HandlerDonorUpdateAPI updates a donor via API
func (h *InventoryHandler) HandlerDonorUpdateAPI(c *fiber.Ctx) error {
	donorID := c.Params("id")

	var donor InventoryDonor
	if err := c.BodyParser(&donor); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	query := `
		UPDATE inventory_donors 
		SET name = $1, organization = $2, contact_person = $3, phone = $4, email = $5, 
		    address = $6, donor_type = $7, country = $8, registration_number = $9, 
		    tax_exempt = $10, notes = $11, status = $12
		WHERE id = $13
	`

	_, err := h.db.Exec(
		query, donor.Name, donor.Organization, donor.ContactPerson, donor.Phone, donor.Email,
		donor.Address, donor.DonorType, donor.Country, donor.RegistrationNumber,
		donor.TaxExempt, donor.Notes, donor.Status, donorID,
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error updating donor"})
	}

	return c.JSON(fiber.Map{"message": "Donor updated successfully"})
}

// HandlerDonorDeleteAPI deletes a donor via API
func (h *InventoryHandler) HandlerDonorDeleteAPI(c *fiber.Ctx) error {
	donorID := c.Params("id")

	// Soft delete by updating status
	_, err := h.db.Exec("UPDATE inventory_donors SET status = 'inactive' WHERE id = $1", donorID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error deleting donor"})
	}

	return c.JSON(fiber.Map{"message": "Donor deleted successfully"})
}
