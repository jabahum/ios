package routes

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"strings"
	"time"
	"path/filepath"
	"net/url"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	_ "github.com/lib/pq"

	fiberSwagger "github.com/arsmn/fiber-swagger/v2"

	"case/internal/handlers"
	"case/internal/middleware"
	"case/internal/models"
	"case/internal/reports"
	"case/internal/services"
)

func SetRoute(app *fiber.App, db *sql.DB, store *session.Store, sl *slog.Logger, config handlers.Config, smsService *services.SMSService, voiceService *services.VoiceService) {

	// Swagger documentation routes
	app.Get("/swagger/*", fiberSwagger.HandlerDefault)
	app.Get("/api/docs/*", fiberSwagger.HandlerDefault)

	// Public routes
	app.Get("/", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "landing")
	})

	//survey routes
	app.Get("/surveys", func(c *fiber.Ctx) error {
		return handlers.HandlerGetSurveys(c, db, sl, store)
	})
	// IVR voice calls routes
	app.Get("/call", func(c *fiber.Ctx) error {
		return models.SendCall(c, db, sl)
	})
	app.Post("/voice/callback", func(c *fiber.Ctx) error {
		return models.HandleVoiceCallback(c, db)
	})
	app.Get("/audios/*", func(c *fiber.Ctx) error {
		// 1. Get the current working directory to anchor the search
		wd, err := os.Getwd()
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Could not determine working directory")
		}

		// 2. Capture the full wildcard path (e.g., "english/01.wav")
		// Fiber uses "*" as the parameter name for the wildcard
		rawPath := c.Params("*")
		filename, _ := url.QueryUnescape(rawPath)
		
		log.Printf("DEBUG: Requested path: %s", filename)

		// 3. Construct the absolute path
		// We join: [Current Working Dir] + [audios folder] + [the wildcard path]
		// use this one for production.
		absPath := filepath.Join(wd, "../..", "audios", filepath.Clean(filename))

		// @jkage: use this one for development on docker
		// absPath := filepath.Join(wd, "audios", filepath.Clean(filename))
		
		fmt.Println("Full system path to file:", absPath)

		// 4. Set Headers for Africa's Talking (No-Cache)
		c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")

		// 5. Verify file existence before sending
		if info, err := os.Stat(absPath); os.IsNotExist(err) || info.IsDir() {
			log.Printf("CRITICAL: File missing or is a directory: %s", absPath)
			return fiber.NewError(fiber.StatusNotFound, "Audio file not found")
		}

		// 6. Send the file
		// Fiber handles Content-Type (audio/wav) and Range requests automatically
		if err := c.SendFile(absPath); err != nil {
			log.Printf("ERROR: Failed to send audio file: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Error transmitting file")
		}

		return nil
	})
	app.Get("/login", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "login")
	})
	app.Post("/login", func(c *fiber.Ctx) error {
		return handlers.HandlerLoginSubmit(c, db, sl, store, config)
	})
	app.Get("/logout", func(c *fiber.Ctx) error {
		return handlers.HandlerLoginOut(c, sl, store, config)
	})
	app.Get("/vhf-cif", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "vhf_cif")
	})
	app.Post("/vhf-cif/save", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFPatientSubmit(c, db, sl, store, config, smsService)
	})
	app.Post("/vhf-cif/update", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFUpdate(c, db, sl, store, config)
	})
	app.Get("/vhf-cif/clinical-signs/:id", func(c *fiber.Ctx) error {
		data := handlers.NewTemplateData(c, store)
		data.Form = fiber.Map{"PatientID": c.Params("id")}
		return handlers.GenerateHTML(c, db, data, "vhf_clinical_signs")
	})
	app.Post("/vhf-cif/clinical-signs/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFClinicalSignsSubmit(c, db, sl, store, config)
	})
	app.Get("/vhf-cif/hospitalization/:id", func(c *fiber.Ctx) error {
		data := handlers.NewTemplateData(c, store)
		data.Form = fiber.Map{"PatientID": c.Params("id")}
		return handlers.GenerateHTML(c, db, data, "vhf_hospitalization")
	})
	app.Post("/vhf-cif/hospitalization/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFHospitalizationSubmit(c, db, sl, store, config)
	})
	app.Get("/vhf-cif/risk-factors/:id", func(c *fiber.Ctx) error {
		data := handlers.NewTemplateData(c, store)
		data.Form = fiber.Map{"PatientID": c.Params("id")}
		return handlers.GenerateHTML(c, db, data, "vhf_risk_factors")
	})
	app.Post("/vhf-cif/risk-factors/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFRiskFactorsSubmit(c, db, sl, store, config)
	})
	app.Get("/vhf-cif/laboratory/:id", func(c *fiber.Ctx) error {
		data := handlers.NewTemplateData(c, store)
		data.Form = fiber.Map{"PatientID": c.Params("id")}
		return handlers.GenerateHTML(c, db, data, "vhf_laboratory")
	})
	app.Post("/vhf-cif/laboratory/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFLaboratorySubmit(c, db, sl, store, config, smsService)
	})
	app.Get("/vhf-cif/investigator/:id", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "vhf_investigator")
	})
	app.Post("/vhf-cif/investigator/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFInvestigatorSubmit(c, db, sl, store, config)
	})
	app.Get("/vhf-cif/success", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFSuccess(c, db, sl, store, config)
	})
	app.Get("/vhf-cif/view/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFView(c, db, sl, store, config)
	})
	app.Get("/measles_success", func(c *fiber.Ctx) error { return handlers.HandlerMeaslesSuccess(c, store) })
	app.Get("/measles_cif", func(c *fiber.Ctx) error {
		return handlers.HandlerMeaslesCIF(c, db, store)
	})
	app.Post("/measles_cif", func(c *fiber.Ctx) error {
		return handlers.HandlerMeaslesCIF(c, db, store)
	})
	app.Get("/measles_cif/view/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerMeaslesCIFView(c, db, sl, store, config)
	})
	app.Get("/measles_cif/edit/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerMeaslesCIFEdit(c, db, sl, store, config)
	})

	// Polio CIF routes (public) - must be defined BEFORE the protected routes group
	app.Get("/polio-cif", func(c *fiber.Ctx) error {
		return handlers.HandlerPolioCIF(c, db, sl, store, config)
	})
	app.Post("/polio-cif/save", func(c *fiber.Ctx) error {
		return handlers.HandlerPolioCIFSubmit(c, db, sl, store, config)
	})
	app.Get("/polio-cif/success", func(c *fiber.Ctx) error {
		return handlers.HandlerPolioCIFSuccess(c, db, sl, store, config)
	})
	app.Get("/polio-cif/view/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerPolioCIFView(c, db, sl, store, config)
	})
	app.Get("/polio-cif/edit/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerPolioCIFEdit(c, db, sl, store, config)
	})

	// Location API routes
	app.Get("/api/locations/districts", func(c *fiber.Ctx) error {
		return handlers.HandlerGetDistricts(c, db, sl)
	})
	app.Get("/api/locations/subcounties/:district_id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetSubcountiesByDistrict(c, db, sl)
	})
	app.Get("/api/locations/parishes/:subcounty_id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetParishesBySubcounty(c, db, sl)
	})
	app.Get("/api/locations/parishes/district/:district_id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetParishesByDistrict(c, db, sl)
	})
	app.Get("/api/locations/villages/:parish_id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetVillagesByParish(c, db, sl)
	})
	app.Get("/api/locations/villages/district/:district_id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetVillagesByDistrict(c, db, sl)
	})
	app.Get("/api/locations/villages/subcounty/:subcounty_id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetVillagesBySubcounty(c, db, sl)
	})

	// Facilities API route
	app.Get("/api/facilities", func(c *fiber.Ctx) error {
		return handlers.HandlerGetFacilities(c, db, sl)
	})

	// Outbreaks API route
	app.Get("/api/outbreaks", func(c *fiber.Ctx) error {
		return handlers.HandlerGetOutbreaksAPI(c, db, sl, store)
	})

	// Backward-compatible CIF aliases (public API style to match existing clients)
	app.Get("/api/cif", func(c *fiber.Ctx) error {
		return handlers.HandlerVhfCIFByCaseCode(c, db, sl, store, config)
	})
	app.Get("/api/cif/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVhfCIFByID(c, db, sl, store, config)
	})

	// Public alias for creating cases to avoid 404 during Admit flow
	app.Post("/api/cases", func(c *fiber.Ctx) error {
		return handlers.HandlerCasesSubmitAPI(c, db, sl, store, config)
	})

	// RBAC API endpoints - Admin only access
	app.Get("/api/roles", middleware.PermissionRequired(store, db, sl, "admin", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerGetRoles(c, db, sl)
	})
	app.Get("/api/roles/:id", middleware.PermissionRequired(store, db, sl, "admin", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerGetRole(c, db, sl)
	})
	app.Put("/api/roles/:id", middleware.PermissionRequired(store, db, sl, "admin", "update"), func(c *fiber.Ctx) error {
		return handlers.HandlerUpdateRole(c, db, sl, store, config)
	})
	app.Delete("/api/roles/:id", middleware.PermissionRequired(store, db, sl, "admin", "delete"), func(c *fiber.Ctx) error {
		return handlers.HandlerDeleteRole(c, db, sl, store, config)
	})
	app.Get("/api/permissions", middleware.PermissionRequired(store, db, sl, "admin", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerGetPermissions(c, db, sl)
	})
	app.Get("/api/permissions/:id", middleware.PermissionRequired(store, db, sl, "admin", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerGetPermission(c, db, sl)
	})
	app.Put("/api/permissions/:id", middleware.PermissionRequired(store, db, sl, "admin", "update"), func(c *fiber.Ctx) error {
		return handlers.HandlerUpdatePermission(c, db, sl, store, config)
	})
	app.Delete("/api/permissions/:id", middleware.PermissionRequired(store, db, sl, "admin", "delete"), func(c *fiber.Ctx) error {
		return handlers.HandlerDeletePermission(c, db, sl, store, config)
	})
	app.Get("/api/rbac/migration-status", middleware.PermissionRequired(store, db, sl, "admin", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerGetMigrationStatus(c, db, sl)
	})

	// RBAC API Routes - Admin only access
	app.Get("/api/rbac/stats", middleware.PermissionRequired(store, db, sl, "admin", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerGetRBACStats(c, db, sl)
	})
	app.Get("/api/rbac/roles", middleware.PermissionRequired(store, db, sl, "admin", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerGetRoles(c, db, sl)
	})
	app.Post("/api/rbac/roles", middleware.PermissionRequired(store, db, sl, "admin", "create"), func(c *fiber.Ctx) error {
		return handlers.HandlerCreateRole(c, db, sl, store, config)
	})
	app.Get("/api/rbac/roles/:id", middleware.PermissionRequired(store, db, sl, "admin", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerGetRole(c, db, sl)
	})
	app.Put("/api/rbac/roles/:id", middleware.PermissionRequired(store, db, sl, "admin", "update"), func(c *fiber.Ctx) error {
		return handlers.HandlerUpdateRole(c, db, sl, store, config)
	})
	app.Delete("/api/rbac/roles/:id", middleware.PermissionRequired(store, db, sl, "admin", "delete"), func(c *fiber.Ctx) error {
		return handlers.HandlerDeleteRole(c, db, sl, store, config)
	})

	app.Get("/api/rbac/permissions", middleware.PermissionRequired(store, db, sl, "admin", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerGetPermissions(c, db, sl)
	})
	app.Post("/api/rbac/permissions", middleware.PermissionRequired(store, db, sl, "admin", "create"), func(c *fiber.Ctx) error {
		return handlers.HandlerCreatePermission(c, db, sl, store, config)
	})
	app.Get("/api/rbac/permissions/:id", middleware.PermissionRequired(store, db, sl, "admin", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerGetPermission(c, db, sl)
	})
	app.Put("/api/rbac/permissions/:id", middleware.PermissionRequired(store, db, sl, "admin", "update"), func(c *fiber.Ctx) error {
		return handlers.HandlerUpdatePermission(c, db, sl, store, config)
	})
	app.Delete("/api/rbac/permissions/:id", middleware.PermissionRequired(store, db, sl, "admin", "delete"), func(c *fiber.Ctx) error {
		return handlers.HandlerDeletePermission(c, db, sl, store, config)
	})

	app.Get("/api/rbac/users", middleware.PermissionRequired(store, db, sl, "users", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerGetUsers(c, db, sl, store)
	})
	app.Get("/api/rbac/users/:user_id/roles", middleware.PermissionRequired(store, db, sl, "users", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerGetUserRoles(c, db, sl)
	})
	app.Put("/api/rbac/users/:user_id/roles", middleware.PermissionRequired(store, db, sl, "users", "update"), func(c *fiber.Ctx) error {
		return handlers.HandlerUpdateUserRoles(c, db, sl, store, config)
	})
	app.Post("/api/rbac/bulk-assign-roles", middleware.PermissionRequired(store, db, sl, "users", "update"), func(c *fiber.Ctx) error {
		return handlers.HandlerBulkAssignRoles(c, db, sl, store, config)
	})
	app.Post("/api/rbac/user-roles", middleware.PermissionRequired(store, db, sl, "users", "update"), func(c *fiber.Ctx) error {
		return handlers.HandlerAssignUserRole(c, db, sl, store, config)
	})
	app.Delete("/api/rbac/user-roles/:user_id/:role_id", middleware.PermissionRequired(store, db, sl, "users", "update"), func(c *fiber.Ctx) error {
		return handlers.HandlerRemoveUserRole(c, db, sl, store, config)
	})
	app.Get("/api/rbac/role-stats", middleware.PermissionRequired(store, db, sl, "admin", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerGetRBACRoleStats(c, db, sl)
	})
	app.Get("/api/rbac/permission-stats", middleware.PermissionRequired(store, db, sl, "admin", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerGetRBACPermissionStats(c, db, sl)
	})

	app.Get("/api/users", middleware.PermissionRequired(store, db, sl, "users", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerUserListAPI(c, db, sl, store, config)
	})
	app.Get("/api/users/:id/permissions", middleware.PermissionRequired(store, db, sl, "users", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerGetUserPermissions(c, db, sl)
	})
	app.Post("/api/users/roles", middleware.PermissionRequired(store, db, sl, "users", "update"), func(c *fiber.Ctx) error {
		return handlers.HandlerAssignUserRole(c, db, sl, store, config)
	})

	app.Get("/api/resource-management/summary", middleware.PermissionRequired(store, db, sl, "resource_management", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerResourceManagementSummaryAPI(c, db, store)
	})
	app.Get("/api/resource-management/pillars", middleware.PermissionRequired(store, db, sl, "resource_management", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerResourceManagementPillarsAPI(c, db, store)
	})
	app.Get("/api/resource-management/rrt-teams", middleware.PermissionRequired(store, db, sl, "resource_management", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerResourceManagementRRTTeamsAPI(c, db, store)
	})
	app.Get("/api/resource-management/rrt-deployments", middleware.PermissionRequired(store, db, sl, "resource_management", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerResourceManagementRRTDeploymentsAPI(c, db, store)
	})
	app.Get("/api/resource-management/resources", middleware.PermissionRequired(store, db, sl, "resource_management", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerResourceManagementResourcesAPI(c, db, store)
	})
	app.Get("/api/resource-management/requisitions", middleware.PermissionRequired(store, db, sl, "resource_management", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerResourceManagementRequisitionsAPI(c, db, store)
	})
	app.Get("/api/resource-management/activity-logs", middleware.PermissionRequired(store, db, sl, "resource_management", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerResourceManagementActivityLogsAPI(c, db, store)
	})

	// Protected routes
	appGroup := app.Group("/")
	appGroup.Use(AuthRequired(store))
	{
		appGroup.Get("/home", func(c *fiber.Ctx) error { return handlers.HandlerHome(c, db, sl, store, config) })
		appGroup.Get("/alerts", middleware.PermissionRequired(store, db, sl, "alerts", "read"), func(c *fiber.Ctx) error {
			return handlers.HandlerAlerts(c, db, sl, store, config)
		})

		// Alerts API endpoint for paginated data
		appGroup.Get("/api/alerts", middleware.PermissionRequired(store, db, sl, "alerts", "read"), func(c *fiber.Ctx) error {
			return handlers.HandlerAlertsAPI(c, db, sl, store, config)
		})

		// 6767 Alerts from DHIS2
		appGroup.Get("/api/alerts/6767", middleware.PermissionRequired(store, db, sl, "alerts", "read"), func(c *fiber.Ctx) error {
			return handlers.HandlerAlerts6767API(c, db, sl, store, config)
		})

		// Alerts debug endpoint
		appGroup.Get("/api/alerts/debug", middleware.PermissionRequired(store, db, sl, "alerts", "read"), func(c *fiber.Ctx) error {
			return handlers.HandlerAlertsDebug(c, db, sl, store, config)
		})
		// VHF CIF routes with RBAC protection
		appGroup.Get("/vhf-cif", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "vhf_cif")
		})
		appGroup.Post("/vhf-cif", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
			return handlers.HandlerVHFPatientSubmit(c, db, sl, store, config, smsService)
		})
		appGroup.Get("/vhf-cif/clinical-signs/:id", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "vhf_clinical_signs")
		})
		appGroup.Post("/vhf-cif/clinical-signs/:id", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
			return handlers.HandlerVHFClinicalSignsSubmit(c, db, sl, store, config)
		})
		appGroup.Get("/vhf-cif/hospitalization/:id", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "vhf_hospitalization")
		})
		appGroup.Post("/vhf-cif/hospitalization/:id", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
			return handlers.HandlerVHFHospitalizationSubmit(c, db, sl, store, config)
		})
		appGroup.Get("/vhf-cif/risk-factors/:id", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "vhf_risk_factors")
		})
		appGroup.Post("/vhf-cif/risk-factors/:id", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
			return handlers.HandlerVHFRiskFactorsSubmit(c, db, sl, store, config)
		})
		appGroup.Get("/vhf-cif/laboratory/:id", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "vhf_laboratory")
		})
		appGroup.Post("/vhf-cif/laboratory/:id", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
			return handlers.HandlerVHFLaboratorySubmit(c, db, sl, store, config, smsService)
		})
		appGroup.Get("/vhf-cif/investigator/:id", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "vhf_investigator")
		})
		appGroup.Post("/vhf-cif/investigator/:id", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
			return handlers.HandlerVHFInvestigatorSubmit(c, db, sl, store, config)
		})
		appGroup.Get("/vhf-cif/success", middleware.PermissionRequired(store, db, sl, "vhf_patients", "read"), func(c *fiber.Ctx) error {
			return handlers.HandlerVHFSuccess(c, db, sl, store, config)
		})
		appGroup.Get("/vhf-list", middleware.PermissionRequired(store, db, sl, "vhf_cases", "read"), func(c *fiber.Ctx) error {
			return handlers.HandlerVHFList(c, db, sl, store, config)
		})
		appGroup.Get("/vhf-cif/view/:id", middleware.PermissionRequired(store, db, sl, "vhf_patients", "read"), func(c *fiber.Ctx) error {
			return handlers.HandlerVHFView(c, db, sl, store, config)
		})
		appGroup.Get("/vhf-lab/:id", middleware.PermissionRequired(store, db, sl, "vhf_lab", "read"), func(c *fiber.Ctx) error {
			return handlers.HandlerVHFLabForm(c, db, sl, store, config)
		})
		appGroup.Post("/vhf-lab/:id", middleware.PermissionRequired(store, db, sl, "vhf_lab", "create"), func(c *fiber.Ctx) error {
			return handlers.HandlerVHFLabSave(c, db, sl, store, config)
		})

		// Inventory routes
		inventoryHandler := handlers.NewInventoryHandler(db, store)

		// Inventory dashboard
		appGroup.Get("/inventory", func(c *fiber.Ctx) error {
			log.Printf("DEBUG: /inventory route accessed")
			return inventoryHandler.HandlerInventoryDashboard(c)
		})

		// Inventory items
		appGroup.Get("/inventory/items", func(c *fiber.Ctx) error {
			return inventoryHandler.HandlerInventoryItemsList(c)
		})
		appGroup.Get("/inventory/items/new", func(c *fiber.Ctx) error {
			return inventoryHandler.HandlerInventoryItemForm(c)
		})
		appGroup.Get("/inventory/items/edit/:id", func(c *fiber.Ctx) error {
			return inventoryHandler.HandlerInventoryItemForm(c)
		})
		appGroup.Post("/inventory/items/save", func(c *fiber.Ctx) error {
			return inventoryHandler.HandlerInventoryItemSave(c)
		})

		// Inventory stock management
		appGroup.Get("/inventory/stock", func(c *fiber.Ctx) error {
			return inventoryHandler.HandlerInventoryStockForm(c)
		})
		appGroup.Post("/inventory/stock/save", func(c *fiber.Ctx) error {
			return inventoryHandler.HandlerInventoryStockSave(c)
		})

		// Purchase orders
		appGroup.Get("/inventory/purchase-orders", func(c *fiber.Ctx) error {
			return inventoryHandler.HandlerInventoryPurchaseOrderForm(c)
		})
		appGroup.Post("/inventory/purchase-orders/save", func(c *fiber.Ctx) error {
			return inventoryHandler.HandlerInventoryPurchaseOrderSave(c)
		})

		// Requisitions
		appGroup.Get("/inventory/requisitions", func(c *fiber.Ctx) error {
			return inventoryHandler.HandlerInventoryRequisitionForm(c)
		})
		appGroup.Post("/inventory/requisitions/save", func(c *fiber.Ctx) error {
			return inventoryHandler.HandlerInventoryRequisitionSave(c)
		})

		// Reports
		appGroup.Get("/inventory/reports", func(c *fiber.Ctx) error {
			return inventoryHandler.HandlerInventoryReports(c)
		})

		// Donation routes
		appGroup.Get("/inventory/donations", func(c *fiber.Ctx) error {
			return inventoryHandler.HandlerDonationsList(c)
		})
		appGroup.Get("/inventory/donations/new", func(c *fiber.Ctx) error {
			return inventoryHandler.HandlerDonationForm(c)
		})
		appGroup.Post("/inventory/donations/save", func(c *fiber.Ctx) error {
			return inventoryHandler.HandlerDonationSave(c)
		})
		appGroup.Get("/inventory/donations/:id", func(c *fiber.Ctx) error {
			return inventoryHandler.HandlerDonationView(c)
		})
		appGroup.Get("/inventory/donors", func(c *fiber.Ctx) error {
			return inventoryHandler.HandlerDonorsList(c)
		})
		appGroup.Get("/inventory/donors/new", func(c *fiber.Ctx) error {
			return inventoryHandler.HandlerDonorForm(c)
		})
		appGroup.Post("/inventory/donors/save", func(c *fiber.Ctx) error {
			return inventoryHandler.HandlerDonorSave(c)
		})

		// Inventory API routes
		appGroup.Get("/api/inventory/items", func(c *fiber.Ctx) error {
			return inventoryHandler.HandlerInventoryAPIItems(c)
		})
		appGroup.Get("/api/inventory/stock-levels", func(c *fiber.Ctx) error {
			return inventoryHandler.HandlerInventoryAPIStockLevels(c)
		})
		appGroup.Get("/api/inventory/low-stock", func(c *fiber.Ctx) error {
			return inventoryHandler.HandlerInventoryAPILowStock(c)
		})
		appGroup.Get("/api/inventory/categories", func(c *fiber.Ctx) error {
			return inventoryHandler.HandlerInventoryAPICategories(c)
		})
		appGroup.Get("/api/inventory/suppliers", func(c *fiber.Ctx) error {
			return inventoryHandler.HandlerInventoryAPISuppliers(c)
		})
		appGroup.Get("/api/inventory/treatment-sites", func(c *fiber.Ctx) error {
			return inventoryHandler.HandlerInventoryAPITreatmentSites(c)
		})
		appGroup.Get("/api/inventory/transaction-types", func(c *fiber.Ctx) error {
			return inventoryHandler.HandlerInventoryAPITransactionTypes(c)
		})

		// Resource Management routes
		resourceHandler := handlers.NewResourceManagementHandler(db, store)

		// Resource Management dashboard
		appGroup.Get("/resource-management", middleware.PermissionRequired(store, db, sl, "resource_management", "read"), func(c *fiber.Ctx) error {
			return resourceHandler.HandlerResourceDashboard(c)
		})

		// Pillars Management
		pillarsHandler := handlers.NewPillarsHandler(db, store)
		appGroup.Get("/resource-management/pillars", middleware.PermissionRequired(store, db, sl, "resource_management", "read"), func(c *fiber.Ctx) error {
			return pillarsHandler.HandlerPillarsList(c)
		})
		appGroup.Get("/resource-management/pillars/new", middleware.PermissionRequired(store, db, sl, "resource_management", "create"), func(c *fiber.Ctx) error {
			return pillarsHandler.HandlerPillarForm(c)
		})
		appGroup.Get("/resource-management/pillars/edit/:id", middleware.PermissionRequired(store, db, sl, "resource_management", "update"), func(c *fiber.Ctx) error {
			return pillarsHandler.HandlerPillarForm(c)
		})
		appGroup.Post("/resource-management/pillars/save", middleware.PermissionRequired(store, db, sl, "resource_management", "create"), func(c *fiber.Ctx) error {
			return pillarsHandler.HandlerPillarSave(c)
		})
		appGroup.Get("/resource-management/pillars/changes/:id", middleware.PermissionRequired(store, db, sl, "resource_management", "read"), func(c *fiber.Ctx) error {
			return pillarsHandler.HandlerPillarChanges(c)
		})
		appGroup.Get("/api/pillars", middleware.PermissionRequired(store, db, sl, "resource_management", "read"), func(c *fiber.Ctx) error {
			return pillarsHandler.HandlerPillarsAPI(c)
		})

		// RRT Teams
		appGroup.Get("/resource-management/rrt-teams", middleware.PermissionRequired(store, db, sl, "resource_management", "read"), func(c *fiber.Ctx) error {
			return resourceHandler.HandlerRRTTeamsList(c)
		})
		appGroup.Get("/resource-management/rrt-teams/new", middleware.PermissionRequired(store, db, sl, "resource_management", "create"), func(c *fiber.Ctx) error {
			return resourceHandler.HandlerRRTTeamForm(c)
		})
		appGroup.Get("/resource-management/rrt-teams/edit/:id", middleware.PermissionRequired(store, db, sl, "resource_management", "update"), func(c *fiber.Ctx) error {
			return resourceHandler.HandlerRRTTeamForm(c)
		})
		appGroup.Post("/resource-management/rrt-teams/save", middleware.PermissionRequired(store, db, sl, "resource_management", "create"), func(c *fiber.Ctx) error {
			return resourceHandler.HandlerRRTTeamSave(c)
		})

		// RRT Deployments
		appGroup.Get("/resource-management/rrt-deployments", middleware.PermissionRequired(store, db, sl, "resource_management", "read"), func(c *fiber.Ctx) error {
			return resourceHandler.HandlerRRTDeploymentsList(c)
		})
		appGroup.Get("/resource-management/rrt-deployments/new", middleware.PermissionRequired(store, db, sl, "resource_management", "create"), func(c *fiber.Ctx) error {
			return resourceHandler.HandlerRRTDeploymentForm(c)
		})
		appGroup.Get("/resource-management/rrt-deployments/edit/:id", middleware.PermissionRequired(store, db, sl, "resource_management", "update"), func(c *fiber.Ctx) error {
			return resourceHandler.HandlerRRTDeploymentForm(c)
		})
		appGroup.Post("/resource-management/rrt-deployments/save", middleware.PermissionRequired(store, db, sl, "resource_management", "create"), func(c *fiber.Ctx) error {
			return resourceHandler.HandlerRRTDeploymentSave(c)
		})

		// Resources
		appGroup.Get("/resource-management/resources", middleware.PermissionRequired(store, db, sl, "resource_management", "read"), func(c *fiber.Ctx) error {
			return resourceHandler.HandlerResourcesList(c)
		})
		appGroup.Get("/resource-management/resources/new", middleware.PermissionRequired(store, db, sl, "resource_management", "create"), func(c *fiber.Ctx) error {
			return resourceHandler.HandlerResourceForm(c)
		})
		appGroup.Get("/resource-management/resources/edit/:id", middleware.PermissionRequired(store, db, sl, "resource_management", "update"), func(c *fiber.Ctx) error {
			return resourceHandler.HandlerResourceForm(c)
		})
		appGroup.Post("/resource-management/resources/save", middleware.PermissionRequired(store, db, sl, "resource_management", "create"), func(c *fiber.Ctx) error {
			return resourceHandler.HandlerResourceSave(c)
		})

		// Requisitions
		appGroup.Get("/resource-management/requisitions", middleware.PermissionRequired(store, db, sl, "resource_management", "read"), func(c *fiber.Ctx) error {
			return resourceHandler.HandlerRequisitionsList(c)
		})
		appGroup.Get("/resource-management/requisitions/new", middleware.PermissionRequired(store, db, sl, "resource_management", "create"), func(c *fiber.Ctx) error {
			return resourceHandler.HandlerRequisitionForm(c)
		})
		appGroup.Get("/resource-management/requisitions/edit/:id", middleware.PermissionRequired(store, db, sl, "resource_management", "update"), func(c *fiber.Ctx) error {
			return resourceHandler.HandlerRequisitionForm(c)
		})
		appGroup.Post("/resource-management/requisitions/save", middleware.PermissionRequired(store, db, sl, "resource_management", "create"), func(c *fiber.Ctx) error {
			return resourceHandler.HandlerRequisitionSave(c)
		})

		// Activity Logs
		appGroup.Get("/resource-management/activity-logs", middleware.PermissionRequired(store, db, sl, "resource_management", "read"), func(c *fiber.Ctx) error {
			return resourceHandler.HandlerActivityLogsList(c)
		})
		appGroup.Get("/resource-management/activity-logs/new", middleware.PermissionRequired(store, db, sl, "resource_management", "create"), func(c *fiber.Ctx) error {
			return resourceHandler.HandlerActivityLogForm(c)
		})
		appGroup.Get("/resource-management/activity-logs/edit/:id", middleware.PermissionRequired(store, db, sl, "resource_management", "update"), func(c *fiber.Ctx) error {
			return resourceHandler.HandlerActivityLogForm(c)
		})
		appGroup.Post("/resource-management/activity-logs/save", middleware.PermissionRequired(store, db, sl, "resource_management", "create"), func(c *fiber.Ctx) error {
			return resourceHandler.HandlerActivityLogSave(c)
		})

		// SitRep Generation
		appGroup.Get("/resource-management/sitrep", middleware.PermissionRequired(store, db, sl, "resource_management", "read"), func(c *fiber.Ctx) error {
			return resourceHandler.HandlerSitRepGeneration(c)
		})
		appGroup.Post("/resource-management/sitrep/generate", middleware.PermissionRequired(store, db, sl, "resource_management", "create"), func(c *fiber.Ctx) error {
			return resourceHandler.HandlerGenerateSitRep(c)
		})

		// RRT Team Members routes
		rrtTeamMemberHandler := handlers.NewRRTTeamMembersHandler(db, store)

		// RRT Team Members
		appGroup.Get("/resource-management/rrt-team-members", middleware.PermissionRequired(store, db, sl, "resource_management", "read"), func(c *fiber.Ctx) error {
			return rrtTeamMemberHandler.HandlerRRTTeamMembersList(c)
		})
		appGroup.Get("/resource-management/rrt-team-members/new", middleware.PermissionRequired(store, db, sl, "resource_management", "create"), func(c *fiber.Ctx) error {
			return rrtTeamMemberHandler.HandlerRRTTeamMemberForm(c)
		})
		appGroup.Get("/resource-management/rrt-team-members/edit/:id", middleware.PermissionRequired(store, db, sl, "resource_management", "update"), func(c *fiber.Ctx) error {
			return rrtTeamMemberHandler.HandlerRRTTeamMemberForm(c)
		})
		appGroup.Post("/resource-management/rrt-team-members/save", middleware.PermissionRequired(store, db, sl, "resource_management", "create"), func(c *fiber.Ctx) error {
			return rrtTeamMemberHandler.HandlerRRTTeamMemberSave(c)
		})

		// RRT Team Member Assignments
		appGroup.Get("/resource-management/rrt-team-member-assignments", middleware.PermissionRequired(store, db, sl, "resource_management", "read"), func(c *fiber.Ctx) error {
			return rrtTeamMemberHandler.HandlerRRTTeamMemberAssignmentsList(c)
		})
		appGroup.Get("/resource-management/rrt-team-member-assignments/new", middleware.PermissionRequired(store, db, sl, "resource_management", "create"), func(c *fiber.Ctx) error {
			return rrtTeamMemberHandler.HandlerRRTTeamMemberAssignmentForm(c)
		})
		appGroup.Get("/resource-management/rrt-team-member-assignments/edit/:id", middleware.PermissionRequired(store, db, sl, "resource_management", "update"), func(c *fiber.Ctx) error {
			return rrtTeamMemberHandler.HandlerRRTTeamMemberAssignmentForm(c)
		})
		appGroup.Post("/resource-management/rrt-team-member-assignments/save", middleware.PermissionRequired(store, db, sl, "resource_management", "create"), func(c *fiber.Ctx) error {
			return rrtTeamMemberHandler.HandlerRRTTeamMemberAssignmentSave(c)
		})

		// RRT Deployment Proposals
		appGroup.Get("/resource-management/deployment-proposals", middleware.PermissionRequired(store, db, sl, "resource_management", "read"), func(c *fiber.Ctx) error {
			return rrtTeamMemberHandler.HandlerRRTDeploymentProposalsList(c)
		})
		appGroup.Get("/resource-management/deployment-proposals/new", middleware.PermissionRequired(store, db, sl, "resource_management", "create"), func(c *fiber.Ctx) error {
			return rrtTeamMemberHandler.HandlerRRTDeploymentProposalForm(c)
		})
		appGroup.Get("/resource-management/deployment-proposals/view/:id", middleware.PermissionRequired(store, db, sl, "resource_management", "read"), func(c *fiber.Ctx) error {
			return rrtTeamMemberHandler.HandlerRRTDeploymentProposalView(c)
		})
		appGroup.Get("/resource-management/deployment-proposals/edit/:id", middleware.PermissionRequired(store, db, sl, "resource_management", "update"), func(c *fiber.Ctx) error {
			return rrtTeamMemberHandler.HandlerRRTDeploymentProposalForm(c)
		})
		appGroup.Post("/resource-management/deployment-proposals/save", middleware.PermissionRequired(store, db, sl, "resource_management", "create"), func(c *fiber.Ctx) error {
			return rrtTeamMemberHandler.HandlerRRTDeploymentProposalSave(c)
		})
		appGroup.Post("/resource-management/deployment-proposals/approve/:id", middleware.PermissionRequired(store, db, sl, "resource_management", "update"), func(c *fiber.Ctx) error {
			return rrtTeamMemberHandler.HandlerRRTDeploymentProposalApprove(c)
		})
		appGroup.Post("/resource-management/deployment-proposals/reject/:id", middleware.PermissionRequired(store, db, sl, "resource_management", "update"), func(c *fiber.Ctx) error {
			return rrtTeamMemberHandler.HandlerRRTDeploymentProposalReject(c)
		})

		// RRT Deployment Extensions
		appGroup.Get("/resource-management/deployment-extensions", middleware.PermissionRequired(store, db, sl, "resource_management", "read"), func(c *fiber.Ctx) error {
			return rrtTeamMemberHandler.HandlerRRTDeploymentExtensionsList(c)
		})
		appGroup.Get("/resource-management/deployment-extensions/new", middleware.PermissionRequired(store, db, sl, "resource_management", "create"), func(c *fiber.Ctx) error {
			return rrtTeamMemberHandler.HandlerRRTDeploymentExtensionForm(c)
		})
		appGroup.Get("/resource-management/deployment-extensions/edit/:id", middleware.PermissionRequired(store, db, sl, "resource_management", "update"), func(c *fiber.Ctx) error {
			return rrtTeamMemberHandler.HandlerRRTDeploymentExtensionForm(c)
		})
		appGroup.Post("/resource-management/deployment-extensions/save", middleware.PermissionRequired(store, db, sl, "resource_management", "create"), func(c *fiber.Ctx) error {
			return rrtTeamMemberHandler.HandlerRRTDeploymentExtensionSave(c)
		})
		appGroup.Post("/resource-management/deployment-extensions/approve/:id", middleware.PermissionRequired(store, db, sl, "resource_management", "update"), func(c *fiber.Ctx) error {
			return rrtTeamMemberHandler.HandlerRRTDeploymentExtensionApprove(c)
		})
		appGroup.Post("/resource-management/deployment-extensions/reject/:id", middleware.PermissionRequired(store, db, sl, "resource_management", "update"), func(c *fiber.Ctx) error {
			return rrtTeamMemberHandler.HandlerRRTDeploymentExtensionReject(c)
		})

		// RRT Field Role Assignments
		appGroup.Get("/resource-management/field-role-assignments", middleware.PermissionRequired(store, db, sl, "resource_management", "read"), func(c *fiber.Ctx) error {
			return rrtTeamMemberHandler.HandlerRRTFieldRoleAssignmentsList(c)
		})
		appGroup.Get("/resource-management/field-role-assignments/new", middleware.PermissionRequired(store, db, sl, "resource_management", "create"), func(c *fiber.Ctx) error {
			return rrtTeamMemberHandler.HandlerRRTFieldRoleAssignmentForm(c)
		})
		appGroup.Get("/resource-management/field-role-assignments/edit/:id", middleware.PermissionRequired(store, db, sl, "resource_management", "update"), func(c *fiber.Ctx) error {
			return rrtTeamMemberHandler.HandlerRRTFieldRoleAssignmentForm(c)
		})
		appGroup.Post("/resource-management/field-role-assignments/save", middleware.PermissionRequired(store, db, sl, "resource_management", "create"), func(c *fiber.Ctx) error {
			return rrtTeamMemberHandler.HandlerRRTFieldRoleAssignmentSave(c)
		})

		// Add more protected routes...

		api := app.Group("/api") // Group for all API routes

		enk := api.Group("/encounter")
		sym := api.Group("/sym")
		mob := api.Group("/mob")
		rus := api.Group("/rush")
		lab := api.Group("/lab")
		sta := api.Group("/status")

		emp := appGroup.Group("/employees") // Employees
		usr := appGroup.Group("/users")     // users
		hfs := app.Group("/secure")         // Health facilities
		cse := app.Group("/cases")

		//enc := app.Group("/encounter")
		dis := app.Group("/discharge")
		rpt := app.Group("/reports")
		srv := app.Group("/surveillance")

		// Additional routes
		RouteFacilities(hfs, db, sl, store, config)
		RouteUsers(usr, db, sl, store, config)
		RouteCases(cse, db, sl, store, config, smsService, voiceService)
		RouteMorbidity(mob, db, sl, store, config)
		RouteSymptoms(sym, db, sl, store, config)
		RouteRush(rus, db, sl, store, config)

		RouteEmployees(emp, db, sl, store, config)
		RouteDischarge(dis, db, sl, store, config)

		RouteReports(rpt, db, sl, store, config)
		RouteSurveillance(srv, db, sl, store, config)

		RouteAPIEncounter(enk, db, sl, store, config)
		RouteAPIStatus(sta, db, sl, store, config)

		// Add missing routes for home.html navigation
		appGroup.Get("/vhf", func(c *fiber.Ctx) error { return handlers.HandlerVHFList(c, db, sl, store, config) })
		appGroup.Get("/vhf/new", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "vhf_cif")
		})
		// Register /cases/list and /cases BEFORE /cases/:outbreak_id so they match first (param route would otherwise catch "list")
		appGroup.Get("/cases/list", middleware.PermissionRequired(store, db, sl, "vhf_patients", "read"), func(c *fiber.Ctx) error {
			return handlers.HandlerCasesList(c, db, sl, store, config)
		})
		appGroup.Get("/cases", middleware.PermissionRequired(store, db, sl, "vhf_patients", "read"), func(c *fiber.Ctx) error {
			return handlers.HandlerCasesList(c, db, sl, store, config)
		})
		appGroup.Get("/cases/new", func(c *fiber.Ctx) error { return handlers.HandlerCasesForm(c, db, sl, store, config, smsService, voiceService) })
		appGroup.Get("/cases/:outbreak_id", func(c *fiber.Ctx) error {
			// Set the outbreak ID in session and redirect to cases list
			outbreakID, err := strconv.Atoi(c.Params("outbreak_id"))
			if err != nil {
				return c.Status(400).SendString("Invalid outbreak ID")
			}

			// Set outbreak in session
			sess, err := store.Get(c)
			if err != nil {
				return c.Status(500).SendString("Failed to get session")
			}
			sess.Set("outbreak_id", outbreakID)
			sess.Set("selected_outbreak", outbreakID) // Set both keys for consistency
			if err := sess.Save(); err != nil {
				return c.Status(500).SendString("Failed to save session")
			}

			// Else Redirect to cases list
			return c.Redirect("/cases/list")
		})
		appGroup.Get("/lab", func(c *fiber.Ctx) error { return handlers.HandlerLabList(c, db, sl, store, config) })
		appGroup.Get("/change-password", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "change_password")
		})
		appGroup.Get("/outbreaks/assignments", func(c *fiber.Ctx) error {
			// Create outbreak assignment handler directly
			userService := models.NewUserService(db)
			userOutbreakService := models.NewUserOutbreakService(db)
			patientRoleService := models.NewPatientManagementRoleService(db)
			outbreakService := models.NewOutbreakService(db)
			facilityService := models.NewFacilityService(db)

			handler := handlers.NewOutbreakAssignmentHandler(
				userOutbreakService, patientRoleService, userService, outbreakService, facilityService, store,
			)
			return handler.ShowOutbreakAssignments(c)
		})
		appGroup.Get("/outbreaks/assign", func(c *fiber.Ctx) error {
			// Create outbreak assignment handler directly
			userService := models.NewUserService(db)
			userOutbreakService := models.NewUserOutbreakService(db)
			patientRoleService := models.NewPatientManagementRoleService(db)
			outbreakService := models.NewOutbreakService(db)
			facilityService := models.NewFacilityService(db)

			handler := handlers.NewOutbreakAssignmentHandler(
				userOutbreakService, patientRoleService, userService, outbreakService, facilityService, store,
			)
			return handler.ShowAssignFormFiber(c)
		})
		appGroup.Post("/outbreaks/assign", func(c *fiber.Ctx) error {
			// Create outbreak assignment handler directly
			userService := models.NewUserService(db)
			userOutbreakService := models.NewUserOutbreakService(db)
			patientRoleService := models.NewPatientManagementRoleService(db)
			outbreakService := models.NewOutbreakService(db)
			facilityService := models.NewFacilityService(db)

			handler := handlers.NewOutbreakAssignmentHandler(
				userOutbreakService, patientRoleService, userService, outbreakService, facilityService, store,
			)
			return handler.HandleAssignFormSubmission(c)
		})
		appGroup.Delete("/api/outbreaks/:outbreak_id/users/:user_id", func(c *fiber.Ctx) error {
			// Create outbreak assignment handler directly
			userService := models.NewUserService(db)
			userOutbreakService := models.NewUserOutbreakService(db)
			patientRoleService := models.NewPatientManagementRoleService(db)
			outbreakService := models.NewOutbreakService(db)
			facilityService := models.NewFacilityService(db)

			handler := handlers.NewOutbreakAssignmentHandler(
				userOutbreakService, patientRoleService, userService, outbreakService, facilityService, store,
			)
			return handler.RemoveUserFromOutbreak(c)
		})
		appGroup.Get("/outbreaks/:id", func(c *fiber.Ctx) error {
			return handlers.HandlerOutbreakForm(c, db, sl, store, config)
		})
		appGroup.Get("/patient-roles", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "patient_roles")
		})
		appGroup.Get("/roles", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "list_roles")
		})
		appGroup.Get("/roles/new", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "form_role")
		})
		appGroup.Get("/rbac-dashboard", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateDataWithDB(c, store, db), "rbac_dashboard")
		})

		// Development endpoint to create default admin user
		appGroup.Get("/setup-admin", func(c *fiber.Ctx) error {
			handler := handlers.NewRBACManagementHandler(db, sl, store, config)
			return handler.CreateDefaultAdminUser(c)
		})

		// Temporary: Make RBAC API endpoints public for development
		// appGroup.Get("/api/roles", func(c *fiber.Ctx) error { return handlers.HandlerGetRoles(c, db, sl) })
		// appGroup.Get("/api/permissions", func(c *fiber.Ctx) error { return handlers.HandlerGetPermissions(c, db, sl) })
		// appGroup.Get("/api/rbac/migration-status", func(c *fiber.Ctx) error { return handlers.HandlerGetMigrationStatus(c, db, sl) })
		// appGroup.Get("/api/users", func(c *fiber.Ctx) error { return handlers.HandlerGetUsers(c, db, sl) })
		// appGroup.Post("/api/users/roles", func(c *fiber.Ctx) error { return handlers.HandlerAssignUserRole(c, db, sl, store, config) })
		appGroup.Get("/permissions", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "permissions")
		})
		appGroup.Get("/employees/statistics", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "employee_statistics")
		})
		appGroup.Get("/employees/export", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "employee_export")
		})

		// User routes are handled by RouteUsers function

		// RBAC API routes (protected)
		// appGroup.Get("/api/roles", func(c *fiber.Ctx) error { return handlers.HandlerGetRoles(c, db, sl) })
		appGroup.Post("/api/roles", func(c *fiber.Ctx) error { return handlers.HandlerCreateRole(c, db, sl, store, config) })
		appGroup.Put("/api/roles/:id", func(c *fiber.Ctx) error { return handlers.HandlerUpdateRole(c, db, sl, store, config) })
		appGroup.Delete("/api/roles/:id", func(c *fiber.Ctx) error { return handlers.HandlerDeleteRole(c, db, sl, store, config) })

		appGroup.Get("/api/permissions", func(c *fiber.Ctx) error { return handlers.HandlerGetPermissions(c, db, sl) })
		appGroup.Post("/api/permissions", func(c *fiber.Ctx) error { return handlers.HandlerCreatePermission(c, db, sl, store, config) })
		appGroup.Put("/api/permissions/:id", func(c *fiber.Ctx) error { return handlers.HandlerUpdatePermission(c, db, sl, store, config) })
		appGroup.Delete("/api/permissions/:id", func(c *fiber.Ctx) error { return handlers.HandlerDeletePermission(c, db, sl, store, config) })

		appGroup.Get("/api/user-roles/:user_id", func(c *fiber.Ctx) error { return handlers.HandlerGetUserRoles(c, db, sl) })
		appGroup.Post("/api/user-roles", func(c *fiber.Ctx) error { return handlers.HandlerAssignUserRole(c, db, sl, store, config) })
		appGroup.Delete("/api/user-roles/:user_id/:role_id", func(c *fiber.Ctx) error { return handlers.HandlerRemoveUserRole(c, db, sl, store, config) })

		appGroup.Get("/api/role-permissions/:role_id", func(c *fiber.Ctx) error { return handlers.HandlerGetRolePermissions(c, db, sl) })
		appGroup.Post("/api/role-permissions", func(c *fiber.Ctx) error { return handlers.HandlerAssignRolePermission(c, db, sl, store, config) })
		appGroup.Delete("/api/role-permissions/:role_id/:permission_id", func(c *fiber.Ctx) error { return handlers.HandlerRemoveRolePermission(c, db, sl, store, config) })

		appGroup.Get("/api/rbac/migration-status", func(c *fiber.Ctx) error { return handlers.HandlerGetMigrationStatus(c, db, sl) })
		appGroup.Post("/scripts/migrate_user_rights_to_rbac", func(c *fiber.Ctx) error { return handlers.HandlerMigrateUserRightsToRBAC(c, db, sl, store, config) })

		// Outbreak routes (protected)
		appGroup.Get("/outbreaks", func(c *fiber.Ctx) error { return handlers.HandlerOutbreakList(c, db, sl, store, config) })
		appGroup.Get("/outbreaks/new/:i", func(c *fiber.Ctx) error { return handlers.HandlerOutbreakForm(c, db, sl, store, config) })
		appGroup.Get("/outbreaks/edit/:i", func(c *fiber.Ctx) error { return handlers.HandlerOutbreakForm(c, db, sl, store, config) })
		appGroup.Post("/outbreaks/save", func(c *fiber.Ctx) error { return handlers.HandlerOutbreakSubmit(c, db, sl, store, config) })
		appGroup.Get("/outbreaks/close/:i", func(c *fiber.Ctx) error { return handlers.HandlerOutbreakClose(c, db, sl, store, config) })
		appGroup.Post("/outbreaks/select/:i", func(c *fiber.Ctx) error {
			id, err := strconv.Atoi(c.Params("i"))
			if err != nil {
				return c.Status(400).SendString("Invalid outbreak ID")
			}
			if err := handlers.SetSelectedOutbreak(c, store, id); err != nil {
				return c.Status(500).SendString("Failed to select outbreak")
			}
			return c.SendStatus(200)
		})

		// New routes for mpox daily follow-up
		cse.Get("/encounters/mpox-admission/new/:i", func(c *fiber.Ctx) error { return handlers.HandlerMpoxAdmissionForm(c, db, sl, store, config) })
		cse.Post("/encounters/mpox-admission/save", func(c *fiber.Ctx) error { return handlers.HandlerMpoxAdmissionSubmit(c, db, sl, store, config) })

		// Mpox daily follow-up routes
		cse.Get("/encounters/mpox-daily-follow-up/new/:i", func(c *fiber.Ctx) error { return handlers.HandlerMpoxDailyFollowUpForm(c, db, sl, store, config) })
		cse.Post("/encounters/mpox-daily-follow-up/save", func(c *fiber.Ctx) error { return handlers.HandlerMpoxDailyFollowUpSubmit(c, db, sl, store, config) })

		// VHF API routes
		lab.Get("/api/vhf-cases", func(c *fiber.Ctx) error { return handlers.HandlerVHFList(c, db, sl, store, config) })
		lab.Get("/api/vhf-cases/:id", func(c *fiber.Ctx) error { return handlers.HandlerVHFView(c, db, sl, store, config) })

		// Protected API routes
		protected := api.Group("/vhf")
		protected.Get("/cases", func(c *fiber.Ctx) error { return handlers.HandlerVHFList(c, db, sl, store, config) })
		protected.Get("/cases/:id", func(c *fiber.Ctx) error { return handlers.HandlerVHFView(c, db, sl, store, config) })

		// RBAC Management routes
		rbacHandler := handlers.NewRBACManagementHandler(db, sl, store, config)

		// Role management
		protected.Get("/roles", rbacHandler.ListRoles)
		protected.Get("/roles/:id", rbacHandler.GetRole)
		protected.Post("/roles", rbacHandler.CreateRole)
		protected.Put("/roles/:id", rbacHandler.UpdateRole)
		protected.Delete("/roles/:id", rbacHandler.DeleteRole)

		// Permission management
		protected.Get("/permissions", rbacHandler.ListPermissions)
		protected.Get("/permissions/:id", rbacHandler.GetPermission)
		protected.Post("/permissions", rbacHandler.CreatePermission)
		protected.Put("/permissions/:id", rbacHandler.UpdatePermission)
		protected.Delete("/permissions/:id", rbacHandler.DeletePermission)

		// User-role assignment
		protected.Post("/users/roles", rbacHandler.AssignUserRole)
		protected.Delete("/users/roles", rbacHandler.RemoveUserRole)

		// ==================== ROLE-BASED ROUTE GROUPS ====================

		// Super Admin routes - Full system access
		superAdminGroup := appGroup.Group("/admin", middleware.RoleRequired(store, db, sl, handlers.RoleSuperAdmin))
		superAdminGroup.Get("/system", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "admin_system")
		})
		superAdminGroup.Get("/users", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "admin_users")
		})
		superAdminGroup.Get("/roles", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "admin_roles")
		})
		superAdminGroup.Get("/permissions", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "admin_permissions")
		})

		// Admin routes - Management access
		adminGroup := appGroup.Group("/management", middleware.RoleRequiredAny(store, db, sl, handlers.RoleAdmin, handlers.RoleSuperAdmin))
		adminGroup.Get("/dashboard", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "admin_dashboard")
		})
		adminGroup.Get("/outbreaks", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "admin_outbreaks")
		})
		adminGroup.Get("/facilities", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "admin_facilities")
		})

		// Inventory Manager routes
		inventoryManagerGroup := appGroup.Group("/inventory", middleware.RoleRequiredAny(store, db, sl, handlers.RoleInventoryManager, handlers.RoleAdmin, handlers.RoleSuperAdmin))
		inventoryManagerGroup.Get("/dashboard", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "inventory_dashboard")
		})
		inventoryManagerGroup.Get("/categories", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "inventory_categories")
		})
		inventoryManagerGroup.Get("/items", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "inventory_items")
		})
		inventoryManagerGroup.Get("/suppliers", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "inventory_suppliers")
		})
		inventoryManagerGroup.Get("/transactions", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "inventory_transactions")
		})
		inventoryManagerGroup.Get("/reports", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "inventory_reports")
		})

		// Inventory Clerk routes
		inventoryClerkGroup := appGroup.Group("/inventory/clerk", middleware.RoleRequiredAny(store, db, sl, handlers.RoleInventoryClerk, handlers.RoleInventoryManager, handlers.RoleAdmin, handlers.RoleSuperAdmin))
		inventoryClerkGroup.Get("/stock-entry", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "inventory_stock_entry")
		})
		inventoryClerkGroup.Get("/stock-levels", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "inventory_stock_levels")
		})

		// Outbreak Coordinator routes
		outbreakCoordinatorGroup := appGroup.Group("/outbreak", middleware.RoleRequiredAny(store, db, sl, handlers.RoleOutbreakCoordinator, handlers.RoleAdmin, handlers.RoleSuperAdmin))
		outbreakCoordinatorGroup.Get("/dashboard", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "outbreak_dashboard")
		})
		outbreakCoordinatorGroup.Get("/cases", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "outbreak_cases")
		})
		outbreakCoordinatorGroup.Get("/admissions", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "outbreak_admissions")
		})

		// Case Manager routes
		caseManagerGroup := appGroup.Group("/cases", middleware.RoleRequiredAny(store, db, sl, handlers.RoleCaseManager, handlers.RoleOutbreakCoordinator, handlers.RoleAdmin, handlers.RoleSuperAdmin))
		caseManagerGroup.Get("/dashboard", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "case_dashboard")
		})
		caseManagerGroup.Get("/tracking", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "case_tracking")
		})
		caseManagerGroup.Get("/reports", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "case_reports")
		})

		// Reports routes
		reportsGroup := appGroup.Group("/reports", middleware.RoleRequiredAny(store, db, sl, handlers.RoleReports, handlers.RoleAdmin, handlers.RoleSuperAdmin))
		reportsGroup.Get("/dashboard", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "reports_dashboard")
		})
		reportsGroup.Get("/analytics", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "reports_analytics")
		})
		reportsGroup.Get("/exports", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "reports_exports")
		})

		// Lab Technician routes
		labTechnicianGroup := appGroup.Group("/lab", middleware.RoleRequiredAny(store, db, sl, handlers.RoleLabTechnician, handlers.RoleAdmin, handlers.RoleSuperAdmin))
		labTechnicianGroup.Get("/dashboard", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "lab_dashboard")
		})
		labTechnicianGroup.Get("/tests", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "lab_tests")
		})
		labTechnicianGroup.Get("/results", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "lab_results")
		})

		// Surveillance Officer routes
		surveillanceGroup := appGroup.Group("/surveillance", middleware.RoleRequiredAny(store, db, sl, handlers.RoleSurveillanceOfficer, handlers.RoleAdmin, handlers.RoleSuperAdmin))
		surveillanceGroup.Get("/dashboard", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "surveillance_dashboard")
		})
		surveillanceGroup.Get("/monitoring", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "surveillance_monitoring")
		})
		surveillanceGroup.Get("/alerts", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "surveillance_alerts")
		})

		// Data Entry routes
		dataEntryGroup := appGroup.Group("/data", middleware.RoleRequiredAny(store, db, sl, handlers.RoleDataEntry, handlers.RoleAdmin, handlers.RoleSuperAdmin))
		dataEntryGroup.Get("/dashboard", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "data_dashboard")
		})
		dataEntryGroup.Get("/entry", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "data_entry")
		})
		dataEntryGroup.Get("/forms", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "data_forms")
		})

		// Viewer routes (read-only access)
		viewerGroup := appGroup.Group("/view", middleware.RoleRequiredAny(store, db, sl, handlers.RoleViewer, handlers.RoleAdmin, handlers.RoleSuperAdmin))
		viewerGroup.Get("/dashboard", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "viewer_dashboard")
		})
		viewerGroup.Get("/reports", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "viewer_reports")
		})
		viewerGroup.Get("/data", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "viewer_data")
		})
		protected.Get("/users/:user_id/roles", rbacHandler.GetUserRoles)

		// Migration status
		protected.Get("/rbac/migration-status", rbacHandler.GetMigrationStatus)

	}

	// MPOX CIF routes (public)
	app.Get("/mpox-cif", func(c *fiber.Ctx) error {
		caseID := c.Query("case_id")
		data := handlers.NewTemplateData(c, store)
		data.Form = map[string]interface{}{"case_id": caseID}
		return handlers.GenerateHTML(c, db, data, "mpox_cif")
	})
	app.Post("/mpox-cif/save", func(c *fiber.Ctx) error {
		return handlers.HandlerMpoxCIFSubmit(c, db, sl)
	})
	app.Get("/mpox-cif/success", func(c *fiber.Ctx) error {
		return handlers.HandlerMpoxCIFSuccess(c, db, sl, store)
	})
	app.Get("/mpox-cif/view/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerMpoxCIFView(c, db, sl, store, config)
	})
	app.Get("/mpox-cif/edit/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerMpoxCIFEdit(c, db, sl, store, config)
	})

	// API endpoints for dropdown data (protected)
	protectedAPI := app.Group("/api")
	protectedAPI.Use(AuthRequired(store))
	{
		protectedAPI.Get("/outbreaks", func(c *fiber.Ctx) error {
			// Add debugging
			fmt.Printf("API /outbreaks called - checking authentication...\n")

			// Check session
			sess, err := store.Get(c)
			if err != nil {
				fmt.Printf("Session error: %v\n", err)
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Session error"})
			}

			userID := sess.Get("user")
			if userID == nil {
				fmt.Printf("No user ID in session\n")
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "No user ID in session"})
			}

			fmt.Printf("User authenticated: %v\n", userID)

			outbreakService := models.NewOutbreakService(db)
			outbreaks, err := outbreakService.GetAllOutbreaks()
			if err != nil {
				fmt.Printf("Database error: %v\n", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
			}

			fmt.Printf("Found %d outbreaks in database\n", len(outbreaks))
			return c.JSON(outbreaks)
		})

		protectedAPI.Get("/users", func(c *fiber.Ctx) error {
			// Add debugging
			fmt.Printf("API /users called - checking authentication...\n")

			// Check session
			sess, err := store.Get(c)
			if err != nil {
				fmt.Printf("Session error: %v\n", err)
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Session error"})
			}

			userID := sess.Get("user")
			if userID == nil {
				fmt.Printf("No user ID in session\n")
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "No user ID in session"})
			}

			fmt.Printf("User authenticated: %v\n", userID)

			userService := models.NewUserService(db)
			users, err := userService.GetAllUsers()
			if err != nil {
				fmt.Printf("Database error: %v\n", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
			}

			fmt.Printf("Found %d users in database\n", len(users))
			return c.JSON(users)
		})
	}

	// Add missing routes to prevent broken links
	app.Get("/about", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "about")
	})
	app.Get("/contact", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "contact")
	})
	app.Get("/patienthelp", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "patienthelp")
	})

	// Reports AJAX routes (public for authenticated users)
	app.Get("/reports/quick-stats", func(c *fiber.Ctx) error {
		return reports.GetQuickStats(c, db, sl, store, config)
	})
	app.Get("/reports/chart-data/:type", func(c *fiber.Ctx) error {
		return reports.GetChartData(c, db, sl, store, config)
	})
	app.Get("/reports/table-data", func(c *fiber.Ctx) error {
		return reports.GetTableData(c, db, sl, store, config)
	})
	app.Post("/reports/export", func(c *fiber.Ctx) error {
		return reports.ExportReport(c, db, sl, store, config)
	})
}

// AuthRequired middleware checks if user is authenticated and has required role
func AuthRequired(store *session.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip auth for public routes
		publicRoutes := []string{
			"/",
			"/login",
			"/logout",
			"/vhf-cif",
			"/vhf-cif/save",
			"/vhf-cif/success",
			"/mpox-cif",
			"/mpox-cif/save",
			"/mpox-cif/success",
			"/polio-cif",
			"/polio-cif/save",
			"/polio-cif/success",
			"/test-simple",
			"/inventory-test",
			"/api/locations/districts",
			"/api/locations/subcounties/:district_id",
			"/api/locations/parishes/:subcounty_id",
			"/api/locations/parishes/district/:district_id",
			"/api/locations/villages/:parish_id",
			"/api/locations/villages/district/:district_id",
			"/api/locations/villages/subcounty/:subcounty_id",
			"/api/lab/blood-types",
			"/api/lab/blood-types/category/:category",
			"/api/lab/swab-types",
			"/api/lab/urine-types",
			"/api/test-lab",
		}

		path := c.Path()
		log.Printf("DEBUG: AuthRequired checking path: %s", path)

		// Allow static assets (CSS, JS, etc.) so public pages like vhf-cif and mpox-cif can load Bootstrap
		if strings.HasPrefix(path, "/static") {
			log.Printf("DEBUG: Path %s matches public prefix /static", path)
			return c.Next()
		}
		// Allow API used by public CIF forms (e.g. facilities, locations)
		if strings.HasPrefix(path, "/api/facilities") || strings.HasPrefix(path, "/api/locations") {
			return c.Next()
		}

		for _, route := range publicRoutes {
			if matchesRoute(path, route) {
				log.Printf("DEBUG: Path %s matches public route %s", path, route)
				return c.Next()
			}
		}

		log.Printf("DEBUG: Path %s requires authentication", path)

		// Check if user is authenticated
		sess, err := store.Get(c)
		if err != nil {
			return c.Redirect("/login")
		}

		userID := sess.Get("user")
		if userID == nil {
			return c.Redirect("/login")
		}

		if sess.Get("isAuthenticated") != true {
			return c.Redirect("/login")
		}

		return c.Next()
	}
}

// matchesRoute checks if a path matches a route pattern with parameters
func matchesRoute(path, pattern string) bool {
	// Exact match
	if path == pattern {
		return true
	}

	// Split both path and pattern into segments
	pathSegments := strings.Split(path, "/")
	patternSegments := strings.Split(pattern, "/")

	// Different number of segments means no match
	if len(pathSegments) != len(patternSegments) {
		return false
	}

	// Check each segment
	for i, patternSeg := range patternSegments {
		pathSeg := pathSegments[i]

		// If pattern segment starts with ":", it's a parameter - always match
		if strings.HasPrefix(patternSeg, ":") {
			continue
		}

		// Otherwise, segments must match exactly
		if pathSeg != patternSeg {
			return false
		}
	}

	return true
}

func RouteAPIEncounter(v fiber.Router, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) {
	// Add RBAC permission checks for API encounter
	v.Get("/", middleware.PermissionRequired(store, db, sl, "vhf_patients", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerAPIGetEncounter(c, db, sl, store, config)
	})
}

func RouteAPIStatus(v fiber.Router, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) {
	// Add RBAC permission checks for API status
	v.Get("/list", middleware.PermissionRequired(store, db, sl, "vhf_patients", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerAPIGetStatuses(c, db, sl, store, config)
	})
	v.Post("/save", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
		return handlers.HandlerAPIPostStatus(c, db, sl, store, config)
	})
}

func RouteDischarge(v fiber.Router, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) {
	// Add RBAC permission checks for discharge management
	v.Get("/list", middleware.PermissionRequired(store, db, sl, "vhf_patients", "read"), func(c *fiber.Ctx) error {
		return handlers.GetDischarge(c, db, sl, store, config)
	})
	v.Get("/certificate", middleware.PermissionRequired(store, db, sl, "vhf_patients", "read"), func(c *fiber.Ctx) error {
		return handlers.Certificate(c, db, sl, store, config)
	})
	v.Post("/save", middleware.PermissionRequired(store, db, sl, "vhf_patients", "update"), func(c *fiber.Ctx) error {
		return handlers.Discharge(c, db, sl, store, config)
	})
}

func RouteHome(app *fiber.App, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config, smsService *services.SMSService) {
	// Landing page
	app.Get("/", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "landing")
	})

	// Login routes
	app.Get("/login", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "login")
	})
	app.Post("/login", func(c *fiber.Ctx) error {
		return handlers.HandlerLoginSubmit(c, db, sl, store, config)
	})
	app.Get("/logout", func(c *fiber.Ctx) error {
		return handlers.HandlerLoginOut(c, sl, store, config)
	})

	// VHF CIF routes
	app.Get("/vhf-cif", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "vhf_cif")
	})
	app.Post("/vhf-cif/save", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFPatientSubmit(c, db, sl, store, config, smsService)
	})
	app.Get("/vhf-cif/clinical-signs/:id", func(c *fiber.Ctx) error {
		data := handlers.NewTemplateData(c, store)
		data.Form = fiber.Map{"PatientID": c.Params("id")}
		return handlers.GenerateHTML(c, db, data, "vhf_clinical_signs")
	})
	app.Post("/vhf-cif/clinical-signs/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFClinicalSignsSubmit(c, db, sl, store, config)
	})
	app.Get("/vhf-cif/hospitalization/:id", func(c *fiber.Ctx) error {
		data := handlers.NewTemplateData(c, store)
		data.Form = fiber.Map{"PatientID": c.Params("id")}
		return handlers.GenerateHTML(c, db, data, "vhf_hospitalization")
	})
	app.Post("/vhf-cif/hospitalization/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFHospitalizationSubmit(c, db, sl, store, config)
	})
	app.Get("/vhf-cif/risk-factors/:id", func(c *fiber.Ctx) error {
		data := handlers.NewTemplateData(c, store)
		data.Form = fiber.Map{"PatientID": c.Params("id")}
		return handlers.GenerateHTML(c, db, data, "vhf_risk_factors")
	})
	app.Post("/vhf-cif/risk-factors/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFRiskFactorsSubmit(c, db, sl, store, config)
	})
	app.Get("/vhf-cif/laboratory/:id", func(c *fiber.Ctx) error {
		data := handlers.NewTemplateData(c, store)
		data.Form = fiber.Map{"PatientID": c.Params("id")}
		return handlers.GenerateHTML(c, db, data, "vhf_laboratory")
	})
	app.Post("/vhf-cif/laboratory/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFLaboratorySubmit(c, db, sl, store, config, smsService)
	})
	app.Get("/vhf-cif/investigator/:id", func(c *fiber.Ctx) error {
		data := handlers.NewTemplateData(c, store)
		data.Form = fiber.Map{"PatientID": c.Params("id")}
		return handlers.GenerateHTML(c, db, data, "vhf_investigator")
	})
	app.Post("/vhf-cif/investigator/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFInvestigatorSubmit(c, db, sl, store, config)
	})
	app.Get("/vhf-cif/success", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFSuccess(c, db, sl, store, config)
	})
	app.Get("/vhf-cif/view/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFView(c, db, sl, store, config)
	})
}

func RouteVerify(app *fiber.App, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) {
	app.Get("/verify/discharges/:i", func(c *fiber.Ctx) error { return handlers.VerifyDischarge2(c, db, sl, store, config) })
}

func RouteFacilities(v fiber.Router, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) {
	// Add RBAC permission checks for facility management
	v.Get("/new/:i", middleware.PermissionRequired(store, db, sl, "facilities", "create"), func(c *fiber.Ctx) error {
		return handlers.HandlerFacilityForm(c, db, sl, store, config)
	})
	v.Post("/save", middleware.PermissionRequired(store, db, sl, "facilities", "create"), func(c *fiber.Ctx) error {
		return handlers.HandlerFacilitySubmit(c, db, sl, store, config)
	})
	v.Post("/filter", middleware.PermissionRequired(store, db, sl, "facilities", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerFacilityList(c, db, sl, store, config)
	})
	v.Get("/list", middleware.PermissionRequired(store, db, sl, "facilities", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerFacilityList(c, db, sl, store, config)
	})
	v.Get("/", middleware.PermissionRequired(store, db, sl, "facilities", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerFacilityList(c, db, sl, store, config)
	})
}

func RouteUsers(v fiber.Router, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) {
	// Add RBAC permission checks for user management
	v.Get("/new/:i", middleware.PermissionRequired(store, db, sl, "users", "create"), func(c *fiber.Ctx) error {
		return handlers.HandlerUserForm(c, db, sl, store, config)
	})
	v.Post("/save", middleware.PermissionRequired(store, db, sl, "users", "create"), func(c *fiber.Ctx) error {
		return handlers.HandlerUserSubmit(c, db, sl, store, config)
	})
	v.Post("/filter", middleware.PermissionRequired(store, db, sl, "users", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerUserList(c, db, sl, store, config)
	})
	v.Get("/list", middleware.PermissionRequired(store, db, sl, "users", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerUserList(c, db, sl, store, config)
	})
	v.Get("/", middleware.PermissionRequired(store, db, sl, "users", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerUserList(c, db, sl, store, config)
	})
}

func RouteEmployees(v fiber.Router, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) {
	// Add RBAC permission checks for employee management
	v.Get("/new/:i", middleware.PermissionRequired(store, db, sl, "employees", "create"), func(c *fiber.Ctx) error {
		return handlers.HandlerEmployeeForm(c, db, sl, store, config)
	})
	v.Post("/save", middleware.PermissionRequired(store, db, sl, "employees", "create"), func(c *fiber.Ctx) error {
		return handlers.HandlerEmployeeSubmit(c, db, sl, store, config)
	})
	v.Post("/filter", middleware.PermissionRequired(store, db, sl, "employees", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerEmployeeList(c, db, sl, store, config)
	})
	v.Get("/list", middleware.PermissionRequired(store, db, sl, "employees", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerEmployeeList(c, db, sl, store, config)
	})
	v.Get("/", middleware.PermissionRequired(store, db, sl, "employees", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerEmployeeList(c, db, sl, store, config)
	})

	// API routes for employee management
	v.Get("/api/employees/:id", middleware.PermissionRequired(store, db, sl, "employees", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerGetEmployee(c, db, sl, store, config)
	})
	v.Delete("/api/employees/:id", middleware.PermissionRequired(store, db, sl, "employees", "delete"), func(c *fiber.Ctx) error {
		return handlers.HandlerDeleteEmployee(c, db, sl, store, config)
	})
}

func RouteCases(v fiber.Router, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config, smsService *services.SMSService, voiceService *services.VoiceService) {
	// Add RBAC permission checks for case management
	v.Get("/profile/:i", middleware.PermissionRequired(store, db, sl, "vhf_patients", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerPatientProfile(c, db, sl, store, config, smsService, voiceService)
	})
	v.Get("/new/:i", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
		return handlers.HandlerCasesForm(c, db, sl, store, config, smsService, voiceService)
	})
	v.Post("/save", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
		return handlers.HandlerCasesSubmit(c, db, sl, store, config)
	})
	v.Post("/filter", middleware.PermissionRequired(store, db, sl, "vhf_patients", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerCasesList(c, db, sl, store, config)
	})
	v.Get("/list", middleware.PermissionRequired(store, db, sl, "vhf_patients", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerCasesList(c, db, sl, store, config)
	})
	v.Get("/", middleware.PermissionRequired(store, db, sl, "vhf_patients", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerCasesList(c, db, sl, store, config)
	})

	// Add route for case manager redirect with outbreak ID
	v.Get("/:outbreak_id", middleware.PermissionRequired(store, db, sl, "vhf_patients", "read"), func(c *fiber.Ctx) error {
		// Set the outbreak ID in session and redirect to cases list
		outbreakID, err := strconv.Atoi(c.Params("outbreak_id"))
		if err != nil {
			return c.Status(400).SendString("Invalid outbreak ID")
		}

		// Set outbreak in session
		sess, err := store.Get(c)
		if err != nil {
			return c.Status(500).SendString("Failed to get session")
		}
		sess.Set("outbreak_id", outbreakID)
		sess.Set("selected_outbreak", outbreakID) // Set both keys for consistency
		if err := sess.Save(); err != nil {
			return c.Status(500).SendString("Failed to save session")
		}

		// Redirect to cases list
		return c.Redirect("/cases/list")
	})

	v.Get("/encounters/list/:i", middleware.PermissionRequired(store, db, sl, "vhf_patients", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerCaseEncounterForm(c, db, sl, store, config)
	})
	v.Get("/encounters/new/:i", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
		return handlers.HandlerCaseEncounterForm(c, db, sl, store, config)
	})
	v.Get("/encounters/new/:i/:j", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
		return handlers.HandlerCaseEncounterForm(c, db, sl, store, config)
	})
	v.Post("/encounters/save", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
		return handlers.HandlerCaseEncounterSubmit(c, db, sl, store, config)
	})

	v.Get("/encounters/mpox-admission/new/:i", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
		return handlers.HandlerMpoxAdmissionForm(c, db, sl, store, config)
	})
	v.Post("/encounters//save", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
		return handlers.HandlerMpoxAdmissionSubmit(c, db, sl, store, config)
	})
}

func RouteCaseDischarge(v fiber.Router, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config, smsService *services.SMSService, voiceService *services.VoiceService) {
	// Add RBAC permission checks for case discharge
	v.Get("/view/:i/:j", middleware.PermissionRequired(store, db, sl, "vhf_patients", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerCasesForm(c, db, sl, store, config, smsService, voiceService)
	})
	v.Get("/new/:i/:j", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
		return handlers.HandlerCasesForm(c, db, sl, store, config, smsService, voiceService)
	})
	v.Post("/save/:i/:j", middleware.PermissionRequired(store, db, sl, "vhf_patients", "update"), func(c *fiber.Ctx) error {
		return handlers.HandlerCasesSubmit(c, db, sl, store, config)
	})
}

func RouteSymptoms(v fiber.Router, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) {
	// Add RBAC permission checks for symptoms management
	v.Get("/new/:i", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
		return handlers.HandlerSymptomsForm(c, db, sl, store, config)
	})
	v.Post("/save", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
		return handlers.HandlerSymptomsSubmit(c, db, sl, store, config)
	})
	v.Post("/filter", middleware.PermissionRequired(store, db, sl, "vhf_patients", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerSymptomsList(c, db, sl, store, config)
	})
	v.Get("/list", middleware.PermissionRequired(store, db, sl, "vhf_patients", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerSymptomsList(c, db, sl, store, config)
	})
	v.Get("/", middleware.PermissionRequired(store, db, sl, "vhf_patients", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerSymptomsList(c, db, sl, store, config)
	})
}

func RouteMorbidity(v fiber.Router, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) {
	// Add RBAC permission checks for morbidity management
	v.Get("/new/:i", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
		return handlers.HandlerMorbidityForm(c, db, sl, store, config)
	})
	v.Post("/save", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
		return handlers.HandlerMorbiditySubmit(c, db, sl, store, config)
	})
	v.Post("/filter", middleware.PermissionRequired(store, db, sl, "vhf_patients", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerMorbidityList(c, db, sl, store, config)
	})
	v.Get("/list", middleware.PermissionRequired(store, db, sl, "vhf_patients", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerMorbidityList(c, db, sl, store, config)
	})
	v.Get("/", middleware.PermissionRequired(store, db, sl, "vhf_patients", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerMorbidityList(c, db, sl, store, config)
	})
}

func RouteRush(v fiber.Router, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) {
	// Add RBAC permission checks for rush management
	v.Get("/new/:i", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
		return handlers.HandlerRushForm(c, db, sl, store, config)
	})
	v.Post("/save", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
		return handlers.HandlerRushSubmit(c, db, sl, store, config)
	})
	v.Post("/filter", middleware.PermissionRequired(store, db, sl, "vhf_patients", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerRushList(c, db, sl, store, config)
	})
	v.Get("/list", middleware.PermissionRequired(store, db, sl, "vhf_patients", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerRushList(c, db, sl, store, config)
	})
	v.Get("/", middleware.PermissionRequired(store, db, sl, "vhf_patients", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerRushList(c, db, sl, store, config)
	})
}

func RouteLab(v fiber.Router, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) {
	// Add RBAC permission checks for laboratory management
	v.Get("/new/:i", middleware.PermissionRequired(store, db, sl, "laboratory", "create"), func(c *fiber.Ctx) error {
		return handlers.HandlerLabForm(c, db, sl, store, config)
	})
	v.Post("/save", middleware.PermissionRequired(store, db, sl, "laboratory", "create"), func(c *fiber.Ctx) error {
		return handlers.HandlerLabSubmit(c, db, sl, store, config)
	})
	v.Post("/filter", middleware.PermissionRequired(store, db, sl, "laboratory", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerLabList(c, db, sl, store, config)
	})
	v.Get("/list", middleware.PermissionRequired(store, db, sl, "laboratory", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerLabList(c, db, sl, store, config)
	})
	v.Get("/", middleware.PermissionRequired(store, db, sl, "laboratory", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerLabList(c, db, sl, store, config)
	})
}

func RouteReports(v fiber.Router, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) { //+
	// Add RBAC permission checks for reports
	v.Get("/view", middleware.PermissionRequired(store, db, sl, "reports", "read"), func(c *fiber.Ctx) error {
		return reports.GenerateReport(c, db, sl, store, config)
	}) //+
	v.Get("/", middleware.PermissionRequired(store, db, sl, "reports", "read"), func(c *fiber.Ctx) error {
		return reports.ReportsHome(c, db, sl, store, config)
	})

	// New comprehensive reports routes
	v.Post("/generate", middleware.PermissionRequired(store, db, sl, "reports", "read"), func(c *fiber.Ctx) error {
		return reports.GenerateReport(c, db, sl, store, config)
	})

	// CIF-specific reports
	v.Get("/cif/:type", middleware.PermissionRequired(store, db, sl, "reports", "read"), func(c *fiber.Ctx) error {
		return reports.CIFReports(c, db, sl, store, config)
	})

	// API endpoints for AJAX data loading
	v.Get("/api/stats", middleware.PermissionRequired(store, db, sl, "reports", "read"), func(c *fiber.Ctx) error {
		return reports.GetQuickStats(c, db, sl, store, config)
	})

	v.Get("/api/chart-data", middleware.PermissionRequired(store, db, sl, "reports", "read"), func(c *fiber.Ctx) error {
		return reports.GetChartData(c, db, sl, store, config)
	})

	v.Get("/api/table-data", middleware.PermissionRequired(store, db, sl, "reports", "read"), func(c *fiber.Ctx) error {
		return reports.GetTableData(c, db, sl, store, config)
	})

	// Export functionality
	v.Post("/export", middleware.PermissionRequired(store, db, sl, "reports", "read"), func(c *fiber.Ctx) error {
		return reports.ExportReport(c, db, sl, store, config)
	})
}

// Add this new function for outbreak routes
func RouteOutbreaks(app *fiber.App, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) {
	// This function is kept for compatibility but routes are now defined in SetRoute
	// to avoid duplicate route definitions
}

// SetupRoutes configures all routes for the application
func SetupRoutes(app *fiber.App, db *sql.DB, store *session.Store, sl *slog.Logger, config handlers.Config, smsService *services.SMSService) {
	// Public routes
	app.Get("/", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "landing")
	})

	// Login routes
	app.Get("/login", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "login")
	})
	app.Post("/login", func(c *fiber.Ctx) error {
		return handlers.HandlerLoginSubmit(c, db, sl, store, config)
	})
	app.Get("/logout", func(c *fiber.Ctx) error {
		return handlers.HandlerLoginOut(c, sl, store, config)
	})

	app.Get("/vhf-cif", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "vhf_cif")
	})
	app.Post("/vhf-cif/save", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFPatientSubmit(c, db, sl, store, config, smsService)
	})
	app.Post("/vhf-cif/update", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFUpdate(c, db, sl, store, config)
	})
	app.Get("/vhf-cif/clinical-signs/:id", func(c *fiber.Ctx) error {
		data := handlers.NewTemplateData(c, store)
		data.Form = fiber.Map{"PatientID": c.Params("id")}
		return handlers.GenerateHTML(c, db, data, "vhf_clinical_signs")
	})
	app.Post("/vhf-cif/clinical-signs/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFClinicalSignsSubmit(c, db, sl, store, config)
	})
	app.Get("/vhf-cif/hospitalization/:id", func(c *fiber.Ctx) error {
		data := handlers.NewTemplateData(c, store)
		data.Form = fiber.Map{"PatientID": c.Params("id")}
		return handlers.GenerateHTML(c, db, data, "vhf_hospitalization")
	})
	app.Post("/vhf-cif/hospitalization/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFHospitalizationSubmit(c, db, sl, store, config)
	})
	app.Get("/vhf-cif/risk-factors/:id", func(c *fiber.Ctx) error {
		data := handlers.NewTemplateData(c, store)
		data.Form = fiber.Map{"PatientID": c.Params("id")}
		return handlers.GenerateHTML(c, db, data, "vhf_risk_factors")
	})
	app.Post("/vhf-cif/risk-factors/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFRiskFactorsSubmit(c, db, sl, store, config)
	})
	app.Get("/vhf-cif/laboratory/:id", func(c *fiber.Ctx) error {
		data := handlers.NewTemplateData(c, store)
		data.Form = fiber.Map{"PatientID": c.Params("id")}
		return handlers.GenerateHTML(c, db, data, "vhf_laboratory")
	})
	app.Post("/vhf-cif/laboratory/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFLaboratorySubmit(c, db, sl, store, config, smsService)
	})
	app.Get("/vhf-cif/investigator/:id", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "vhf_investigator")
	})
	app.Post("/vhf-cif/investigator/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFInvestigatorSubmit(c, db, sl, store, config)
	})
	app.Get("/vhf-cif/success", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFSuccess(c, db, sl, store, config)
	})
	app.Get("/vhf-cif/view/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFView(c, db, sl, store, config)
	})
	app.Get("/measles_success", func(c *fiber.Ctx) error { return handlers.HandlerMeaslesSuccess(c, store) })
	app.Get("/measles_cif", func(c *fiber.Ctx) error {
		return handlers.HandlerMeaslesCIF(c, db, store)
	})
	app.Post("/measles_cif", func(c *fiber.Ctx) error {
		return handlers.HandlerMeaslesCIF(c, db, store)
	})

	// Location API routes
	app.Get("/api/locations/districts", func(c *fiber.Ctx) error {
		return handlers.HandlerGetDistricts(c, db, sl)
	})
	app.Get("/api/locations/subcounties/:district_id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetSubcountiesByDistrict(c, db, sl)
	})
	app.Get("/api/locations/parishes/:subcounty_id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetParishesBySubcounty(c, db, sl)
	})
	app.Get("/api/locations/parishes/district/:district_id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetParishesByDistrict(c, db, sl)
	})
	app.Get("/api/locations/villages/:parish_id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetVillagesByParish(c, db, sl)
	})
	app.Get("/api/locations/villages/district/:district_id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetVillagesByDistrict(c, db, sl)
	})
	app.Get("/api/locations/villages/subcounty/:subcounty_id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetVillagesBySubcounty(c, db, sl)
	})

	// Test routes (defined before protected group to avoid conflicts)
	app.Get("/test-simple", func(c *fiber.Ctx) error {
		return c.SendString("Simple route working!")
	})

	// Inventory routes (moved to top to avoid conflicts)
	inventoryHandler := handlers.NewInventoryHandler(db, store)

	// Inventory dashboard - test without middleware first
	app.Get("/inventory-test", func(c *fiber.Ctx) error {
		log.Printf("DEBUG: /inventory-test route accessed")
		return inventoryHandler.HandlerInventoryDashboard(c)
	})

	// Test endpoint to verify the route is working
	app.Get("/api/test-lab", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "success",
			"message":   "Lab API test endpoint working",
			"timestamp": time.Now(),
		})
	})

	// Surveillance routes - Added before protected routes group
	app.Get("/surveillance/community-mortality", AuthRequired(store), func(c *fiber.Ctx) error {
		log.Printf("DEBUG: /surveillance/community-mortality route accessed")
		return handlers.CommunityMortalitySurveillance(c, db, store, config)
	})
	app.Get("/surveillance/facility-mortality", AuthRequired(store), func(c *fiber.Ctx) error {
		log.Printf("DEBUG: /surveillance/facility-mortality route accessed")
		return handlers.FacilityMortalitySurveillance(c, db, store, config)
	})

	// Protected routes
	protected := app.Group("/", AuthRequired(store))

	// Inventory dashboard - protected by authentication
	protected.Get("/inventory", func(c *fiber.Ctx) error {
		log.Printf("DEBUG: /inventory route accessed")
		return inventoryHandler.HandlerInventoryDashboard(c)
	})

	// Test protected route
	protected.Get("/test-protected", func(c *fiber.Ctx) error {
		return c.SendString("Protected route working!")
	})

	// Lab Sample Types API routes (protected - require authentication)
	protected.Get("/api/lab/swab-types", func(c *fiber.Ctx) error {
		return handlers.HandlerAPIGetLabSwabTypes(c, db, sl)
	})
	protected.Get("/api/lab/urine-types", func(c *fiber.Ctx) error {
		return handlers.HandlerAPIGetLabUrineTypes(c, db, sl)
	})
	protected.Get("/api/lab/blood-types", func(c *fiber.Ctx) error {
		return handlers.HandlerAPIGetLabBloodTypes(c, db, sl)
	})
	protected.Get("/api/lab/blood-types/category/:category", func(c *fiber.Ctx) error {
		return handlers.HandlerAPIGetLabBloodTypesByCategory(c, db, sl)
	})

	// Lab sample selections routes (protected - require authentication)
	protected.Post("/api/lab/sample-selections", func(c *fiber.Ctx) error {
		return handlers.HandlerAPISaveLabSampleSelections(c, db, sl)
	})
	protected.Get("/api/lab/sample-selections/:lab_id", func(c *fiber.Ctx) error {
		return handlers.HandlerAPIGetLabSampleSelections(c, db, sl)
	})

	// VHF CIF routes
	protected.Get("/vhf-cif", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "vhf_cif")
	})
	protected.Post("/vhf-cif", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFPatientSubmit(c, db, sl, store, config, smsService)
	})
	protected.Get("/vhf-cif/clinical-signs/:id", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "vhf_clinical_signs")
	})
	protected.Post("/vhf-cif/clinical-signs/:id", func(c *fiber.Ctx) error { return handlers.HandlerVHFClinicalSignsSubmit(c, db, sl, store, config) })
	protected.Get("/vhf-cif/hospitalization/:id", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "vhf_hospitalization")
	})
	protected.Post("/vhf-cif/hospitalization/:id", func(c *fiber.Ctx) error { return handlers.HandlerVHFHospitalizationSubmit(c, db, sl, store, config) })
	protected.Get("/vhf-cif/risk-factors/:id", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "vhf_risk_factors")
	})
	protected.Post("/vhf-cif/risk-factors/:id", func(c *fiber.Ctx) error { return handlers.HandlerVHFRiskFactorsSubmit(c, db, sl, store, config) })
	protected.Get("/vhf-cif/laboratory/:id", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "vhf_laboratory")
	})
	protected.Post("/vhf-cif/laboratory/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFLaboratorySubmit(c, db, sl, store, config, smsService)
	})
	protected.Get("/vhf-cif/investigator/:id", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "vhf_investigator")
	})
	protected.Post("/vhf-cif/investigator/:id", func(c *fiber.Ctx) error { return handlers.HandlerVHFInvestigatorSubmit(c, db, sl, store, config) })
	protected.Get("/vhf-cif/success", func(c *fiber.Ctx) error { return handlers.HandlerVHFSuccess(c, db, sl, store, config) })
	protected.Get("/vhf/list", func(c *fiber.Ctx) error { return handlers.HandlerVHFList(c, db, sl, store, config) })
	protected.Get("/measles/list", func(c *fiber.Ctx) error { return handlers.HandlerMeaslesList(c, db, sl, store, config) })
	protected.Get("/mpox/list", func(c *fiber.Ctx) error { return handlers.HandlerMpoxList(c, db, sl, store, config) })
	protected.Get("/vhf/view/:id", func(c *fiber.Ctx) error { return handlers.HandlerVHFView(c, db, sl, store, config) })
	protected.Get("/vhf-lab/:id", func(c *fiber.Ctx) error { return handlers.HandlerVHFLabForm(c, db, sl, store, config) })
	protected.Post("/vhf-lab/:id", func(c *fiber.Ctx) error { return handlers.HandlerVHFLabSave(c, db, sl, store, config) })

	// Location routes
	app.Get("/api/districts", func(c *fiber.Ctx) error {
		return handlers.HandlerGetDistricts(c, db, sl)
	})

	app.Get("/api/subcounties/:district_id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetSubcountiesByDistrict(c, db, sl)
	})

	app.Get("/api/parishes/:subcounty_id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetParishesBySubcounty(c, db, sl)
	})

	app.Get("/api/parishes/district/:district_id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetParishesByDistrict(c, db, sl)
	})

	app.Get("/api/villages/:parish_id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetVillagesByParish(c, db, sl)
	})

	app.Get("/api/villages/district/:district_id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetVillagesByDistrict(c, db, sl)
	})

	app.Get("/api/villages/subcounty/:subcounty_id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetVillagesBySubcounty(c, db, sl)
	})

	// VHF API routes
	protected.Get("/api/vhf-cases", func(c *fiber.Ctx) error { return handlers.HandlerVHFList(c, db, sl, store, config) })
	protected.Get("/api/vhf-cases/:id", func(c *fiber.Ctx) error { return handlers.HandlerVHFView(c, db, sl, store, config) })

	// Inventory items
	protected.Get("/inventory", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "inventory_dashboard")
	})
	protected.Get("/inventory/items", func(c *fiber.Ctx) error { return inventoryHandler.HandlerInventoryItemsList(c) })
	protected.Get("/inventory/items/new", func(c *fiber.Ctx) error { return inventoryHandler.HandlerInventoryItemForm(c) })
	protected.Get("/inventory/items/edit/:id", func(c *fiber.Ctx) error { return inventoryHandler.HandlerInventoryItemForm(c) })
	protected.Post("/inventory/items/save", func(c *fiber.Ctx) error { return inventoryHandler.HandlerInventoryItemSave(c) })

	// Inventory stock management
	protected.Get("/inventory/stock", func(c *fiber.Ctx) error { return inventoryHandler.HandlerInventoryStockForm(c) })
	protected.Post("/inventory/stock/save", func(c *fiber.Ctx) error { return inventoryHandler.HandlerInventoryStockSave(c) })

	// Purchase orders
	protected.Get("/inventory/purchase-orders", func(c *fiber.Ctx) error { return inventoryHandler.HandlerInventoryPurchaseOrderForm(c) })
	protected.Post("/inventory/purchase-orders/save", func(c *fiber.Ctx) error { return inventoryHandler.HandlerInventoryPurchaseOrderSave(c) })

	// Requisitions
	protected.Get("/inventory/requisitions", func(c *fiber.Ctx) error { return inventoryHandler.HandlerInventoryRequisitionForm(c) })
	protected.Post("/inventory/requisitions/save", func(c *fiber.Ctx) error { return inventoryHandler.HandlerInventoryRequisitionSave(c) })

	// Reports
	protected.Get("/inventory/reports", func(c *fiber.Ctx) error { return inventoryHandler.HandlerInventoryReports(c) })

	// Donation routes
	protected.Get("/inventory/donations", func(c *fiber.Ctx) error { return inventoryHandler.HandlerDonationsList(c) })
	protected.Get("/inventory/donations/new", func(c *fiber.Ctx) error { return inventoryHandler.HandlerDonationForm(c) })
	protected.Post("/inventory/donations/save", func(c *fiber.Ctx) error { return inventoryHandler.HandlerDonationSave(c) })
	protected.Get("/inventory/donations/:id", func(c *fiber.Ctx) error { return inventoryHandler.HandlerDonationView(c) })
	protected.Get("/inventory/donors", func(c *fiber.Ctx) error { return inventoryHandler.HandlerDonorsList(c) })
	protected.Get("/inventory/donors/new", func(c *fiber.Ctx) error { return inventoryHandler.HandlerDonorForm(c) })
	protected.Post("/inventory/donors/save", func(c *fiber.Ctx) error { return inventoryHandler.HandlerDonorSave(c) })

	// Inventory API routes
	protected.Get("/api/inventory/items", func(c *fiber.Ctx) error { return inventoryHandler.HandlerInventoryAPIItems(c) })
	protected.Get("/api/inventory/stock-levels", func(c *fiber.Ctx) error { return inventoryHandler.HandlerInventoryAPIStockLevels(c) })
	protected.Get("/api/inventory/low-stock", func(c *fiber.Ctx) error { return inventoryHandler.HandlerInventoryAPILowStock(c) })

	// Reports routes - More specific routes must come first
	protected.Get("/reports/dashboard", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "reports_navigation")
	})
	protected.Get("/reports/cif-dashboard", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "cif_dashboard")
	})
	protected.Get("/reports/overview", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "reports_navigation")
	})
	protected.Get("/reports/trend-analysis", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "trend_analysis")
	})
	protected.Get("/reports/demographics", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "demographics_report")
	})
	protected.Get("/reports/outcome-analysis", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "reports_navigation")
	})
	protected.Get("/reports/vhf-cif", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "vhf_cif_report")
	})
	protected.Get("/reports/measles-cif", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "reports_navigation")
	})
	protected.Get("/reports/polio-cif", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "reports_navigation")
	})
	protected.Get("/reports/mpox-cif", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "reports_navigation")
	})
	protected.Get("/reports/treatment-protocols", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "reports_navigation")
	})
	protected.Get("/reports/test-results", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "reports_navigation")
	})
	protected.Get("/reports/specimen-analysis", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "reports_navigation")
	})
	protected.Get("/reports/case-distribution", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "reports_navigation")
	})
	protected.Get("/reports/hotspot-analysis", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "reports_navigation")
	})

	// API routes for reports data
	protected.Get("/reports/quick-stats", func(c *fiber.Ctx) error {
		return reports.GetQuickStats(c, db, sl, store, config)
	})
	protected.Get("/reports/health-indicators", func(c *fiber.Ctx) error {
		return reports.GetHealthIndicators(c, db, sl, store, config)
	})
	protected.Get("/reports/chart-data/:type", func(c *fiber.Ctx) error {
		return reports.GetChartData(c, db, sl, store, config)
	})
	protected.Get("/reports/table-data", func(c *fiber.Ctx) error {
		return reports.GetTableData(c, db, sl, store, config)
	})

	// VHF-specific API routes
	protected.Get("/reports/vhf-stats", func(c *fiber.Ctx) error {
		return reports.GetVHFStats(c, db, sl, store, config)
	})
	protected.Get("/reports/vhf-trends", func(c *fiber.Ctx) error {
		return reports.GetVHFTrends(c, db, sl, store, config)
	})
	protected.Get("/reports/vhf-status", func(c *fiber.Ctx) error {
		return reports.GetVHFStatus(c, db, sl, store, config)
	})
	protected.Get("/reports/vhf-gender", func(c *fiber.Ctx) error {
		return reports.GetVHFGender(c, db, sl, store, config)
	})
	protected.Get("/reports/vhf-age", func(c *fiber.Ctx) error {
		return reports.GetVHFAge(c, db, sl, store, config)
	})
	protected.Get("/reports/vhf-districts", func(c *fiber.Ctx) error {
		return reports.GetVHFDistricts(c, db, sl, store, config)
	})
	protected.Get("/reports/vhf-cases", func(c *fiber.Ctx) error {
		return reports.GetVHFCases(c, db, sl, store, config)
	})

	// Demographics API routes
	protected.Get("/reports/demographics-stats", func(c *fiber.Ctx) error {
		return reports.GetDemographicsStats(c, db, sl, store, config)
	})
	protected.Get("/reports/gender-distribution", func(c *fiber.Ctx) error {
		return reports.GetGenderDistribution(c, db, sl, store, config)
	})
	protected.Get("/reports/age-group-distribution", func(c *fiber.Ctx) error {
		return reports.GetAgeGroupDistribution(c, db, sl, store, config)
	})
	protected.Get("/reports/age-distribution", func(c *fiber.Ctx) error {
		return reports.GetAgeDistribution(c, db, sl, store, config)
	})
	protected.Get("/reports/district-distribution", func(c *fiber.Ctx) error {
		return reports.GetDistrictDistribution(c, db, sl, store, config)
	})
	protected.Get("/reports/facility-distribution", func(c *fiber.Ctx) error {
		return reports.GetFacilityDistribution(c, db, sl, store, config)
	})
	protected.Get("/reports/occupation-distribution", func(c *fiber.Ctx) error {
		return reports.GetOccupationDistribution(c, db, sl, store, config)
	})
	protected.Get("/reports/demographics-table", func(c *fiber.Ctx) error {
		return reports.GetDemographicsTable(c, db, sl, store, config)
	})

	// Trend analysis API routes
	protected.Get("/reports/trend-stats", func(c *fiber.Ctx) error {
		return reports.GetTrendStats(c, db, sl, store, config)
	})
	protected.Get("/reports/trend-data", func(c *fiber.Ctx) error {
		return reports.GetTrendData(c, db, sl, store, config)
	})
	protected.Get("/reports/weekly-comparison", func(c *fiber.Ctx) error {
		return reports.GetWeeklyComparison(c, db, sl, store, config)
	})
	protected.Get("/reports/disease-distribution", func(c *fiber.Ctx) error {
		return reports.GetDiseaseDistribution(c, db, sl, store, config)
	})
	protected.Get("/reports/geographic-trends", func(c *fiber.Ctx) error {
		return reports.GetGeographicTrends(c, db, sl, store, config)
	})

	// CIF API routes
	protected.Get("/reports/cif-stats", func(c *fiber.Ctx) error {
		return reports.GetCIFStats(c, db, sl, store, config)
	})
	protected.Get("/reports/cif-status-chart", func(c *fiber.Ctx) error {
		return reports.GetCIFStatusChart(c, db, sl, store, config)
	})
	protected.Get("/reports/cif-type-chart", func(c *fiber.Ctx) error {
		return reports.GetCIFTypeChart(c, db, sl, store, config)
	})
	protected.Get("/reports/recent-cifs", func(c *fiber.Ctx) error {
		return reports.GetRecentCIFs(c, db, sl, store, config)
	})

	protected.Post("/reports/generate", func(c *fiber.Ctx) error {
		return reports.GenerateReport(c, db, sl, store, config)
	})
	protected.Post("/reports/export", func(c *fiber.Ctx) error {
		return reports.ExportReport(c, db, sl, store, config)
	})
	protected.Get("/reports", func(c *fiber.Ctx) error {
		return reports.ReportsHome(c, db, sl, store, config)
	})

	// ===== COMPREHENSIVE API ENDPOINTS FOR ALL ROUTES =====
	// These endpoints mirror all existing routes for Next.js migration

	// Authentication & User Management APIs
	protected.Get("/api/auth/user", func(c *fiber.Ctx) error {
		return handlers.HandlerGetCurrentUser(c, db, sl, store, config)
	})
	protected.Post("/api/auth/logout", func(c *fiber.Ctx) error {
		return handlers.HandlerLoginOut(c, sl, store, config)
	})
	protected.Post("/api/auth/change-password", func(c *fiber.Ctx) error {
		return handlers.HandlerChangePassword(c, db, sl, store, config)
	})

	// Dashboard & Home APIs
	protected.Get("/api/dashboard/home", func(c *fiber.Ctx) error {
		return handlers.HandlerHomeAPI(c, db, sl, store, config)
	})
	protected.Get("/api/dashboard/stats", func(c *fiber.Ctx) error {
		return handlers.HandlerDashboardStats(c, db, sl, store, config)
	})

	// VHF Case Management APIs
	protected.Get("/api/vhf/patients", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFListAPI(c, db, sl, store, config)
	})
	protected.Get("/api/vhf/patients/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFViewAPI(c, db, sl, store, config)
	})
	protected.Post("/api/vhf/patients", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFPatientSubmitAPI(c, db, sl, store, config, smsService)
	})
	protected.Put("/api/vhf/patients/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFUpdateAPI(c, db, sl, store, config)
	})
	protected.Delete("/api/vhf/patients/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFDeleteAPI(c, db, sl, store, config)
	})

	// VHF Clinical Signs APIs
	protected.Get("/api/vhf/patients/:id/clinical-signs", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFClinicalSignsAPI(c, db, sl, store, config)
	})
	protected.Post("/api/vhf/patients/:id/clinical-signs", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFClinicalSignsSubmitAPI(c, db, sl, store, config)
	})

	// VHF Hospitalization APIs
	protected.Get("/api/vhf/patients/:id/hospitalization", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFHospitalizationAPI(c, db, sl, store, config)
	})
	protected.Post("/api/vhf/patients/:id/hospitalization", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFHospitalizationSubmitAPI(c, db, sl, store, config)
	})

	// VHF Risk Factors APIs
	protected.Get("/api/vhf/patients/:id/risk-factors", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFRiskFactorsAPI(c, db, sl, store, config)
	})
	protected.Post("/api/vhf/patients/:id/risk-factors", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFRiskFactorsSubmitAPI(c, db, sl, store, config)
	})

	// VHF Laboratory APIs
	protected.Get("/api/vhf/patients/:id/laboratory", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFLaboratoryAPI(c, db, sl, store, config)
	})
	protected.Post("/api/vhf/patients/:id/laboratory", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFLaboratorySubmitAPI(c, db, sl, store, config, smsService)
	})

	// VHF Investigator APIs
	protected.Get("/api/vhf/patients/:id/investigator", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFInvestigatorAPI(c, db, sl, store, config)
	})
	protected.Post("/api/vhf/patients/:id/investigator", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFInvestigatorSubmitAPI(c, db, sl, store, config)
	})

	// VHF Lab Form APIs
	protected.Get("/api/vhf/lab/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFLabFormAPI(c, db, sl, store, config)
	})
	protected.Post("/api/vhf/lab/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFLabSaveAPI(c, db, sl, store, config)
	})

	// Employee Management APIs (session + employees:* RBAC)
	protected.Get("/api/employees", middleware.PermissionRequired(store, db, sl, "employees", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerEmployeeListAPI(c, db, sl, store, config)
	})
	protected.Get("/api/employees/:id", middleware.PermissionRequired(store, db, sl, "employees", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerGetEmployeeAPI(c, db, sl, store, config)
	})
	protected.Post("/api/employees", middleware.PermissionRequired(store, db, sl, "employees", "create"), func(c *fiber.Ctx) error {
		return handlers.HandlerEmployeeSubmitAPI(c, db, sl, store, config)
	})
	protected.Put("/api/employees/:id", middleware.PermissionRequired(store, db, sl, "employees", "update"), func(c *fiber.Ctx) error {
		return handlers.HandlerEmployeeUpdateAPI(c, db, sl, store, config)
	})
	protected.Delete("/api/employees/:id", middleware.PermissionRequired(store, db, sl, "employees", "delete"), func(c *fiber.Ctx) error {
		return handlers.HandlerDeleteEmployeeAPI(c, db, sl, store, config)
	})

	// User Management APIs
	protected.Get("/api/users", func(c *fiber.Ctx) error {
		return handlers.HandlerUserListAPI(c, db, sl, store, config)
	})
	protected.Get("/api/users/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetUserAPI(c, db, sl, store, config)
	})
	protected.Post("/api/users", func(c *fiber.Ctx) error {
		return handlers.HandlerUserSubmitAPI(c, db, sl, store, config)
	})
	protected.Put("/api/users/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerUserUpdateAPI(c, db, sl, store, config)
	})
	protected.Delete("/api/users/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerUserDeleteAPI(c, db, sl, store, config)
	})

	// Facility Management APIs
	protected.Get("/api/facilities", func(c *fiber.Ctx) error {
		return handlers.HandlerFacilityListAPI(c, db, sl, store, config)
	})
	protected.Get("/api/facilities/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetFacilityAPI(c, db, sl, store, config)
	})
	protected.Post("/api/facilities", func(c *fiber.Ctx) error {
		return handlers.HandlerFacilitySubmitAPI(c, db, sl, store, config)
	})
	protected.Put("/api/facilities/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerFacilityUpdateAPI(c, db, sl, store, config)
	})
	protected.Delete("/api/facilities/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerFacilityDeleteAPI(c, db, sl, store, config)
	})

	// Outbreak Management APIs
	protected.Get("/api/outbreaks", func(c *fiber.Ctx) error {
		return handlers.HandlerOutbreakListAPI(c, db, sl, store, config)
	})
	protected.Get("/api/outbreaks/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetOutbreakAPI(c, db, sl, store, config)
	})
	protected.Post("/api/outbreaks", func(c *fiber.Ctx) error {
		return handlers.HandlerOutbreakSubmitAPI(c, db, sl, store, config)
	})
	protected.Put("/api/outbreaks/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerOutbreakUpdateAPI(c, db, sl, store, config)
	})
	protected.Delete("/api/outbreaks/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerOutbreakDeleteAPI(c, db, sl, store, config)
	})
	protected.Post("/api/outbreaks/:id/close", func(c *fiber.Ctx) error {
		return handlers.HandlerOutbreakCloseAPI(c, db, sl, store, config)
	})
	protected.Post("/api/outbreaks/:id/select", func(c *fiber.Ctx) error {
		return handlers.HandlerOutbreakSelectAPI(c, db, sl, store, config)
	})

	// Case Management APIs
	protected.Get("/api/cases", func(c *fiber.Ctx) error {
		return handlers.HandlerCasesListAPI(c, db, sl, store, config)
	})
	protected.Get("/api/cases/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetCaseAPI(c, db, sl, store, config)
	})
	protected.Post("/api/cases", func(c *fiber.Ctx) error {
		return handlers.HandlerCasesSubmitAPI(c, db, sl, store, config)
	})
	protected.Put("/api/cases/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerCaseUpdateAPI(c, db, sl, store, config)
	})
	protected.Delete("/api/cases/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerCaseDeleteAPI(c, db, sl, store, config)
	})

	// Disease-specific CIF APIs
	// VHF
	protected.Get("/api/vhf/cif/:id", func(c *fiber.Ctx) error { return handlers.HandlerVhfCIFByID(c, db, sl, store, config) })
	protected.Get("/api/vhf/cif", func(c *fiber.Ctx) error { return handlers.HandlerVhfCIFByCaseCode(c, db, sl, store, config) })
	// Backward-compatible aliases
	protected.Get("/api/cif/:id", func(c *fiber.Ctx) error { return handlers.HandlerVhfCIFByID(c, db, sl, store, config) })
	protected.Get("/api/cif", func(c *fiber.Ctx) error { return handlers.HandlerVhfCIFByCaseCode(c, db, sl, store, config) })
	// Measles
	protected.Get("/api/measles/cif/:id", func(c *fiber.Ctx) error { return handlers.HandlerMeaslesCIFByID(c, db, sl, store, config) })
	protected.Get("/api/measles/cif", func(c *fiber.Ctx) error { return handlers.HandlerMeaslesCIFByCaseCode(c, db, sl, store, config) })
	// Polio
	protected.Get("/api/polio/cif/:id", func(c *fiber.Ctx) error { return handlers.HandlerPolioCIFByID(c, db, sl, store, config) })
	protected.Get("/api/polio/cif", func(c *fiber.Ctx) error { return handlers.HandlerPolioCIFByCaseCode(c, db, sl, store, config) })
	// Mpox
	protected.Get("/api/mpox/cif/:id", func(c *fiber.Ctx) error { return handlers.HandlerMpoxCIFByID(c, db, sl, store, config) })
	protected.Get("/api/mpox/cif", func(c *fiber.Ctx) error { return handlers.HandlerMpoxCIFByCaseCode(c, db, sl, store, config) })

	// Case Encounter APIs
	protected.Get("/api/cases/:id/encounters", func(c *fiber.Ctx) error {
		return handlers.HandlerCaseEncounterListAPI(c, db, sl, store, config)
	})
	protected.Get("/api/cases/:id/encounters/:encounter_id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetCaseEncounterAPI(c, db, sl, store, config)
	})
	protected.Post("/api/cases/:id/encounters", func(c *fiber.Ctx) error {
		return handlers.HandlerCaseEncounterSubmitAPI(c, db, sl, store, config)
	})
	protected.Put("/api/cases/:id/encounters/:encounter_id", func(c *fiber.Ctx) error {
		return handlers.HandlerCaseEncounterUpdateAPI(c, db, sl, store, config)
	})
	protected.Delete("/api/cases/:id/encounters/:encounter_id", func(c *fiber.Ctx) error {
		return handlers.HandlerCaseEncounterDeleteAPI(c, db, sl, store, config)
	})

	// Discharge Management APIs
	protected.Get("/api/discharges", func(c *fiber.Ctx) error {
		return handlers.GetDischargeAPI(c, db, sl, store, config)
	})
	protected.Get("/api/discharges/:id", func(c *fiber.Ctx) error {
		return handlers.GetDischargeByIdAPI(c, db, sl, store, config)
	})
	protected.Post("/api/discharges", func(c *fiber.Ctx) error {
		return handlers.DischargeAPI(c, db, sl, store, config)
	})
	protected.Get("/api/discharges/:id/certificate", func(c *fiber.Ctx) error {
		return handlers.CertificateAPI(c, db, sl, store, config)
	})

	// Laboratory Management APIs
	protected.Get("/api/laboratory", func(c *fiber.Ctx) error {
		return handlers.HandlerLabListAPI(c, db, sl, store, config)
	})
	protected.Get("/api/laboratory/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetLabAPI(c, db, sl, store, config)
	})
	protected.Post("/api/laboratory", func(c *fiber.Ctx) error {
		return handlers.HandlerLabSubmitAPI(c, db, sl, store, config)
	})
	protected.Put("/api/laboratory/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerLabUpdateAPI(c, db, sl, store, config)
	})
	protected.Delete("/api/laboratory/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerLabDeleteAPI(c, db, sl, store, config)
	})

	// Symptoms Management APIs
	protected.Get("/api/symptoms", func(c *fiber.Ctx) error {
		return handlers.HandlerSymptomsListAPI(c, db, sl, store, config)
	})
	protected.Get("/api/symptoms/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetSymptomsAPI(c, db, sl, store, config)
	})
	protected.Post("/api/symptoms", func(c *fiber.Ctx) error {
		return handlers.HandlerSymptomsSubmitAPI(c, db, sl, store, config)
	})
	protected.Put("/api/symptoms/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerSymptomsUpdateAPI(c, db, sl, store, config)
	})
	protected.Delete("/api/symptoms/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerSymptomsDeleteAPI(c, db, sl, store, config)
	})

	// Morbidity Management APIs
	protected.Get("/api/morbidity", func(c *fiber.Ctx) error {
		return handlers.HandlerMorbidityListAPI(c, db, sl, store, config)
	})
	protected.Get("/api/morbidity/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetMorbidityAPI(c, db, sl, store, config)
	})
	protected.Post("/api/morbidity", func(c *fiber.Ctx) error {
		return handlers.HandlerMorbiditySubmitAPI(c, db, sl, store, config)
	})
	protected.Put("/api/morbidity/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerMorbidityUpdateAPI(c, db, sl, store, config)
	})
	protected.Delete("/api/morbidity/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerMorbidityDeleteAPI(c, db, sl, store, config)
	})

	// Rush Management APIs
	protected.Get("/api/rush", func(c *fiber.Ctx) error {
		return handlers.HandlerRushListAPI(c, db, sl, store, config)
	})
	protected.Get("/api/rush/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetRushAPI(c, db, sl, store, config)
	})
	protected.Post("/api/rush", func(c *fiber.Ctx) error {
		return handlers.HandlerRushSubmitAPI(c, db, sl, store, config)
	})
	protected.Put("/api/rush/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerRushUpdateAPI(c, db, sl, store, config)
	})
	protected.Delete("/api/rush/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerRushDeleteAPI(c, db, sl, store, config)
	})

	// Inventory Management APIs
	protected.Get("/api/inventory/dashboard", func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerInventoryDashboardAPI(c)
	})
	protected.Get("/api/inventory/items", func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerInventoryAPIItems(c)
	})
	protected.Get("/api/inventory/items/:id", func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerGetInventoryItemAPI(c)
	})
	protected.Post("/api/inventory/items", func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerInventoryItemSaveAPI(c)
	})
	protected.Put("/api/inventory/items/:id", func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerInventoryItemUpdateAPI(c)
	})
	protected.Delete("/api/inventory/items/:id", func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerInventoryItemDeleteAPI(c)
	})

	// Inventory Stock APIs
	protected.Get("/api/inventory/stock", func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerInventoryAPIStockLevels(c)
	})
	protected.Get("/api/inventory/stock/:id", func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerGetInventoryStockAPI(c)
	})
	protected.Post("/api/inventory/stock", func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerInventoryStockSaveAPI(c)
	})
	protected.Put("/api/inventory/stock/:id", func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerInventoryStockUpdateAPI(c)
	})

	// Inventory Purchase Orders APIs
	protected.Get("/api/inventory/purchase-orders", func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerInventoryPurchaseOrdersAPI(c)
	})
	protected.Get("/api/inventory/purchase-orders/:id", func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerGetPurchaseOrderAPI(c)
	})
	protected.Post("/api/inventory/purchase-orders", func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerInventoryPurchaseOrderSaveAPI(c)
	})
	protected.Put("/api/inventory/purchase-orders/:id", func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerPurchaseOrderUpdateAPI(c)
	})

	// Inventory Requisitions APIs
	protected.Get("/api/inventory/requisitions", func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerInventoryRequisitionsAPI(c)
	})
	protected.Get("/api/inventory/requisitions/:id", func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerGetRequisitionAPI(c)
	})
	protected.Post("/api/inventory/requisitions", func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerInventoryRequisitionSaveAPI(c)
	})
	protected.Put("/api/inventory/requisitions/:id", func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerRequisitionUpdateAPI(c)
	})

	// Inventory Donations APIs
	protected.Get("/api/inventory/donations", func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerDonationsListAPI(c)
	})
	protected.Get("/api/inventory/donations/:id", func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerDonationViewAPI(c)
	})
	protected.Post("/api/inventory/donations", func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerDonationSaveAPI(c)
	})
	protected.Put("/api/inventory/donations/:id", func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerDonationUpdateAPI(c)
	})
	protected.Delete("/api/inventory/donations/:id", func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerDonationDeleteAPI(c)
	})

	// Inventory Donors APIs
	protected.Get("/api/inventory/donors", func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerDonorsListAPI(c)
	})
	protected.Get("/api/inventory/donors/:id", func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerGetDonorAPI(c)
	})
	protected.Post("/api/inventory/donors", func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerDonorSaveAPI(c)
	})
	protected.Put("/api/inventory/donors/:id", func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerDonorUpdateAPI(c)
	})
	protected.Delete("/api/inventory/donors/:id", func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerDonorDeleteAPI(c)
	})

	// Surveillance APIs
	protected.Get("/api/surveillance/community-mortality", func(c *fiber.Ctx) error {
		return handlers.CommunityMortalitySurveillanceAPI(c, db, store, config)
	})
	protected.Get("/api/surveillance/facility-mortality", func(c *fiber.Ctx) error {
		return handlers.FacilityMortalitySurveillanceAPI(c, db, store, config)
	})

	// Outbreak Assignment APIs
	protected.Get("/api/outbreaks/assignments", func(c *fiber.Ctx) error {
		userService := models.NewUserService(db)
		userOutbreakService := models.NewUserOutbreakService(db)
		patientRoleService := models.NewPatientManagementRoleService(db)
		outbreakService := models.NewOutbreakService(db)
		facilityService := models.NewFacilityService(db)
		handler := handlers.NewOutbreakAssignmentHandler(
			userOutbreakService, patientRoleService, userService, outbreakService, facilityService, store,
		)
		return handler.ShowOutbreakAssignmentsAPI(c)
	})
	protected.Post("/api/outbreaks/assign", func(c *fiber.Ctx) error {
		userService := models.NewUserService(db)
		userOutbreakService := models.NewUserOutbreakService(db)
		patientRoleService := models.NewPatientManagementRoleService(db)
		outbreakService := models.NewOutbreakService(db)
		facilityService := models.NewFacilityService(db)
		handler := handlers.NewOutbreakAssignmentHandler(
			userOutbreakService, patientRoleService, userService, outbreakService, facilityService, store,
		)
		return handler.HandleAssignFormSubmissionAPI(c)
	})
	protected.Delete("/api/outbreaks/:outbreak_id/users/:user_id", func(c *fiber.Ctx) error {
		userService := models.NewUserService(db)
		userOutbreakService := models.NewUserOutbreakService(db)
		patientRoleService := models.NewPatientManagementRoleService(db)
		outbreakService := models.NewOutbreakService(db)
		facilityService := models.NewFacilityService(db)
		handler := handlers.NewOutbreakAssignmentHandler(
			userOutbreakService, patientRoleService, userService, outbreakService, facilityService, store,
		)
		return handler.RemoveUserFromOutbreakAPI(c)
	})

	// Mpox APIs
	protected.Get("/api/mpox/patients", func(c *fiber.Ctx) error {
		return handlers.HandlerMpoxListAPI(c, db, sl, store, config)
	})
	protected.Get("/api/mpox/patients/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetMpoxAPI(c, db, sl, store, config)
	})
	protected.Post("/api/mpox/patients", func(c *fiber.Ctx) error {
		return handlers.HandlerMpoxCIFSubmitAPI(c, db, sl)
	})
	protected.Put("/api/mpox/patients/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerMpoxUpdateAPI(c, db, sl, store, config)
	})
	protected.Delete("/api/mpox/patients/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerMpoxDeleteAPI(c, db, sl, store, config)
	})

	// Mpox Admission APIs
	protected.Get("/api/mpox/patients/:id/admission", func(c *fiber.Ctx) error {
		return handlers.HandlerMpoxAdmissionFormAPI(c, db, sl, store, config)
	})
	protected.Post("/api/mpox/patients/:id/admission", func(c *fiber.Ctx) error {
		return handlers.HandlerMpoxAdmissionSubmitAPI(c, db, sl, store, config)
	})

	// Mpox Daily Follow-up APIs
	protected.Get("/api/mpox/patients/:id/daily-follow-up", func(c *fiber.Ctx) error {
		return handlers.HandlerMpoxDailyFollowUpFormAPI(c, db, sl, store, config)
	})
	protected.Post("/api/mpox/patients/:id/daily-follow-up", func(c *fiber.Ctx) error {
		return handlers.HandlerMpoxDailyFollowUpSubmitAPI(c, db, sl, store, config)
	})

	// Measles APIs
	protected.Get("/api/measles/patients", func(c *fiber.Ctx) error {
		return handlers.HandlerMeaslesListAPI(c, db, sl, store, config)
	})
	protected.Get("/api/measles/patients/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetMeaslesAPI(c, db, sl, store, config)
	})
	protected.Post("/api/measles/patients", func(c *fiber.Ctx) error {
		return handlers.HandlerMeaslesCIFAPI(c, db, store)
	})
	protected.Put("/api/measles/patients/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerMeaslesUpdateAPI(c, db, sl, store, config)
	})
	protected.Delete("/api/measles/patients/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerMeaslesDeleteAPI(c, db, sl, store, config)
	})

	// Polio APIs
	protected.Get("/api/polio/patients", func(c *fiber.Ctx) error {
		return handlers.HandlerPolioListAPI(c, db, sl, store, config)
	})
	protected.Get("/api/polio/patients/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetPolioAPI(c, db, sl, store, config)
	})
	protected.Post("/api/polio/patients", func(c *fiber.Ctx) error {
		return handlers.HandlerPolioCIFSubmitAPI(c, db, sl, store, config)
	})
	protected.Put("/api/polio/patients/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerPolioUpdateAPI(c, db, sl, store, config)
	})
	protected.Delete("/api/polio/patients/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerPolioDeleteAPI(c, db, sl, store, config)
	})

	// Patient Roles APIs
	protected.Get("/api/patient-roles", func(c *fiber.Ctx) error {
		return handlers.HandlerPatientRolesAPI(c, db, sl, store, config)
	})
	protected.Get("/api/patient-roles/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetPatientRoleAPI(c, db, sl, store, config)
	})
	protected.Post("/api/patient-roles", func(c *fiber.Ctx) error {
		return handlers.HandlerPatientRoleSubmitAPI(c, db, sl, store, config)
	})
	protected.Put("/api/patient-roles/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerPatientRoleUpdateAPI(c, db, sl, store, config)
	})
	protected.Delete("/api/patient-roles/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerPatientRoleDeleteAPI(c, db, sl, store, config)
	})

	// Alerts APIs
	protected.Get("/api/alerts", func(c *fiber.Ctx) error {
		return handlers.HandlerAlertsAPI(c, db, sl, store, config)
	})
	protected.Get("/api/alerts/6767", func(c *fiber.Ctx) error {
		return handlers.HandlerAlerts6767API(c, db, sl, store, config)
	})
	protected.Get("/api/alerts/debug", func(c *fiber.Ctx) error {
		return handlers.HandlerAlertsDebug(c, db, sl, store, config)
	})
	protected.Get("/api/alerts/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerGetAlertAPI(c, db, sl, store, config)
	})
	protected.Post("/api/alerts", func(c *fiber.Ctx) error {
		return handlers.HandlerAlertSubmitAPI(c, db, sl, store, config)
	})
	protected.Put("/api/alerts/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerAlertUpdateAPI(c, db, sl, store, config)
	})
	protected.Delete("/api/alerts/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerAlertDeleteAPI(c, db, sl, store, config)
	})

}

// RouteSurveillance defines the surveillance routes
func RouteSurveillance(v fiber.Router, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) {
	// Community Mortality Surveillance
	v.Get("/community-mortality", middleware.PermissionRequired(store, db, sl, "surveillance", "read"), func(c *fiber.Ctx) error {
		return handlers.CommunityMortalitySurveillance(c, db, store, config)
	})

	// Facility Mortality Surveillance
	v.Get("/facility-mortality", middleware.PermissionRequired(store, db, sl, "surveillance", "read"), func(c *fiber.Ctx) error {
		return handlers.FacilityMortalitySurveillance(c, db, store, config)
	})
}
