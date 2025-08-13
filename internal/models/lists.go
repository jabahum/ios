package models

import (
	"context"
	"database/sql"
	"log"
	"strconv"
)

func (c *Client) SetAsExists() {
	c._exists = true
}

func (c *User) SetAsExists() {
	c._exists = true
}

func (c *UserRight) SetAsExists() {
	c._exists = true
}

func (c *Encounter) SetAsExists() {
	c._exists = true
}

func (c *Clinical) SetAsExists() {
	c._exists = true
}

func (c *Vital) SetAsExists() {
	c._exists = true
}

func (c *Lab) SetAsExists() {
	c._exists = true
}

func (c *Treatment) SetAsExists() {
	c._exists = true
}

func (c *Employee) SetAsExists() {
	c._exists = true
}

func (c *Status) SetAsExists() {
	c._exists = true
}

func (c *Discharge) SetAsExists() {
	c._exists = true
}

type ClientEncounter struct {
	EncounterID   int
	EncounterType sql.NullInt64
	EmployeeFname sql.NullString
	EmployeeLname sql.NullString
	EncounterDate sql.NullString
	EncounterTime sql.NullString
	ClinicalTeam  sql.NullString
	ManagedBy     sql.NullInt64
	ClientID      int
}

type Statuz struct {
	ZeStatus Status
	FName    sql.NullString
	LName    sql.NullString
}

func Statusez(ctx context.Context, db DB, flt string) ([]Statuz, error) {
	// query
	sqlstr := ` SELECT 
					s.status_id, s.client_id, s.status_date, s.status, s.status_notes, s.updated_by, s.updated_on,
					e.employee_fname, e.employee_lname
				FROM status s left join employee e on s.updated_by = e.employee_id `

	var args []interface{}
	if flt != "" {
		sqlstr += " WHERE " + flt
	}

	// Log the query
	logf(sqlstr)

	// Execute query
	rows, err := db.QueryContext(ctx, sqlstr, args...)
	if err != nil {
		return nil, logerror(err)
	}
	defer rows.Close()

	// Slice to hold clients
	var statuses []Statuz

	// Iterate through rows
	for rows.Next() {
		var s Statuz
		if err := rows.Scan(
			&s.ZeStatus.StatusID, &s.ZeStatus.ClientID, &s.ZeStatus.StatusDate, &s.ZeStatus.Status, &s.ZeStatus.StatusNotes, &s.ZeStatus.UpdatedBy, &s.ZeStatus.UpdatedOn,
			&s.FName, &s.LName,
		); err != nil {
			return nil, logerror(err)
		}
		statuses = append(statuses, s)
	}

	// Check for iteration errors
	if err := rows.Err(); err != nil {
		return nil, logerror(err)
	}

	return statuses, nil

}

func DischargeByClientID(ctx context.Context, db DB, clientID int) (*Discharge, error) {
	// query
	const sqlstr = `SELECT ` +
		`discharge_id, client_id, discharge_date, final_diagnosis, final_diagnosis_other, discharge_outcome, discharge_seq_heari, discharge_seq_pregn, discharge_seq_ocula, discharge_seq_extre, discharge_seq_arthr, discharge_seq_neuro, discharge_seq_others, counselling_provided, discharging_officer, discharge_facility, discharge_seq_others_aza, entered_on, entered_by, updated_by, updated_on ` +
		`FROM public.discharge ` +
		`WHERE client_id = $1`
	// run
	logf(sqlstr, clientID)
	d := Discharge{
		_exists: true,
	}
	if err := db.QueryRowContext(ctx, sqlstr, clientID).Scan(&d.DischargeID, &d.ClientID, &d.DischargeDate, &d.FinalDiagnosis, &d.FinalDiagnosisOther, &d.DischargeOutcome, &d.DischargeSeqHeari, &d.DischargeSeqPregn, &d.DischargeSeqOcula, &d.DischargeSeqExtre, &d.DischargeSeqArthr, &d.DischargeSeqNeuro, &d.DischargeSeqOthers, &d.CounsellingProvided, &d.DischargingOfficer, &d.DischargeFacility, &d.DischargeSeqOthersAza, &d.EnteredOn, &d.EnteredBy, &d.UpdatedBy, &d.UpdatedOn); err != nil {
		return nil, logerror(err)
	}
	return &d, nil
}

func Statuses(ctx context.Context, db DB, flt string) ([]Status, error) {
	// query
	sqlstr := ` SELECT 
					status_id, client_id, status_date, status, status_notes, updated_by, updated_on
				FROM status `

	var args []interface{}
	if flt != "" {
		sqlstr += " WHERE " + flt
	}

	// Log the query
	logf(sqlstr)

	// Execute query
	rows, err := db.QueryContext(ctx, sqlstr, args...)
	if err != nil {
		return nil, logerror(err)
	}
	defer rows.Close()

	// Slice to hold clients
	var statuses []Status

	// Iterate through rows
	for rows.Next() {
		var s Status
		if err := rows.Scan(
			&s.StatusID, &s.ClientID, &s.StatusDate, &s.Status, &s.StatusNotes, &s.UpdatedBy, &s.UpdatedOn,
		); err != nil {
			return nil, logerror(err)
		}
		statuses = append(statuses, s)
	}

	// Check for iteration errors
	if err := rows.Err(); err != nil {
		return nil, logerror(err)
	}

	return statuses, nil

}

func ClientEncounters(ctx context.Context, db DB, flt string, outbreakID int) ([]ClientEncounter, error) {
	// query - handle both specific outbreak_id and NULL outbreak_id
	sqlstr := ` SELECT 
					encounter.encounter_id, encounter.encounter_type, employee.employee_fname, employee.employee_lname, encounter.encounter_date, encounter.encounter_time, encounter.client_id, encounter.clinical_team, encounter.managed_by
				FROM encounter 
				LEFT JOIN employee on employee.employee_id = encounter.managed_by 
				WHERE (encounter.outbreak_id = $1 OR encounter.outbreak_id IS NULL)`
	var args []interface{}
	args = append(args, outbreakID)
	if flt != "" {
		sqlstr += " AND " + flt
	}

	// Log the query with more details
	log.Printf("ClientEncounters query: %s with args: %v", sqlstr, args)
	logf(sqlstr)

	// Execute query
	rows, err := db.QueryContext(ctx, sqlstr, args...)
	if err != nil {
		return nil, logerror(err)
	}
	defer rows.Close()

	// Slice to hold clients
	var clientencounters []ClientEncounter

	// Iterate through rows
	for rows.Next() {
		var e ClientEncounter
		if err := rows.Scan(
			&e.EncounterID, &e.EncounterType, &e.EmployeeFname, &e.EmployeeLname, &e.EncounterDate, &e.EncounterTime, &e.ClientID, &e.ClinicalTeam, &e.ManagedBy,
		); err != nil {
			return nil, logerror(err)
		}
		clientencounters = append(clientencounters, e)
	}

	// Check for iteration errors
	if err := rows.Err(); err != nil {
		return nil, logerror(err)
	}

	return clientencounters, nil
}

func ClientEncounterz(ctx context.Context, db DB, flt string, outbreakID int) ([]ClientEncounter, error) {
	// query - handle both specific outbreak_id and NULL outbreak_id
	sqlstr := ` SELECT DISTINCT 
					employee_fname, employee_lname, encounter_date, client_id, clinical_team
				FROM encounter 
				LEFT JOIN employee ON employee.employee_id = encounter.managed_by 
				WHERE (encounter.outbreak_id = $1 OR encounter.outbreak_id IS NULL)`
	var args []interface{}
	args = append(args, outbreakID)
	if flt != "" {
		sqlstr += " AND " + flt
	}

	// Log the query
	logf(sqlstr)

	// Execute query
	rows, err := db.QueryContext(ctx, sqlstr, args...)
	if err != nil {
		return nil, logerror(err)
	}
	defer rows.Close()

	// Slice to hold clients
	var clientencounters []ClientEncounter

	// Iterate through rows
	for rows.Next() {
		var e ClientEncounter
		if err := rows.Scan(
			&e.EmployeeFname, &e.EmployeeLname, &e.EncounterDate, &e.ClientID, &e.ClinicalTeam,
		); err != nil {
			return nil, logerror(err)
		}
		clientencounters = append(clientencounters, e)
	}

	// Check for iteration errors
	if err := rows.Err(); err != nil {
		return nil, logerror(err)
	}

	return clientencounters, nil
}

func ClinicalByEncounterID(ctx context.Context, db DB, encounterID int) (*Clinical, error) {
	// query
	const sqlstr = `SELECT ` +
		`clinical_id, encounter_id, fever, fatigue, weakness, malaise, myalgia, anorexia, sore_throat, headache, nausea, chest_pain, joint_pain, hiccups, cough, difficulty_breathing, difficulty_swallowing, abdominal_pain, diarrhoea, vomiting, irritability, dysphagia, unusual_bleeding, dehydration, shock, anuria, disorientation, agitation, seizure, meningitis, confusion, coma, bacteraemia, hyperglycemia, hypoglycemia, other_complications, aza_complications_specif, pharyngeal_erythema, pharyngeal_exudate, conjunctival_injection, oedema_face, tender_abdomen, sunken_eyes, tenting_skin, palpable_liver, palpable_spleen, jaundice, enlarged_lymph_nodes, lower_extremity_oedema, bleeding, bleeding_nose, bleeding_mouth, bleeding_vagina, bleeding_rectum, bleeding_sputum, bleeding_urine, bleeding_iv_site, bleeding_other, bleeding_other_specif ` +
		`FROM public.clinical ` +
		`WHERE encounter_id = $1`
	// run
	logf(sqlstr, encounterID)
	c := Clinical{
		_exists: true,
	}
	if err := db.QueryRowContext(ctx, sqlstr, encounterID).Scan(&c.ClinicalID, &c.EncounterID, &c.Fever, &c.Fatigue, &c.Weakness, &c.Malaise, &c.Myalgia, &c.Anorexia, &c.SoreThroat, &c.Headache, &c.Nausea, &c.ChestPain, &c.JointPain, &c.Hiccups, &c.Cough, &c.DifficultyBreathing, &c.DifficultySwallowing, &c.AbdominalPain, &c.Diarrhoea, &c.Vomiting, &c.Irritability, &c.Dysphagia, &c.UnusualBleeding, &c.Dehydration, &c.Shock, &c.Anuria, &c.Disorientation, &c.Agitation, &c.Seizure, &c.Meningitis, &c.Confusion, &c.Coma, &c.Bacteraemia, &c.Hyperglycemia, &c.Hypoglycemia, &c.OtherComplications, &c.AzaComplicationsSpecif, &c.PharyngealErythema, &c.PharyngealExudate, &c.ConjunctivalInjection, &c.OedemaFace, &c.TenderAbdomen, &c.SunkenEyes, &c.TentingSkin, &c.PalpableLiver, &c.PalpableSpleen, &c.Jaundice, &c.EnlargedLymphNodes, &c.LowerExtremityOedema, &c.Bleeding, &c.BleedingNose, &c.BleedingMouth, &c.BleedingVagina, &c.BleedingRectum, &c.BleedingSputum, &c.BleedingUrine, &c.BleedingIvSite, &c.BleedingOther, &c.BleedingOtherSpecif); err != nil {
		return nil, logerror(err)
	}
	return &c, nil
}

func VitalByEncounterID(ctx context.Context, db DB, encounterID int) (*Vital, error) {
	// query
	const sqlstr = `SELECT ` +
		`vitals_id, encounter_id, heart_rate, bp_systolic, bp_diastolic, capillary_refill, respiratory_rate, saturation, weight, height, temperature, lowest_consciousness, mental_status, muac, bleeding, shock, meningitis, confusion, seizure, coma, bacteraemia, hyperglycemia, hypoglycemia, other ` +
		`FROM public.vitals ` +
		`WHERE encounter_id = $1`
	// run
	logf(sqlstr, encounterID)
	v := Vital{
		_exists: true,
	}
	if err := db.QueryRowContext(ctx, sqlstr, encounterID).Scan(&v.VitalsID, &v.EncounterID, &v.HeartRate, &v.BpSystolic, &v.BpDiastolic, &v.CapillaryRefill, &v.RespiratoryRate, &v.Saturation, &v.Weight, &v.Height, &v.Temperature, &v.LowestConsciousness, &v.MentalStatus, &v.Muac, &v.Bleeding, &v.Shock, &v.Meningitis, &v.Confusion, &v.Seizure, &v.Coma, &v.Bacteraemia, &v.Hyperglycemia, &v.Hypoglycemia, &v.Other); err != nil {
		return nil, logerror(err)
	}
	return &v, nil
}

func LabByEncounterID(ctx context.Context, db DB, encounterID int) (*Lab, error) {
	// query
	const sqlstr = `SELECT ` +
		`lab_id, encounter_id, specimen, sample_blood, sample_urine, sample_swab, sample_aza, ebola_rdt, ebola_rdt_date, ebola_rdt_results, ebola_pcr, ebola_pcr_aza, ebola_pcr_date, ebola_pcr_gp, ebola_pcr_gp_ct, ebola_pcr_np, ebola_pcr_np_ct, ebola_pcr_indeterminate, malaria_rdt_date, malaria_rdt_result, blood_culture_date, blood_culture_result, test_pos_infection, test_pos_infection_aza, haemoglobinuria, proteinuria, hematuria, blood_gas, ph, pco2, pao2, hco3, oxygen_therapy, alt_sgpt, ast_sgo, creatinine, potassium, urea, creatinine_kinase, calcium, sodium, alt_sgpt_nd, ast_sgo_nd, creatinine_nd, potassium_nd, urea_nd, creatinine_kinase_nd, calcium_nd, sodium_nd, glucose, lactate, haemoglobin, total_bilirubin, wbc_count, platelets, pt, aptt, glucose_nd, lactate_nd, haemoglobin_nd, total_bilirubin_nd, wbc_count_nd, platelets_nd, pt_nd, aptt_nd, ebola_rdt_aza ` +
		`FROM public.lab ` +
		`WHERE encounter_id = $1`
	// run
	logf(sqlstr, encounterID)
	l := Lab{
		_exists: true,
	}
	if err := db.QueryRowContext(ctx, sqlstr, encounterID).Scan(&l.LabID, &l.EncounterID, &l.Specimen, &l.SampleBlood, &l.SampleUrine, &l.SampleSwab, &l.SampleAza, &l.EbolaRdt, &l.EbolaRdtDate, &l.EbolaRdtResults, &l.EbolaPcr, &l.EbolaPcrAza, &l.EbolaPcrDate, &l.EbolaPcrGp, &l.EbolaPcrGpCt, &l.EbolaPcrNp, &l.EbolaPcrNpCt, &l.EbolaPcrIndeterminate, &l.MalariaRdtDate, &l.MalariaRdtResult, &l.BloodCultureDate, &l.BloodCultureResult, &l.TestPosInfection, &l.TestPosInfectionAza, &l.Haemoglobinuria, &l.Proteinuria, &l.Hematuria, &l.BloodGas, &l.Ph, &l.Pco2, &l.Pao2, &l.Hco3, &l.OxygenTherapy, &l.AltSgpt, &l.AstSgo, &l.Creatinine, &l.Potassium, &l.Urea, &l.CreatinineKinase, &l.Calcium, &l.Sodium, &l.AltSgptNd, &l.AstSgoNd, &l.CreatinineNd, &l.PotassiumNd, &l.UreaNd, &l.CreatinineKinaseNd, &l.CalciumNd, &l.SodiumNd, &l.Glucose, &l.Lactate, &l.Haemoglobin, &l.TotalBilirubin, &l.WbcCount, &l.Platelets, &l.Pt, &l.Aptt, &l.GlucoseNd, &l.LactateNd, &l.HaemoglobinNd, &l.TotalBilirubinNd, &l.WbcCountNd, &l.PlateletsNd, &l.PtNd, &l.ApttNd, &l.EbolaRdtAza); err != nil {
		return nil, logerror(err)
	}
	return &l, nil
}

func TreatmentByEncounterID(ctx context.Context, db DB, encounterID int) (*Treatment, error) {
	// query
	const sqlstr = `SELECT treatment_id FROM public.treatment WHERE encounter_id = $1`
	// run
	logf(sqlstr, encounterID)
	t := Treatment{
		_exists: true,
	}
	if err := db.QueryRowContext(ctx, sqlstr, encounterID).Scan(&t.TreatmentID); err != nil {
		return nil, logerror(err)
	}
	return &t, nil
}

func (u *User) Update_NoPass(ctx context.Context, db DB) error {

	// update with composite primary key
	const sqlstr = `UPDATE public.users SET ` +
		`user_name = $1, user_employee = $2 ` +
		`WHERE user_id = $3`
	// run
	logf(sqlstr, u.UserName, u.UserEmployee, u.UserID)
	if _, err := db.ExecContext(ctx, sqlstr, u.UserName, u.UserEmployee, u.UserID); err != nil {
		return logerror(err)
	}
	return nil
}

func (u *User) Update_Pass(ctx context.Context, db DB) error {

	// update with composite primary key
	const sqlstr = `UPDATE public.users SET ` +
		`user_pass = $1 ` +
		`WHERE user_id = $2`
	// run
	logf(sqlstr, u.UserPass, u.UserID)
	if _, err := db.ExecContext(ctx, sqlstr, u.UserPass, u.UserID); err != nil {
		return logerror(err)
	}
	return nil
}

func GetFields(ctx context.Context, db DB, sql_statement string) (map[int][]string, error) {
	var args []interface{}
	// Log the query
	logf(sql_statement)

	// Execute query
	rows, err := db.QueryContext(ctx, sql_statement, args...)
	if err != nil {
		return nil, logerror(err)
	}
	defer rows.Close()

	zaResults := make(map[int][]string)
	var i, id int
	var labs string

	for rows.Next() {
		if err := rows.Scan(&id, &labs); err != nil {
			log.Println("Error scanning row:", err)
			continue
		}

		// Append to the map
		zaResults[i] = []string{strconv.Itoa(id), labs}
		i++
	}

	return zaResults, nil
}

func Clients(ctx context.Context, db DB, flt string) ([]Client, error) {
	// Base SQL query
	sqlstr := `SELECT 
		id, uuid, firstname, lastname, othername, gender, date_of_birth, age, marital, nin, nationality, adm_date, adm_from, lab_no, cif_no, etu_no, case_no, occupation, occupation_aza, date_symptom_onset, date_isolation, pregnant, adm_ward, tb, asplenia, hep, diabetes, hiv, liver, malignancy, heart, pulmonary, kidney, neurologic, other, status, enter_on, enter_by, edit_on, edit_by, transfer, site 
	FROM public.clients`

	// Add filter condition if `flt` is not empty
	var args []interface{}
	if flt != "" {
		sqlstr += " WHERE " + flt
	}

	// Log the query
	logf(sqlstr)

	// Execute query
	rows, err := db.QueryContext(ctx, sqlstr, args...)
	if err != nil {
		return nil, logerror(err)
	}
	defer rows.Close()

	// Slice to hold clients
	var clients []Client

	// Iterate through rows
	for rows.Next() {
		var c Client
		c._exists = true
		if err := rows.Scan(
			&c.ID, &c.UUID, &c.Firstname, &c.Lastname, &c.Othername, &c.Gender, &c.DateOfBirth, &c.Age, &c.Marital, &c.Nin, &c.Nationality, &c.AdmDate, &c.AdmFrom, &c.LabNo, &c.CifNo, &c.EtuNo, &c.CaseNo, &c.Occupation, &c.OccupationAza, &c.DateSymptomOnset, &c.DateIsolation, &c.Pregnant, &c.AdmWard, &c.Tb, &c.Asplenia, &c.Hep, &c.Diabetes, &c.Hiv, &c.Liver, &c.Malignancy, &c.Heart, &c.Pulmonary, &c.Kidney, &c.Neurologic, &c.Other, &c.Status, &c.EnterOn, &c.EnterBy, &c.EditOn, &c.EditBy, &c.Transfer, &c.Site,
		); err != nil {
			return nil, logerror(err)
		}
		clients = append(clients, c)
	}

	// Check for iteration errors
	if err := rows.Err(); err != nil {
		return nil, logerror(err)
	}

	return clients, nil
}

func Users(ctx context.Context, db DB, flt string) ([]User, error) {
	// Base SQL query
	sqlstr := `SELECT user_id, user_name, user_pass, user_employee FROM public.users`

	// Add filter condition if `flt` is not empty
	var args []interface{}
	if flt != "" {
		sqlstr += " WHERE " + flt
	}

	// Log the query
	logf(sqlstr)

	// Execute query
	rows, err := db.QueryContext(ctx, sqlstr, args...)
	if err != nil {
		return nil, logerror(err)
	}
	defer rows.Close()

	// Slice to hold clients
	var users []User

	// Iterate through rows
	for rows.Next() {
		var u User
		u._exists = true
		if err := rows.Scan(
			&u.UserID, &u.UserName, &u.UserPass, &u.UserEmployee,
		); err != nil {
			return nil, logerror(err)
		}

		users = append(users, u)
	}

	// Check for iteration errors
	if err := rows.Err(); err != nil {
		return nil, logerror(err)
	}

	return users, nil
}

type Metumx struct {
	MetaID          int            `json:"meta_id"`          // meta_id
	MetaCategory    sql.NullInt64  `json:"meta_category"`    // meta_category
	MetaName        sql.NullString `json:"meta_name"`        // meta_name
	MetaOrder       sql.NullInt64  `json:"meta_order"`       // meta_order
	MetaDescription sql.NullString `json:"meta_description"` // meta_description
	MetaLink        sql.NullString `json:"meta_link"`        // meta_link
	Scope           sql.NullInt64  `json:"function_scope"`   // meta_order
	UserID          sql.NullInt64  `json:"user_id"`          // meta_order
	// xo fields
	_exists, _deleted bool
}

func Metums(ctx context.Context, db DB, flt string) ([]Metumx, error) {
	// Base SQL query
	sqlstr := `SELECT meta_id, meta_category, meta_name, meta_order, meta_description, meta_link, f.function_scope, f.user_id 
	           FROM public.meta,
				( 
					Select user_id,function_scope , user_rights_function, 
						user_rights_can_create+user_rights_can_view+user_rights_can_edit+user_rights_can_remove as func
					From public.user_right
					WHERE user_rights_can_create+user_rights_can_view+user_rights_can_edit+user_rights_can_remove > 0 
				) f
				Where f.user_rights_function = meta_id`

	// Add filter condition if `flt` is not empty
	var args []interface{}
	if flt != "" {
		sqlstr += " AND " + flt
	}

	// Log the query
	logf(sqlstr)

	// Execute query
	rows, err := db.QueryContext(ctx, sqlstr, args...)
	if err != nil {
		return nil, logerror(err)
	}
	defer rows.Close()

	// Slice to hold clients
	var metums []Metumx

	// Iterate through rows
	for rows.Next() {
		var m Metumx
		m._exists = true
		if err := rows.Scan(
			&m.MetaID, &m.MetaCategory, &m.MetaName, &m.MetaOrder, &m.MetaDescription, &m.MetaLink, &m.Scope, &m.UserID,
		); err != nil {
			return nil, logerror(err)
		}

		metums = append(metums, m)
	}

	// Check for iteration errors
	if err := rows.Err(); err != nil {
		return nil, logerror(err)
	}

	return metums, nil
}
