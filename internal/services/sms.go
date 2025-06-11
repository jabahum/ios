package services

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

type SMSConfig struct {
	BaseURL  string
	Username string
	Password string
}

type SMSService struct {
	config SMSConfig
	logger *slog.Logger
}

func NewSMSService(config SMSConfig) *SMSService {
	return &SMSService{
		config: config,
		logger: slog.Default(),
	}
}

func (s *SMSService) SendSMS(to, message string) error {
	// Format phone number - remove '+' if present and ensure it starts with '256'
	phoneNumber := to
	if len(phoneNumber) > 0 && phoneNumber[0] == '+' {
		phoneNumber = phoneNumber[1:]
	}
	if len(phoneNumber) > 0 && phoneNumber[0] == '0' {
		phoneNumber = "256" + phoneNumber[1:]
	}

	// Create the request URL with query parameters
	params := url.Values{}
	params.Add("text", message)
	params.Add("to", phoneNumber)

	// Construct the full SMS URL
	smsURL := s.config.BaseURL + "/sendsms"

	// Log the request details
	s.logger.Info("Sending SMS request",
		"url", smsURL,
		"original_phone", to,
		"formatted_phone", phoneNumber,
		"username", s.config.Username,
		"message", message)

	// Create the request
	req, err := http.NewRequest("GET", smsURL+"?"+params.Encode(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Add basic auth
	req.SetBasicAuth(s.config.Username, s.config.Password)

	// Create client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second, // 10 second timeout
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	// Log response details
	s.logger.Info("SMS response received",
		"status", resp.Status,
		"body", string(body))

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SMS sending failed with status: %s, body: %s", resp.Status, string(body))
	}

	return nil
}
