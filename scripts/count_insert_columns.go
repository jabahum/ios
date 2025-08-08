package main

import (
	"fmt"
	"strings"
)

func main() {
	// The INSERT statement from the model
	insertQuery := `
		INSERT INTO mpox_onset_vitals (
			demographics_id, symptom_onset, fever, fever_onset_date, sore_throat, sore_throat_onset_date,
			headache, headache_onset_date, muscle_aches, muscle_aches_onset_date, cough, cough_onset_date,
			fatigue, fatigue_onset_date, oral_pain, oral_pain_onset_date, nausea, nausea_onset_date,
			vomiting, vomiting_onset_date, diarrhea, diarrhea_onset_date, rectal_pain, rectal_pain_onset_date,
			lesions, lesions_onset_date, lymphadenopathy, lymphadenopathy_onset_date, pruritis, pruritis_onset_date,
			pain_swallowing, pain_swallowing_onset_date, difficulty_swallowing, difficulty_swallowing_onset_date,
			urethritis, urethritis_onset_date, chest_pain, chest_pain_onset_date, decreased_urine, decreased_urine_onset_date,
			dizziness, dizziness_onset_date, joint_pain, joint_pain_onset_date, psychological_disturbance, psychological_disturbance_onset_date,
			temperature, heart_rate, respiratory_rate, bp_systolic, bp_diastolic, dehydration, avpu, height_cm, weight_kg,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30,
			$31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, $42, $43, $44,
			$45, $46, $47, $48, $49, $50, $51, $52, $53, $54, $55, $56, $57, $58,
			CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		) RETURNING id`

	// Extract column names from the INSERT statement
	startIndex := strings.Index(insertQuery, "(") + 1
	endIndex := strings.Index(insertQuery, ")")
	columnsPart := insertQuery[startIndex:endIndex]

	// Split by comma and count
	columns := strings.Split(columnsPart, ",")
	fmt.Printf("Number of columns in INSERT statement: %d\n", len(columns))

	fmt.Println("\nColumns in INSERT statement:")
	for i, col := range columns {
		col = strings.TrimSpace(col)
		fmt.Printf("%d. %s\n", i+1, col)
	}
}
