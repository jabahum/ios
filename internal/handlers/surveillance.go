package handlers

import (
	"case/internal/config"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// DHIS2Event represents a single event from the DHIS2 API
type DHIS2Event struct {
	Event         string                 `json:"event"`
	EventDate     string                 `json:"eventDate"`
	Status        string                 `json:"status"`
	OrgUnit       string                 `json:"orgUnit"`
	OrgUnitName   string                 `json:"orgUnitName"`
	Program       string                 `json:"program"`
	ProgramStage  string                 `json:"programStage"`
	DataValues    []DHIS2DataValue       `json:"dataValues"`
	Attributes    map[string]interface{} `json:"attributes"`
	Geometry      map[string]interface{} `json:"geometry"`
	Created       string                 `json:"created"`
	LastUpdated   string                 `json:"lastUpdated"`
	StoredBy      string                 `json:"storedBy"`
	Coordinate    map[string]interface{} `json:"coordinate"`
	FollowUp      bool                   `json:"followUp"`
	Deleted       bool                   `json:"deleted"`
	AssignedUser  string                 `json:"assignedUser"`
	CompletedDate string                 `json:"completedDate"`
	CompletedBy   string                 `json:"completedBy"`
}

// DHIS2DataValue represents a data value within an event
type DHIS2DataValue struct {
	DataElement       string `json:"dataElement"`
	Value             string `json:"value"`
	ProvidedElsewhere bool   `json:"providedElsewhere"`
	StoredBy          string `json:"storedBy"`
	Created           string `json:"created"`
	LastUpdated       string `json:"lastUpdated"`
}

// DHIS2Response represents the response from DHIS2 API
type DHIS2Response struct {
	Pager   map[string]interface{} `json:"pager"`
	Events  []DHIS2Event           `json:"events"`
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
}

// SurveillanceData represents processed surveillance data
type SurveillanceData struct {
	TotalEvents     int                    `json:"totalEvents"`
	EventsByDate    map[string]int         `json:"eventsByDate"`
	EventsByOrgUnit map[string]int         `json:"eventsByOrgUnit"`
	RecentEvents    []ProcessedEvent       `json:"recentEvents"`
	Summary         map[string]interface{} `json:"summary"`
}

// ProcessedEvent represents a processed event for display
type ProcessedEvent struct {
	ID          string                 `json:"id"`
	Date        string                 `json:"date"`
	OrgUnit     string                 `json:"orgUnit"`
	OrgUnitName string                 `json:"orgUnitName"`
	Status      string                 `json:"status"`
	DataValues  map[string]interface{} `json:"dataValues"`
}

// CommunityMortalitySurveillance handles the community mortality surveillance page
func CommunityMortalitySurveillance(c *fiber.Ctx, db *sql.DB, store *session.Store, config config.Config) error {
	sess, err := store.Get(c)
	if err != nil {
		return c.Status(500).SendString("Failed to get session")
	}

	// Check if user is authenticated
	if sess.Get("isAuthenticated") != true {
		return c.Redirect("/login")
	}

	// Get date range from query parameters
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	// Fetch data from DHIS2 API - if no dates provided, fetch all data
	var surveillanceData *SurveillanceData

	if startDate != "" && endDate != "" {
		// Fetch data with date filters
		var err error
		surveillanceData, err = fetchDHIS2Data(config, "community", startDate, endDate)
		log.Printf("DEBUG: Fetching community mortality data with filters: %s to %s", startDate, endDate)
		if err != nil {
			log.Printf("Error fetching DHIS2 data: %v", err)
			// Return empty data structure if API fails
			surveillanceData = &SurveillanceData{
				TotalEvents:     0,
				EventsByDate:    make(map[string]int),
				EventsByOrgUnit: make(map[string]int),
				RecentEvents:    []ProcessedEvent{},
				Summary:         make(map[string]interface{}),
			}
		}
	} else {
		// Fetch all data without date filters
		var err error
		surveillanceData, err = fetchDHIS2Data(config, "community", "", "")
		log.Printf("DEBUG: Fetching all community mortality data without filters")
		if err != nil {
			log.Printf("Error fetching DHIS2 data: %v", err)
			// Return empty data structure if API fails
			surveillanceData = &SurveillanceData{
				TotalEvents:     0,
				EventsByDate:    make(map[string]int),
				EventsByOrgUnit: make(map[string]int),
				RecentEvents:    []ProcessedEvent{},
				Summary:         make(map[string]interface{}),
			}
		}
	}

	templateData := NewTemplateDataWithDB(c, store, db)
	templateData.Title = "Community Mortality Surveillance"
	templateData.SurveillanceData = surveillanceData
	templateData.StartDate = startDate
	templateData.EndDate = endDate

	return GenerateHTML(c, db, templateData, "surveillance_community")
}

// FacilityMortalitySurveillance handles the facility mortality surveillance page
func FacilityMortalitySurveillance(c *fiber.Ctx, db *sql.DB, store *session.Store, config config.Config) error {
	sess, err := store.Get(c)
	if err != nil {
		return c.Status(500).SendString("Failed to get session")
	}

	// Check if user is authenticated
	if sess.Get("isAuthenticated") != true {
		return c.Redirect("/login")
	}

	// Get date range from query parameters
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	// Fetch data from DHIS2 API - if no dates provided, fetch all data
	var surveillanceData *SurveillanceData

	if startDate != "" && endDate != "" {
		// Fetch data with date filters
		var err error
		surveillanceData, err = fetchDHIS2Data(config, "facility", startDate, endDate)
		log.Printf("DEBUG: Fetching facility mortality data with filters: %s to %s", startDate, endDate)
		if err != nil {
			log.Printf("Error fetching DHIS2 data: %v", err)
			// Return empty data structure if API fails
			surveillanceData = &SurveillanceData{
				TotalEvents:     0,
				EventsByDate:    make(map[string]int),
				EventsByOrgUnit: make(map[string]int),
				RecentEvents:    []ProcessedEvent{},
				Summary:         make(map[string]interface{}),
			}
		}
	} else {
		// Fetch all data without date filters
		var err error
		surveillanceData, err = fetchDHIS2Data(config, "facility", "", "")
		log.Printf("DEBUG: Fetching all facility mortality data without filters")
		if err != nil {
			log.Printf("Error fetching DHIS2 data: %v", err)
			// Return empty data structure if API fails
			surveillanceData = &SurveillanceData{
				TotalEvents:     0,
				EventsByDate:    make(map[string]int),
				EventsByOrgUnit: make(map[string]int),
				RecentEvents:    []ProcessedEvent{},
				Summary:         make(map[string]interface{}),
			}
		}
	}

	templateData := NewTemplateDataWithDB(c, store, db)
	templateData.Title = "Facility Mortality Surveillance"
	templateData.SurveillanceData = surveillanceData
	templateData.StartDate = startDate
	templateData.EndDate = endDate

	return GenerateHTML(c, db, templateData, "surveillance_facility")
}

// fetchDHIS2Data fetches data from the DHIS2 API
func fetchDHIS2Data(config config.Config, surveillanceType, startDate, endDate string) (*SurveillanceData, error) {
	// Construct the API URL based on the surveillance type
	var endpoint string
	if surveillanceType == "community" {
		// Community Mortality Surveillance - Community Death LineList
		if startDate != "" && endDate != "" {
			endpoint = fmt.Sprintf("%s/events.json?dimension=ou:LEVEL-5;akV6429SUqu&orgUnit=akV6429SUqu&program=tByt2gf3UFe&ouMode=DESCENDANTS&startDate=%s&endDate=%s&skipPaging=true",
				config.DHIS2API.BaseURL, startDate, endDate)
		} else {
			// Fetch all data without date filters
			endpoint = fmt.Sprintf("%s/events.json?dimension=ou:LEVEL-5;akV6429SUqu&orgUnit=akV6429SUqu&program=tByt2gf3UFe&ouMode=DESCENDANTS&skipPaging=true",
				config.DHIS2API.BaseURL)
		}
		log.Printf("DEBUG: Community mortality endpoint: %s", endpoint)
	} else {
		// Facility Mortality Surveillance - Facility Death LineList
		if startDate != "" && endDate != "" {
			endpoint = fmt.Sprintf("%s/events.json?dimension=ou:LEVEL-4;akV6429SUqu&orgUnit=akV6429SUqu&program=CsXyqsb2cpK&ouMode=DESCENDANTS&startDate=%s&endDate=%s&skipPaging=true",
				config.DHIS2API.BaseURL, startDate, endDate)
		} else {
			// Fetch all data without date filters
			endpoint = fmt.Sprintf("%s/events.json?dimension=ou:LEVEL-4;akV6429SUqu&orgUnit=akV6429SUqu&program=CsXyqsb2cpK&ouMode=DESCENDANTS&skipPaging=true",
				config.DHIS2API.BaseURL)
		}
		log.Printf("DEBUG: Facility mortality endpoint: %s", endpoint)
	}

	// Create HTTP client
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create request
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Add basic authentication
	auth := config.DHIS2API.Username + ":" + config.DHIS2API.Password
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Check if response is successful
	if resp.StatusCode != http.StatusOK {
		log.Printf("DEBUG: DHIS2 API request failed with status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("DEBUG: DHIS2 API response received, status: %d, body length: %d", resp.StatusCode, len(body))

	// Parse the response
	var dhis2Response DHIS2Response
	if err := json.Unmarshal(body, &dhis2Response); err != nil {
		log.Printf("DEBUG: Failed to parse JSON response: %v", err)
		return nil, fmt.Errorf("failed to parse JSON response: %v", err)
	}

	log.Printf("DEBUG: Parsed DHIS2 response - Events count: %d", len(dhis2Response.Events))

	// Process the data
	return processDHIS2Data(dhis2Response), nil
}

// processDHIS2Data processes the raw DHIS2 data into a structured format
func processDHIS2Data(response DHIS2Response) *SurveillanceData {
	data := &SurveillanceData{
		TotalEvents:     len(response.Events),
		EventsByDate:    make(map[string]int),
		EventsByOrgUnit: make(map[string]int),
		RecentEvents:    []ProcessedEvent{},
		Summary:         make(map[string]interface{}),
	}

	// Process each event
	for _, event := range response.Events {
		// Count by date
		if event.EventDate != "" {
			date := event.EventDate[:10] // Extract YYYY-MM-DD
			data.EventsByDate[date]++
		}

		// Count by organization unit
		if event.OrgUnitName != "" {
			data.EventsByOrgUnit[event.OrgUnitName]++
		}

		// Add to recent events (limit to last 50)
		if len(data.RecentEvents) < 50 {
			processedEvent := ProcessedEvent{
				ID:          event.Event,
				Date:        event.EventDate,
				OrgUnit:     event.OrgUnit,
				OrgUnitName: event.OrgUnitName,
				Status:      event.Status,
				DataValues:  make(map[string]interface{}),
			}

			// Process data values
			for _, dv := range event.DataValues {
				processedEvent.DataValues[dv.DataElement] = dv.Value
			}

			data.RecentEvents = append(data.RecentEvents, processedEvent)
		}
	}

	// Calculate summary statistics
	data.Summary["totalEvents"] = data.TotalEvents
	data.Summary["uniqueOrgUnits"] = len(data.EventsByOrgUnit)
	data.Summary["dateRange"] = len(data.EventsByDate)

	return data
}
