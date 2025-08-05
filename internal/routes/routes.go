package routes

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	_ "github.com/lib/pq"

	"case/internal/handlers"
	"case/internal/middleware"
	"case/internal/models"
	"case/internal/reports"
	"case/internal/services"
)

func SetRoute(app *fiber.App, db *sql.DB, store *session.Store, sl *slog.Logger, config handlers.Config, smsService *services.SMSService) {

	// Public routes
	app.Get("/", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "landing")
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
		return handlers.HandlerGetUsers(c, db, sl)
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
		return handlers.HandlerGetUsers(c, db, sl)
	})
	app.Get("/api/users/:id/permissions", middleware.PermissionRequired(store, db, sl, "users", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerGetUserPermissions(c, db, sl)
	})
	app.Post("/api/users/roles", middleware.PermissionRequired(store, db, sl, "users", "update"), func(c *fiber.Ctx) error {
		return handlers.HandlerAssignUserRole(c, db, sl, store, config)
	})

	// Protected routes
	appGroup := app.Group("/")
	appGroup.Use(AuthRequired(store))
	{
		appGroup.Get("/home", func(c *fiber.Ctx) error { return handlers.HandlerHome(c, db, sl, store, config) })
		appGroup.Get("/alerts", middleware.PermissionRequired(store, db, sl, "alerts", "read"), func(c *fiber.Ctx) error {
			return handlers.HandlerAlerts(c, db, sl, store, config)
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

		// Additional routes
		RouteFacilities(hfs, db, sl, store, config)
		RouteUsers(usr, db, sl, store, config)
		RouteCases(cse, db, sl, store, config)
		RouteMorbidity(mob, db, sl, store, config)
		RouteSymptoms(sym, db, sl, store, config)
		RouteRush(rus, db, sl, store, config)
		RouteLab(lab, db, sl, store, config)

		RouteEmployees(emp, db, sl, store, config)
		RouteDischarge(dis, db, sl, store, config)

		RouteReports(rpt, db, sl, store, config)

		RouteAPIEncounter(enk, db, sl, store, config)
		RouteAPIStatus(sta, db, sl, store, config)

		// Add missing routes for home.html navigation
		appGroup.Get("/vhf", func(c *fiber.Ctx) error { return handlers.HandlerVHFList(c, db, sl, store, config) })
		appGroup.Get("/vhf/new", func(c *fiber.Ctx) error {
			return handlers.GenerateHTML(c, db, handlers.NewTemplateData(c, store), "vhf_cif")
		})
		appGroup.Get("/cases/new", func(c *fiber.Ctx) error { return handlers.HandlerCasesForm(c, db, sl, store, config) })
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
		}

		path := c.Path()
		for _, route := range publicRoutes {
			if path == route {
				return c.Next()
			}
		}

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
}

func RouteCases(v fiber.Router, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) {
	// Add RBAC permission checks for case management
	v.Get("/new/:i", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
		return handlers.HandlerCasesForm(c, db, sl, store, config)
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

func RouteCaseDischarge(v fiber.Router, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) {
	// Add RBAC permission checks for case discharge
	v.Get("/view/:i/:j", middleware.PermissionRequired(store, db, sl, "vhf_patients", "read"), func(c *fiber.Ctx) error {
		return handlers.HandlerCasesForm(c, db, sl, store, config)
	})
	v.Get("/new/:i/:j", middleware.PermissionRequired(store, db, sl, "vhf_patients", "create"), func(c *fiber.Ctx) error {
		return handlers.HandlerCasesForm(c, db, sl, store, config)
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
		return reports.ReportView(c, db, sl, store, config)
	}) //+
	v.Get("/", middleware.PermissionRequired(store, db, sl, "reports", "read"), func(c *fiber.Ctx) error {
		return reports.ReportHome(c, db, sl, store, config)
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
}
