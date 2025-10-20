package handlers

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// ExternalAlert represents an alert from the external API
type ExternalAlert struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Severity      string    `json:"severity"`
	Category      string    `json:"category"`
	Location      string    `json:"location"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Status        string    `json:"status"`
	Reporter      string    `json:"reporter"`
	ContactNumber string    `json:"contact_number"`
	Source        string    `json:"source"`
	Response      string    `json:"response"`
	Verified      bool      `json:"verified"`
	CIFNumber     string    `json:"cif_number,omitempty"`
	CaseCode      string    `json:"case_code,omitempty"`

	// API response fields (for unmarshaling)
	IDAPI int `json:"id"`

	// Additional fields from actual API response
	Date                       string `json:"date"`
	Time                       string `json:"time"`
	CallTaker                  string `json:"callTaker"`
	CifNo                      string `json:"cifNo"`
	PersonReporting            string `json:"personReporting"`
	Village                    string `json:"village"`
	SubCounty                  string `json:"subCounty"`
	ContactNumberAPI           string `json:"contactNumber"`
	SourceOfAlert              string `json:"sourceOfAlert"`
	AlertCaseName              string `json:"alertCaseName"`
	AlertCaseAge               int    `json:"alertCaseAge"`
	AlertCaseSex               string `json:"alertCaseSex"`
	AlertCasePregnantDuration  int    `json:"alertCasePregnantDuration"`
	AlertCaseVillage           string `json:"alertCaseVillage"`
	AlertCaseParish            string `json:"alertCaseParish"`
	AlertCaseSubCounty         string `json:"alertCaseSubCounty"`
	AlertCaseDistrict          string `json:"alertCaseDistrict"`
	AlertCaseNationality       string `json:"alertCaseNationality"`
	PointOfContactName         string `json:"pointOfContactName"`
	PointOfContactRelationship string `json:"pointOfContactRelationship"`
	PointOfContactPhone        string `json:"pointOfContactPhone"`
	History                    string `json:"history"`
	HealthFacilityVisit        string `json:"healthFacilityVisit"`
	TraditionalHealerVisit     string `json:"traditionalHealerVisit"`
	Symptoms                   string `json:"symptoms"`
	Actions                    string `json:"actions"`
	CaseVerificationDesk       string `json:"caseVerificationDesk"`
	FieldVerification          string `json:"fieldVerification"`
	FieldVerificationDecision  string `json:"fieldVerificationDecision"`
	Feedback                   string `json:"feedback"`
	LabResult                  string `json:"labResult"`
	LabResultDate              string `json:"labResultDate"`
	IsHighlighted              bool   `json:"isHighlighted"`
	AssignedTo                 string `json:"assignedTo"`
	AlertReportedBefore        string `json:"alertReportedBefore"`
	AlertFrom                  string `json:"alertFrom"`
	VerifiedAPI                *bool  `json:"verified"`
	Comments                   string `json:"comments"`
	VerificationDate           string `json:"verificationDate"`
	VerificationTime           string `json:"verificationTime"`
	ResponseAPI                string `json:"response"`
	Narrative                  string `json:"narrative"`
	FacilityType               string `json:"facilityType"`
	Facility                   string `json:"facility"`
	IsVerified                 bool   `json:"isVerified"`
	VerifiedBy                 string `json:"verifiedBy"`
	Region                     string `json:"region"`
	CaseCodeAPI                string `json:"caseCode"`
	CreatedAtAPI               string `json:"createdAt"`
	UpdatedAtAPI               string `json:"updatedAt"`
}

// AlertsResponse represents the response from the external alerts API
type AlertsResponse struct {
	Success bool            `json:"success"`
	Data    []ExternalAlert `json:"data"`
	Message string          `json:"message"`
}

// PaginatedAlertsResponse represents the paginated response for alerts API
type PaginatedAlertsResponse struct {
	Success    bool            `json:"success"`
	Data       []ExternalAlert `json:"data"`
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
	Message    string          `json:"message"`
}

// TokenResponse represents the response from the authentication endpoint
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// TokenCache holds cached token information
type TokenCache struct {
	Token     string
	ExpiresAt time.Time
	mutex     sync.RWMutex
}

// Global token cache
var tokenCache = &TokenCache{}

// mapAPIAlertToFrontend maps API response fields to frontend expected fields
func mapAPIAlertToFrontend(apiAlert ExternalAlert) ExternalAlert {
	// Parse dates
	var createdAt time.Time
	if apiAlert.CreatedAtAPI != "" {
		if parsed, err := time.Parse("2006-01-02T15:04:05-07:00", apiAlert.CreatedAtAPI); err == nil {
			createdAt = parsed
		} else if parsed, err := time.Parse("2006-01-02T15:04:05.000Z", apiAlert.CreatedAtAPI); err == nil {
			createdAt = parsed
		}
	}

	var updatedAt time.Time
	if apiAlert.UpdatedAtAPI != "" && apiAlert.UpdatedAtAPI != "0001-01-01T00:00:00Z" {
		if parsed, err := time.Parse("2006-01-02T15:04:05-07:00", apiAlert.UpdatedAtAPI); err == nil {
			updatedAt = parsed
		} else if parsed, err := time.Parse("2006-01-02T15:04:05.000Z", apiAlert.UpdatedAtAPI); err == nil {
			updatedAt = parsed
		}
	}

	// Map fields to frontend expected format
	return ExternalAlert{
		ID:            fmt.Sprintf("%d", apiAlert.IDAPI),
		Title:         apiAlert.AlertCaseName,
		Description:   apiAlert.Symptoms,
		Severity:      "Medium", // Default since not provided by API
		Category:      apiAlert.ResponseAPI,
		Location:      apiAlert.AlertCaseDistrict,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
		Status:        apiAlert.Status,
		Reporter:      apiAlert.PersonReporting,
		ContactNumber: apiAlert.ContactNumberAPI,
		Source:        apiAlert.SourceOfAlert,
		Response:      apiAlert.ResponseAPI,
		Verified:      apiAlert.IsVerified,
		CIFNumber:     apiAlert.CifNo,
		CaseCode:      apiAlert.CaseCodeAPI,
	}
}

// HandlerAlerts handles the alerts page
func HandlerAlerts(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	// Fetch alerts from external API
	alerts, err := fetchExternalAlerts(config)
	if err != nil {
		sl.Error("Failed to fetch external alerts", "error", err)
		// Continue with empty alerts rather than failing completely
		alerts = []ExternalAlert{}
	}

	data := NewTemplateDataWithDB(c, store, db)
	data.Form = fiber.Map{
		"Title":    "Health Alerts",
		"Alerts":   alerts,
		"Error":    err,
		"HasError": err != nil,
	}

	return GenerateHTML(c, db, data, "alerts")
}

// HandlerAlertsAPI handles the API endpoint for paginated alerts
func HandlerAlertsAPI(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	// Get query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "10"))
	search := c.Query("search", "")
	severity := c.Query("severity", "")
	status := c.Query("status", "")
	location := c.Query("location", "")

	// Validate parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 0 || pageSize > 10000 {
		pageSize = 0
	}

	// Fetch all alerts from external API
	allAlerts, err := fetchExternalAlerts(config)
	if err != nil {
		sl.Error("Failed to fetch external alerts", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch alerts",
			"error":   err.Error(),
		})
	}

	// Filter alerts based on search criteria
	filteredAlerts := filterAlerts(allAlerts, search, severity, status, location)

	// If pageSize is 0, return all alerts without pagination
	if pageSize == 0 {
		response := PaginatedAlertsResponse{
			Success:    true,
			Data:       filteredAlerts,
			Total:      len(filteredAlerts),
			Page:       1,
			PageSize:   len(filteredAlerts),
			TotalPages: 1,
			Message:    "All alerts retrieved successfully",
		}
		return c.JSON(response)
	}

	// Calculate pagination
	total := len(filteredAlerts)
	totalPages := (total + pageSize - 1) / pageSize

	// Apply pagination
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}

	var paginatedAlerts []ExternalAlert
	if start < total {
		paginatedAlerts = filteredAlerts[start:end]
	}

	response := PaginatedAlertsResponse{
		Success:    true,
		Data:       paginatedAlerts,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		Message:    "Alerts retrieved successfully",
	}

	return c.JSON(response)
}

// HandlerAlertsDebug handles debugging the external API
func HandlerAlertsDebug(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	// Test the external API directly
	alerts, err := fetchExternalAlerts(config)

	debugInfo := fiber.Map{
		"success":      err == nil,
		"error":        nil,
		"alerts_count": len(alerts),
		"sample_data":  err != nil,
	}

	if err != nil {
		debugInfo["error"] = err.Error()
	}

	if len(alerts) > 0 {
		debugInfo["first_alert"] = alerts[0]
	}

	return c.JSON(debugInfo)
}

// filterAlerts filters alerts based on search criteria
func filterAlerts(alerts []ExternalAlert, search, severity, status, location string) []ExternalAlert {
	var filtered []ExternalAlert

	for _, alert := range alerts {
		// Search filter
		if search != "" {
			searchLower := strings.ToLower(search)
			matches := strings.Contains(strings.ToLower(alert.Title), searchLower) ||
				strings.Contains(strings.ToLower(alert.Description), searchLower) ||
				strings.Contains(strings.ToLower(alert.Reporter), searchLower) ||
				strings.Contains(strings.ToLower(alert.Location), searchLower) ||
				strings.Contains(strings.ToLower(alert.CIFNumber), searchLower) ||
				strings.Contains(strings.ToLower(alert.CaseCode), searchLower)

			if !matches {
				continue
			}
		}

		// Severity filter
		if severity != "" && alert.Severity != severity {
			continue
		}

		// Status filter
		if status != "" && alert.Status != status {
			continue
		}

		// Location filter
		if location != "" && !strings.Contains(strings.ToLower(alert.Location), strings.ToLower(location)) {
			continue
		}

		filtered = append(filtered, alert)
	}

	return filtered
}

// getBearerToken returns the provided test token for now
func getBearerToken(config Config) (string, error) {
	// Using the provided test token
	testToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTc2NjIzMTcsInVzZXJfaWQiOjEsInVzZXJuYW1lIjoicHdhaXN3YSJ9.WevC8m0gFOlQmI3or8QdVuDRGx3-D5M44DV4gIABXTM"

	// Cache the token with expiration
	tokenCache.mutex.Lock()
	tokenCache.Token = testToken
	tokenCache.ExpiresAt = time.Now().Add(55 * time.Minute) // Default to 55 minutes
	tokenCache.mutex.Unlock()

	return testToken, nil
}

// fetchExternalAlerts fetches alerts from the external API using token-based authentication
func fetchExternalAlerts(config Config) ([]ExternalAlert, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	url := config.AlertsAPI.BaseURL + "/alerts"

	// Get bearer token using proper authentication
	token, err := getBearerToken(config)
	if err != nil {
		return getSampleAlerts(), fmt.Errorf("failed to obtain authentication token: %w", err)
	}

	// Make authenticated request with bearer token
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return getSampleAlerts(), fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return getSampleAlerts(), fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return getSampleAlerts(), fmt.Errorf("failed to read response body: %w", err)
		}

		// Try parsing directly as array first (this is what the API actually returns)
		var apiAlerts []ExternalAlert
		if err := json.Unmarshal(bodyBytes, &apiAlerts); err == nil {
			// Map API response to frontend expected format
			var mappedAlerts []ExternalAlert
			for _, apiAlert := range apiAlerts {
				mappedAlerts = append(mappedAlerts, mapAPIAlertToFrontend(apiAlert))
			}
			return mappedAlerts, nil
		}

		// Try parsing as structured response as fallback
		var alertsResponse AlertsResponse
		if err := json.Unmarshal(bodyBytes, &alertsResponse); err == nil && alertsResponse.Success {
			// Map API response to frontend expected format
			var mappedAlerts []ExternalAlert
			for _, apiAlert := range alertsResponse.Data {
				mappedAlerts = append(mappedAlerts, mapAPIAlertToFrontend(apiAlert))
			}
			return mappedAlerts, nil
		}

		return getSampleAlerts(), fmt.Errorf("failed to parse API response, returning samples")
	}

	// If token authentication failed, try basic auth as fallback
	auth := config.AlertsAPI.Username + ":" + config.AlertsAPI.Password
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return getSampleAlerts(), fmt.Errorf("failed to create fallback request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
	req.Header.Set("Accept", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		return getSampleAlerts(), fmt.Errorf("failed to make fallback request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return getSampleAlerts(), fmt.Errorf("failed to read fallback response body: %w", err)
		}

		// Try parsing directly as array first (this is what the API actually returns)
		var apiAlerts []ExternalAlert
		if err := json.Unmarshal(bodyBytes, &apiAlerts); err == nil {
			// Map API response to frontend expected format
			var mappedAlerts []ExternalAlert
			for _, apiAlert := range apiAlerts {
				mappedAlerts = append(mappedAlerts, mapAPIAlertToFrontend(apiAlert))
			}
			return mappedAlerts, nil
		}

		// Try parsing as structured response as fallback
		var alertsResponse AlertsResponse
		if err := json.Unmarshal(bodyBytes, &alertsResponse); err == nil && alertsResponse.Success {
			// Map API response to frontend expected format
			var mappedAlerts []ExternalAlert
			for _, apiAlert := range alertsResponse.Data {
				mappedAlerts = append(mappedAlerts, mapAPIAlertToFrontend(apiAlert))
			}
			return mappedAlerts, nil
		}
	}

	return getSampleAlerts(), fmt.Errorf("failed to fetch real alerts, returning samples")
}

// DHIS2 minimal structures for 6767 alerts
type dhis2Event struct {
	Event       string           `json:"event"`
	EventDate   string           `json:"eventDate"`
	Status      string           `json:"status"`
	OrgUnit     string           `json:"orgUnit"`
	OrgUnitName string           `json:"orgUnitName"`
	DataValues  []dhis2DataValue `json:"dataValues"`
}

type dhis2DataValue struct {
	DataElement string `json:"dataElement"`
	Value       string `json:"value"`
}

type dhis2EventsResponse struct {
	Events []dhis2Event `json:"events"`
}

// fetchDHIS2Events retrieves events for a given DHIS2 program
func fetchDHIS2Events(config Config, programID, startDate, endDate string) ([]dhis2Event, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	// Build endpoint
	endpoint := fmt.Sprintf("%s/events.json?program=%s&orgUnit=akV6429SUqu&ouMode=DESCENDANTS&skipPaging=true", config.DHIS2API.BaseURL, programID)
	if startDate != "" && endDate != "" {
		endpoint = fmt.Sprintf("%s&startDate=%s&endDate=%s", endpoint, startDate, endDate)
	}

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	// Basic auth
	auth := config.DHIS2API.Username + ":" + config.DHIS2API.Password
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("dhis2 error: %d %s", resp.StatusCode, string(b))
	}

	var dr dhis2EventsResponse
	if err := json.NewDecoder(resp.Body).Decode(&dr); err != nil {
		return nil, err
	}
	return dr.Events, nil
}

// HandlerAlerts6767API returns alerts sourced from DHIS2 program iaN1DovM5em
func HandlerAlerts6767API(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	startDate := c.Query("startDate", "")
	endDate := c.Query("endDate", "")

	events, err := fetchDHIS2Events(config, "iaN1DovM5em", startDate, endDate)
	if err != nil {
		sl.Error("Failed to fetch 6767 alerts from DHIS2", "error", err)
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to fetch 6767 alerts", "error": err.Error()})
	}

	// Convert to ExternalAlert slice for frontend reuse
	var alerts []ExternalAlert
	for _, ev := range events {
		t, _ := time.Parse("2006-01-02T15:04:05.000", ev.EventDate)
		if t.IsZero() {
			// fallback to date only
			t, _ = time.Parse("2006-01-02", ev.EventDate)
		}
		alerts = append(alerts, ExternalAlert{
			ID:          ev.Event,
			Title:       "6767 Alert",
			Description: "Imported from DHIS2",
			Severity:    "Medium",
			Category:    "6767",
			Location:    ev.OrgUnitName,
			CreatedAt:   t,
			UpdatedAt:   t,
			Status:      ev.Status,
			Reporter:    "",
			Source:      "DHIS2",
			Response:    "6767",
			Verified:    false,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    alerts,
		"total":   len(alerts),
		"message": "6767 alerts retrieved successfully",
	})
}

// getSampleAlerts returns sample alerts for demonstration purposes
func getSampleAlerts() []ExternalAlert {
	now := time.Now()
	return []ExternalAlert{
		{
			ID:            "ALT3755",
			Title:         "Measles Outbreak Alert",
			Description:   "Multiple cases of measles reported in Kampala district. Immediate vaccination campaign recommended.",
			Severity:      "High",
			Category:      "Infectious Disease",
			Location:      "Kasese District",
			CreatedAt:     now.Add(-2 * time.Hour),
			UpdatedAt:     now.Add(-30 * time.Minute),
			Status:        "Alive",
			Reporter:      "Dr. Sarah Nakimera",
			ContactNumber: "0782388634",
			Source:        "Community",
			Response:      "Measles",
			Verified:      true,
			CIFNumber:     "CIF-2023-001",
			CaseCode:      "CASE-2023-001",
		},
		{
			ID:            "ALT3754",
			Title:         "Cholera Prevention Notice",
			Description:   "Heavy rains expected in Eastern region. Communities advised to ensure clean water sources.",
			Severity:      "Medium",
			Category:      "Waterborne Disease",
			Location:      "Kasese District",
			CreatedAt:     now.Add(-4 * time.Hour),
			UpdatedAt:     now.Add(-1 * time.Hour),
			Status:        "Alive",
			Reporter:      "Nurse Mary Okello",
			ContactNumber: "0782388634",
			Source:        "Facility",
			Response:      "Cholera",
			Verified:      true,
			CIFNumber:     "CIF-2023-002",
			CaseCode:      "CASE-2023-002",
		},
		{
			ID:            "ALT3753",
			Title:         "COVID-19 Variant Update",
			Description:   "New variant detected in border regions. Enhanced surveillance and testing recommended.",
			Severity:      "Critical",
			Category:      "Respiratory Disease",
			Location:      "Kasese District",
			CreatedAt:     now.Add(-6 * time.Hour),
			UpdatedAt:     now.Add(-2 * time.Hour),
			Status:        "Alive",
			Reporter:      "Dr. John Muwonge",
			ContactNumber: "0782388634",
			Source:        "Facility",
			Response:      "COVID-19",
			Verified:      true,
			CIFNumber:     "CIF-2023-003",
			CaseCode:      "CASE-2023-003",
		},
		{
			ID:            "ALT3752",
			Title:         "Malaria Prevention Campaign",
			Description:   "Seasonal increase in malaria cases. Distribution of mosquito nets and awareness campaigns.",
			Severity:      "Medium",
			Category:      "Vector-borne Disease",
			Location:      "Kasese District",
			CreatedAt:     now.Add(-8 * time.Hour),
			UpdatedAt:     now.Add(-3 * time.Hour),
			Status:        "Alive",
			Reporter:      "Dr. Grace Nalukenge",
			ContactNumber: "0782388634",
			Source:        "Facility",
			Response:      "Malaria",
			Verified:      true,
			CIFNumber:     "CIF-2023-004",
			CaseCode:      "CASE-2023-004",
		},
		{
			ID:            "ALT3751",
			Title:         "Food Safety Alert",
			Description:   "Contaminated food products reported in central markets. Public advised to check food sources.",
			Severity:      "High",
			Category:      "Food Safety",
			Location:      "Kasese District",
			CreatedAt:     now.Add(-12 * time.Hour),
			UpdatedAt:     now.Add(-6 * time.Hour),
			Status:        "Alive",
			Reporter:      "Dr. Robert Ssebunya",
			ContactNumber: "0782388634",
			Source:        "Facility",
			Response:      "Food Safety",
			Verified:      true,
			CIFNumber:     "CIF-2023-005",
			CaseCode:      "CASE-2023-005",
		},
		{
			ID:            "ALT3749",
			Title:         "Ebola Outbreak Alert",
			Description:   "Suspected Ebola cases reported in Bundibugyo district. Immediate isolation and testing required.",
			Severity:      "Critical",
			Category:      "Hemorrhagic Fever",
			Location:      "Bundibugyo District",
			CreatedAt:     now.Add(-14 * time.Hour),
			UpdatedAt:     now.Add(-8 * time.Hour),
			Status:        "Alive",
			Reporter:      "Dr. Alice Namukasa",
			ContactNumber: "0775677566",
			Source:        "Facility",
			Response:      "EVD",
			Verified:      true,
			CIFNumber:     "CIF-2023-006",
			CaseCode:      "CASE-2023-006",
		},
		{
			ID:            "ALT3748",
			Title:         "Ebola Case Confirmation",
			Description:   "Confirmed Ebola case in Mukono district. Contact tracing and quarantine measures implemented.",
			Severity:      "Critical",
			Category:      "Hemorrhagic Fever",
			Location:      "Mukono District",
			CreatedAt:     now.Add(-16 * time.Hour),
			UpdatedAt:     now.Add(-10 * time.Hour),
			Status:        "Alive",
			Reporter:      "Dr. Peter Kato",
			ContactNumber: "0771234567",
			Source:        "Community",
			Response:      "EVD",
			Verified:      true,
			CIFNumber:     "CIF-2023-007",
			CaseCode:      "CASE-2023-007",
		},
		{
			ID:            "ALT3756",
			Title:         "Mpox Outbreak Alert",
			Description:   "Multiple Mpox cases confirmed in Bunyangabu district. Vaccination campaign initiated.",
			Severity:      "High",
			Category:      "Viral Disease",
			Location:      "Bunyangabu District",
			CreatedAt:     now.Add(-18 * time.Hour),
			UpdatedAt:     now.Add(-12 * time.Hour),
			Status:        "Alive",
			Reporter:      "Dr. Betty Nalwoga",
			ContactNumber: "0772931104",
			Source:        "Facility",
			Response:      "Mpox",
			Verified:      true,
			CIFNumber:     "CIF-2023-008",
			CaseCode:      "CASE-2023-008",
		},
		{
			ID:            "ALT3750",
			Title:         "Mpox Case Investigation",
			Description:   "Suspected Mpox case under investigation in Bunyangabu district. Samples sent for testing.",
			Severity:      "Medium",
			Category:      "Viral Disease",
			Location:      "Bunyangabu District",
			CreatedAt:     now.Add(-20 * time.Hour),
			UpdatedAt:     now.Add(-14 * time.Hour),
			Status:        "Alive",
			Reporter:      "Dr. James Muwonge",
			ContactNumber: "0705352032",
			Source:        "Facility",
			Response:      "Mpox",
			Verified:      true,
			CIFNumber:     "CIF-2023-009",
			CaseCode:      "CASE-2023-009",
		},
		{
			ID:            "ALT3747",
			Title:         "Mpox Prevention Campaign",
			Description:   "Mpox prevention campaign launched in Kyegegwa district. Public awareness and vaccination.",
			Severity:      "Medium",
			Category:      "Viral Disease",
			Location:      "Kyegegwa District",
			CreatedAt:     now.Add(-22 * time.Hour),
			UpdatedAt:     now.Add(-16 * time.Hour),
			Status:        "Alive",
			Reporter:      "Dr. Christine Nalukenge",
			ContactNumber: "0780115709",
			Source:        "Facility",
			Response:      "Mpox",
			Verified:      true,
			CIFNumber:     "CIF-2023-010",
			CaseCode:      "CASE-2023-010",
		},
	}
}
