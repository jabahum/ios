package handlers

import (
	"database/sql"
	"log/slog"

	"case/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func newOutbreakAssignmentHandlerForAPI(db *sql.DB, store *session.Store) *OutbreakAssignmentHandler {
	return NewOutbreakAssignmentHandler(
		models.NewUserOutbreakService(db),
		models.NewPatientManagementRoleService(db),
		models.NewUserService(db),
		models.NewOutbreakService(db),
		models.NewFacilityService(db),
		store,
	)
}

// --- Outbreaks (CRUD + assignments; mirrors routes on app in routes.go) ---

// SwaggerGetOutbreak godoc
// @Summary Get outbreak by ID
// @Description Requires outbreaks:read and user access to the outbreak.
// @Tags Outbreaks
// @Produce json
// @Security SessionAuth
// @Param id path int true "Outbreak ID"
// @Success 200 {object} map[string]interface{} "outbreak object"
// @Router /api/outbreaks/{id} [get]
func SwaggerGetOutbreak(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerGetOutbreakAPI(c, db, sl, store, config)
}

// SwaggerCreateOutbreak godoc
// @Summary Create outbreak
// @Description Requires outbreaks:create. Body: name, start_date (YYYY-MM-DD), optional description, end_date, status, outbreak_type, outbreak_category.
// @Tags Outbreaks
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param request body APIOutbreakCreateRequest true "Outbreak"
// @Success 201 {object} map[string]interface{} "Created"
// @Router /api/outbreaks [post]
func SwaggerCreateOutbreak(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerOutbreakSubmitAPI(c, db, sl, store, config)
}

// SwaggerUpdateOutbreak godoc
// @Summary Update outbreak
// @Description Requires outbreaks:update and manage permission on the outbreak.
// @Tags Outbreaks
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path int true "Outbreak ID"
// @Param request body APIOutbreakCreateRequest true "Fields to update"
// @Success 200 {object} map[string]interface{} "OK"
// @Router /api/outbreaks/{id} [put]
func SwaggerUpdateOutbreak(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerOutbreakUpdateAPI(c, db, sl, store, config)
}

// SwaggerDeleteOutbreak godoc
// @Summary Delete outbreak
// @Description Requires outbreaks:delete and manage permission.
// @Tags Outbreaks
// @Produce json
// @Security SessionAuth
// @Param id path int true "Outbreak ID"
// @Success 200 {object} map[string]interface{} "OK"
// @Router /api/outbreaks/{id} [delete]
func SwaggerDeleteOutbreak(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerOutbreakDeleteAPI(c, db, sl, store, config)
}

// SwaggerCloseOutbreak godoc
// @Summary Close outbreak (status=closed)
// @Tags Outbreaks
// @Produce json
// @Security SessionAuth
// @Param id path int true "Outbreak ID"
// @Success 200 {object} map[string]interface{} "OK"
// @Router /api/outbreaks/{id}/close [post]
func SwaggerCloseOutbreak(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerOutbreakCloseAPI(c, db, sl, store, config)
}

// SwaggerSelectOutbreak godoc
// @Summary Set session selected outbreak
// @Tags Outbreaks
// @Produce json
// @Security SessionAuth
// @Param id path int true "Outbreak ID"
// @Success 200 {object} map[string]interface{} "OK"
// @Router /api/outbreaks/{id}/select [post]
func SwaggerSelectOutbreak(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerOutbreakSelectAPI(c, db, sl, store, config)
}

// SwaggerListOutbreakAssignments godoc
// @Summary List user–outbreak assignments
// @Tags Outbreaks
// @Produce json
// @Security SessionAuth
// @Success 200 {object} map[string]interface{} "assignments"
// @Router /api/outbreaks/assignments [get]
func SwaggerListOutbreakAssignments(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store) error {
	return newOutbreakAssignmentHandlerForAPI(db, store).ShowOutbreakAssignmentsAPI(c)
}

// SwaggerAssignUserToOutbreak godoc
// @Summary Assign user to outbreak
// @Description Requires outbreaks:update. Body: outbreak_id, user_id.
// @Tags Outbreaks
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param request body APIOutbreakAssignRequest true "Assignment"
// @Success 200 {object} map[string]interface{} "OK"
// @Router /api/outbreaks/assign [post]
func SwaggerAssignUserToOutbreak(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return newOutbreakAssignmentHandlerForAPI(db, store).HandleAssignFormSubmissionAPI(c)
}

// SwaggerRemoveUserFromOutbreak godoc
// @Summary Remove user from outbreak
// @Tags Outbreaks
// @Produce json
// @Security SessionAuth
// @Param outbreak_id path int true "Outbreak ID"
// @Param user_id path int true "User ID"
// @Success 200 {object} map[string]interface{} "OK"
// @Router /api/outbreaks/{outbreak_id}/users/{user_id} [delete]
func SwaggerRemoveUserFromOutbreak(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return newOutbreakAssignmentHandlerForAPI(db, store).RemoveUserFromOutbreakAPI(c)
}

// SwaggerListUsersForOutbreak godoc
// @Summary Users assigned to an outbreak
// @Tags Outbreaks
// @Produce json
// @Security SessionAuth
// @Param id path int true "Outbreak ID"
// @Success 200 {object} map[string]interface{} "users"
// @Router /api/outbreaks/{id}/users [get]
func SwaggerListUsersForOutbreak(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store) error {
	return HandlerOutbreakUsersAPI(c, db, sl, store)
}

// --- Resource management CRUD (JSON) ---

// SwaggerGetResourceManagementPillarByID godoc
// @Summary Get pillar by ID
// @Tags Resource Management
// @Produce json
// @Security SessionAuth
// @Param id path int true "Pillar ID"
// @Success 200 {object} map[string]interface{} "pillar"
// @Router /api/resource-management/pillars/{id} [get]
func SwaggerGetResourceManagementPillarByID(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementPillarGetAPI(c, db, store)
}

// SwaggerCreateResourceManagementPillar godoc
// @Summary Create pillar
// @Tags Resource Management
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param body body APIResourceManagementPillarWrite true "Pillar"
// @Success 201 {object} map[string]interface{} "Created"
// @Router /api/resource-management/pillars [post]
func SwaggerCreateResourceManagementPillar(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementPillarCreateAPI(c, db, store)
}

// SwaggerUpdateResourceManagementPillar godoc
// @Summary Update pillar
// @Tags Resource Management
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path int true "Pillar ID"
// @Param body body APIResourceManagementPillarWrite true "Pillar"
// @Success 200 {object} map[string]interface{} "OK"
// @Router /api/resource-management/pillars/{id} [put]
func SwaggerUpdateResourceManagementPillar(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementPillarUpdateAPI(c, db, store)
}

// SwaggerDeleteResourceManagementPillar godoc
// @Summary Delete pillar
// @Tags Resource Management
// @Produce json
// @Security SessionAuth
// @Param id path int true "Pillar ID"
// @Success 200 {object} map[string]interface{} "OK"
// @Router /api/resource-management/pillars/{id} [delete]
func SwaggerDeleteResourceManagementPillar(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementPillarDeleteAPI(c, db, store)
}

// SwaggerGetResourceManagementRRTTeamByID godoc
// @Summary Get RRT team by ID
// @Tags Resource Management
// @Produce json
// @Security SessionAuth
// @Param id path int true "Team ID"
// @Success 200 {object} map[string]interface{} "team"
// @Router /api/resource-management/rrt-teams/{id} [get]
func SwaggerGetResourceManagementRRTTeamByID(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementRRTTeamGetAPI(c, db, store)
}

// SwaggerCreateResourceManagementRRTTeam godoc
// @Summary Create RRT team
// @Tags Resource Management
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param body body APIResourceManagementRRTTeamWrite true "RRT team (team_type and team_lead_name required by DB)"
// @Success 201 {object} map[string]interface{} "Created"
// @Router /api/resource-management/rrt-teams [post]
func SwaggerCreateResourceManagementRRTTeam(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementRRTTeamCreateAPI(c, db, store)
}

// SwaggerUpdateResourceManagementRRTTeam godoc
// @Summary Update RRT team
// @Tags Resource Management
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path int true "Team ID"
// @Param body body APIResourceManagementRRTTeamWrite true "RRT team"
// @Success 200 {object} map[string]interface{} "OK"
// @Router /api/resource-management/rrt-teams/{id} [put]
func SwaggerUpdateResourceManagementRRTTeam(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementRRTTeamUpdateAPI(c, db, store)
}

// SwaggerDeleteResourceManagementRRTTeam godoc
// @Summary Delete RRT team
// @Tags Resource Management
// @Produce json
// @Security SessionAuth
// @Param id path int true "Team ID"
// @Success 200 {object} map[string]interface{} "OK"
// @Router /api/resource-management/rrt-teams/{id} [delete]
func SwaggerDeleteResourceManagementRRTTeam(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementRRTTeamDeleteAPI(c, db, store)
}

// SwaggerGetResourceManagementRRTDeploymentByID godoc
// @Summary Get RRT deployment by ID
// @Tags Resource Management
// @Produce json
// @Security SessionAuth
// @Param id path int true "Deployment ID"
// @Success 200 {object} map[string]interface{} "deployment"
// @Router /api/resource-management/rrt-deployments/{id} [get]
func SwaggerGetResourceManagementRRTDeploymentByID(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementRRTDeploymentGetAPI(c, db, store)
}

// SwaggerCreateResourceManagementRRTDeployment godoc
// @Summary Create RRT deployment
// @Tags Resource Management
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param body body APIResourceManagementRRTDeploymentWrite true "Deployment"
// @Success 201 {object} map[string]interface{} "Created"
// @Router /api/resource-management/rrt-deployments [post]
func SwaggerCreateResourceManagementRRTDeployment(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementRRTDeploymentCreateAPI(c, db, store)
}

// SwaggerUpdateResourceManagementRRTDeployment godoc
// @Summary Update RRT deployment
// @Tags Resource Management
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path int true "Deployment ID"
// @Param body body APIResourceManagementRRTDeploymentWrite true "Deployment"
// @Success 200 {object} map[string]interface{} "OK"
// @Router /api/resource-management/rrt-deployments/{id} [put]
func SwaggerUpdateResourceManagementRRTDeployment(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementRRTDeploymentUpdateAPI(c, db, store)
}

// SwaggerDeleteResourceManagementRRTDeployment godoc
// @Summary Delete RRT deployment
// @Tags Resource Management
// @Produce json
// @Security SessionAuth
// @Param id path int true "Deployment ID"
// @Success 200 {object} map[string]interface{} "OK"
// @Router /api/resource-management/rrt-deployments/{id} [delete]
func SwaggerDeleteResourceManagementRRTDeployment(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementRRTDeploymentDeleteAPI(c, db, store)
}

// SwaggerGetResourceManagementResourceByID godoc
// @Summary Get resource catalog item by ID
// @Tags Resource Management
// @Produce json
// @Security SessionAuth
// @Param id path int true "Resource ID"
// @Success 200 {object} map[string]interface{} "resource"
// @Router /api/resource-management/resources/{id} [get]
func SwaggerGetResourceManagementResourceByID(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementResourceGetAPI(c, db, store)
}

// SwaggerCreateResourceManagementResource godoc
// @Summary Create resource catalog item
// @Tags Resource Management
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param body body APIResourceManagementResourceWrite true "Resource"
// @Success 201 {object} map[string]interface{} "Created"
// @Router /api/resource-management/resources [post]
func SwaggerCreateResourceManagementResource(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementResourceCreateAPI(c, db, store)
}

// SwaggerUpdateResourceManagementResource godoc
// @Summary Update resource catalog item
// @Tags Resource Management
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path int true "Resource ID"
// @Param body body APIResourceManagementResourceWrite true "Resource"
// @Success 200 {object} map[string]interface{} "OK"
// @Router /api/resource-management/resources/{id} [put]
func SwaggerUpdateResourceManagementResource(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementResourceUpdateAPI(c, db, store)
}

// SwaggerDeleteResourceManagementResource godoc
// @Summary Delete resource catalog item
// @Tags Resource Management
// @Produce json
// @Security SessionAuth
// @Param id path int true "Resource ID"
// @Success 200 {object} map[string]interface{} "OK"
// @Router /api/resource-management/resources/{id} [delete]
func SwaggerDeleteResourceManagementResource(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementResourceDeleteAPI(c, db, store)
}

// SwaggerListResourceManagementResourceCategories godoc
// @Summary List resource categories
// @Description Use category `id` values when creating catalog resources.
// @Tags Resource Management
// @Produce json
// @Security SessionAuth
// @Success 200 {object} map[string]interface{} "{ \"resource_categories\": [...] }"
// @Router /api/resource-management/resource-categories [get]
func SwaggerListResourceManagementResourceCategories(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementResourceCategoriesAPI(c, db, store)
}

// SwaggerGetResourceManagementResourceCategory godoc
// @Summary Get resource category by ID
// @Tags Resource Management
// @Produce json
// @Security SessionAuth
// @Param id path int true "Category ID"
// @Success 200 {object} map[string]interface{} "{ \"resource_category\": {...} }"
// @Router /api/resource-management/resource-categories/{id} [get]
func SwaggerGetResourceManagementResourceCategory(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementResourceCategoryGetAPI(c, db, store)
}

// SwaggerCreateResourceManagementResourceCategory godoc
// @Summary Create resource category
// @Tags Resource Management
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param body body APIResourceCategoryWrite true "Category"
// @Success 201 {object} map[string]interface{} "Created"
// @Router /api/resource-management/resource-categories [post]
func SwaggerCreateResourceManagementResourceCategory(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementResourceCategoryCreateAPI(c, db, store)
}

// SwaggerGetResourceManagementRequisitionByID godoc
// @Summary Get requisition by ID
// @Tags Resource Management
// @Produce json
// @Security SessionAuth
// @Param id path int true "Requisition ID"
// @Success 200 {object} map[string]interface{} "requisition"
// @Router /api/resource-management/requisitions/{id} [get]
func SwaggerGetResourceManagementRequisitionByID(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementRequisitionGetAPI(c, db, store)
}

// SwaggerCreateResourceManagementRequisition godoc
// @Summary Create requisition
// @Tags Resource Management
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param body body APIResourceManagementRequisitionWrite true "Requisition"
// @Success 201 {object} map[string]interface{} "Created"
// @Router /api/resource-management/requisitions [post]
func SwaggerCreateResourceManagementRequisition(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementRequisitionCreateAPI(c, db, store)
}

// SwaggerUpdateResourceManagementRequisition godoc
// @Summary Update requisition
// @Tags Resource Management
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path int true "Requisition ID"
// @Param body body APIResourceManagementRequisitionWrite true "Requisition"
// @Success 200 {object} map[string]interface{} "OK"
// @Router /api/resource-management/requisitions/{id} [put]
func SwaggerUpdateResourceManagementRequisition(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementRequisitionUpdateAPI(c, db, store)
}

// SwaggerDeleteResourceManagementRequisition godoc
// @Summary Delete requisition
// @Tags Resource Management
// @Produce json
// @Security SessionAuth
// @Param id path int true "Requisition ID"
// @Success 200 {object} map[string]interface{} "OK"
// @Router /api/resource-management/requisitions/{id} [delete]
func SwaggerDeleteResourceManagementRequisition(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementRequisitionDeleteAPI(c, db, store)
}

// SwaggerGetResourceManagementActivityLogByID godoc
// @Summary Get activity log by ID
// @Tags Resource Management
// @Produce json
// @Security SessionAuth
// @Param id path int true "Activity log ID"
// @Success 200 {object} map[string]interface{} "activity_log"
// @Router /api/resource-management/activity-logs/{id} [get]
func SwaggerGetResourceManagementActivityLogByID(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementActivityLogGetAPI(c, db, store)
}

// SwaggerCreateResourceManagementActivityLog godoc
// @Summary Create activity log
// @Tags Resource Management
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param body body APIResourceManagementActivityLogWrite true "Activity log"
// @Success 201 {object} map[string]interface{} "Created"
// @Router /api/resource-management/activity-logs [post]
func SwaggerCreateResourceManagementActivityLog(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementActivityLogCreateAPI(c, db, store)
}

// SwaggerUpdateResourceManagementActivityLog godoc
// @Summary Update activity log
// @Tags Resource Management
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path int true "Activity log ID"
// @Param body body APIResourceManagementActivityLogWrite true "Activity log"
// @Success 200 {object} map[string]interface{} "OK"
// @Router /api/resource-management/activity-logs/{id} [put]
func SwaggerUpdateResourceManagementActivityLog(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementActivityLogUpdateAPI(c, db, store)
}

// SwaggerDeleteResourceManagementActivityLog godoc
// @Summary Delete activity log
// @Tags Resource Management
// @Produce json
// @Security SessionAuth
// @Param id path int true "Activity log ID"
// @Success 200 {object} map[string]interface{} "OK"
// @Router /api/resource-management/activity-logs/{id} [delete]
func SwaggerDeleteResourceManagementActivityLog(c *fiber.Ctx, db *sql.DB, store *session.Store) error {
	return HandlerResourceManagementActivityLogDeleteAPI(c, db, store)
}

// --- Departments (CRUD; admin permissions, same as /api/roles) ---

// SwaggerListDepartments godoc
// @Summary List departments
// @Description Requires admin:read
// @Tags Departments
// @Produce json
// @Security SessionAuth
// @Success 200 {object} map[string]interface{} "{ \"departments\": [...] }"
// @Router /api/departments [get]
func SwaggerListDepartments(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerDepartmentListAPI(c, db, sl)
}

// SwaggerGetDepartment godoc
// @Summary Get department by ID
// @Description Requires admin:read
// @Tags Departments
// @Produce json
// @Security SessionAuth
// @Param id path int true "Department ID"
// @Success 200 {object} map[string]interface{} "{ \"department\": {...} }"
// @Router /api/departments/{id} [get]
func SwaggerGetDepartment(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerDepartmentGetAPI(c, db, sl)
}

// SwaggerCreateDepartment godoc
// @Summary Create department
// @Description Requires admin:create
// @Tags Departments
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param body body APIDepartmentWriteRequest true "Department"
// @Success 201 {object} map[string]interface{} "Created"
// @Router /api/departments [post]
func SwaggerCreateDepartment(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerDepartmentCreateAPI(c, db, sl)
}

// SwaggerUpdateDepartment godoc
// @Summary Update department
// @Description Requires admin:update
// @Tags Departments
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path int true "Department ID"
// @Param body body APIDepartmentWriteRequest true "Department"
// @Success 200 {object} map[string]interface{} "OK"
// @Router /api/departments/{id} [put]
func SwaggerUpdateDepartment(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerDepartmentUpdateAPI(c, db, sl)
}

// SwaggerDeleteDepartment godoc
// @Summary Delete department
// @Description Requires admin:delete. Fails with 409 if still referenced.
// @Tags Departments
// @Produce json
// @Security SessionAuth
// @Param id path int true "Department ID"
// @Success 200 {object} map[string]interface{} "OK"
// @Router /api/departments/{id} [delete]
func SwaggerDeleteDepartment(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	return HandlerDepartmentDeleteAPI(c, db, sl)
}
