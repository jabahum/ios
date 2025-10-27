package models

import (
	"database/sql"
	"time"
)

type Survey struct {
	ID        int               `json:"id"`
	Title     string            `json:"title"`
	Questions []SurveyQuestion  `json:"questions"`
	CompletedMessage string     `json:"completed_message"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type SurveyQuestion struct {
	ID         int           `json:"id"`
	SurveyID   int           `json:"survey_id"`
	QuestionNo int           `json:"question_no"`
	Question   string        `json:"question"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

func GetSurveys(db *sql.DB) ([]Survey, error) {
	rows, err := db.Query("SELECT * FROM surveys")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var surveys []Survey
	for rows.Next() {
		var s Survey
		if err := rows.Scan(&s.ID, &s.Title, &s.CompletedMessage, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		surveys = append(surveys, s)
	}
	return surveys, nil
}

// Get survey_questions by survey ID
func GetQuestionsBySurvey(db *sql.DB, surveyID int) ([]SurveyQuestion, error) {
	rows, err := db.Query("SELECT id, survey_id, question_no, question, created_at, updated_at FROM subcounties WHERE survey_id = $1 ORDER BY question_no", surveyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var survey_questions []SurveyQuestion
	for rows.Next() {
		var sq SurveyQuestion
		if err := rows.Scan(&sq.ID, &sq.SurveyID, &sq.QuestionNo, &sq.Question, &sq.CreatedAt, &sq.UpdatedAt); err != nil {
			return nil, err
		}
		survey_questions = append(survey_questions, sq)
	}
	return survey_questions, nil
}
