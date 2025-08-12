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

	"html/template"
	"os"
	"path/filepath"

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
	allowedRoles := []string{"admin", "reports", "vhf_lab_technician", "case_manager", "data_analyst"}
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

	data := handlers.NewTemplateData(c, store)
	data.Form = map[string]interface{}{
		"Title":       "Reports Dashboard",
		"UserID":      userName,
		"Outbreaks":   outbreaks,
		"Facilities":  facilities,
		"Districts":   districts,
		"AccessLevel": getAccessLevel(userRole, userFacility, userDistrict),
		"Filters":     filterState, // Add current filter state
	}

	return handlers.GenerateHTML(c, db, data, "reports_dashboard")
}

// GenerateReport handles report generation with filter persistence
func GenerateReport(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) error {
	userID, _ := handlers.GetUser(c, sl, store)
	if userID == 0 {
		return c.Redirect("/login")
	}

	// Parse filters from form
	filters := parseReportFilters(c)

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
			v.first_name,
			v.last_name,
			v.age,
			v.gender,
			v.date_of_onset,
			v.case_classification,
			v.outbreak_id,
			f.facility_name,
			d.name as district_name
		FROM vhf_case_investigation_form v
		LEFT JOIN afi_facilities f ON v.reporting_health_facility_name = f.name
		LEFT JOIN districts d ON f.district = d.name
		WHERE v.outbreak_id IN (SELECT id FROM outbreaks WHERE outbreak_type = 'vhf')
	`

	var args []interface{}
	argCount := 1

	if filters.StartDate != "" {
		query += fmt.Sprintf(" AND v.date_of_onset >= $%d", argCount)
		args = append(args, filters.StartDate)
		argCount++
	}

	if filters.EndDate != "" {
		query += fmt.Sprintf(" AND v.date_of_onset <= $%d", argCount)
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

	query += " ORDER BY v.date_of_onset DESC"

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

// GetChartData returns chart data based on filters
func GetChartData(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config handlers.Config) error {
	userID, _ := handlers.GetUser(c, sl, store)
	if userID == 0 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
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

// renderReportView renders the report view template as a standalone HTML document
func renderReportView(c *fiber.Ctx, db *sql.DB, data interface{}) error {
	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		return c.Status(500).SendString(fmt.Sprintf("Failed to get working directory: %v", err))
	}

	// Try multiple possible locations for the templates
	possiblePaths := []string{
		filepath.Join(wd, "ui", "html"),                    // Current directory
		filepath.Join(wd, "..", "ui", "html"),              // One level up
		filepath.Join(wd, "..", "..", "ui", "html"),        // Two levels up
		filepath.Join(wd, "cmd", "ui", "html"),             // cmd directory
		filepath.Join(wd, "..", "cmd", "ui", "html"),       // cmd directory one level up
		filepath.Join(wd, "..", "..", "cmd", "ui", "html"), // cmd directory two levels up
	}

	var basePath string
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			basePath = path
			break
		}
	}

	if basePath == "" {
		return c.Status(500).SendString(fmt.Sprintf("Template directory not found. Tried paths: %v", possiblePaths))
	}

	// Load the report_view template directly
	templateFile := filepath.Join(basePath, "report_view.html")
	content, err := os.ReadFile(templateFile)
	if err != nil {
		return c.Status(500).SendString(fmt.Sprintf("Failed to read report_view template: %v", err))
	}

	// Create a new template without layout
	tmpl := template.New("report_view").Funcs(handlers.CreateTemplateFunctions(c, db))

	// Parse the template content
	_, err = tmpl.Parse(string(content))
	if err != nil {
		return c.Status(500).SendString(fmt.Sprintf("Failed to parse report_view template: %v", err))
	}

	// Execute template and write output
	c.Set("Content-Type", "text/html")
	return tmpl.Execute(c.Response().BodyWriter(), data)
}
