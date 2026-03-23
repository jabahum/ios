package routes

import (
	"database/sql"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"

	"case/internal/handlers"
	"case/internal/models"
	"case/internal/services"
)

// JSONAPIAuthRequired returns 401 JSON when the session is not authenticated (no HTML redirect).
// Accepts session keys "user" or "user_id" as int, int64, or float64.
func JSONAPIAuthRequired(store *session.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Authentication required"})
		}
		if sess.Get("isAuthenticated") != true {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Authentication required"})
		}
		if apiSessionUserID(sess) == 0 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user session"})
		}
		return c.Next()
	}
}

func apiSessionUserID(sess *session.Session) int {
	for _, key := range []string{"user_id", "user"} {
		v := sess.Get(key)
		if v == nil {
			continue
		}
		switch n := v.(type) {
		case int:
			if n > 0 {
				return n
			}
		case int64:
			if n > 0 {
				return int(n)
			}
		case float64:
			if n > 0 {
				return int(n)
			}
		}
	}
	return 0
}

// registerAuthenticatedJSONAPIRoutes registers JSON /api/* routes on the root app with session auth.
// This avoids relying only on app.Group("/", AuthRequired(store)), which can behave differently per deploy
// and only checks sess.Get("user") (not user_id).
func registerAuthenticatedJSONAPIRoutes(app *fiber.App, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config, smsService *services.SMSService, inventoryHandler *handlers.InventoryHandler) {
	auth := JSONAPIAuthRequired(store)

	// Lab Sample Types API routes (protected - require authentication)
	app.Get("/api/lab/swab-types", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerAPIGetLabSwabTypes(c, db, sl)
	})
	app.Get("/api/lab/urine-types", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerAPIGetLabUrineTypes(c, db, sl)
	})
	app.Get("/api/lab/blood-types", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerAPIGetLabBloodTypes(c, db, sl)
	})
	app.Get("/api/lab/blood-types/category/:category", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerAPIGetLabBloodTypesByCategory(c, db, sl)
	})

	// Lab sample selections routes (protected - require authentication)
	app.Post("/api/lab/sample-selections", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerAPISaveLabSampleSelections(c, db, sl)
	})
	app.Get("/api/lab/sample-selections/:lab_id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerAPIGetLabSampleSelections(c, db, sl)
	})

	// VHF API routes
	app.Get("/api/vhf-cases", auth, func(c *fiber.Ctx) error { return handlers.HandlerVHFList(c, db, sl, store, config) })
	app.Get("/api/vhf-cases/:id", auth, func(c *fiber.Ctx) error { return handlers.HandlerVHFView(c, db, sl, store, config) })

	// ===== COMPREHENSIVE API ENDPOINTS FOR ALL ROUTES =====
	// These endpoints mirror all existing routes for Next.js migration

	// Authentication & User Management APIs
	app.Get("/api/auth/user", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerGetCurrentUser(c, db, sl, store, config)
	})
	app.Post("/api/auth/logout", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerLoginOut(c, sl, store, config)
	})
	app.Post("/api/auth/change-password", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerChangePassword(c, db, sl, store, config)
	})

	// Dashboard & Home APIs
	app.Get("/api/dashboard/home", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerHomeAPI(c, db, sl, store, config)
	})
	app.Get("/api/dashboard/stats", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerDashboardStats(c, db, sl, store, config)
	})

	// VHF Case Management APIs
	app.Get("/api/vhf/patients", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerVHFListAPI(c, db, sl, store, config)
	})
	app.Get("/api/vhf/patients/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerVHFViewAPI(c, db, sl, store, config)
	})
	app.Post("/api/vhf/patients", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerVHFPatientSubmitAPI(c, db, sl, store, config, smsService)
	})
	app.Put("/api/vhf/patients/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerVHFUpdateAPI(c, db, sl, store, config)
	})
	app.Delete("/api/vhf/patients/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerVHFDeleteAPI(c, db, sl, store, config)
	})

	// VHF Clinical Signs APIs
	app.Get("/api/vhf/patients/:id/clinical-signs", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerVHFClinicalSignsAPI(c, db, sl, store, config)
	})
	app.Post("/api/vhf/patients/:id/clinical-signs", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerVHFClinicalSignsSubmitAPI(c, db, sl, store, config)
	})

	// VHF Hospitalization APIs
	app.Get("/api/vhf/patients/:id/hospitalization", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerVHFHospitalizationAPI(c, db, sl, store, config)
	})
	app.Post("/api/vhf/patients/:id/hospitalization", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerVHFHospitalizationSubmitAPI(c, db, sl, store, config)
	})

	// VHF Risk Factors APIs
	app.Get("/api/vhf/patients/:id/risk-factors", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerVHFRiskFactorsAPI(c, db, sl, store, config)
	})
	app.Post("/api/vhf/patients/:id/risk-factors", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerVHFRiskFactorsSubmitAPI(c, db, sl, store, config)
	})

	// VHF Laboratory APIs
	app.Get("/api/vhf/patients/:id/laboratory", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerVHFLaboratoryAPI(c, db, sl, store, config)
	})
	app.Post("/api/vhf/patients/:id/laboratory", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerVHFLaboratorySubmitAPI(c, db, sl, store, config, smsService)
	})

	// VHF Investigator APIs
	app.Get("/api/vhf/patients/:id/investigator", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerVHFInvestigatorAPI(c, db, sl, store, config)
	})
	app.Post("/api/vhf/patients/:id/investigator", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerVHFInvestigatorSubmitAPI(c, db, sl, store, config)
	})

	// VHF Lab Form APIs
	app.Get("/api/vhf/lab/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerVHFLabFormAPI(c, db, sl, store, config)
	})
	app.Post("/api/vhf/lab/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerVHFLabSaveAPI(c, db, sl, store, config)
	})

	// Employee JSON APIs: registered on app with PermissionRequired (see /api/users block).

	// User Management APIs
	app.Get("/api/users/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerGetUserAPI(c, db, sl, store, config)
	})
	app.Post("/api/users", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerUserSubmitAPI(c, db, sl, store, config)
	})
	app.Put("/api/users/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerUserUpdateAPI(c, db, sl, store, config)
	})
	app.Delete("/api/users/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerUserDeleteAPI(c, db, sl, store, config)
	})

	// Facility Management APIs
	app.Get("/api/facilities/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerGetFacilityAPI(c, db, sl, store, config)
	})
	app.Post("/api/facilities", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerFacilitySubmitAPI(c, db, sl, store, config)
	})
	app.Put("/api/facilities/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerFacilityUpdateAPI(c, db, sl, store, config)
	})
	app.Delete("/api/facilities/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerFacilityDeleteAPI(c, db, sl, store, config)
	})

	// Outbreak Management APIs
	app.Get("/api/outbreaks/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerGetOutbreakAPI(c, db, sl, store, config)
	})
	app.Post("/api/outbreaks", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerOutbreakSubmitAPI(c, db, sl, store, config)
	})
	app.Put("/api/outbreaks/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerOutbreakUpdateAPI(c, db, sl, store, config)
	})
	app.Delete("/api/outbreaks/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerOutbreakDeleteAPI(c, db, sl, store, config)
	})
	app.Post("/api/outbreaks/:id/close", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerOutbreakCloseAPI(c, db, sl, store, config)
	})
	app.Post("/api/outbreaks/:id/select", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerOutbreakSelectAPI(c, db, sl, store, config)
	})

	// Case Management APIs
	app.Get("/api/cases", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerCasesListAPI(c, db, sl, store, config)
	})
	app.Get("/api/cases/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerGetCaseAPI(c, db, sl, store, config)
	})
	app.Put("/api/cases/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerCaseUpdateAPI(c, db, sl, store, config)
	})
	app.Delete("/api/cases/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerCaseDeleteAPI(c, db, sl, store, config)
	})

	// Disease-specific CIF APIs (VHF /api/vhf/cif* and /api/cif* are on app — see top of SetRoute)
	// Measles
	app.Get("/api/measles/cif/:id", auth, func(c *fiber.Ctx) error { return handlers.HandlerMeaslesCIFByID(c, db, sl, store, config) })
	app.Get("/api/measles/cif", auth, func(c *fiber.Ctx) error { return handlers.HandlerMeaslesCIFByCaseCode(c, db, sl, store, config) })
	// Polio
	app.Get("/api/polio/cif/:id", auth, func(c *fiber.Ctx) error { return handlers.HandlerPolioCIFByID(c, db, sl, store, config) })
	app.Get("/api/polio/cif", auth, func(c *fiber.Ctx) error { return handlers.HandlerPolioCIFByCaseCode(c, db, sl, store, config) })
	// Mpox
	app.Get("/api/mpox/cif/:id", auth, func(c *fiber.Ctx) error { return handlers.HandlerMpoxCIFByID(c, db, sl, store, config) })
	app.Get("/api/mpox/cif", auth, func(c *fiber.Ctx) error { return handlers.HandlerMpoxCIFByCaseCode(c, db, sl, store, config) })

	// Case Encounter APIs
	app.Get("/api/cases/:id/encounters", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerCaseEncounterListAPI(c, db, sl, store, config)
	})
	app.Get("/api/cases/:id/encounters/:encounter_id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerGetCaseEncounterAPI(c, db, sl, store, config)
	})
	app.Post("/api/cases/:id/encounters", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerCaseEncounterSubmitAPI(c, db, sl, store, config)
	})
	app.Put("/api/cases/:id/encounters/:encounter_id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerCaseEncounterUpdateAPI(c, db, sl, store, config)
	})
	app.Delete("/api/cases/:id/encounters/:encounter_id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerCaseEncounterDeleteAPI(c, db, sl, store, config)
	})

	// Discharge Management APIs
	app.Get("/api/discharges", auth, func(c *fiber.Ctx) error {
		return handlers.GetDischargeAPI(c, db, sl, store, config)
	})
	app.Get("/api/discharges/:id", auth, func(c *fiber.Ctx) error {
		return handlers.GetDischargeByIdAPI(c, db, sl, store, config)
	})
	app.Post("/api/discharges", auth, func(c *fiber.Ctx) error {
		return handlers.DischargeAPI(c, db, sl, store, config)
	})
	app.Get("/api/discharges/:id/certificate", auth, func(c *fiber.Ctx) error {
		return handlers.CertificateAPI(c, db, sl, store, config)
	})

	// Laboratory Management APIs
	app.Get("/api/laboratory", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerLabListAPI(c, db, sl, store, config)
	})
	app.Get("/api/laboratory/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerGetLabAPI(c, db, sl, store, config)
	})
	app.Post("/api/laboratory", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerLabSubmitAPI(c, db, sl, store, config)
	})
	app.Put("/api/laboratory/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerLabUpdateAPI(c, db, sl, store, config)
	})
	app.Delete("/api/laboratory/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerLabDeleteAPI(c, db, sl, store, config)
	})

	// Symptoms Management APIs
	app.Get("/api/symptoms", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerSymptomsListAPI(c, db, sl, store, config)
	})
	app.Get("/api/symptoms/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerGetSymptomsAPI(c, db, sl, store, config)
	})
	app.Post("/api/symptoms", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerSymptomsSubmitAPI(c, db, sl, store, config)
	})
	app.Put("/api/symptoms/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerSymptomsUpdateAPI(c, db, sl, store, config)
	})
	app.Delete("/api/symptoms/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerSymptomsDeleteAPI(c, db, sl, store, config)
	})

	// Morbidity Management APIs
	app.Get("/api/morbidity", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerMorbidityListAPI(c, db, sl, store, config)
	})
	app.Get("/api/morbidity/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerGetMorbidityAPI(c, db, sl, store, config)
	})
	app.Post("/api/morbidity", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerMorbiditySubmitAPI(c, db, sl, store, config)
	})
	app.Put("/api/morbidity/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerMorbidityUpdateAPI(c, db, sl, store, config)
	})
	app.Delete("/api/morbidity/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerMorbidityDeleteAPI(c, db, sl, store, config)
	})

	// Rush Management APIs
	app.Get("/api/rush", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerRushListAPI(c, db, sl, store, config)
	})
	app.Get("/api/rush/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerGetRushAPI(c, db, sl, store, config)
	})
	app.Post("/api/rush", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerRushSubmitAPI(c, db, sl, store, config)
	})
	app.Put("/api/rush/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerRushUpdateAPI(c, db, sl, store, config)
	})
	app.Delete("/api/rush/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerRushDeleteAPI(c, db, sl, store, config)
	})

	// Inventory Management APIs
	app.Get("/api/inventory/dashboard", auth, func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerInventoryDashboardAPI(c)
	})
	app.Get("/api/inventory/items", auth, func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerInventoryAPIItems(c)
	})
	app.Get("/api/inventory/items/:id", auth, func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerGetInventoryItemAPI(c)
	})
	app.Post("/api/inventory/items", auth, func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerInventoryItemSaveAPI(c)
	})
	app.Put("/api/inventory/items/:id", auth, func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerInventoryItemUpdateAPI(c)
	})
	app.Delete("/api/inventory/items/:id", auth, func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerInventoryItemDeleteAPI(c)
	})

	// Inventory Stock APIs
	app.Get("/api/inventory/stock", auth, func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerInventoryAPIStockLevels(c)
	})
	app.Get("/api/inventory/stock/:id", auth, func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerGetInventoryStockAPI(c)
	})
	app.Post("/api/inventory/stock", auth, func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerInventoryStockSaveAPI(c)
	})
	app.Put("/api/inventory/stock/:id", auth, func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerInventoryStockUpdateAPI(c)
	})

	// Inventory Purchase Orders APIs
	app.Get("/api/inventory/purchase-orders", auth, func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerInventoryPurchaseOrdersAPI(c)
	})
	app.Get("/api/inventory/purchase-orders/:id", auth, func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerGetPurchaseOrderAPI(c)
	})
	app.Post("/api/inventory/purchase-orders", auth, func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerInventoryPurchaseOrderSaveAPI(c)
	})
	app.Put("/api/inventory/purchase-orders/:id", auth, func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerPurchaseOrderUpdateAPI(c)
	})

	// Inventory Requisitions APIs
	app.Get("/api/inventory/requisitions", auth, func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerInventoryRequisitionsAPI(c)
	})
	app.Get("/api/inventory/requisitions/:id", auth, func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerGetRequisitionAPI(c)
	})
	app.Post("/api/inventory/requisitions", auth, func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerInventoryRequisitionSaveAPI(c)
	})
	app.Put("/api/inventory/requisitions/:id", auth, func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerRequisitionUpdateAPI(c)
	})

	// Inventory Donations APIs
	app.Get("/api/inventory/donations", auth, func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerDonationsListAPI(c)
	})
	app.Get("/api/inventory/donations/:id", auth, func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerDonationViewAPI(c)
	})
	app.Post("/api/inventory/donations", auth, func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerDonationSaveAPI(c)
	})
	app.Put("/api/inventory/donations/:id", auth, func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerDonationUpdateAPI(c)
	})
	app.Delete("/api/inventory/donations/:id", auth, func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerDonationDeleteAPI(c)
	})

	// Inventory Donors APIs
	app.Get("/api/inventory/donors", auth, func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerDonorsListAPI(c)
	})
	app.Get("/api/inventory/donors/:id", auth, func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerGetDonorAPI(c)
	})
	app.Post("/api/inventory/donors", auth, func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerDonorSaveAPI(c)
	})
	app.Put("/api/inventory/donors/:id", auth, func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerDonorUpdateAPI(c)
	})
	app.Delete("/api/inventory/donors/:id", auth, func(c *fiber.Ctx) error {
		return inventoryHandler.HandlerDonorDeleteAPI(c)
	})

	// Surveillance APIs
	app.Get("/api/surveillance/community-mortality", auth, func(c *fiber.Ctx) error {
		return handlers.CommunityMortalitySurveillanceAPI(c, db, store, config)
	})
	app.Get("/api/surveillance/facility-mortality", auth, func(c *fiber.Ctx) error {
		return handlers.FacilityMortalitySurveillanceAPI(c, db, store, config)
	})

	// Outbreak Assignment APIs
	app.Get("/api/outbreaks/assignments", auth, func(c *fiber.Ctx) error {
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
	app.Post("/api/outbreaks/assign", auth, func(c *fiber.Ctx) error {
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
	app.Delete("/api/outbreaks/:outbreak_id/users/:user_id", auth, func(c *fiber.Ctx) error {
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
	app.Get("/api/mpox/patients", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerMpoxListAPI(c, db, sl, store, config)
	})
	app.Get("/api/mpox/patients/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerGetMpoxAPI(c, db, sl, store, config)
	})
	app.Post("/api/mpox/patients", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerMpoxCIFSubmitAPI(c, db, sl)
	})
	app.Put("/api/mpox/patients/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerMpoxUpdateAPI(c, db, sl, store, config)
	})
	app.Delete("/api/mpox/patients/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerMpoxDeleteAPI(c, db, sl, store, config)
	})

	// Mpox Admission APIs
	app.Get("/api/mpox/patients/:id/admission", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerMpoxAdmissionFormAPI(c, db, sl, store, config)
	})
	app.Post("/api/mpox/patients/:id/admission", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerMpoxAdmissionSubmitAPI(c, db, sl, store, config)
	})

	// Mpox Daily Follow-up APIs
	app.Get("/api/mpox/patients/:id/daily-follow-up", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerMpoxDailyFollowUpFormAPI(c, db, sl, store, config)
	})
	app.Post("/api/mpox/patients/:id/daily-follow-up", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerMpoxDailyFollowUpSubmitAPI(c, db, sl, store, config)
	})

	// Measles APIs
	app.Get("/api/measles/patients", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerMeaslesListAPI(c, db, sl, store, config)
	})
	app.Get("/api/measles/patients/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerGetMeaslesAPI(c, db, sl, store, config)
	})
	app.Post("/api/measles/patients", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerMeaslesCIFAPI(c, db, store)
	})
	app.Put("/api/measles/patients/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerMeaslesUpdateAPI(c, db, sl, store, config)
	})
	app.Delete("/api/measles/patients/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerMeaslesDeleteAPI(c, db, sl, store, config)
	})

	// Polio APIs
	app.Get("/api/polio/patients", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerPolioListAPI(c, db, sl, store, config)
	})
	app.Get("/api/polio/patients/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerGetPolioAPI(c, db, sl, store, config)
	})
	app.Post("/api/polio/patients", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerPolioCIFSubmitAPI(c, db, sl, store, config)
	})
	app.Put("/api/polio/patients/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerPolioUpdateAPI(c, db, sl, store, config)
	})
	app.Delete("/api/polio/patients/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerPolioDeleteAPI(c, db, sl, store, config)
	})

	// Patient Roles APIs
	app.Get("/api/patient-roles", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerPatientRolesAPI(c, db, sl, store, config)
	})
	app.Get("/api/patient-roles/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerGetPatientRoleAPI(c, db, sl, store, config)
	})
	app.Post("/api/patient-roles", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerPatientRoleSubmitAPI(c, db, sl, store, config)
	})
	app.Put("/api/patient-roles/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerPatientRoleUpdateAPI(c, db, sl, store, config)
	})
	app.Delete("/api/patient-roles/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerPatientRoleDeleteAPI(c, db, sl, store, config)
	})

	// Alerts APIs
	app.Get("/api/alerts", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerAlertsAPI(c, db, sl, store, config)
	})
	app.Get("/api/alerts/6767", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerAlerts6767API(c, db, sl, store, config)
	})
	app.Get("/api/alerts/debug", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerAlertsDebug(c, db, sl, store, config)
	})
	app.Get("/api/alerts/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerGetAlertAPI(c, db, sl, store, config)
	})
	app.Post("/api/alerts", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerAlertSubmitAPI(c, db, sl, store, config)
	})
	app.Put("/api/alerts/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerAlertUpdateAPI(c, db, sl, store, config)
	})
	app.Delete("/api/alerts/:id", auth, func(c *fiber.Ctx) error {
		return handlers.HandlerAlertDeleteAPI(c, db, sl, store, config)
	})

}
