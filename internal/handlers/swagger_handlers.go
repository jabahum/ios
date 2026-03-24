package handlers

import (
	"database/sql"
	"log/slog"

	"case/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// Login godoc
// @Summary User login
// @Description Creates a session cookie (fiber_sess). For API/Swagger use Content-Type application/json and body {"username","password"}. HTML forms may POST as application/x-www-form-urlencoded with the same fields.
// @Tags Authentication
// @Accept application/json
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param credentials body LoginCredentials true "Credentials (required for JSON requests)"
// @Success 200 {object} map[string]interface{} "Login successful (JSON mode); includes redirect path for the web app"
// @Failure 400 {object} map[string]string "Missing username or password"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Router /login [post]
func SwaggerLogin(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerLoginSubmit(c, db, sl, store, config)
}

// Logout godoc
// @Summary User logout
// @Description End user session
// @Tags Authentication
// @Success 302 "Redirect to login"
// @Router /logout [get]
func SwaggerLogout(c *fiber.Ctx, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerLoginOut(c, sl, store, config)
}

// GetDistricts godoc
// @Summary Get all districts
// @Description Retrieve list of all districts in Uganda
// @Tags Locations
// @Produce json
// @Success 200 {array} map[string]interface{} "List of districts"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/locations/districts [get]
func SwaggerGetDistricts(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	return HandlerGetDistricts(c, db, sl)
}

// GetSubcounties godoc
// @Summary Get subcounties by district
// @Description Retrieve subcounties for a specific district
// @Tags Locations
// @Produce json
// @Param district_id path int true "District ID"
// @Success 200 {array} map[string]interface{} "List of subcounties"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/locations/subcounties/{district_id} [get]
func SwaggerGetSubcounties(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	return HandlerGetSubcountiesByDistrict(c, db, sl)
}

// GetParishes godoc
// @Summary Get parishes by subcounty
// @Description Retrieve parishes for a specific subcounty
// @Tags Locations
// @Produce json
// @Param subcounty_id path int true "Subcounty ID"
// @Success 200 {array} map[string]interface{} "List of parishes"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/locations/parishes/{subcounty_id} [get]
func SwaggerGetParishes(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	return HandlerGetParishesBySubcounty(c, db, sl)
}

// GetVillages godoc
// @Summary Get villages by parish
// @Description Retrieve villages for a specific parish
// @Tags Locations
// @Produce json
// @Param parish_id path int true "Parish ID"
// @Success 200 {array} map[string]interface{} "List of villages"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/locations/villages/{parish_id} [get]
func SwaggerGetVillages(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	return HandlerGetVillagesByParish(c, db, sl)
}

// GetFacilities godoc
// @Summary Get all facilities
// @Description Retrieve list of health facilities
// @Tags Facilities
// @Produce json
// @Param district_id query int false "Filter by district"
// @Success 200 {array} map[string]interface{} "List of facilities"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/facilities [get]
func SwaggerGetFacilities(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	return HandlerGetFacilities(c, db, sl)
}

// GetOutbreaks godoc
// @Summary Get all outbreaks
// @Description Accessible outbreaks for the current user. Response: `{ "outbreaks": [ ... ] }`.
// @Tags Outbreaks
// @Produce json
// @Security SessionAuth
// @Success 200 {object} map[string]interface{} "outbreaks array"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/outbreaks [get]
func SwaggerGetOutbreaks(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store) error {
	return HandlerGetOutbreaksAPI(c, db, sl, store)
}

// GetVHFCases godoc
// @Summary Get VHF cases
// @Description Retrieve list of Viral Hemorrhagic Fever cases
// @Tags VHF
// @Produce json
// @Security SessionAuth
// @Param outbreak_id query int false "Filter by outbreak"
// @Param status query string false "Filter by status"
// @Success 200 {array} map[string]interface{} "List of VHF cases"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/vhf/patients [get]
func SwaggerGetVHFCases(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerVHFListAPI(c, db, sl, store, config)
}

// GetVHFCase godoc
// @Summary Get VHF case by ID
// @Description Retrieve detailed information about a specific VHF case
// @Tags VHF
// @Produce json
// @Security SessionAuth
// @Param id path int true "Case ID"
// @Success 200 {object} map[string]interface{} "VHF case details"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Case not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/vhf/patients/{id} [get]
func SwaggerGetVHFCase(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerVHFViewAPI(c, db, sl, store, config)
}

// CreateVHFCase godoc
// @Summary Create VHF case
// @Description Create a new Viral Hemorrhagic Fever case
// @Tags VHF
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param case body APIVHFPatientCreateRequest true "VHF case data"
// @Success 201 {object} map[string]interface{} "Case created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/vhf/patients [post]
func SwaggerCreateVHFCase(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config, sms *services.SMSService) error {
	return HandlerVHFPatientSubmitAPI(c, db, sl, store, config, sms)
}

// UpdateVHFCase godoc
// @Summary Update VHF case
// @Description Update an existing Viral Hemorrhagic Fever case by patient ID
// @Tags VHF
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path int true "Patient ID"
// @Param body body object true "VHF case fields"
// @Success 200 {object} map[string]interface{} "Updated case"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/vhf/patients/{id} [put]
func SwaggerUpdateVHFCase(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerVHFUpdateAPI(c, db, sl, store, config)
}

// DeleteVHFCase godoc
// @Summary Delete VHF case
// @Description Delete a VHF patient record by ID
// @Tags VHF
// @Produce json
// @Security SessionAuth
// @Param id path int true "Patient ID"
// @Success 200 {object} map[string]interface{} "Deleted"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/vhf/patients/{id} [delete]
func SwaggerDeleteVHFCase(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerVHFDeleteAPI(c, db, sl, store, config)
}

// GetUsers godoc
// @Summary Get all users
// @Description Paginated user list (users:read). Response shape: users (array), pagination { page, limit, total, total_pages }.
// @Tags Users
// @Produce json
// @Security SessionAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Page size" default(20)
// @Param search query string false "Search username, email, or name"
// @Param department_id query string false "Filter by department id"
// @Param role_id query string false "Filter by role id"
// @Param is_active query bool false "Filter active users"
// @Success 200 {object} map[string]interface{} "users and pagination"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/users [get]
func SwaggerGetUsers(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerUserListAPI(c, db, sl, store, config)
}

// GetEmployees godoc
// @Summary Get all employees
// @Description Returns `{ "employees": [...] }`. Requires `employees:read`.
// @Tags Employees
// @Produce json
// @Security SessionAuth
// @Success 200 {object} map[string]interface{} "employees array"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/employees [get]
func SwaggerGetEmployees(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerEmployeeListAPI(c, db, sl, store, config)
}

// GetInventoryItems godoc
// @Summary Get inventory items
// @Description Retrieve list of inventory items
// @Tags Inventory
// @Produce json
// @Security SessionAuth
// @Success 200 {array} map[string]interface{} "List of items"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/inventory/items [get]
func SwaggerGetInventoryItems(c *fiber.Ctx) error {
	// This would call the inventory handler
	return c.JSON(fiber.Map{"message": "Inventory items endpoint"})
}

// GetStockLevels godoc
// @Summary Get stock levels
// @Description Retrieve current stock levels for all items
// @Tags Inventory
// @Produce json
// @Security SessionAuth
// @Success 200 {array} map[string]interface{} "Stock levels"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/inventory/stock-levels [get]
func SwaggerGetStockLevels(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Stock levels endpoint"})
}

// GetAlerts godoc
// @Summary Get alerts
// @Description Retrieve list of alerts with pagination
// @Tags Alerts
// @Produce json
// @Security SessionAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{} "Paginated alerts"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/alerts [get]
func SwaggerGetAlerts(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerAlertsAPI(c, db, sl, store, config)
}

// GetRoles godoc
// @Summary Get all roles
// @Description Retrieve list of all RBAC roles
// @Tags RBAC
// @Produce json
// @Security SessionAuth
// @Success 200 {array} map[string]interface{} "List of roles"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/rbac/roles [get]
func SwaggerGetRoles(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	return HandlerGetRoles(c, db, sl)
}

// GetPermissions godoc
// @Summary Get all permissions
// @Description Retrieve list of all RBAC permissions
// @Tags RBAC
// @Produce json
// @Security SessionAuth
// @Success 200 {array} map[string]interface{} "List of permissions"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/rbac/permissions [get]
func SwaggerGetPermissions(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	return HandlerGetPermissions(c, db, sl)
}

// GetReports godoc
// @Summary Get quick statistics
// @Description Retrieve dashboard quick statistics
// @Tags Reports
// @Produce json
// @Security SessionAuth
// @Success 200 {object} map[string]interface{} "Statistics"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /reports/quick-stats [get]
func SwaggerGetReports(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Reports endpoint"})
}

// GetPillars godoc
// @Summary List resource pillars
// @Description Returns `{ "pillars": [...] }` (requires resource_management:read)
// @Tags Resource Management
// @Produce json
// @Security SessionAuth
// @Success 200 {object} map[string]interface{} "Pillars payload"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/pillars [get]
func SwaggerGetPillars(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementPillarsAPI(c, db, store)
}

// GetResourceManagementPillars godoc
// @Summary List resource pillars (alternate path)
// @Description Same as GET /api/pillars (`{ "pillars": [...] }`); session cookie + resource_management:read
// @Tags Resource Management
// @Produce json
// @Security SessionAuth
// @Success 200 {object} map[string]interface{} "Pillars payload"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/resource-management/pillars [get]
func SwaggerGetResourceManagementPillars(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementPillarsAPI(c, db, store)
}

// GetMpoxCases godoc
// @Summary Get Mpox cases
// @Description Retrieve list of Mpox cases
// @Tags Mpox
// @Produce json
// @Security SessionAuth
// @Param outbreak_id query int false "Filter by outbreak"
// @Success 200 {array} map[string]interface{} "List of Mpox cases"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/mpox/patients [get]
func SwaggerGetMpoxCases(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerMpoxListAPI(c, db, sl, store, config)
}

// GetMeaslesCases godoc
// @Summary Get Measles cases
// @Description Retrieve list of Measles cases
// @Tags Measles
// @Produce json
// @Security SessionAuth
// @Param outbreak_id query int false "Filter by outbreak"
// @Success 200 {array} map[string]interface{} "List of Measles cases"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/measles/patients [get]
func SwaggerGetMeaslesCases(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerMeaslesListAPI(c, db, sl, store, config)
}

// GetPolioCases godoc
// @Summary Get Polio cases
// @Description Retrieve list of Polio cases
// @Tags Polio
// @Produce json
// @Security SessionAuth
// @Param outbreak_id query int false "Filter by outbreak"
// @Success 200 {array} map[string]interface{} "List of Polio cases"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/polio/patients [get]
func SwaggerGetPolioCases(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerPolioListAPI(c, db, sl, store, config)
}

// --- VHF CIF bundle (JSON) ---

// GetVHFCIFByID godoc
// @Summary VHF CIF bundle by patient ID
// @Description Full CIF JSON: patient, clinical signs, hospitalization, risk factors, laboratory, investigator
// @Tags VHF
// @Produce json
// @Security SessionAuth
// @Param id path int true "VHF patient ID"
// @Success 200 {object} map[string]interface{} "CIF bundle"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/vhf/cif/{id} [get]
func SwaggerGetVHFCIFByID(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerVhfCIFByID(c, db, sl, store, config)
}

// GetVHFCIFByCaseCode godoc
// @Summary VHF CIF bundle by case code
// @Description Same bundle as by ID; pass case_code query parameter
// @Tags VHF
// @Produce json
// @Security SessionAuth
// @Param case_code query string true "Case code"
// @Success 200 {object} map[string]interface{} "CIF bundle"
// @Failure 400 {object} map[string]string "Missing or invalid case_code"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not found"
// @Router /api/vhf/cif [get]
func SwaggerGetVHFCIFByCaseCode(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerVhfCIFByCaseCode(c, db, sl, store, config)
}

// GetVHFClinicalSigns godoc
// @Summary Get VHF clinical signs
// @Tags VHF
// @Produce json
// @Security SessionAuth
// @Param id path int true "Patient ID"
// @Success 200 {object} map[string]interface{} "Clinical signs"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/vhf/patients/{id}/clinical-signs [get]
func SwaggerGetVHFClinicalSigns(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerVHFClinicalSignsAPI(c, db, sl, store, config)
}

// SaveVHFClinicalSigns godoc
// @Summary Save VHF clinical signs
// @Tags VHF
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path int true "Patient ID"
// @Param body body APIVHFClinicalSignsRequest true "Clinical signs payload"
// @Success 200 {object} map[string]interface{} "Saved"
// @Failure 400 {object} map[string]string "Invalid request"
// @Router /api/vhf/patients/{id}/clinical-signs [post]
func SwaggerSaveVHFClinicalSigns(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerVHFClinicalSignsSubmitAPI(c, db, sl, store, config)
}

// GetVHFHospitalization godoc
// @Summary Get VHF hospitalization
// @Tags VHF
// @Produce json
// @Security SessionAuth
// @Param id path int true "Patient ID"
// @Success 200 {object} map[string]interface{} "Hospitalization"
// @Router /api/vhf/patients/{id}/hospitalization [get]
func SwaggerGetVHFHospitalization(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerVHFHospitalizationAPI(c, db, sl, store, config)
}

// SaveVHFHospitalization godoc
// @Summary Save VHF hospitalization
// @Tags VHF
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path int true "Patient ID"
// @Param body body APIVHFHospitalizationRequest true "Hospitalization payload"
// @Success 200 {object} map[string]interface{} "Saved"
// @Router /api/vhf/patients/{id}/hospitalization [post]
func SwaggerSaveVHFHospitalization(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerVHFHospitalizationSubmitAPI(c, db, sl, store, config)
}

// GetVHFRiskFactors godoc
// @Summary Get VHF risk factors
// @Tags VHF
// @Produce json
// @Security SessionAuth
// @Param id path int true "Patient ID"
// @Success 200 {object} map[string]interface{} "Risk factors"
// @Router /api/vhf/patients/{id}/risk-factors [get]
func SwaggerGetVHFRiskFactors(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerVHFRiskFactorsAPI(c, db, sl, store, config)
}

// SaveVHFRiskFactors godoc
// @Summary Save VHF risk factors
// @Tags VHF
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path int true "Patient ID"
// @Param body body APIVHFRiskFactorsRequest true "Risk factors payload"
// @Success 200 {object} map[string]interface{} "Saved"
// @Router /api/vhf/patients/{id}/risk-factors [post]
func SwaggerSaveVHFRiskFactors(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerVHFRiskFactorsSubmitAPI(c, db, sl, store, config)
}

// GetVHFLaboratorySection godoc
// @Summary Get VHF laboratory (CIF section)
// @Tags VHF
// @Produce json
// @Security SessionAuth
// @Param id path int true "Patient ID"
// @Success 200 {object} map[string]interface{} "Laboratory"
// @Router /api/vhf/patients/{id}/laboratory [get]
func SwaggerGetVHFLaboratorySection(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerVHFLaboratoryAPI(c, db, sl, store, config)
}

// SaveVHFLaboratorySection godoc
// @Summary Save VHF laboratory (CIF section)
// @Tags VHF
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path int true "Patient ID"
// @Param body body APIVHFLaboratoryRequest true "Laboratory payload"
// @Success 200 {object} map[string]interface{} "Saved"
// @Router /api/vhf/patients/{id}/laboratory [post]
func SwaggerSaveVHFLaboratorySection(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config, sms *services.SMSService) error {
	return HandlerVHFLaboratorySubmitAPI(c, db, sl, store, config, sms)
}

// GetVHFInvestigator godoc
// @Summary Get VHF investigator
// @Tags VHF
// @Produce json
// @Security SessionAuth
// @Param id path int true "Patient ID"
// @Success 200 {object} map[string]interface{} "Investigator"
// @Router /api/vhf/patients/{id}/investigator [get]
func SwaggerGetVHFInvestigator(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerVHFInvestigatorAPI(c, db, sl, store, config)
}

// SaveVHFInvestigator godoc
// @Summary Save VHF investigator
// @Tags VHF
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path int true "Patient ID"
// @Param body body APIVHFInvestigatorRequest true "Investigator payload"
// @Success 200 {object} map[string]interface{} "Saved"
// @Router /api/vhf/patients/{id}/investigator [post]
func SwaggerSaveVHFInvestigator(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerVHFInvestigatorSubmitAPI(c, db, sl, store, config)
}

// GetVHFLabForm godoc
// @Summary Get VHF lab form (by case/patient id)
// @Tags VHF
// @Produce json
// @Security SessionAuth
// @Param id path int true "Case or patient ID"
// @Success 200 {object} map[string]interface{} "Lab form"
// @Router /api/vhf/lab/{id} [get]
func SwaggerGetVHFLabForm(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerVHFLabFormAPI(c, db, sl, store, config)
}

// SaveVHFLabForm godoc
// @Summary Submit VHF lab form
// @Tags VHF
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path int true "Case or patient ID"
// @Param body body APIVHFLaboratoryRequest true "Lab form payload"
// @Success 200 {object} map[string]interface{} "Saved"
// @Router /api/vhf/lab/{id} [post]
func SwaggerSaveVHFLabForm(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerVHFLabSaveAPI(c, db, sl, store, config)
}

// CreateMpoxCIFCase godoc
// @Summary Create Mpox CIF case
// @Description Submit a new Mpox CIF payload.
// @Tags Mpox
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param body body APIMpoxCIFRequest true "Mpox CIF payload"
// @Success 201 {object} map[string]interface{} "Created"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/mpox/patients [post]
func SwaggerCreateMpoxCIFCase(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	return HandlerMpoxCIFSubmitAPI(c, db, sl)
}

// CreateMeaslesCIFCase godoc
// @Summary Create Measles CIF case
// @Description Submit a new Measles CIF payload.
// @Tags Measles
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param body body APIMeaslesCIFRequest true "Measles CIF payload"
// @Success 201 {object} map[string]interface{} "Created"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/measles/patients [post]
func SwaggerCreateMeaslesCIFCase(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerMeaslesCIFAPI(c, db, store)
}

// CreatePolioCIFCase godoc
// @Summary Create Polio CIF case
// @Description Submit a new Polio CIF payload.
// @Tags Polio
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param body body APIPolioCIFRequest true "Polio CIF payload"
// @Success 201 {object} map[string]interface{} "Created"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/polio/patients [post]
func SwaggerCreatePolioCIFCase(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerPolioCIFSubmitAPI(c, db, sl, store, config)
}

// --- Users (CRUD + permissions) ---

// GetUser godoc
// @Summary Get user by ID
// @Tags Users
// @Produce json
// @Security SessionAuth
// @Param id path int true "User ID"
// @Success 200 {object} map[string]interface{} "User"
// @Router /api/users/{id} [get]
func SwaggerGetUser(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerGetUserAPI(c, db, sl, store, config)
}

// CreateUser godoc
// @Summary Create user
// @Tags Users
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param request body APIUserCreateRequest true "User + optional role_ids"
// @Success 201 {object} map[string]interface{} "Created"
// @Router /api/users [post]
func SwaggerCreateUser(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerUserSubmitAPI(c, db, sl, store, config)
}

// UpdateUser godoc
// @Summary Update user
// @Tags Users
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path int true "User ID"
// @Param request body APIUserUpdateRequest true "Fields to update"
// @Success 200 {object} map[string]interface{} "Updated"
// @Router /api/users/{id} [put]
func SwaggerUpdateUser(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerUserUpdateAPI(c, db, sl, store, config)
}

// DeleteUser godoc
// @Summary Delete user
// @Tags Users
// @Produce json
// @Security SessionAuth
// @Param id path int true "User ID"
// @Success 200 {object} map[string]interface{} "Deleted"
// @Router /api/users/{id} [delete]
func SwaggerDeleteUser(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerUserDeleteAPI(c, db, sl, store, config)
}

// GetUserPermissionsDoc godoc
// @Summary Effective permissions for a user
// @Description Permissions derived from role assignments
// @Tags Users
// @Produce json
// @Security SessionAuth
// @Param id path int true "User ID"
// @Success 200 {object} map[string]interface{} "permissions array"
// @Router /api/users/{id}/permissions [get]
func SwaggerGetUserPermissionsDoc(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	return HandlerGetUserPermissions(c, db, sl)
}

// AssignUserRoleApp godoc
// @Summary Assign a role to a user
// @Description Same handler as POST /api/rbac/user-roles; requires users:update
// @Tags Users
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param body body object true "user_id, role_id"
// @Success 200 {object} map[string]interface{} "OK"
// @Router /api/users/roles [post]
func SwaggerAssignUserRoleApp(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerAssignUserRole(c, db, sl, store, config)
}

// --- Employees ---

// GetEmployee godoc
// @Summary Get employee by ID
// @Description Returns `{ "employee": { ... } }` with nullable fields. Requires `employees:read`.
// @Tags Employees
// @Produce json
// @Security SessionAuth
// @Param id path int true "Employee ID"
// @Success 200 {object} map[string]interface{} "employee object"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Not found"
// @Router /api/employees/{id} [get]
func SwaggerGetEmployee(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerGetEmployeeAPI(c, db, sl, store, config)
}

// CreateEmployee godoc
// @Summary Create employee
// @Description Requires `employees:create`. Body field names match `models.Employee` JSON tags.
// @Tags Employees
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param body body APIEmployeeWriteRequest true "Employee"
// @Success 201 {object} map[string]interface{} "Created"
// @Failure 403 {object} map[string]string "Forbidden"
// @Router /api/employees [post]
func SwaggerCreateEmployee(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerEmployeeSubmitAPI(c, db, sl, store, config)
}

// UpdateEmployee godoc
// @Summary Update employee
// @Description Requires `employees:update`.
// @Tags Employees
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path int true "Employee ID"
// @Param body body APIEmployeeWriteRequest true "Employee"
// @Success 200 {object} map[string]interface{} "Updated"
// @Failure 403 {object} map[string]string "Forbidden"
// @Router /api/employees/{id} [put]
func SwaggerUpdateEmployee(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerEmployeeUpdateAPI(c, db, sl, store, config)
}

// DeleteEmployee godoc
// @Summary Delete employee
// @Description Requires `employees:delete`.
// @Tags Employees
// @Produce json
// @Security SessionAuth
// @Param id path int true "Employee ID"
// @Success 200 {object} map[string]interface{} "Deleted"
// @Failure 403 {object} map[string]string "Forbidden"
// @Router /api/employees/{id} [delete]
func SwaggerDeleteEmployee(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerDeleteEmployeeAPI(c, db, sl, store, config)
}

// --- Resource management (JSON lists) ---

// GetResourceManagementSummary godoc
// @Summary Resource management dashboard stats
// @Tags Resource Management
// @Produce json
// @Security SessionAuth
// @Success 200 {object} map[string]interface{} "Counts and summary fields"
// @Router /api/resource-management/summary [get]
func SwaggerGetResourceManagementSummary(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementSummaryAPI(c, db, store)
}

// GetResourceManagementRRTTeams godoc
// @Summary List RRT teams
// @Tags Resource Management
// @Produce json
// @Security SessionAuth
// @Success 200 {object} map[string]interface{} "teams"
// @Router /api/resource-management/rrt-teams [get]
func SwaggerGetResourceManagementRRTTeams(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementRRTTeamsAPI(c, db, store)
}

// GetResourceManagementRRTDeployments godoc
// @Summary List RRT deployments
// @Tags Resource Management
// @Produce json
// @Security SessionAuth
// @Success 200 {object} map[string]interface{} "deployments"
// @Router /api/resource-management/rrt-deployments [get]
func SwaggerGetResourceManagementRRTDeployments(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementRRTDeploymentsAPI(c, db, store)
}

// GetResourceManagementResources godoc
// @Summary List resources
// @Tags Resource Management
// @Produce json
// @Security SessionAuth
// @Success 200 {object} map[string]interface{} "resources"
// @Router /api/resource-management/resources [get]
func SwaggerGetResourceManagementResources(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementResourcesAPI(c, db, store)
}

// GetResourceManagementRequisitions godoc
// @Summary List requisitions
// @Tags Resource Management
// @Produce json
// @Security SessionAuth
// @Success 200 {object} map[string]interface{} "requisitions"
// @Router /api/resource-management/requisitions [get]
func SwaggerGetResourceManagementRequisitions(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementRequisitionsAPI(c, db, store)
}

// GetResourceManagementActivityLogs godoc
// @Summary List activity logs
// @Tags Resource Management
// @Produce json
// @Security SessionAuth
// @Success 200 {object} map[string]interface{} "activity_logs"
// @Router /api/resource-management/activity-logs [get]
func SwaggerGetResourceManagementActivityLogs(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementActivityLogsAPI(c, db, store)
}

// --- RBAC (admin / users) ---

// GetRBACStatsDoc godoc
// @Summary RBAC aggregate stats
// @Tags RBAC
// @Produce json
// @Security SessionAuth
// @Success 200 {object} map[string]interface{} "Stats"
// @Router /api/rbac/stats [get]
func SwaggerGetRBACStatsDoc(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	return HandlerGetRBACStats(c, db, sl)
}

// GetRBACMigrationStatus godoc
// @Summary RBAC migration status
// @Tags RBAC
// @Produce json
// @Security SessionAuth
// @Success 200 {object} map[string]interface{} "Status"
// @Router /api/rbac/migration-status [get]
func SwaggerGetRBACMigrationStatus(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	return HandlerGetMigrationStatus(c, db, sl)
}

// CreateRBACRole godoc
// @Summary Create role
// @Tags RBAC
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param body body object true "Role"
// @Success 201 {object} map[string]interface{} "Created"
// @Router /api/rbac/roles [post]
func SwaggerCreateRBACRole(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerCreateRole(c, db, sl, store, config)
}

// GetRBACRole godoc
// @Summary Get role by ID
// @Tags RBAC
// @Produce json
// @Security SessionAuth
// @Param id path int true "Role ID"
// @Success 200 {object} map[string]interface{} "Role"
// @Router /api/rbac/roles/{id} [get]
func SwaggerGetRBACRole(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	return HandlerGetRole(c, db, sl)
}

// UpdateRBACRole godoc
// @Summary Update role
// @Tags RBAC
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path int true "Role ID"
// @Param body body object true "Role"
// @Success 200 {object} map[string]interface{} "Updated"
// @Router /api/rbac/roles/{id} [put]
func SwaggerUpdateRBACRole(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerUpdateRole(c, db, sl, store, config)
}

// DeleteRBACRole godoc
// @Summary Delete role
// @Tags RBAC
// @Produce json
// @Security SessionAuth
// @Param id path int true "Role ID"
// @Success 200 {object} map[string]interface{} "Deleted"
// @Router /api/rbac/roles/{id} [delete]
func SwaggerDeleteRBACRole(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerDeleteRole(c, db, sl, store, config)
}

// CreateRBACPermission godoc
// @Summary Create permission
// @Tags RBAC
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param body body object true "Permission"
// @Success 201 {object} map[string]interface{} "Created"
// @Router /api/rbac/permissions [post]
func SwaggerCreateRBACPermission(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerCreatePermission(c, db, sl, store, config)
}

// GetRBACPermission godoc
// @Summary Get permission by ID
// @Tags RBAC
// @Produce json
// @Security SessionAuth
// @Param id path int true "Permission ID"
// @Success 200 {object} map[string]interface{} "Permission"
// @Router /api/rbac/permissions/{id} [get]
func SwaggerGetRBACPermission(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	return HandlerGetPermission(c, db, sl)
}

// UpdateRBACPermission godoc
// @Summary Update permission
// @Tags RBAC
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path int true "Permission ID"
// @Param body body object true "Permission"
// @Success 200 {object} map[string]interface{} "Updated"
// @Router /api/rbac/permissions/{id} [put]
func SwaggerUpdateRBACPermission(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerUpdatePermission(c, db, sl, store, config)
}

// DeleteRBACPermission godoc
// @Summary Delete permission
// @Tags RBAC
// @Produce json
// @Security SessionAuth
// @Param id path int true "Permission ID"
// @Success 200 {object} map[string]interface{} "Deleted"
// @Router /api/rbac/permissions/{id} [delete]
func SwaggerDeleteRBACPermission(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerDeletePermission(c, db, sl, store, config)
}

// ListRBACUsers godoc
// @Summary List users (RBAC path)
// @Description Same user list as user management; requires users:read
// @Tags RBAC
// @Produce json
// @Security SessionAuth
// @Success 200 {object} map[string]interface{} "Users"
// @Router /api/rbac/users [get]
func SwaggerListRBACUsers(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store) error {
	return HandlerGetUsers(c, db, sl, store)
}

// GetRBACUserRoles godoc
// @Summary Roles assigned to a user
// @Tags RBAC
// @Produce json
// @Security SessionAuth
// @Param user_id path int true "User ID"
// @Success 200 {object} map[string]interface{} "Roles"
// @Router /api/rbac/users/{user_id}/roles [get]
func SwaggerGetRBACUserRoles(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	return HandlerGetUserRoles(c, db, sl)
}

// UpdateRBACUserRoles godoc
// @Summary Replace all roles for a user
// @Tags RBAC
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param user_id path int true "User ID"
// @Param body body object true "role_ids"
// @Success 200 {object} map[string]interface{} "OK"
// @Router /api/rbac/users/{user_id}/roles [put]
func SwaggerUpdateRBACUserRoles(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerUpdateUserRoles(c, db, sl, store, config)
}

// BulkAssignRBACRoles godoc
// @Summary Bulk assign roles
// @Tags RBAC
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param body body object true "Bulk assign payload"
// @Success 200 {object} map[string]interface{} "OK"
// @Router /api/rbac/bulk-assign-roles [post]
func SwaggerBulkAssignRBACRoles(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerBulkAssignRoles(c, db, sl, store, config)
}

// AssignRBACUserRole godoc
// @Summary Assign one role to a user
// @Tags RBAC
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param body body object true "user_id, role_id"
// @Success 200 {object} map[string]interface{} "OK"
// @Router /api/rbac/user-roles [post]
func SwaggerAssignRBACUserRole(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerAssignUserRole(c, db, sl, store, config)
}

// RemoveRBACUserRole godoc
// @Summary Remove role from user
// @Tags RBAC
// @Produce json
// @Security SessionAuth
// @Param user_id path int true "User ID"
// @Param role_id path int true "Role ID"
// @Success 200 {object} map[string]interface{} "OK"
// @Router /api/rbac/user-roles/{user_id}/{role_id} [delete]
func SwaggerRemoveRBACUserRole(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerRemoveUserRole(c, db, sl, store, config)
}

// GetRBACRoleStatsDoc godoc
// @Summary Role assignment statistics
// @Tags RBAC
// @Produce json
// @Security SessionAuth
// @Success 200 {object} map[string]interface{} "Stats"
// @Router /api/rbac/role-stats [get]
func SwaggerGetRBACRoleStatsDoc(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	return HandlerGetRBACRoleStats(c, db, sl)
}

// GetRBACPermissionStatsDoc godoc
// @Summary Permission usage statistics
// @Tags RBAC
// @Produce json
// @Security SessionAuth
// @Success 200 {object} map[string]interface{} "Stats"
// @Router /api/rbac/permission-stats [get]
func SwaggerGetRBACPermissionStatsDoc(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	return HandlerGetRBACPermissionStats(c, db, sl)
}
