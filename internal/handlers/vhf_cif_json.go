package handlers

import (
	"database/sql"

	"case/internal/models"

	"github.com/gofiber/fiber/v2"
)

// BuildVHFCIFJSON returns the same payload as /api/vhf/cif/:id (full CIF bundle).
func BuildVHFCIFJSON(db *sql.DB, id int64) (fiber.Map, error) {
	patient, err := models.GetVHFPatient(db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	signs, _ := models.GetVHFClinicalSigns(db, id)
	hosp, _ := models.GetVHFHospitalization(db, id)
	risk, _ := models.GetVHFRiskFactors(db, id)
	lab, _ := models.GetVHFLaboratory(db, id)
	inv, _ := models.GetVHFInvestigator(db, id)
	return fiber.Map{
		"patient":         patient,
		"clinical_signs":  signs,
		"hospitalization": hosp,
		"risk_factors":    risk,
		"laboratory":      lab,
		"investigator":    inv,
	}, nil
}
