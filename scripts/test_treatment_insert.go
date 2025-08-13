package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// Connect to database
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/ios?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Test the exact INSERT statement from the model
	query := `INSERT INTO public.treatment (
		encounter_id, antibacterial, amoxicillin, ceftriaxone, cefixime, ampicillin, chloramphenicol, amoxiclav, azithromycin, cefotaxime, ceftazidime, ciprofloxacin, tetracycline, imipenem, cotrimoxazole, gentamicin, metronidazole, antibacterial_other, antibacterial_dose, antibacterial_route, antibacterial_freq, antimalarial, antimalarial_artesunate, antimalarial_arthemeter, antimalarial_al, antimalarial_aa, antimalarial_dose, antimalarial_route, antimalarial_freq, other_meds_specify, other_meds_dose, other_meds_route, other_meds_freq, ebola_experimental, ebola_experimental_if, oral, oral_ors, oral_ors_qty, oral_water, oral_water_qty, oral_other, oral_other_qty, iv, iv_qty, iv_using, iv_aza, access_type, blood_trans, oxygen_therapy, oxygen_qty, oxygen_with, vasopressors, renal, invasive, ebola_rdt_aza, ebola_experimental_if_zmap, ebola_experimental_if_remd, ebola_experimental_if_regn, ebola_experimental_if_favi, ebola_experimental_if_mab, oral_other_aza, antibacterial_aza, antimalarial_artesunate_dose, antimalarial_artesunate_route, antimalarial_artesunate_freq, antimalarial_arthemeter_dose, antimalarial_arthemeter_route, antimalarial_arthemeter_freq, antimalarial_al_dose, antimalarial_al_route, antimalarial_al_freq, antimalarial_aa_dose, antimalarial_aa_route, antimalarial_aa_freq, amoxicillin_dose, amoxicillin_route, amoxicillin_freq, ceftriaxone_dose, ceftriaxone_route, ceftriaxone_freq, cefixime_dose, cefixime_route, cefixime_freq, ampicillin_dose, ampicillin_route, ampicillin_freq, chloramphenicol_dose, chloramphenicol_route, chloramphenicol_freq, amoxiclav_dose, amoxiclav_route, amoxiclav_freq, azithromycin_dose, azithromycin_route, azithromycin_freq, cefotaxime_dose, cefotaxime_route, cefotaxime_freq, ceftazidime_dose, ceftazidime_route, ceftazidime_freq, ciprofloxacin_dose, ciprofloxacin_route, ciprofloxacin_freq, tetracycline_dose, tetracycline_route, tetracycline_freq, imipenem_dose, imipenem_route, imipenem_freq, cotrimoxazole_dose, cotrimoxazole_route, cotrimoxazole_freq, gentamicin_dose, gentamicin_route, gentamicin_freq, metronidazole_dose, metronidazole_route, metronidazole_freq
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, $42, $43, $44, $45, $46, $47, $48, $49, $50, $51, $52, $53, $54, $55, $56, $57, $58, $59, $60, $61, $62, $63, $64, $65, $66, $67, $68, $69, $70, $71, $72, $73, $74, $75, $76, $77, $78, $79, $80, $81, $82, $83, $84, $85, $86, $87, $88, $89, $90, $91, $92, $93, $94, $95, $96, $97, $98, $99, $100, $101, $102, $103, $104, $105, $106, $107, $108
	) RETURNING treatment_id`

	// Count the columns in the INSERT statement
	columnCount := 108
	fmt.Printf("INSERT statement has %d columns\n", columnCount)

	// Try to execute the query with dummy values
	args := make([]interface{}, columnCount)
	for i := range args {
		args[i] = nil
	}

	ctx := context.Background()
	var treatmentID int
	err = db.QueryRowContext(ctx, query, args...).Scan(&treatmentID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Success! Treatment ID: %d\n", treatmentID)
	}
}
