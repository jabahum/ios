package models

// import (
// 	"case/internal/database"
// 	"github.com/gofiber/fiber/v2"
// 	"net/http"
// 	"bytes"
// 	"encoding/json"
// 	"time"
// 	"log"
// 	"strconv"
// )

// // Message represents a WhatsApp message stored in DB
// type Message struct {
// 	ID         uint      `gorm:"primaryKey"`
// 	From       string
// 	Body       string
// 	Reply      string
// 	Timestamp  time.Time
// 	QuestionNo int
// }

// func HandleWebhookMessage(c *fiber.Ctx) error {
// 	var payload map[string]interface{}

// 	if err := c.BodyParser(&payload); err != nil {
// 		log.Println("Failed to parse payload:", err)
// 		return c.SendStatus(fiber.StatusBadRequest)
// 	}

// 	// Dig into the payload to extract message
// 	entries := payload["entry"].([]interface{})
// 	changes := entries[0].(map[string]interface{})["changes"].([]interface{})
// 	value := changes[0].(map[string]interface{})["value"].(map[string]interface{})

// 	if value["messages"] == nil {
// 		return nil
// 	}

// 	messages := value["messages"].([]interface{})
// 	msg := messages[0].(map[string]interface{})
// 	from := msg["from"].(string)
// 	body := msg["text"].(map[string]interface{})["body"].(string)

// 	// Determine next question
// 	var lastMsg models.Message
// 	database.DB.Db.Where("messages.from = ?", "+"+from).Order("id desc").First(&lastMsg)

// 	nextQuestionNo := lastMsg.QuestionNo + 1
// 	if nextQuestionNo > 6 {
// 		// Survey complete — collect answers into SurveyResult
// 		var answers []models.Message
// 		database.DB.Db.Where("messages.from = ?", "+"+from).Order("question_no asc").Find(&answers)

// 		// Map answers to fields (assuming survey has name, age, sex)
// 		assessment := models.Assessment{PhoneNumber: "+"+from, PatientName: "Patient", Platform: "Whatsapp", SessionId: "sess_"+time.Now().Format("2006-01-02"), CompletedAt: time.Now().Format("2006-01-02 11:25:36") }
// 		for _, ans := range answers {
// 			var q models.SurveyQuestion
// 			database.DB.Db.Where("order = ?", ans.QuestionNo).First(&q)
// 			switch q.QuestionNo {
// 			case 1:
// 				assessment.GeneralStatus = ans.Body
// 			case 2:
// 				assessment.PainLevel, _ = strconv.Atoi(ans.Body)
// 			case 3:
// 				assessment.SensitiveLesions, _ = strconv.ParseBool(ans.Body)
// 			case 4:
// 				assessment.Comorbidities, _ = strconv.ParseBool(ans.Body)
// 			case 5:
// 				assessment.NewLesions, _ = strconv.ParseBool(ans.Body)
// 			case 6:
// 				assessment.ExposureAlert, _ = strconv.ParseBool(ans.Body)
// 			}
// 		}
// 		database.DB.Db.Create(&assessment)

// 		reply := "Thank you for completing your daily Mpox check-in. Your responses have been recorded and a health worker will contact you if needed. Stay safe."
// 		storeMessage(from, body, reply, 0)
// 		sendReply(from, reply)
// 		return c.SendStatus(fiber.StatusOK)
// 	}

// 	var qns []models.SurveyQuestion
// 	database.DB.Db.Where("survey_id = 5").Find(&qns)
// 	questions := map[int]string{}
// 	for _,value := range qns {
// 		questions[value.QuestionNo] = value.Question
// 	}

// 	reply := questions[nextQuestionNo]
// 	storeMessage(from, body, reply, nextQuestionNo)
// 	sendReply(from, reply)

// 	return c.SendStatus(fiber.StatusOK)
// }

// func saveWhatsappMessage(phone, content, direction string) {
// 	database.DB.Db.Create(&models.Message{
// 		Phone:      "+"+phone,
// 		Content:    content,
// 		Direction:  direction,
// 		Platform:   "whatsapp",
// 	})
// }

// func storeMessage(from, body, reply string, qno int) {
// 	database.DB.Db.Create(&models.Message{
// 		From:       "+"+from,
// 		Body:       body,
// 		Reply:      reply,
// 		Timestamp:  time.Now(),
// 		QuestionNo: qno,
// 	})
// }

// func sendReply(to, reply string) {

// 	url := "https://jkage-consultants.work.gd/send"
// 	method := "POST"
// 	payload := map[string]interface{}{"to": to, "message": reply}
// 	jsonData, _ := json.Marshal(payload)

// 	req, _ := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
// 	req.Header.Set("Content-Type", "application/json")

// 	client := &http.Client{Timeout: 10 * time.Second}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		log.Println("Failed to send reply:", err)
// 		return
// 	}
// 	defer resp.Body.Close()


// }