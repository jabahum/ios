package handlers

import (
	"case/internal/models"
	"case/internal/services"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// HandlerVHFPatientSubmit handles the submission of patient information
func HandlerVHFPatientSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config, smsService *services.SMSService) error {
	// Parse form data
	patient := &models.VHFPatient{
		Surname:                     c.FormValue("surname"),
		OtherNames:                  c.FormValue("other_names"),
		DateOfBirth:                 sql.NullTime{Time: parseDate(c.FormValue("dob")), Valid: true},
		AgeYears:                    sql.NullInt32{Int32: parseInt32(c.FormValue("age_years")), Valid: true},
		AgeMonths:                   sql.NullInt32{Int32: parseInt32(c.FormValue("age_months")), Valid: true},
		Gender:                      c.FormValue("gender"),
		PatientPhone:                c.FormValue("patient_phone"),
		PhoneOwner:                  c.FormValue("phone_owner"),
		NextOfKin:                   c.FormValue("next_of_kin"),
		NextOfKinPhone:              c.FormValue("next_of_kin_phone"),
		RelationshipToPatient:       c.FormValue("relationship_to_patient"),
		DataCapturerName:            sql.NullString{String: c.FormValue("data_capturer_name"), Valid: c.FormValue("data_capturer_name") != ""},
		DataCapturerPhone:           sql.NullString{String: c.FormValue("data_capturer_phone"), Valid: c.FormValue("data_capturer_phone") != ""},
		ReportingHealthFacilityName: sql.NullString{String: c.FormValue("reporting_health_facility_name"), Valid: c.FormValue("reporting_health_facility_name") != ""},
		CaseCode:                    c.FormValue("case_code"),
		Status:                      c.FormValue("status"),
		HeadOfHousehold:             c.FormValue("head_of_household"),
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
		AdmissionDate:      sql.NullTime{Time: parseDate(c.FormValue("admission_date")), Valid: true},
		HealthFacilityName: c.FormValue("health_facility_name"),
		InIsolation:        parseBool(c.FormValue("in_isolation")),
		IsolationDate:      sql.NullTime{Time: parseDate(c.FormValue("isolation_date")), Valid: true},
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
			patient.CaseCode, patient.ReportingHealthFacilityName.String, patient.Surname, patient.OtherNames)
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
					patient.CaseCode, patient.ReportingHealthFacilityName.String, patient.District, patient.Surname, patient.OtherNames)
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
	return c.Redirect(fmt.Sprintf("/vhf-cif/success?case_code=%s", patient.CaseCode))
}

// Helper functions for parsing form values
func parseDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}
	}
	return t
}

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
func HandlerVHFClinicalSignsSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
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

	// Parse dates
	parseDate := func(dateStr string) sql.NullTime {
		var result sql.NullTime
		if dateStr != "" {
			if t, err := time.Parse("2006-01-02", dateStr); err == nil {
				result.Time = t
				result.Valid = true
			}
		}
		return result
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
		DateInitialOnset: parseDate(c.FormValue("date_initial_onset")),
		TempSource: sql.NullString{
			String: c.FormValue("temp_source"),
			Valid:  c.FormValue("temp_source") != "",
		},
		Temperature: temperature,
		Fever: sql.NullBool{
			Bool:  c.FormValue("fever") == "true",
			Valid: c.FormValue("fever") != "",
		},
		DateFever:     parseDate(c.FormValue("date_fever")),
		DurationFever: parseDuration(c.FormValue("duration_fever")),
		Vomiting: sql.NullBool{
			Bool:  c.FormValue("vomiting") == "true",
			Valid: c.FormValue("vomiting") != "",
		},
		DateVomiting:     parseDate(c.FormValue("date_vomiting")),
		DurationVomiting: parseDuration(c.FormValue("duration_vomiting")),
		Nausea: sql.NullBool{
			Bool:  c.FormValue("nausea") == "true",
			Valid: c.FormValue("nausea") != "",
		},
		DateNausea:     parseDate(c.FormValue("date_nausea")),
		DurationNausea: parseDuration(c.FormValue("duration_nausea")),
		Diarrhea: sql.NullBool{
			Bool:  c.FormValue("diarrhea") == "true",
			Valid: c.FormValue("diarrhea") != "",
		},
		DateDiarrhea:     parseDate(c.FormValue("date_diarrhea")),
		DurationDiarrhea: parseDuration(c.FormValue("duration_diarrhea")),
		IntenseFatigueGeneralWeakness: sql.NullBool{
			Bool:  c.FormValue("intense_fatigue_general_weakness") == "true",
			Valid: c.FormValue("intense_fatigue_general_weakness") != "",
		},
		DateIntenseFatigueGeneralWeakness:     parseDate(c.FormValue("date_intense_fatigue_general_weakness")),
		DurationIntenseFatigueGeneralWeakness: parseDuration(c.FormValue("duration_intense_fatigue_general_weakness")),
		EpigastricPain: sql.NullBool{
			Bool:  c.FormValue("epigastric_pain") == "true",
			Valid: c.FormValue("epigastric_pain") != "",
		},
		DateEpigastricPain:     parseDate(c.FormValue("date_epigastric_pain")),
		DurationEpigastricPain: parseDuration(c.FormValue("duration_epigastric_pain")),
		LowerAbdominalPain: sql.NullBool{
			Bool:  c.FormValue("lower_abdominal_pain") == "true",
			Valid: c.FormValue("lower_abdominal_pain") != "",
		},
		DateLowerAbdominalPain:     parseDate(c.FormValue("date_lower_abdominal_pain")),
		DurationLowerAbdominalPain: parseDuration(c.FormValue("duration_lower_abdominal_pain")),
		ChestPain: sql.NullBool{
			Bool:  c.FormValue("chest_pain") == "true",
			Valid: c.FormValue("chest_pain") != "",
		},
		DateChestPain:     parseDate(c.FormValue("date_chest_pain")),
		DurationChestPain: parseDuration(c.FormValue("duration_chest_pain")),
		MusclePain: sql.NullBool{
			Bool:  c.FormValue("muscle_pain") == "true",
			Valid: c.FormValue("muscle_pain") != "",
		},
		DateMusclePain:     parseDate(c.FormValue("date_muscle_pain")),
		DurationMusclePain: parseDuration(c.FormValue("duration_muscle_pain")),
		JointPain: sql.NullBool{
			Bool:  c.FormValue("joint_pain") == "true",
			Valid: c.FormValue("joint_pain") != "",
		},
		DateJointPain:     parseDate(c.FormValue("date_joint_pain")),
		DurationJointPain: parseDuration(c.FormValue("duration_joint_pain")),
		Headache: sql.NullBool{
			Bool:  c.FormValue("headache") == "true",
			Valid: c.FormValue("headache") != "",
		},
		DateHeadache:     parseDate(c.FormValue("date_headache")),
		DurationHeadache: parseDuration(c.FormValue("duration_headache")),
		Cough: sql.NullBool{
			Bool:  c.FormValue("cough") == "true",
			Valid: c.FormValue("cough") != "",
		},
		DateCough:     parseDate(c.FormValue("date_cough")),
		DurationCough: parseDuration(c.FormValue("duration_cough")),
		DifficultyBreathing: sql.NullBool{
			Bool:  c.FormValue("difficulty_breathing") == "true",
			Valid: c.FormValue("difficulty_breathing") != "",
		},
		DateDifficultyBreathing:     parseDate(c.FormValue("date_difficulty_breathing")),
		DurationDifficultyBreathing: parseDuration(c.FormValue("duration_difficulty_breathing")),
		DifficultySwallowing: sql.NullBool{
			Bool:  c.FormValue("difficulty_swallowing") == "true",
			Valid: c.FormValue("difficulty_swallowing") != "",
		},
		DateDifficultySwallowing:     parseDate(c.FormValue("date_difficulty_swallowing")),
		DurationDifficultySwallowing: parseDuration(c.FormValue("duration_difficulty_swallowing")),
		SoreThroat: sql.NullBool{
			Bool:  c.FormValue("sore_throat") == "true",
			Valid: c.FormValue("sore_throat") != "",
		},
		DateSoreThroat:     parseDate(c.FormValue("date_sore_throat")),
		DurationSoreThroat: parseDuration(c.FormValue("duration_sore_throat")),
		Jaundice: sql.NullBool{
			Bool:  c.FormValue("jaundice") == "true",
			Valid: c.FormValue("jaundice") != "",
		},
		DateJaundice:     parseDate(c.FormValue("date_jaundice")),
		DurationJaundice: parseDuration(c.FormValue("duration_jaundice")),
		Conjunctivitis: sql.NullBool{
			Bool:  c.FormValue("conjunctivitis") == "true",
			Valid: c.FormValue("conjunctivitis") != "",
		},
		DateConjunctivitis:     parseDate(c.FormValue("date_conjunctivitis")),
		DurationConjunctivitis: parseDuration(c.FormValue("duration_conjunctivitis")),
		SkinRash: sql.NullBool{
			Bool:  c.FormValue("skin_rash") == "true",
			Valid: c.FormValue("skin_rash") != "",
		},
		DateSkinRash:     parseDate(c.FormValue("date_skin_rash")),
		DurationSkinRash: parseDuration(c.FormValue("duration_skin_rash")),
		Hiccups: sql.NullBool{
			Bool:  c.FormValue("hiccups") == "true",
			Valid: c.FormValue("hiccups") != "",
		},
		DateHiccups:     parseDate(c.FormValue("date_hiccups")),
		DurationHiccups: parseDuration(c.FormValue("duration_hiccups")),
		PainBehindEyes: sql.NullBool{
			Bool:  c.FormValue("pain_behind_eyes") == "true",
			Valid: c.FormValue("pain_behind_eyes") != "",
		},
		DatePainBehindEyes:     parseDate(c.FormValue("date_pain_behind_eyes")),
		DurationPainBehindEyes: parseDuration(c.FormValue("duration_pain_behind_eyes")),
		SensitiveToLight: sql.NullBool{
			Bool:  c.FormValue("sensitive_to_light") == "true",
			Valid: c.FormValue("sensitive_to_light") != "",
		},
		DateSensitiveToLight:     parseDate(c.FormValue("date_sensitive_to_light")),
		DurationSensitiveToLight: parseDuration(c.FormValue("duration_sensitive_to_light")),
		ComaUnconscious: sql.NullBool{
			Bool:  c.FormValue("coma_unconscious") == "true",
			Valid: c.FormValue("coma_unconscious") != "",
		},
		DateComaUnconscious:     parseDate(c.FormValue("date_coma_unconscious")),
		DurationComaUnconscious: parseDuration(c.FormValue("duration_coma_unconscious")),
		ConfusedOrDisoriented: sql.NullBool{
			Bool:  c.FormValue("confused_or_disoriented") == "true",
			Valid: c.FormValue("confused_or_disoriented") != "",
		},
		DateConfusedOrDisoriented:     parseDate(c.FormValue("date_confused_or_disoriented")),
		DurationConfusedOrDisoriented: parseDuration(c.FormValue("duration_confused_or_disoriented")),
		Convulsions: sql.NullBool{
			Bool:  c.FormValue("convulsions") == "true",
			Valid: c.FormValue("convulsions") != "",
		},
		DateConvulsions:     parseDate(c.FormValue("date_convulsions")),
		DurationConvulsions: parseDuration(c.FormValue("duration_convulsions")),
		UnexplainedBleeding: sql.NullBool{
			Bool:  c.FormValue("unexplained_bleeding") == "true",
			Valid: c.FormValue("unexplained_bleeding") != "",
		},
		DateUnexplainedBleeding:     parseDate(c.FormValue("date_unexplained_bleeding")),
		DurationUnexplainedBleeding: parseDuration(c.FormValue("duration_unexplained_bleeding")),
		BleedingOfTheGums: sql.NullBool{
			Bool:  c.FormValue("bleeding_of_the_gums") == "true",
			Valid: c.FormValue("bleeding_of_the_gums") != "",
		},
		DateBleedingOfTheGums:     parseDate(c.FormValue("date_bleeding_of_the_gums")),
		DurationBleedingOfTheGums: parseDuration(c.FormValue("duration_bleeding_of_the_gums")),
		BleedingFromInjectionSite: sql.NullBool{
			Bool:  c.FormValue("bleeding_from_injection_site") == "true",
			Valid: c.FormValue("bleeding_from_injection_site") != "",
		},
		DateBleedingFromInjectionSite:     parseDate(c.FormValue("date_bleeding_from_injection_site")),
		DurationBleedingFromInjectionSite: parseDuration(c.FormValue("duration_bleeding_from_injection_site")),
		NoseBleedEpistaxis: sql.NullBool{
			Bool:  c.FormValue("nose_bleed_epistaxis") == "true",
			Valid: c.FormValue("nose_bleed_epistaxis") != "",
		},
		DateNoseBleedEpistaxis:     parseDate(c.FormValue("date_nose_bleed_epistaxis")),
		DurationNoseBleedEpistaxis: parseDuration(c.FormValue("duration_nose_bleed_epistaxis")),
		BloodyStool: sql.NullBool{
			Bool:  c.FormValue("bloody_stool") == "true",
			Valid: c.FormValue("bloody_stool") != "",
		},
		DateBloodyStool:     parseDate(c.FormValue("date_bloody_stool")),
		DurationBloodyStool: parseDuration(c.FormValue("duration_bloody_stool")),
		BloodInVomit: sql.NullBool{
			Bool:  c.FormValue("blood_in_vomit") == "true",
			Valid: c.FormValue("blood_in_vomit") != "",
		},
		DateBloodInVomit:     parseDate(c.FormValue("date_blood_in_vomit")),
		DurationBloodInVomit: parseDuration(c.FormValue("duration_blood_in_vomit")),
		CoughingUpBloodHemoptysis: sql.NullBool{
			Bool:  c.FormValue("coughing_up_blood_hemoptysis") == "true",
			Valid: c.FormValue("coughing_up_blood_hemoptysis") != "",
		},
		DateCoughingUpBloodHemoptysis:     parseDate(c.FormValue("date_coughing_up_blood_hemoptysis")),
		DurationCoughingUpBloodHemoptysis: parseDuration(c.FormValue("duration_coughing_up_blood_hemoptysis")),
		BleedingFromVagina: sql.NullBool{
			Bool:  c.FormValue("bleeding_from_vagina") == "true",
			Valid: c.FormValue("bleeding_from_vagina") != "",
		},
		DateBleedingFromVagina:     parseDate(c.FormValue("date_bleeding_from_vagina")),
		DurationBleedingFromVagina: parseDuration(c.FormValue("duration_bleeding_from_vagina")),
		BruisingOfTheSkin: sql.NullBool{
			Bool:  c.FormValue("bruising_of_the_skin") == "true",
			Valid: c.FormValue("bruising_of_the_skin") != "",
		},
		DateBruisingOfTheSkin:     parseDate(c.FormValue("date_bruising_of_the_skin")),
		DurationBruisingOfTheSkin: parseDuration(c.FormValue("duration_bruising_of_the_skin")),
		BloodInUrine: sql.NullBool{
			Bool:  c.FormValue("blood_in_urine") == "true",
			Valid: c.FormValue("blood_in_urine") != "",
		},
		DateBloodInUrine:     parseDate(c.FormValue("date_blood_in_urine")),
		DurationBloodInUrine: parseDuration(c.FormValue("duration_blood_in_urine")),
		OtherHemorrhagicSymptoms: sql.NullBool{
			Bool:  c.FormValue("other_hemorrhagic_symptoms") == "true",
			Valid: c.FormValue("other_hemorrhagic_symptoms") != "",
		},
		DateOtherHemorrhagicSymptoms:     parseDate(c.FormValue("date_other_hemorrhagic_symptoms")),
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
func HandlerVHFHospitalizationSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
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
func HandlerVHFRiskFactorsSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
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

// HandlerVHFLaboratorySubmit handles the submission of laboratory information
func HandlerVHFLaboratorySubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config, smsService *services.SMSService) error {
	patientID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid patient ID")
	}

	laboratory := &models.VHFLaboratory{
		PatientID:            patientID,
		SampleCollectionDate: sql.NullTime{Time: parseDate(c.FormValue("sample_collection_date")), Valid: true},
		SampleCollectionTime: sql.NullString{String: c.FormValue("sample_collection_time"), Valid: c.FormValue("sample_collection_time") != ""},
		SampleType:           sql.NullString{String: c.FormValue("sample_type"), Valid: c.FormValue("sample_type") != ""},
		OtherSampleType:      sql.NullString{String: c.FormValue("other_sample_type"), Valid: c.FormValue("other_sample_type") != ""},
		RequestedTest:        sql.NullString{String: c.FormValue("requested_test"), Valid: c.FormValue("requested_test") != ""},
		Serology:             sql.NullString{String: c.FormValue("serology"), Valid: c.FormValue("serology") != ""},
		MalariaRDT:           sql.NullString{String: c.FormValue("malaria_rdt"), Valid: c.FormValue("malaria_rdt") != ""},
		HIVRDT:               sql.NullString{String: c.FormValue("hiv_rdt"), Valid: c.FormValue("hiv_rdt") != ""},
		CreatedAt:            time.Now(),
	}

	if err := models.SaveVHFLaboratory(db, laboratory); err != nil {
		sl.Error("Failed to save laboratory data", "error", err)
		return c.Status(500).SendString("Failed to save laboratory data")
	}
	// Send SMS notification to CPHL if phone number is provided
	if laboratory.SampleType.String != "" {
		// Get patient details first
		patient, err := models.GetVHFPatient(db, patientID)
		if err != nil {
			sl.Error("Failed to get patient details for SMS", "error", err)
			return c.Status(500).SendString("Failed to get patient details")
		}

		message := fmt.Sprintf("A suspected VHF Case %s has been notified at %s and sample has been dispatched to CPHL with Case Details: %s %s",
			patient.CaseCode, patient.ReportingHealthFacilityName.String, patient.Surname, patient.OtherNames)
		// Send SMS notification
		if err := smsService.SendSMS("256783261162", message); err != nil {
			sl.Error("Failed to send SMS notification", "error", err)
		}
	}
	return c.Redirect(fmt.Sprintf("/vhf/view/%d", patientID))
}

// HandlerVHFInvestigatorSubmit handles the submission of investigator information
func HandlerVHFInvestigatorSubmit(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
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
	return c.Redirect(fmt.Sprintf("/vhf-cif/success?case_code=%s", patient.CaseCode))
}

// HandlerVHFList handles the listing of all VHF cases
func HandlerVHFList(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
	// Get current user information
	userID, _ := GetUser(c, sl, store)

	// Check if user has role ID 65 (vhf_lab_technician) and get their facility
	var facilityFilter string
	var args []interface{}

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

	// If user has role ID 65, check their facility assignment
	if roleCount > 0 {
		// Get user's facility ID from employee table
		facilityQuery := `
			SELECT e.facility 
			FROM employee e
			JOIN users u ON e.employee_email = u.email
			WHERE u.user_id = $1
			LIMIT 1
		`
		var userFacilityID sql.NullInt64
		err := db.QueryRowContext(c.Context(), facilityQuery, userID).Scan(&userFacilityID)
		if err != nil && err != sql.ErrNoRows {
			sl.Error("Failed to get user facility", "error", err)
			return c.Status(500).SendString("Failed to get user facility information")
		}

		// If user has a facility assigned, filter by that facility
		if userFacilityID.Valid && userFacilityID.Int64 > 0 {
			facilityFilter = "AND vc.facility_id = $1"
			args = append(args, userFacilityID.Int64)
			sl.Info("Filtering VHF cases by user facility", "user_id", userID, "facility_id", userFacilityID.Int64)
		} else {
			sl.Info("User has role 65 but no facility assigned, showing all VHF cases", "user_id", userID)
		}
	} else {
		sl.Info("User does not have role 65, showing all VHF cases", "user_id", userID)
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

	// Return HTML response
	data := NewTemplateData(c, store)
	data.Form = fiber.Map{
		"Title": "VHF Cases",
		"Cases": cases,
	}
	return GenerateHTML(c, db, data, "vhf_list")
}

// HandlerVHFView handles viewing a single VHF case
func HandlerVHFView(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
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

	//I will comment this out for now aka nga i want to retunr or parse this data into a template
	// Format clinical signs data
	// clinicalSignsData := fiber.Map{}
	// if clinicalSigns != nil {
	// 	clinicalSignsData = fiber.Map{
	// 		"DateInitialOnset":                      clinicalSigns.DateInitialOnset.Time.Format("2006-01-02"),
	// 		"TempSource":                            clinicalSigns.TempSource.String,
	// 		"Temperature":                           clinicalSigns.Temperature.Float64,
	// 		"Fever":                                 clinicalSigns.Fever.Bool,
	// 		"DateFever":                             clinicalSigns.DateFever.Time.Format("2006-01-02"),
	// 		"DurationFever":                         clinicalSigns.DurationFever.Int32,
	// 		"Vomiting":                              clinicalSigns.Vomiting.Bool,
	// 		"DateVomiting":                          clinicalSigns.DateVomiting.Time.Format("2006-01-02"),
	// 		"DurationVomiting":                      clinicalSigns.DurationVomiting.Int32,
	// 		"Nausea":                                clinicalSigns.Nausea.Bool,
	// 		"DateNausea":                            clinicalSigns.DateNausea.Time.Format("2006-01-02"),
	// 		"DurationNausea":                        clinicalSigns.DurationNausea.Int32,
	// 		"Diarrhea":                              clinicalSigns.Diarrhea.Bool,
	// 		"DateDiarrhea":                          clinicalSigns.DateDiarrhea.Time.Format("2006-01-02"),
	// 		"DurationDiarrhea":                      clinicalSigns.DurationDiarrhea.Int32,
	// 		"IntenseFatigueGeneralWeakness":         clinicalSigns.IntenseFatigueGeneralWeakness.Bool,
	// 		"DateIntenseFatigueGeneralWeakness":     clinicalSigns.DateIntenseFatigueGeneralWeakness.Time.Format("2006-01-02"),
	// 		"DurationIntenseFatigueGeneralWeakness": clinicalSigns.DurationIntenseFatigueGeneralWeakness.Int32,
	// 		"EpigastricPain":                        clinicalSigns.EpigastricPain.Bool,
	// 		"DateEpigastricPain":                    clinicalSigns.DateEpigastricPain.Time.Format("2006-01-02"),
	// 		"DurationEpigastricPain":                clinicalSigns.DurationEpigastricPain.Int32,
	// 		"LowerAbdominalPain":                    clinicalSigns.LowerAbdominalPain.Bool,
	// 		"DateLowerAbdominalPain":                clinicalSigns.DateLowerAbdominalPain.Time.Format("2006-01-02"),
	// 		"DurationLowerAbdominalPain":            clinicalSigns.DurationLowerAbdominalPain.Int32,
	// 		"ChestPain":                             clinicalSigns.ChestPain.Bool,
	// 		"DateChestPain":                         clinicalSigns.DateChestPain.Time.Format("2006-01-02"),
	// 		"DurationChestPain":                     clinicalSigns.DurationChestPain.Int32,
	// 		"MusclePain":                            clinicalSigns.MusclePain.Bool,
	// 		"DateMusclePain":                        clinicalSigns.DateMusclePain.Time.Format("2006-01-02"),
	// 		"DurationMusclePain":                    clinicalSigns.DurationMusclePain.Int32,
	// 		"JointPain":                             clinicalSigns.JointPain.Bool,
	// 		"DateJointPain":                         clinicalSigns.DateJointPain.Time.Format("2006-01-02"),
	// 		"DurationJointPain":                     clinicalSigns.DurationJointPain.Int32,
	// 		"Headache":                              clinicalSigns.Headache.Bool,
	// 		"DateHeadache":                          clinicalSigns.DateHeadache.Time.Format("2006-01-02"),
	// 		"DurationHeadache":                      clinicalSigns.DurationHeadache.Int32,
	// 		"Cough":                                 clinicalSigns.Cough.Bool,
	// 		"DateCough":                             clinicalSigns.DateCough.Time.Format("2006-01-02"),
	// 		"DurationCough":                         clinicalSigns.DurationCough.Int32,
	// 		"DifficultyBreathing":                   clinicalSigns.DifficultyBreathing.Bool,
	// 		"DateDifficultyBreathing":               clinicalSigns.DateDifficultyBreathing.Time.Format("2006-01-02"),
	// 		"DurationDifficultyBreathing":           clinicalSigns.DurationDifficultyBreathing.Int32,
	// 		"DifficultySwallowing":                  clinicalSigns.DifficultySwallowing.Bool,
	// 		"DateDifficultySwallowing":              clinicalSigns.DateDifficultySwallowing.Time.Format("2006-01-02"),
	// 		"DurationDifficultySwallowing":          clinicalSigns.DurationDifficultySwallowing.Int32,
	// 		"SoreThroat":                            clinicalSigns.SoreThroat.Bool,
	// 		"DateSoreThroat":                        clinicalSigns.DateSoreThroat.Time.Format("2006-01-02"),
	// 		"DurationSoreThroat":                    clinicalSigns.DurationSoreThroat.Int32,
	// 		"Jaundice":                              clinicalSigns.Jaundice.Bool,
	// 		"DateJaundice":                          clinicalSigns.DateJaundice.Time.Format("2006-01-02"),
	// 		"DurationJaundice":                      clinicalSigns.DurationJaundice.Int32,
	// 		"Conjunctivitis":                        clinicalSigns.Conjunctivitis.Bool,
	// 		"DateConjunctivitis":                    clinicalSigns.DateConjunctivitis.Time.Format("2006-01-02"),
	// 		"DurationConjunctivitis":                clinicalSigns.DurationConjunctivitis.Int32,
	// 		"SkinRash":                              clinicalSigns.SkinRash.Bool,
	// 		"DateSkinRash":                          clinicalSigns.DateSkinRash.Time.Format("2006-01-02"),
	// 		"DurationSkinRash":                      clinicalSigns.DurationSkinRash.Int32,
	// 		"Hiccups":                               clinicalSigns.Hiccups.Bool,
	// 		"DateHiccups":                           clinicalSigns.DateHiccups.Time.Format("2006-01-02"),
	// 		"DurationHiccups":                       clinicalSigns.DurationHiccups.Int32,
	// 		"PainBehindEyes":                        clinicalSigns.PainBehindEyes.Bool,
	// 		"DatePainBehindEyes":                    clinicalSigns.DatePainBehindEyes.Time.Format("2006-01-02"),
	// 		"DurationPainBehindEyes":                clinicalSigns.DurationPainBehindEyes.Int32,
	// 		"SensitiveToLight":                      clinicalSigns.SensitiveToLight.Bool,
	// 		"DateSensitiveToLight":                  clinicalSigns.DateSensitiveToLight.Time.Format("2006-01-02"),
	// 		"DurationSensitiveToLight":              clinicalSigns.DurationSensitiveToLight.Int32,
	// 		"ComaUnconscious":                       clinicalSigns.ComaUnconscious.Bool,
	// 		"DateComaUnconscious":                   clinicalSigns.DateComaUnconscious.Time.Format("2006-01-02"),
	// 		"DurationComaUnconscious":               clinicalSigns.DurationComaUnconscious.Int32,
	// 		"ConfusedOrDisoriented":                 clinicalSigns.ConfusedOrDisoriented.Bool,
	// 		"DateConfusedOrDisoriented":             clinicalSigns.DateConfusedOrDisoriented.Time.Format("2006-01-02"),
	// 		"DurationConfusedOrDisoriented":         clinicalSigns.DurationConfusedOrDisoriented.Int32,
	// 		"Convulsions":                           clinicalSigns.Convulsions.Bool,
	// 		"DateConvulsions":                       clinicalSigns.DateConvulsions.Time.Format("2006-01-02"),
	// 		"DurationConvulsions":                   clinicalSigns.DurationConvulsions.Int32,
	// 		"UnexplainedBleeding":                   clinicalSigns.UnexplainedBleeding.Bool,
	// 		"DateUnexplainedBleeding":               clinicalSigns.DateUnexplainedBleeding.Time.Format("2006-01-02"),
	// 		"DurationUnexplainedBleeding":           clinicalSigns.DurationUnexplainedBleeding.Int32,
	// 		"BleedingOfTheGums":                     clinicalSigns.BleedingOfTheGums.Bool,
	// 		"DateBleedingOfTheGums":                 clinicalSigns.DateBleedingOfTheGums.Time.Format("2006-01-02"),
	// 		"DurationBleedingOfTheGums":             clinicalSigns.DurationBleedingOfTheGums.Int32,
	// 		"BleedingFromInjectionSite":             clinicalSigns.BleedingFromInjectionSite.Bool,
	// 		"DateBleedingFromInjectionSite":         clinicalSigns.DateBleedingFromInjectionSite.Time.Format("2006-01-02"),
	// 		"DurationBleedingFromInjectionSite":     clinicalSigns.DurationBleedingFromInjectionSite.Int32,
	// 		"NoseBleedEpistaxis":                    clinicalSigns.NoseBleedEpistaxis.Bool,
	// 		"DateNoseBleedEpistaxis":                clinicalSigns.DateNoseBleedEpistaxis.Time.Format("2006-01-02"),
	// 		"DurationNoseBleedEpistaxis":            clinicalSigns.DurationNoseBleedEpistaxis.Int32,
	// 		"BloodyStool":                           clinicalSigns.BloodyStool.Bool,
	// 		"DateBloodyStool":                       clinicalSigns.DateBloodyStool.Time.Format("2006-01-02"),
	// 		"DurationBloodyStool":                   clinicalSigns.DurationBloodyStool.Int32,
	// 		"BloodInVomit":                          clinicalSigns.BloodInVomit.Bool,
	// 		"DateBloodInVomit":                      clinicalSigns.DateBloodInVomit.Time.Format("2006-01-02"),
	// 		"DurationBloodInVomit":                  clinicalSigns.DurationBloodInVomit.Int32,
	// 		"CoughingUpBloodHemoptysis":             clinicalSigns.CoughingUpBloodHemoptysis.Bool,
	// 		"DateCoughingUpBloodHemoptysis":         clinicalSigns.DateCoughingUpBloodHemoptysis.Time.Format("2006-01-02"),
	// 		"DurationCoughingUpBloodHemoptysis":     clinicalSigns.DurationCoughingUpBloodHemoptysis.Int32,
	// 		"BleedingFromVagina":                    clinicalSigns.BleedingFromVagina.Bool,
	// 		"DateBleedingFromVagina":                clinicalSigns.DateBleedingFromVagina.Time.Format("2006-01-02"),
	// 		"DurationBleedingFromVagina":            clinicalSigns.DurationBleedingFromVagina.Int32,
	// 		"BruisingOfTheSkin":                     clinicalSigns.BruisingOfTheSkin.Bool,
	// 		"DateBruisingOfTheSkin":                 clinicalSigns.DateBruisingOfTheSkin.Time.Format("2006-01-02"),
	// 		"DurationBruisingOfTheSkin":             clinicalSigns.DurationBruisingOfTheSkin.Int32,
	// 		"BloodInUrine":                          clinicalSigns.BloodInUrine.Bool,
	// 		"DateBloodInUrine":                      clinicalSigns.DateBloodInUrine.Time.Format("2006-01-02"),
	// 		"DurationBloodInUrine":                  clinicalSigns.DurationBloodInUrine.Int32,
	// 		"OtherHemorrhagicSymptoms":              clinicalSigns.OtherHemorrhagicSymptoms.Bool,
	// 		"DateOtherHemorrhagicSymptoms":          clinicalSigns.DateOtherHemorrhagicSymptoms.Time.Format("2006-01-02"),
	// 		"DurationOtherHemorrhagicSymptoms":      clinicalSigns.DurationOtherHemorrhagicSymptoms.Int32,
	// 	}
	// }

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
	data.Form = fiber.Map{
		"Title":           "Update VHF Case Investigation Form",
		"Case":            patient,
		"Lab":             labData,
		"ClinicalSigns":   clinicalSigns,
		"Hospitalization": hospitalization,
		"RiskFactors":     riskFactors,
		"Investigator":    investigatorData,
	}
	return GenerateHTML(c, db, data, "update_vhf_cif")
}

// HandlerVHFSuccess handles the success page after form submission
func HandlerVHFSuccess(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
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
func HandlerVHFLabForm(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
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

	// Return HTML response with patient and lab data
	data := NewTemplateData(c, store)
	data.Form = fiber.Map{
		"Title": "VHF Lab Form",
		"Patient": fiber.Map{
			"ID":           patient.ID,
			"CaseCode":     patient.CaseCode,
			"Surname":      patient.Surname,
			"OtherNames":   patient.OtherNames,
			"PatientPhone": patient.PatientPhone,
		},
		"Lab": fiber.Map{
			"SampleCollectionDate": lab.SampleCollectionDate,
			"SampleCollectionTime": lab.SampleCollectionTime,
			"SampleType":           lab.SampleType,
			"OtherSampleType":      lab.OtherSampleType,
			"RequestedTest":        lab.RequestedTest,
			"Serology":             lab.Serology,
			"MalariaRDT":           lab.MalariaRDT,
			"HIVRDT":               lab.HIVRDT,
			"TestResult":           lab.TestResult,
			"DateTested":           lab.DateTested,
			"LabName":              lab.LabName,
		},
	}
	return GenerateHTML(c, db, data, "vhf_lab_form")
}

// HandlerVHFLabSave handles the submission of laboratory information
func HandlerVHFLabSave(c *fiber.Ctx, db *sql.DB, sl *slog.Logger, store *session.Store, config Config) error {
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
