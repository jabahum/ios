package models

import (
	"database/sql"
)

// UserService provides methods for managing users
type UserService struct {
	db *sql.DB
}

// NewUserService creates a new user service
func NewUserService(db *sql.DB) *UserService {
	return &UserService{db: db}
}

// GetUserByID gets a user by ID
func (s *UserService) GetUserByID(userID int64) (*User, error) {
	query := `SELECT user_id, user_name, user_pass, user_employee FROM users WHERE user_id = $1`
	user := &User{}
	err := s.db.QueryRow(query, userID).Scan(&user.UserID, &user.UserName, &user.UserPass, &user.UserEmployee)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByEmail gets a user by email (assuming email is stored in user_name for now)
func (s *UserService) GetUserByEmail(email string) (*User, error) {
	query := `SELECT user_id, user_name, user_pass, user_employee FROM users WHERE user_name = $1`
	user := &User{}
	err := s.db.QueryRow(query, email).Scan(&user.UserID, &user.UserName, &user.UserPass, &user.UserEmployee)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetAllUsers gets all users
func (s *UserService) GetAllUsers() ([]*User, error) {
	query := `SELECT user_id, user_name, user_pass, user_employee FROM users ORDER BY user_name`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.UserID, &user.UserName, &user.UserPass, &user.UserEmployee)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// UpdatePassword updates a user's password
func (s *UserService) UpdatePassword(userID int64, passwordHash, passwordSalt string) error {
	query := `UPDATE users SET user_pass = $1 WHERE user_id = $2`
	_, err := s.db.Exec(query, passwordHash, userID)
	return err
}

// CreatePasswordChangeRequest creates a password change request
func (s *UserService) CreatePasswordChangeRequest(request *PasswordChangeRequest) error {
	query := `
		INSERT INTO password_change_requests (user_id, request_token, current_password_hash, 
		                                     new_password_hash, new_password_salt, expires_at, 
		                                     ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP)
		RETURNING id
	`
	err := s.db.QueryRow(
		query,
		request.UserID, request.RequestToken, request.CurrentPasswordHash,
		request.NewPasswordHash, request.NewPasswordSalt, request.ExpiresAt,
		request.IPAddress, request.UserAgent,
	).Scan(&request.ID)
	return err
}

// GetPasswordChangeRequest gets a password change request by token
func (s *UserService) GetPasswordChangeRequest(token string) (*PasswordChangeRequest, error) {
	query := `
		SELECT id, user_id, request_token, current_password_hash, new_password_hash, 
		       new_password_salt, is_approved, approved_by, approved_at, expires_at, 
		       created_at, ip_address, user_agent
		FROM password_change_requests 
		WHERE request_token = $1 AND expires_at > CURRENT_TIMESTAMP
	`
	request := &PasswordChangeRequest{}
	err := s.db.QueryRow(query, token).Scan(
		&request.ID, &request.UserID, &request.RequestToken, &request.CurrentPasswordHash,
		&request.NewPasswordHash, &request.NewPasswordSalt, &request.IsApproved,
		&request.ApprovedBy, &request.ApprovedAt, &request.ExpiresAt, &request.CreatedAt,
		&request.IPAddress, &request.UserAgent,
	)
	if err != nil {
		return nil, err
	}
	return request, nil
}

// UpdatePasswordChangeRequest updates a password change request
func (s *UserService) UpdatePasswordChangeRequest(request *PasswordChangeRequest) error {
	query := `
		UPDATE password_change_requests SET 
			new_password_hash = $1, new_password_salt = $2, is_approved = $3, 
			approved_at = $4
		WHERE id = $5
	`
	_, err := s.db.Exec(query, request.NewPasswordHash, request.NewPasswordSalt,
		request.IsApproved, request.ApprovedAt, request.ID)
	return err
}

// OutbreakService provides methods for managing outbreaks
type OutbreakService struct {
	db *sql.DB
}

// NewOutbreakService creates a new outbreak service
func NewOutbreakService(db *sql.DB) *OutbreakService {
	return &OutbreakService{db: db}
}

// GetOutbreakByID gets an outbreak by ID
func (s *OutbreakService) GetOutbreakByID(outbreakID int64) (*Outbreak, error) {
	query := `
		SELECT id, name, description, start_date, end_date, status, enter_on, enter_by, edit_on, edit_by
		FROM outbreaks WHERE id = $1
	`
	outbreak := &Outbreak{}
	err := s.db.QueryRow(query, outbreakID).Scan(
		&outbreak.ID, &outbreak.Name, &outbreak.Description, &outbreak.StartDate,
		&outbreak.EndDate, &outbreak.Status, &outbreak.EnterOn, &outbreak.EnterBy,
		&outbreak.EditOn, &outbreak.EditBy,
	)
	if err != nil {
		return nil, err
	}
	return outbreak, nil
}

// GetOutbreak gets an outbreak by ID (alias for GetOutbreakByID)
func (s *OutbreakService) GetOutbreak(outbreakID int64) (*Outbreak, error) {
	return s.GetOutbreakByID(outbreakID)
}

// GetActiveOutbreaks gets all active outbreaks
func (s *OutbreakService) GetActiveOutbreaks() ([]*Outbreak, error) {
	query := `
		SELECT id, name, description, start_date, end_date, status, enter_on, enter_by, edit_on, edit_by
		FROM outbreaks WHERE status != 'closed' ORDER BY start_date DESC
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var outbreaks []*Outbreak
	for rows.Next() {
		outbreak := &Outbreak{}
		err := rows.Scan(
			&outbreak.ID, &outbreak.Name, &outbreak.Description, &outbreak.StartDate,
			&outbreak.EndDate, &outbreak.Status, &outbreak.EnterOn, &outbreak.EnterBy,
			&outbreak.EditOn, &outbreak.EditBy,
		)
		if err != nil {
			return nil, err
		}
		outbreaks = append(outbreaks, outbreak)
	}
	return outbreaks, nil
}

// GetAllOutbreaks gets all outbreaks
func (s *OutbreakService) GetAllOutbreaks() ([]*Outbreak, error) {
	query := `
		SELECT id, name, description, start_date, end_date, status, enter_on, enter_by, edit_on, edit_by
		FROM outbreaks ORDER BY start_date DESC
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var outbreaks []*Outbreak
	for rows.Next() {
		outbreak := &Outbreak{}
		err := rows.Scan(
			&outbreak.ID, &outbreak.Name, &outbreak.Description, &outbreak.StartDate,
			&outbreak.EndDate, &outbreak.Status, &outbreak.EnterOn, &outbreak.EnterBy,
			&outbreak.EditOn, &outbreak.EditBy,
		)
		if err != nil {
			return nil, err
		}
		outbreaks = append(outbreaks, outbreak)
	}
	return outbreaks, nil
}

// FacilityService provides methods for managing facilities
type FacilityService struct {
	db *sql.DB
}

// NewFacilityService creates a new facility service
func NewFacilityService(db *sql.DB) *FacilityService {
	return &FacilityService{db: db}
}

// GetFacilityByID gets a facility by ID
func (s *FacilityService) GetFacilityByID(facilityID int64) (*Facility, error) {
	query := `SELECT facility_id, facility_name, facility_level FROM facility WHERE facility_id = $1`
	facility := &Facility{}
	err := s.db.QueryRow(query, facilityID).Scan(&facility.FacilityID, &facility.FacilityName, &facility.FacilityLevel)
	if err != nil {
		return nil, err
	}
	return facility, nil
}

// GetAllFacilities gets all facilities
func (s *FacilityService) GetAllFacilities() ([]*Facility, error) {
	query := `SELECT facility_id, facility_name, facility_level FROM facility ORDER BY facility_name`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var facilities []*Facility
	for rows.Next() {
		facility := &Facility{}
		err := rows.Scan(&facility.FacilityID, &facility.FacilityName, &facility.FacilityLevel)
		if err != nil {
			return nil, err
		}
		facilities = append(facilities, facility)
	}
	return facilities, nil
}
