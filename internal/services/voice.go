package services

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	// "net/url"
	"time"
	"strings"
	"os"
	"encoding/json"
	"bytes"
)

type VoiceConfig struct {
	VoiceURL  string
}

type VoiceService struct {
	config VoiceConfig
	logger *slog.Logger
}

func NewVoiceService(config VoiceConfig) *VoiceService {
	return &VoiceService{
		config: config,
		logger: slog.Default(),
	}
}

func (s *VoiceService) MakeVoiceCall(to string) error {
	phoneNumbers := strings.Split(to, ",")

	for _, phoneNumber := range phoneNumbers {
		// Format phone number - make sure it begins with '+256'
		// phoneNumber := to
		if len(phoneNumber) > 0 && phoneNumber[0] == '2' {
			phoneNumber = "+"+phoneNumber[0:]
		}
		if len(phoneNumber) > 0 && phoneNumber[0] == '0' {
			phoneNumber = "+256" + phoneNumber[1:]
		}
	}

		// Construct the full Voice URL
	voiceURL := s.config.VoiceURL

	// Log the request details
	s.logger.Info("Sending Voice call request",
		"url", voiceURL,
		"recipients", phoneNumbers)
	
	// Build payload
    payloadData := map[string]interface{}{
        "from":     os.Getenv("AT_PHONE"),
        "to":       phoneNumbers,
        "username": os.Getenv("AT_USERNAME"),
    }

    payloadBytes, err := json.Marshal(payloadData)
    if err != nil {
        fmt.Errorf("Error marshaling payload:", err)
    }

    payload := bytes.NewReader(payloadBytes)
	// *********

	// Debug print
	fmt.Println("Final JSON payload:", string(payloadBytes))

	// Create the request
	req, err := http.NewRequest("POST", voiceURL, payload)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("apiKey", os.Getenv("AT_API_KEY"))
	req.Header.Add("Content-Type", "application/json")

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
	s.logger.Info("Voice call response received",
		"status", resp.Status,
		"body", string(body))

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Voice call failed with status: %s, body: %s", resp.Status, string(body))
	}

	return nil
}
