package reports

import (
	"case/internal/handlers"
	"case/internal/security"
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// ReportFilters represents the filter criteria for reports
type ReportFilters struct {
	StartDate     string
	EndDate       string
	OutbreakID    int
	FacilityID    int
	DistrictID    int
	TreatmentSite string
	Symptoms      string
	Treatment     string
	PatientType   string // case, suspect, etc.
	Outcome       string // recovered, died, in_treatment
	ReportType    string // vhf, measles, polio, mpox, general
}

// ReportData represents the data structure for reports
type ReportData struct {
	Filters      ReportFilters
	Summary      map[string]interface{}
	Charts       map[string]interface{}
	Tables       map[string]interface{}
	UserRole     string
	UserFacility int
	UserDistrict int
	AccessLevel  string // full, facility, district
	Error        string // For error reporting
}

// ReportsHome handles the main reports dashboard
func ReportsHome(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) error {
	userID, userName := handlers.GetUser(c, sl, store)
	if userID == 0 {
		return c.Redirect("/login")
	}

	// Check if user has reports access using proper database query
	allowedRoles := []string{"admin", "reports", "vhf_lab_technician", "case_manager", "data_analyst", "cases_viewer", "cif_viewer"}
	if !security.HasAnyRole(db, userID, allowedRoles) {
		return c.Status(403).SendString("Access denied: Reports permission required")
	}

	// Get user roles for access level determination
	userRoles := security.GetUserRoles(db, userID)
	userRole := strings.Join(userRoles, ",") // Join multiple roles with comma
	userFacility := handlers.GetCurrentFacility(c, db, sl, store)
	userDistrict := getUserDistrict(c, db, userID)

	// Get or restore filter state from session
	sess, _ := store.Get(c)
	filterState := getFilterStateFromSession(sess)

	// If no filter state in session, initialize with defaults
	if filterState == nil {
		filterState = &ReportFilters{
			StartDate:  time.Now().AddDate(0, -1, 0).Format("2006-01-02"), // Last 30 days
			EndDate:    time.Now().Format("2006-01-02"),
			ReportType: "general",
		}
	}

	// Get accessible data based on user permissions
	outbreaks, facilities, districts := getAccessibleData(c, db, userID, userRole, userFacility, userDistrict)

	// Check if user has CIF-only access
	hasCIFOnlyAccess := security.HasRole(db, userID, "cif_viewer") && !security.HasAnyRole(db, userID, []string{"admin", "reports", "data_analyst"})

	data := handlers.NewTemplateData(c, store)
	data.Form = map[string]interface{}{
		"Title":         "Reports Dashboard",
		"UserID":        userName,
		"Outbreaks":     outbreaks,
		"Facilities":    facilities,
		"Districts":     districts,
		"AccessLevel":   getAccessLevel(userRole, userFacility, userDistrict),
		"Filters":       filterState, // Add current filter state
		"CIFOnlyAccess": hasCIFOnlyAccess,
	}

	return handlers.GenerateHTML(c, db, data, "reports_dashboard")
}

// GenerateReport handles report generation with filter persistence
func GenerateReport(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) error {
	userID, _ := handlers.GetUser(c, sl, store)
	if userID == 0 {
		return c.Redirect("/login")
	}

	// Log the form data for debugging
	sl.Info("Report generation request",
		"method", c.Method(),
		"content_type", c.Get("Content-Type"),
		"form_values", c.FormValue("report_type"),
		"query_values", c.Query("report_type"),
	)

	// Parse filters from form
	filters := parseReportFilters(c)

	// Log the parsed filters
	sl.Info("Parsed report filters",
		"report_type", filters.ReportType,
		"start_date", filters.StartDate,
		"end_date", filters.EndDate,
		"outbreak_id", filters.OutbreakID,
	)

	// Store filter state in session for persistence
	sess, _ := store.Get(c)
	saveFilterStateToSession(sess, filters)
	sess.Save()

	// Get user roles for access level determination
	userRoles := security.GetUserRoles(db, userID)
	userRole := strings.Join(userRoles, ",")
	userFacility := handlers.GetCurrentFacility(c, db, sl, store)
	userDistrict := getUserDistrict(c, db, userID)

	// Apply access restrictions
	filters = applyAccessRestrictions(filters, userRole, userFacility, userDistrict)

	// Generate report based on type
	var reportData ReportData
	switch filters.ReportType {
	case "vhf":
		reportData = generateVHFReport(c, db, filters)
	case "measles":
		reportData = generateMeaslesReport(c, db, filters)
	case "polio":
		reportData = generatePolioReport(c, db, filters)
	case "mpox":
		reportData = generateMpoxReport(c, db, filters)
	case "indicators":
		reportData = generateIndicatorsReport(c, db, filters)
	case "general":
		reportData = generateGeneralReport(c, db, filters)
	default:
		reportData = generateGeneralReport(c, db, filters)
	}

	// Get accessible data for filter dropdowns
	outbreaks, facilities, districts := getAccessibleData(c, db, userID, userRole, userFacility, userDistrict)

	data := handlers.NewTemplateData(c, store)
	data.Form = map[string]interface{}{
		"Title":       "Report Results",
		"UserID":      userID,
		"Outbreaks":   outbreaks,
		"Facilities":  facilities,
		"Districts":   districts,
		"AccessLevel": getAccessLevel(userRole, userFacility, userDistrict),
		"Filters":     filters,
		"Summary":     reportData.Summary,
		"Tables":      reportData.Tables,
		"Error":       reportData.Error,
	}

	return renderReportView(c, db, data)
}

// CIFReports handles CIF-specific reports
func CIFReports(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) error {
	userID, userName := handlers.GetUser(c, sl, store)
	if userID == 0 {
		return c.Redirect("/login", 302)
	}

	// Get user roles for access level determination
	userRoles := security.GetUserRoles(db, userID)
	userRole := strings.Join(userRoles, ",") // Join multiple roles with comma
	userFacility := handlers.GetCurrentFacility(c, db, sl, store)
	userDistrict := getUserDistrict(c, db, userID)

	// Special access for lab technicians
	if strings.Contains(userRole, "vhf_lab_technician") {
		return generateLabTechnicianReport(c, db, userDistrict)
	}

	// Regular CIF reports
	cifType := c.Params("type") // vhf, measles, polio, mpox
	filters := parseReportFilters(c)
	filters.ReportType = cifType

	reportData := generateCIFReport(c, db, filters, cifType)
	reportData.UserRole = userRole
	reportData.UserFacility = userFacility
	reportData.UserDistrict = userDistrict

	data := handlers.NewTemplateData(c, store)
	data.User = userName
	data.Form = reportData

	return handlers.GenerateHTML(c, db, data, "cif_report")
}

// Helper functions

func getUserDistrict(c *fiber.Ctx, db *sql.DB, userID int) int {
	var districtID int
	query := `SELECT district_id FROM employee WHERE employee_id = (SELECT user_employee FROM users WHERE user_id = $1)`
	err := db.QueryRow(query, userID).Scan(&districtID)
	if err != nil {
		return 0
	}
	return districtID
}

func getAccessLevel(role string, facility, district int) string {
	if strings.Contains(role, "admin") {
		return "full"
	}
	if facility > 0 {
		return "facility"
	}
	if district > 0 {
		return "district"
	}
	return "limited"
}

func getAccessibleData(c *fiber.Ctx, db *sql.DB, userID int, role string, facility, district int) ([]map[string]interface{}, []map[string]interface{}, []map[string]interface{}) {
	var outbreaks, facilities, districts []map[string]interface{}

	// Get outbreaks - using correct column names: id, name
	outbreakQuery := `SELECT id, name FROM outbreaks WHERE status = 'active'`
	rows, err := db.Query(outbreakQuery)
	if err != nil {
		// Log the error but don't fail completely
		fmt.Printf("Error querying outbreaks: %v\n", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var id int
			var name string
			if rows.Scan(&id, &name) == nil {
				outbreaks = append(outbreaks, map[string]interface{}{"id": id, "name": name})
			}
		}
	}

	// Get facilities - using facility table with correct column names: facility_id, facility_name
	facilityQuery := `SELECT facility_id, facility_name FROM facility WHERE is_active = true`
	var facilityArgs []interface{}
	if district > 0 {
		facilityQuery += ` AND district = $1`
		facilityArgs = append(facilityArgs, district)
	}
	rows, err = db.Query(facilityQuery, facilityArgs...)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id int
			var name string
			if rows.Scan(&id, &name) == nil {
				facilities = append(facilities, map[string]interface{}{"id": id, "name": name})
			}
		}
	}

	// Get districts - using correct column names: id, name
	districtQuery := `SELECT id, name FROM districts WHERE 1=1`
	rows, err = db.Query(districtQuery)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id int
			var name string
			if rows.Scan(&id, &name) == nil {
				districts = append(districts, map[string]interface{}{"id": id, "name": name})
			}
		}
	}

	return outbreaks, facilities, districts
}

// getCIFFacilities returns facilities from afi_facilities table for CIF forms
func getCIFFacilities(c *fiber.Ctx, db *sql.DB, district int) []map[string]interface{} {
	var facilities []map[string]interface{}

	// Get CIF facilities from afi_facilities table - using id column
	facilityQuery := `SELECT id, name FROM afi_facilities WHERE 1=1`
	var facilityArgs []interface{}
	if district > 0 {
		// Note: afi_facilities might not have district_id column, so we'll skip district filtering for now
		// facilityQuery += ` AND district_id = $1`
		// facilityArgs = append(facilityArgs, district)
	}
	rows, err := db.Query(facilityQuery, facilityArgs...)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id int
			var name string
			if rows.Scan(&id, &name) == nil {
				facilities = append(facilities, map[string]interface{}{"id": id, "name": name})
			}
		}
	}

	return facilities
}

func parseReportFilters(c *fiber.Ctx) ReportFilters {
	// Try form values first, then query parameters
	outbreakIDStr := c.FormValue("outbreak_id")
	if outbreakIDStr == "" {
		outbreakIDStr = c.Query("outbreak_id")
	}

	facilityIDStr := c.FormValue("facility_id")
	if facilityIDStr == "" {
		facilityIDStr = c.Query("facility_id")
	}

	districtIDStr := c.FormValue("district_id")
	if districtIDStr == "" {
		districtIDStr = c.Query("district_id")
	}

	startDate := c.FormValue("start_date")
	if startDate == "" {
		startDate = c.Query("start_date")
	}

	endDate := c.FormValue("end_date")
	if endDate == "" {
		endDate = c.Query("end_date")
	}

	reportType := c.FormValue("report_type")
	if reportType == "" {
		reportType = c.Query("report_type")
	}

	patientType := c.FormValue("patient_type")
	if patientType == "" {
		patientType = c.Query("patient_type")
	}

	outcome := c.FormValue("outcome")
	if outcome == "" {
		outcome = c.Query("outcome")
	}

	outbreakID, _ := strconv.Atoi(outbreakIDStr)
	facilityID, _ := strconv.Atoi(facilityIDStr)
	districtID, _ := strconv.Atoi(districtIDStr)

	return ReportFilters{
		StartDate:     startDate,
		EndDate:       endDate,
		OutbreakID:    outbreakID,
		FacilityID:    facilityID,
		DistrictID:    districtID,
		TreatmentSite: c.FormValue("treatment_site"),
		Symptoms:      c.FormValue("symptoms"),
		Treatment:     c.FormValue("treatment"),
		PatientType:   patientType,
		Outcome:       outcome,
		ReportType:    reportType,
	}
}

func applyAccessRestrictions(filters ReportFilters, role string, facility, district int) ReportFilters {
	if strings.Contains(role, "admin") {
		return filters // Full access
	}
	if facility > 0 {
		filters.FacilityID = facility
	}
	if district > 0 {
		filters.DistrictID = district
	}
	return filters
}

// Report generation functions with real data
func generateVHFReport(c *fiber.Ctx, db *sql.DB, filters ReportFilters) ReportData {
	// Get VHF cases from VHF CIF table
	query := `
		SELECT 
			v.id,
			v.case_code,
			v.surname,
			v.other_names,
			v.age_years,
			v.age_months,
			v.gender,
			v.date_of_birth,
			v.status,
			v.created_at,
			v.outbreak_id,
			f.facility_name,
			d.name as district_name,
			v.district
		FROM vhf_patients v
		LEFT JOIN afi_facilities f ON v.reporting_health_facility_name = f.facility_name
		
		WHERE v.outbreak_id IN (SELECT id FROM outbreaks WHERE outbreak = 'vhf')
	`

	var args []interface{}
	argCount := 1

	if filters.StartDate != "" {
		query += fmt.Sprintf(" AND v.created_at >= $%d", argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		query += fmt.Sprintf(" AND v.created_at <= $%d", argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	if filters.OutbreakID > 0 {
		query += fmt.Sprintf(" AND v.outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	if filters.FacilityID > 0 {
		query += fmt.Sprintf(" AND f.id = $%d", argCount)
		args = append(args, filters.FacilityID)
		argCount++
	}

	query += " ORDER BY v.created_at DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return ReportData{Filters: filters, Error: "Database error: " + err.Error()}
	}
	defer rows.Close()

	var cases []map[string]interface{}
	for rows.Next() {
		var id int
		var caseCode, surname, otherNames, gender, dateOfBirth, status sql.NullString
		var age sql.NullFloat64
		var outbreakID sql.NullInt64
		var facilityName, districtName sql.NullString
		var district sql.NullString
		var labStatus sql.NullBool
		err := rows.Scan(&id, &caseCode, &surname, &otherNames, &age, &gender, &dateOfBirth, &status, &outbreakID, &facilityName, &districtName, &district, &labStatus)
		if err != nil {
			continue
		}

		cases = append(cases, map[string]interface{}{
			"id":             id,
			"case_code":      caseCode.String,
			"name":           fmt.Sprintf("%s %s", surname.String, otherNames.String),
			"age":            age.Float64,
			"gender":         gender.String,
			"date_of_birth":  dateOfBirth.String,
			"status":         status.String,
			"outbreak_id":    outbreakID.Int64,
			"facility":       facilityName.String,
			"district":       districtName.String,
			"district_other": district.String,
			"lab_status":     labStatus.Bool,
		})
	}

	return ReportData{
		Filters: filters,
		Summary: map[string]interface{}{
			"total_cases":  len(cases),
			"active_cases": countByStatus(cases, "active"),
			"recovered":    countByStatus(cases, "recovered"),
			"died":         countByStatus(cases, "died"),
		},
		Tables: map[string]interface{}{
			"cases": cases,
		},
	}
}

func generateMeaslesReport(c *fiber.Ctx, db *sql.DB, filters ReportFilters) ReportData {
	// Get Measles cases from Measles CIF table
	query := `
		SELECT 
			m.id,
			m.case_code,
			m.first_name,
			m.last_name,
			m.age,
			m.gender,
			m.date_of_onset,
			m.case_classification,
			m.outbreak_id,
			f.facility_name,
			d.name as district_name
		FROM measles_case_investigation_form m
		LEFT JOIN afi_facilities f ON m.reporting_health_facility_name = f.name
		LEFT JOIN districts d ON f.district = d.name
		WHERE m.outbreak_id IN (SELECT id FROM outbreaks WHERE outbreak_type = 'measles')
	`

	var args []interface{}
	argCount := 1

	if filters.StartDate != "" {
		query += fmt.Sprintf(" AND m.date_of_onset >= $%d", argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		query += fmt.Sprintf(" AND m.date_of_onset <= $%d", argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	if filters.OutbreakID > 0 {
		query += fmt.Sprintf(" AND m.outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	if filters.FacilityID > 0 {
		query += fmt.Sprintf(" AND f.id = $%d", argCount)
		args = append(args, filters.FacilityID)
		argCount++
	}

	query += " ORDER BY m.date_of_onset DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return ReportData{Filters: filters, Error: "Database error: " + err.Error()}
	}
	defer rows.Close()

	var cases []map[string]interface{}
	for rows.Next() {
		var id int
		var caseCode, firstName, lastName, gender, dateOfOnset, caseClassification sql.NullString
		var age sql.NullFloat64
		var outbreakID sql.NullInt64
		var facilityName, districtName sql.NullString

		err := rows.Scan(&id, &caseCode, &firstName, &lastName, &age, &gender, &dateOfOnset, &caseClassification, &outbreakID, &facilityName, &districtName)
		if err != nil {
			continue
		}

		cases = append(cases, map[string]interface{}{
			"id":                  id,
			"case_code":           caseCode.String,
			"name":                fmt.Sprintf("%s %s", firstName.String, lastName.String),
			"age":                 age.Float64,
			"gender":              gender.String,
			"date_of_onset":       dateOfOnset.String,
			"case_classification": caseClassification.String,
			"outbreak_id":         outbreakID.Int64,
			"facility":            facilityName.String,
			"district":            districtName.String,
		})
	}

	return ReportData{
		Filters: filters,
		Summary: map[string]interface{}{
			"total_cases":     len(cases),
			"confirmed_cases": countByClassification(cases, "Confirmed"),
			"suspect_cases":   countByClassification(cases, "Suspect"),
			"probable_cases":  countByClassification(cases, "Probable"),
		},
		Tables: map[string]interface{}{
			"cases": cases,
		},
	}
}

func generatePolioReport(c *fiber.Ctx, db *sql.DB, filters ReportFilters) ReportData {
	// Get Polio cases from Polio CIF table
	query := `
		SELECT 
			p.id,
			p.case_code,
			p.first_name,
			p.last_name,
			p.age,
			p.gender,
			p.date_of_onset,
			p.case_classification,
			p.outbreak_id,
			f.facility_name,
			d.name as district_name
		FROM polio_case_investigation_form p
		LEFT JOIN afi_facilities f ON p.reporting_health_facility_name = f.name
		LEFT JOIN districts d ON f.district = d.name
		WHERE p.outbreak_id IN (SELECT id FROM outbreaks WHERE outbreak_type = 'polio')
	`

	var args []interface{}
	argCount := 1

	if filters.StartDate != "" {
		query += fmt.Sprintf(" AND p.date_of_onset >= $%d", argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		query += fmt.Sprintf(" AND p.date_of_onset <= $%d", argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	if filters.OutbreakID > 0 {
		query += fmt.Sprintf(" AND p.outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	if filters.FacilityID > 0 {
		query += fmt.Sprintf(" AND f.id = $%d", argCount)
		args = append(args, filters.FacilityID)
		argCount++
	}

	query += " ORDER BY p.date_of_onset DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return ReportData{Filters: filters, Error: "Database error: " + err.Error()}
	}
	defer rows.Close()

	var cases []map[string]interface{}
	for rows.Next() {
		var id int
		var caseCode, firstName, lastName, gender, dateOfOnset, caseClassification sql.NullString
		var age sql.NullFloat64
		var outbreakID sql.NullInt64
		var facilityName, districtName sql.NullString

		err := rows.Scan(&id, &caseCode, &firstName, &lastName, &age, &gender, &dateOfOnset, &caseClassification, &outbreakID, &facilityName, &districtName)
		if err != nil {
			continue
		}

		cases = append(cases, map[string]interface{}{
			"id":                  id,
			"case_code":           caseCode.String,
			"name":                fmt.Sprintf("%s %s", firstName.String, lastName.String),
			"age":                 age.Float64,
			"gender":              gender.String,
			"date_of_onset":       dateOfOnset.String,
			"case_classification": caseClassification.String,
			"outbreak_id":         outbreakID.Int64,
			"facility":            facilityName.String,
			"district":            districtName.String,
		})
	}

	return ReportData{
		Filters: filters,
		Summary: map[string]interface{}{
			"total_cases":     len(cases),
			"confirmed_cases": countByClassification(cases, "Confirmed"),
			"suspect_cases":   countByClassification(cases, "Suspect"),
			"probable_cases":  countByClassification(cases, "Probable"),
		},
		Tables: map[string]interface{}{
			"cases": cases,
		},
	}
}

func generateMpoxReport(c *fiber.Ctx, db *sql.DB, filters ReportFilters) ReportData {
	// Get Mpox cases from Mpox CIF table
	query := `
		SELECT 
			m.id,
			m.case_code,
			m.first_name,
			m.last_name,
			m.age,
			m.gender,
			m.date_of_onset,
			m.case_classification,
			m.outbreak_id,
			f.facility_name,
			d.name as district_name
		FROM mpox_case_investigation_form m
		LEFT JOIN afi_facilities f ON m.reporting_health_facility_name = f.name
		LEFT JOIN districts d ON f.district = d.name
		WHERE m.outbreak_id IN (SELECT id FROM outbreaks WHERE outbreak_type = 'mpox')
	`

	var args []interface{}
	argCount := 1

	if filters.StartDate != "" {
		query += fmt.Sprintf(" AND m.date_of_onset >= $%d", argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		query += fmt.Sprintf(" AND m.date_of_onset <= $%d", argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	if filters.OutbreakID > 0 {
		query += fmt.Sprintf(" AND m.outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	if filters.FacilityID > 0 {
		query += fmt.Sprintf(" AND f.id = $%d", argCount)
		args = append(args, filters.FacilityID)
		argCount++
	}

	query += " ORDER BY m.date_of_onset DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return ReportData{Filters: filters, Error: "Database error: " + err.Error()}
	}
	defer rows.Close()

	var cases []map[string]interface{}
	for rows.Next() {
		var id int
		var caseCode, firstName, lastName, gender, dateOfOnset, caseClassification sql.NullString
		var age sql.NullFloat64
		var outbreakID sql.NullInt64
		var facilityName, districtName sql.NullString

		err := rows.Scan(&id, &caseCode, &firstName, &lastName, &age, &gender, &dateOfOnset, &caseClassification, &outbreakID, &facilityName, &districtName)
		if err != nil {
			continue
		}

		cases = append(cases, map[string]interface{}{
			"id":                  id,
			"case_code":           caseCode.String,
			"name":                fmt.Sprintf("%s %s", firstName.String, lastName.String),
			"age":                 age.Float64,
			"gender":              gender.String,
			"date_of_onset":       dateOfOnset.String,
			"case_classification": caseClassification.String,
			"outbreak_id":         outbreakID.Int64,
			"facility":            facilityName.String,
			"district":            districtName.String,
		})
	}

	return ReportData{
		Filters: filters,
		Summary: map[string]interface{}{
			"total_cases":     len(cases),
			"confirmed_cases": countByClassification(cases, "Confirmed"),
			"suspect_cases":   countByClassification(cases, "Suspect"),
			"probable_cases":  countByClassification(cases, "Probable"),
		},
		Tables: map[string]interface{}{
			"cases": cases,
		},
	}
}

func generateGeneralReport(c *fiber.Ctx, db *sql.DB, filters ReportFilters) ReportData {
	// Get all cases with real data
	query := `
		SELECT 
			c.id,
			c.firstname,
			c.lastname,
			c.age,
			c.gender,
			c.adm_date,
			c.status,
			c.outbreak_id,
			o.name as outbreak_name,
			f.facility_name,
			d.name as district_name
		FROM clients c
		LEFT JOIN outbreaks o ON c.outbreak_id = o.id
		LEFT JOIN facility f ON c.site = f.facility_id
		LEFT JOIN districts d ON f.district = d.name
		WHERE 1=1
	`

	var args []interface{}
	argCount := 1

	if filters.StartDate != "" {
		query += fmt.Sprintf(" AND c.adm_date >= $%d", argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		query += fmt.Sprintf(" AND c.adm_date <= $%d", argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	if filters.OutbreakID > 0 {
		query += fmt.Sprintf(" AND c.outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	if filters.FacilityID > 0 {
		query += fmt.Sprintf(" AND c.site = $%d", argCount)
		args = append(args, filters.FacilityID)
		argCount++
	}

	if filters.DistrictID > 0 {
		query += fmt.Sprintf(" AND d.id = $%d", argCount)
		args = append(args, filters.DistrictID)
		argCount++
	}

	query += " ORDER BY c.adm_date DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return ReportData{Filters: filters, Error: "Database error: " + err.Error()}
	}
	defer rows.Close()

	var cases []map[string]interface{}
	for rows.Next() {
		var id int
		var firstName, lastName, gender, admDate, status sql.NullString
		var age sql.NullFloat64
		var outbreakID sql.NullInt64
		var outbreakName, facilityName, districtName sql.NullString

		err := rows.Scan(&id, &firstName, &lastName, &age, &gender, &admDate, &status, &outbreakID, &outbreakName, &facilityName, &districtName)
		if err != nil {
			continue
		}

		cases = append(cases, map[string]interface{}{
			"id":            id,
			"name":          fmt.Sprintf("%s %s", firstName.String, lastName.String),
			"age":           age.Float64,
			"gender":        gender.String,
			"adm_date":      admDate.String,
			"status":        status.String,
			"outbreak_id":   outbreakID.Int64,
			"outbreak_name": outbreakName.String,
			"facility":      facilityName.String,
			"district":      districtName.String,
		})
	}

	return ReportData{
		Filters: filters,
		Summary: map[string]interface{}{
			"total_cases":  len(cases),
			"active_cases": countByStatus(cases, "active"),
			"recovered":    countByStatus(cases, "recovered"),
			"died":         countByStatus(cases, "died"),
		},
		Tables: map[string]interface{}{
			"cases": cases,
		},
	}
}

// generateIndicatorsReport generates comprehensive health indicators report
func generateIndicatorsReport(c *fiber.Ctx, db *sql.DB, filters ReportFilters) ReportData {
	indicators := make(map[string]interface{})

	// 1. New admissions (daily)
	newAdmissions, err := GetNewAdmissionsDaily(db, filters)
	if err == nil {
		indicators["new_admissions_daily"] = newAdmissions
	}

	// 2. Cumulative confirmed cases
	cumulativeConfirmed, err := GetCumulativeConfirmedCases(db, filters)
	if err == nil {
		indicators["cumulative_confirmed"] = cumulativeConfirmed
	}

	// 3. Cumulative suspected cases
	cumulativeSuspected, err := GetCumulativeSuspectedCases(db, filters)
	if err == nil {
		indicators["cumulative_suspected"] = cumulativeSuspected
	}

	// 4. Cumulative deaths
	cumulativeDeaths, err := GetCumulativeDeaths(db, filters)
	if err == nil {
		indicators["cumulative_deaths"] = cumulativeDeaths
	}

	// 5. Case Fatality Rate (CFR)
	cfr, err := GetCaseFatalityRate(db, filters)
	if err == nil {
		indicators["case_fatality_rate"] = cfr
	}

	// 6. Current admissions (active cases in hospital)
	currentAdmissions, err := GetCurrentAdmissions(db, filters)
	if err == nil {
		indicators["current_admissions"] = currentAdmissions
	}

	// 7. Discharges (cumulative)
	cumulativeDischarges, err := GetCumulativeDischarges(db, filters)
	if err == nil {
		indicators["cumulative_discharges"] = cumulativeDischarges
	}

	// 8. Severe cases admitted
	severeCases, err := GetSevereCasesAdmitted(db, filters)
	if err == nil {
		indicators["severe_cases"] = severeCases
	}

	// 9. Critical cases admitted
	criticalCases, err := GetCriticalCasesAdmitted(db, filters)
	if err == nil {
		indicators["critical_cases"] = criticalCases
	}

	// 10. Cases by sex
	casesBySex, err := GetCasesBySex(db, filters)
	if err == nil {
		indicators["cases_by_sex"] = casesBySex
	}

	// 11. Cases by age group
	casesByAge, err := GetCasesByAgeGroup(db, filters)
	if err == nil {
		indicators["cases_by_age"] = casesByAge
	}

	// 12. Cases by district/facility
	casesByLocation, err := GetCasesByLocation(db, filters)
	if err == nil {
		indicators["cases_by_location"] = casesByLocation
	}

	// 13. Healthcare worker (HCW) infections
	hcwInfections, err := GetHCWInfections(db, filters)
	if err == nil {
		indicators["hcw_infections"] = hcwInfections
	}

	return ReportData{
		Filters: filters,
		Summary: indicators,
		Tables: map[string]interface{}{
			"indicators": indicators,
		},
	}
}

// GetNewAdmissionsDaily returns daily new admissions count
func GetNewAdmissionsDaily(db *sql.DB, filters ReportFilters) (map[string]interface{}, error) {
	query := `
		SELECT 
			DATE(c.adm_date) as admission_date,
			COUNT(*) as daily_admissions
		FROM clients c
		WHERE c.adm_date IS NOT NULL
	`

	var args []interface{}
	argCount := 1

	if filters.StartDate != "" {
		query += fmt.Sprintf(" AND c.adm_date >= $%d", argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		query += fmt.Sprintf(" AND c.adm_date <= $%d", argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	if filters.OutbreakID > 0 {
		query += fmt.Sprintf(" AND c.outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	query += " GROUP BY DATE(c.adm_date) ORDER BY admission_date"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dates []string
	var counts []int

	for rows.Next() {
		var date sql.NullString
		var count int

		err := rows.Scan(&date, &count)
		if err != nil {
			continue
		}

		dates = append(dates, date.String)
		counts = append(counts, count)
	}

	return map[string]interface{}{
		"dates":  dates,
		"counts": counts,
		"total":  len(counts),
	}, nil
}

// GetCumulativeConfirmedCases returns cumulative confirmed cases
func GetCumulativeConfirmedCases(db *sql.DB, filters ReportFilters) (map[string]interface{}, error) {
	query := `
		SELECT COUNT(*) as confirmed_count
		FROM clients c
		LEFT JOIN discharge d ON c.id = d.client_id
		WHERE c.status = 'confirmed' OR d.final_diagnosis = 'Mpox Positive'
	`

	var args []interface{}
	argCount := 1

	if filters.StartDate != "" {
		query += fmt.Sprintf(" AND c.adm_date >= $%d", argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		query += fmt.Sprintf(" AND c.adm_date <= $%d", argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	if filters.OutbreakID > 0 {
		query += fmt.Sprintf(" AND c.outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	var count int
	err := db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"confirmed_cases": count,
	}, nil
}

// GetCumulativeSuspectedCases returns cumulative suspected cases
func GetCumulativeSuspectedCases(db *sql.DB, filters ReportFilters) (map[string]interface{}, error) {
	query := `
		SELECT COUNT(*) as suspected_count
		FROM clients c
		WHERE c.status = 'suspected' OR c.status = 'probable'
	`

	var args []interface{}
	argCount := 1

	if filters.StartDate != "" {
		query += fmt.Sprintf(" AND c.adm_date >= $%d", argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		query += fmt.Sprintf(" AND c.adm_date <= $%d", argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	if filters.OutbreakID > 0 {
		query += fmt.Sprintf(" AND c.outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	var count int
	err := db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"suspected_cases": count,
	}, nil
}

// GetCumulativeDeaths returns cumulative deaths
func GetCumulativeDeaths(db *sql.DB, filters ReportFilters) (map[string]interface{}, error) {
	query := `
		SELECT COUNT(*) as death_count
		FROM discharge d
		LEFT JOIN clients c ON d.client_id = c.id
		WHERE d.discharge_outcome = 'Death'
	`

	var args []interface{}
	argCount := 1

	if filters.StartDate != "" {
		query += fmt.Sprintf(" AND d.discharge_date >= $%d", argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		query += fmt.Sprintf(" AND d.discharge_date <= $%d", argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	if filters.OutbreakID > 0 {
		query += fmt.Sprintf(" AND c.outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	var count int
	err := db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"deaths": count,
	}, nil
}

// GetCaseFatalityRate returns case fatality rate
func GetCaseFatalityRate(db *sql.DB, filters ReportFilters) (map[string]interface{}, error) {
	// Get deaths among confirmed cases
	deathsQuery := `
		SELECT COUNT(*) as death_count
		FROM discharge d
		LEFT JOIN clients c ON d.client_id = c.id
		WHERE d.discharge_outcome = 'Death' 
		AND (c.status = 'confirmed' OR d.final_diagnosis = 'Mpox Positive')
	`

	// Get total confirmed cases
	confirmedQuery := `
		SELECT COUNT(*) as confirmed_count
		FROM clients c
		LEFT JOIN discharge d ON c.id = d.client_id
		WHERE c.status = 'confirmed' OR d.final_diagnosis = 'Mpox Positive'
	`

	var args []interface{}
	argCount := 1

	if filters.StartDate != "" {
		deathsQuery += fmt.Sprintf(" AND d.discharge_date >= $%d", argCount)
		confirmedQuery += fmt.Sprintf(" AND c.adm_date >= $%d", argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		deathsQuery += fmt.Sprintf(" AND d.discharge_date <= $%d", argCount)
		confirmedQuery += fmt.Sprintf(" AND c.adm_date <= $%d", argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	if filters.OutbreakID > 0 {
		deathsQuery += fmt.Sprintf(" AND c.outbreak_id = $%d", argCount)
		confirmedQuery += fmt.Sprintf(" AND c.outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	var deaths int
	var confirmed int

	err := db.QueryRow(deathsQuery, args...).Scan(&deaths)
	if err != nil {
		return nil, err
	}

	err = db.QueryRow(confirmedQuery, args...).Scan(&confirmed)
	if err != nil {
		return nil, err
	}

	var cfr float64
	if confirmed > 0 {
		cfr = float64(deaths) / float64(confirmed) * 100
	}

	return map[string]interface{}{
		"deaths":      deaths,
		"confirmed":   confirmed,
		"cfr":         cfr,
		"cfr_percent": fmt.Sprintf("%.2f%%", cfr),
	}, nil
}

// GetCurrentAdmissions returns current active admissions
func GetCurrentAdmissions(db *sql.DB, filters ReportFilters) (map[string]interface{}, error) {
	query := `
		SELECT COUNT(*) as current_admissions
		FROM clients c
		LEFT JOIN discharge d ON c.id = d.client_id
		WHERE c.adm_date IS NOT NULL 
		AND d.discharge_date IS NULL
		AND d.discharge_outcome != 'Death'
	`

	var args []interface{}
	argCount := 1

	if filters.StartDate != "" {
		query += fmt.Sprintf(" AND c.adm_date >= $%d", argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		query += fmt.Sprintf(" AND c.adm_date <= $%d", argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	if filters.OutbreakID > 0 {
		query += fmt.Sprintf(" AND c.outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	var count int
	err := db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"current_admissions": count,
	}, nil
}

// GetCumulativeDischarges returns cumulative discharges
func GetCumulativeDischarges(db *sql.DB, filters ReportFilters) (map[string]interface{}, error) {
	query := `
		SELECT COUNT(*) as discharge_count
		FROM discharge d
		LEFT JOIN clients c ON d.client_id = c.id
		WHERE d.discharge_outcome = 'Discharged alive'
	`

	var args []interface{}
	argCount := 1

	if filters.StartDate != "" {
		query += fmt.Sprintf(" AND d.discharge_date >= $%d", argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		query += fmt.Sprintf(" AND d.discharge_date <= $%d", argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	if filters.OutbreakID > 0 {
		query += fmt.Sprintf(" AND c.outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	var count int
	err := db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"discharges": count,
	}, nil
}

// GetSevereCasesAdmitted returns severe cases percentage
func GetSevereCasesAdmitted(db *sql.DB, filters ReportFilters) (map[string]interface{}, error) {
	// This would need to be adapted based on your specific severity classification system
	// For now, using a placeholder query
	query := `
		SELECT 
			COUNT(*) as total_admitted,
			COUNT(CASE WHEN c.status = 'severe' THEN 1 END) as severe_cases
		FROM clients c
		WHERE c.adm_date IS NOT NULL
	`

	var args []interface{}
	argCount := 1

	if filters.StartDate != "" {
		query += fmt.Sprintf(" AND c.adm_date >= $%d", argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		query += fmt.Sprintf(" AND c.adm_date <= $%d", argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	if filters.OutbreakID > 0 {
		query += fmt.Sprintf(" AND c.outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	var total, severe int
	err := db.QueryRow(query, args...).Scan(&total, &severe)
	if err != nil {
		return nil, err
	}

	var percentage float64
	if total > 0 {
		percentage = float64(severe) / float64(total) * 100
	}

	return map[string]interface{}{
		"total_admitted": total,
		"severe_cases":   severe,
		"percentage":     percentage,
		"percentage_str": fmt.Sprintf("%.2f%%", percentage),
	}, nil
}

// GetCriticalCasesAdmitted returns critical cases percentage
func GetCriticalCasesAdmitted(db *sql.DB, filters ReportFilters) (map[string]interface{}, error) {
	// This would need to be adapted based on your specific severity classification system
	query := `
		SELECT 
			COUNT(*) as total_admitted,
			COUNT(CASE WHEN c.status = 'critical' THEN 1 END) as critical_cases
		FROM clients c
		WHERE c.adm_date IS NOT NULL
	`

	var args []interface{}
	argCount := 1

	if filters.StartDate != "" {
		query += fmt.Sprintf(" AND c.adm_date >= $%d", argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		query += fmt.Sprintf(" AND c.adm_date <= $%d", argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	if filters.OutbreakID > 0 {
		query += fmt.Sprintf(" AND c.outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	var total, critical int
	err := db.QueryRow(query, args...).Scan(&total, &critical)
	if err != nil {
		return nil, err
	}

	var percentage float64
	if total > 0 {
		percentage = float64(critical) / float64(total) * 100
	}

	return map[string]interface{}{
		"total_admitted": total,
		"critical_cases": critical,
		"percentage":     percentage,
		"percentage_str": fmt.Sprintf("%.2f%%", percentage),
	}, nil
}

// GetCasesBySex returns cases distribution by sex
func GetCasesBySex(db *sql.DB, filters ReportFilters) (map[string]interface{}, error) {
	query := `
		SELECT 
			CASE 
				WHEN c.gender = 1 THEN 'Male'
				WHEN c.gender = 2 THEN 'Female'
				ELSE 'Unknown'
			END as sex,
			COUNT(*) as count
		FROM clients c
		WHERE c.adm_date IS NOT NULL
	`

	var args []interface{}
	argCount := 1

	if filters.StartDate != "" {
		query += fmt.Sprintf(" AND c.adm_date >= $%d", argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		query += fmt.Sprintf(" AND c.adm_date <= $%d", argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	if filters.OutbreakID > 0 {
		query += fmt.Sprintf(" AND c.outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	query += " GROUP BY sex ORDER BY sex"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var labels []string
	var data []int

	for rows.Next() {
		var sex sql.NullString
		var count int

		err := rows.Scan(&sex, &count)
		if err != nil {
			continue
		}

		labels = append(labels, sex.String)
		data = append(data, count)
	}

	return map[string]interface{}{
		"labels": labels,
		"data":   data,
	}, nil
}

// GetCasesByAgeGroup returns cases distribution by age group
func GetCasesByAgeGroup(db *sql.DB, filters ReportFilters) (map[string]interface{}, error) {
	query := `
		SELECT 
			CASE 
				WHEN c.age < 5 THEN '<5'
				WHEN c.age BETWEEN 5 AND 17 THEN '5-17'
				WHEN c.age BETWEEN 18 AND 35 THEN '18-35'
				WHEN c.age BETWEEN 36 AND 59 THEN '36-59'
				WHEN c.age >= 60 THEN '60+'
				ELSE 'Unknown'
			END as age_group,
			COUNT(*) as count
		FROM clients c
		WHERE c.adm_date IS NOT NULL
	`

	var args []interface{}
	argCount := 1

	if filters.StartDate != "" {
		query += fmt.Sprintf(" AND c.adm_date >= $%d", argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		query += fmt.Sprintf(" AND c.adm_date <= $%d", argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	if filters.OutbreakID > 0 {
		query += fmt.Sprintf(" AND c.outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	query += " GROUP BY age_group ORDER BY age_group"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var labels []string
	var data []int

	for rows.Next() {
		var ageGroup sql.NullString
		var count int

		err := rows.Scan(&ageGroup, &count)
		if err != nil {
			continue
		}

		labels = append(labels, ageGroup.String)
		data = append(data, count)
	}

	return map[string]interface{}{
		"labels": labels,
		"data":   data,
	}, nil
}

// GetCasesByLocation returns cases by district/facility
func GetCasesByLocation(db *sql.DB, filters ReportFilters) (map[string]interface{}, error) {
	query := `
		SELECT 
			f.facility_name,
			d.name as district_name,
			COUNT(*) as case_count
		FROM clients c
		LEFT JOIN facility f ON c.site = f.facility_id
		LEFT JOIN districts d ON f.district = d.name
		WHERE c.adm_date IS NOT NULL
	`

	var args []interface{}
	argCount := 1

	if filters.StartDate != "" {
		query += fmt.Sprintf(" AND c.adm_date >= $%d", argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		query += fmt.Sprintf(" AND c.adm_date <= $%d", argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	if filters.OutbreakID > 0 {
		query += fmt.Sprintf(" AND c.outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	if filters.FacilityID > 0 {
		query += fmt.Sprintf(" AND f.facility_id = $%d", argCount)
		args = append(args, filters.FacilityID)
		argCount++
	}

	if filters.DistrictID > 0 {
		query += fmt.Sprintf(" AND d.id = $%d", argCount)
		args = append(args, filters.DistrictID)
		argCount++
	}

	query += " GROUP BY f.facility_name, d.name ORDER BY case_count DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []map[string]interface{}

	for rows.Next() {
		var facilityName, districtName sql.NullString
		var count int

		err := rows.Scan(&facilityName, &districtName, &count)
		if err != nil {
			continue
		}

		locations = append(locations, map[string]interface{}{
			"facility": facilityName.String,
			"district": districtName.String,
			"count":    count,
		})
	}

	return map[string]interface{}{
		"locations": locations,
	}, nil
}

// GetHCWInfections returns healthcare worker infections
func GetHCWInfections(db *sql.DB, filters ReportFilters) (map[string]interface{}, error) {
	// This assumes you have a field to identify healthcare workers
	// You may need to adapt this based on your actual data structure
	query := `
		SELECT 
			COUNT(*) as total_cases,
			COUNT(CASE WHEN c.occupation = 1 THEN 1 END) as hcw_cases
		FROM clients c
		WHERE c.adm_date IS NOT NULL
	`

	var args []interface{}
	argCount := 1

	if filters.StartDate != "" {
		query += fmt.Sprintf(" AND c.adm_date >= $%d", argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		query += fmt.Sprintf(" AND c.adm_date <= $%d", argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	if filters.OutbreakID > 0 {
		query += fmt.Sprintf(" AND c.outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	var total, hcw int
	err := db.QueryRow(query, args...).Scan(&total, &hcw)
	if err != nil {
		return nil, err
	}

	var percentage float64
	if total > 0 {
		percentage = float64(hcw) / float64(total) * 100
	}

	return map[string]interface{}{
		"total_cases":    total,
		"hcw_cases":      hcw,
		"percentage":     percentage,
		"percentage_str": fmt.Sprintf("%.2f%%", percentage),
	}, nil
}

// Helper function to count cases by status
func countByStatus(cases []map[string]interface{}, status string) int {
	count := 0
	for _, c := range cases {
		if c["status"] == status {
			count++
		}
	}
	return count
}

// Helper function to count cases by classification
func countByClassification(cases []map[string]interface{}, classification string) int {
	count := 0
	for _, c := range cases {
		if c["case_classification"] == classification {
			count++
		}
	}
	return count
}

func generateCIFReport(c *fiber.Ctx, db *sql.DB, filters ReportFilters, cifType string) ReportData {
	// Get CIF data based on type
	var query string
	switch cifType {
	case "vhf":
		query = `
			SELECT 
				c.id,
				c.firstname,
				c.lastname,
				c.age,
				c.gender,
				c.adm_date,
				c.status,
				c.outbreak_id,
				f.facility_name,
				d.name as district_name
			FROM clients c
			LEFT JOIN facility f ON c.site = f.facility_id
			LEFT JOIN districts d ON f.district = d.name
			WHERE c.outbreak_id IN (SELECT id FROM outbreaks WHERE outbreak_type = 'vhf')
		`
	case "measles":
		query = `
			SELECT 
				c.id,
				c.firstname,
				c.lastname,
				c.age,
				c.gender,
				c.adm_date,
				c.status,
				c.outbreak_id,
				f.facility_name,
				d.name as district_name
			FROM clients c
			LEFT JOIN facility f ON c.site = f.facility_id
			LEFT JOIN districts d ON f.district = d.name
			WHERE c.outbreak_id IN (SELECT id FROM outbreaks WHERE outbreak_type = 'measles')
		`
	case "polio":
		query = `
			SELECT 
				c.id,
				c.firstname,
				c.lastname,
				c.age,
				c.gender,
				c.adm_date,
				c.status,
				c.outbreak_id,
				f.facility_name,
				d.name as district_name
			FROM clients c
			LEFT JOIN facility f ON c.site = f.facility_id
			LEFT JOIN districts d ON f.district = d.name
			WHERE c.outbreak_id IN (SELECT id FROM outbreaks WHERE outbreak_type = 'polio')
		`
	case "mpox":
		query = `
			SELECT 
				c.id,
				c.firstname,
				c.lastname,
				c.age,
				c.gender,
				c.adm_date,
				c.status,
				c.outbreak_id,
				f.facility_name,
				d.name as district_name
			FROM clients c
			LEFT JOIN facility f ON c.site = f.facility_id
			LEFT JOIN districts d ON f.district = d.name
			WHERE c.outbreak_id IN (SELECT id FROM outbreaks WHERE outbreak_type = 'mpox')
		`
	default:
		return ReportData{Filters: filters, Error: "Invalid CIF type"}
	}

	var args []interface{}
	argCount := 1

	if filters.StartDate != "" {
		query += fmt.Sprintf(" AND c.adm_date >= $%d", argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		query += fmt.Sprintf(" AND c.adm_date <= $%d", argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	if filters.OutbreakID > 0 {
		query += fmt.Sprintf(" AND c.outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	if filters.FacilityID > 0 {
		query += fmt.Sprintf(" AND c.site = $%d", argCount)
		args = append(args, filters.FacilityID)
		argCount++
	}

	query += " ORDER BY c.adm_date DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return ReportData{Filters: filters, Error: "Database error: " + err.Error()}
	}
	defer rows.Close()

	var cases []map[string]interface{}
	for rows.Next() {
		var id int
		var firstName, lastName, gender, admDate, status sql.NullString
		var age sql.NullFloat64
		var outbreakID sql.NullInt64
		var facilityName, districtName sql.NullString

		err := rows.Scan(&id, &firstName, &lastName, &age, &gender, &admDate, &status, &outbreakID, &facilityName, &districtName)
		if err != nil {
			continue
		}

		cases = append(cases, map[string]interface{}{
			"id":          id,
			"name":        fmt.Sprintf("%s %s", firstName.String, lastName.String),
			"age":         age.Float64,
			"gender":      gender.String,
			"adm_date":    admDate.String,
			"status":      status.String,
			"outbreak_id": outbreakID.Int64,
			"facility":    facilityName.String,
			"district":    districtName.String,
		})
	}

	return ReportData{
		Filters: filters,
		Summary: map[string]interface{}{
			"total_cases":  len(cases),
			"active_cases": countByStatus(cases, "active"),
			"recovered":    countByStatus(cases, "recovered"),
			"died":         countByStatus(cases, "died"),
		},
		Tables: map[string]interface{}{
			"cases": cases,
		},
	}
}

func generateLabTechnicianReport(c *fiber.Ctx, db *sql.DB, district int) error {
	// Generate lab technician specific report for their district
	query := `
		SELECT 
			c.id,
			c.firstname,
			c.lastname,
			c.age,
			c.gender,
			c.adm_date,
			c.status,
			c.outbreak_id,
			f.facility_name,
			d.name as district_name
		FROM clients c
		LEFT JOIN facility f ON c.site = f.facility_id
		LEFT JOIN districts d ON f.district = d.name
		WHERE d.id = $1 AND c.outbreak_id IN (SELECT id FROM outbreaks WHERE outbreak_type = 'vhf')
		ORDER BY c.adm_date DESC
	`

	rows, err := db.Query(query, district)
	if err != nil {
		return err
	}
	defer rows.Close()

	var cases []map[string]interface{}
	for rows.Next() {
		var id int
		var firstName, lastName, gender, admDate, status sql.NullString
		var age sql.NullFloat64
		var outbreakID sql.NullInt64
		var facilityName, districtName sql.NullString

		err := rows.Scan(&id, &firstName, &lastName, &age, &gender, &admDate, &status, &outbreakID, &facilityName, &districtName)
		if err != nil {
			continue
		}

		cases = append(cases, map[string]interface{}{
			"id":          id,
			"name":        fmt.Sprintf("%s %s", firstName.String, lastName.String),
			"age":         age.Float64,
			"gender":      gender.String,
			"adm_date":    admDate.String,
			"status":      status.String,
			"outbreak_id": outbreakID.Int64,
			"facility":    facilityName.String,
			"district":    districtName.String,
		})
	}

	// Return the report data
	reportData := ReportData{
		Summary: map[string]interface{}{
			"total_cases":  len(cases),
			"active_cases": countByStatus(cases, "active"),
			"recovered":    countByStatus(cases, "recovered"),
			"died":         countByStatus(cases, "died"),
		},
		Tables: map[string]interface{}{
			"cases": cases,
		},
	}

	// Create a minimal template data structure
	data := &handlers.TemplateData{
		Form: reportData,
	}

	return handlers.GenerateHTML(c, db, data, "report_view")
}

// API Functions for AJAX data loading

// GetQuickStats returns quick statistics for the dashboard
func GetQuickStats(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) error {
	userID, _ := handlers.GetUser(c, sl, store)
	if userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Check if user has reports access using proper database query
	allowedRoles := []string{"admin", "reports", "vhf_lab_technician", "case_manager", "data_analyst", "cases_viewer", "cif_viewer"}
	if !security.HasAnyRole(db, userID, allowedRoles) {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied: Reports permission required"})
	}

	filters := parseReportFilters(c)

	// Get user roles for access level determination
	userRoles := security.GetUserRoles(db, userID)
	userRole := strings.Join(userRoles, ",")
	userFacility := handlers.GetCurrentFacility(c, db, sl, store)
	userDistrict := getUserDistrict(c, db, userID)

	filters = applyAccessRestrictions(filters, userRole, userFacility, userDistrict)

	// Generate quick stats based on report type
	var stats fiber.Map
	switch filters.ReportType {
	case "vhf", "measles", "polio", "mpox":
		stats = getCIFQuickStats(c, db, filters)
	default:
		stats = getGeneralQuickStats(c, db, filters)
	}

	return c.JSON(stats)
}

func getGeneralQuickStats(c *fiber.Ctx, db *sql.DB, filters ReportFilters) fiber.Map {
	// Use only clients table for now to fix infinite rendering
	query := `
		SELECT 
			COUNT(*) as total_cases,
			COUNT(CASE WHEN status = 'active' THEN 1 END) as active_cases,
			COUNT(CASE WHEN status = 'recovered' THEN 1 END) as recovered,
			COUNT(CASE WHEN status = 'died' THEN 1 END) as died
		FROM clients 
		WHERE 1=1
	`

	var args []interface{}
	argCount := 1

	if filters.StartDate != "" {
		query += fmt.Sprintf(" AND adm_date >= $%d", argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		query += fmt.Sprintf(" AND adm_date <= $%d", argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	if filters.OutbreakID > 0 {
		query += fmt.Sprintf(" AND outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	row := db.QueryRow(query, args...)
	var totalCases, activeCases, recovered, died int
	err := row.Scan(&totalCases, &activeCases, &recovered, &died)
	if err != nil {
		return fiber.Map{
			"total_cases":  0,
			"active_cases": 0,
			"recovered":    0,
			"died":         0,
		}
	}

	return fiber.Map{
		"total_cases":  totalCases,
		"active_cases": activeCases,
		"recovered":    recovered,
		"died":         died,
	}
}

func getCIFQuickStats(c *fiber.Ctx, db *sql.DB, filters ReportFilters) fiber.Map {
	// Use only clients table for now to fix infinite rendering
	query := `
		SELECT 
			COUNT(*) as total_cases,
			COUNT(CASE WHEN status = 'active' THEN 1 END) as confirmed_cases,
			COUNT(CASE WHEN status = 'recovered' THEN 1 END) as probable_cases,
			COUNT(CASE WHEN status = 'died' THEN 1 END) as suspect_cases
		FROM clients 
		WHERE 1=1
	`

	var args []interface{}
	argCount := 1

	if filters.StartDate != "" {
		query += fmt.Sprintf(" AND adm_date >= $%d", argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		query += fmt.Sprintf(" AND adm_date <= $%d", argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	if filters.OutbreakID > 0 {
		query += fmt.Sprintf(" AND outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	row := db.QueryRow(query, args...)
	var totalCases, confirmedCases, probableCases, suspectCases int
	err := row.Scan(&totalCases, &confirmedCases, &probableCases, &suspectCases)
	if err != nil {
		return fiber.Map{
			"total_cases":     0,
			"confirmed_cases": 0,
			"probable_cases":  0,
			"suspect_cases":   0,
		}
	}

	return fiber.Map{
		"total_cases":     totalCases,
		"confirmed_cases": confirmedCases,
		"probable_cases":  probableCases,
		"suspect_cases":   suspectCases,
	}
}

// GetHealthIndicators returns all 14 health indicators for the dashboard
func GetHealthIndicators(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) error {
	userID, _ := handlers.GetUser(c, sl, store)
	if userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Check if user has reports access using proper database query
	allowedRoles := []string{"admin", "reports", "vhf_lab_technician", "case_manager", "data_analyst", "cases_viewer", "cif_viewer"}
	if !security.HasAnyRole(db, userID, allowedRoles) {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied: Reports permission required"})
	}

	// Parse filters from query parameters
	filters := parseReportFilters(c)

	// Get user access restrictions
	userRoles := security.GetUserRoles(db, userID)
	userRole := strings.Join(userRoles, ",")
	userFacility := handlers.GetCurrentFacility(c, db, sl, store)
	userDistrict := getUserDistrict(c, db, userID)

	// Apply access restrictions
	filters = applyAccessRestrictions(filters, userRole, userFacility, userDistrict)

	// Build base query with filters
	baseQuery := "FROM clients WHERE 1=1"
	var args []interface{}
	argCount := 1

	// Apply date filters
	if filters.StartDate != "" {
		baseQuery += fmt.Sprintf(" AND adm_date >= $%d", argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		baseQuery += fmt.Sprintf(" AND adm_date <= $%d", argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	// Apply outbreak filter
	if filters.OutbreakID > 0 {
		baseQuery += fmt.Sprintf(" AND outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	// Apply facility filter
	if filters.FacilityID > 0 {
		baseQuery += fmt.Sprintf(" AND facility_id = $%d", argCount)
		args = append(args, filters.FacilityID)
		argCount++
	}

	// Apply district filter
	if filters.DistrictID > 0 {
		baseQuery += fmt.Sprintf(" AND district_id = $%d", argCount)
		args = append(args, filters.DistrictID)
		argCount++
	}

	// Apply patient type filter
	if filters.PatientType != "" {
		baseQuery += fmt.Sprintf(" AND patient_type = $%d", argCount)
		args = append(args, filters.PatientType)
		argCount++
	}

	// Apply outcome filter
	if filters.Outcome != "" {
		baseQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, filters.Outcome)
		argCount++
	}

	// 1. New admissions (daily) - only if date filter is applied
	var newAdmissions int
	if filters.StartDate != "" && filters.EndDate != "" {
		// If date range is specified, count admissions in that range
		newAdmissionsQuery := fmt.Sprintf("SELECT COUNT(*) %s", baseQuery)
		err := db.QueryRow(newAdmissionsQuery, args...).Scan(&newAdmissions)
		if err != nil {
			sl.Error("Error fetching new admissions", "error", err)
			newAdmissions = 0
		}
	} else {
		// If no date filter, show 0 for new admissions
		newAdmissions = 0
	}

	// 2. Cumulative confirmed cases
	confirmedQuery := fmt.Sprintf("SELECT COUNT(*) %s AND status = 'confirmed'", baseQuery)
	var confirmedCases int
	if err := db.QueryRow(confirmedQuery, args...).Scan(&confirmedCases); err != nil {
		sl.Error("Error fetching confirmed cases", "error", err)
		confirmedCases = 0
	}

	// 3. Cumulative suspected cases
	suspectedQuery := fmt.Sprintf("SELECT COUNT(*) %s AND status = 'suspected'", baseQuery)
	var suspectedCases int
	if err := db.QueryRow(suspectedQuery, args...).Scan(&suspectedCases); err != nil {
		sl.Error("Error fetching suspected cases", "error", err)
		suspectedCases = 0
	}

	// 4. Cumulative deaths
	deathsQuery := fmt.Sprintf("SELECT COUNT(*) %s AND status = 'died'", baseQuery)
	var cumulativeDeaths int
	if err := db.QueryRow(deathsQuery, args...).Scan(&cumulativeDeaths); err != nil {
		sl.Error("Error fetching cumulative deaths", "error", err)
		cumulativeDeaths = 0
	}

	// 5. Case Fatality Rate (CFR)
	var cfr float64
	if confirmedCases > 0 {
		cfr = (float64(cumulativeDeaths) / float64(confirmedCases)) * 100
	}

	// 6. Current admissions (active cases in hospital)
	currentQuery := fmt.Sprintf("SELECT COUNT(*) %s AND status = 'active'", baseQuery)
	var currentAdmissions int
	if err := db.QueryRow(currentQuery, args...).Scan(&currentAdmissions); err != nil {
		sl.Error("Error fetching current admissions", "error", err)
		currentAdmissions = 0
	}

	// 7. Discharges (cumulative)
	dischargesQuery := fmt.Sprintf("SELECT COUNT(*) %s AND status = 'discharged'", baseQuery)
	var discharges int
	if err := db.QueryRow(dischargesQuery, args...).Scan(&discharges); err != nil {
		sl.Error("Error fetching discharges", "error", err)
		discharges = 0
	}

	// 8. Severe cases admitted
	severeQuery := fmt.Sprintf("SELECT COUNT(*) %s AND severity = 'severe'", baseQuery)
	var severeCases int
	if err := db.QueryRow(severeQuery, args...).Scan(&severeCases); err != nil {
		sl.Error("Error fetching severe cases", "error", err)
		severeCases = 0
	}

	// 9. Critical cases admitted
	criticalQuery := fmt.Sprintf("SELECT COUNT(*) %s AND severity = 'critical'", baseQuery)
	var criticalCases int
	if err := db.QueryRow(criticalQuery, args...).Scan(&criticalCases); err != nil {
		sl.Error("Error fetching critical cases", "error", err)
		criticalCases = 0
	}

	// 10. Cases by sex
	sexQuery := fmt.Sprintf("SELECT COUNT(*) %s AND gender IS NOT NULL", baseQuery)
	var casesBySex int
	if err := db.QueryRow(sexQuery, args...).Scan(&casesBySex); err != nil {
		sl.Error("Error fetching cases by sex", "error", err)
		casesBySex = 0
	}

	// 11. Cases by age group
	ageQuery := fmt.Sprintf("SELECT COUNT(*) %s AND age IS NOT NULL", baseQuery)
	var casesByAge int
	if err := db.QueryRow(ageQuery, args...).Scan(&casesByAge); err != nil {
		sl.Error("Error fetching cases by age", "error", err)
		casesByAge = 0
	}

	// 12. Cases by district/facility
	locationQuery := fmt.Sprintf("SELECT COUNT(*) %s AND (district_id IS NOT NULL OR facility_id IS NOT NULL)", baseQuery)
	var casesByLocation int
	if err := db.QueryRow(locationQuery, args...).Scan(&casesByLocation); err != nil {
		sl.Error("Error fetching cases by location", "error", err)
		casesByLocation = 0
	}

	// 13. Healthcare worker (HCW) infections
	hcwQuery := fmt.Sprintf("SELECT COUNT(*) %s AND is_healthcare_worker = true", baseQuery)
	var hcwInfections int
	if err := db.QueryRow(hcwQuery, args...).Scan(&hcwInfections); err != nil {
		sl.Error("Error fetching HCW infections", "error", err)
		hcwInfections = 0
	}

	return c.JSON(fiber.Map{
		"new_admissions":     newAdmissions,
		"confirmed_cases":    confirmedCases,
		"suspected_cases":    suspectedCases,
		"cumulative_deaths":  cumulativeDeaths,
		"cfr":                cfr,
		"current_admissions": currentAdmissions,
		"discharges":         discharges,
		"severe_cases":       severeCases,
		"critical_cases":     criticalCases,
		"cases_by_sex":       casesBySex,
		"cases_by_age":       casesByAge,
		"cases_by_location":  casesByLocation,
		"hcw_infections":     hcwInfections,
	})
}

// GetChartData returns chart data based on filters
func GetChartData(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) error {
	userID, _ := handlers.GetUser(c, sl, store)
	if userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Check if user has reports access using proper database query
	allowedRoles := []string{"admin", "reports", "vhf_lab_technician", "case_manager", "data_analyst", "cases_viewer", "cif_viewer"}
	if !security.HasAnyRole(db, userID, allowedRoles) {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied: Reports permission required"})
	}

	chartType := c.Params("type") // Get from URL parameter, not query parameter
	filters := parseReportFilters(c)

	// Get user roles for access level determination
	userRoles := security.GetUserRoles(db, userID)
	userRole := strings.Join(userRoles, ",") // Join multiple roles with comma
	userFacility := handlers.GetCurrentFacility(c, db, sl, store)
	userDistrict := getUserDistrict(c, db, userID)

	filters = applyAccessRestrictions(filters, userRole, userFacility, userDistrict)

	var chartData interface{}
	var err error

	switch chartType {
	case "trends":
		chartData, err = getTrendsData(c, db, filters)
	case "outcomes":
		chartData, err = getOutcomesData(c, db, filters)
	case "facilities":
		chartData, err = getFacilitiesData(c, db, filters)
	case "age_groups":
		chartData, err = getAgeGroupsData(c, db, filters)
	default:
		return c.Status(400).JSON(fiber.Map{"error": "Invalid chart type"})
	}

	if err != nil {
		sl.Error("Error getting chart data", "error", err, "chart_type", chartType)
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	return c.JSON(chartData)
}

// GetTableData returns table data based on filters
func GetTableData(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) error {
	userID, _ := handlers.GetUser(c, sl, store)
	if userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Check if user has reports access using proper database query
	allowedRoles := []string{"admin", "reports", "vhf_lab_technician", "case_manager", "data_analyst", "cases_viewer", "cif_viewer"}
	if !security.HasAnyRole(db, userID, allowedRoles) {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied: Reports permission required"})
	}

	filters := parseReportFilters(c)

	// Get user roles for access level determination
	userRoles := security.GetUserRoles(db, userID)
	userRole := strings.Join(userRoles, ",") // Join multiple roles with comma
	userFacility := handlers.GetCurrentFacility(c, db, sl, store)
	userDistrict := getUserDistrict(c, db, userID)

	filters = applyAccessRestrictions(filters, userRole, userFacility, userDistrict)

	// Build query with filters
	query := `SELECT 
		id, firstname, lastname, age, gender, adm_date, status
		FROM clients WHERE 1=1`

	var args []interface{}
	argCount := 1

	if filters.StartDate != "" {
		query += fmt.Sprintf(" AND adm_date >= $%d", argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		query += fmt.Sprintf(" AND adm_date <= $%d", argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	if filters.OutbreakID > 0 {
		query += fmt.Sprintf(" AND outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	if userFacility > 0 {
		query += fmt.Sprintf(" AND site = $%d", argCount)
		args = append(args, userFacility)
		argCount++
	}

	query += " ORDER BY adm_date DESC LIMIT 100"

	rows, err := db.Query(query, args...)
	if err != nil {
		sl.Error("Error getting table data", "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}
	defer rows.Close()

	var data []fiber.Map
	for rows.Next() {
		var id int
		var firstName, lastName, gender, admDate, status sql.NullString
		var age sql.NullFloat64

		err := rows.Scan(&id, &firstName, &lastName, &age, &gender, &admDate, &status)
		if err != nil {
			continue
		}

		data = append(data, fiber.Map{
			"id":       id,
			"name":     fmt.Sprintf("%s %s", firstName.String, lastName.String),
			"age":      age.Float64,
			"gender":   gender.String,
			"adm_date": admDate.String,
			"status":   status.String,
		})
	}

	return c.JSON(fiber.Map{"data": data})
}

// ExportReport handles report export functionality
func ExportReport(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) error {
	userID, _ := handlers.GetUser(c, sl, store)
	if userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	exportType := c.FormValue("type") // csv, excel, pdf
	filters := parseReportFilters(c)

	// Get user roles for access level determination
	userRoles := security.GetUserRoles(db, userID)
	userRole := strings.Join(userRoles, ",") // Join multiple roles with comma
	userFacility := handlers.GetCurrentFacility(c, db, sl, store)
	userDistrict := getUserDistrict(c, db, userID)

	filters = applyAccessRestrictions(filters, userRole, userFacility, userDistrict)

	// TODO: Implement actual export functionality
	// For now, return a placeholder response
	return c.JSON(fiber.Map{
		"message": fmt.Sprintf("Export functionality for %s will be implemented", exportType),
		"filters": filters,
	})
}

// Helper functions for filter persistence
func getFilterStateFromSession(sess *session.Session) *ReportFilters {
	if sess == nil {
		return nil
	}

	filterData := sess.Get("report_filters")
	if filterData == nil {
		return nil
	}

	// Convert session data back to ReportFilters
	if filterMap, ok := filterData.(map[string]interface{}); ok {
		return &ReportFilters{
			StartDate:   getStringFromMap(filterMap, "start_date"),
			EndDate:     getStringFromMap(filterMap, "end_date"),
			OutbreakID:  getIntFromMap(filterMap, "outbreak_id"),
			FacilityID:  getIntFromMap(filterMap, "facility_id"),
			DistrictID:  getIntFromMap(filterMap, "district_id"),
			ReportType:  getStringFromMap(filterMap, "report_type"),
			PatientType: getStringFromMap(filterMap, "patient_type"),
			Outcome:     getStringFromMap(filterMap, "outcome"),
		}
	}

	return nil
}

func saveFilterStateToSession(sess *session.Session, filters ReportFilters) {
	if sess == nil {
		return
	}

	filterMap := map[string]interface{}{
		"start_date":   filters.StartDate,
		"end_date":     filters.EndDate,
		"outbreak_id":  filters.OutbreakID,
		"facility_id":  filters.FacilityID,
		"district_id":  filters.DistrictID,
		"report_type":  filters.ReportType,
		"patient_type": filters.PatientType,
		"outcome":      filters.Outcome,
	}

	sess.Set("report_filters", filterMap)
}

func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getIntFromMap(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return 0
}

// Helper functions for chart data

func getTrendsData(c *fiber.Ctx, db *sql.DB, filters ReportFilters) (interface{}, error) {
	var query string
	var args []interface{}
	argCount := 1

	// Determine which table to query based on report type
	switch filters.ReportType {
	case "vhf":
		query = `
			SELECT 
				DATE(date_of_birth) as date,
				COUNT(*) as cases
			FROM vhf_patients 
			WHERE date_of_birth IS NOT NULL
		`
	case "measles":
		query = `
			SELECT 
				DATE(visit_date) as date,
				COUNT(*) as cases
			FROM measles_case_investigations 
			WHERE visit_date IS NOT NULL
		`
	case "polio":
		query = `
			SELECT 
				DATE(received_date) as date,
				COUNT(*) as cases
			FROM polio_case_investigation 
			WHERE received_date IS NOT NULL
		`
	case "mpox":
		query = `
			SELECT 
				DATE(created_at) as date,
				COUNT(*) as cases
			FROM mpox_case_investigation 
			WHERE created_at IS NOT NULL
		`
	default:
		query = `
			SELECT 
				DATE(adm_date) as date,
				COUNT(*) as cases
			FROM clients 
			WHERE adm_date IS NOT NULL
		`
	}

	// Add date filters based on the table being queried
	if filters.StartDate != "" {
		var dateColumn string
		switch filters.ReportType {
		case "vhf":
			dateColumn = "date_of_birth"
		case "measles":
			dateColumn = "visit_date"
		case "polio":
			dateColumn = "received_date"
		case "mpox":
			dateColumn = "created_at"
		default:
			dateColumn = "adm_date"
		}
		query += fmt.Sprintf(" AND %s >= $%d", dateColumn, argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		var dateColumn string
		switch filters.ReportType {
		case "vhf":
			dateColumn = "date_of_birth"
		case "measles":
			dateColumn = "visit_date"
		case "polio":
			dateColumn = "received_date"
		case "mpox":
			dateColumn = "created_at"
		default:
			dateColumn = "adm_date"
		}
		query += fmt.Sprintf(" AND %s <= $%d", dateColumn, argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	// Add outbreak filter if specified
	if filters.OutbreakID > 0 {
		query += fmt.Sprintf(" AND outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	// Add GROUP BY clause based on the table being queried
	var dateColumn string
	switch filters.ReportType {
	case "vhf":
		dateColumn = "DATE(date_of_birth)"
	case "measles":
		dateColumn = "DATE(visit_date)"
	case "polio":
		dateColumn = "DATE(received_date)"
	case "mpox":
		dateColumn = "DATE(created_at)"
	default:
		dateColumn = "DATE(adm_date)"
	}
	query += fmt.Sprintf(" GROUP BY %s ORDER BY date", dateColumn)

	rows, err := db.Query(query, args...)
	if err != nil {
		// If the CIF table doesn't exist, fall back to clients table
		if filters.ReportType != "general" {
			return getTrendsDataFallback(c, db, filters)
		}
		return nil, err
	}
	defer rows.Close()

	var labels []string
	var data []int

	for rows.Next() {
		var date sql.NullString
		var cases int

		err := rows.Scan(&date, &cases)
		if err != nil {
			continue
		}

		labels = append(labels, date.String)
		data = append(data, cases)
	}

	return fiber.Map{
		"labels": labels,
		"data":   data,
	}, nil
}

func getTrendsDataFallback(c *fiber.Ctx, db *sql.DB, filters ReportFilters) (interface{}, error) {
	// Fallback to clients table if CIF table doesn't exist
	query := `
		SELECT 
			DATE(adm_date) as date,
			COUNT(*) as cases
		FROM clients 
		WHERE adm_date IS NOT NULL
	`

	var args []interface{}
	argCount := 1

	if filters.StartDate != "" {
		query += fmt.Sprintf(" AND adm_date >= $%d", argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		query += fmt.Sprintf(" AND adm_date <= $%d", argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	if filters.OutbreakID > 0 {
		query += fmt.Sprintf(" AND outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	query += " GROUP BY DATE(adm_date) ORDER BY date"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var labels []string
	var data []int

	for rows.Next() {
		var date sql.NullString
		var cases int

		err := rows.Scan(&date, &cases)
		if err != nil {
			continue
		}

		labels = append(labels, date.String)
		data = append(data, cases)
	}

	return fiber.Map{
		"labels": labels,
		"data":   data,
	}, nil
}

func getOutcomesData(c *fiber.Ctx, db *sql.DB, filters ReportFilters) (interface{}, error) {
	var query string
	var args []interface{}
	argCount := 1

	// Determine which table to query based on report type
	switch filters.ReportType {
	case "vhf":
		query = `
			SELECT 
				status,
				COUNT(*) as cases
			FROM vhf_patients 
			WHERE status IS NOT NULL
		`
	case "measles":
		query = `
			SELECT 
				CASE 
					WHEN final_classification = 1 THEN 'Confirmed'
					WHEN final_classification = 2 THEN 'Probable'
					WHEN final_classification = 3 THEN 'Suspect'
					ELSE 'Unknown'
				END as status,
				COUNT(*) as cases
			FROM measles_case_investigations 
			WHERE final_classification IS NOT NULL
		`
	case "polio":
		query = `
			SELECT 
				'AFP Case' as status,
				COUNT(*) as cases
			FROM polio_case_investigation 
			WHERE 1=1
		`
	case "mpox":
		query = `
			SELECT 
				'Case' as status,
				COUNT(*) as cases
			FROM mpox_case_investigation 
			WHERE 1=1
		`
	default:
		query = `
			SELECT 
				status,
				COUNT(*) as cases
			FROM clients 
			WHERE status IS NOT NULL
		`
	}

	// Add date filters based on the table being queried
	if filters.StartDate != "" {
		var dateColumn string
		switch filters.ReportType {
		case "vhf":
			dateColumn = "date_of_birth"
		case "measles":
			dateColumn = "visit_date"
		case "polio":
			dateColumn = "received_date"
		case "mpox":
			dateColumn = "created_at"
		default:
			dateColumn = "adm_date"
		}
		query += fmt.Sprintf(" AND %s >= $%d", dateColumn, argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		var dateColumn string
		switch filters.ReportType {
		case "vhf":
			dateColumn = "date_of_birth"
		case "measles":
			dateColumn = "visit_date"
		case "polio":
			dateColumn = "received_date"
		case "mpox":
			dateColumn = "created_at"
		default:
			dateColumn = "adm_date"
		}
		query += fmt.Sprintf(" AND %s <= $%d", dateColumn, argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	// Add outbreak filter if specified
	if filters.OutbreakID > 0 {
		query += fmt.Sprintf(" AND outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	// Add GROUP BY clause
	if filters.ReportType == "measles" {
		query += " GROUP BY final_classification ORDER BY cases DESC"
	} else {
		query += " GROUP BY status ORDER BY cases DESC"
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		// If the CIF table doesn't exist, fall back to clients table
		if filters.ReportType != "general" {
			return getOutcomesDataFallback(c, db, filters)
		}
		return nil, err
	}
	defer rows.Close()

	var labels []string
	var data []int

	for rows.Next() {
		var status sql.NullString
		var cases int

		err := rows.Scan(&status, &cases)
		if err != nil {
			continue
		}

		labels = append(labels, status.String)
		data = append(data, cases)
	}

	return fiber.Map{
		"labels": labels,
		"data":   data,
	}, nil
}

func getOutcomesDataFallback(c *fiber.Ctx, db *sql.DB, filters ReportFilters) (interface{}, error) {
	// Fallback to clients table if CIF table doesn't exist
	query := `
		SELECT 
			status,
			COUNT(*) as cases
		FROM clients 
		WHERE status IS NOT NULL
	`

	var args []interface{}
	argCount := 1

	if filters.StartDate != "" {
		query += fmt.Sprintf(" AND adm_date >= $%d", argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		query += fmt.Sprintf(" AND adm_date <= $%d", argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	if filters.OutbreakID > 0 {
		query += fmt.Sprintf(" AND outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	query += " GROUP BY status ORDER BY cases DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var labels []string
	var data []int

	for rows.Next() {
		var status sql.NullString
		var cases int

		err := rows.Scan(&status, &cases)
		if err != nil {
			continue
		}

		labels = append(labels, status.String)
		data = append(data, cases)
	}

	return fiber.Map{
		"labels": labels,
		"data":   data,
	}, nil
}

func getFacilitiesData(c *fiber.Ctx, db *sql.DB, filters ReportFilters) (interface{}, error) {
	// Use only clients table for now to fix infinite rendering
	query := `
		SELECT 
			f.facility_name,
			COUNT(*) as cases
		FROM clients c
		LEFT JOIN facility f ON c.facility_id = f.facility_id
		WHERE f.facility_name IS NOT NULL
	`

	var args []interface{}
	argCount := 1

	if filters.StartDate != "" {
		query += fmt.Sprintf(" AND c.adm_date >= $%d", argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		query += fmt.Sprintf(" AND c.adm_date <= $%d", argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	if filters.OutbreakID > 0 {
		query += fmt.Sprintf(" AND c.outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	query += " GROUP BY f.facility_name ORDER BY cases DESC LIMIT 10"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var labels []string
	var data []int

	for rows.Next() {
		var facilityName sql.NullString
		var cases int

		err := rows.Scan(&facilityName, &cases)
		if err != nil {
			continue
		}

		labels = append(labels, facilityName.String)
		data = append(data, cases)
	}

	return fiber.Map{
		"labels": labels,
		"data":   data,
	}, nil
}

func getAgeGroupsData(c *fiber.Ctx, db *sql.DB, filters ReportFilters) (interface{}, error) {
	// Use only clients table for now to fix infinite rendering
	query := `
		SELECT 
			CASE 
				WHEN age < 5 THEN '0-4'
				WHEN age >= 5 AND age < 15 THEN '5-14'
				WHEN age >= 15 AND age < 25 THEN '15-24'
				WHEN age >= 25 AND age < 35 THEN '25-34'
				WHEN age >= 35 AND age < 45 THEN '35-44'
				WHEN age >= 45 AND age < 55 THEN '45-54'
				WHEN age >= 55 AND age < 65 THEN '55-64'
				ELSE '65+'
			END as age_group,
			COUNT(*) as cases
		FROM clients 
		WHERE age IS NOT NULL
	`

	var args []interface{}
	argCount := 1

	if filters.StartDate != "" {
		query += fmt.Sprintf(" AND adm_date >= $%d", argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		query += fmt.Sprintf(" AND adm_date <= $%d", argCount)
		args = append(args, filters.EndDate)
		argCount++
	}

	if filters.OutbreakID > 0 {
		query += fmt.Sprintf(" AND outbreak_id = $%d", argCount)
		args = append(args, filters.OutbreakID)
		argCount++
	}

	query += " GROUP BY age_group ORDER BY age_group"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var labels []string
	var data []int

	for rows.Next() {
		var ageGroup sql.NullString
		var cases int

		err := rows.Scan(&ageGroup, &cases)
		if err != nil {
			continue
		}

		labels = append(labels, ageGroup.String)
		data = append(data, cases)
	}

	return fiber.Map{
		"labels": labels,
		"data":   data,
	}, nil
}

// renderReportView renders the report view template using the standard layout
func renderReportView(c *fiber.Ctx, db *sql.DB, data interface{}) error {
	// Convert data to TemplateData if it's not already
	var templateData *handlers.TemplateData
	if td, ok := data.(*handlers.TemplateData); ok {
		templateData = td
	} else if td, ok := data.(handlers.TemplateData); ok {
		templateData = &td
	} else {
		// If it's not TemplateData, create a new one
		templateData = handlers.NewTemplateData(c, nil)
		if formMap, ok := data.(map[string]interface{}); ok {
			templateData.Form = formMap
		}
	}

	// Determine which template to use based on report type
	var templateName string
	if formData, ok := templateData.Form.(map[string]interface{}); ok {
		// Check if ReportType is directly in the form data
		if reportType, ok := formData["ReportType"].(string); ok {
			switch reportType {
			case "indicators":
				templateName = "indicators_report"
			case "vhf", "measles", "polio", "mpox":
				templateName = "cif_report"
			default:
				templateName = "report_view"
			}
		} else if filters, ok := formData["Filters"].(ReportFilters); ok {
			// Check if ReportType is in the Filters struct
			switch filters.ReportType {
			case "indicators":
				templateName = "indicators_report"
			case "vhf", "measles", "polio", "mpox":
				templateName = "cif_report"
			default:
				templateName = "report_view"
			}
		} else {
			templateName = "report_view"
		}
	} else {
		templateName = "report_view"
	}

	// Use the standard GenerateHTML function with the layout
	return handlers.GenerateHTML(c, db, templateData, templateName)
}
