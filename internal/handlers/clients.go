package handlers

import (
	"case/internal/models"
	"case/internal/security"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// Define a struct for the encounter form page data
type EncounterPageData struct {
	FormRef       models.Client
	Form          []models.ClientEncounter
	Date          string
	FormChild1    []models.Clinical
	FormChild2    []models.Vital
	FormChild3    []models.Lab
	FormChild4    []models.Treatment
	AllEncounters []models.ClientEncounter // Add field for all encounters
	Optionz       map[string]map[string]string
}

func HandlerCasesForm(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	DoZaLogging("INFO", "Starting Client form", nil)

	userID, userName := GetUser(c, sl, store)
	role := security.GetRoles(userID, "admin")
	id, err := strconv.Atoi(c.Params("i"))
	data := NewTemplateData(c, store)

	var client models.Client

	if err != nil || id < 1 {
		client.ID = 0
		data.IsIDPos = false
	} else {
		c, err := models.ClientByID(c.Context(), db, id)
		if err == nil {
			client = *c
		}

		data.IsIDPos = true
	}

	// Get outbreak ID from session
	sess, err := store.Get(c)
	if err != nil {
		return c.Status(400).SendString("Failed to get session")
	}
	outbreakID := sess.Get("outbreak_id")
	if outbreakID == nil {
		return c.Status(400).SendString("No outbreak selected")
	}

	// Set outbreak ID for new cases
	data.OutbreakID = outbreakID.(int)
	data.IsOutbreakID = data.OutbreakID > 0
	if client.ID == 0 {
		client.OutbreakID = sql.NullInt64{Int64: int64(outbreakID.(int)), Valid: true}
	}

	cE, err := models.ClientEncounterz(c.Context(), db, "client_id="+strconv.Itoa(id), outbreakID.(int))
	if err != nil {
		DoZaLogging("ERROR", "Failed to get encounters", err)
	}

	st, err := models.Statuses(c.Context(), db, "client_id="+strconv.Itoa(id))
	if err != nil {
		DoZaLogging("ERROR", "Failed to get statuses", err)
	}

	// Check if there is an Mpox admission for this client
	hasAdmission := false
	var admissionID int
	err = db.QueryRow("SELECT id FROM mpox_demographics WHERE client_id = $1 LIMIT 1", client.ID).Scan(&admissionID)
	if err == nil {
		hasAdmission = true
	}
	data.HasMpoxAdmission = hasAdmission
	data.MpoxAdmissionID = admissionID

	data.User = userName
	data.Role = role
	data.Optionz = Get_Client_Optionz()
	data.Form = client
	data.FormChild1 = cE
	data.FormChild2 = st

	DoZaLogging("INFO", "Load Client form", err)
	return GenerateHTML(c, db, data, "form_patients")
}

func HandlerCasesSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {

	id, er := strconv.Atoi(c.FormValue("id"))
	if er != nil {
		id = 0
	}

	client := models.Client{
		ID:               id,
		Firstname:        ParseNullString(c.FormValue("firstname")),
		Lastname:         ParseNullString(c.FormValue("lastname")),
		Othername:        ParseNullString(c.FormValue("othername")),
		Gender:           ParseNullInt(c.FormValue("gender")),
		DateOfBirth:      ParseNullString(c.FormValue("date_of_birth")),
		Age:              ParseNullFloat(c.FormValue("age")),
		Marital:          ParseNullInt(c.FormValue("marital")),
		Nin:              ParseNullString(c.FormValue("nin")),
		Nationality:      ParseNullInt(c.FormValue("nationality")),
		AdmDate:          ParseNullString(c.FormValue("adm_date")),
		AdmFrom:          ParseNullString(c.FormValue("adm_from")),
		LabNo:            ParseNullString(c.FormValue("lab_no")),
		CifNo:            ParseNullString(c.FormValue("cif_no")),
		EtuNo:            ParseNullString(c.FormValue("etu_no")),
		CaseNo:           ParseNullString(c.FormValue("case_no")),
		Occupation:       ParseNullInt(c.FormValue("occupation")),
		OccupationAza:    ParseNullString(c.FormValue("occupation_aza")),
		DateSymptomOnset: ParseNullString(c.FormValue("date_symptom_onset")),
		DateIsolation:    ParseNullString(c.FormValue("date_isolation")),
		Pregnant:         ParseNullInt(c.FormValue("pregnant")),
		AdmWard:          ParseNullString(c.FormValue("adm_ward")),
		Tb:               ParseNullInt(c.FormValue("tb")),
		Asplenia:         ParseNullInt(c.FormValue("asplenia")),
		Hep:              ParseNullInt(c.FormValue("hep")),
		Diabetes:         ParseNullInt(c.FormValue("diabetes")),
		Hiv:              ParseNullInt(c.FormValue("hiv")),
		Liver:            ParseNullInt(c.FormValue("liver")),
		Malignancy:       ParseNullInt(c.FormValue("malignancy")),
		Heart:            ParseNullInt(c.FormValue("heart")),
		Pulmonary:        ParseNullInt(c.FormValue("pulmonary")),
		Kidney:           ParseNullInt(c.FormValue("kidney")),
		Neurologic:       ParseNullInt(c.FormValue("neurologic")),
		Other:            ParseNullString(c.FormValue("other")),
		Transfer:         ParseNullInt(c.FormValue("transfer")),
		Site:             ParseNullInt(c.FormValue("site")),
		Status:           ParseNullString(c.FormValue("status")),

		//Status: ParseNullString(c.FormValue("status")),
	}

	//visID, _ := utilities.GetSequentialVisitID()
	userID := GetCurrentUser(c, store)

	client.EditOn.Valid = true
	client.EditBy.Valid = true

	client.EditBy.Int64 = int64(userID)
	client.EditOn.Time = time.Now()

	if client.ID == 0 {

		client.EnterOn.Valid = true
		client.EnterBy.Valid = true

		client.EnterBy.Int64 = int64(userID)
		client.EnterOn.Time = time.Now()

		client.UUID.Valid = true
		client.UUID.String = models.CreateUUID()

		//appID := models.CreateUUID()
		//client.UUID.String = appID

	}

	if client.ID == 0 {
		err := client.Insert(c.Context(), db)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		client.SetAsExists()
		err := client.Update(c.Context(), db)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	urlx := "/cases/new/" + strconv.Itoa(client.ID)

	return c.Redirect(urlx)
}

func HandlerCasesList(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	fmt.Println("starting case list")

	userID, userName := GetUser(c, sl, store)
	role := security.GetRoles(userID, "admin")

	data := NewTemplateData(c, store)
	data.User = userName
	data.Role = role

	fmt.Println("loading case list page")

	// Get outbreak ID from session using helper function
	outbreakID := GetCurrentOutbreak(c, store)
	if outbreakID == 0 {
		sl.Error("No outbreak selected for user", "user_id", userID)
		return c.Status(400).SendString("No outbreak selected. Please select an outbreak first.")
	}

	sl.Info("Loading cases for outbreak", "user_id", userID, "outbreak_id", outbreakID)

	// Get user's facility from session
	userFacility := GetCurrentFacility(c, db, sl, store)

	// Build filter based on outbreak and facility
	filter := fmt.Sprintf("outbreak_id = %d", outbreakID)

	// If user has a facility assigned, filter by that facility
	if userFacility > 0 {
		filter += fmt.Sprintf(" AND site = %d", userFacility)
		sl.Info("Filtering cases by user facility", "user_id", userID, "facility_id", userFacility)
	} else {
		sl.Info("No facility assigned to user, showing all cases for outbreak", "user_id", userID)
	}

	clients, err := models.Clients(c.Context(), db, filter)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			fmt.Println("error loading case list: ", err.Error())
		} else {
			fmt.Println("error loading case list: ", err.Error())
		}
	}

	data.Form = clients

	return GenerateHTML(c, db, data, "list_patients")
}

func HandlerCaseEncounterForm(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	// Get client ID from URL path parameter
	clientIDStr := c.Params("i")
	if clientIDStr == "" {
		sl.Error("Client ID is missing from URL path")
		return c.Status(400).SendString("Client ID is required")
	}

	// Convert client ID to int
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		sl.Error("Invalid client ID", "error", err, "clientID", clientIDStr)
		return c.Status(400).SendString("Invalid client ID")
	}

	// Get outbreak ID from session
	sess, err := store.Get(c)
	if err != nil {
		sl.Error("Failed to get session", "error", err)
		return c.Status(500).SendString("Failed to get session")
	}
	outbreakID := sess.Get("outbreak_id")
	if outbreakID == nil {
		sl.Error("No outbreak selected")
		return c.Status(400).SendString("No outbreak selected")
	}

	// Get encounter date from query parameter
	encounterDate := c.Query("dte")

	// Validate date format
	if encounterDate == "" || encounterDate == "0000-00-00" {
		// If no date provided or invalid date, use current date
		encounterDate = time.Now().Format("2006-01-02")
	} else {
		// Try to parse the date to validate it
		if _, err := time.Parse("2006-01-02", encounterDate); err != nil {
			// If date is invalid, use current date
			encounterDate = time.Now().Format("2006-01-02")
		}
	}

	// Get client details first
	client, err := models.ClientByID(c.Context(), db, clientID)
	if err != nil {
		sl.Error("Failed to get client", "error", err, "clientID", clientID)
		return c.Status(500).SendString("Failed to get client details")
	}

	// Get all encounters for this client (not filtered by date)
	sl.Info("Fetching all encounters", "clientID", clientID, "outbreakID", outbreakID.(int))
	allEncounters, err := models.ClientEncounters(c.Context(), db, fmt.Sprintf("client_id = %d", clientID), outbreakID.(int))
	if err != nil {
		sl.Error("Failed to get all encounters", "error", err, "clientID", clientID)
		allEncounters = []models.ClientEncounter{}
	} else {
		sl.Info("Successfully fetched encounters", "count", len(allEncounters), "clientID", clientID)
	}

	// Get encounters for the specific date (for editing existing encounters)
	dateEncounters, err := models.ClientEncounters(c.Context(), db, fmt.Sprintf("client_id = %d AND encounter_date = '%s'", clientID, encounterDate), outbreakID.(int))
	if err != nil {
		sl.Error("Failed to get encounters for date", "error", err, "clientID", clientID, "date", encounterDate)
		dateEncounters = []models.ClientEncounter{}
	}

	// Use date-specific encounters if they exist, otherwise create empty encounter
	var encounters []models.ClientEncounter
	if len(dateEncounters) > 0 {
		encounters = dateEncounters
	} else {
		// Create an empty encounter with the current date
		emptyEncounter := models.ClientEncounter{
			EncounterID:   0,
			EncounterType: sql.NullInt64{Int64: 0, Valid: false},
			EmployeeFname: sql.NullString{String: "", Valid: true},
			EmployeeLname: sql.NullString{String: "", Valid: true},
			EncounterDate: sql.NullString{String: encounterDate, Valid: true},
			EncounterTime: sql.NullString{String: "", Valid: true},
			ClinicalTeam:  sql.NullString{String: "", Valid: true},
			ManagedBy:     sql.NullInt64{Int64: 0, Valid: false},
			ClientID:      clientID,
		}
		encounters = append(encounters, emptyEncounter)
	}

	// Get clinical data for the first encounter
	var clinical []models.Clinical
	if len(encounters) > 0 && encounters[0].EncounterID > 0 {
		clinicalData, err := models.ClinicalByEncounterID(c.Context(), db, encounters[0].EncounterID)
		if err == nil && clinicalData != nil {
			clinical = append(clinical, *clinicalData)
		}
	}
	if len(clinical) == 0 {
		// Add empty clinical data
		clinical = append(clinical, models.Clinical{
			ClinicalID:            0,
			PharyngealErythema:    sql.NullInt64{Int64: 0, Valid: true},
			PharyngealExudate:     sql.NullInt64{Int64: 0, Valid: true},
			ConjunctivalInjection: sql.NullInt64{Int64: 0, Valid: true},
			OedemaFace:            sql.NullInt64{Int64: 0, Valid: true},
			TenderAbdomen:         sql.NullInt64{Int64: 0, Valid: true},
			SunkenEyes:            sql.NullInt64{Int64: 0, Valid: true},
			TentingSkin:           sql.NullInt64{Int64: 0, Valid: true},
			PalpableLiver:         sql.NullInt64{Int64: 0, Valid: true},
			PalpableSpleen:        sql.NullInt64{Int64: 0, Valid: true},
			Jaundice:              sql.NullInt64{Int64: 0, Valid: true},
			EnlargedLymphNodes:    sql.NullInt64{Int64: 0, Valid: true},
			LowerExtremityOedema:  sql.NullInt64{Int64: 0, Valid: true},
			Bleeding:              sql.NullInt64{Int64: 0, Valid: true},
			BleedingNose:          sql.NullInt64{Int64: 0, Valid: true},
			BleedingMouth:         sql.NullInt64{Int64: 0, Valid: true},
			BleedingVagina:        sql.NullInt64{Int64: 0, Valid: true},
			BleedingRectum:        sql.NullInt64{Int64: 0, Valid: true},
			Shock:                 sql.NullInt64{Int64: 0, Valid: true},
			Meningitis:            sql.NullInt64{Int64: 0, Valid: true},
			Confusion:             sql.NullInt64{Int64: 0, Valid: true},
			Seizure:               sql.NullInt64{Int64: 0, Valid: true},
			Coma:                  sql.NullInt64{Int64: 0, Valid: true},
			Bacteraemia:           sql.NullInt64{Int64: 0, Valid: true},
			Hyperglycemia:         sql.NullInt64{Int64: 0, Valid: true},
			Hypoglycemia:          sql.NullInt64{Int64: 0, Valid: true},
		})
	}
	// Ensure we have 3 clinical records for morning, afternoon, evening
	for len(clinical) < 3 {
		clinical = append(clinical, models.Clinical{
			ClinicalID:            0,
			PharyngealErythema:    sql.NullInt64{Int64: 0, Valid: true},
			PharyngealExudate:     sql.NullInt64{Int64: 0, Valid: true},
			ConjunctivalInjection: sql.NullInt64{Int64: 0, Valid: true},
			OedemaFace:            sql.NullInt64{Int64: 0, Valid: true},
			TenderAbdomen:         sql.NullInt64{Int64: 0, Valid: true},
			SunkenEyes:            sql.NullInt64{Int64: 0, Valid: true},
			TentingSkin:           sql.NullInt64{Int64: 0, Valid: true},
			PalpableLiver:         sql.NullInt64{Int64: 0, Valid: true},
			PalpableSpleen:        sql.NullInt64{Int64: 0, Valid: true},
			Jaundice:              sql.NullInt64{Int64: 0, Valid: true},
			EnlargedLymphNodes:    sql.NullInt64{Int64: 0, Valid: true},
			LowerExtremityOedema:  sql.NullInt64{Int64: 0, Valid: true},
			Bleeding:              sql.NullInt64{Int64: 0, Valid: true},
			BleedingNose:          sql.NullInt64{Int64: 0, Valid: true},
			BleedingMouth:         sql.NullInt64{Int64: 0, Valid: true},
			BleedingVagina:        sql.NullInt64{Int64: 0, Valid: true},
			BleedingRectum:        sql.NullInt64{Int64: 0, Valid: true},
			Shock:                 sql.NullInt64{Int64: 0, Valid: true},
			Meningitis:            sql.NullInt64{Int64: 0, Valid: true},
			Confusion:             sql.NullInt64{Int64: 0, Valid: true},
			Seizure:               sql.NullInt64{Int64: 0, Valid: true},
			Coma:                  sql.NullInt64{Int64: 0, Valid: true},
			Bacteraemia:           sql.NullInt64{Int64: 0, Valid: true},
			Hyperglycemia:         sql.NullInt64{Int64: 0, Valid: true},
			Hypoglycemia:          sql.NullInt64{Int64: 0, Valid: true},
		})
	}

	// Get vitals data for the first encounter
	var vitals []models.Vital
	if len(encounters) > 0 && encounters[0].EncounterID > 0 {
		vitalsData, err := models.VitalByEncounterID(c.Context(), db, encounters[0].EncounterID)
		if err == nil && vitalsData != nil {
			vitals = append(vitals, *vitalsData)
		}
	}
	if len(vitals) == 0 {
		// Add empty vitals data
		vitals = append(vitals, models.Vital{
			VitalsID:        0,
			HeartRate:       sql.NullFloat64{Float64: 0, Valid: true},
			BpSystolic:      sql.NullFloat64{Float64: 0, Valid: true},
			BpDiastolic:     sql.NullFloat64{Float64: 0, Valid: true},
			RespiratoryRate: sql.NullFloat64{Float64: 0, Valid: true},
			Saturation:      sql.NullFloat64{Float64: 0, Valid: true},
			Weight:          sql.NullFloat64{Float64: 0, Valid: true},
			Height:          sql.NullFloat64{Float64: 0, Valid: true},
			Temperature:     sql.NullFloat64{Float64: 0, Valid: true},
			MentalStatus:    sql.NullString{String: "", Valid: true},
			Muac:            sql.NullFloat64{Float64: 0, Valid: true},
		})
	}
	// Ensure we have 3 vitals records for morning, afternoon, evening
	for len(vitals) < 3 {
		vitals = append(vitals, models.Vital{
			VitalsID:        0,
			HeartRate:       sql.NullFloat64{Float64: 0, Valid: true},
			BpSystolic:      sql.NullFloat64{Float64: 0, Valid: true},
			BpDiastolic:     sql.NullFloat64{Float64: 0, Valid: true},
			RespiratoryRate: sql.NullFloat64{Float64: 0, Valid: true},
			Saturation:      sql.NullFloat64{Float64: 0, Valid: true},
			Weight:          sql.NullFloat64{Float64: 0, Valid: true},
			Height:          sql.NullFloat64{Float64: 0, Valid: true},
			Temperature:     sql.NullFloat64{Float64: 0, Valid: true},
			MentalStatus:    sql.NullString{String: "", Valid: true},
			Muac:            sql.NullFloat64{Float64: 0, Valid: true},
		})
	}

	// Get lab data for the first encounter
	var labs []models.Lab
	if len(encounters) > 0 && encounters[0].EncounterID > 0 {
		labData, err := models.LabByEncounterID(c.Context(), db, encounters[0].EncounterID)
		if err == nil && labData != nil {
			labs = append(labs, *labData)
		}
	}
	if len(labs) == 0 {
		// Add empty lab data
		labs = append(labs, models.Lab{
			LabID: 0,
		})
	}
	// Ensure we have at least 1 lab record
	for len(labs) < 1 {
		labs = append(labs, models.Lab{
			LabID: 0,
		})
	}

	// Get treatment data for the first encounter
	var treatments []models.Treatment
	if len(encounters) > 0 && encounters[0].EncounterID > 0 {
		treatmentData, err := models.TreatmentByEncounterID(c.Context(), db, encounters[0].EncounterID)
		if err == nil && treatmentData != nil {
			treatments = append(treatments, *treatmentData)
		}
	}
	if len(treatments) == 0 {
		// Add empty treatment data
		treatments = append(treatments, models.Treatment{
			TreatmentID: 0,
		})
	}
	// Ensure we have at least 1 treatment record
	for len(treatments) < 1 {
		treatments = append(treatments, models.Treatment{
			TreatmentID: 0,
		})
	}

	// Prepare strongly typed data for the template
	data := EncounterPageData{
		FormRef:       *client,
		Form:          encounters,
		Date:          encounterDate,
		FormChild1:    clinical,
		FormChild2:    vitals,
		FormChild3:    labs,
		FormChild4:    treatments,
		AllEncounters: allEncounters, // Add all encounters
		Optionz:       Get_Client_Optionz(),
	}

	// Debug: Log the client data being passed to template
	sl.Info("Client data for template",
		"clientID", client.ID,
		"firstName", client.Firstname.String,
		"lastName", client.Lastname.String,
		"genderValid", client.Gender.Valid,
		"genderValue", client.Gender.Int64,
		"outbreakIDValid", client.OutbreakID.Valid,
		"outbreakIDValue", client.OutbreakID.Int64,
	)

	return GenerateHTML(c, db, data, "form_encounters")
}

func HandlerCaseEncounterList(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	// Get client ID from query parameter
	ClientIDStr := c.Query("client_id")
	if ClientIDStr == "" {
		return c.Status(400).SendString("Client ID is required")
	}

	// Convert client ID to int
	ClientID, err := strconv.Atoi(ClientIDStr)
	if err != nil {
		return c.Status(400).SendString("Invalid client ID")
	}

	// Get outbreak ID from session
	sess, err := store.Get(c)
	if err != nil {
		return c.Status(400).SendString("Failed to get session")
	}
	outbreakID := sess.Get("outbreak_id")
	if outbreakID == nil {
		return c.Status(400).SendString("No outbreak selected")
	}

	// Get encounters for this client
	encounters, err := models.ClientEncounterz(c.Context(), db, fmt.Sprintf("client_id = %d", ClientID), outbreakID.(int))
	if err != nil {
		sl.Error("Failed to get encounters", "error", err)
		return c.Status(500).SendString("Failed to get encounters")
	}

	// Get client details
	client, err := models.ClientByID(c.Context(), db, ClientID)
	if err != nil {
		sl.Error("Failed to get client", "error", err)
		return c.Status(500).SendString("Failed to get client")
	}

	data := NewTemplateData(c, store)
	data.Form = fiber.Map{
		"Client":     client,
		"Encounters": encounters,
	}

	return GenerateHTML(c, db, data, "list_case_encounter")
}

func saveEncounter(c *fiber.Ctx, db *sql.DB, userID int, cid, dte string) (int, int, int, error) {
	// Convert client ID to int
	clientID, err := strconv.Atoi(cid)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid client ID: %v", err)
	}

	// Get outbreak ID from session
	sess, err := session.New().Get(c)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get session: %v", err)
	}
	outbreakID := sess.Get("outbreak_id")
	if outbreakID == nil {
		return 0, 0, 0, fmt.Errorf("no outbreak selected")
	}

	// Create encounter
	ClinicalTeam := "0"
	encounter := models.Encounter{
		EncounterType: sql.NullInt64{Int64: 1, Valid: true}, // Assuming 1 is the default encounter type
		EncounterTime: sql.NullString{String: time.Now().Format("15:04:05"), Valid: true},
		ClientID:      sql.NullInt64{Int64: int64(clientID), Valid: true},
		EncounterDate: sql.NullString{String: dte, Valid: true},
		ManagedBy:     sql.NullInt64{Int64: int64(userID), Valid: true},
		EnterOn:       sql.NullTime{Time: time.Now(), Valid: true},
		EnterBy:       sql.NullInt64{Int64: int64(userID), Valid: true},
		OutbreakID:    sql.NullInt64{Int64: int64(outbreakID.(int)), Valid: true},
		ClinicalTeam:  sql.NullString{String: ClinicalTeam, Valid: true},
	}

	// Save encounter
	if err := encounter.Insert(c.Context(), db); err != nil {
		return 0, 0, 0, fmt.Errorf("failed to save encounter: %v", err)
	}

	return encounter.EncounterID, clientID, outbreakID.(int), nil
}

func saveVitals(c *fiber.Ctx, db *sql.DB, id1, id2, id3 int) error {
	vital_id, err := strconv.Atoi(c.FormValue("vital_id"))
	if err != nil {
		vital_id = 0
	}

	vital := models.Vital{
		VitalsID:            vital_id,
		EncounterID:         sql.NullInt64{Int64: int64(id1), Valid: true},
		HeartRate:           ParseNullFloat(c.FormValue("heart_rate")),
		BpSystolic:          ParseNullFloat(c.FormValue("bp_systolic")),
		BpDiastolic:         ParseNullFloat(c.FormValue("bp_diastolic")),
		CapillaryRefill:     ParseNullInt(c.FormValue("capillary_refill")),
		RespiratoryRate:     ParseNullFloat(c.FormValue("respiratory_rate")),
		Saturation:          ParseNullFloat(c.FormValue("saturation")),
		Weight:              ParseNullFloat(c.FormValue("weight")),
		Height:              ParseNullFloat(c.FormValue("height")),
		Temperature:         ParseNullFloat(c.FormValue("temperature")),
		LowestConsciousness: ParseNullString(c.FormValue("lowest_consciousness")),
		MentalStatus:        ParseNullString(c.FormValue("mental_status")),
		Muac:                ParseNullFloat(c.FormValue("muac")),
		Bleeding:            ParseNullInt(c.FormValue("bleeding")),
		Shock:               ParseNullInt(c.FormValue("shock")),
		Meningitis:          ParseNullInt(c.FormValue("meningitis")),
		Confusion:           ParseNullInt(c.FormValue("confusion")),
		Seizure:             ParseNullInt(c.FormValue("seizure")),
		Coma:                ParseNullInt(c.FormValue("coma")),
		Bacteraemia:         ParseNullInt(c.FormValue("bacteraemia")),
		Hyperglycemia:       ParseNullInt(c.FormValue("hyperglycemia")),
		Hypoglycemia:        ParseNullInt(c.FormValue("hypoglycemia")),
		Other:               ParseNullString(c.FormValue("other")),
	}

	if vital_id == 0 {
		return vital.Insert(c.Context(), db)
	} else {
		vital.SetAsExists()
		return vital.Update(c.Context(), db)
	}
}

func getZaFormValue(c *fiber.Ctx, zname string, i int) string {
	return c.FormValue(fmt.Sprintf("%s%d", zname, i))
}

func saveClinical(c *fiber.Ctx, db *sql.DB, id1, id2, id3 int) error {
	clinical_id, err := strconv.Atoi(c.FormValue("clinical_id"))
	if err != nil {
		clinical_id = 0
	}

	clinical := models.Clinical{
		ClinicalID:              clinical_id,
		EncounterID:             sql.NullInt64{Int64: int64(id1), Valid: true},
		Fever:                   ParseNullInt(c.FormValue("fever")),
		Fatigue:                 ParseNullInt(c.FormValue("fatigue")),
		Weakness:                ParseNullInt(c.FormValue("weakness")),
		Malaise:                 ParseNullInt(c.FormValue("malaise")),
		Myalgia:                 ParseNullInt(c.FormValue("myalgia")),
		Anorexia:                ParseNullInt(c.FormValue("anorexia")),
		SoreThroat:              ParseNullInt(c.FormValue("sore_throat")),
		Headache:                ParseNullInt(c.FormValue("headache")),
		Nausea:                  ParseNullInt(c.FormValue("nausea")),
		ChestPain:               ParseNullInt(c.FormValue("chest_pain")),
		JointPain:               ParseNullInt(c.FormValue("joint_pain")),
		Hiccups:                 ParseNullInt(c.FormValue("hiccups")),
		Cough:                   ParseNullInt(c.FormValue("cough")),
		DifficultyBreathing:     ParseNullInt(c.FormValue("difficulty_breathing")),
		DifficultySwallowing:    ParseNullInt(c.FormValue("difficulty_swallowing")),
		AbdominalPain:           ParseNullInt(c.FormValue("abdominal_pain")),
		Diarrhoea:               ParseNullInt(c.FormValue("diarrhoea")),
		Vomiting:                ParseNullInt(c.FormValue("vomiting")),
		Irritability:            ParseNullInt(c.FormValue("irritability")),
		Dysphagia:               ParseNullInt(c.FormValue("dysphagia")),
		UnusualBleeding:         ParseNullInt(c.FormValue("unusual_bleeding")),
		Dehydration:             ParseNullInt(c.FormValue("dehydration")),
		Shock:                   ParseNullInt(c.FormValue("shock")),
		Anuria:                  ParseNullInt(c.FormValue("anuria")),
		Disorientation:          ParseNullInt(c.FormValue("disorientation")),
		Agitation:               ParseNullInt(c.FormValue("agitation")),
		Seizure:                 ParseNullInt(c.FormValue("seizure")),
		Meningitis:              ParseNullInt(c.FormValue("meningitis")),
		Confusion:               ParseNullInt(c.FormValue("confusion")),
		Coma:                    ParseNullInt(c.FormValue("coma")),
		Bacteraemia:             ParseNullInt(c.FormValue("bacteraemia")),
		Hyperglycemia:           ParseNullInt(c.FormValue("hyperglycemia")),
		Hypoglycemia:            ParseNullInt(c.FormValue("hypoglycemia")),
		OtherComplications:      ParseNullInt(c.FormValue("other_complications")),
		AzaComplicationsSpecif:  ParseNullString(c.FormValue("aza_complications_specif")),
		PharyngealErythema:      ParseNullInt(c.FormValue("pharyngeal_erythema")),
		PharyngealExudate:       ParseNullInt(c.FormValue("pharyngeal_exudate")),
		ConjunctivalInjection:   ParseNullInt(c.FormValue("conjunctival_injection")),
		OedemaFace:              ParseNullInt(c.FormValue("oedema_face")),
		TenderAbdomen:           ParseNullInt(c.FormValue("tender_abdomen")),
		SunkenEyes:              ParseNullInt(c.FormValue("sunken_eyes")),
		TentingSkin:             ParseNullInt(c.FormValue("tenting_skin")),
		PalpableLiver:           ParseNullInt(c.FormValue("palpable_liver")),
		PalpableSpleen:          ParseNullInt(c.FormValue("palpable_spleen")),
		Jaundice:                ParseNullInt(c.FormValue("jaundice")),
		EnlargedLymphNodes:      ParseNullInt(c.FormValue("enlarged_lymph_nodes")),
		LowerExtremityOedema:    ParseNullInt(c.FormValue("lower_extremity_oedema")),
		Bleeding:                ParseNullInt(c.FormValue("bleeding")),
		BleedingNose:            ParseNullInt(c.FormValue("bleeding_nose")),
		BleedingMouth:           ParseNullInt(c.FormValue("bleeding_mouth")),
		BleedingVagina:          ParseNullInt(c.FormValue("bleeding_vagina")),
		BleedingRectum:          ParseNullInt(c.FormValue("bleeding_rectum")),
		BleedingSputum:          ParseNullInt(c.FormValue("bleeding_sputum")),
		BleedingUrine:           ParseNullInt(c.FormValue("bleeding_urine")),
		BleedingIvSite:          ParseNullInt(c.FormValue("bleeding_iv_site")),
		BleedingOther:           ParseNullInt(c.FormValue("bleeding_other")),
		BleedingOtherSpecif:     ParseNullString(c.FormValue("bleeding_other_specif")),
		FinalDiagnosis:          ParseNullInt(c.FormValue("final_diagnosis")),
		FinalDiagnosisAza:       ParseNullString(c.FormValue("final_diagnosis_aza")),
		OutcomeDischarge:        ParseNullInt(c.FormValue("outcome_discharge")),
		OutcomeDischargeIfHear:  ParseNullInt(c.FormValue("outcome_discharge_if_hear")),
		OutcomeDischargeIfArth:  ParseNullInt(c.FormValue("outcome_discharge_if_arth")),
		OutcomeDischargeIfAbor:  ParseNullInt(c.FormValue("outcome_discharge_if_abor")),
		OutcomeDischargeIfNeur:  ParseNullInt(c.FormValue("outcome_discharge_if_neur")),
		OutcomeDischargeIfOcul:  ParseNullInt(c.FormValue("outcome_discharge_if_ocul")),
		OutcomeDischargeIfExtr:  ParseNullInt(c.FormValue("outcome_discharge_if_extr")),
		OutcomeDischargeIfOthe:  ParseNullInt(c.FormValue("outcome_discharge_if_othe")),
		OutcomeDischargeIfAza:   ParseNullString(c.FormValue("outcome_discharge_if_aza")),
		OutcomeReferredFacility: ParseNullString(c.FormValue("outcome_referred_facility")),
		DischargeDate:           ParseNullString(c.FormValue("discharge_date")),
		SurvivorCounselling:     ParseNullInt(c.FormValue("survivor_counselling")),
	}

	if clinical_id == 0 {
		return clinical.Insert(c.Context(), db)
	} else {
		clinical.SetAsExists()
		return clinical.Update(c.Context(), db)
	}
}

func HandlerCaseEncounterSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	// Get user ID from session
	sess, err := store.Get(c)
	if err != nil {
		sl.Error("Failed to get session", "error", err)
		return c.Status(500).SendString("Internal server error")
	}
	userID := sess.Get("user_id")
	if userID == nil {
		sl.Error("User not authenticated")
		return c.Status(401).SendString("Not authenticated")
	}

	// Get client ID and validate
	cid := c.FormValue("client_id")
	if cid == "" {
		sl.Error("Missing client ID")
		return c.Status(400).SendString("Missing client ID")
	}

	// Get encounter date and validate format
	dte := c.FormValue("encounter_date")
	if dte == "" {
		sl.Error("Missing encounter date")
		return c.Status(400).SendString("Missing encounter date")
	}
	if _, err := time.Parse("2006-01-02", dte); err != nil {
		sl.Error("Invalid encounter date format", "error", err)
		return c.Status(400).SendString("Invalid encounter date format")
	}

	// Get outbreak ID from session
	outbreakID := sess.Get("outbreak_id")
	if outbreakID == nil {
		sl.Error("No outbreak selected")
		return c.Status(400).SendString("No outbreak selected")
	}

	// Save encounter and get IDs
	id1, id2, id3, err := saveEncounter(c, db, userID.(int), cid, dte)
	if err != nil {
		sl.Error("Failed to save encounter", "error", err)
		return c.Status(500).SendString("Failed to save encounter")
	}

	// Save vitals
	saveVitals(c, db, id1, id2, id3)

	// Save clinical data
	saveClinical(c, db, id1, id2, id3)

	// Save lab data
	if err := saveLab(c, db, id1); err != nil {
		sl.Error("Failed to save lab data", "error", err)
		return c.Status(500).SendString("Failed to save lab data")
	}

	// Save treatment data
	if err := saveTreatment(c, db, id1); err != nil {
		sl.Error("Failed to save treatment data", "error", err)
		return c.Status(500).SendString("Failed to save treatment data")
	}

	return c.Redirect(fmt.Sprintf("/cases/encounters/%d", id2))
}

// Helper function to save lab data
func saveLab(c *fiber.Ctx, db *sql.DB, encounterID int) error {
	lab_id, err := strconv.Atoi(c.FormValue("lab_id"))
	if err != nil {
		lab_id = 0
	}

	lab := models.Lab{
		LabID:                 lab_id,
		EncounterID:           sql.NullInt64{Int64: int64(encounterID), Valid: true},
		Specimen:              ParseNullInt(c.FormValue("specimen")),
		SampleBlood:           ParseNullInt(c.FormValue("sample_blood")),
		SampleUrine:           ParseNullInt(c.FormValue("sample_urine")),
		SampleSwab:            ParseNullInt(c.FormValue("sample_swab")),
		SampleAza:             ParseNullString(c.FormValue("sample_aza")),
		EbolaRdt:              ParseNullInt(c.FormValue("ebola_rdt")),
		EbolaRdtDate:          ParseNullString(c.FormValue("ebola_rdt_date")),
		EbolaRdtResults:       ParseNullInt(c.FormValue("ebola_rdt_results")),
		EbolaPcr:              ParseNullInt(c.FormValue("ebola_pcr")),
		EbolaPcrAza:           ParseNullString(c.FormValue("ebola_pcr_aza")),
		EbolaPcrDate:          ParseNullString(c.FormValue("ebola_pcr_date")),
		EbolaPcrGp:            ParseNullInt(c.FormValue("ebola_pcr_gp")),
		EbolaPcrGpCt:          ParseNullFloat(c.FormValue("ebola_pcr_gp_ct")),
		EbolaPcrNp:            ParseNullInt(c.FormValue("ebola_pcr_np")),
		EbolaPcrNpCt:          ParseNullFloat(c.FormValue("ebola_pcr_np_ct")),
		EbolaPcrIndeterminate: ParseNullInt(c.FormValue("ebola_pcr_indeterminate")),
		MalariaRdtDate:        ParseNullString(c.FormValue("malaria_rdt_date")),
		MalariaRdtResult:      ParseNullInt(c.FormValue("malaria_rdt_result")),
		BloodCultureDate:      ParseNullString(c.FormValue("blood_culture_date")),
		BloodCultureResult:    ParseNullInt(c.FormValue("blood_culture_result")),
		TestPosInfection:      ParseNullInt(c.FormValue("test_pos_infection")),
		TestPosInfectionAza:   ParseNullString(c.FormValue("test_pos_infection_aza")),
		Haemoglobinuria:       ParseNullInt(c.FormValue("haemoglobinuria")),
		Proteinuria:           ParseNullInt(c.FormValue("proteinuria")),
		Hematuria:             ParseNullInt(c.FormValue("hematuria")),
		BloodGas:              ParseNullInt(c.FormValue("blood_gas")),
		Ph:                    ParseNullFloat(c.FormValue("ph")),
		Pco2:                  ParseNullFloat(c.FormValue("pco2")),
		Pao2:                  ParseNullFloat(c.FormValue("pao2")),
		Hco3:                  ParseNullFloat(c.FormValue("hco3")),
		OxygenTherapy:         ParseNullFloat(c.FormValue("oxygen_therapy")),
		AltSgpt:               ParseNullFloat(c.FormValue("alt_sgpt")),
		AstSgo:                ParseNullFloat(c.FormValue("ast_sgo")),
		Creatinine:            ParseNullFloat(c.FormValue("creatinine")),
		Potassium:             ParseNullFloat(c.FormValue("potassium")),
		Urea:                  ParseNullFloat(c.FormValue("urea")),
		CreatinineKinase:      ParseNullFloat(c.FormValue("creatinine_kinase")),
		Calcium:               ParseNullFloat(c.FormValue("calcium")),
		Sodium:                ParseNullFloat(c.FormValue("sodium")),
		AltSgptNd:             ParseNullInt(c.FormValue("alt_sgpt_nd")),
		AstSgoNd:              ParseNullInt(c.FormValue("ast_sgo_nd")),
		CreatinineNd:          ParseNullInt(c.FormValue("creatinine_nd")),
		PotassiumNd:           ParseNullInt(c.FormValue("potassium_nd")),
		UreaNd:                ParseNullInt(c.FormValue("urea_nd")),
		CreatinineKinaseNd:    ParseNullInt(c.FormValue("creatinine_kinase_nd")),
		CalciumNd:             ParseNullInt(c.FormValue("calcium_nd")),
		SodiumNd:              ParseNullInt(c.FormValue("sodium_nd")),
		Glucose:               ParseNullFloat(c.FormValue("glucose")),
		Lactate:               ParseNullFloat(c.FormValue("lactate")),
		Haemoglobin:           ParseNullFloat(c.FormValue("haemoglobin")),
		TotalBilirubin:        ParseNullFloat(c.FormValue("total_bilirubin")),
		WbcCount:              ParseNullFloat(c.FormValue("wbc_count")),
		Platelets:             ParseNullFloat(c.FormValue("platelets")),
		Pt:                    ParseNullFloat(c.FormValue("pt")),
		Aptt:                  ParseNullFloat(c.FormValue("aptt")),
		GlucoseNd:             ParseNullInt(c.FormValue("glucose_nd")),
		LactateNd:             ParseNullInt(c.FormValue("lactate_nd")),
		HaemoglobinNd:         ParseNullInt(c.FormValue("haemoglobin_nd")),
		TotalBilirubinNd:      ParseNullInt(c.FormValue("total_bilirubin_nd")),
		WbcCountNd:            ParseNullInt(c.FormValue("wbc_count_nd")),
		PlateletsNd:           ParseNullInt(c.FormValue("platelets_nd")),
		PtNd:                  ParseNullInt(c.FormValue("pt_nd")),
		ApttNd:                ParseNullInt(c.FormValue("aptt_nd")),
	}

	if lab_id == 0 {
		return lab.Insert(c.Context(), db)
	} else {
		lab.SetAsExists()
		return lab.Update(c.Context(), db)
	}
}

// Helper function to save treatment data
func saveTreatment(c *fiber.Ctx, db *sql.DB, encounterID int) error {
	treat_id, err := strconv.Atoi(c.FormValue("treat_id"))
	if err != nil {
		treat_id = 0
	}

	treat := &models.Treatment{
		TreatmentID:                 treat_id,
		EncounterID:                 sql.NullInt64{Int64: int64(encounterID), Valid: true},
		Antibacterial:               ParseNullInt2(c.FormValue("antibacterial")),
		Amoxicillin:                 ParseNullInt2(c.FormValue("amoxicillin")),
		Ceftriaxone:                 ParseNullInt2(c.FormValue("ceftriaxone")),
		Cefixime:                    ParseNullInt2(c.FormValue("cefixime")),
		AntibacterialOther:          ParseNullString2(c.FormValue("antibacterial_other")),
		AntibacterialDose:           ParseNullString2(c.FormValue("antibacterial_dose")),
		AntibacterialRoute:          ParseNullString2(c.FormValue("antibacterial_route")),
		AntibacterialFreq:           ParseNullString2(c.FormValue("antibacterial_freq")),
		Antimalarial:                ParseNullInt2(c.FormValue("antimalarial")),
		AntimalarialArtesunate:      ParseNullInt2(c.FormValue("antimalarial_artesunate")),
		AntimalarialArthemeter:      ParseNullInt2(c.FormValue("antimalarial_arthemeter")),
		AntimalarialAl:              ParseNullInt2(c.FormValue("antimalarial_al")),
		AntimalarialAa:              ParseNullInt2(c.FormValue("antimalarial_aa")),
		AntimalarialDose:            ParseNullString2(c.FormValue("antimalarial_dose")),
		AntimalarialRoute:           ParseNullString2(c.FormValue("antimalarial_route")),
		AntimalarialFreq:            ParseNullString2(c.FormValue("antimalarial_freq")),
		OtherMedsSpecify:            ParseNullString2(c.FormValue("other_meds_specify")),
		OtherMedsDose:               ParseNullString2(c.FormValue("other_meds_dose")),
		OtherMedsRoute:              ParseNullString2(c.FormValue("other_meds_route")),
		OtherMedsFreq:               ParseNullString2(c.FormValue("other_meds_freq")),
		EbolaExperimental:           ParseNullInt2(c.FormValue("ebola_experimental")),
		EbolaExperimentalIf:         ParseNullString2(c.FormValue("ebola_experimental_if")),
		Oral:                        ParseNullInt2(c.FormValue("oral")),
		OralOrs:                     ParseNullInt2(c.FormValue("oral_ors")),
		OralOrsQty:                  ParseNullFloat(c.FormValue("oral_ors_qty")),
		OralWater:                   ParseNullInt2(c.FormValue("oral_water")),
		OralWaterQty:                ParseNullFloat(c.FormValue("oral_water_qty")),
		OralOther:                   ParseNullInt2(c.FormValue("oral_other")),
		OralOtherQty:                ParseNullFloat(c.FormValue("oral_other_qty")),
		Iv:                          ParseNullInt2(c.FormValue("iv")),
		IvQty:                       ParseNullString2(c.FormValue("iv_qty")),
		IvUsing:                     ParseNullString2(c.FormValue("iv_using")),
		IvAza:                       ParseNullString2(c.FormValue("iv_aza")),
		AccessType:                  ParseNullInt2(c.FormValue("access_type")),
		BloodTrans:                  ParseNullInt2(c.FormValue("blood_trans")),
		OxygenTherapy:               ParseNullInt2(c.FormValue("oxygen_therapy")),
		OxygenQty:                   ParseNullFloat(c.FormValue("oxygen_qty")),
		OxygenWith:                  ParseNullString2(c.FormValue("oxygen_with")),
		Vasopressors:                ParseNullInt2(c.FormValue("vasopressors")),
		Renal:                       ParseNullInt2(c.FormValue("renal")),
		Invasive:                    ParseNullInt2(c.FormValue("invasive")),
		EbolaExperimentalIfZmap:     ParseNullInt2(c.FormValue("ebola_experimental_if_zmap")),
		EbolaExperimentalIfRemd:     ParseNullInt2(c.FormValue("ebola_experimental_if_remd")),
		EbolaExperimentalIfRegn:     ParseNullInt2(c.FormValue("ebola_experimental_if_regn")),
		EbolaExperimentalIfFavi:     ParseNullInt2(c.FormValue("ebola_experimental_if_favi")),
		EbolaExperimentalIfMab:      ParseNullInt2(c.FormValue("ebola_experimental_if_mab")),
		OralOtherAza:                ParseNullString2(c.FormValue("oral_other_aza")),
		AntibacterialAza:            ParseNullInt2(c.FormValue("antibacterial_aza")),
		AntimalarialArtesunateDose:  ParseNullString2(c.FormValue("antimalarial_artesunate_dose")),
		AntimalarialArtesunateRoute: ParseNullString2(c.FormValue("antimalarial_artesunate_route")),
		AntimalarialArtesunateFreq:  ParseNullString2(c.FormValue("antimalarial_artesunate_freq")),
		AntimalarialArthemeterDose:  ParseNullString2(c.FormValue("antimalarial_arthemeter_dose")),
		AntimalarialArthemeterRoute: ParseNullString2(c.FormValue("antimalarial_arthemeter_route")),
		AntimalarialArthemeterFreq:  ParseNullString2(c.FormValue("antimalarial_arthemeter_freq")),
		AntimalarialAlDose:          ParseNullString2(c.FormValue("antimalarial_al_dose")),
		AntimalarialAlRoute:         ParseNullString2(c.FormValue("antimalarial_al_route")),
		AntimalarialAlFreq:          ParseNullString2(c.FormValue("antimalarial_al_freq")),
		AntimalarialAaDose:          ParseNullString2(c.FormValue("antimalarial_aa_dose")),
		AntimalarialAaRoute:         ParseNullString2(c.FormValue("antimalarial_aa_route")),
		AntimalarialAaFreq:          ParseNullString2(c.FormValue("antimalarial_aa_freq")),
	}

	if treat_id == 0 {
		return treat.Insert(c.Context(), db)
	} else {
		treat.SetAsExists()
		return treat.Update(c.Context(), db)
	}
}

func HandlerAPIGetEncounter(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	// Get ID from the query parameter

	id := c.Query("id")

	if id == "" {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "",
		})
	}

	encounter_id, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "",
		})
	}

	var clinical = &models.Clinical{}
	var vital = &models.Vital{}

	clinical, _ = models.ClinicalByEncounterID(c.Context(), db, encounter_id)
	vital, _ = models.VitalByEncounterID(c.Context(), db, encounter_id)

	rtnStr := ` Vitals<br />
				<table class="full-width" border="1">
					<tr>
						<td>Weight: ` + fmt.Sprintf("%.2f", vital.Weight.Float64) + `</td>
						<td>Height: ` + fmt.Sprintf("%.2f", vital.Height.Float64) + `</td>
					</tr>
				</table>
				Symptomms<br/>
				<table class="full-width" border="1">
					<tr>
						<td valign="top">
							Fever: ` + strconv.Itoa(int(clinical.Fever.Int64)) + `<br/>
							Fatigue:<br/>
							Weakness:<br/>
							Malaise:<br/>
							Myalgia:<br/>
							Anorexia:<br/>
							Sore throat
						</td>
						<td valign="top">
							Headache:<br/> 
							Nausea:<br/> 
							Chest pain:<br/> 
							Joint Pain:<br/> 
							Hiccups:<br/>
							Cough:<br/>
						</td>
						<td valign="top">
							Chest pain:<br/>
							Difficulty breathing:<br/>
							Difficulty swallowing:<br/> 
							Abdominal pain:<br/> 
							Diarrhoea:<br/>
							Vomiting:<br/>
							Irritability / Confusion:<br/> 
						</td>
					</tr>
				</table>

				<br/>
				Signs<br/>
				<table class="full-width" border="1">
					<tr>
						<td valign="top">
							Pharyngeal erythema:<br/>  
							Pharyngeal exudate:<br/>  
							Conjunctival injection/bleeding:<br/>  
							Oedema of face/neck:<br/> 
							Tender abdomen:<br/> 
							Sunken eyes or fontanelle:<br/>  
							Tenting on skin pinch:<br/>  
							Palpable liver:<br/> 
							Palpable spleen Rash:<br/> 
							Jaundice:<br/> 

						</td>
						<td valign="top">
							Enlarged lymph nodes:<br/>
							Lower extremity oedema :<br/> 
							Bleeding:<br/> 
						</td>
					</tr>
				</table>
				<br/>
				Specimen <br/>
				<table class="full-width" border="1">
					<tr>
						<td valign="top">
						</td>
					</tr>
				</table>
				<br/>
				Lab Results <br/>
				<table class="full-width" border="1">
					<tr>
						<td valign="top">
						</td>
					</tr>
				</table>
				<br/>
				Medications <br/>
				<table class="full-width" border="1">
					<tr>
						<td valign="top">
						</td>
					</tr>
				</table>`

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": rtnStr,
	})

}

func HandlerAPIGetStatuses(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	userID := GetCurrentUser(c, store)

	// Check if user is logged in
	if userID == 0 {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	ClientID := c.Query("client_id")
	if ClientID == "" {
		ClientID = "0"
	}

	statuses, er := models.Statusez(c.Context(), db, " client_id = "+ClientID)
	if er != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error fetching statuses",
		})
	}

	return c.JSON(statuses)

}

func HandlerAPIPostStatus(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {

	//=================

	userID := GetCurrentUser(c, store)
	// Check if user is logged in
	if userID == 0 {
		fmt.Println("unauthorized")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	//=============================

	var formData map[string]interface{}

	if err := c.BodyParser(&formData); err != nil {
		fmt.Println("JSON parsing failed:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	var s models.Status

	s.StatusID = int(ParseNullInt2(formData["status_id"]).Int64)
	s.ClientID = ParseNullInt2(formData["client_id"])
	s.StatusDate = ParseNullString2(formData["status_date"])
	s.Status = ParseNullString2(formData["status"])
	s.StatusNotes = ParseNullString2(formData["status_notes"])

	s.UpdatedBy.Valid = true
	s.UpdatedBy.Int64 = int64(userID)

	s.UpdatedOn.Valid = true
	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02")
	s.UpdatedOn.String = formattedTime

	// Check if it's a new record (StatusID == 0)
	if s.StatusID > 0 {
		s.SetAsExists()
		err := s.Update(c.Context(), db)
		if err != nil {
			fmt.Println("update fail:", err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	} else {

		err := s.Insert(c.Context(), db)
		if err != nil {
			fmt.Println("insert fail:", err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "success",
	})
}
