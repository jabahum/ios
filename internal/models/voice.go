package models

import (
	"database/sql"
	// "case/internal/database"
	// "case/internal/models"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Session represents call state
type Session struct {
	SessionID    string            `json:"session_id"`
	PhoneNumber  string            `json:"phone_number"`
	PatientName  string            `json:"patient_name"`
	CurrentStep  int               `json:"current_step"`
	Responses    map[string]string `json:"responses"`
	LastActivity time.Time         `json:"last_activity"`
}

// Africa's Talking callback request
type VoiceCallback struct {
	SessionID         string `json:"sessionId"`
	CallSessionState  string `json:"callSessionState"`
	CallerNumber      string `json:"callerNumber"`
	DtmfDigits        string `json:"dtmfDigits"`
	Direction         string `json:"direction"`
	Carrier           string `json:"callerCarrierName"`
	CountryCode       string `json:"callerCountryCode"`
	DurationInSeconds string `json:"durationInSeconds"`
	Amount            string `json:"amount"`
}

// Africa's Talking XML response structures
type ATResponse struct {
	XMLName   xml.Name   `xml:"Response"`
	GetDigits *GetDigits `xml:"GetDigits,omitempty"`
	Say       *Say       `xml:"Say,omitempty"`
}

type GetDigits struct {
	Timeout int `xml:"timeout,attr"`
	// FinishOnKey string `xml:"finishOnKey,attr"`
	NumDigits   int    `xml:"numDigits,attr"`
	CallbackUrl string `xml:"callbackUrl,attr"`
	Say         Say    `xml:"Say"`
}

type Say struct {
	Voice    string `xml:"voice,attr"`
	PlayBeep string `xml:"playBeep,attr"`
	Text     string `xml:",chardata"`
}

type MpoxAssessment struct {
	ID               int       `json:"id"`
	ClientId         int       `json:"client_id"`
	CallSessionId    int       `json:"call_session_id"`
	Platform         string    `json:"platform"`
	Comorbidities    bool      `json:"comorbidities"`
	CompletedAt      string    `json:"completed_at"`
	ExposureAlert    bool      `json:"exposure_alert"`
	PatientCondition string    `json:"patient_condition"`
	NewLesions       bool      `json:"new_lesions"`
	PainLevel        int       `json:"pain_level"`
	PatientName      string    `json:"patient_name"`
	PhoneNumber      string    `json:"phone_number"`
	LesionsDry       bool      `json:"lesions_dry"`
	SessionId        string    `json:"session_id"`
	CreatedAt        time.Time `json:"created_at"`
}

type CallSession struct {
	PhoneNumber       string `json:"phone_number"`
	DurationInSeconds int    `json:"duration_in_seconds"`
	Amount            int    `json:"amount"`
}

var stepPrompts = map[int]string{}

func SendCall(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	phoneNumber := c.Query("phone")
	var phone = [...]string{phoneNumber}

	surveyID := 1 // Mpox survey id = 1, need to do a db call here to get active survey or maybe use outbreak id.

	qns, err := GetQuestionsBySurvey(db, surveyID)
	if err != nil {
		sl.Error("Error getting survey_questions", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get survey_questions",
		})
	}

	for _, value := range qns {
		stepPrompts[value.QuestionNo] = value.Question
	}

	fmt.Println("steps %v", stepPrompts)

	url := "https://voice.africastalking.com/call"
	method := "POST"

	// Build payload
	payloadData := map[string]interface{}{
		"from":     os.Getenv("AT_PHONE"),
		"to":       phone,
		"username": os.Getenv("AT_USERNAME"),
	}

	payloadBytes, err := json.Marshal(payloadData)
	if err != nil {
		log.Fatal("Error marshaling payload:", err)
	}

	payload := bytes.NewReader(payloadBytes)
	// *********

	// Debug print
	fmt.Println("Final JSON payload:", string(payloadBytes))

	client := &http.Client{}

	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("apiKey", os.Getenv("AT_API_KEY"))
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		// return err
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": "failed",
			"error":  err.Error(),
		})
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(string(body))

	// return c.Status(200).JSON(body)
	return c.JSON(fiber.Map{
		"status": "call initiated",
		"phone":  phone,
	})
}

// Global session store (use Redis in production)
var (
	sessions = make(map[string]*Session)
	mutex    = &sync.RWMutex{}
)

func HandleVoiceCallback(c *fiber.Ctx, db *sql.DB) error {
	var callback VoiceCallback
	if err := c.BodyParser(&callback); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}

	fmt.Printf("Callback: SessionID=%s, Phone=%s, DTMF=%s ",
		callback.SessionID, callback.CallerNumber, callback.DtmfDigits)

	// Get or create session
	session := getOrCreateSession(callback.SessionID, callback.CallerNumber)
	var response ATResponse

	// Process DTMF input
	if callback.DtmfDigits != "" {
		if !processInput(session, callback.DtmfDigits) {
			// Invalid input - retry current step
			response = createGetDigitsResponse(session.CurrentStep, session.PatientName, true)
		} else {
			// Valid input - move to next step
			session.CurrentStep++
			if session.CurrentStep > 6 {
				// Assessment complete
				response = createCompletionResponse()
			} else {
				// Continue to next step
				response = createGetDigitsResponse(session.CurrentStep, session.PatientName, false)
			}
		}
	} else {
		// Assessment complete
		if callback.CallSessionState == "Completed" {
			client, _ := ClientByHBCPhone(c.Context(), db, callback.CallerNumber[1:])
			saveMpoxResponse(session, callback.Amount, callback.DurationInSeconds, db, client)
			deleteSession(session.SessionID)
		} else {
			// Assessment not complete
			// Start or continue current step
			response = createGetDigitsResponse(session.CurrentStep, session.PatientName, false)
		}
	}

	session.LastActivity = time.Now()
	updateSession(session)

	// Marshal response to XML
	output, err := xml.Marshal(response)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error encoding XML")
	}

	// Send XML response with header
	c.Type("xml")
	return c.SendString(`<?xml version="1.0" encoding="UTF-8"?>` + string(output))
}

func getOrCreateSession(sessionID, phoneNumber string) *Session {
	mutex.Lock()
	defer mutex.Unlock()

	if session, exists := sessions[sessionID]; exists {
		return session
	}

	// Create new session - get patient name from DB in production
	session := &Session{
		SessionID:    sessionID,
		PhoneNumber:  phoneNumber,
		PatientName:  "Patient", // Fetch from database: getPatientName(phoneNumber)
		CurrentStep:  1,
		Responses:    make(map[string]string),
		LastActivity: time.Now(),
	}

	sessions[sessionID] = session
	return session
}

func updateSession(session *Session) {
	mutex.Lock()
	defer mutex.Unlock()
	sessions[session.SessionID] = session
}

func deleteSession(sessionID string) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(sessions, sessionID)
}

func processInput(session *Session, input string) bool {
	input = strings.TrimSpace(input)
	step := session.CurrentStep

	switch step {
	case 1: // Patient Condition (1-5)
		if input >= "1" && input <= "5" {
			session.Responses["patient_condition"] = input
			return true
		}
	case 2: // Pain level (0-10)
		if level, err := strconv.Atoi(input); err == nil && level >= 0 && level <= 10 {
			session.Responses["pain_level"] = input
			return true
		}
	case 3, 4, 5, 6: // Y/N questions (1=Yes, 2=No)
		if input == "1" || input == "2" {
			stepKey := getStepKey(step)
			session.Responses[stepKey] = input
			return true
		}
	}

	return false
}

func getStepKey(step int) string {
	keys := map[int]string{
		3: "new_lesions",
		4: "comorbidities",
		5: "lesions_dry",
		6: "exposure_alert",
	}
	return keys[step]
}

func createGetDigitsResponse(step int, patientName string, isRetry bool) ATResponse {
	text := fmt.Sprintf(stepPrompts[step])

	if isRetry {
		text = "I didn't understand your response. " + text
	}

	numDigits := 1
	if step == 2 { // Pain level can be 2 digits
		numDigits = 2
	}

	return ATResponse{
		GetDigits: &GetDigits{
			Timeout: 8,
			// FinishOnKey: "#",
			NumDigits:   numDigits,
			CallbackUrl: "https://localhost:3000/voice/callback", // Replace with your domain
			Say: Say{
				Voice:    "woman",
				PlayBeep: "true",
				Text:     text,
			},
		},
	}
}

func createCompletionResponse() ATResponse {
	return ATResponse{
		Say: &Say{
			Voice:    "woman",
			PlayBeep: "false",
			Text:     "Thank you for completing your daily Mpox check-in. Your responses have been recorded and a health worker will contact you if needed. Stay safe.",
		},
	}
}

func saveMpoxResponse(session *Session, amount string, duration string, db *sql.DB, client *Client) {
	painLevel, _ := strconv.Atoi(session.Responses["pain_level"])
	fmt.Println("\n Mpox Responses for Session:<>%v \n", session.Responses)

	// Create structured response
	response := map[string]interface{}{
		"session_id":        session.SessionID,
		"phone_number":      session.PhoneNumber,
		"patient_name":      session.PatientName,
		"patient_condition": session.Responses["patient_condition"],
		"pain_level":        painLevel,
		"new_lesions":       session.Responses["new_lesions"] == "1",
		"comorbidities":     session.Responses["comorbidities"] == "1",
		"lesions_dry":       session.Responses["lesions_dry"] == "1",
		"exposure_alert":    session.Responses["exposure_alert"] == "1",
		"completed_at":      time.Now(),
		"amount":            amount,
		"duration":          duration,
		"platform":          "ivr",
	}

	fmt.Printf("Mpox Assessment Complete: %+v", response)

	// Check for urgent conditions and send alerts
	if session.Responses["patient_condition"] == "3" || // Feel worse
		session.Responses["patient_condition"] == "5" || // Need health worker
		painLevel >= 8 || // High pain
		session.Responses["exposure_alert"] == "1" { // Exposure risk

		sendUrgentAlert(session.PatientName, session.PhoneNumber, response)
	}

	// In production: save to database
	// db.Create(&MpoxResponse{...})
	jsonBytes, _ := json.Marshal(response)
	var assessmentData MpoxAssessment
	// convert jsonBytes to Json and bind it to assessmentData
	err := json.Unmarshal(jsonBytes, &assessmentData)
	if err != nil {
		panic(err)
	}

	amt, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return
	}
	intAmount := math.Round(amt)

	var callID int
	query := `INSERT INTO call_sessions (phone_number, duration_in_seconds, amount) VALUES ($1, $2, $3) RETURNING id`
	err = db.QueryRow(query, assessmentData.PhoneNumber, duration, intAmount).Scan(&callID)
	if err != nil {
		fmt.Println("Error saving call session: %v", err)
		return
	}

	query1 := `INSERT INTO mpox_assessments ` +
		`(client_id, call_session_id, platform, phone_number, patient_condition, pain_level, new_lesions, comorbidities, lesions_dry, exposure_alert) ` +
		`VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	// save assessment to database
	_, err1 := db.Exec(query1, client.ID, callID, assessmentData.Platform, assessmentData.PhoneNumber, assessmentData.PatientCondition, assessmentData.PainLevel, assessmentData.NewLesions, assessmentData.Comorbidities, assessmentData.LesionsDry, assessmentData.ExposureAlert)
	if err1 != nil {
		fmt.Println("Error saving mpox assessment: %v", err)
		return
	}

	return
}

func sendUrgentAlert(patientName, phoneNumber string, response map[string]interface{}) {
	fmt.Printf("🚨 URGENT ALERT: Patient %s (%s) requires immediate attention", patientName, phoneNumber)

	// In production, implement:
	// - SMS to health worker: sendSMS(healthWorkerNumber, alertMessage)
	// - Email notification: sendEmail(healthWorkerEmail, alertDetails)
	// - Push notification: sendPushNotification(alertData)
	// - Create support ticket: createTicket(patientData, response)
}
