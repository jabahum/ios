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
	sync.Mutex
	SessionID    string            `json:"session_id"`
	PhoneNumber  string            `json:"phone_number"`
	PatientName  string            `json:"patient_name"`
	Language     string            `json:"language"`
	CurrentStep  int               `json:"current_step"`
	Responses    MpoxAssessment    `json:"responses"`
	LastActivity time.Time         `json:"last_activity"`
	ClientID     int               `json:"client_id"`
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
	Play      *Play      `xml:"Play,omitempty"`
}

type GetDigits struct {
	Timeout int `xml:"timeout,attr"`
	// FinishOnKey string `xml:"finishOnKey,attr"`
	NumDigits   int    `xml:"numDigits,attr"`
	CallbackUrl string `xml:"callbackUrl,attr"`
	// Say         Say    `xml:"Say"`
	Play Play `xml:"Play"`
}

type Say struct {
	Voice    string `xml:"voice,attr"`
	PlayBeep string `xml:"playBeep,attr"`
	Text     string `xml:",chardata"`
}

type Play struct {
	Url string `xml:"url,attr"`
}

type MpoxAssessment struct {
	ID               int       `json:"id"`
	ClientId         int       `json:"client_id"`
	CallSessionId    int       `json:"call_session_id"`
	Platform         string    `json:"platform"`
	Comorbidities    bool      `json:"comorbidities"`
	CompletedAt      string    `json:"completed_at"`
	ExposureAlert    bool      `json:"exposure_alert"`
	PatientCondition int       `json:"patient_condition"`
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

// Config represents the application configuration
type Config struct {
	AT_PHONE	 string `json:"AT_PHONE"`
	AT_USERNAME  string `json:"AT_USERNAME"`
	AT_API_KEY   string `json:"AT_API_KEY"`
}

var stepPrompts = map[int]string{}

func SendCall(c *fiber.Ctx, db *sql.DB, sl *slog.Logger) error {
	phoneNumber := c.Query("phone")
	fmt.Println("phone number received:", phoneNumber)
	// var phone = [...]string{phoneNumber}
	phone := strings.Split(phoneNumber, ", ") // Split by comma and space to get individual numbers
	fmt.Println("phone number:", phone)

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

	fmt.Printf("steps %v", stepPrompts)

	// Load configuration from config.json without killing the server on error
	config, err := loadConfig()
	if err != nil {
		sl.Error("Error loading voice config", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": "failed",
			"error":  "Failed to load voice configuration",
		})
	}
	fmt.Printf("Voice config loaded: %+v\n", config)

	url := "https://voice.africastalking.com/call"
	method := "POST"

	// Build payload
	payloadData := map[string]interface{}{
		"from":     config.AT_PHONE, //os.Getenv("AT_PHONE"),
		"to":       phone,
		"username": config.AT_USERNAME, //os.Getenv("AT_USERNAME"),
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
	req.Header.Add("apiKey", config.AT_API_KEY) // os.Getenv("AT_API_KEY"))
	req.Header.Add("Content-Type", "application/json")

	var res *http.Response
	res, err = client.Do(req)
	if err != nil {
		fmt.Println(err)
		// return err
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": "failed",
			"error":  err.Error(),
		})
	}
	defer res.Body.Close()

	var body []byte
	body, err = io.ReadAll(res.Body)
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
	var clientLanguage int64 = 1 // default to English
	var clientID int
	if err := c.BodyParser(&callback); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}

	fmt.Printf("Callback: SessionID=%s, Phone=%s, DTMF=%s ",
		callback.SessionID, callback.CallerNumber, callback.DtmfDigits)

	if callback.CallSessionState == "Answered" {
		// fmt.Println(">>>>> Call answered")
		client, _ := ClientByHBCPhone(c.Context(), db, callback.CallerNumber[1:])
		clientLanguage = client.HbcLanguage.Int64
		clientID = client.ID
	}

	// Get or create session
	session := getOrCreateSession(callback.SessionID, callback.CallerNumber, clientLanguage, clientID) // session created with client's preferred language, and saved in session. After that session language is used for that client for that session.
	// fmt.Printf("Session: %+v\n", session)
	var response ATResponse

	// Process DTMF input
	if callback.DtmfDigits != "" {
		// fmt.Printf(">>>>Processing input for step %d: %s\n", session.CurrentStep, callback.DtmfDigits)
		if !processInput(session, callback.DtmfDigits) {
			// Invalid input - retry current step
			response = createGetDigitsResponse(session.CurrentStep, session.PatientName, session.Language, true)
		} else {
			// Valid input - move to next step
			session.CurrentStep++
			if session.CurrentStep > 6 {
				// fmt.Printf("<<<<<??????Session responses before completion response: %+v\n", session.Responses)
				// Assessment complete
				response = createCompletionResponse()
			} else {
				// Continue to next step
				// fmt.Printf("<<<<<[[[]]]>>>>Session responses so far: %+v\n", session.Responses)
				response = createGetDigitsResponse(session.CurrentStep, session.PatientName, session.Language, false)
			}
		}
	} else {
		// Assessment complete
		if callback.CallSessionState == "Completed" {
			fmt.Printf(">>>>>Session responses at call completion: %+v\n", session.Responses)
			saveMpoxResponse(session, callback.Amount, callback.DurationInSeconds, db)
			deleteSession(session.SessionID)
		} else {
			// Assessment not complete
			// Start or continue current step
			response = createGetDigitsResponse(session.CurrentStep, session.PatientName, session.Language, false)
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

func getOrCreateSession(sessionID, phoneNumber string, languageID int64, clientID int) *Session {
	mutex.Lock()
	defer mutex.Unlock()

	if session, exists := sessions[sessionID]; exists {
		return session
	}

	languages := map[int64]string{
		1: "english",
		2: "luganda",
		3: "other",
	}

	// Create new session - get patient name from DB in production
	session := &Session{
		SessionID:    sessionID,
		PhoneNumber:  phoneNumber,
		PatientName:  "Patient",             // Fetch from database: getPatientName(phoneNumber)
		Language:     languages[languageID], // Default language; fetch from DB if needed
		CurrentStep:  1,
		Responses:    MpoxAssessment{},
		LastActivity: time.Now(),
		ClientID:     clientID,
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
			session.Responses.PatientCondition, _ = strconv.Atoi(input)
			return true
		}
	case 2: // Pain level (0-10)
		if level, err := strconv.Atoi(input); err == nil && level >= 0 && level <= 10 {
			session.Responses.PainLevel = level
			return true
		}
	case 3: // New Lesions
		if input == "1" || input == "2" {
			session.Responses.NewLesions = (input == "1")
			return true
		}
	case 4: // Comorbidities
		if input == "1" || input == "2" {
			session.Responses.Comorbidities = (input == "1")
			return true
		}
	case 5: // Lesions Dry
		if input == "1" || input == "2" {
			session.Responses.LesionsDry = (input == "1")
			return true
		}
	case 6: // Exposure Alert
		if input == "1" || input == "2" {
			session.Responses.ExposureAlert = (input == "1")
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

func createGetDigitsResponse(step int, patientName string, language string, isRetry bool) ATResponse {
	text := fmt.Sprintf(stepPrompts[step])
	voice_url := "https://response.health.go.ug/audios/" + language + "/0" + strconv.Itoa(step) + ".wav"
	fmt.Println("Voice URL for step", step, ":", voice_url)

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
			NumDigits:   numDigits,
			CallbackUrl: "https://response.health.go.ug/voice/callback", //"https://pxvs54rm-3001.uks1.devtunnels.ms/voice/callback", // Replace with your domain
			Play: Play{
				Url: voice_url,
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

func saveMpoxResponse(session *Session, amount string, duration string, db *sql.DB) {
	// 1. Parse amount to float64 and then to int	
	amt, _ := strconv.ParseFloat(amount, 64)
	intAmount := int(math.Round(amt)) // Convert to int for DB consistency

	// 2. Map directly to your struct (No JSON overhead)
	assessmentData := MpoxAssessment{
		Platform:         "ivr",
		PhoneNumber:      session.PhoneNumber,
		// Pull values directly from the fixed memory slots
		PatientCondition: session.Responses.PatientCondition,
		PainLevel:        session.Responses.PainLevel,
		NewLesions:       session.Responses.NewLesions,
		Comorbidities:    session.Responses.Comorbidities,
		LesionsDry:       session.Responses.LesionsDry,
		ExposureAlert:    session.Responses.ExposureAlert,
	}

	fmt.Printf("Mpox Assessment Prepared: %+v\n", assessmentData)

	// 3. Urgent Alert Logic
	if session.Responses.PatientCondition == 3 || session.Responses.PatientCondition == 5 || session.Responses.PainLevel >= 8 || assessmentData.ExposureAlert {
		// Pass the struct directly
		sendUrgentAlert(session.PatientName, session.PhoneNumber)
	}

	// 4. Save Call Session
	var callID int
	queryCall := `INSERT INTO call_sessions (phone_number, duration_in_seconds, amount) 
	              VALUES ($1, $2, $3) RETURNING id`
	
	err := db.QueryRow(queryCall, assessmentData.PhoneNumber, duration, intAmount).Scan(&callID)
	if err != nil {
		log.Printf("Error saving call session: %v", err)
		return
	}

	// 5. Save Assessment (Using the verified callID and session.ClientID)
	queryAss := `INSERT INTO mpox_assessments 
		(client_id, call_session_id, platform, phone_number, patient_condition, pain_level, new_lesions, comorbidities, lesions_dry, exposure_alert) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	
	_, err = db.Exec(queryAss, 
		session.ClientID, 
		callID, 
		assessmentData.Platform, 
		assessmentData.PhoneNumber, 
		assessmentData.PatientCondition, 
		assessmentData.PainLevel, 
		assessmentData.NewLesions, 
		assessmentData.Comorbidities, 
		assessmentData.LesionsDry, 
		assessmentData.ExposureAlert,
	)

	if err != nil {
		log.Printf("Error saving mpox assessment: %v", err)
	}
}

func sendUrgentAlert(patientName, phoneNumber string) {
	fmt.Printf("🚨 URGENT ALERT: Patient %s (%s) requires immediate attention", patientName, phoneNumber)

	// In production, implement:
	// - SMS to health worker: sendSMS(healthWorkerNumber, alertMessage)
	// - Email notification: sendEmail(healthWorkerEmail, alertDetails)
	// - Push notification: sendPushNotification(alertData)
	// - Create support ticket: createTicket(patientData, response)
}

func loadConfig() (*Config, error) {
	// Try multiple possible config file locations (same as cmd/web/main.go)
	configPaths := []string{
		"config.json",
		"cmd/web/config.json",
		"../../cmd/web/config.json",
	}
	var configData []byte
	var err error
	for _, p := range configPaths {
		configData, err = os.ReadFile(p)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var config Config
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	return &config, nil
}
