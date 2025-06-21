package models

import (
	"database/sql"
	"time"
)

type SurveillanceFocalPerson struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Phone      string    `json:"phone"`
	DistrictID string    `json:"district_id"`
	Email      string    `json:"email"`
	Position   string    `json:"position"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// GetFocalPersonByDistrict returns the active focal person for a given district
func GetFocalPersonByDistrict(db *sql.DB, districtID int) (*SurveillanceFocalPerson, error) {
	focalPerson := &SurveillanceFocalPerson{}
	query := `
		SELECT id, name, phone, district_id, email, position, is_active, created_at, updated_at
		FROM surveillance_focal_persons
		WHERE district_id = $1 AND is_active = true
		LIMIT 1`

	err := db.QueryRow(query, districtID).Scan(
		&focalPerson.ID,
		&focalPerson.Name,
		&focalPerson.Phone,
		&focalPerson.DistrictID,
		&focalPerson.Email,
		&focalPerson.Position,
		&focalPerson.IsActive,
		&focalPerson.CreatedAt,
		&focalPerson.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return focalPerson, nil
}

// Get active surveillance focal persons by district code
func GetActiveSurveillanceFocalPersonsByDistrict(db *sql.DB, districtCode string) ([]SurveillanceFocalPerson, error) {
	rows, err := db.Query(`
		SELECT id, name, phone, district_id, email, position, is_active, created_at, updated_at
		FROM surveillance_focal_persons
		WHERE district_id = $1 AND is_active = true
		ORDER BY name
	`, districtCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var persons []SurveillanceFocalPerson
	for rows.Next() {
		var p SurveillanceFocalPerson
		if err := rows.Scan(&p.ID, &p.Name, &p.Phone, &p.DistrictID, &p.Email, &p.Position, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		persons = append(persons, p)
	}
	return persons, nil
}
