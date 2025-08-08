package models

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type MpoxDailyFollowUp struct {
	ID                 int             `db:"id"`
	ClientID           int             `db:"client_id"`
	FollowUpDate       sql.NullTime    `db:"followup_date"`
	EncounterType      pq.StringArray  `db:"encounter_type"` // use pq.StringArray for PostgreSQL array
	OtherSite          sql.NullString  `db:"other_site"`
	SpO2               sql.NullInt64   `db:"spo2"`
	NewLesions         sql.NullBool    `db:"new_lesions"`
	DiseaseProgression sql.NullString  `db:"disease_progression"`
	ProgressionOther   sql.NullString  `db:"progression_other"`
	LesionFace         sql.NullString  `db:"lesion_face"`
	LesionMouth        sql.NullString  `db:"lesion_mouth"`
	LesionChest        sql.NullString  `db:"lesion_chest"`
	LesionAbdomen      sql.NullString  `db:"lesion_abdomen"`
	LesionBack         sql.NullString  `db:"lesion_back"`
	LesionArms         sql.NullString  `db:"lesion_arms"`
	LesionPalms        sql.NullString  `db:"lesion_palms"`
	LesionForearms     sql.NullString  `db:"lesion_forearms"`
	LesionThighs       sql.NullString  `db:"lesion_thighs"`
	LesionLegs         sql.NullString  `db:"lesion_legs"`
	LesionSoles        sql.NullString  `db:"lesion_soles"`
	LesionGenitalia    sql.NullString  `db:"lesion_genitalia"`
	LesionPerianal     sql.NullString  `db:"lesion_perianal"`
	LesionOther        sql.NullString  `db:"lesion_other"`
	LesionSpecifyWhere sql.NullString  `db:"lesion_specify_where"`
	TypeMacule         sql.NullString  `db:"type_macule"`
	TypePapule         sql.NullString  `db:"type_papule"`
	TypeVesicle        sql.NullString  `db:"type_vesicle"`
	TypePustule        sql.NullString  `db:"type_pustule"`
	TypeUmbilicated    sql.NullString  `db:"type_umbilicated"`
	TypeUlcerated      sql.NullString  `db:"type_ulcerated"`
	TypeCrusting       sql.NullString  `db:"type_crusting"`
	TypePartialScab    sql.NullString  `db:"type_partialscab"`
	TypeOther          sql.NullString  `db:"type_other"`
	PainPresent        sql.NullBool    `db:"pain_present"`
	PainScore          sql.NullInt64   `db:"pain_score"`
	PainDescription    sql.NullString  `db:"pain_description"`
	Temperature        sql.NullFloat64 `db:"temperature"`
	HeartRate          sql.NullInt64   `db:"heart_rate"`
	RespiratoryRate    sql.NullInt64   `db:"respiratory_rate"`
	BpSystolic         sql.NullInt64   `db:"bp_systolic"`
	BpDiastolic        sql.NullInt64   `db:"bp_diastolic"`
	Consciousness      sql.NullString  `db:"consciousness"`
	DataEntrant        sql.NullString  `db:"data_entrant"`
	CreatedAt          time.Time       `db:"created_at"`
}

func (m *MpoxDailyFollowUp) Insert(db *sql.DB) error {
	query := `
		INSERT INTO mpox_daily_followup (
			client_id, followup_date, encounter_type, other_site, spo2, new_lesions, 
			disease_progression, progression_other, lesion_face, lesion_mouth, lesion_chest, 
			lesion_abdomen, lesion_back, lesion_arms, lesion_palms, lesion_forearms, 
			lesion_thighs, lesion_legs, lesion_soles, lesion_genitalia, lesion_perianal, 
			lesion_other, lesion_specify_where, type_macule, type_papule, type_vesicle, 
			type_pustule, type_umbilicated, type_ulcerated, type_crusting, type_partialscab, 
			type_other, pain_present, pain_score, pain_description, temperature, heart_rate, 
			respiratory_rate, bp_systolic, bp_diastolic, consciousness, data_entrant, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, $42, NOW()
		) RETURNING id`
	return db.QueryRow(query,
		m.ClientID, m.FollowUpDate, m.EncounterType, m.OtherSite, m.SpO2, m.NewLesions,
		m.DiseaseProgression, m.ProgressionOther, m.LesionFace, m.LesionMouth, m.LesionChest,
		m.LesionAbdomen, m.LesionBack, m.LesionArms, m.LesionPalms, m.LesionForearms,
		m.LesionThighs, m.LesionLegs, m.LesionSoles, m.LesionGenitalia, m.LesionPerianal,
		m.LesionOther, m.LesionSpecifyWhere, m.TypeMacule, m.TypePapule, m.TypeVesicle,
		m.TypePustule, m.TypeUmbilicated, m.TypeUlcerated, m.TypeCrusting, m.TypePartialScab,
		m.TypeOther, m.PainPresent, m.PainScore, m.PainDescription, m.Temperature, m.HeartRate,
		m.RespiratoryRate, m.BpSystolic, m.BpDiastolic, m.Consciousness, m.DataEntrant,
	).Scan(&m.ID)
}
