package models

import (
	"database/sql"
	"time"
)

// UserOutbreak represents a user's assignment to an outbreak
type UserOutbreak struct {
	ID         int64
	UserID     int64
	OutbreakID int64
	AssignedAt time.Time
	AssignedBy sql.NullInt64
	IsActive   bool
	// Related data
	User           *User
	Outbreak       *Outbreak
	AssignedByUser *User
}

// PatientManagementRole represents a user's role in patient management
type PatientManagementRole struct {
	ID         int64
	UserID     int64
	RoleType   string // 'registration', 'admission', 'discharge', 'care'
	OutbreakID sql.NullInt64
	FacilityID sql.NullInt64
	IsActive   bool
	CreatedAt  time.Time
	CreatedBy  sql.NullInt64
	// Related data
	User          *User
	Outbreak      *Outbreak
	Facility      *Facility
	CreatedByUser *User
}

// PasswordChangeRequest represents a password change request
type PasswordChangeRequest struct {
	ID                  int64
	UserID              int64
	RequestToken        string
	CurrentPasswordHash string
	NewPasswordHash     string
	NewPasswordSalt     string
	IsApproved          bool
	ApprovedBy          sql.NullInt64
	ApprovedAt          sql.NullTime
	ExpiresAt           time.Time
	CreatedAt           time.Time
	IPAddress           sql.NullString
	UserAgent           sql.NullString
	// Related data
	User           *User
	ApprovedByUser *User
}

// EmployeeAssignment represents an employee's assignment
type EmployeeAssignment struct {
	ID             int64
	EmployeeID     int64
	AssignmentType string // 'facility', 'outbreak', 'department', 'project'
	AssignmentID   int64
	AssignmentName string
	StartDate      time.Time
	EndDate        sql.NullTime
	IsPrimary      bool
	IsActive       bool
	CreatedAt      time.Time
	CreatedBy      sql.NullInt64
	// Related data
	Employee      *Employee
	CreatedByUser *User
}

// EmployeeDepartment represents an employee department
type EmployeeDepartment struct {
	ID          int64
	Name        string
	Description sql.NullString
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// EmployeeTitle represents an employee title
type EmployeeTitle struct {
	ID          int64
	Name        string
	Description sql.NullString
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UserOutbreakService provides methods for managing user outbreak assignments
type UserOutbreakService struct {
	db *sql.DB
}

// NewUserOutbreakService creates a new user outbreak service
func NewUserOutbreakService(db *sql.DB) *UserOutbreakService {
	return &UserOutbreakService{db: db}
}

// AssignUserToOutbreak assigns a user to an outbreak
func (s *UserOutbreakService) AssignUserToOutbreak(userID, outbreakID, assignedBy int64) error {
	query := `
		INSERT INTO user_outbreaks (user_id, outbreak_id, assigned_by, is_active)
		VALUES ($1, $2, $3, true)
		ON CONFLICT (user_id, outbreak_id) 
		DO UPDATE SET is_active = true, assigned_by = $3, assigned_at = CURRENT_TIMESTAMP
	`
	_, err := s.db.Exec(query, userID, outbreakID, assignedBy)
	return err
}

// RemoveUserFromOutbreak removes a user from an outbreak
func (s *UserOutbreakService) RemoveUserFromOutbreak(userID, outbreakID int64) error {
	query := `UPDATE user_outbreaks SET is_active = false WHERE user_id = $1 AND outbreak_id = $2`
	_, err := s.db.Exec(query, userID, outbreakID)
	return err
}

// GetUserOutbreaks gets all outbreaks assigned to a user
func (s *UserOutbreakService) GetUserOutbreaks(userID int64) ([]UserOutbreak, error) {
	query := `
		SELECT uo.id, uo.user_id, uo.outbreak_id, uo.assigned_at, uo.assigned_by, uo.is_active,
		       o.name, o.description, o.start_date, o.end_date, o.status, o.status as outbreak_active
		FROM user_outbreaks uo
		JOIN outbreaks o ON uo.outbreak_id = o.id
		WHERE uo.user_id = $1 AND uo.is_active = true
		ORDER BY uo.assigned_at DESC
	`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var outbreaks []UserOutbreak
	for rows.Next() {
		var uo UserOutbreak
		var outbreak Outbreak
		err := rows.Scan(
			&uo.ID, &uo.UserID, &uo.OutbreakID, &uo.AssignedAt, &uo.AssignedBy, &uo.IsActive,
			&outbreak.Name, &outbreak.Description, &outbreak.StartDate, &outbreak.EndDate,
			&outbreak.Status, &outbreak.Status,
		)
		if err != nil {
			return nil, err
		}
		uo.Outbreak = &outbreak
		outbreaks = append(outbreaks, uo)
	}
	return outbreaks, nil
}

// GetAllOutbreakAssignments gets all outbreak assignments for display
func (s *UserOutbreakService) GetAllOutbreakAssignments() ([]UserOutbreak, error) {
	query := `
		SELECT uo.id, uo.user_id, uo.outbreak_id, uo.assigned_at, uo.assigned_by, uo.is_active,
		       o.name, o.description, o.start_date, o.end_date, o.status, o.outbreak_type,
		       u.user_name
		FROM user_outbreaks uo
		JOIN outbreaks o ON uo.outbreak_id = o.id
		JOIN users u ON uo.user_id = u.user_id
		WHERE uo.is_active = true
		ORDER BY uo.assigned_at DESC
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []UserOutbreak
	for rows.Next() {
		var uo UserOutbreak
		var outbreak Outbreak
		var user User
		err := rows.Scan(
			&uo.ID, &uo.UserID, &uo.OutbreakID, &uo.AssignedAt, &uo.AssignedBy, &uo.IsActive,
			&outbreak.Name, &outbreak.Description, &outbreak.StartDate, &outbreak.EndDate,
			&outbreak.Status, &outbreak.OutbreakType, &user.UserName,
		)
		if err != nil {
			return nil, err
		}
		uo.Outbreak = &outbreak
		uo.User = &user
		assignments = append(assignments, uo)
	}
	return assignments, nil
}

// GetOutbreakUsers gets all users assigned to an outbreak
func (s *UserOutbreakService) GetOutbreakUsers(outbreakID int64) ([]User, error) {
	query := `
		SELECT u.user_id, u.user_name, u.user_name, u.user_name, u.user_name, u.user_name,
		       uo.assigned_at, uo.assigned_by
		FROM user_outbreaks uo
		JOIN users u ON uo.user_id = u.user_id
		WHERE uo.outbreak_id = $1 AND uo.is_active = true
		ORDER BY uo.assigned_at DESC
	`
	rows, err := s.db.Query(query, outbreakID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		var assignedAt time.Time
		var assignedBy sql.NullInt64
		err := rows.Scan(
			&user.UserID, &user.UserName, &user.UserName, &user.UserName, &user.UserName, &user.UserName,
			&assignedAt, &assignedBy,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// PatientManagementRoleService provides methods for managing patient management roles
type PatientManagementRoleService struct {
	db *sql.DB
}

// NewPatientManagementRoleService creates a new patient management role service
func NewPatientManagementRoleService(db *sql.DB) *PatientManagementRoleService {
	return &PatientManagementRoleService{db: db}
}

// AssignPatientRole assigns a patient management role to a user
func (s *PatientManagementRoleService) AssignPatientRole(userID int64, roleType string, outbreakID, facilityID *int64, createdBy int64) error {
	query := `
		INSERT INTO patient_management_roles (user_id, role_type, outbreak_id, facility_id, created_by, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
		ON CONFLICT (user_id, role_type, outbreak_id, facility_id) 
		DO UPDATE SET is_active = true, created_by = $5
	`
	_, err := s.db.Exec(query, userID, roleType, outbreakID, facilityID, createdBy)
	return err
}

// RemovePatientRole removes a patient management role from a user
func (s *PatientManagementRoleService) RemovePatientRole(userID int64, roleType string, outbreakID, facilityID *int64) error {
	query := `UPDATE patient_management_roles SET is_active = false WHERE user_id = $1 AND role_type = $2 AND outbreak_id IS NOT DISTINCT FROM $3 AND facility_id IS NOT DISTINCT FROM $4`
	_, err := s.db.Exec(query, userID, roleType, outbreakID, facilityID)
	return err
}

// GetUserPatientRoles gets all patient management roles for a user
func (s *PatientManagementRoleService) GetUserPatientRoles(userID int64) ([]PatientManagementRole, error) {
	query := `
		SELECT pmr.id, pmr.user_id, pmr.role_type, pmr.outbreak_id, pmr.facility_id, 
		       pmr.is_active, pmr.created_at, pmr.created_by,
		       o.name as outbreak_name, f.facility_name as facility_name
		FROM patient_management_roles pmr
		LEFT JOIN outbreaks o ON pmr.outbreak_id = o.id
		LEFT JOIN facility f ON pmr.facility_id = f.facility_id
		WHERE pmr.user_id = $1 AND pmr.is_active = true
		ORDER BY pmr.created_at DESC
	`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []PatientManagementRole
	for rows.Next() {
		var role PatientManagementRole
		var outbreak Outbreak
		var facility Facility
		err := rows.Scan(
			&role.ID, &role.UserID, &role.RoleType, &role.OutbreakID, &role.FacilityID,
			&role.IsActive, &role.CreatedAt, &role.CreatedBy,
			&outbreak.Name, &facility.FacilityName,
		)
		if err != nil {
			return nil, err
		}
		if role.OutbreakID.Valid {
			role.Outbreak = &outbreak
		}
		if role.FacilityID.Valid {
			role.Facility = &facility
		}
		roles = append(roles, role)
	}
	return roles, nil
}

// CheckPatientPermission checks if a user has a specific patient management permission
func (s *PatientManagementRoleService) CheckPatientPermission(userID int64, roleType string, outbreakID, facilityID *int64) bool {
	query := `
		SELECT COUNT(*) FROM patient_management_roles 
		WHERE user_id = $1 AND role_type = $2 AND is_active = true
		AND (outbreak_id IS NULL OR outbreak_id = $3)
		AND (facility_id IS NULL OR facility_id = $4)
	`
	var count int
	err := s.db.QueryRow(query, userID, roleType, outbreakID, facilityID).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}
