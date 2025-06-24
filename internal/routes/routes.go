package routes

import (
	"database/sql"
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	_ "github.com/lib/pq"

	"case/internal/handlers"
	"case/internal/reports"
	"case/internal/services"
)

func SetRoute(app *fiber.App, db *sql.DB, store *session.Store, sl *slog.Logger, config handlers.Config, smsService *services.SMSService) {
	// Public routes
	app.Get("/", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, nil, "landing")
	})
	app.Get("/login", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, nil, "login")
	})
	app.Post("/login", func(c *fiber.Ctx) error {
		return handlers.HandlerLoginSubmit(c, db, sl, store, config)
	})
	app.Get("/logout", func(c *fiber.Ctx) error {
		return handlers.HandlerLoginOut(c, sl, store, config)
	})
	app.Get("/vhf-cif", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, nil, "vhf_cif")
	})
	app.Post("/vhf-cif/save", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFPatientSubmit(c, db, sl, store, config, smsService)
	})
	app.Get("/vhf-cif/clinical-signs/:id", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, fiber.Map{"PatientID": c.Params("id")}, "vhf_clinical_signs")
	})
	app.Post("/vhf-cif/clinical-signs/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFClinicalSignsSubmit(c, db, sl, store, config)
	})
	app.Get("/vhf-cif/hospitalization/:id", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, fiber.Map{"PatientID": c.Params("id")}, "vhf_hospitalization")
	})
	app.Post("/vhf-cif/hospitalization/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFHospitalizationSubmit(c, db, sl, store, config)
	})
	app.Get("/vhf-cif/risk-factors/:id", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, fiber.Map{"PatientID": c.Params("id")}, "vhf_risk_factors")
	})
	app.Post("/vhf-cif/risk-factors/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFRiskFactorsSubmit(c, db, sl, store, config)
	})
	app.Get("/vhf-cif/laboratory/:id", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, fiber.Map{"PatientID": c.Params("id")}, "vhf_laboratory")
	})
	app.Post("/vhf-cif/laboratory/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFLaboratorySubmit(c, db, sl, store, config)
	})
	app.Get("/vhf-cif/investigator/:id", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, fiber.Map{"PatientID": c.Params("id")}, "vhf_investigator")
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
	app.Get("/vhf-list", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFList(c, db, sl, store, config)
	})

	// Add outbreak routes
	app.Get("/outbreaks", func(c *fiber.Ctx) error { return handlers.HandlerOutbreakList(c, db, sl, store, config) })
	app.Get("/outbreaks/new/:i", func(c *fiber.Ctx) error { return handlers.HandlerOutbreakForm(c, db, sl, store, config) })
	app.Get("/outbreaks/edit/:i", func(c *fiber.Ctx) error { return handlers.HandlerOutbreakForm(c, db, sl, store, config) })
	app.Post("/outbreaks/save", func(c *fiber.Ctx) error { return handlers.HandlerOutbreakSubmit(c, db, sl, store, config) })
	app.Get("/outbreaks/close/:i", func(c *fiber.Ctx) error { return handlers.HandlerOutbreakClose(c, db, sl, store, config) })
	app.Post("/outbreaks/select/:i", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("i"))
		if err != nil {
			return c.Status(400).SendString("Invalid outbreak ID")
		}
		if err := handlers.SetSelectedOutbreak(c, store, id); err != nil {
			return c.Status(500).SendString("Failed to select outbreak")
		}
		return c.SendStatus(200)
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

	// Protected routes
	appGroup := app.Group("/")
	appGroup.Use(AuthRequired(store))
	{
		appGroup.Get("/home", func(c *fiber.Ctx) error { return handlers.HandlerHome(c, db, sl, store, config) })
		appGroup.Get("/vhf-cif", func(c *fiber.Ctx) error { return handlers.GenerateHTML(c, db, nil, "vhf_cif") })
		appGroup.Post("/vhf-cif", func(c *fiber.Ctx) error {
			return handlers.HandlerVHFPatientSubmit(c, db, sl, store, config, smsService)
		})
		appGroup.Get("/vhf-cif/clinical-signs/:id", func(c *fiber.Ctx) error { return handlers.GenerateHTML(c, db, nil, "vhf_clinical_signs") })
		appGroup.Post("/vhf-cif/clinical-signs/:id", func(c *fiber.Ctx) error { return handlers.HandlerVHFClinicalSignsSubmit(c, db, sl, store, config) })
		appGroup.Get("/vhf-cif/hospitalization/:id", func(c *fiber.Ctx) error { return handlers.GenerateHTML(c, db, nil, "vhf_hospitalization") })
		appGroup.Post("/vhf-cif/hospitalization/:id", func(c *fiber.Ctx) error { return handlers.HandlerVHFHospitalizationSubmit(c, db, sl, store, config) })
		appGroup.Get("/vhf-cif/risk-factors/:id", func(c *fiber.Ctx) error { return handlers.GenerateHTML(c, db, nil, "vhf_risk_factors") })
		appGroup.Post("/vhf-cif/risk-factors/:id", func(c *fiber.Ctx) error { return handlers.HandlerVHFRiskFactorsSubmit(c, db, sl, store, config) })
		appGroup.Get("/vhf-cif/laboratory/:id", func(c *fiber.Ctx) error { return handlers.GenerateHTML(c, db, nil, "vhf_laboratory") })
		appGroup.Post("/vhf-cif/laboratory/:id", func(c *fiber.Ctx) error { return handlers.HandlerVHFLaboratorySubmit(c, db, sl, store, config) })
		appGroup.Get("/vhf-cif/investigator/:id", func(c *fiber.Ctx) error { return handlers.GenerateHTML(c, db, nil, "vhf_investigator") })
		appGroup.Post("/vhf-cif/investigator/:id", func(c *fiber.Ctx) error { return handlers.HandlerVHFInvestigatorSubmit(c, db, sl, store, config) })
		appGroup.Get("/vhf-cif/success", func(c *fiber.Ctx) error { return handlers.HandlerVHFSuccess(c, db, sl, store, config) })
		appGroup.Get("/vhf/list", func(c *fiber.Ctx) error { return handlers.HandlerVHFList(c, db, sl, store, config) })
		appGroup.Get("/vhf/view/:id", func(c *fiber.Ctx) error { return handlers.HandlerVHFView(c, db, sl, store, config) })
		appGroup.Get("/vhf-lab/:id", func(c *fiber.Ctx) error { return handlers.HandlerVHFLabForm(c, db) })
		appGroup.Post("/vhf-lab/:id", func(c *fiber.Ctx) error { return handlers.HandlerVHFLabSave(c, db, sl, store, config) })

		// Add more protected routes...

		api := app.Group("/api") // Group for all API routes

		enk := api.Group("/encounter")
		sym := api.Group("/sym")
		mob := api.Group("/mob")
		rus := api.Group("/rush")
		lab := api.Group("/lab")
		sta := api.Group("/status")

		emp := app.Group("/employees") // Employees
		usr := app.Group("/users")     // users
		hfs := app.Group("/secure")    // Health facilities
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

		// New routes for mpox daily follow-up
		cse.Get("/encounters/mpox-admission/new/:i", func(c *fiber.Ctx) error { return handlers.HandlerMpoxAdmissionForm(c, db, sl, store, config) })
		cse.Post("/encounters/mpox-admission/save", func(c *fiber.Ctx) error { return handlers.HandlerMpoxAdmissionSubmit(c, db, sl, store, config) })

		// VHF API routes
		lab.Get("/api/vhf-cases", func(c *fiber.Ctx) error { return handlers.HandlerVHFList(c, db, sl, store, config) })
		lab.Get("/api/vhf-cases/:id", func(c *fiber.Ctx) error { return handlers.HandlerVHFView(c, db, sl, store, config) })

		// Protected API routes
		protected := api.Group("/vhf")
		protected.Get("/cases", func(c *fiber.Ctx) error { return handlers.HandlerVHFList(c, db, sl, store, config) })
		protected.Get("/cases/:id", func(c *fiber.Ctx) error { return handlers.HandlerVHFView(c, db, sl, store, config) })
	}

	// MPOX CIF routes (public)
	app.Get("/mpox-cif", func(c *fiber.Ctx) error {
		// Generate a unique case_id (e.g., using timestamp or UUID)
		caseID := "MPOX-" + strconv.FormatInt(time.Now().UnixNano(), 10)
		return handlers.GenerateHTML(c, db, fiber.Map{"case_id": caseID}, "mpox_cif")
	})
	app.Post("/mpox-cif/save", func(c *fiber.Ctx) error {
		return handlers.HandlerMpoxCIFSubmit(c, db, sl)
	})
	app.Get("/mpox-cif/success", func(c *fiber.Ctx) error {
		return handlers.HandlerMpoxCIFSuccess(c, db, sl)
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
	v.Get("/", func(c *fiber.Ctx) error { return handlers.HandlerAPIGetEncounter(c, db, sl, store, config) })
}

func RouteAPIStatus(v fiber.Router, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) {
	v.Get("/list", func(c *fiber.Ctx) error { return handlers.HandlerAPIGetStatuses(c, db, sl, store, config) })
	v.Post("/save", func(c *fiber.Ctx) error { return handlers.HandlerAPIPostStatus(c, db, sl, store, config) })
}

func RouteDischarge(v fiber.Router, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) {
	v.Get("/list", func(c *fiber.Ctx) error { return handlers.GetDischarge(c, db, sl, store, config) })
	v.Get("/certificate", func(c *fiber.Ctx) error { return handlers.Certificate(c, db, sl, store, config) })
	v.Post("/save", func(c *fiber.Ctx) error { return handlers.Discharge(c, db, sl, store, config) })
}

func RouteHome(app *fiber.App, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config, smsService *services.SMSService) {
	// Landing page
	app.Get("/", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, nil, "landing")
	})

	// Login routes
	app.Get("/login", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, nil, "login")
	})
	app.Post("/login", func(c *fiber.Ctx) error {
		return handlers.HandlerLoginSubmit(c, db, sl, store, config)
	})
	app.Get("/logout", func(c *fiber.Ctx) error {
		return handlers.HandlerLoginOut(c, sl, store, config)
	})

	// VHF CIF routes
	app.Get("/vhf-cif", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, nil, "vhf_cif")
	})
	app.Post("/vhf-cif/save", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFPatientSubmit(c, db, sl, store, config, smsService)
	})
	app.Get("/vhf-cif/clinical-signs/:id", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, fiber.Map{"PatientID": c.Params("id")}, "vhf_clinical_signs")
	})
	app.Post("/vhf-cif/clinical-signs/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFClinicalSignsSubmit(c, db, sl, store, config)
	})
	app.Get("/vhf-cif/hospitalization/:id", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, fiber.Map{"PatientID": c.Params("id")}, "vhf_hospitalization")
	})
	app.Post("/vhf-cif/hospitalization/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFHospitalizationSubmit(c, db, sl, store, config)
	})
	app.Get("/vhf-cif/risk-factors/:id", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, fiber.Map{"PatientID": c.Params("id")}, "vhf_risk_factors")
	})
	app.Post("/vhf-cif/risk-factors/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFRiskFactorsSubmit(c, db, sl, store, config)
	})
	app.Get("/vhf-cif/laboratory/:id", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, fiber.Map{"PatientID": c.Params("id")}, "vhf_laboratory")
	})
	app.Post("/vhf-cif/laboratory/:id", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFLaboratorySubmit(c, db, sl, store, config)
	})
	app.Get("/vhf-cif/investigator/:id", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, fiber.Map{"PatientID": c.Params("id")}, "vhf_investigator")
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
	v.Get("/new/:i", func(c *fiber.Ctx) error { return handlers.HandlerFacilityForm(c, db, sl, store, config) })
	v.Post("/save", func(c *fiber.Ctx) error { return handlers.HandlerFacilitySubmit(c, db, sl, store, config) })
	v.Post("/filter", func(c *fiber.Ctx) error { return handlers.HandlerFacilityList(c, db, sl, store, config) })
	v.Get("/list", func(c *fiber.Ctx) error { return handlers.HandlerFacilityList(c, db, sl, store, config) })
	v.Get("/", func(c *fiber.Ctx) error { return handlers.HandlerFacilityList(c, db, sl, store, config) })
}

func RouteUsers(v fiber.Router, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) {
	v.Get("/new/:i", func(c *fiber.Ctx) error { return handlers.HandlerUserForm(c, db, sl, store, config) })
	v.Post("/save", func(c *fiber.Ctx) error { return handlers.HandlerUserSubmit(c, db, sl, store, config) })
	v.Post("/filter", func(c *fiber.Ctx) error { return handlers.HandlerUserList(c, db, sl, store, config) })
	v.Get("/list", func(c *fiber.Ctx) error { return handlers.HandlerUserList(c, db, sl, store, config) })
	v.Get("/", func(c *fiber.Ctx) error { return handlers.HandlerUserList(c, db, sl, store, config) })
}

func RouteEmployees(v fiber.Router, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) {
	v.Get("/new/:i", func(c *fiber.Ctx) error { return handlers.HandlerEmployeeForm(c, db, sl, store, config) })
	v.Post("/save", func(c *fiber.Ctx) error { return handlers.HandlerEmployeeSubmit(c, db, sl, store, config) })
	v.Post("/filter", func(c *fiber.Ctx) error { return handlers.HandlerEmployeeList(c, db, sl, store, config) })
	v.Get("/list", func(c *fiber.Ctx) error { return handlers.HandlerEmployeeList(c, db, sl, store, config) })
	v.Get("/", func(c *fiber.Ctx) error { return handlers.HandlerEmployeeList(c, db, sl, store, config) })
}

func RouteCases(v fiber.Router, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) {
	v.Get("/new/:i", func(c *fiber.Ctx) error { return handlers.HandlerCasesForm(c, db, sl, store, config) })
	v.Post("/save", func(c *fiber.Ctx) error { return handlers.HandlerCasesSubmit(c, db, sl, store, config) })
	v.Post("/filter", func(c *fiber.Ctx) error { return handlers.HandlerCasesList(c, db, sl, store, config) })
	v.Get("/list", func(c *fiber.Ctx) error { return handlers.HandlerCasesList(c, db, sl, store, config) })
	v.Get("/", func(c *fiber.Ctx) error { return handlers.HandlerCasesList(c, db, sl, store, config) })

	v.Get("/encounters/list/:i", func(c *fiber.Ctx) error { return handlers.HandlerCaseEncounterForm(c, db, sl, store, config) })
	v.Get("/encounters/new/:i", func(c *fiber.Ctx) error { return handlers.HandlerCaseEncounterForm(c, db, sl, store, config) })
	v.Get("/encounters/new/:i/:j", func(c *fiber.Ctx) error { return handlers.HandlerCaseEncounterForm(c, db, sl, store, config) })
	v.Post("/encounters/save", func(c *fiber.Ctx) error { return handlers.HandlerCaseEncounterSubmit(c, db, sl, store, config) })

	v.Get("/encounters/mpox-admission/new/:i", func(c *fiber.Ctx) error { return handlers.HandlerMpoxAdmissionForm(c, db, sl, store, config) })
	v.Post("/encounters//save", func(c *fiber.Ctx) error { return handlers.HandlerMpoxAdmissionSubmit(c, db, sl, store, config) })
}

func RouteCaseDischarge(v fiber.Router, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) {
	v.Get("/view/:i/:j", func(c *fiber.Ctx) error { return handlers.HandlerCasesForm(c, db, sl, store, config) })
	v.Get("/new/:i/:j", func(c *fiber.Ctx) error { return handlers.HandlerCasesForm(c, db, sl, store, config) })
	v.Post("/save/:i/:j", func(c *fiber.Ctx) error { return handlers.HandlerCasesSubmit(c, db, sl, store, config) })
}

func RouteSymptoms(v fiber.Router, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) {
	v.Get("/new/:i", func(c *fiber.Ctx) error { return handlers.HandlerSymptomsForm(c, db, sl, store, config) })
	v.Post("/save", func(c *fiber.Ctx) error { return handlers.HandlerSymptomsSubmit(c, db, sl, store, config) })
	v.Post("/filter", func(c *fiber.Ctx) error { return handlers.HandlerSymptomsList(c, db, sl, store, config) })
	v.Get("/list", func(c *fiber.Ctx) error { return handlers.HandlerSymptomsList(c, db, sl, store, config) })
	v.Get("/", func(c *fiber.Ctx) error { return handlers.HandlerSymptomsList(c, db, sl, store, config) })
}

func RouteMorbidity(v fiber.Router, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) {
	v.Get("/new/:i", func(c *fiber.Ctx) error { return handlers.HandlerMorbidityForm(c, db, sl, store, config) })
	v.Post("/save", func(c *fiber.Ctx) error { return handlers.HandlerMorbiditySubmit(c, db, sl, store, config) })
	v.Post("/filter", func(c *fiber.Ctx) error { return handlers.HandlerMorbidityList(c, db, sl, store, config) })
	v.Get("/list", func(c *fiber.Ctx) error { return handlers.HandlerMorbidityList(c, db, sl, store, config) })
	v.Get("/", func(c *fiber.Ctx) error { return handlers.HandlerMorbidityList(c, db, sl, store, config) })
}

func RouteRush(v fiber.Router, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) {
	v.Get("/new/:i", func(c *fiber.Ctx) error { return handlers.HandlerRushForm(c, db, sl, store, config) })
	v.Post("/save", func(c *fiber.Ctx) error { return handlers.HandlerRushSubmit(c, db, sl, store, config) })
	v.Post("/filter", func(c *fiber.Ctx) error { return handlers.HandlerRushList(c, db, sl, store, config) })
	v.Get("/list", func(c *fiber.Ctx) error { return handlers.HandlerRushList(c, db, sl, store, config) })
	v.Get("/", func(c *fiber.Ctx) error { return handlers.HandlerRushList(c, db, sl, store, config) })
}

func RouteLab(v fiber.Router, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) {
	v.Get("/new/:i", func(c *fiber.Ctx) error { return handlers.HandlerLabForm(c, db, sl, store, config) })
	v.Post("/save", func(c *fiber.Ctx) error { return handlers.HandlerLabSubmit(c, db, sl, store, config) })
	v.Post("/filter", func(c *fiber.Ctx) error { return handlers.HandlerLabList(c, db, sl, store, config) })
	v.Get("/list", func(c *fiber.Ctx) error { return handlers.HandlerLabList(c, db, sl, store, config) })
	v.Get("/", func(c *fiber.Ctx) error { return handlers.HandlerLabList(c, db, sl, store, config) })
}

func RouteReports(v fiber.Router, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) { //+
	//+
	v.Get("/view", func(c *fiber.Ctx) error { return reports.ReportView(c, db, sl, store, config) }) //+
	v.Get("/", func(c *fiber.Ctx) error { return reports.ReportHome(c, db, sl, store, config) })
}

// Add this new function for outbreak routes
func RouteOutbreaks(app *fiber.App, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) {
	// Public routes
	app.Get("/outbreaks", func(c *fiber.Ctx) error { return handlers.HandlerOutbreakList(c, db, sl, store, config) })
	app.Get("/outbreaks/new/:i", func(c *fiber.Ctx) error { return handlers.HandlerOutbreakForm(c, db, sl, store, config) })
	app.Get("/outbreaks/edit/:i", func(c *fiber.Ctx) error { return handlers.HandlerOutbreakForm(c, db, sl, store, config) })
	app.Post("/outbreaks/save", func(c *fiber.Ctx) error { return handlers.HandlerOutbreakSubmit(c, db, sl, store, config) })
	app.Get("/outbreaks/close/:i", func(c *fiber.Ctx) error { return handlers.HandlerOutbreakClose(c, db, sl, store, config) })
	app.Post("/outbreaks/select/:i", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("i"))
		if err != nil {
			return c.Status(400).SendString("Invalid outbreak ID")
		}
		if err := handlers.SetSelectedOutbreak(c, store, id); err != nil {
			return c.Status(500).SendString("Failed to select outbreak")
		}
		return c.SendStatus(200)
	})
}

// SetupRoutes configures all routes for the application
func SetupRoutes(app *fiber.App, db *sql.DB, store *session.Store, sl *slog.Logger, config handlers.Config, smsService *services.SMSService) {
	// Public routes
	app.Get("/", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, nil, "landing")
	})

	// Login routes
	app.Get("/login", func(c *fiber.Ctx) error {
		return handlers.GenerateHTML(c, db, nil, "login")
	})
	app.Post("/login", func(c *fiber.Ctx) error {
		return handlers.HandlerLoginSubmit(c, db, sl, store, config)
	})
	app.Get("/logout", func(c *fiber.Ctx) error {
		return handlers.HandlerLoginOut(c, sl, store, config)
	})

	// Protected routes
	protected := app.Group("/", AuthRequired(store))

	// VHF CIF routes
	protected.Get("/vhf-cif", func(c *fiber.Ctx) error { return handlers.GenerateHTML(c, db, nil, "vhf_cif") })
	protected.Post("/vhf-cif", func(c *fiber.Ctx) error {
		return handlers.HandlerVHFPatientSubmit(c, db, sl, store, config, smsService)
	})
	protected.Get("/vhf-cif/clinical-signs/:id", func(c *fiber.Ctx) error { return handlers.GenerateHTML(c, db, nil, "vhf_clinical_signs") })
	protected.Post("/vhf-cif/clinical-signs/:id", func(c *fiber.Ctx) error { return handlers.HandlerVHFClinicalSignsSubmit(c, db, sl, store, config) })
	protected.Get("/vhf-cif/hospitalization/:id", func(c *fiber.Ctx) error { return handlers.GenerateHTML(c, db, nil, "vhf_hospitalization") })
	protected.Post("/vhf-cif/hospitalization/:id", func(c *fiber.Ctx) error { return handlers.HandlerVHFHospitalizationSubmit(c, db, sl, store, config) })
	protected.Get("/vhf-cif/risk-factors/:id", func(c *fiber.Ctx) error { return handlers.GenerateHTML(c, db, nil, "vhf_risk_factors") })
	protected.Post("/vhf-cif/risk-factors/:id", func(c *fiber.Ctx) error { return handlers.HandlerVHFRiskFactorsSubmit(c, db, sl, store, config) })
	protected.Get("/vhf-cif/laboratory/:id", func(c *fiber.Ctx) error { return handlers.GenerateHTML(c, db, nil, "vhf_laboratory") })
	protected.Post("/vhf-cif/laboratory/:id", func(c *fiber.Ctx) error { return handlers.HandlerVHFLaboratorySubmit(c, db, sl, store, config) })
	protected.Get("/vhf-cif/investigator/:id", func(c *fiber.Ctx) error { return handlers.GenerateHTML(c, db, nil, "vhf_investigator") })
	protected.Post("/vhf-cif/investigator/:id", func(c *fiber.Ctx) error { return handlers.HandlerVHFInvestigatorSubmit(c, db, sl, store, config) })
	protected.Get("/vhf-cif/success", func(c *fiber.Ctx) error { return handlers.HandlerVHFSuccess(c, db, sl, store, config) })
	protected.Get("/vhf/list", func(c *fiber.Ctx) error { return handlers.HandlerVHFList(c, db, sl, store, config) })
	protected.Get("/vhf/view/:id", func(c *fiber.Ctx) error { return handlers.HandlerVHFView(c, db, sl, store, config) })
	protected.Get("/vhf-lab/:id", func(c *fiber.Ctx) error { return handlers.HandlerVHFLabForm(c, db) })
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
}
