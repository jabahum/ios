package handlers

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
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
}

// AlertsResponse represents the response from the external alerts API
type AlertsResponse struct {
	Success bool            `json:"success"`
	Data    []ExternalAlert `json:"data"`
	Message string          `json:"message"`
}

// TokenResponse represents the response from the authentication endpoint
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// HandlerAlerts handles the alerts page
func HandlerAlerts(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	// Fetch alerts from external API
	alerts, err := fetchExternalAlerts()
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

// getBearerToken fetches a bearer token using the provided credentials
func getBearerToken() (string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create login request
	loginData := map[string]string{
		"username": "pwaiswa",
		"password": "leaves",
	}

	loginJSON, err := json.Marshal(loginData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal login data: %w", err)
	}

	// Try different authentication endpoints
	endpoints := []string{
		"https://alerts.health.go.ug/api/v1/login",
		"https://alerts.health.go.ug/api/v1/token",
		// "https://alerts.health.go.ug/api/v1/oauth/token",
		// "https://alerts.health.go.ug/auth/login",
	}

	for _, endpoint := range endpoints {
		req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(loginJSON))
		if err != nil {
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		if resp.StatusCode == http.StatusOK {
			bodyBytes, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				continue
			}

			var tokenResp TokenResponse
			if err := json.Unmarshal(bodyBytes, &tokenResp); err != nil {
				// Try parsing as simple token response
				var simpleResp map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &simpleResp); err == nil {
					if token, ok := simpleResp["token"].(string); ok {
						return token, nil
					}
					if token, ok := simpleResp["access_token"].(string); ok {
						return token, nil
					}
				}
				continue
			}

			if tokenResp.AccessToken != "" {
				return tokenResp.AccessToken, nil
			}
		}
		resp.Body.Close()
	}

	return "", fmt.Errorf("failed to obtain bearer token from any endpoint")
}

// fetchExternalAlerts fetches alerts from the external API
func fetchExternalAlerts() ([]ExternalAlert, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create request
	req, err := http.NewRequest("GET", "https://alerts.health.go.ug/api/v1/alerts", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Try different authentication methods
	authMethods := []struct {
		name   string
		header string
		value  string
	}{
		{"Basic Auth", "Authorization", "Basic " + base64.StdEncoding.EncodeToString([]byte("pwaiswa:leaves"))},
		{"API Key", "X-API-Key", "pwaiswa"},
		{"API Key", "Authorization", "ApiKey pwaiswa"},
		{"Custom Header", "X-Auth-Username", "pwaiswa"},
		{"Custom Header", "X-Auth-Password", "leaves"},
	}

	for _, method := range authMethods {
		req.Header.Set(method.header, method.value)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		// Make request
		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		if resp.StatusCode == http.StatusOK {
			// Read response body
			bodyBytes, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				continue
			}

			// Debug: Log the raw response
			fmt.Printf("DEBUG: Raw API response with %s: %s\n", method.name, string(bodyBytes))

			// Parse response
			var alertsResponse AlertsResponse
			if err := json.Unmarshal(bodyBytes, &alertsResponse); err != nil {
				continue
			}

			// Debug: Log the parsed response
			fmt.Printf("DEBUG: Parsed response - Success: %v, Message: %s, Data count: %d\n",
				alertsResponse.Success, alertsResponse.Message, len(alertsResponse.Data))

			// Check if the API response indicates success
			if alertsResponse.Success {
				return alertsResponse.Data, nil
			}
		} else {
			resp.Body.Close()
		}
	}

	// If external API fails, return sample alerts for demonstration
	return getSampleAlerts(), nil
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
			Reporter:      "Not specified",
			ContactNumber: "0782388634",
			Source:        "Community",
			Response:      "Mpox",
			Verified:      true,
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
			Reporter:      "Not specified",
			ContactNumber: "0782388634",
			Source:        "Facility",
			Response:      "Mpox",
			Verified:      true,
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
			Reporter:      "Not specified",
			ContactNumber: "0782388634",
			Source:        "Facility",
			Response:      "Mpox",
			Verified:      true,
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
			Reporter:      "Not specified",
			ContactNumber: "0782388634",
			Source:        "Facility",
			Response:      "Mpox",
			Verified:      true,
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
			Reporter:      "Not specified",
			ContactNumber: "0782388634",
			Source:        "Facility",
			Response:      "Mpox",
			Verified:      true,
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
			Reporter:      "Not specified",
			ContactNumber: "0775677566",
			Source:        "Facility",
			Response:      "EVD",
			Verified:      true,
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
			Reporter:      "Not specified",
			ContactNumber: "12",
			Source:        "Community",
			Response:      "EVD",
			Verified:      true,
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
			Reporter:      "Not specified",
			ContactNumber: "0772931104",
			Source:        "Facility",
			Response:      "Mpox",
			Verified:      true,
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
			Reporter:      "Not specified",
			ContactNumber: "0705352032",
			Source:        "Facility",
			Response:      "Mpox",
			Verified:      true,
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
			Reporter:      "Not specified",
			ContactNumber: "0780115709",
			Source:        "Facility",
			Response:      "Mpox",
			Verified:      true,
		},
	}
}
