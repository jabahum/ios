package models

import (
	// "database/sql"
	// "case/internal/database"
	// "case/internal/models"
	"github.com/gofiber/fiber/v2"
	"fmt"
	"net/http"
	"io"
	// "strings"
	"os"
	"log"
	"time"
	// "sync"
	// "strconv"
	"encoding/xml"
	"encoding/json"
    "bytes"

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
	SessionID    string `json:"sessionId"`
	CallerNumber string `json:"callerNumber"`
	DtmfDigits   string `json:"dtmfDigits"`
	Direction    string `json:"direction"`
	Carrier      string `json:"callerCarrierName"`
	CountryCode  string `json:"callerCountryCode"`
	DurationInSeconds int `json:"durationInSeconds"`
	Amount int            `json:"amount"`
}

// Africa's Talking XML response structures
type ATResponse struct {
	XMLName xml.Name `xml:"Response"`
	GetDigits *GetDigits `xml:"GetDigits,omitempty"`
	Say       *Say       `xml:"Say,omitempty"`
}

type GetDigits struct {
	Timeout     int    `xml:"timeout,attr"`
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

var stepPrompts = map[int]string{}

func SendCall(c *fiber.Ctx) error {
	phoneNumber := c.Query("phone")
	var phone =[...]string{phoneNumber}

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

	client := &http.Client {}

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
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(string(body))

	return c.Status(200).JSON(body)
}

// func MakeVoiceCall(c *fiber.Ctx) error {
// 	// get survey from database
// 	var qns []models.SurveyQuestion
// 	database.DB.Db.Where("survey_id = 5").Find(&qns)
// 	for _,value := range qns {
// 		stepPrompts[value.QuestionNo] = value.Question
// 	}

//   url := "https://voice.africastalking.com/call"
//   method := "POST"

//   // *********
// 	// Query phone numbers
// 	var recipients []models.Patient
// 	database.DB.Db.Where("pathway = 'ivr'").Find(&recipients)
//     // Collect phone numbers into slice
// 	var toNumbers []string
// 	for _, r := range recipients {
// 		toNumbers = append(toNumbers, r.PhoneNumber)
// 	}

// 	// Build payload
//     payloadData := map[string]interface{}{
//         "from":     "+256323200977",
//         "to":       toNumbers,
//         "username": "hbc_user",
//     }

//     payloadBytes, err := json.Marshal(payloadData)
//     if err != nil {
//         log.Fatal("Error marshaling payload:", err)
//     }

//     payload := bytes.NewReader(payloadBytes)
//   // *********
//   // Debug print
//   fmt.Println("Final JSON payload:", string(payloadBytes))
// // panic(payload)

// //   payload := strings.NewReader(`{
// //       "from": "+256323200977",
// //       "to": ["+256782866580"],
// //       "username": "hbc_user"
// //   }`)

//   client := &http.Client {
//   }
//   req, err := http.NewRequest(method, url, payload)

//   if err != nil {
//     fmt.Println(err)
//     return err
//   }
//   req.Header.Add("Accept", "application/json")
//   req.Header.Add("apiKey", os.Getenv("AT_API_KEY"))
//   req.Header.Add("Content-Type", "application/json")

//   res, err := client.Do(req)
//   if err != nil {
//     fmt.Println(err)
//     return err
//   }
//   defer res.Body.Close()

//   body, err := io.ReadAll(res.Body)
//   if err != nil {
//     fmt.Println(err)
//     return err
//   }
//   fmt.Println(string(body))

//   return c.Status(200).JSON(body)
// }

// // Global session store (use Redis in production)
// var (
// 	sessions = make(map[string]*models.Session)
// 	mutex    = &sync.RWMutex{}
// )

// // Step prompts for Mpox monitoring
// // var stepPrompts = map[int]string{
// // 	1: "Hello, this is your daily Mpox check-in. Have you developed any new lesions or symptoms in the past 24 hours? Press 1 for No new symptoms, 2 for Yes new symptoms, 3 for I feel worse, 4 for I feel better, or 5 if you need to speak to a health worker.",
// // 	2: "On a scale of 0 to 10, what is your pain level today? 0 means no pain, 10 means worst pain imaginable. Press the number.",
// // 	3: "Have you noticed new lesions today in sensitive areas like mouth, genitals, or eyes? Press 1 for Yes or 2 for No .",
// // 	4: "Are you experiencing complications related to another illness such as HIV, diabetes, or high blood pressure? Press 1 for Yes or 2 for No .",
// // 	5: "Any new lesions today in sensitive areas like mouth, genitals, or eyes? Press 1 for Yes or 2 for No .",
// // 	6: "Has anyone in your home or neighbourhood developed Mpox-like symptoms in the last 24 hours? Press 1 for Yes or 2 for No .",
// // }

// func HandleVoiceCallback(c *fiber.Ctx) error {
// 	var callback models.VoiceCallback
// 	if err := c.BodyParser(&callback); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Invalid JSON",
// 		})
// 	}

// 	log.Printf("Callback: SessionID=%s, Phone=%s, DTMF=%s",
// 		callback.SessionID, callback.CallerNumber, callback.DtmfDigits)

// 	// Get or create session
// 	session := getOrCreateSession(callback.SessionID, callback.CallerNumber)
// 	var response models.ATResponse

// 	// Process DTMF input
// 	if callback.DtmfDigits != "" {
// 		if !processInput(session, callback.DtmfDigits) {
// 			// Invalid input - retry current step
// 			response = createGetDigitsResponse(session.CurrentStep, session.PatientName, true)
// 		} else {
// 			// Valid input - move to next step
// 			session.CurrentStep++
// 			if session.CurrentStep > 6 {
// 				// Assessment complete
// 				saveMpoxResponse(session, callback.Amount, callback.DurationInSeconds)
// 				response = createCompletionResponse()
// 				deleteSession(session.SessionID)
// 			} else {
// 				// Continue to next step
// 				response = createGetDigitsResponse(session.CurrentStep, session.PatientName, false)
// 			}
// 		}
// 	} else {
// 		// Start or continue current step
// 		response = createGetDigitsResponse(session.CurrentStep, session.PatientName, false)
// 	}

// 	session.LastActivity = time.Now()
// 	updateSession(session)

// 	// Marshal response to XML
// 	output, err := xml.Marshal(response)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).SendString("Error encoding XML")
// 	}

// 	// Send XML response with header
// 	c.Type("xml")
// 	return c.SendString(`<?xml version="1.0" encoding="UTF-8"?>` + string(output))
// }

// func getOrCreateSession(sessionID, phoneNumber string) *models.Session {
// 	mutex.Lock()
// 	defer mutex.Unlock()

// 	if session, exists := sessions[sessionID]; exists {
// 		return session
// 	}

// 	// Create new session - get patient name from DB in production
// 	session := &models.Session{
// 		SessionID:    sessionID,
// 		PhoneNumber:  phoneNumber,
// 		PatientName:  "Patient", // Fetch from database: getPatientName(phoneNumber)
// 		CurrentStep:  1,
// 		Responses:    make(map[string]string),
// 		LastActivity: time.Now(),
// 	}

// 	sessions[sessionID] = session
// 	return session
// }

// func updateSession(session *models.Session) {
// 	mutex.Lock()
// 	defer mutex.Unlock()
// 	sessions[session.SessionID] = session
// }

// func deleteSession(sessionID string) {
// 	mutex.Lock()
// 	defer mutex.Unlock()
// 	delete(sessions, sessionID)
// }

// func processInput(session *models.Session, input string) bool {
// 	input = strings.TrimSpace(input)
// 	step := session.CurrentStep
	
// 	switch step {
// 	case 1: // General status (1-5)
// 		if input >= "1" && input <= "5" {
// 			session.Responses["general_status"] = input
// 			return true
// 		}
// 	case 2: // Pain level (0-10)
// 		if level, err := strconv.Atoi(input); err == nil && level >= 0 && level <= 10 {
// 			session.Responses["pain_level"] = input
// 			return true
// 		}
// 	case 3, 4, 5, 6: // Y/N questions (1=Yes, 2=No)
// 		if input == "1" || input == "2" {
// 			stepKey := getStepKey(step)
// 			session.Responses[stepKey] = input
// 			return true
// 		}
// 	}
	
// 	return false
// }

// func getStepKey(step int) string {
// 	keys := map[int]string{
// 		3: "sensitive_lesions",
// 		4: "comorbidities",
// 		5: "new_lesions", 
// 		6: "exposure_alert",
// 	}
// 	return keys[step]
// }

// func createGetDigitsResponse(step int, patientName string, isRetry bool) models.ATResponse {
// 	text := fmt.Sprintf(stepPrompts[step])
	
// 	if isRetry {
// 		text = "I didn't understand your response. " + text
// 	}

// 	numDigits := 1
// 	if step == 2 { // Pain level can be 2 digits
// 		numDigits = 2
// 	}

// 	return models.ATResponse{
// 		GetDigits: &models.GetDigits{
// 			Timeout:     8,
// 			// FinishOnKey: "#",
// 			NumDigits:   numDigits,
// 			CallbackUrl: "https://pxvs54rm-3000.uks1.devtunnels.ms/voice/callback", // Replace with your domain 
// 			Say: models.Say{
// 				Voice:    "woman",
// 				PlayBeep: "true",
// 				Text:     text,
// 			},
// 		},
// 	}
// }

// func createCompletionResponse() models.ATResponse {
// 	return models.ATResponse{
// 		Say: &models.Say{
// 			Voice: "woman",
// 			PlayBeep: "false",
// 			Text:  "Thank you for completing your daily Mpox check-in. Your responses have been recorded and a health worker will contact you if needed. Stay safe.",
// 		},
// 	}
// }

// func saveMpoxResponse(session *models.Session, amount int, duration int) {
// 	painLevel, _ := strconv.Atoi(session.Responses["pain_level"])
	
// 	// Create structured response
// 	response := map[string]interface{}{
// 		"session_id":        session.SessionID,
// 		"phone_number":      session.PhoneNumber,
// 		"patient_name":      session.PatientName,
// 		"general_status":    session.Responses["general_status"],
// 		"pain_level":        painLevel,
// 		"sensitive_lesions": session.Responses["sensitive_lesions"] == "1",
// 		"comorbidities":     session.Responses["comorbidities"] == "1",
// 		"new_lesions":       session.Responses["new_lesions"] == "1",
// 		"exposure_alert":    session.Responses["exposure_alert"] == "1",
// 		"completed_at":      time.Now(),
// 		"amount":            amount,
// 		"duration":          duration,
// 		"platform":          "voice",
// 	}

// 	log.Printf("Mpox Assessment Complete: %+v", response)
	
// 	// Check for urgent conditions and send alerts
// 	if session.Responses["general_status"] == "3" || // Feel worse
// 	   session.Responses["general_status"] == "5" || // Need health worker
// 	   painLevel >= 8 ||                             // High pain
// 	   session.Responses["exposure_alert"] == "1" {  // Exposure risk
		
// 		sendUrgentAlert(session.PatientName, session.PhoneNumber, response)
// 	}
	
// 	// In production: save to database
// 	// db.Create(&MpoxResponse{...})
// 	jsonBytes, _ := json.Marshal(response)
// 	var assessmentData models.Assessment
// 	// convert jsonBytes to Json and bind it to assessmentData
//     err := json.Unmarshal(jsonBytes, &assessmentData)
//     if err != nil {
//         panic(err)
//     }
// 	// save assessment to database
// 	result := database.DB.Db.Create(&assessmentData)
// 	if result.Error != nil {
// 		log.Println("Error: %v", result.Error)
// 		return
// 	}

// 	return
// }

// func sendUrgentAlert(patientName, phoneNumber string, response map[string]interface{}) {
// 	log.Printf("🚨 URGENT ALERT: Patient %s (%s) requires immediate attention", patientName, phoneNumber)
	
// 	// In production, implement:
// 	// - SMS to health worker: sendSMS(healthWorkerNumber, alertMessage)
// 	// - Email notification: sendEmail(healthWorkerEmail, alertDetails)
// 	// - Push notification: sendPushNotification(alertData)
// 	// - Create support ticket: createTicket(patientData, response)
// }