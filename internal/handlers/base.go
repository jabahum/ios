package handlers

import (
	"bytes"
	"case/internal/models"
	"database/sql"
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
	CSRFToken       string // Add a CSRFToken field.

	// Custom fields for Mpox admission logic
	HasMpoxAdmission bool
	MpoxAdmissionID  int
}

func NewTemplateData(c *fiber.Ctx, store *session.Store) *TemplateData {
	//log.Printf("template data")
	return &TemplateData{
		CurrentYear:     time.Now().Year(),
		IsAuthenticated: IzAuthenticated(c, store),
		//CSRFToken:       c.Locals("csrf").(string), // Add the CSRF token.
	}
}

// Initialize a template.FuncMap object and store it in a global variable. This is
// essentially a string-keyed map which acts as a lookup between the names of our
// custom template functions and the functions themselves.
func CreateTemplateFunctions(c *fiber.Ctx, db *sql.DB) template.FuncMap {
	return template.FuncMap{
		"humanDate":            HumanDate,
		"humanDateTime":        HumanDateTime,
		"IsNullStringEmpty":    IsNullStringEmpty,
		"GetClientOptionLabel": GetClientOptionLabel,
		"seq":                  Seq,
		"GetOptionField":       GetOptionField,
		"GetDBOptions": func(table, cat, deflt, fld_name, fld_lab string, deflt_int int64) string {
			return GetDBOptions(c, db, table, cat, deflt, fld_name, fld_lab, deflt_int)
		},
		"GetDBLabel": func(table, namesFld, indexFld string, indexID int64) string {
			return GetDBLabel(c, db, table, namesFld, indexFld, indexID)
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

	// Try to parse using a standard format (adjust based on your expected input)
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

	userID := sess.Get("user")
	userInt, ok := userID.(int)
	if !ok {
		fmt.Println("User not found or not an int")
		log.Println("User not found or not an int")
		return 0
	}
	return userInt
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

	if table == "facility" {
		if defaultvalue == 1 {
			zaDefa1 = "selected"
		}
		if defaultvalue == 2 {
			zaDefa2 = "selected"
		}
		if defaultvalue == 3 {
			zaDefa3 = "selected"
		}

		optionz = `<option value=""> -- select -- </option>
					<option value="1" ` + zaDefa1 + `>Mulago ETU</option>
					<option value="2" ` + zaDefa2 + `>Mbale ETU</option>
					<option value="3" ` + zaDefa3 + `>Fort Portal ETU</option>`
	}

	if table == "Status" {
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
		sql = "SELECT facility_id as code, facility_name as lab FROM public.facility"
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
	var label string
	err := db.QueryRowContext(c.Context(), query, indexID).Scan(&label)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("No rows found for ID:", indexID)
			return ""
		}
		log.Println("Error executing query:", err)
		return ""
	}
	return label
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
	opt["ward"] = map[string]string{"": " -- ", "1": "Ward", "2": "ICU"}
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

	userID, ok := sess.Get("user").(int)
	if !ok {
		fmt.Println("Failed to convert session value to int")
		return 0, ""
	}

	username, ok := sess.Get("username").(string)
	if !ok {
		fmt.Println("Failed to convert session value to string")
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

	var facilityID sql.NullInt64
	err := db.QueryRowContext(c.Context(), "SELECT facility FROM public.employee WHERE employee_id = $1", userID).Scan(&facilityID)
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
