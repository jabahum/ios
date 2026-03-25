package handlers

import (
	"case/internal/config"
	"case/internal/models"
	"case/internal/security"
	"case/internal/services"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// HandlerVHFPatientSubmit handles the submission of patient information
func HandlerVHFPatientSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config, smsService *services.SMSService) error {
	// Parse form data
	patient := &models.VHFPatient{
		Surname:                     c.FormValue("surname"),
		OtherNames:                  c.FormValue("other_names"),
		DateOfBirth:                 parseNullTime(c.FormValue("dob")),
		AgeYears:                    sql.NullInt32{Int32: parseInt32(c.FormValue("age_years")), Valid: true},
		AgeMonths:                   sql.NullInt32{Int32: parseInt32(c.FormValue("age_months")), Valid: true},
		Gender:                      sql.NullString{String: c.FormValue("gender"), Valid: c.FormValue("gender") != ""},
		PatientPhone:                sql.NullString{String: c.FormValue("patient_phone"), Valid: c.FormValue("patient_phone") != ""},
		PhoneOwner:                  sql.NullString{String: c.FormValue("phone_owner"), Valid: c.FormValue("phone_owner") != ""},
		NextOfKin:                   sql.NullString{String: c.FormValue("next_of_kin"), Valid: c.FormValue("next_of_kin") != ""},
		NextOfKinPhone:              sql.NullString{String: c.FormValue("next_of_kin_phone"), Valid: c.FormValue("next_of_kin_phone") != ""},
		RelationshipToPatient:       sql.NullString{String: c.FormValue("relationship_to_patient"), Valid: c.FormValue("relationship_to_patient") != ""},
		DataCapturerName:            sql.NullString{String: c.FormValue("data_capturer_name"), Valid: c.FormValue("data_capturer_name") != ""},
		DataCapturerPhone:           sql.NullString{String: c.FormValue("data_capturer_phone"), Valid: c.FormValue("data_capturer_phone") != ""},
		ReportingHealthFacilityName: sql.NullString{String: c.FormValue("reporting_health_facility_name"), Valid: c.FormValue("reporting_health_facility_name") != ""},
		CaseCode:                    sql.NullString{String: c.FormValue("case_code"), Valid: c.FormValue("case_code") != ""},
		Status:                      sql.NullString{String: c.FormValue("status"), Valid: c.FormValue("status") != ""},
		HeadOfHousehold:             sql.NullString{String: c.FormValue("head_of_household"), Valid: c.FormValue("head_of_household") != ""},
		VillageTown:                 c.FormValue("village_town"),
		Parish:                      c.FormValue("parish"),
		Subcounty:                   c.FormValue("subcounty"),
		District:                    c.FormValue("district"),
		CountryOfResidence:          c.FormValue("country_of_residence"),
		Occupation:                  c.FormValue("occupation"),
		IllVillageTown:              c.FormValue("ill_village_town"),
		IllSubcounty:                c.FormValue("ill_subcounty"),
		IllDistrict:                 c.FormValue("ill_district"),
		CreatedAt:                   time.Now(),
	}

	// Debug logging for location fields
	sl.Info("Location fields received",
		"district", patient.District,
		"subcounty", patient.Subcounty,
		"parish", patient.Parish,
		"village_town", patient.VillageTown)

	// Save patient data
	if err := models.SaveVHFPatient(db, patient); err != nil {
		sl.Error("Failed to save patient data", "error", err)
		return c.Status(500).SendString("Failed to save patient data")
	}

	// Parse and save clinical signs
	clinicalSigns := &models.VHFClinicalSigns{
		PatientID: patient.ID,
		DateInitialOnset: func() sql.NullTime {
			val := c.FormValue("date_initial_onset")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		TempSource: sql.NullString{String: c.FormValue("temp_source"), Valid: c.FormValue("temp_source") != ""},
		Temperature: func() sql.NullFloat64 {
			val := c.FormValue("temperature")
			f, err := strconv.ParseFloat(val, 64)
			return sql.NullFloat64{Float64: f, Valid: err == nil && val != ""}
		}(),
		Fever: sql.NullBool{Bool: c.FormValue("fever") == "Yes", Valid: c.FormValue("fever") != ""},
		DateFever: func() sql.NullTime {
			val := c.FormValue("date_fever")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationFever: func() sql.NullInt32 {
			val := c.FormValue("duration_fever")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		Vomiting: sql.NullBool{Bool: c.FormValue("vomiting") == "Yes", Valid: c.FormValue("vomiting") != ""},
		DateVomiting: func() sql.NullTime {
			val := c.FormValue("date_vomiting")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationVomiting: func() sql.NullInt32 {
			val := c.FormValue("duration_vomiting")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		Nausea: sql.NullBool{Bool: c.FormValue("nausea") == "Yes", Valid: c.FormValue("nausea") != ""},
		DateNausea: func() sql.NullTime {
			val := c.FormValue("date_nausea")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationNausea: func() sql.NullInt32 {
			val := c.FormValue("duration_nausea")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		Diarrhea: sql.NullBool{Bool: c.FormValue("diarrhea") == "Yes", Valid: c.FormValue("diarrhea") != ""},
		DateDiarrhea: func() sql.NullTime {
			val := c.FormValue("date_diarrhea")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationDiarrhea: func() sql.NullInt32 {
			val := c.FormValue("duration_diarrhea")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		IntenseFatigueGeneralWeakness: sql.NullBool{Bool: c.FormValue("intense_fatigue_general_weakness") == "Yes", Valid: c.FormValue("intense_fatigue_general_weakness") != ""},
		DateIntenseFatigueGeneralWeakness: func() sql.NullTime {
			val := c.FormValue("date_intense_fatigue_general_weakness")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationIntenseFatigueGeneralWeakness: func() sql.NullInt32 {
			val := c.FormValue("duration_intense_fatigue_general_weakness")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		EpigastricPain: sql.NullBool{Bool: c.FormValue("epigastric_pain") == "Yes", Valid: c.FormValue("epigastric_pain") != ""},
		DateEpigastricPain: func() sql.NullTime {
			val := c.FormValue("date_epigastric_pain")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationEpigastricPain: func() sql.NullInt32 {
			val := c.FormValue("duration_epigastric_pain")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		LowerAbdominalPain: sql.NullBool{Bool: c.FormValue("lower_abdominal_pain") == "Yes", Valid: c.FormValue("lower_abdominal_pain") != ""},
		DateLowerAbdominalPain: func() sql.NullTime {
			val := c.FormValue("date_lower_abdominal_pain")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationLowerAbdominalPain: func() sql.NullInt32 {
			val := c.FormValue("duration_lower_abdominal_pain")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		ChestPain: sql.NullBool{Bool: c.FormValue("chest_pain") == "Yes", Valid: c.FormValue("chest_pain") != ""},
		DateChestPain: func() sql.NullTime {
			val := c.FormValue("date_chest_pain")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationChestPain: func() sql.NullInt32 {
			val := c.FormValue("duration_chest_pain")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		MusclePain: sql.NullBool{Bool: c.FormValue("muscle_pain") == "Yes", Valid: c.FormValue("muscle_pain") != ""},
		DateMusclePain: func() sql.NullTime {
			val := c.FormValue("date_muscle_pain")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationMusclePain: func() sql.NullInt32 {
			val := c.FormValue("duration_muscle_pain")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		JointPain: sql.NullBool{Bool: c.FormValue("joint_pain") == "Yes", Valid: c.FormValue("joint_pain") != ""},
		DateJointPain: func() sql.NullTime {
			val := c.FormValue("date_joint_pain")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationJointPain: func() sql.NullInt32 {
			val := c.FormValue("duration_joint_pain")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		Headache: sql.NullBool{Bool: c.FormValue("headache") == "Yes", Valid: c.FormValue("headache") != ""},
		DateHeadache: func() sql.NullTime {
			val := c.FormValue("date_headache")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationHeadache: func() sql.NullInt32 {
			val := c.FormValue("duration_headache")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		Cough: sql.NullBool{Bool: c.FormValue("cough") == "Yes", Valid: c.FormValue("cough") != ""},
		DateCough: func() sql.NullTime {
			val := c.FormValue("date_cough")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationCough: func() sql.NullInt32 {
			val := c.FormValue("duration_cough")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		DifficultyBreathing: sql.NullBool{Bool: c.FormValue("difficulty_breathing") == "Yes", Valid: c.FormValue("difficulty_breathing") != ""},
		DateDifficultyBreathing: func() sql.NullTime {
			val := c.FormValue("date_difficulty_breathing")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationDifficultyBreathing: func() sql.NullInt32 {
			val := c.FormValue("duration_difficulty_breathing")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		DifficultySwallowing: sql.NullBool{Bool: c.FormValue("difficulty_swallowing") == "Yes", Valid: c.FormValue("difficulty_swallowing") != ""},
		DateDifficultySwallowing: func() sql.NullTime {
			val := c.FormValue("date_difficulty_swallowing")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationDifficultySwallowing: func() sql.NullInt32 {
			val := c.FormValue("duration_difficulty_swallowing")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		SoreThroat: sql.NullBool{Bool: c.FormValue("sore_throat") == "Yes", Valid: c.FormValue("sore_throat") != ""},
		DateSoreThroat: func() sql.NullTime {
			val := c.FormValue("date_sore_throat")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationSoreThroat: func() sql.NullInt32 {
			val := c.FormValue("duration_sore_throat")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		Jaundice: sql.NullBool{Bool: c.FormValue("jaundice") == "Yes", Valid: c.FormValue("jaundice") != ""},
		DateJaundice: func() sql.NullTime {
			val := c.FormValue("date_jaundice")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationJaundice: func() sql.NullInt32 {
			val := c.FormValue("duration_jaundice")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		Conjunctivitis: sql.NullBool{Bool: c.FormValue("conjunctivitis") == "Yes", Valid: c.FormValue("conjunctivitis") != ""},
		DateConjunctivitis: func() sql.NullTime {
			val := c.FormValue("date_conjunctivitis")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationConjunctivitis: func() sql.NullInt32 {
			val := c.FormValue("duration_conjunctivitis")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		SkinRash: sql.NullBool{Bool: c.FormValue("skin_rash") == "Yes", Valid: c.FormValue("skin_rash") != ""},
		DateSkinRash: func() sql.NullTime {
			val := c.FormValue("date_skin_rash")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationSkinRash: func() sql.NullInt32 {
			val := c.FormValue("duration_skin_rash")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		Hiccups: sql.NullBool{Bool: c.FormValue("hiccups") == "Yes", Valid: c.FormValue("hiccups") != ""},
		DateHiccups: func() sql.NullTime {
			val := c.FormValue("date_hiccups")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationHiccups: func() sql.NullInt32 {
			val := c.FormValue("duration_hiccups")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		PainBehindEyes: sql.NullBool{Bool: c.FormValue("pain_behind_eyes") == "Yes", Valid: c.FormValue("pain_behind_eyes") != ""},
		DatePainBehindEyes: func() sql.NullTime {
			val := c.FormValue("date_pain_behind_eyes")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationPainBehindEyes: func() sql.NullInt32 {
			val := c.FormValue("duration_pain_behind_eyes")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		SensitiveToLight: sql.NullBool{Bool: c.FormValue("sensitive_to_light") == "Yes", Valid: c.FormValue("sensitive_to_light") != ""},
		DateSensitiveToLight: func() sql.NullTime {
			val := c.FormValue("date_sensitive_to_light")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationSensitiveToLight: func() sql.NullInt32 {
			val := c.FormValue("duration_sensitive_to_light")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		ComaUnconscious: sql.NullBool{Bool: c.FormValue("coma_unconscious") == "Yes", Valid: c.FormValue("coma_unconscious") != ""},
		DateComaUnconscious: func() sql.NullTime {
			val := c.FormValue("date_coma_unconscious")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationComaUnconscious: func() sql.NullInt32 {
			val := c.FormValue("duration_coma_unconscious")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		ConfusedOrDisoriented: sql.NullBool{Bool: c.FormValue("confused_or_disoriented") == "Yes", Valid: c.FormValue("confused_or_disoriented") != ""},
		DateConfusedOrDisoriented: func() sql.NullTime {
			val := c.FormValue("date_confused_or_disoriented")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationConfusedOrDisoriented: func() sql.NullInt32 {
			val := c.FormValue("duration_confused_or_disoriented")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		Convulsions: sql.NullBool{Bool: c.FormValue("convulsions") == "Yes", Valid: c.FormValue("convulsions") != ""},
		DateConvulsions: func() sql.NullTime {
			val := c.FormValue("date_convulsions")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationConvulsions: func() sql.NullInt32 {
			val := c.FormValue("duration_convulsions")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		UnexplainedBleeding: sql.NullBool{Bool: c.FormValue("unexplained_bleeding") == "Yes", Valid: c.FormValue("unexplained_bleeding") != ""},
		DateUnexplainedBleeding: func() sql.NullTime {
			val := c.FormValue("date_unexplained_bleeding")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationUnexplainedBleeding: func() sql.NullInt32 {
			val := c.FormValue("duration_unexplained_bleeding")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		BleedingOfTheGums: sql.NullBool{Bool: c.FormValue("bleeding_of_the_gums") == "Yes", Valid: c.FormValue("bleeding_of_the_gums") != ""},
		DateBleedingOfTheGums: func() sql.NullTime {
			val := c.FormValue("date_bleeding_of_the_gums")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationBleedingOfTheGums: func() sql.NullInt32 {
			val := c.FormValue("duration_bleeding_of_the_gums")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		BleedingFromInjectionSite: sql.NullBool{Bool: c.FormValue("bleeding_from_injection_site") == "Yes", Valid: c.FormValue("bleeding_from_injection_site") != ""},
		DateBleedingFromInjectionSite: func() sql.NullTime {
			val := c.FormValue("date_bleeding_from_injection_site")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationBleedingFromInjectionSite: func() sql.NullInt32 {
			val := c.FormValue("duration_bleeding_from_injection_site")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		NoseBleedEpistaxis: sql.NullBool{Bool: c.FormValue("nose_bleed_epistaxis") == "Yes", Valid: c.FormValue("nose_bleed_epistaxis") != ""},
		DateNoseBleedEpistaxis: func() sql.NullTime {
			val := c.FormValue("date_nose_bleed_epistaxis")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationNoseBleedEpistaxis: func() sql.NullInt32 {
			val := c.FormValue("duration_nose_bleed_epistaxis")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		BloodyStool: sql.NullBool{Bool: c.FormValue("bloody_stool") == "Yes", Valid: c.FormValue("bloody_stool") != ""},
		DateBloodyStool: func() sql.NullTime {
			val := c.FormValue("date_bloody_stool")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationBloodyStool: func() sql.NullInt32 {
			val := c.FormValue("duration_bloody_stool")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		BloodInVomit: sql.NullBool{Bool: c.FormValue("blood_in_vomit") == "Yes", Valid: c.FormValue("blood_in_vomit") != ""},
		DateBloodInVomit: func() sql.NullTime {
			val := c.FormValue("date_blood_in_vomit")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationBloodInVomit: func() sql.NullInt32 {
			val := c.FormValue("duration_blood_in_vomit")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		CoughingUpBloodHemoptysis: sql.NullBool{Bool: c.FormValue("coughing_up_blood_hemoptysis") == "Yes", Valid: c.FormValue("coughing_up_blood_hemoptysis") != ""},
		DateCoughingUpBloodHemoptysis: func() sql.NullTime {
			val := c.FormValue("date_coughing_up_blood_hemoptysis")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationCoughingUpBloodHemoptysis: func() sql.NullInt32 {
			val := c.FormValue("duration_coughing_up_blood_hemoptysis")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		BleedingFromVagina: sql.NullBool{Bool: c.FormValue("bleeding_from_vagina") == "Yes", Valid: c.FormValue("bleeding_from_vagina") != ""},
		DateBleedingFromVagina: func() sql.NullTime {
			val := c.FormValue("date_bleeding_from_vagina")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationBleedingFromVagina: func() sql.NullInt32 {
			val := c.FormValue("duration_bleeding_from_vagina")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		BruisingOfTheSkin: sql.NullBool{Bool: c.FormValue("bruising_of_the_skin") == "Yes", Valid: c.FormValue("bruising_of_the_skin") != ""},
		DateBruisingOfTheSkin: func() sql.NullTime {
			val := c.FormValue("date_bruising_of_the_skin")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationBruisingOfTheSkin: func() sql.NullInt32 {
			val := c.FormValue("duration_bruising_of_the_skin")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		BloodInUrine: sql.NullBool{Bool: c.FormValue("blood_in_urine") == "Yes", Valid: c.FormValue("blood_in_urine") != ""},
		DateBloodInUrine: func() sql.NullTime {
			val := c.FormValue("date_blood_in_urine")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationBloodInUrine: func() sql.NullInt32 {
			val := c.FormValue("duration_blood_in_urine")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		OtherHemorrhagicSymptoms: sql.NullBool{Bool: c.FormValue("other_hemorrhagic_symptoms") == "Yes", Valid: c.FormValue("other_hemorrhagic_symptoms") != ""},
		DateOtherHemorrhagicSymptoms: func() sql.NullTime {
			val := c.FormValue("date_other_hemorrhagic_symptoms")
			t, err := time.Parse("2006-01-02", val)
			return sql.NullTime{Time: t, Valid: err == nil && val != ""}
		}(),
		DurationOtherHemorrhagicSymptoms: func() sql.NullInt32 {
			val := c.FormValue("duration_other_hemorrhagic_symptoms")
			i, err := strconv.ParseInt(val, 10, 32)
			return sql.NullInt32{Int32: int32(i), Valid: err == nil && val != ""}
		}(),
		CreatedAt: time.Now(),
	}

	if err := models.SaveVHFClinicalSigns(db, clinicalSigns); err != nil {
		sl.Error("Failed to save clinical signs", "error", err)
		return c.Status(500).SendString("Failed to save clinical signs")
	}

	// Parse and save hospitalization data
	hospitalization := &models.VHFHospitalization{
		PatientID:          patient.ID,
		Hospitalized:       parseBool(c.FormValue("hospitalized")),
		AdmissionDate:      parseNullTime(c.FormValue("admission_date")),
		HealthFacilityName: c.FormValue("health_facility_name"),
		InIsolation:        parseBool(c.FormValue("in_isolation")),
		IsolationDate:      parseNullTime(c.FormValue("isolation_date")),
		CreatedAt:          time.Now(),
	}

	if err := models.SaveVHFHospitalization(db, hospitalization); err != nil {
		sl.Error("Failed to save hospitalization data", "error", err)
		return c.Status(500).SendString("Failed to save hospitalization data")
	}

	// Parse and save risk factors
	riskFactors := &models.VHFRiskFactors{
		PatientID:       patient.ID,
		ContactWithCase: sql.NullBool{Bool: parseBool(c.FormValue("contactWithCase")), Valid: true},
		ContactName:     c.FormValue("contact_name"),
		ContactRelation: c.FormValue("contact_relation"),
		ContactDates:    c.FormValue("contact_dates"),
		ContactVillage:  c.FormValue("contact_village"),
		ContactDistrict: c.FormValue("contact_district"),
		ContactStatus:   c.FormValue("contact_status"),
		ContactTypes:    c.FormValue("contact_types"),
		CreatedAt:       time.Now(),
	}

	// Parse contact death date if provided
	if deathDate := c.FormValue("contact_death_date"); deathDate != "" {
		if t, err := time.Parse("2006-01-02", deathDate); err == nil {
			riskFactors.ContactDeathDate = sql.NullTime{Time: t, Valid: true}
		}
	}

	if err := models.SaveVHFRiskFactors(db, riskFactors); err != nil {
		sl.Error("Failed to save risk factors", "error", err)
		return c.Status(500).SendString("Failed to save risk factors")
	}

	// Parse and save investigator data
	investigator := &models.VHFInvestigator{
		PatientID:         patient.ID,
		InvestigatorName:  c.FormValue("investigator_name"),
		Phone:             c.FormValue("phone"),
		Email:             c.FormValue("email"),
		Position:          c.FormValue("position"),
		District:          c.FormValue("district"),
		HealthFacility:    c.FormValue("health_facility"),
		InformationSource: c.FormValue("information_source"),
		ProxyName:         c.FormValue("proxy_name"),
		ProxyRelation:     c.FormValue("proxy_relation"),
		CreatedAt:         time.Now(),
	}

	if err := models.SaveVHFInvestigator(db, investigator); err != nil {
		sl.Error("Failed to save investigator data", "error", err)
		return c.Status(500).SendString("Failed to save investigator data")
	}

	// Send SMS notification if phone number is provided
	if patient.DataCapturerPhone.String != "" {
		message := fmt.Sprintf("You have notified a suspected VHF Case %s and %s. Case DetailsPatient: %s %s",
			patient.CaseCode.String, patient.ReportingHealthFacilityName.String, patient.Surname, patient.OtherNames)
		if err := smsService.SendSMS(patient.DataCapturerPhone.String, message); err != nil {
			sl.Error("Failed to send SMS notification", "error", err)
			// Don't return error here, as the form was still saved successfully
		}
	}

	// Send SMS notification to DSFP if district is provided
	if patient.District != "" {
		// Get district ID from district name
		district, err := models.GetDistrictByName(db, patient.District)
		if err != nil {
			sl.Error("Failed to get district by name", "district", patient.District, "error", err)
		} else {
			// Get DSFP for the district
			focalPerson, err := models.GetFocalPersonByDistrict(db, district.ID)
			if err != nil {
				sl.Error("Failed to get focal person for district", "district_id", district.ID, "error", err)
			} else if focalPerson != nil && focalPerson.Phone != "" {
				dsfpMessage := fmt.Sprintf("A suspected VHF Case %s has been notified at %s in %s district. Patient: %s %s and a sample has been dispatched to CPHL. Please track the sample and results.",
					patient.CaseCode.String, patient.ReportingHealthFacilityName.String, patient.District, patient.Surname, patient.OtherNames)
				if err := smsService.SendSMS(focalPerson.Phone, dsfpMessage); err != nil {
					sl.Error("Failed to send SMS notification to DSFP", "error", err)
					// Don't return error here, as the form was still saved successfully
				} else {
					sl.Info("SMS sent to DSFP", "district", patient.District, "dsfp_phone", focalPerson.Phone)
				}
			}
		}
	}

	// Redirect to success page with case code
	return c.Redirect(fmt.Sprintf("/vhf-cif/success?case_code=%s", patient.CaseCode.String))
}

// Helper functions for parsing form values

func parseInt32(str string) int32 {
	if str == "" {
		return 0
	}
	i, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return 0
	}
	return int32(i)
}

func parseFloat(str string) float64 {
	if str == "" {
		return 0
	}
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0
	}
	return f
}

func parseBool(str string) bool {
	return str == "Yes"
}

// HandlerVHFClinicalSignsSubmit handles the submission of clinical signs
func HandlerVHFClinicalSignsSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	patientID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid patient ID")
	}

	// Parse temperature
	var temperature sql.NullFloat64
	if tempStr := c.FormValue("temperature"); tempStr != "" {
		if temp, err := strconv.ParseFloat(tempStr, 64); err == nil {
			temperature.Float64 = temp
			temperature.Valid = true
		}
	}

	// Parse durations
	parseDuration := func(durationStr string) sql.NullInt32 {
		var result sql.NullInt32
		if durationStr != "" {
			if d, err := strconv.ParseInt(durationStr, 10, 32); err == nil {
				result.Int32 = int32(d)
				result.Valid = true
			}
		}
		return result
	}

	// Create clinical signs struct with proper null handling
	signs := &models.VHFClinicalSigns{
		PatientID:        patientID,
		DateInitialOnset: parseNullTime(c.FormValue("date_initial_onset")),
		TempSource: sql.NullString{
			String: c.FormValue("temp_source"),
			Valid:  c.FormValue("temp_source") != "",
		},
		Temperature: temperature,
		Fever: sql.NullBool{
			Bool:  c.FormValue("fever") == "true",
			Valid: c.FormValue("fever") != "",
		},
		DateFever:     parseNullTime(c.FormValue("date_fever")),
		DurationFever: parseDuration(c.FormValue("duration_fever")),
		Vomiting: sql.NullBool{
			Bool:  c.FormValue("vomiting") == "true",
			Valid: c.FormValue("vomiting") != "",
		},
		DateVomiting:     parseNullTime(c.FormValue("date_vomiting")),
		DurationVomiting: parseDuration(c.FormValue("duration_vomiting")),
		Nausea: sql.NullBool{
			Bool:  c.FormValue("nausea") == "true",
			Valid: c.FormValue("nausea") != "",
		},
		DateNausea:     parseNullTime(c.FormValue("date_nausea")),
		DurationNausea: parseDuration(c.FormValue("duration_nausea")),
		Diarrhea: sql.NullBool{
			Bool:  c.FormValue("diarrhea") == "true",
			Valid: c.FormValue("diarrhea") != "",
		},
		DateDiarrhea:     parseNullTime(c.FormValue("date_diarrhea")),
		DurationDiarrhea: parseDuration(c.FormValue("duration_diarrhea")),
		IntenseFatigueGeneralWeakness: sql.NullBool{
			Bool:  c.FormValue("intense_fatigue_general_weakness") == "true",
			Valid: c.FormValue("intense_fatigue_general_weakness") != "",
		},
		DateIntenseFatigueGeneralWeakness:     parseNullTime(c.FormValue("date_intense_fatigue_general_weakness")),
		DurationIntenseFatigueGeneralWeakness: parseDuration(c.FormValue("duration_intense_fatigue_general_weakness")),
		EpigastricPain: sql.NullBool{
			Bool:  c.FormValue("epigastric_pain") == "true",
			Valid: c.FormValue("epigastric_pain") != "",
		},
		DateEpigastricPain:     parseNullTime(c.FormValue("date_epigastric_pain")),
		DurationEpigastricPain: parseDuration(c.FormValue("duration_epigastric_pain")),
		LowerAbdominalPain: sql.NullBool{
			Bool:  c.FormValue("lower_abdominal_pain") == "true",
			Valid: c.FormValue("lower_abdominal_pain") != "",
		},
		DateLowerAbdominalPain:     parseNullTime(c.FormValue("date_lower_abdominal_pain")),
		DurationLowerAbdominalPain: parseDuration(c.FormValue("duration_lower_abdominal_pain")),
		ChestPain: sql.NullBool{
			Bool:  c.FormValue("chest_pain") == "true",
			Valid: c.FormValue("chest_pain") != "",
		},
		DateChestPain:     parseNullTime(c.FormValue("date_chest_pain")),
		DurationChestPain: parseDuration(c.FormValue("duration_chest_pain")),
		MusclePain: sql.NullBool{
			Bool:  c.FormValue("muscle_pain") == "true",
			Valid: c.FormValue("muscle_pain") != "",
		},
		DateMusclePain:     parseNullTime(c.FormValue("date_muscle_pain")),
		DurationMusclePain: parseDuration(c.FormValue("duration_muscle_pain")),
		JointPain: sql.NullBool{
			Bool:  c.FormValue("joint_pain") == "true",
			Valid: c.FormValue("joint_pain") != "",
		},
		DateJointPain:     parseNullTime(c.FormValue("date_joint_pain")),
		DurationJointPain: parseDuration(c.FormValue("duration_joint_pain")),
		Headache: sql.NullBool{
			Bool:  c.FormValue("headache") == "true",
			Valid: c.FormValue("headache") != "",
		},
		DateHeadache:     parseNullTime(c.FormValue("date_headache")),
		DurationHeadache: parseDuration(c.FormValue("duration_headache")),
		Cough: sql.NullBool{
			Bool:  c.FormValue("cough") == "true",
			Valid: c.FormValue("cough") != "",
		},
		DateCough:     parseNullTime(c.FormValue("date_cough")),
		DurationCough: parseDuration(c.FormValue("duration_cough")),
		DifficultyBreathing: sql.NullBool{
			Bool:  c.FormValue("difficulty_breathing") == "true",
			Valid: c.FormValue("difficulty_breathing") != "",
		},
		DateDifficultyBreathing:     parseNullTime(c.FormValue("date_difficulty_breathing")),
		DurationDifficultyBreathing: parseDuration(c.FormValue("duration_difficulty_breathing")),
		DifficultySwallowing: sql.NullBool{
			Bool:  c.FormValue("difficulty_swallowing") == "true",
			Valid: c.FormValue("difficulty_swallowing") != "",
		},
		DateDifficultySwallowing:     parseNullTime(c.FormValue("date_difficulty_swallowing")),
		DurationDifficultySwallowing: parseDuration(c.FormValue("duration_difficulty_swallowing")),
		SoreThroat: sql.NullBool{
			Bool:  c.FormValue("sore_throat") == "true",
			Valid: c.FormValue("sore_throat") != "",
		},
		DateSoreThroat:     parseNullTime(c.FormValue("date_sore_throat")),
		DurationSoreThroat: parseDuration(c.FormValue("duration_sore_throat")),
		Jaundice: sql.NullBool{
			Bool:  c.FormValue("jaundice") == "true",
			Valid: c.FormValue("jaundice") != "",
		},
		DateJaundice:     parseNullTime(c.FormValue("date_jaundice")),
		DurationJaundice: parseDuration(c.FormValue("duration_jaundice")),
		Conjunctivitis: sql.NullBool{
			Bool:  c.FormValue("conjunctivitis") == "true",
			Valid: c.FormValue("conjunctivitis") != "",
		},
		DateConjunctivitis:     parseNullTime(c.FormValue("date_conjunctivitis")),
		DurationConjunctivitis: parseDuration(c.FormValue("duration_conjunctivitis")),
		SkinRash: sql.NullBool{
			Bool:  c.FormValue("skin_rash") == "true",
			Valid: c.FormValue("skin_rash") != "",
		},
		DateSkinRash:     parseNullTime(c.FormValue("date_skin_rash")),
		DurationSkinRash: parseDuration(c.FormValue("duration_skin_rash")),
		Hiccups: sql.NullBool{
			Bool:  c.FormValue("hiccups") == "true",
			Valid: c.FormValue("hiccups") != "",
		},
		DateHiccups:     parseNullTime(c.FormValue("date_hiccups")),
		DurationHiccups: parseDuration(c.FormValue("duration_hiccups")),
		PainBehindEyes: sql.NullBool{
			Bool:  c.FormValue("pain_behind_eyes") == "true",
			Valid: c.FormValue("pain_behind_eyes") != "",
		},
		DatePainBehindEyes:     parseNullTime(c.FormValue("date_pain_behind_eyes")),
		DurationPainBehindEyes: parseDuration(c.FormValue("duration_pain_behind_eyes")),
		SensitiveToLight: sql.NullBool{
			Bool:  c.FormValue("sensitive_to_light") == "true",
			Valid: c.FormValue("sensitive_to_light") != "",
		},
		DateSensitiveToLight:     parseNullTime(c.FormValue("date_sensitive_to_light")),
		DurationSensitiveToLight: parseDuration(c.FormValue("duration_sensitive_to_light")),
		ComaUnconscious: sql.NullBool{
			Bool:  c.FormValue("coma_unconscious") == "true",
			Valid: c.FormValue("coma_unconscious") != "",
		},
		DateComaUnconscious:     parseNullTime(c.FormValue("date_coma_unconscious")),
		DurationComaUnconscious: parseDuration(c.FormValue("duration_coma_unconscious")),
		ConfusedOrDisoriented: sql.NullBool{
			Bool:  c.FormValue("confused_or_disoriented") == "true",
			Valid: c.FormValue("confused_or_disoriented") != "",
		},
		DateConfusedOrDisoriented:     parseNullTime(c.FormValue("date_confused_or_disoriented")),
		DurationConfusedOrDisoriented: parseDuration(c.FormValue("duration_confused_or_disoriented")),
		Convulsions: sql.NullBool{
			Bool:  c.FormValue("convulsions") == "true",
			Valid: c.FormValue("convulsions") != "",
		},
		DateConvulsions:     parseNullTime(c.FormValue("date_convulsions")),
		DurationConvulsions: parseDuration(c.FormValue("duration_convulsions")),
		UnexplainedBleeding: sql.NullBool{
			Bool:  c.FormValue("unexplained_bleeding") == "true",
			Valid: c.FormValue("unexplained_bleeding") != "",
		},
		DateUnexplainedBleeding:     parseNullTime(c.FormValue("date_unexplained_bleeding")),
		DurationUnexplainedBleeding: parseDuration(c.FormValue("duration_unexplained_bleeding")),
		BleedingOfTheGums: sql.NullBool{
			Bool:  c.FormValue("bleeding_of_the_gums") == "true",
			Valid: c.FormValue("bleeding_of_the_gums") != "",
		},
		DateBleedingOfTheGums:     parseNullTime(c.FormValue("date_bleeding_of_the_gums")),
		DurationBleedingOfTheGums: parseDuration(c.FormValue("duration_bleeding_of_the_gums")),
		BleedingFromInjectionSite: sql.NullBool{
			Bool:  c.FormValue("bleeding_from_injection_site") == "true",
			Valid: c.FormValue("bleeding_from_injection_site") != "",
		},
		DateBleedingFromInjectionSite:     parseNullTime(c.FormValue("date_bleeding_from_injection_site")),
		DurationBleedingFromInjectionSite: parseDuration(c.FormValue("duration_bleeding_from_injection_site")),
		NoseBleedEpistaxis: sql.NullBool{
			Bool:  c.FormValue("nose_bleed_epistaxis") == "true",
			Valid: c.FormValue("nose_bleed_epistaxis") != "",
		},
		DateNoseBleedEpistaxis:     parseNullTime(c.FormValue("date_nose_bleed_epistaxis")),
		DurationNoseBleedEpistaxis: parseDuration(c.FormValue("duration_nose_bleed_epistaxis")),
		BloodyStool: sql.NullBool{
			Bool:  c.FormValue("bloody_stool") == "true",
			Valid: c.FormValue("bloody_stool") != "",
		},
		DateBloodyStool:     parseNullTime(c.FormValue("date_bloody_stool")),
		DurationBloodyStool: parseDuration(c.FormValue("duration_bloody_stool")),
		BloodInVomit: sql.NullBool{
			Bool:  c.FormValue("blood_in_vomit") == "true",
			Valid: c.FormValue("blood_in_vomit") != "",
		},
		DateBloodInVomit:     parseNullTime(c.FormValue("date_blood_in_vomit")),
		DurationBloodInVomit: parseDuration(c.FormValue("duration_blood_in_vomit")),
		CoughingUpBloodHemoptysis: sql.NullBool{
			Bool:  c.FormValue("coughing_up_blood_hemoptysis") == "true",
			Valid: c.FormValue("coughing_up_blood_hemoptysis") != "",
		},
		DateCoughingUpBloodHemoptysis:     parseNullTime(c.FormValue("date_coughing_up_blood_hemoptysis")),
		DurationCoughingUpBloodHemoptysis: parseDuration(c.FormValue("duration_coughing_up_blood_hemoptysis")),
		BleedingFromVagina: sql.NullBool{
			Bool:  c.FormValue("bleeding_from_vagina") == "true",
			Valid: c.FormValue("bleeding_from_vagina") != "",
		},
		DateBleedingFromVagina:     parseNullTime(c.FormValue("date_bleeding_from_vagina")),
		DurationBleedingFromVagina: parseDuration(c.FormValue("duration_bleeding_from_vagina")),
		BruisingOfTheSkin: sql.NullBool{
			Bool:  c.FormValue("bruising_of_the_skin") == "true",
			Valid: c.FormValue("bruising_of_the_skin") != "",
		},
		DateBruisingOfTheSkin:     parseNullTime(c.FormValue("date_bruising_of_the_skin")),
		DurationBruisingOfTheSkin: parseDuration(c.FormValue("duration_bruising_of_the_skin")),
		BloodInUrine: sql.NullBool{
			Bool:  c.FormValue("blood_in_urine") == "true",
			Valid: c.FormValue("blood_in_urine") != "",
		},
		DateBloodInUrine:     parseNullTime(c.FormValue("date_blood_in_urine")),
		DurationBloodInUrine: parseDuration(c.FormValue("duration_blood_in_urine")),
		OtherHemorrhagicSymptoms: sql.NullBool{
			Bool:  c.FormValue("other_hemorrhagic_symptoms") == "true",
			Valid: c.FormValue("other_hemorrhagic_symptoms") != "",
		},
		DateOtherHemorrhagicSymptoms:     parseNullTime(c.FormValue("date_other_hemorrhagic_symptoms")),
		DurationOtherHemorrhagicSymptoms: parseDuration(c.FormValue("duration_other_hemorrhagic_symptoms")),
		CreatedAt:                        time.Now(),
	}

	// Save clinical signs
	if err := models.SaveVHFClinicalSigns(db, signs); err != nil {
		sl.Error("Failed to save clinical signs", "error", err)
		return c.Status(500).SendString("Failed to save clinical signs")
	}

	// Store patient ID in session
	sess, err := store.Get(c)
	if err != nil {
		sl.Error("Failed to get session", "error", err)
		return c.Status(500).SendString("Failed to get session")
	}
	sess.Set("patient_id", patientID)
	if err := sess.Save(); err != nil {
		sl.Error("Failed to save session", "error", err)
		return c.Status(500).SendString("Failed to save session")
	}

	// Redirect to hospitalization form
	return c.Redirect(fmt.Sprintf("/vhf-cif/hospitalization/%d", patientID))
}

// HandlerVHFHospitalizationSubmit handles the submission of hospitalization information
func HandlerVHFHospitalizationSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	patientID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid patient ID")
	}

	hospitalization := &models.VHFHospitalization{
		PatientID:          patientID,
		Hospitalized:       c.FormValue("hospitalized") == "Yes",
		HealthFacilityName: c.FormValue("health_facility_name"),
		InIsolation:        c.FormValue("isolation") == "Yes",
	}

	// Parse date fields
	if admission := c.FormValue("admission_date"); admission != "" {
		if t, err := time.Parse("2006-01-02", admission); err == nil {
			hospitalization.AdmissionDate = sql.NullTime{Time: t, Valid: true}
		}
	}

	if isolation := c.FormValue("isolation_date"); isolation != "" {
		if t, err := time.Parse("2006-01-02", isolation); err == nil {
			hospitalization.IsolationDate = sql.NullTime{Time: t, Valid: true}
		}
	}

	// Save hospitalization data
	if err := models.SaveVHFHospitalization(db, hospitalization); err != nil {
		sl.Error("Failed to save hospitalization data", "error", err)
		return c.Status(500).SendString("Failed to save hospitalization data")
	}

	// Store patient ID in session
	sess, err := store.Get(c)
	if err != nil {
		sl.Error("Failed to get session", "error", err)
		return c.Status(500).SendString("Failed to get session")
	}
	sess.Set("patient_id", patientID)
	if err := sess.Save(); err != nil {
		sl.Error("Failed to save session", "error", err)
		return c.Status(500).SendString("Failed to save session")
	}

	// Redirect to risk factors form
	return c.Redirect(fmt.Sprintf("/vhf-cif/risk-factors/%d", patientID))
}

// HandlerVHFRiskFactorsSubmit handles the submission of risk factors
func HandlerVHFRiskFactorsSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	patientID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid patient ID")
	}

	riskFactors := &models.VHFRiskFactors{
		PatientID:       patientID,
		ContactWithCase: sql.NullBool{Bool: parseBool(c.FormValue("contactWithCase")), Valid: true},
		ContactName:     c.FormValue("contact_name"),
		ContactRelation: c.FormValue("contact_relation"),
		ContactDates:    c.FormValue("contact_dates"),
		ContactVillage:  c.FormValue("contact_village"),
		ContactDistrict: c.FormValue("contact_district"),
		ContactStatus:   c.FormValue("contact_status"),
		ContactTypes:    c.FormValue("contact_types"),
		CreatedAt:       time.Now(),
	}

	// Parse contact death date if provided
	if deathDate := c.FormValue("contact_death_date"); deathDate != "" {
		if t, err := time.Parse("2006-01-02", deathDate); err == nil {
			riskFactors.ContactDeathDate = sql.NullTime{Time: t, Valid: true}
		}
	}

	// Save risk factors data
	if err := models.SaveVHFRiskFactors(db, riskFactors); err != nil {
		sl.Error("Failed to save risk factors", "error", err)
		return c.Status(500).SendString("Failed to save risk factors")
	}

	// Store patient ID in session
	sess, err := store.Get(c)
	if err != nil {
		sl.Error("Failed to get session", "error", err)
		return c.Status(500).SendString("Failed to get session")
	}
	sess.Set("patient_id", patientID)
	if err := sess.Save(); err != nil {
		sl.Error("Failed to save session", "error", err)
		return c.Status(500).SendString("Failed to save session")
	}

	// Redirect to laboratory form
	return c.Redirect(fmt.Sprintf("/vhf-cif/laboratory/%d", patientID))
}

//Last known save lab data function that does not send sms
// HandlerVHFLaboratorySubmit handles the submission of laboratory information
// func HandlerVHFLaboratorySubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config, smsService *services.SMSService) error {
// 	patientID, err := strconv.ParseInt(c.Params("id"), 10, 64)
// 	if err != nil {
// 		return c.Status(400).SendString("Invalid patient ID")
// 	}

// 	laboratory := &models.VHFLaboratory{
// 		PatientID:            patientID,
// 		SampleCollectionDate: parseNullTime(c.FormValue("sample_collection_date")),
// 		SampleCollectionTime: sql.NullString{String: c.FormValue("sample_collection_time"), Valid: c.FormValue("sample_collection_time") != ""},
// 		SampleType:           sql.NullString{String: c.FormValue("sample_type"), Valid: c.FormValue("sample_type") != ""},
// 		OtherSampleType:      sql.NullString{String: c.FormValue("other_sample_type"), Valid: c.FormValue("other_sample_type") != ""},
// 		RequestedTest:        sql.NullString{String: c.FormValue("requested_test"), Valid: c.FormValue("requested_test") != ""},
// 		Serology:             sql.NullString{String: c.FormValue("serology"), Valid: c.FormValue("serology") != ""},
// 		MalariaRDT:           sql.NullString{String: c.FormValue("malaria_rdt"), Valid: c.FormValue("malaria_rdt") != ""},
// 		HIVRDT:               sql.NullString{String: c.FormValue("hiv_rdt"), Valid: c.FormValue("hiv_rdt") != ""},
// 		CreatedAt:            time.Now(),
// 	}

// 	if err := models.SaveVHFLaboratory(db, laboratory); err != nil {
// 		sl.Error("Failed to save laboratory data", "error", err)
// 		return c.Status(500).SendString("Failed to save laboratory data")
// 	}
// 	// Send SMS notification to CPHL if phone number is provided
// 	if laboratory.SampleType.String != "" {
// 		// Get patient details first
// 		patient, err := models.GetVHFPatient(db, patientID)
// 		if err != nil {
// 			sl.Error("Failed to get patient details for SMS", "error", err)
// 			return c.Status(500).SendString("Failed to get patient details")
// 		}
// 		var labPhoneNumber string = "256783261162"
// 		message := fmt.Sprintf("A suspected VHF Case %s has been notified at %s and sample has been dispatched to CPHL with Case Details: %s %s",
// 			labPhoneNumber, patient.CaseCode.String, patient.ReportingHealthFacilityName.String, patient.Surname, patient.OtherNames)
// 		// Send SMS notification

// 		if err := smsService.SendSMS(labPhoneNumber, message); err != nil {
// 			sl.Error("Failed to send SMS notification", "error", err)
// 		}
// 	}
// 	return c.Redirect(fmt.Sprintf("/vhf/view/%d", patientID))
// }

// Refactored lab submission function that sends sms
func HandlerVHFLaboratorySubmit(
	c *fiber.Ctx,
	db *sql.DB,
	sl *slog.Logger,
	store *session.Store,
	config config.Config,
	smsService *services.SMSService,
) error {
	patientID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid patient ID")
	}

	laboratory := &models.VHFLaboratory{
		PatientID:            patientID,
		SampleCollectionDate: parseNullTime(c.FormValue("sample_collection_date")),
		SampleCollectionTime: sql.NullString{String: c.FormValue("sample_collection_time"), Valid: c.FormValue("sample_collection_time") != ""},
		SampleType:           sql.NullString{String: c.FormValue("sample_type"), Valid: c.FormValue("sample_type") != ""},
		OtherSampleType:      sql.NullString{String: c.FormValue("other_sample_type"), Valid: c.FormValue("other_sample_type") != ""},
		RequestedTest:        sql.NullString{String: c.FormValue("requested_test"), Valid: c.FormValue("requested_test") != ""},
		Serology:             sql.NullString{String: c.FormValue("serology"), Valid: c.FormValue("serology") != ""},
		MalariaRDT:           sql.NullString{String: c.FormValue("malaria_rdt"), Valid: c.FormValue("malaria_rdt") != ""},
		HIVRDT:               sql.NullString{String: c.FormValue("hiv_rdt"), Valid: c.FormValue("hiv_rdt") != ""},
		CreatedAt:            time.Now(),
	}

	// Save the lab record
	if err := models.SaveVHFLaboratory(db, laboratory); err != nil {
		sl.Error("Failed to save laboratory data", "error", err)
		return c.Status(500).SendString("Failed to save laboratory data")
	}
	sl.Info("Lab record saved successfully", "patientID", patientID)

	// Fetch patient details for SMS
	patient, err := models.GetVHFPatient(db, patientID)
	if err != nil {
		sl.Error("Failed to get patient details for SMS", "error", err)
		return c.Status(500).SendString("Failed to get patient details")
	}

	// Prepare SMS
	labPhoneNumber := "0783261162"
	message := fmt.Sprintf(
		"A suspected VHF Case %s has been notified at %s. Sample has been dispatched to CPHL. Case details: %s %s",
		patient.CaseCode.String,
		patient.ReportingHealthFacilityName.String,
		patient.Surname,
		patient.OtherNames,
	)

	// Log before sending
	sl.Info("Attempting to send SMS", "to", labPhoneNumber, "message", message)

	// Send SMS
	smsService.SendSMS("256783261162", "Test message from handler")
	if err := smsService.SendSMS(labPhoneNumber, message); err != nil {
		sl.Error("Failed to send SMS notification", "error", err)
	} else {
		sl.Info("SMS sent successfully", "to", labPhoneNumber)
	}

	return c.Redirect(fmt.Sprintf("/vhf/view/%d", patientID))
}

// HandlerVHFInvestigatorSubmit handles the submission of investigator information
func HandlerVHFInvestigatorSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	patientID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid patient ID")
	}

	investigator := &models.VHFInvestigator{
		PatientID:         patientID,
		InvestigatorName:  c.FormValue("investigator_name"),
		Phone:             c.FormValue("phone"),
		Email:             c.FormValue("email"),
		Position:          c.FormValue("position"),
		District:          c.FormValue("district"),
		HealthFacility:    c.FormValue("health_facility"),
		InformationSource: c.FormValue("information_source"),
		ProxyName:         c.FormValue("proxy_name"),
		ProxyRelation:     c.FormValue("proxy_relation"),
		CreatedAt:         time.Now(),
	}

	if err := models.SaveVHFInvestigator(db, investigator); err != nil {
		sl.Error("Failed to save investigator information", "error", err)
		return c.Status(500).SendString("Failed to save investigator information")
	}

	// Get the case code from the patient record
	patient, err := models.GetVHFPatient(db, patientID)
	if err != nil {
		sl.Error("Failed to get patient", "error", err)
		return c.Status(500).SendString("Failed to get patient information")
	}

	// Redirect to success page with case code
	return c.Redirect(fmt.Sprintf("/vhf-cif/success?case_code=%s", patient.CaseCode.String))
}

// HandlerVHFList handles the listing of all VHF cases
func HandlerVHFList(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	// Get current user information
	userID, _ := GetUser(c, sl, store)

	// Check if user has role ID 65 (vhf_lab_technician) and get their facility
	var facilityFilter string
	var args []interface{}

	// Admins should not be restricted by facility
	isAdmin := security.HasAnyRole(db, userID, []string{"admin", "super_admin"})

	// Check if user has vhf_lab_technician role (ID 65)
	roleQuery := `
		SELECT COUNT(*) FROM user_roles ur 
		JOIN roles r ON ur.role_id = r.id 
		WHERE ur.user_id = $1 AND r.id = 65 AND r.is_active = true
	`
	var roleCount int
	err := db.QueryRowContext(c.Context(), roleQuery, userID).Scan(&roleCount)
	if err != nil {
		sl.Error("Failed to check user role", "error", err)
		return c.Status(500).SendString("Failed to check user permissions")
	}

	// Check if user has any VHF-related roles or regional/district assignments
	hasVHFRole := roleCount > 0

	// Also check if user has regional/district assignments (even without role 65)
	var hasRegionalAssignment bool
	regionalQuery := `
		SELECT COUNT(*) FROM employee e
		JOIN users u ON e.employee_id = u.user_employee
		WHERE u.user_id = $1 AND (
			(e.afi_region IS NOT NULL AND e.afi_region != '') OR 
			(e.afi_district IS NOT NULL AND e.afi_district != '') OR
			(e.afi_facility IS NOT NULL AND e.afi_facility != '')
		)
	`
	var regionalCount int
	err = db.QueryRowContext(c.Context(), regionalQuery, userID).Scan(&regionalCount)
	if err != nil {
		sl.Error("Failed to check user regional assignments", "error", err)
	}
	hasRegionalAssignment = regionalCount > 0

	// Apply filtering if user has VHF role OR regional/district assignments (unless admin)
	if (hasVHFRole || hasRegionalAssignment) && !isAdmin {
		// Get user's AFI facility, region, and district directly from employee table
		afiQuery := `
			SELECT e.afi_facility, e.afi_region, e.afi_district
			FROM employee e
			JOIN users u ON e.employee_id = u.user_employee
			WHERE u.user_id = $1
			LIMIT 1
		`
		var userAFIFacility, userAFIRegion, userAFIDistrict sql.NullString
		err := db.QueryRowContext(c.Context(), afiQuery, userID).Scan(&userAFIFacility, &userAFIRegion, &userAFIDistrict)
		if err != nil && err != sql.ErrNoRows {
			sl.Error("Failed to get user AFI facility/region/district", "error", err)
			return c.Status(500).SendString("Failed to get user facility information")
		}

		// Build filter based on what's available
		filterConditions := []string{}
		paramIndex := 1

		if userAFIFacility.Valid && userAFIFacility.String != "" {
			// First priority: exact facility match
			filterConditions = append(filterConditions, fmt.Sprintf("LOWER(TRIM(vc.reporting_health_facility_name)) = LOWER(TRIM($%d))", paramIndex))
			args = append(args, userAFIFacility.String)
			paramIndex++
			sl.Info("Filtering VHF cases by AFI facility name", "user_id", userID, "afi_facility", userAFIFacility.String)
		}

		if userAFIDistrict.Valid && userAFIDistrict.String != "" {
			// Second priority: same district
			filterConditions = append(filterConditions, fmt.Sprintf("LOWER(TRIM(vc.district)) = LOWER(TRIM($%d))", paramIndex))
			args = append(args, userAFIDistrict.String)
			paramIndex++
			sl.Info("Filtering VHF cases by AFI district", "user_id", userID, "afi_district", userAFIDistrict.String)
		}

		if userAFIRegion.Valid && userAFIRegion.String != "" {
			// Third priority: same region - simplified logic
			filterConditions = append(filterConditions, fmt.Sprintf("EXISTS (SELECT 1 FROM afi_facilities af WHERE LOWER(TRIM(af.facility_name)) = LOWER(TRIM(vc.reporting_health_facility_name)) AND LOWER(TRIM(af.region)) = LOWER(TRIM($%d)))", paramIndex))
			args = append(args, userAFIRegion.String)
			paramIndex++
			sl.Info("Filtering VHF cases by AFI region", "user_id", userID, "afi_region", userAFIRegion.String)
		}

		if len(filterConditions) > 0 {
			facilityFilter = "AND (" + strings.Join(filterConditions, " OR ") + ")"
			sl.Info("Applied VHF filtering", "user_id", userID, "conditions", len(filterConditions), "filter", facilityFilter)
		} else {
			// Strict requirement: no AFI facility/region/district set => no records
			facilityFilter = "AND 1 = 0"
			sl.Info("User without AFI facility/region/district; no VHF records will be shown", "user_id", userID)
		}
	} else {
		if isAdmin {
			sl.Info("Admin user - bypassing facility filter for VHF list", "user_id", userID)
		} else {
			sl.Info("User without VHF role or regional assignments - showing all VHF cases", "user_id", userID)
		}
	}

	// Build the query with optional facility filtering
	query := `
		SELECT 
			vc.id,
			vc.case_code,
			vc.surname,
			vc.other_names,
			vc.date_of_birth,
			vc.age_years,
			vc.age_months,
			vc.gender,
			vc.district,
			vc.status,
			vc.created_at,
			CASE WHEN vl.id IS NOT NULL THEN true ELSE false END as lab_status
		FROM vhf_patients vc
		LEFT JOIN vhf_laboratory vl ON vc.id = vl.patient_id
		WHERE 1=1 ` + facilityFilter + `
		ORDER BY vc.created_at DESC`

	rows, err := db.QueryContext(c.Context(), query, args...)
	if err != nil {
		sl.Error("Failed to query VHF cases", "error", err)
		return c.Status(500).SendString("Failed to retrieve VHF cases")
	}
	defer rows.Close()

	var cases []fiber.Map
	for rows.Next() {
		var (
			id         int64
			caseCode   sql.NullString
			surname    string
			otherNames sql.NullString
			dob        sql.NullTime
			ageYears   sql.NullInt32
			ageMonths  sql.NullInt32
			gender     string
			district   string
			status     string
			createdAt  time.Time
			labStatus  bool
		)

		err := rows.Scan(
			&id,
			&caseCode,
			&surname,
			&otherNames,
			&dob,
			&ageYears,
			&ageMonths,
			&gender,
			&district,
			&status,
			&createdAt,
			&labStatus,
		)
		if err != nil {
			sl.Error("Failed to scan VHF case", "error", err)
			continue
		}

		// Calculate age display
		var ageDisplay string
		if ageYears.Valid {
			ageDisplay = fmt.Sprintf("%d years", ageYears.Int32)
		} else if ageMonths.Valid {
			ageDisplay = fmt.Sprintf("%d months", ageMonths.Int32)
		}

		// Format name
		name := surname
		if otherNames.Valid && otherNames.String != "" {
			name = fmt.Sprintf("%s %s", surname, otherNames.String)
		}

		cases = append(cases, fiber.Map{
			"ID":        id,
			"CaseCode":  caseCode.String,
			"Name":      name,
			"Age":       ageDisplay,
			"Gender":    gender,
			"District":  district,
			"Status":    status,
			"LabStatus": labStatus,
			"CreatedAt": createdAt.Format("2006-01-02 15:04"),
		})
	}

	if err = rows.Err(); err != nil {
		sl.Error("Error iterating VHF cases", "error", err)
		return c.Status(500).SendString("Error retrieving VHF cases")
	}

	// Check if this is an API request
	if c.Get("Accept") == "application/json" {
		return c.JSON(cases)
	}

	// Load measles cases as well
	measlesQuery := `
		SELECT 
			mp.patient_id,
			mp.measles_code,
			mp.patient_name,
			mp.sex,
			mp.dob,
			md.age_months,
			md.onset_district,
			mp.created_at,
			CASE WHEN mr.id IS NOT NULL THEN true ELSE false END as results_status
		FROM measles_patients mp
		LEFT JOIN measles_demographics md ON mp.patient_id = md.patient_id
		LEFT JOIN measles_results mr ON mp.patient_id = mr.patient_id
		ORDER BY mp.created_at DESC`

	measlesRows, err := db.QueryContext(c.Context(), measlesQuery)
	if err != nil {
		sl.Error("Failed to query measles cases", "error", err)
		// Continue with VHF cases only if measles query fails
	}

	var measlesCases []fiber.Map
	if measlesRows != nil {
		defer measlesRows.Close()
		for measlesRows.Next() {
			var (
				patientID     string
				measlesCode   sql.NullString
				patientName   string
				sex           string
				dob           sql.NullTime
				ageMonths     sql.NullInt32
				onsetDistrict sql.NullString
				createdAt     time.Time
				resultsStatus bool
			)

			err := measlesRows.Scan(
				&patientID,
				&measlesCode,
				&patientName,
				&sex,
				&dob,
				&ageMonths,
				&onsetDistrict,
				&createdAt,
				&resultsStatus,
			)
			if err != nil {
				sl.Error("Failed to scan measles case", "error", err)
				continue
			}

			// Calculate age display
			var ageDisplay string
			if ageMonths.Valid {
				ageDisplay = fmt.Sprintf("%d months", ageMonths.Int32)
			} else if dob.Valid {
				age := time.Now().Year() - dob.Time.Year()
				if time.Now().YearDay() < dob.Time.YearDay() {
					age--
				}
				ageDisplay = fmt.Sprintf("%d years", age)
			}

			measlesCases = append(measlesCases, fiber.Map{
				"ID":            patientID,
				"MeaslesCode":   measlesCode.String,
				"Name":          patientName,
				"Age":           ageDisplay,
				"Gender":        sex,
				"District":      onsetDistrict.String,
				"Status":        "Active", // Default status for measles cases
				"ResultsStatus": resultsStatus,
				"CreatedAt":     createdAt.Format("2006-01-02 15:04"),
			})
		}
	}

	// Load MPOX cases as well
	mpoxQuery := `
		SELECT 
			mci.id,
			mci.case_id,
			mci.case_status,
			mci.case_classification,
			mci.date,
			mpd.surname,
			mpd.other_names,
			mpd.sex,
			mpd.age,
			mpd.parish,
			mpd.sub_county,
			mpd.onset_date,
			mpd.rash_onset_date,
			CASE WHEN mli.id IS NOT NULL THEN true ELSE false END as lab_status
		FROM mpox_case_investigation mci
		LEFT JOIN mpox_patient_demographics mpd ON mci.case_id = mpd.case_id
		LEFT JOIN mpox_lab_investigation mli ON mci.case_id = mli.case_id
		ORDER BY mci.date DESC`

	mpoxRows, err := db.QueryContext(c.Context(), mpoxQuery)
	if err != nil {
		sl.Error("Failed to query MPOX cases", "error", err)
		// Continue with other cases if MPOX query fails
	}

	var mpoxCases []fiber.Map
	if mpoxRows != nil {
		defer mpoxRows.Close()
		for mpoxRows.Next() {
			var (
				id                 int64
				caseID             string
				caseStatus         sql.NullString
				caseClassification sql.NullString
				date               time.Time
				surname            string
				otherNames         sql.NullString
				sex                string
				age                int
				parish             sql.NullString
				subCounty          sql.NullString
				onsetDate          sql.NullTime
				rashOnsetDate      sql.NullTime
				labStatus          bool
			)

			err := mpoxRows.Scan(
				&id,
				&caseID,
				&caseStatus,
				&caseClassification,
				&date,
				&surname,
				&otherNames,
				&sex,
				&age,
				&parish,
				&subCounty,
				&onsetDate,
				&rashOnsetDate,
				&labStatus,
			)
			if err != nil {
				sl.Error("Failed to scan MPOX case", "error", err)
				continue
			}

			// Format name
			name := surname
			if otherNames.Valid && otherNames.String != "" {
				name = fmt.Sprintf("%s %s", surname, otherNames.String)
			}

			// Format location
			location := ""
			if parish.Valid && parish.String != "" {
				location = parish.String
			}
			if subCounty.Valid && subCounty.String != "" {
				if location != "" {
					location += ", "
				}
				location += subCounty.String
			}

			// Determine status
			status := "Active"
			if caseStatus.Valid && caseStatus.String != "" {
				status = caseStatus.String
			}

			// Format onset date
			onsetDateStr := ""
			if onsetDate.Valid {
				onsetDateStr = onsetDate.Time.Format("2006-01-02")
			}

			mpoxCases = append(mpoxCases, fiber.Map{
				"ID":             id,
				"CaseID":         caseID,
				"Name":           name,
				"Age":            fmt.Sprintf("%d years", age),
				"Gender":         sex,
				"Location":       location,
				"Status":         status,
				"Classification": caseClassification.String,
				"LabStatus":      labStatus,
				"OnsetDate":      onsetDateStr,
				"CreatedAt":      date.Format("2006-01-02 15:04"),
			})
		}
	}

	// Load Polio cases as well
	polioQuery := `
		SELECT 
			pci.id,
			pci.case_id,
			pfe.final_classification as case_status,
			pfe.final_classification as case_classification,
			pci.created_at as date,
			pi.patient_name,
			pi.sex,
			pi.age_years,
			pi.age_months,
			pi.district
		FROM polio_case_investigation pci
		LEFT JOIN polio_identification pi ON pci.case_id = pi.case_id
		LEFT JOIN polio_follow_up_examination pfe ON pci.case_id = pfe.case_id
		ORDER BY pci.created_at DESC`

	polioRows, err := db.QueryContext(c.Context(), polioQuery)
	if err != nil {
		sl.Error("Failed to query Polio cases", "error", err)
		// Continue with other cases if Polio query fails
	}

	var polioCases []fiber.Map
	if polioRows != nil {
		defer polioRows.Close()
		for polioRows.Next() {
			var (
				id                 int64
				caseID             string
				caseStatus         sql.NullString
				caseClassification sql.NullString
				date               time.Time
				patientName        string
				sex                string
				ageYears           sql.NullInt32
				ageMonths          sql.NullInt32
				district           string
			)

			err := polioRows.Scan(
				&id,
				&caseID,
				&caseStatus,
				&caseClassification,
				&date,
				&patientName,
				&sex,
				&ageYears,
				&ageMonths,
				&district,
			)
			if err != nil {
				sl.Error("Failed to scan Polio case", "error", err)
				continue
			}

			// Calculate age display
			var ageDisplay string
			if ageYears.Valid {
				ageDisplay = fmt.Sprintf("%d years", ageYears.Int32)
			} else if ageMonths.Valid {
				ageDisplay = fmt.Sprintf("%d months", ageMonths.Int32)
			}

			// Determine status
			status := "Active"
			if caseStatus.Valid && caseStatus.String != "" {
				status = caseStatus.String
			}

			polioCases = append(polioCases, fiber.Map{
				"ID":             id,
				"CaseID":         caseID,
				"Name":           patientName,
				"Age":            ageDisplay,
				"Gender":         sex,
				"District":       district,
				"Status":         status,
				"Classification": caseClassification.String,
				"CreatedAt":      date.Format("2006-01-02 15:04"),
			})
		}
	}

	// Return HTML response
	data := NewTemplateData(c, store)
	data.Form = fiber.Map{
		"Title":        "Case Management",
		"VHFCases":     cases,
		"MeaslesCases": measlesCases,
		"MpoxCases":    mpoxCases,
		"PolioCases":   polioCases,
	}
	return GenerateHTML(c, db, data, "vhf_list")
}

// HandlerVHFView handles viewing a single VHF case
func HandlerVHFView(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid patient ID")
	}

	patient, err := models.GetVHFPatient(db, id)
	if err != nil {
		sl.Error("Failed to get patient", "error", err)
		return c.Status(500).SendString("Failed to retrieve patient information")
	}

	// Get laboratory data
	lab, err := models.GetVHFLaboratory(db, id)
	if err != nil && err != sql.ErrNoRows {
		sl.Error("Failed to get laboratory data", "error", err)
		return c.Status(500).SendString("Failed to retrieve laboratory information")
	}

	// Get clinical signs
	clinicalSigns, err := models.GetVHFClinicalSigns(db, id)
	if err != nil && err != sql.ErrNoRows {
		sl.Error("Failed to get clinical signs", "error", err)
		return c.Status(500).SendString("Failed to retrieve clinical signs")
	}

	// Get hospitalization data
	hospitalization, err := models.GetVHFHospitalization(db, id)
	if err != nil && err != sql.ErrNoRows {
		sl.Error("Failed to get hospitalization data", "error", err)
		return c.Status(500).SendString("Failed to retrieve hospitalization information")
	}

	// Get risk factors
	riskFactors, err := models.GetVHFRiskFactors(db, id)
	if err != nil && err != sql.ErrNoRows {
		sl.Error("Failed to get risk factors", "error", err)
		return c.Status(500).SendString("Failed to retrieve risk factors")
	}

	// Get investigator data
	investigator, err := models.GetVHFInvestigator(db, id)
	if err != nil && err != sql.ErrNoRows {
		sl.Error("Failed to get investigator data", "error", err)
		return c.Status(500).SendString("Failed to retrieve investigator information")
	}

	// Get location data for dropdowns
	districts, err := models.GetDistricts(db)
	if err != nil {
		sl.Error("Failed to get districts", "error", err)
		return c.Status(500).SendString("Failed to retrieve districts")
	}

	// Extract district names
	var districtNames []string
	for _, d := range districts {
		districtNames = append(districtNames, d.Name)
	}

	// Get subcounties for the patient's district (if available)
	var subcountyNames []string
	var subcounties []models.Subcounty
	if patient.District != "" {
		district, err := models.GetDistrictByName(db, patient.District)
		if err == nil && district != nil {
			subcounties, err = models.GetSubcountiesByDistrict(db, district.ID)
			if err == nil {
				for _, s := range subcounties {
					subcountyNames = append(subcountyNames, s.Name)
				}
			}
		}
	}

	// Get parishes for the patient's subcounty (if available)
	var parishNames []string
	var parishes []models.Parish
	if patient.Subcounty != "" && len(subcountyNames) > 0 {
		// Find the subcounty ID
		for _, s := range subcounties {
			if s.Name == patient.Subcounty {
				parishes, err = models.GetParishesBySubcounty(db, s.ID)
				if err == nil {
					for _, p := range parishes {
						parishNames = append(parishNames, p.Name)
					}
				}
				break
			}
		}
	}

	// Get villages for the patient's parish (if available)
	var villageNames []string
	if patient.Parish != "" && len(parishNames) > 0 {
		// Find the parish ID
		for _, p := range parishes {
			if p.Name == patient.Parish {
				villages, err := models.GetVillagesByParish(db, p.ID)
				if err == nil {
					for _, v := range villages {
						villageNames = append(villageNames, v.Name)
					}
				}
				break
			}
		}
	}

	// Format lab data for display
	labData := fiber.Map{}
	if lab != nil {
		labData = fiber.Map{
			"SampleCollectionDate": lab.SampleCollectionDate.Time.Format("2006-01-02"),
			"SampleCollectionTime": lab.SampleCollectionTime.String,
			"SampleType":           lab.SampleType.String,
			"OtherSampleType":      lab.OtherSampleType.String,
			"RequestedTest":        lab.RequestedTest.String,
			"Serology":             lab.Serology.String,
			"MalariaRDT":           lab.MalariaRDT.String,
			"HIVRDT":               lab.HIVRDT.String,
			"TestResult":           lab.TestResult.String,
			"DateTested":           lab.DateTested.Time.Format("2006-01-02"),
			"LabName":              lab.LabName.String,
		}
	}

	// Format clinical signs data
	clinicalSignsData := fiber.Map{}
	if clinicalSigns != nil {
		clinicalSignsData = fiber.Map{
			"DateInitialOnset":                      clinicalSigns.DateInitialOnset.Time.Format("2006-01-02"),
			"TempSource":                            clinicalSigns.TempSource.String,
			"Temperature":                           clinicalSigns.Temperature.Float64,
			"Fever":                                 clinicalSigns.Fever.Bool,
			"DateFever":                             clinicalSigns.DateFever.Time.Format("2006-01-02"),
			"DurationFever":                         clinicalSigns.DurationFever.Int32,
			"Vomiting":                              clinicalSigns.Vomiting.Bool,
			"DateVomiting":                          clinicalSigns.DateVomiting.Time.Format("2006-01-02"),
			"DurationVomiting":                      clinicalSigns.DurationVomiting.Int32,
			"Nausea":                                clinicalSigns.Nausea.Bool,
			"DateNausea":                            clinicalSigns.DateNausea.Time.Format("2006-01-02"),
			"DurationNausea":                        clinicalSigns.DurationNausea.Int32,
			"Diarrhea":                              clinicalSigns.Diarrhea.Bool,
			"DateDiarrhea":                          clinicalSigns.DateDiarrhea.Time.Format("2006-01-02"),
			"DurationDiarrhea":                      clinicalSigns.DurationDiarrhea.Int32,
			"IntenseFatigueGeneralWeakness":         clinicalSigns.IntenseFatigueGeneralWeakness.Bool,
			"DateIntenseFatigueGeneralWeakness":     clinicalSigns.DateIntenseFatigueGeneralWeakness.Time.Format("2006-01-02"),
			"DurationIntenseFatigueGeneralWeakness": clinicalSigns.DurationIntenseFatigueGeneralWeakness.Int32,
			"EpigastricPain":                        clinicalSigns.EpigastricPain.Bool,
			"DateEpigastricPain":                    clinicalSigns.DateEpigastricPain.Time.Format("2006-01-02"),
			"DurationEpigastricPain":                clinicalSigns.DurationEpigastricPain.Int32,
			"LowerAbdominalPain":                    clinicalSigns.LowerAbdominalPain.Bool,
			"DateLowerAbdominalPain":                clinicalSigns.DateLowerAbdominalPain.Time.Format("2006-01-02"),
			"DurationLowerAbdominalPain":            clinicalSigns.DurationLowerAbdominalPain.Int32,
			"ChestPain":                             clinicalSigns.ChestPain.Bool,
			"DateChestPain":                         clinicalSigns.DateChestPain.Time.Format("2006-01-02"),
			"DurationChestPain":                     clinicalSigns.DurationChestPain.Int32,
			"MusclePain":                            clinicalSigns.MusclePain.Bool,
			"DateMusclePain":                        clinicalSigns.DateMusclePain.Time.Format("2006-01-02"),
			"DurationMusclePain":                    clinicalSigns.DurationMusclePain.Int32,
			"JointPain":                             clinicalSigns.JointPain.Bool,
			"DateJointPain":                         clinicalSigns.DateJointPain.Time.Format("2006-01-02"),
			"DurationJointPain":                     clinicalSigns.DurationJointPain.Int32,
			"Headache":                              clinicalSigns.Headache.Bool,
			"DateHeadache":                          clinicalSigns.DateHeadache.Time.Format("2006-01-02"),
			"DurationHeadache":                      clinicalSigns.DurationHeadache.Int32,
			"Cough":                                 clinicalSigns.Cough.Bool,
			"DateCough":                             clinicalSigns.DateCough.Time.Format("2006-01-02"),
			"DurationCough":                         clinicalSigns.DurationCough.Int32,
			"DifficultyBreathing":                   clinicalSigns.DifficultyBreathing.Bool,
			"DateDifficultyBreathing":               clinicalSigns.DateDifficultyBreathing.Time.Format("2006-01-02"),
			"DurationDifficultyBreathing":           clinicalSigns.DurationDifficultyBreathing.Int32,
			"DifficultySwallowing":                  clinicalSigns.DifficultySwallowing.Bool,
			"DateDifficultySwallowing":              clinicalSigns.DateDifficultySwallowing.Time.Format("2006-01-02"),
			"DurationDifficultySwallowing":          clinicalSigns.DurationDifficultySwallowing.Int32,
			"SoreThroat":                            clinicalSigns.SoreThroat.Bool,
			"DateSoreThroat":                        clinicalSigns.DateSoreThroat.Time.Format("2006-01-02"),
			"DurationSoreThroat":                    clinicalSigns.DurationSoreThroat.Int32,
			"Jaundice":                              clinicalSigns.Jaundice.Bool,
			"DateJaundice":                          clinicalSigns.DateJaundice.Time.Format("2006-01-02"),
			"DurationJaundice":                      clinicalSigns.DurationJaundice.Int32,
			"Conjunctivitis":                        clinicalSigns.Conjunctivitis.Bool,
			"DateConjunctivitis":                    clinicalSigns.DateConjunctivitis.Time.Format("2006-01-02"),
			"DurationConjunctivitis":                clinicalSigns.DurationConjunctivitis.Int32,
			"SkinRash":                              clinicalSigns.SkinRash.Bool,
			"DateSkinRash":                          clinicalSigns.DateSkinRash.Time.Format("2006-01-02"),
			"DurationSkinRash":                      clinicalSigns.DurationSkinRash.Int32,
			"Hiccups":                               clinicalSigns.Hiccups.Bool,
			"DateHiccups":                           clinicalSigns.DateHiccups.Time.Format("2006-01-02"),
			"DurationHiccups":                       clinicalSigns.DurationHiccups.Int32,
			"PainBehindEyes":                        clinicalSigns.PainBehindEyes.Bool,
			"DatePainBehindEyes":                    clinicalSigns.DatePainBehindEyes.Time.Format("2006-01-02"),
			"DurationPainBehindEyes":                clinicalSigns.DurationPainBehindEyes.Int32,
			"SensitiveToLight":                      clinicalSigns.SensitiveToLight.Bool,
			"DateSensitiveToLight":                  clinicalSigns.DateSensitiveToLight.Time.Format("2006-01-02"),
			"DurationSensitiveToLight":              clinicalSigns.DurationSensitiveToLight.Int32,
			"ComaUnconscious":                       clinicalSigns.ComaUnconscious.Bool,
			"DateComaUnconscious":                   clinicalSigns.DateComaUnconscious.Time.Format("2006-01-02"),
			"DurationComaUnconscious":               clinicalSigns.DurationComaUnconscious.Int32,
			"ConfusedOrDisoriented":                 clinicalSigns.ConfusedOrDisoriented.Bool,
			"DateConfusedOrDisoriented":             clinicalSigns.DateConfusedOrDisoriented.Time.Format("2006-01-02"),
			"DurationConfusedOrDisoriented":         clinicalSigns.DurationConfusedOrDisoriented.Int32,
			"Convulsions":                           clinicalSigns.Convulsions.Bool,
			"DateConvulsions":                       clinicalSigns.DateConvulsions.Time.Format("2006-01-02"),
			"DurationConvulsions":                   clinicalSigns.DurationConvulsions.Int32,
			"UnexplainedBleeding":                   clinicalSigns.UnexplainedBleeding.Bool,
			"DateUnexplainedBleeding":               clinicalSigns.DateUnexplainedBleeding.Time.Format("2006-01-02"),
			"DurationUnexplainedBleeding":           clinicalSigns.DurationUnexplainedBleeding.Int32,
			"BleedingOfTheGums":                     clinicalSigns.BleedingOfTheGums.Bool,
			"DateBleedingOfTheGums":                 clinicalSigns.DateBleedingOfTheGums.Time.Format("2006-01-02"),
			"DurationBleedingOfTheGums":             clinicalSigns.DurationBleedingOfTheGums.Int32,
			"BleedingFromInjectionSite":             clinicalSigns.BleedingFromInjectionSite.Bool,
			"DateBleedingFromInjectionSite":         clinicalSigns.DateBleedingFromInjectionSite.Time.Format("2006-01-02"),
			"DurationBleedingFromInjectionSite":     clinicalSigns.DurationBleedingFromInjectionSite.Int32,
			"NoseBleedEpistaxis":                    clinicalSigns.NoseBleedEpistaxis.Bool,
			"DateNoseBleedEpistaxis":                clinicalSigns.DateNoseBleedEpistaxis.Time.Format("2006-01-02"),
			"DurationNoseBleedEpistaxis":            clinicalSigns.DurationNoseBleedEpistaxis.Int32,
			"BloodyStool":                           clinicalSigns.BloodyStool.Bool,
			"DateBloodyStool":                       clinicalSigns.DateBloodyStool.Time.Format("2006-01-02"),
			"DurationBloodyStool":                   clinicalSigns.DurationBloodyStool.Int32,
			"BloodInVomit":                          clinicalSigns.BloodInVomit.Bool,
			"DateBloodInVomit":                      clinicalSigns.DateBloodInVomit.Time.Format("2006-01-02"),
			"DurationBloodInVomit":                  clinicalSigns.DurationBloodInVomit.Int32,
			"CoughingUpBloodHemoptysis":             clinicalSigns.CoughingUpBloodHemoptysis.Bool,
			"DateCoughingUpBloodHemoptysis":         clinicalSigns.DateCoughingUpBloodHemoptysis.Time.Format("2006-01-02"),
			"DurationCoughingUpBloodHemoptysis":     clinicalSigns.DurationCoughingUpBloodHemoptysis.Int32,
			"BleedingFromVagina":                    clinicalSigns.BleedingFromVagina.Bool,
			"DateBleedingFromVagina":                clinicalSigns.DateBleedingFromVagina.Time.Format("2006-01-02"),
			"DurationBleedingFromVagina":            clinicalSigns.DurationBleedingFromVagina.Int32,
			"BruisingOfTheSkin":                     clinicalSigns.BruisingOfTheSkin.Bool,
			"DateBruisingOfTheSkin":                 clinicalSigns.DateBruisingOfTheSkin.Time.Format("2006-01-02"),
			"DurationBruisingOfTheSkin":             clinicalSigns.DurationBruisingOfTheSkin.Int32,
			"BloodInUrine":                          clinicalSigns.BloodInUrine.Bool,
			"DateBloodInUrine":                      clinicalSigns.DateBloodInUrine.Time.Format("2006-01-02"),
			"DurationBloodInUrine":                  clinicalSigns.DurationBloodInUrine.Int32,
			"OtherHemorrhagicSymptoms":              clinicalSigns.OtherHemorrhagicSymptoms.Bool,
			"DateOtherHemorrhagicSymptoms":          clinicalSigns.DateOtherHemorrhagicSymptoms.Time.Format("2006-01-02"),
			"DurationOtherHemorrhagicSymptoms":      clinicalSigns.DurationOtherHemorrhagicSymptoms.Int32,
		}
	}

	// Format investigator data
	investigatorData := fiber.Map{}
	if investigator != nil {
		investigatorData = fiber.Map{
			"InvestigatorName":  investigator.InvestigatorName,
			"Phone":             investigator.Phone,
			"Email":             investigator.Email,
			"Position":          investigator.Position,
			"District":          investigator.District,
			"HealthFacility":    investigator.HealthFacility,
			"InformationSource": investigator.InformationSource,
			"ProxyName":         investigator.ProxyName,
			"ProxyRelation":     investigator.ProxyRelation,
		}
	}

	data := NewTemplateData(c, store)
	data.Case = patient
	data.ClinicalSigns = clinicalSigns
	data.Hospitalization = hospitalization
	data.RiskFactors = riskFactors
	data.Investigator = investigator
	data.Districts = districtNames
	data.Subcounties = subcountyNames
	data.Parishes = parishNames
	data.Villages = villageNames
	data.Form = fiber.Map{
		"Title":           "Update VHF Case Investigation Form",
		"Lab":             labData,
		"ClinicalSigns":   clinicalSignsData,
		"Hospitalization": hospitalization,
		"RiskFactors":     riskFactors,
		"Investigator":    investigatorData,
	}
	return GenerateHTML(c, db, data, "update_vhf_cif")
}

// HandlerVHFSuccess handles the success page after form submission
func HandlerVHFSuccess(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	// Get the case code from the query parameters
	caseCode := c.Query("case_code")
	if caseCode == "" {
		sl.Error("No case code provided in success page")
		return c.Status(400).SendString("No case code provided")
	}

	data := NewTemplateData(c, store)
	data.Form = fiber.Map{
		"CaseCode": caseCode,
	}
	return GenerateHTML(c, db, data, "vhf_success")
}

// HandlerVHFLabForm handles displaying the lab form for a VHF case
func HandlerVHFLabForm(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	caseID := c.Params("id")
	if caseID == "" {
		return c.Status(400).SendString("Case ID is required")
	}
	var patient models.VHFPatient
	var lab models.VHFLaboratory
	err := db.QueryRow(`
		SELECT p.id, p.surname, p.other_names, p.date_of_birth,
		p.age_years, p.age_months, p.gender, p.patient_phone,
		p.phone_owner, p.next_of_kin, p.next_of_kin_phone,
		p.data_capturer_name, p.data_capturer_phone, p.reporting_health_facility_name,
		p.case_code, p.status, p.date_of_death, p.head_of_household,
		p.village_town, p.parish, p.subcounty, p.district,
		p.country_of_residence, p.occupation, p.ill_village_town,
		p.ill_subcounty, p.ill_district, p.latitude, p.longitude,
		p.date_residing_from, p.date_residing_to, p.created_at,
		l.sample_collection_date, l.sample_collection_time, l.sample_type, 
		l.other_sample_type, l.requested_test, l.serology, l.malaria_rdt, l.hiv_rdt,
		l.test_result, l.date_tested, l.lab_name
		FROM vhf_patients p
		LEFT JOIN vhf_laboratory l ON p.id = l.patient_id
		WHERE p.id = $1
	`, caseID).Scan(
		&patient.ID, &patient.Surname, &patient.OtherNames, &patient.DateOfBirth,
		&patient.AgeYears, &patient.AgeMonths, &patient.Gender, &patient.PatientPhone,
		&patient.PhoneOwner, &patient.NextOfKin, &patient.NextOfKinPhone,
		&patient.DataCapturerName, &patient.DataCapturerPhone, &patient.ReportingHealthFacilityName,
		&patient.CaseCode, &patient.Status, &patient.DateOfDeath, &patient.HeadOfHousehold,
		&patient.VillageTown, &patient.Parish, &patient.Subcounty, &patient.District,
		&patient.CountryOfResidence, &patient.Occupation, &patient.IllVillageTown,
		&patient.IllSubcounty, &patient.IllDistrict, &patient.Latitude, &patient.Longitude,
		&patient.DateResidingFrom, &patient.DateResidingTo, &patient.CreatedAt,
		&lab.SampleCollectionDate, &lab.SampleCollectionTime, &lab.SampleType,
		&lab.OtherSampleType, &lab.RequestedTest, &lab.Serology,
		&lab.MalariaRDT, &lab.HIVRDT, &lab.TestResult,
		&lab.DateTested, &lab.LabName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No patient found with ID: %s", caseID)
			return c.Status(404).SendString("Case not found")
		}
		log.Printf("Database error for patient ID %s: %v", caseID, err)
		return c.Status(500).SendString("Database error")
	}

	// Log successful retrieval
	log.Printf("Successfully retrieved patient with ID: %s", caseID)
	log.Printf("DEBUG: Patient data - ID: %d, CaseCode: %s, Surname: %s", patient.ID, patient.CaseCode.String, patient.Surname)

	// Return HTML response with patient and lab data
	data := NewTemplateData(c, store)
	data.Patient = patient
	data.Lab = lab
	data.Form = fiber.Map{
		"Title":   "VHF Lab Form",
		"Patient": patient, // Also add to Form as fallback
		"Lab":     lab,     // Also add to Form as fallback
	}

	// Also set Case field for compatibility
	data.Case = patient

	// Ensure Optionz is initialized
	if data.Optionz == nil {
		data.Optionz = make(map[string]map[string]string)
	}

	// Debug logging
	log.Printf("DEBUG: VHF Lab Form - Patient ID: %d, Case Code: %s", patient.ID, patient.CaseCode.String)
	log.Printf("DEBUG: VHF Lab Form - Data.Patient is set: %v", data.Patient != nil)
	log.Printf("DEBUG: VHF Lab Form - Data type: %T", data)
	log.Printf("DEBUG: VHF Lab Form - Patient type: %T", data.Patient)

	// Additional safety check
	if data.Patient == nil {
		log.Printf("ERROR: Patient data is nil, creating fallback")
		data.Patient = &models.VHFPatient{
			ID:       patient.ID,
			CaseCode: patient.CaseCode,
			Surname:  patient.Surname,
		}
	}

	// Log template data structure
	log.Printf("DEBUG: Template data structure - Patient: %+v", data.Patient)
	log.Printf("DEBUG: Template data structure - Lab: %+v", data.Lab)
	log.Printf("DEBUG: Template data structure - Form: %+v", data.Form)

	// Try to render the template with error handling
	log.Printf("DEBUG: About to render template 'vhf_lab_form'")
	log.Printf("DEBUG: Template data keys - Patient: %v, Lab: %v, Case: %v", data.Patient != nil, data.Lab != nil, data.Case != nil)
	return GenerateHTML(c, db, data, "vhf_lab_form")
}

// HandlerVHFLabSave handles the submission of laboratory information
func HandlerVHFLabSave(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	caseID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid case ID",
		})
	}

	// Parse sample collection date
	sampleCollectionDate, err := time.Parse("2006-01-02", c.FormValue("sample_collection_date"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid sample collection date format",
		})
	}

	// Parse date tested
	var dateTested sql.NullTime
	if dateTestedStr := c.FormValue("date_tested"); dateTestedStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateTestedStr)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid date tested format",
			})
		}
		dateTested = sql.NullTime{Time: parsedDate, Valid: true}
	}

	// Create laboratory record
	laboratory := &models.VHFLaboratory{
		PatientID:            caseID,
		SampleCollectionDate: sql.NullTime{Time: sampleCollectionDate, Valid: true},
		SampleCollectionTime: sql.NullString{String: c.FormValue("sample_collection_time"), Valid: c.FormValue("sample_collection_time") != ""},
		SampleType:           sql.NullString{String: c.FormValue("sample_type"), Valid: c.FormValue("sample_type") != ""},
		OtherSampleType:      sql.NullString{String: c.FormValue("other_sample_type"), Valid: c.FormValue("other_sample_type") != ""},
		RequestedTest:        sql.NullString{String: c.FormValue("requested_test"), Valid: c.FormValue("requested_test") != ""},
		Serology:             sql.NullString{String: c.FormValue("serology"), Valid: c.FormValue("serology") != ""},
		MalariaRDT:           sql.NullString{String: c.FormValue("malaria_rdt"), Valid: c.FormValue("malaria_rdt") != ""},
		HIVRDT:               sql.NullString{String: c.FormValue("hiv_rdt"), Valid: c.FormValue("hiv_rdt") != ""},
		TestResult:           sql.NullString{String: c.FormValue("test_result"), Valid: c.FormValue("test_result") != ""},
		DateTested:           dateTested,
		LabName:              sql.NullString{String: c.FormValue("lab_name"), Valid: c.FormValue("lab_name") != ""},
	}

	// Save to database
	err = models.SaveVHFLaboratory(db, laboratory)
	if err != nil {
		sl.Error("Failed to save laboratory data", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save laboratory data",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Laboratory data saved successfully",
	})
}

// HandlerVHFUpdate handles updating a VHF case (full implementation)
func HandlerVHFUpdate(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config config.Config) error {
	caseCode := c.FormValue("case_code")
	sl.Info("HandlerVHFUpdate called", "case_code", caseCode)

	if caseCode == "" {
		sl.Error("Missing case code")
		return c.Status(400).SendString("Missing case code")
	}

	patient, err := models.GetVHFPatientByCaseCode(db, caseCode)
	if err != nil {
		sl.Error("Failed to get patient by case code", "error", err, "case_code", caseCode)
		return c.Status(500).SendString("Database error: " + err.Error())
	}
	if patient == nil {
		sl.Error("Patient not found", "case_code", caseCode)
		return c.Status(404).SendString("Patient not found")
	}

	sl.Info("Found patient", "patient_id", patient.ID, "case_code", caseCode)
	patientID := patient.ID

	// --- Patient Info ---
	sl.Info("Updating patient info", "surname", c.FormValue("surname"), "other_names", c.FormValue("other_names"))
	patient.Surname = c.FormValue("surname")
	patient.OtherNames = c.FormValue("other_names")
	patient.DateOfBirth = parseNullTime(c.FormValue("date_of_birth"))
	patient.AgeYears = parseNullInt32(c.FormValue("age_years"))
	patient.Gender = parseNullString(c.FormValue("gender"))
	patient.PatientPhone = parseNullString(c.FormValue("patient_phone"))
	patient.PhoneOwner = parseNullString(c.FormValue("phone_owner"))
	patient.NextOfKin = parseNullString(c.FormValue("next_of_kin"))
	patient.NextOfKinPhone = parseNullString(c.FormValue("next_of_kin_phone"))
	patient.HeadOfHousehold = parseNullString(c.FormValue("head_of_household"))
	patient.District = c.FormValue("district")
	patient.Subcounty = c.FormValue("subcounty")
	patient.Parish = c.FormValue("parish")
	patient.VillageTown = c.FormValue("village")
	patient.CountryOfResidence = c.FormValue("country_of_residence")
	patient.Occupation = c.FormValue("occupation")
	patient.Latitude = parseNullFloat64FromCSV(c.FormValue("gps_coordinates"), 0)
	patient.Longitude = parseNullFloat64FromCSV(c.FormValue("gps_coordinates"), 1)
	patient.DateResidingFrom = parseNullTime(c.FormValue("date_residing_from"))
	patient.DateResidingTo = parseNullTime(c.FormValue("date_residing_to"))

	if err := models.SaveVHFPatient(db, patient); err != nil {
		sl.Error("Failed to update patient info", "error", err)
		return c.Status(500).SendString("Failed to update patient info: " + err.Error())
	}
	sl.Info("Patient info updated successfully")

	// --- Clinical Signs ---
	clinicalSigns, _ := models.GetVHFClinicalSigns(db, patientID)
	if clinicalSigns == nil {
		clinicalSigns = &models.VHFClinicalSigns{PatientID: patientID}
	}
	clinicalSigns.DateInitialOnset = parseNullTime(c.FormValue("date_initial_onset"))
	clinicalSigns.TempSource = parseNullString(c.FormValue("temp_source"))
	clinicalSigns.Temperature = parseNullFloat64(c.FormValue("temperature"))
	clinicalSigns.Fever = parseNullBool(c.FormValue("fever"))
	clinicalSigns.DateFever = parseNullTime(c.FormValue("date_fever"))
	clinicalSigns.DurationFever = parseNullInt32(c.FormValue("duration_fever"))
	clinicalSigns.Vomiting = parseNullBool(c.FormValue("vomiting"))
	clinicalSigns.DateVomiting = parseNullTime(c.FormValue("date_vomiting"))
	clinicalSigns.DurationVomiting = parseNullInt32(c.FormValue("duration_vomiting"))
	clinicalSigns.Nausea = parseNullBool(c.FormValue("nausea"))
	clinicalSigns.DateNausea = parseNullTime(c.FormValue("date_nausea"))
	clinicalSigns.DurationNausea = parseNullInt32(c.FormValue("duration_nausea"))
	clinicalSigns.Diarrhea = parseNullBool(c.FormValue("diarrhea"))
	clinicalSigns.DateDiarrhea = parseNullTime(c.FormValue("date_diarrhea"))
	clinicalSigns.DurationDiarrhea = parseNullInt32(c.FormValue("duration_diarrhea"))
	clinicalSigns.IntenseFatigueGeneralWeakness = parseNullBool(c.FormValue("intense_fatigue_general_weakness"))
	clinicalSigns.DateIntenseFatigueGeneralWeakness = parseNullTime(c.FormValue("date_intense_fatigue_general_weakness"))
	clinicalSigns.DurationIntenseFatigueGeneralWeakness = parseNullInt32(c.FormValue("duration_intense_fatigue_general_weakness"))
	clinicalSigns.EpigastricPain = parseNullBool(c.FormValue("epigastric_pain"))
	clinicalSigns.DateEpigastricPain = parseNullTime(c.FormValue("date_epigastric_pain"))
	clinicalSigns.DurationEpigastricPain = parseNullInt32(c.FormValue("duration_epigastric_pain"))
	clinicalSigns.LowerAbdominalPain = parseNullBool(c.FormValue("lower_abdominal_pain"))
	clinicalSigns.DateLowerAbdominalPain = parseNullTime(c.FormValue("date_lower_abdominal_pain"))
	clinicalSigns.DurationLowerAbdominalPain = parseNullInt32(c.FormValue("duration_lower_abdominal_pain"))
	clinicalSigns.ChestPain = parseNullBool(c.FormValue("chest_pain"))
	clinicalSigns.DateChestPain = parseNullTime(c.FormValue("date_chest_pain"))
	clinicalSigns.DurationChestPain = parseNullInt32(c.FormValue("duration_chest_pain"))
	clinicalSigns.MusclePain = parseNullBool(c.FormValue("muscle_pain"))
	clinicalSigns.DateMusclePain = parseNullTime(c.FormValue("date_muscle_pain"))
	clinicalSigns.DurationMusclePain = parseNullInt32(c.FormValue("duration_muscle_pain"))
	clinicalSigns.JointPain = parseNullBool(c.FormValue("joint_pain"))
	clinicalSigns.DateJointPain = parseNullTime(c.FormValue("date_joint_pain"))
	clinicalSigns.DurationJointPain = parseNullInt32(c.FormValue("duration_joint_pain"))
	clinicalSigns.Headache = parseNullBool(c.FormValue("headache"))
	clinicalSigns.DateHeadache = parseNullTime(c.FormValue("date_headache"))
	clinicalSigns.DurationHeadache = parseNullInt32(c.FormValue("duration_headache"))
	clinicalSigns.Cough = parseNullBool(c.FormValue("cough"))
	clinicalSigns.DateCough = parseNullTime(c.FormValue("date_cough"))
	clinicalSigns.DurationCough = parseNullInt32(c.FormValue("duration_cough"))
	clinicalSigns.DifficultyBreathing = parseNullBool(c.FormValue("difficulty_breathing"))
	clinicalSigns.DateDifficultyBreathing = parseNullTime(c.FormValue("date_difficulty_breathing"))
	clinicalSigns.DurationDifficultyBreathing = parseNullInt32(c.FormValue("duration_difficulty_breathing"))
	clinicalSigns.DifficultySwallowing = parseNullBool(c.FormValue("difficulty_swallowing"))
	clinicalSigns.DateDifficultySwallowing = parseNullTime(c.FormValue("date_swallowing"))
	clinicalSigns.DurationDifficultySwallowing = parseNullInt32(c.FormValue("duration_difficulty_swallowing"))
	clinicalSigns.SoreThroat = parseNullBool(c.FormValue("sore_throat"))
	clinicalSigns.DateSoreThroat = parseNullTime(c.FormValue("date_sore_throat"))
	clinicalSigns.DurationSoreThroat = parseNullInt32(c.FormValue("duration_sore_throat"))
	clinicalSigns.Jaundice = parseNullBool(c.FormValue("jaundice"))
	clinicalSigns.DateJaundice = parseNullTime(c.FormValue("date_jaundice"))
	clinicalSigns.DurationJaundice = parseNullInt32(c.FormValue("duration_jaundice"))
	clinicalSigns.Conjunctivitis = parseNullBool(c.FormValue("conjunctivitis"))
	clinicalSigns.DateConjunctivitis = parseNullTime(c.FormValue("date_conjunctivitis"))
	clinicalSigns.DurationConjunctivitis = parseNullInt32(c.FormValue("duration_conjunctivitis"))
	clinicalSigns.SkinRash = parseNullBool(c.FormValue("skin_rash"))
	clinicalSigns.DateSkinRash = parseNullTime(c.FormValue("date_skin_rash"))
	clinicalSigns.DurationSkinRash = parseNullInt32(c.FormValue("duration_skin_rash"))
	clinicalSigns.Hiccups = parseNullBool(c.FormValue("hiccups"))
	clinicalSigns.DateHiccups = parseNullTime(c.FormValue("date_hiccups"))
	clinicalSigns.DurationHiccups = parseNullInt32(c.FormValue("duration_hiccups"))
	clinicalSigns.PainBehindEyes = parseNullBool(c.FormValue("pain_behind_eyes"))
	clinicalSigns.DatePainBehindEyes = parseNullTime(c.FormValue("date_pain_behind_eyes"))
	clinicalSigns.DurationPainBehindEyes = parseNullInt32(c.FormValue("duration_pain_behind_eyes"))
	clinicalSigns.SensitiveToLight = parseNullBool(c.FormValue("sensitive_to_light"))
	clinicalSigns.DateSensitiveToLight = parseNullTime(c.FormValue("date_sensitive_to_light"))
	clinicalSigns.DurationSensitiveToLight = parseNullInt32(c.FormValue("duration_sensitive_to_light"))
	clinicalSigns.ComaUnconscious = parseNullBool(c.FormValue("coma_unconscious"))
	clinicalSigns.DateComaUnconscious = parseNullTime(c.FormValue("date_coma_unconscious"))
	clinicalSigns.DurationComaUnconscious = parseNullInt32(c.FormValue("duration_coma_unconscious"))
	clinicalSigns.ConfusedOrDisoriented = parseNullBool(c.FormValue("confused_or_disoriented"))
	clinicalSigns.DateConfusedOrDisoriented = parseNullTime(c.FormValue("date_confused_or_disoriented"))
	clinicalSigns.DurationConfusedOrDisoriented = parseNullInt32(c.FormValue("duration_confused_or_disoriented"))
	clinicalSigns.Convulsions = parseNullBool(c.FormValue("convulsions"))
	clinicalSigns.DateConvulsions = parseNullTime(c.FormValue("date_convulsions"))
	clinicalSigns.DurationConvulsions = parseNullInt32(c.FormValue("duration_convulsions"))
	clinicalSigns.UnexplainedBleeding = parseNullBool(c.FormValue("unexplained_bleeding"))
	clinicalSigns.DateUnexplainedBleeding = parseNullTime(c.FormValue("date_unexplained_bleeding"))
	clinicalSigns.DurationUnexplainedBleeding = parseNullInt32(c.FormValue("duration_unexplained_bleeding"))
	clinicalSigns.BleedingOfTheGums = parseNullBool(c.FormValue("bleeding_of_the_gums"))
	clinicalSigns.DateBleedingOfTheGums = parseNullTime(c.FormValue("date_bleeding_of_the_gums"))
	clinicalSigns.DurationBleedingOfTheGums = parseNullInt32(c.FormValue("duration_bleeding_of_the_gums"))
	clinicalSigns.BleedingFromInjectionSite = parseNullBool(c.FormValue("bleeding_from_injection_site"))
	clinicalSigns.DateBleedingFromInjectionSite = parseNullTime(c.FormValue("date_bleeding_from_injection_site"))
	clinicalSigns.DurationBleedingFromInjectionSite = parseNullInt32(c.FormValue("duration_bleeding_from_injection_site"))
	clinicalSigns.NoseBleedEpistaxis = parseNullBool(c.FormValue("nose_bleed_epistaxis"))
	clinicalSigns.DateNoseBleedEpistaxis = parseNullTime(c.FormValue("date_nose_bleed_epistaxis"))
	clinicalSigns.DurationNoseBleedEpistaxis = parseNullInt32(c.FormValue("duration_nose_bleed_epistaxis"))
	clinicalSigns.BloodyStool = parseNullBool(c.FormValue("bloody_stool"))
	clinicalSigns.DateBloodyStool = parseNullTime(c.FormValue("date_bloody_stool"))
	clinicalSigns.DurationBloodyStool = parseNullInt32(c.FormValue("duration_bloody_stool"))
	clinicalSigns.BloodInVomit = parseNullBool(c.FormValue("blood_in_vomit"))
	clinicalSigns.DateBloodInVomit = parseNullTime(c.FormValue("date_blood_in_vomit"))
	clinicalSigns.DurationBloodInVomit = parseNullInt32(c.FormValue("duration_blood_in_vomit"))
	clinicalSigns.CoughingUpBloodHemoptysis = parseNullBool(c.FormValue("coughing_up_blood_hemoptysis"))
	clinicalSigns.DateCoughingUpBloodHemoptysis = parseNullTime(c.FormValue("date_coughing_up_blood_hemoptysis"))
	clinicalSigns.DurationCoughingUpBloodHemoptysis = parseNullInt32(c.FormValue("duration_coughing_up_blood_hemoptysis"))
	clinicalSigns.BleedingFromVagina = parseNullBool(c.FormValue("bleeding_from_vagina"))
	clinicalSigns.DateBleedingFromVagina = parseNullTime(c.FormValue("date_bleeding_from_vagina"))
	clinicalSigns.DurationBleedingFromVagina = parseNullInt32(c.FormValue("duration_bleeding_from_vagina"))
	clinicalSigns.BruisingOfTheSkin = parseNullBool(c.FormValue("bruising_of_the_skin"))
	clinicalSigns.DateBruisingOfTheSkin = parseNullTime(c.FormValue("date_bruising_of_the_skin"))
	clinicalSigns.DurationBruisingOfTheSkin = parseNullInt32(c.FormValue("duration_bruising_of_the_skin"))
	clinicalSigns.BloodInUrine = parseNullBool(c.FormValue("blood_in_urine"))
	clinicalSigns.DateBloodInUrine = parseNullTime(c.FormValue("date_blood_in_urine"))
	clinicalSigns.DurationBloodInUrine = parseNullInt32(c.FormValue("duration_blood_in_urine"))
	clinicalSigns.OtherHemorrhagicSymptoms = parseNullBool(c.FormValue("other_hemorrhagic_symptoms"))
	clinicalSigns.DateOtherHemorrhagicSymptoms = parseNullTime(c.FormValue("date_other_hemorrhagic_symptoms"))
	clinicalSigns.DurationOtherHemorrhagicSymptoms = parseNullInt32(c.FormValue("duration_other_hemorrhagic_symptoms"))

	if err := models.SaveVHFClinicalSigns(db, clinicalSigns); err != nil {
		sl.Error("Failed to update clinical signs", "error", err)
		return c.Status(500).SendString("Failed to update clinical signs: " + err.Error())
	}
	sl.Info("Clinical signs updated successfully")

	// --- Hospitalization ---
	sl.Info("Updating hospitalization info")
	hosp, _ := models.GetVHFHospitalization(db, patientID)
	if hosp == nil {
		hosp = &models.VHFHospitalization{PatientID: patientID}
	}
	hosp.Hospitalized = c.FormValue("hospitalized") == "true"
	hosp.AdmissionDate = parseNullTime(c.FormValue("admission_date"))
	hosp.HealthFacilityName = c.FormValue("health_facility_name")
	hosp.InIsolation = c.FormValue("in_isolation") == "true"
	hosp.IsolationDate = parseNullTime(c.FormValue("isolation_date"))
	if err := models.SaveVHFHospitalization(db, hosp); err != nil {
		sl.Error("Failed to update hospitalization", "error", err)
		return c.Status(500).SendString("Failed to update hospitalization: " + err.Error())
	}
	sl.Info("Hospitalization updated successfully")

	// --- Risk Factors ---
	sl.Info("Updating risk factors")
	risk, _ := models.GetVHFRiskFactors(db, patientID)
	if risk == nil {
		risk = &models.VHFRiskFactors{PatientID: patientID}
	}
	risk.ContactWithCase = parseNullBool(c.FormValue("contactWithCase"))
	risk.ContactName = c.FormValue("contact_name")
	risk.ContactRelation = c.FormValue("contact_relation")
	risk.ContactDates = c.FormValue("contact_dates")
	risk.ContactVillage = c.FormValue("contact_village")
	risk.ContactDistrict = c.FormValue("contact_district")
	risk.ContactStatus = c.FormValue("contact_status")
	risk.ContactDeathDate = parseNullTime(c.FormValue("contact_death_date"))
	if err := models.SaveVHFRiskFactors(db, risk); err != nil {
		sl.Error("Failed to update risk factors", "error", err)
		return c.Status(500).SendString("Failed to update risk factors: " + err.Error())
	}
	sl.Info("Risk factors updated successfully")

	// --- Investigator ---
	sl.Info("Updating investigator info")
	investigator, _ := models.GetVHFInvestigator(db, patientID)
	if investigator == nil {
		investigator = &models.VHFInvestigator{PatientID: patientID}
	}
	investigator.InvestigatorName = c.FormValue("investigator_name")
	investigator.Phone = c.FormValue("investigator_phone")
	investigator.Email = c.FormValue("investigator_email")
	investigator.Position = c.FormValue("investigator_position")
	investigator.District = c.FormValue("investigator_district")
	investigator.HealthFacility = c.FormValue("investigator_health_facility")
	if err := models.SaveVHFInvestigator(db, investigator); err != nil {
		sl.Error("Failed to update investigator", "error", err)
		return c.Status(500).SendString("Failed to update investigator: " + err.Error())
	}
	sl.Info("Investigator updated successfully")

	sl.Info("Update successful, redirecting", "patient_id", patientID)
	return c.Redirect("/vhf-cif/view/" + fmt.Sprint(patientID))
}

// --- Helper functions for parsing form values ---
//
//	func parseNullString(val string) sql.NullString {
//		return sql.NullString{String: val, Valid: val != ""}
//	}
func parseNullInt32(val string) sql.NullInt32 {
	i, err := strconv.ParseInt(val, 10, 32)
	return sql.NullInt32{Int32: int32(i), Valid: err == nil}
}

//	func parseNullFloat64(val string) sql.NullFloat64 {
//		f, err := strconv.ParseFloat(val, 64)
//		return sql.NullFloat64{Float64: f, Valid: err == nil}
//	}
//
//	func parseNullBool(val string) sql.NullBool {
//		if val == "" {
//			return sql.NullBool{Valid: false}
//		}
//		return sql.NullBool{Bool: val == "true", Valid: true}
//	}
func parseNullTime(val string) sql.NullTime {
	if val == "" {
		return sql.NullTime{Valid: false}
	}
	t, err := time.Parse("2006-01-02", val)
	return sql.NullTime{Time: t, Valid: err == nil}
}
func parseNullFloat64FromCSV(val string, idx int) sql.NullFloat64 {
	parts := strings.Split(val, ",")
	if len(parts) > idx {
		f, err := strconv.ParseFloat(strings.TrimSpace(parts[idx]), 64)
		return sql.NullFloat64{Float64: f, Valid: err == nil}
	}
	return sql.NullFloat64{Valid: false}
}
