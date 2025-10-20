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
// @Description Authenticate user with username and password
// @Tags Authentication
// @Accept json,application/x-www-form-urlencoded
// @Produce json,html
// @Param username formData string true "Username"
// @Param password formData string true "Password"
// @Success 200 {object} map[string]interface{} "Login successful"
// @Failure 400 {object} map[string]string "Invalid credentials"
// @Failure 401 {object} map[string]string "Unauthorized"
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
// @Description Retrieve list of all outbreaks
// @Tags Outbreaks
// @Produce json
// @Security SessionAuth
// @Success 200 {array} map[string]interface{} "List of outbreaks"
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
// @Param case body object true "VHF case data"
// @Success 201 {object} map[string]interface{} "Case created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/vhf/patients [post]
func SwaggerCreateVHFCase(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config, sms *services.SMSService) error {
	return HandlerVHFPatientSubmitAPI(c, db, sl, store, config, sms)
}

// GetUsers godoc
// @Summary Get all users
// @Description Retrieve list of all users
// @Tags Users
// @Produce json
// @Security SessionAuth
// @Success 200 {array} map[string]interface{} "List of users"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/users [get]
func SwaggerGetUsers(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store) error {
	return HandlerGetUsers(c, db, sl, store)
}

// GetEmployees godoc
// @Summary Get all employees
// @Description Retrieve list of all employees
// @Tags Employees
// @Produce json
// @Security SessionAuth
// @Success 200 {array} map[string]interface{} "List of employees"
// @Failure 401 {object} map[string]string "Unauthorized"
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
// @Summary Get pillars
// @Description Retrieve list of all resource management pillars
// @Tags Resource Management
// @Produce json
// @Security SessionAuth
// @Success 200 {array} map[string]interface{} "List of pillars"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/pillars [get]
func SwaggerGetPillars(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Pillars endpoint"})
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
