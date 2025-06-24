package models

import (
	"database/sql"
	"time"
)

type District struct {
	ID        int            `json:"id"`
	Name      string         `json:"name"`
	Code      sql.NullString `json:"code"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type Subcounty struct {
	ID         int            `json:"id"`
	DistrictID int            `json:"district_id"`
	Name       string         `json:"name"`
	Code       sql.NullString `json:"code"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

type Parish struct {
	ID          int            `json:"id"`
	DistrictID  int            `json:"district_id"`
	SubcountyID int            `json:"subcounty_id"`
	Name        string         `json:"name"`
	Code        sql.NullString `json:"code"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type Village struct {
	ID          int            `json:"id"`
	DistrictID  int            `json:"district_id"`
	SubcountyID int            `json:"subcounty_id"`
	ParishID    int            `json:"parish_id"`
	Name        string         `json:"name"`
	Code        sql.NullString `json:"code"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// Get all districts
func GetDistricts(db *sql.DB) ([]District, error) {
	rows, err := db.Query("SELECT id, name, code, created_at, updated_at FROM districts ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var districts []District
	for rows.Next() {
		var d District
		if err := rows.Scan(&d.ID, &d.Name, &d.Code, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		districts = append(districts, d)
	}
	return districts, nil
}

// Get subcounties by district ID
func GetSubcountiesByDistrict(db *sql.DB, districtID int) ([]Subcounty, error) {
	rows, err := db.Query("SELECT id, district_id, name, code, created_at, updated_at FROM subcounties WHERE district_id = $1 ORDER BY name", districtID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subcounties []Subcounty
	for rows.Next() {
		var s Subcounty
		if err := rows.Scan(&s.ID, &s.DistrictID, &s.Name, &s.Code, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		subcounties = append(subcounties, s)
	}
	return subcounties, nil
}

// Get parishes by subcounty ID
func GetParishesBySubcounty(db *sql.DB, subcountyID int) ([]Parish, error) {
	rows, err := db.Query("SELECT id, district_id, subcounty_id, name, code, created_at, updated_at FROM parishes WHERE subcounty_id = $1 ORDER BY name", subcountyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parishes []Parish
	for rows.Next() {
		var p Parish
		if err := rows.Scan(&p.ID, &p.DistrictID, &p.SubcountyID, &p.Name, &p.Code, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		parishes = append(parishes, p)
	}
	return parishes, nil
}

// Get parishes by district ID
func GetParishesByDistrict(db *sql.DB, districtID int) ([]Parish, error) {
	rows, err := db.Query("SELECT id, district_id, subcounty_id, name, code, created_at, updated_at FROM parishes WHERE district_id = $1 ORDER BY name", districtID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parishes []Parish
	for rows.Next() {
		var p Parish
		if err := rows.Scan(&p.ID, &p.DistrictID, &p.SubcountyID, &p.Name, &p.Code, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		parishes = append(parishes, p)
	}
	return parishes, nil
}

// Get villages by parish ID
func GetVillagesByParish(db *sql.DB, parishID int) ([]Village, error) {
	rows, err := db.Query("SELECT id, district_id, subcounty_id, parish_id, name, code, created_at, updated_at FROM villages WHERE parish_id = $1 ORDER BY name", parishID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var villages []Village
	for rows.Next() {
		var v Village
		if err := rows.Scan(&v.ID, &v.DistrictID, &v.SubcountyID, &v.ParishID, &v.Name, &v.Code, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		villages = append(villages, v)
	}
	return villages, nil
}

// Get villages by district ID
func GetVillagesByDistrict(db *sql.DB, districtID int) ([]Village, error) {
	rows, err := db.Query("SELECT id, district_id, subcounty_id, parish_id, name, code, created_at, updated_at FROM villages WHERE district_id = $1 ORDER BY name", districtID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var villages []Village
	for rows.Next() {
		var v Village
		if err := rows.Scan(&v.ID, &v.DistrictID, &v.SubcountyID, &v.ParishID, &v.Name, &v.Code, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		villages = append(villages, v)
	}
	return villages, nil
}

// Get villages by subcounty ID
func GetVillagesBySubcounty(db *sql.DB, subcountyID int) ([]Village, error) {
	rows, err := db.Query("SELECT id, district_id, subcounty_id, parish_id, name, code, created_at, updated_at FROM villages WHERE subcounty_id = $1 ORDER BY name", subcountyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var villages []Village
	for rows.Next() {
		var v Village
		if err := rows.Scan(&v.ID, &v.DistrictID, &v.SubcountyID, &v.ParishID, &v.Name, &v.Code, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		villages = append(villages, v)
	}
	return villages, nil
}
