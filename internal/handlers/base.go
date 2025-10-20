package handlers

import (
	"bytes"
	"case/internal/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gorilla/schema"
)

// Global variables
var (
	CtxG *fiber.Ctx
	dbG  *sql.DB
)

// Role constants for RBAC system
const (
	RoleSuperAdmin          = "super_admin"
	RoleAdmin               = "admin"
	RoleManager             = "manager"
	RoleDataEntry           = "data_entry"
	RoleViewer              = "viewer"
	RoleLabTechnician       = "lab_technician"
	RoleSurveillanceOfficer = "surveillance_officer"
	RoleInventoryManager    = "inventory_manager"
	RoleInventoryClerk      = "inventory_clerk"
	RoleOutbreakCoordinator = "outbreak_coordinator"
	RoleCaseManager         = "case_manager"
	RoleReports             = "reports"
)

// Permission constants for RBAC system
const (
	// VHF Patients permissions
	PermissionVHFPatientsCreate = "vhf_patients:create"
	PermissionVHFPatientsRead   = "vhf_patients:read"
	PermissionVHFPatientsUpdate = "vhf_patients:update"
	PermissionVHFPatientsDelete = "vhf_patients:delete"
	PermissionVHFPatientsExport = "vhf_patients:export"

	// Users permissions
	PermissionUsersCreate = "users:create"
	PermissionUsersRead   = "users:read"
	PermissionUsersUpdate = "users:update"
	PermissionUsersDelete = "users:delete"

	// Reports permissions
	PermissionReportsRead   = "reports:read"
	PermissionReportsExport = "reports:export"

	// Outbreaks permissions
	PermissionOutbreaksCreate = "outbreaks:create"
	PermissionOutbreaksRead   = "outbreaks:read"
	PermissionOutbreaksUpdate = "outbreaks:update"
	PermissionOutbreaksDelete = "outbreaks:delete"

	// Inventory permissions
	PermissionInventoryCategoriesCreate = "inventory_categories:create"
	PermissionInventoryCategoriesRead   = "inventory_categories:read"
	PermissionInventoryCategoriesUpdate = "inventory_categories:update"
	PermissionInventoryCategoriesDelete = "inventory_categories:delete"

	PermissionInventoryItemsCreate = "inventory_items:create"
	PermissionInventoryItemsRead   = "inventory_items:read"
	PermissionInventoryItemsUpdate = "inventory_items:update"
	PermissionInventoryItemsDelete = "inventory_items:delete"

	PermissionInventorySuppliersCreate = "inventory_suppliers:create"
	PermissionInventorySuppliersRead   = "inventory_suppliers:read"
	PermissionInventorySuppliersUpdate = "inventory_suppliers:update"
	PermissionInventorySuppliersDelete = "inventory_suppliers:delete"

	PermissionInventoryTransactionsCreate = "inventory_transactions:create"
	PermissionInventoryTransactionsRead   = "inventory_transactions:read"
	PermissionInventoryTransactionsUpdate = "inventory_transactions:update"
	PermissionInventoryTransactionsDelete = "inventory_transactions:delete"

	PermissionInventoryReportsRead     = "inventory_reports:read"
	PermissionInventoryReportsGenerate = "inventory_reports:generate"
)

// Role checking helper functions
func HasRole(userRoles []string, role string) bool {
	for _, r := range userRoles {
		if r == role {
			return true
		}
	}
	return false
}

func HasAnyRole(userRoles []string, roles ...string) bool {
	for _, userRole := range userRoles {
		for _, role := range roles {
			if userRole == role {
				return true
			}
		}
	}
	return false
}

func HasPermission(userPermissions map[string][]string, resource, action string) bool {
	if actions, exists := userPermissions[resource]; exists {
		for _, a := range actions {
			if a == action {
				return true
			}
		}
	}
	return false
}

// GetRoleHierarchy returns roles in order of precedence (highest to lowest)
func GetRoleHierarchy() []string {
	return []string{
		RoleSuperAdmin,
		RoleAdmin,
		RoleManager,
		RoleInventoryManager,
		RoleOutbreakCoordinator,
		RoleCaseManager,
		RoleReports,
		RoleLabTechnician,
		RoleSurveillanceOfficer,
		RoleInventoryClerk,
		RoleDataEntry,
		RoleViewer,
	}
}

type Config struct {
	Address      string `json:"Port"`
	ReadTimeout  int64  `json:"ReadTimeout"`
	WriteTimeout int64  `json:"WriteTimeout"`
	Static       string `json:"Static"`
	Ux           string `json:"Ux"`
	Px           string `json:"Px"`
	Dx           string `json:"Dx"`
	LogFile      string `json:"LogFile"`
	LogData      string `json:"LogData"`
	Facility     string `json:"Facility"`
	SMSUsername  string `json:"SMSUser"`
	SMSPassword  string `json:"SMSPassword"`
	SMSBaseURL   string `json:"SMSURL"`
	DHIS2API     struct {
		BaseURL  string `json:"BaseURL"`
		Username string `json:"Username"`
		Password string `json:"Password"`
	} `json:"DHIS2API"`
	AlertsAPI struct {
		BaseURL  string `json:"BaseURL"`
		Username string `json:"Username"`
		Password string `json:"Password"`
	} `json:"AlertsAPI"`
}

type TemplateData struct {
	CurrentYear     int
	User            string
	Role            int
	IsIDPos         bool
	OutbreakID      int
	IsOutbreakID    bool
	AdmissionID     int
	IsAdmissionID   bool
	ClientID        int    // Add ClientID field for templates
	Title           string // Add Title field for page titles
	CurrentPath     string // Add CurrentPath field for navigation context
	Form            any
	FormRef         any
	FormChild1      any
	FormChild2      any
	FormChild3      any
	FormChild4      any
	FormChild5      any
	Ses             any
	Items           []interface{}
	Optionz         map[string]map[string]string
	Flash           string
	Menuz           string
	IsAuthenticated bool
	CSRFToken       string                 // Add a CSRFToken field.
	IsNew           bool                   // Add IsNew field for form templates
	Message         string                 // Add Message field for alerts
	MessageType     string                 // Add MessageType field for alert styling
	From            string                 // Add From field for tracking form source
	Stats           *Stats                 // Add Stats field for dashboard statistics
	Departments     []Department           // Add Departments field for RBAC
	Roles           []Role                 // Add Roles field for RBAC
	Permissions     []Permission           // Add Permissions field for RBAC
	UserPermissions map[string][]string    // Add UserPermissions field for access control
	UserRoles       []string               // Add UserRoles field for role-based access
	UserFacility    int                    // Add UserFacility field for facility-based restrictions
	Outbreaks       []*models.Outbreak     // Correct type for outbreaks
	Users           []*models.EnhancedUser // Correct type for users
	Employee        *EmployeeForm          // Add Employee field for employee forms
	Titles          []EmployeeTitle        // Add Titles field for employee titles
	Facilities      []Facility             // Add Facilities field for facilities
	Employees       []EmployeeForm         // Add Employees field for employee lists
	Case            any                    // Add Case field for VHF case data
	Patient         any                    // Add Patient field for patient data
	Lab             any                    // Add Lab field for laboratory data
	ClinicalSigns   any                    // Add ClinicalSigns field for VHF clinical signs data
	Hospitalization any                    // Add Hospitalization field for VHF hospitalization data
	RiskFactors     any                    // Add RiskFactors field for VHF risk factors data
	Investigator    any                    // Add Investigator field for VHF investigator data
	Districts       []string               // Add Districts field for location dropdowns
	Subcounties     []string               // Add Subcounties field for location dropdowns
	Parishes        []string               // Add Parishes field for location dropdowns
	Villages        []string               // Add Villages field for location dropdowns

	// Custom fields for Mpox admission logic
	HasMpoxAdmission bool
	MpoxAdmissionID  int
	UserFacilityID   string // Add UserFacilityID field for default facility selection

	// Inventory fields
	InventoryStats          *InventoryStats
	InventoryAlerts         []*InventoryAlert
	InventoryTransactions   []*InventoryTransaction
	InventoryItems          []*InventoryItem
	InventoryItem           *InventoryItem
	InventoryCategories     []*InventoryCategory
	InventorySuppliers      []*InventorySupplier
	InventoryTreatmentSites []*InventoryTreatmentSite
	StockLevelReports       []*StockLevelReport
	TransactionReports      []*TransactionReport
	// Donation-related fields
	Donations              []*DonationSummary
	Donation               *InventoryDonation
	Donors                 []*InventoryDonor
	Donor                  *InventoryDonor
	DonationTypes          []*InventoryDonationType
	DonationType           *InventoryDonationType
	DonationItems          []*InventoryDonationItem
	DonationItem           *InventoryDonationItem
	DonationAcknowledgment *InventoryDonationAcknowledgment
	DonationStatistics     *DonationStatistics

	// Surveillance fields
	SurveillanceData *SurveillanceData
	StartDate        string
	EndDate          string

	// Resource Management fields
	ResourceStats      map[string]interface{}     // Resource management statistics
	RRTTeams           []*models.RRTTeam          // RRT teams
	RRTTeam            *models.RRTTeam            // Single RRT team
	RRTDeployments     []*models.RRTDeployment    // RRT deployments
	RRTDeployment      *models.RRTDeployment      // Single RRT deployment
	Resources          []*models.Resource         // Resources
	Resource           *models.Resource           // Single resource
	ResourceCategories []*models.ResourceCategory // Resource categories
	StorageLocations   []*models.StorageLocation  // Storage locations
	ResourceDonors     []*models.Donor            // Resource Management Donors
	Requisitions       []*models.Requisition      // Requisitions
	Requisition        *models.Requisition        // Single requisition
	Dispatches         []*models.Dispatch         // Dispatches
	ActivityLogs       []*models.ActivityLog      // Activity logs
	ActivityLog        *models.ActivityLog        // Single activity log
	GeneratedSitreps   []*models.GeneratedSitRep  // Generated SitReps
	GeneratedSitrep    *models.GeneratedSitRep    // Single generated SitRep

	// RRT Team Members fields
	RRTTeamMembers           []*models.RRTTeamMember           // RRT team members
	RRTTeamMember            *models.RRTTeamMember             // Single RRT team member
	RRTTeamMemberAssignments []*models.RRTTeamMemberAssignment // Team member assignments
	RRTTeamMemberAssignment  *models.RRTTeamMemberAssignment   // Single team member assignment
	RRTDeploymentProposals   []*models.RRTDeploymentProposal   // Deployment proposals
	RRTDeploymentProposal    *models.RRTDeploymentProposal     // Single deployment proposal
	RRTDeploymentExtensions  []*models.RRTDeploymentExtension  // Deployment extensions
	RRTDeploymentExtension   *models.RRTDeploymentExtension    // Single deployment extension
	RRTFieldRoleAssignments  []*models.RRTFieldRoleAssignment  // Field role assignments
	RRTFieldRoleAssignment   *models.RRTFieldRoleAssignment    // Single field role assignment

	// Pillars fields
	Pillars       []*models.Pillar       // Pillars
	Pillar        *models.Pillar         // Single pillar
	PillarChanges []*models.PillarChange // Pillar changes
	PillarChange  *models.PillarChange   // Single pillar change
}

// Department represents a department in the RBAC system
type Department struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Role represents a role in the RBAC system
type Role struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Selected    bool   `json:"selected"`
}

// Permission represents a permission in the RBAC system
type Permission struct {
	Resource string   `json:"resource"`
	Actions  []Action `json:"actions"`
}

// Action represents an action within a permission
type Action struct {
	Action  string `json:"action"`
	Granted bool   `json:"granted"`
}

// Stats contains dashboard statistics
type Stats struct {
	TotalUsers      int
	ActiveUsers     int
	LockedUsers     int
	TotalRoles      int
	TotalOutbreaks  int
	TotalCases      int
	TotalFacilities int
	TotalEmployees  int
}

// EmployeeForm represents an employee for form templates
type EmployeeForm struct {
	EmployeeID         int
	EmployeeFname      sql.NullString
	EmployeeLname      sql.NullString
	EmployeeSex        sql.NullString
	EmployeeEmail      sql.NullString
	EmployeePhone      sql.NullString
	EmployeeCadre      sql.NullString
	EmployeeTitle      sql.NullString
	EmployeeDepartment sql.NullString
	EmployeeSupervisor sql.NullInt64
	EmployeeStatus     sql.NullString
	EmployeeStartDate  sql.NullTime
	EmployeeEndDate    sql.NullTime
	EmployeePhotoURL   sql.NullString
	EmployeeNotes      sql.NullString
	Facility           sql.NullInt64
	AFIFacility        sql.NullString
	AFIRegion          sql.NullString
	AFIDistrict        sql.NullString
	FacilityInfo       struct {
		Name sql.NullString
	}
}

// EmployeeTitle represents an employee title
type EmployeeTitle struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Facility represents a facility
type Facility struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func NewTemplateData(c *fiber.Ctx, store *session.Store) *TemplateData {
	return NewTemplateDataWithDB(c, store, nil)
}

func NewTemplateDataWithDB(c *fiber.Ctx, store *session.Store, db *sql.DB) *TemplateData {
	//log.Printf("template data")

	// Get user's facility ID from session
	userFacilityID := ""
	userPermissions := make(map[string][]string)

	sess, err := store.Get(c)
	if err == nil {
		val := sess.Get("facility_id")
		switch v := val.(type) {
		case int:
			if v > 0 {
				userFacilityID = strconv.Itoa(v)
			}
		case int64:
			if v > 0 {
				userFacilityID = strconv.FormatInt(v, 10)
			}
		case float64:
			if v > 0 {
				userFacilityID = strconv.FormatInt(int64(v), 10)
			}
		case string:
			if v != "" {
				userFacilityID = v
			}
		}

		// Load user permissions if user is authenticated and database is available
		if IzAuthenticated(c, store) && db != nil {
			userID := GetCurrentUser(c, store)
			if userID > 0 {
				permissions, err := getUserPermissions(c, db, userID)
				if err == nil {
					userPermissions = permissions
				}
			}
		}
	}

	// Initialize Optionz with default values to ensure it's always accessible
	optionz := make(map[string]map[string]string)
	optionz["yn"] = map[string]string{"": " -- ", "1": "Yes", "2": "No"}
	optionz["yn_extra"] = map[string]string{"": " -- ", "1": "Yes", "2": "No", "3": "Unknown"}
	optionz["sex"] = map[string]string{"": " -- ", "1": "Male", "2": "Female"}
	optionz["marital"] = map[string]string{"": " -- ", "1": "Married", "2": "Cohabiting", "3": "Widowed", "4": "Separated", "5": "Divorced", "6": "Single"}
	optionz["nationality"] = map[string]string{"": " -- ", "1": "Ugandan", "2": "EAC", "3": "Other"}
	optionz["mental"] = map[string]string{"": " -- ", "a": "A", "v": "V", "p": "P", "u": "U"}
	optionz["preg"] = map[string]string{"": " -- ", "1": "Yes", "2": "No", "3": "ND"}
	optionz["ward"] = map[string]string{"": " -- ", "1": "Ward", "2": "IDU"}
	optionz["result1"] = map[string]string{"": " -- ", "1": "Pos", "2": "Neg", "3": "indeterminate"}
	optionz["result2"] = map[string]string{"": " -- ", "1": "Pos", "2": "Neg", "3": "ND"}

	// Get current outbreak ID from session
	outbreakID := GetCurrentOutbreak(c, store)

	// Get current page path for navigation context
	currentPath := c.Path()

	return &TemplateData{
		CurrentYear:     time.Now().Year(),
		IsAuthenticated: IzAuthenticated(c, store),
		OutbreakID:      outbreakID,
		IsOutbreakID:    outbreakID > 0,
		CurrentPath:     currentPath, // Add current path for navigation context
		Optionz:         optionz,
		UserFacilityID:  userFacilityID,
		UserPermissions: userPermissions,
		//CSRFToken:       c.Locals("csrf").(string), // Add the CSRF token.
	}
}

// Initialize a template.FuncMap object and store it in a global variable. This is
// essentially a string-keyed map which acts as a lookup between the names of my
// custom template functions and the functions themselves.
func CreateTemplateFunctions(c *fiber.Ctx, db *sql.DB) template.FuncMap {
	return template.FuncMap{
		"humanDate":            HumanDate,
		"humanDateTime":        HumanDateTime,
		"IsNullStringEmpty":    IsNullStringEmpty,
		"GetClientOptionLabel": GetClientOptionLabel,
		"seq":                  Seq,
		"GetOptionField":       GetOptionField,
		"title":                Title,
		"first":                First,
		"safe":                 func(s string) template.HTML { return template.HTML(s) },
		"mul":                  func(a, b float64) float64 { return a * b },
		// Convert values to int64 for templates
		"toInt64": func(v interface{}) int64 {
			switch t := v.(type) {
			case int:
				return int64(t)
			case int64:
				return t
			case float64:
				return int64(t)
			case string:
				if n, err := strconv.ParseInt(t, 10, 64); err == nil {
					return n
				}
				return 0
			default:
				return 0
			}
		},
		"GetDBOptions": func(table, cat, deflt, fld_name, fld_lab string, deflt_int int64) string {
			return GetDBOptions(c, db, table, cat, deflt, fld_name, fld_lab, deflt_int)
		},
		"GetFacilityOptions": func(deflt string, fld_name, fld_lab string) template.HTML {
			// Use the existing GetDBOptions with "site" table which queries the facility table
			// and return as template.HTML so it renders as HTML instead of escaped text
			return template.HTML(GetDBOptions(c, db, "site", "", deflt, fld_name, fld_lab, 0))
		},
		"GetDBLabel": func(table, namesFld, indexFld string, indexID int64) string {
			return GetDBLabel(c, db, table, namesFld, indexFld, indexID)
		},
		"hasPrefix": func(s, prefix string) bool {
			return strings.HasPrefix(s, prefix)
		},
		"hasPermission": func(permissions map[string][]string, resource, action string) bool {
			if permissions == nil {
				return false
			}
			actions, exists := permissions[resource]
			if !exists {
				return false
			}
			for _, a := range actions {
				if a == action {
					return true
				}
			}
			return false
		},
		"getOptionzValue": func(optionz map[string]map[string]string, key1, key2 string) string {
			if optionz == nil {
				log.Printf("DEBUG: getOptionzValue called with nil optionz")
				return ""
			}
			if subMap, exists := optionz[key1]; exists {
				if value, found := subMap[key2]; found {
					return value
				}
			}
			return ""
		},
		"safeOptionz": func(optionz map[string]map[string]string) map[string]map[string]string {
			if optionz == nil {
				return make(map[string]map[string]string)
			}
			return optionz
		},
		"isSlice": func(v interface{}) bool {
			if v == nil {
				return false
			}
			val := reflect.ValueOf(v)
			return val.Kind() == reflect.Slice
		},
		"toJSON": func(v interface{}) template.JS {
			if v == nil {
				return template.JS("null")
			}
			// Simple primitives inline
			switch val := v.(type) {
			case string:
				return template.JS(fmt.Sprintf(`"%s"`, val))
			case int, int64, float64, bool:
				return template.JS(fmt.Sprintf("%v", val))
			}
			// Fallback: JSON marshal any complex type (maps, slices, structs)
			b, err := json.Marshal(v)
			if err != nil {
				return template.JS("null")
			}
			return template.JS(string(b))
		},
		"renderSymptomRow": func(name, label string, value sql.NullBool, date sql.NullTime, duration sql.NullInt32) template.HTML {
			var buf bytes.Buffer
			tmpl := `
			<tr>
				<td>{{.Label}}</td>
				<td>
					<div class="d-flex gap-3">
						<div class="form-check">
							<input class="form-check-input" type="radio" name="{{.Name}}" value="true" {{if and .Value.Valid .Value.Bool}}checked{{end}}>
							<label class="form-check-label">Yes</label>
						</div>
						<div class="form-check">
							<input class="form-check-input" type="radio" name="{{.Name}}" value="false" {{if and .Value.Valid (not .Value.Bool)}}checked{{end}}>
							<label class="form-check-label">No</label>
						</div>
						<div class="form-check">
							<input class="form-check-input" type="radio" name="{{.Name}}" value="unknown" {{if not .Value.Valid}}checked{{end}}>
							<label class="form-check-label">Unknown</label>
						</div>
					</div>
				</td>
				<td>
					<input type="date" class="form-control" name="{{.Name}}_date" value="{{if .Date.Valid}}{{.Date.Time.Format "2006-01-02"}}{{end}}">
				</td>
				<td>
					<input type="number" class="form-control" name="{{.Name}}_duration" min="0" value="{{if .Duration.Valid}}{{.Duration.Int32}}{{end}}">
				</td>
			</tr>`
			t := template.Must(template.New("symptomRow").Parse(tmpl))
			data := struct {
				Name     string
				Label    string
				Value    sql.NullBool
				Date     sql.NullTime
				Duration sql.NullInt32
			}{
				Name:     name,
				Label:    label,
				Value:    value,
				Date:     date,
				Duration: duration,
			}
			if err := t.Execute(&buf, data); err != nil {
				return template.HTML(fmt.Sprintf("Error rendering symptom row: %v", err))
			}
			return template.HTML(buf.String())
		},
		// Alerts template functions
		"getSeverityColor": func(severity string) string {
			switch strings.ToLower(severity) {
			case "critical", "high":
				return "danger"
			case "medium":
				return "warning"
			case "low":
				return "info"
			default:
				return "secondary"
			}
		},
		"getSeverityIcon": func(severity string) string {
			switch strings.ToLower(severity) {
			case "critical":
				return "fa-exclamation-triangle"
			case "high":
				return "fa-exclamation-circle"
			case "medium":
				return "fa-exclamation"
			case "low":
				return "fa-info-circle"
			default:
				return "fa-bell"
			}
		},
		"getStatusColor": func(status string) string {
			switch strings.ToLower(status) {
			case "active":
				return "danger"
			case "resolved":
				return "success"
			case "pending":
				return "warning"
			default:
				return "secondary"
			}
		},
		"formatDate": func(t time.Time) string {
			return t.Format("02 Jan 2006 15:04")
		},
		"formatDateOnly": func(t time.Time) string {
			return t.Format("1/2/2006")
		},
		"formatTimeOnly": func(t time.Time) string {
			return t.Format("3:04:05 PM")
		},
		"now": func() time.Time {
			return time.Now()
		},
		"date": func(format string, t time.Time) string {
			return t.Format(format)
		},
		"getSourceColor": func(source string) string {
			switch strings.ToLower(source) {
			case "community":
				return "info"
			case "facility":
				return "primary"
			default:
				return "secondary"
			}
		},
		"getResponseColor": func(response string) string {
			switch strings.ToLower(response) {
			case "mpox":
				return "warning"
			case "evd":
				return "danger"
			default:
				return "secondary"
			}
		},
	}
}

func HumanDate(t time.Time) string {
	return t.Format("02 Jan 2006")
}

func HumanDateTime(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

func Seq(start, end int) []int {
	s := make([]int, end-start+1)
	for i := range s {
		s[i] = start + i
	}
	return s
}

// Title capitalizes the first letter of each word in a string
func Title(s string) string {
	if s == "" {
		return s
	}

	// Split by spaces and capitalize first letter of each word
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}

// First returns the first character of a string
func First(s string) string {
	if s == "" {
		return ""
	}
	return string(s[0])
}

func IsNullStringEmpty(nullable sql.NullString) bool {
	return !nullable.Valid || nullable.String == ""
}

// GenerateHTML renders an HTML template with given data
func GenerateHTML(c *fiber.Ctx, db *sql.DB, zdata interface{}, filenames ...string) error {
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

	// Log the path for debugging
	log.Printf("Using template directory: %s", basePath)

	// First, load the layout template
	layoutFile := filepath.Join(basePath, "layout.html")
	layoutContent, err := os.ReadFile(layoutFile)
	if err != nil {
		return c.Status(500).SendString(fmt.Sprintf("Failed to read layout template: %v", err))
	}

	// Create a new template with the layout
	templates := template.New("layout").Funcs(CreateTemplateFunctions(c, db))

	// Parse the layout template first
	_, err = templates.Parse(string(layoutContent))
	if err != nil {
		return c.Status(500).SendString(fmt.Sprintf("Failed to parse layout template: %v", err))
	}

	// Always include the navigation template
	navigationFile := filepath.Join(basePath, "navigation.html")
	if _, err := os.Stat(navigationFile); err == nil {
		log.Printf("Loading navigation template: %s", navigationFile)
		navContent, err := os.ReadFile(navigationFile)
		if err != nil {
			return c.Status(500).SendString(fmt.Sprintf("Failed to read navigation template: %v", err))
		}
		_, err = templates.Parse(string(navContent))
		if err != nil {
			return c.Status(500).SendString(fmt.Sprintf("Failed to parse navigation template: %v", err))
		}
	} else {
		log.Printf("Navigation template not found: %s", navigationFile)
	}

	// Add the requested templates
	for _, file := range filenames {
		// Ensure we have the full path with .html extension
		filePath := filepath.Join(basePath, file+".html")
		log.Printf("Loading template file: %s", filePath)

		// Check if file exists
		if _, err := os.Stat(filePath); err != nil {
			return c.Status(500).SendString(fmt.Sprintf("Template file not found: %s (error: %v)", filePath, err))
		}

		// Read file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			return c.Status(500).SendString(fmt.Sprintf("Failed to read template file %s: %v", filePath, err))
		}

		// Parse the template content
		_, err = templates.Parse(string(content))
		if err != nil {
			return c.Status(500).SendString(fmt.Sprintf("Failed to parse template %s: %v", filePath, err))
		}
	}

	// Execute template and write output
	c.Set("Content-Type", "text/html")

	// Debug logging
	log.Printf("DEBUG: Executing template with data type: %T", zdata)
	if templateData, ok := zdata.(*TemplateData); ok {
		log.Printf("DEBUG: TemplateData.Optionz initialized: %v, keys: %d", templateData.Optionz != nil, len(templateData.Optionz))
		if templateData.Optionz != nil {
			for key, value := range templateData.Optionz {
				log.Printf("DEBUG: Optionz[%s] has %d items", key, len(value))
			}
		} else {
			log.Printf("DEBUG: TemplateData.Optionz is nil - this might cause issues")
		}
		log.Printf("DEBUG: TemplateData.UserPermissions initialized: %v, keys: %d", templateData.UserPermissions != nil, len(templateData.UserPermissions))
		if templateData.UserPermissions != nil {
			for resource, actions := range templateData.UserPermissions {
				log.Printf("DEBUG: UserPermissions[%s] has actions: %v", resource, actions)
			}
		}
	} else {
		log.Printf("DEBUG: Data is not *TemplateData, actual type: %T", zdata)
		// Try to convert to TemplateData if possible
		if reflect.TypeOf(zdata).Kind() == reflect.Ptr {
			log.Printf("DEBUG: Data is a pointer to: %T", reflect.ValueOf(zdata).Elem().Interface())
		}
		// Log the actual value for debugging
		log.Printf("DEBUG: Data value: %+v", zdata)
	}

	if err := templates.ExecuteTemplate(c.Response().BodyWriter(), "layout", zdata); err != nil {
		return c.Status(500).SendString(fmt.Sprintf("Template execution error: %v", err))
	}

	return nil
}

func GetPath(topFile string) (rtn string) {
	var dirAbsPath string

	ex, err := os.Executable()
	if err == nil {
		dirAbsPath = filepath.Dir(ex)
		rtn = dirAbsPath + "/" + topFile
	} else {
		rtn = topFile
	}

	return
}

func GetParent() (parentPath string) {
	parentPath = filepath.Join("..", "..")
	return
}

type NullableString struct {
	sql.NullString
}

func (ns *NullableString) UnmarshalText(text []byte) error {
	ns.Valid = len(text) > 0
	ns.String = string(text)
	return nil
}

type NullableFloat64 struct {
	sql.NullFloat64
}

type NullableTime struct {
	sql.NullTime
}

func (nf *NullableTime) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		nf.Valid = false
		return nil
	}

	// Try to parse using a standard format (adjust based on the expected input)
	n, err := time.Parse("2006-01-02 15:04:05", string(text))
	if err != nil {
		return err
	}

	nf.Time = n
	nf.Valid = true
	return nil
}

func (nf *NullableFloat64) UnmarshalText(text []byte) error {
	nf.Valid = len(text) > 0
	n, err := strconv.ParseFloat(string(text), 64)
	if err != nil {
		return err
	}
	nf.Float64 = n
	return nil
}

type NullableInt64 struct {
	sql.NullInt64
}

func (ni *NullableInt64) UnmarshalText(text []byte) error {
	ni.Valid = len(text) > 0
	n, err := strconv.ParseInt(string(text), 10, 64)
	if err != nil {
		return err
	}
	ni.Int64 = n
	return nil
}

// DecodeFormData decodes form data into a given struct
func DecodeFormData(c *fiber.Ctx, v interface{}) error {
	// Get all form data and convert it to map[string][]string
	//formData := make(map[string][]string)
	/*
		c.Request().PostArgs().VisitAll(func(key, value []byte) {
			formData[string(key)] = []string{string(value)}
		})
	*/

	postArgs := c.Context().PostArgs()

	// Create a map to store form data
	formData := make(map[string][]string)

	// Visit all arguments and store them in formData map
	postArgs.VisitAll(func(key, value []byte) {
		formData[string(key)] = []string{string(value)}
	})

	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true) // Ignore unknown fields

	// Register converters for sql.Null* types
	decoder.RegisterConverter(sql.NullString{}, func(s string) reflect.Value {
		var ns NullableString
		if err := ns.UnmarshalText([]byte(s)); err != nil {
			return reflect.ValueOf(sql.NullString{})
		}
		return reflect.ValueOf(ns.NullString)
	})

	decoder.RegisterConverter(sql.NullFloat64{}, func(s string) reflect.Value {
		var nf NullableFloat64
		if err := nf.UnmarshalText([]byte(s)); err != nil {
			return reflect.ValueOf(sql.NullFloat64{})
		}
		return reflect.ValueOf(nf.NullFloat64)
	})

	decoder.RegisterConverter(sql.NullTime{}, func(s string) reflect.Value {
		var nf NullableTime
		if err := nf.UnmarshalText([]byte(s)); err != nil {
			return reflect.ValueOf(sql.NullTime{})
		}
		return reflect.ValueOf(nf.NullTime)
	})

	decoder.RegisterConverter(sql.NullInt64{}, func(s string) reflect.Value {
		var ni NullableInt64
		if err := ni.UnmarshalText([]byte(s)); err != nil {
			return reflect.ValueOf(sql.NullInt64{})
		}
		return reflect.ValueOf(ni.NullInt64)
	})

	// Decode form data into the struct
	if err := decoder.Decode(v, formData); err != nil {
		fmt.Println("Decoding error:", err)
		log.Println("Error retrieving session: ", err.Error())
		return err
	}

	return nil
}

// GetCurrentUser retrieves the current user ID from the session
func GetCurrentUser(c *fiber.Ctx, store *session.Store) int {
	sess, err := store.Get(c)
	if err != nil {
		fmt.Println("Error retrieving session:", err)
		log.Println("Error retrieving session: ", err.Error())
		return 0
	}

	// Try to get userID from session with multiple type handling
	userIDVal := sess.Get("user")
	if userIDVal == nil {
		// Try alternative key
		userIDVal = sess.Get("user_id")
	}

	if userIDVal == nil {
		fmt.Println("User not found in session")
		log.Println("User not found in session")
		return 0
	}

	switch v := userIDVal.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i
		} else {
			fmt.Println("Failed to convert session userID string to int:", err)
			log.Println("Failed to convert session userID string to int:", err)
			return 0
		}
	default:
		fmt.Println("Unknown session userID type:", fmt.Sprintf("%T", userIDVal))
		log.Println("Unknown session userID type:", fmt.Sprintf("%T", userIDVal))
		return 0
	}
}

// GetUserFacility gets the facility ID assigned to a user from the employee table
func GetUserFacility(c *fiber.Ctx, db *sql.DB, userID int) (int, error) {
	var facilityID sql.NullInt64
	err := db.QueryRowContext(c.Context(),
		"SELECT facility_id FROM employees WHERE user_id = $1 AND is_active = true",
		userID).Scan(&facilityID)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // No facility assigned
		}
		return 0, err
	}

	if facilityID.Valid {
		return int(facilityID.Int64), nil
	}
	return 0, nil
}

// GetCurrentOutbreak retrieves the current outbreak ID from the session
func GetCurrentOutbreak(c *fiber.Ctx, store *session.Store) int {
	sess, err := store.Get(c)
	if err != nil {
		fmt.Println("Error retrieving session:", err)
		log.Println("Error retrieving session: ", err.Error())
		return 0
	}
	// Try both keys
	for _, key := range []string{"outbreak_id", "selected_outbreak"} {
		val := sess.Get(key)
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		case string:
			i, _ := strconv.Atoi(v)
			return i
		}
	}
	// No outbreak selected - this is normal, not an error
	return 0
}

// IsAuthenticated middleware checks if a user is authenticated
func IsAuthenticated(store *session.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			log.Println("Error retrieving session:", err)
			return c.Redirect("/login", fiber.StatusFound)
		}

		if sess.Get("isAuthenticated") == nil {
			log.Println("No active session")
			return c.Redirect("/login", fiber.StatusFound)
		}

		return c.Next()
	}
}

func IzAuthenticated(c *fiber.Ctx, store *session.Store) bool {

	sess, err := store.Get(c)
	if err != nil {
		log.Println("Error retrieving session:", err)
		return false
	}

	if sess.Get("isAuthenticated") == nil {
		log.Println("No active session")
		return false
	}

	return true
}

func ParseNullString(value string) sql.NullString {
	if value == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: value, Valid: true}
}

func ParseNullInt(value string) sql.NullInt64 {
	if value == "" {
		return sql.NullInt64{Valid: false}
	}
	var i int64
	_, err := fmt.Sscanf(value, "%d", &i)
	if err != nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: i, Valid: true}
}

func ParseNullFloat(value string) sql.NullFloat64 {
	if value == "" {
		return sql.NullFloat64{Valid: false}
	}
	var f float64
	_, err := fmt.Sscanf(value, "%f", &f)
	if err != nil {
		return sql.NullFloat64{Valid: false}
	}
	return sql.NullFloat64{Float64: f, Valid: true}
}

func ParseNullTime(value string) sql.NullTime {
	if value == "" {
		return sql.NullTime{Valid: false}
	}
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: t, Valid: true}
}

// Convert interface{} to sql.NullInt64
func ParseNullInt2(value interface{}) sql.NullInt64 {
	if value == nil {
		return sql.NullInt64{Valid: false}
	}

	switch v := value.(type) {
	case float64: // JSON numbers are decoded as float64
		return sql.NullInt64{Int64: int64(v), Valid: true}
	case string:
		if num, err := strconv.ParseInt(v, 10, 64); err == nil {
			return sql.NullInt64{Int64: num, Valid: true}
		}
	}

	return sql.NullInt64{Valid: false}
}

// Convert interface{} to sql.NullString
func ParseNullString2(value interface{}) sql.NullString {
	if value == nil {
		return sql.NullString{Valid: false}
	}

	if str, ok := value.(string); ok && str != "" {
		return sql.NullString{String: str, Valid: true}
	}

	return sql.NullString{Valid: false}
}

// ConvertFiberToGin converts a Fiber context to a Gin context
func ConvertFiberToGin(fctx *fiber.Ctx) (*gin.Context, error) {
	// Create a new HTTP request using Fiber's request data
	req := fctx.Request()

	// Convert Fiber request to standard *http.Request
	httpReq, err := http.NewRequest(
		string(req.Header.Method()),
		fctx.OriginalURL(),
		bytes.NewReader(req.Body()),
	)
	if err != nil {
		return nil, err
	}

	// Copy headers from Fiber to the new request
	req.Header.VisitAll(func(key, value []byte) {
		httpReq.Header.Set(string(key), string(value))
	})

	// Create a new Gin response recorder
	w := httptest.NewRecorder()

	// Create a new Gin context
	ginCtx, _ := gin.CreateTestContext(w)
	ginCtx.Request = httpReq

	return ginCtx, nil
}

func DoZaLogging(typ, msg string, er error) {
	switch typ {
	case "ERROR":
		log.Printf("ERROR: %s - %s", msg, er)
		fmt.Printf("ERROR: %s - %s", msg, er)
	case "INFO":
		log.Printf("INFO: %s", msg)
		fmt.Printf("INFO: %s", msg)
	case "WARNING":
		log.Printf("WARNING: %s", msg)
		fmt.Printf("WARNING: %s", msg)
	default:
		log.Printf("UNKNOWN: %s - %s", msg, er)
		fmt.Printf("UNKNOWN: %s - %s", msg, er)
	}

}

func GetOptionField(table, field, labs, defaultString string, defaultvalue, whole int64) string {
	zaDefa1 := ""
	zaDefa2 := ""
	zaDefa3 := ""
	optionz := ""

	// Get the options map
	options := Get_Client_Optionz()

	// Handle different option types
	switch table {
	case "facility":
		optionz = `<option value=""> -- select -- </option>`

	case "Status":
		// Handle both string and sql.NullString types
		statusValue := defaultString
		if statusValue == "Suspect" {
			zaDefa1 = "selected"
		}
		if statusValue == "Case" {
			zaDefa2 = "selected"
		}
		if statusValue == "Other" {
			zaDefa3 = "selected"
		}

		optionz = `<option value=""> -- select -- </option>
					<option value="Suspect" ` + zaDefa1 + `>Suspect</option>
					<option value="Case" ` + zaDefa2 + `>Case</option>
					<option value="Other" ` + zaDefa3 + `>Other</option>`

	case "encounter_type":
		// Encounter type options
		optionz = `<option value=""> -- select -- </option>
					<option value="1" ` + zaDefa1 + `>Initial</option>
					<option value="2" ` + zaDefa2 + `>Follow-up</option>
					<option value="3" ` + zaDefa3 + `>Discharge</option>`

	case "yn", "YN":
		// Yes/No options - handle both cases
		if defaultvalue == 1 {
			zaDefa1 = "selected"
		} else if defaultvalue == 2 {
			zaDefa2 = "selected"
		}
		optionz = `<option value=""> -- select -- </option>
					<option value="1" ` + zaDefa1 + `>Yes</option>
					<option value="2" ` + zaDefa2 + `>No</option>`

	case "posx":
		// Positive/Negative/Indeterminate options
		if defaultvalue == 1 {
			zaDefa1 = "selected"
		} else if defaultvalue == 2 {
			zaDefa2 = "selected"
		} else if defaultvalue == 3 {
			zaDefa3 = "selected"
		}
		optionz = `<option value=""> -- select -- </option>
					<option value="1" ` + zaDefa1 + `>Positive</option>
					<option value="2" ` + zaDefa2 + `>Negative</option>
					<option value="3" ` + zaDefa3 + `>Indeterminate</option>`

	case "po":
		// Positive/Negative options
		if defaultvalue == 1 {
			zaDefa1 = "selected"
		} else if defaultvalue == 2 {
			zaDefa2 = "selected"
		}
		optionz = `<option value=""> -- select -- </option>
					<option value="1" ` + zaDefa1 + `>Positive</option>
					<option value="2" ` + zaDefa2 + `>Negative</option>`

	case "pos":
		// Positive/Negative/ND options
		if defaultvalue == 1 {
			zaDefa1 = "selected"
		} else if defaultvalue == 2 {
			zaDefa2 = "selected"
		} else if defaultvalue == 3 {
			zaDefa3 = "selected"
		}
		optionz = `<option value=""> -- select -- </option>
					<option value="1" ` + zaDefa1 + `>Positive</option>
					<option value="2" ` + zaDefa2 + `>Negative</option>
					<option value="3" ` + zaDefa3 + `>ND</option>`

	case "blood":
		// Blood gas options
		if defaultvalue == 1 {
			zaDefa1 = "selected"
		} else if defaultvalue == 2 {
			zaDefa2 = "selected"
		} else if defaultvalue == 3 {
			zaDefa3 = "selected"
		}
		optionz = `<option value=""> -- select -- </option>
					<option value="1" ` + zaDefa1 + `>Normal</option>
					<option value="2" ` + zaDefa2 + `>Abnormal</option>
					<option value="3" ` + zaDefa3 + `>ND</option>`

	case "e_rdt":
		// Ebola RDT options
		if defaultvalue == 1 {
			zaDefa1 = "selected"
		} else if defaultvalue == 2 {
			zaDefa2 = "selected"
		} else if defaultvalue == 3 {
			zaDefa3 = "selected"
		}
		optionz = `<option value=""> -- select -- </option>
					<option value="1" ` + zaDefa1 + `>Standard</option>
					<option value="2" ` + zaDefa2 + `>Rapid</option>
					<option value="3" ` + zaDefa3 + `>Other</option>`

	case "e_pcr":
		// Ebola PCR options
		if defaultvalue == 1 {
			zaDefa1 = "selected"
		} else if defaultvalue == 2 {
			zaDefa2 = "selected"
		} else if defaultvalue == 3 {
			zaDefa3 = "selected"
		}
		optionz = `<option value=""> -- select -- </option>
					<option value="1" ` + zaDefa1 + `>Standard</option>
					<option value="2" ` + zaDefa2 + `>Real-time</option>
					<option value="3" ` + zaDefa3 + `>Other</option>`

	default:
		// Try to get from the options map
		if optMap, exists := options[table]; exists {
			optionz = `<option value=""> -- select -- </option>`
			for key, value := range optMap {
				if key != "" { // Skip empty key
					selected := ""
					if strconv.FormatInt(defaultvalue, 10) == key {
						selected = "selected"
					}
					optionz += `<option value="` + key + `" ` + selected + `>` + value + `</option>`
				}
			}
		} else {
			// Fallback for unknown option types
			optionz = `<option value=""> -- select -- </option>`
		}
	}

	if whole == 1 {
		return `<select class="form-control-sm patient-input form-select" name="` + field + `" id="` + field + `" aria-label="` + labs + `">
				` + optionz + `
				</select>`
	}
	return optionz
}

// GetClientOptionLabel returns the label for a given option key
func GetClientOptionLabel(arrayKey, mapKey string) string {
	options := Get_Client_Optionz()
	if subMap, exists := options[arrayKey]; exists {
		if label, found := subMap[mapKey]; found {
			return label
		}
	}
	return ""
}

// GetDBOptions returns HTML select options for database values
func GetDBOptions(c *fiber.Ctx, db *sql.DB, table, cat, deflt, fld_name, fld_lab string, deflt_int int64) string {
	sql := ""
	rtn := ""
	switch table {
	case "Employee":
		sql = "SELECT employee_id as code, CONCAT(employee_fname, ' ', employee_lname) AS lab FROM public.employee"
	case "function":
		sql = "SELECT function_id as code, function_name as lab FROM public.function"
	case "site":
		sql = "SELECT facility_id as code, facility_name as lab FROM public.facility ORDER BY facility_name ASC"
	case "test":
		// Handle test case
	case "meta":
		sql = "Select meta_id as code, meta_name as lab from meta, meta_category WHERE meta.meta_category=meta_category.meta_category_id AND meta_category_name='" + cat + "'"
	}

	res, er := models.GetFields(c.Context(), db, sql)
	if er != nil {
		log.Println("Error getting fields:", er)
		return ""
	}

	i := 0
	for _, values := range res {
		if deflt == "" {
			zvalue, _ := strconv.ParseInt(values[0], 10, 64)
			if zvalue == deflt_int {
				rtn = rtn + fmt.Sprintf(`<option value="%s" selected>%s</option>`, values[0], values[1])
			} else {
				rtn = rtn + fmt.Sprintf(`<option value="%s">%s</option>`, values[0], values[1])
			}
		} else {
			if values[0] == deflt {
				rtn = rtn + fmt.Sprintf(`<option value="%s" selected>%s</option>`, values[0], values[1])
			} else {
				rtn = rtn + fmt.Sprintf(`<option value="%s">%s</option>`, values[0], values[1])
			}
		}
		i++
	}

	addString := ""
	if i > 1 {
		addString = `<option value=""> -- </option>`
	}

	return `<select class="form-control-sm patient-input form-select" name="` + fld_name + `" id="` + fld_name + `" aria-label="` + fld_lab + `">
			` + addString + rtn + `
			</select>`
}

// GetDBLabel returns the label for a database value
func GetDBLabel(c *fiber.Ctx, db *sql.DB, table, namesFld, indexFld string, indexID int64) string {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", namesFld, table, indexFld)
	var label sql.NullString
	err := db.QueryRowContext(c.Context(), query, indexID).Scan(&label)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("No rows found for ID:", indexID)
			return ""
		}
		log.Println("Error executing query:", err)
		return ""
	}
	if !label.Valid {
		return ""
	}
	return label.String
}

// Get_Client_Optionz returns a map of client options
func Get_Client_Optionz() map[string]map[string]string {
	opt := make(map[string]map[string]string)
	opt["sex"] = map[string]string{"": " -- ", "1": "Male", "2": "Female"}
	opt["sex2"] = map[string]string{"": " -- ", "Male": "Male", "Female": "Female"}
	opt["occup"] = map[string]string{"": " -- ", "1": "Healthcare worker", "2": "Non-Healthcare worker"}
	opt["yn"] = map[string]string{"": " -- ", "1": "Yes", "2": "No"}
	opt["yn_extra"] = map[string]string{"": " -- ", "1": "Yes", "2": "No", "3": "Unknown"}
	opt["marital"] = map[string]string{"": " -- ",
		"1": "Married",
		"2": "Cohabiting",
		"3": "Widowed",
		"4": "Separated",
		"5": "Divorced",
		"6": "Single",
	}
	opt["nationality"] = map[string]string{"": " -- ", "1": "Ugandan", "2": "EAC", "3": "Other"}
	opt["ethnicity"] = map[string]string{"": " -- ", "1": "Black", "2": "Other"}
	opt["mental"] = map[string]string{"": " -- ", "a": "A", "v": "V", "p": "P", "u": "U"}
	opt["preg"] = map[string]string{"": " -- ", "1": "Yes", "2": "No", "3": "ND"}
	opt["ward"] = map[string]string{"": " -- ", "1": "Ward", "2": "Infectious Diseases Unit", "3": "Home Based Care"}
	opt["result1"] = map[string]string{"": " -- ", "1": "Pos", "2": "Neg", "3": "indeterminate"}
	opt["result2"] = map[string]string{"": " -- ", "1": "Pos", "2": "Neg", "3": "ND"}
	return opt
}

// GetUser returns the current user ID and username
func GetUser(c *fiber.Ctx, sl *slog.Logger, store *session.Store) (int, string) {
	sess, err := store.Get(c)
	if err != nil {
		sl.Info("Session error")
		return 0, ""
	}

	// Try to get userID from session with multiple type handling
	userIDVal := sess.Get("user")
	sl.Info("Session debug", "user_key_value", userIDVal, "user_key_type", fmt.Sprintf("%T", userIDVal))

	if userIDVal == nil {
		// Try alternative key
		userIDVal = sess.Get("user_id")
		sl.Info("Session debug", "user_id_key_value", userIDVal, "user_id_key_type", fmt.Sprintf("%T", userIDVal))
	}

	var userID int
	switch v := userIDVal.(type) {
	case int:
		userID = v
	case int64:
		userID = int(v)
	case float64:
		userID = int(v)
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			userID = i
		} else {
			sl.Error("Failed to convert session userID string to int", "error", err, "value", v)
			return 0, ""
		}
	default:
		sl.Error("Unknown session userID type", "type", fmt.Sprintf("%T", userIDVal), "value", userIDVal)
		return 0, ""
	}

	username, ok := sess.Get("username").(string)
	if !ok {
		sl.Error("Failed to convert session username to string", "value", sess.Get("username"))
		return 0, ""
	}

	return userID, username
}

// GetCurrentFacility returns the current facility ID for the user
func GetCurrentFacility(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store) int {
	userID := GetCurrentUser(c, store)
	if userID == 0 {
		return 0
	}

	// First try to get from session
	sess, err := store.Get(c)
	if err == nil {
		val := sess.Get("facility_id")
		switch v := val.(type) {
		case int:
			if v > 0 {
				return v
			}
		case int64:
			if v > 0 {
				return int(v)
			}
		case float64:
			if v > 0 {
				return int(v)
			}
		case string:
			if i, err := strconv.Atoi(v); err == nil && i > 0 {
				return i
			}
		}
	}

	// Fallback to database query
	var facilityID sql.NullInt64
	err = db.QueryRowContext(c.Context(), `
		SELECT e.facility 
		FROM employee e
		JOIN users u ON e.employee_id = u.user_employee
		WHERE u.user_id = $1 AND e.employee_status = 'active'
	`, userID).Scan(&facilityID)
	if err != nil {
		if err != sql.ErrNoRows {
			sl.Error("Error getting facility", "error", err)
		}
		return 0
	}

	if !facilityID.Valid {
		return 0
	}
	return int(facilityID.Int64)
}

// GetDBInt retrieves an integer value from the database based on the given parameters
func GetDBInt(table, field, cat, filter string, defaultValue int) int {
	if dbG == nil || CtxG == nil {
		return defaultValue
	}

	query := fmt.Sprintf("SELECT %s FROM %s", field, table)
	if cat != "" {
		query += " WHERE " + cat
	}
	if filter != "" {
		if cat != "" {
			query += " AND " + filter
		} else {
			query += " WHERE " + filter
		}
	}

	var value int
	err := dbG.QueryRowContext(CtxG.Context(), query).Scan(&value)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Error getting DB int: %v", err)
		}
		return defaultValue
	}
	return value
}

type MeaslesCaseInvestigation struct {
	ID          int
	EpidNo      string
	LabNo       string
	VisitDate   string
	HealthUnit  string
	District    string
	PatientType string
	PatientNo   string
	// ... all other fields as per schema above
}
