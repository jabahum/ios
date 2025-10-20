package models

import (
	"context"
	"database/sql"
	"time"
)

// EnhancedUser represents a user with RBAC functionality
// This extends the base User model with additional fields
type EnhancedUser struct {
	UserID              int            `json:"user_id"`
	UserName            sql.NullString `json:"user_name"`
	UserPass            sql.NullString `json:"user_pass"`
	UserEmployee        sql.NullInt64  `json:"user_employee"`
	Email               sql.NullString `json:"email"`
	FirstName           sql.NullString `json:"first_name"`
	LastName            sql.NullString `json:"last_name"`
	PasswordHash        sql.NullString `json:"-"` // Never expose in JSON
	PasswordSalt        sql.NullString `json:"-"` // Never expose in JSON
	IsActive            bool           `json:"is_active"`
	IsLocked            bool           `json:"is_locked"`
	FailedLoginAttempts int            `json:"-"`
	LastLoginAt         sql.NullTime   `json:"last_login_at"`
	PasswordChangedAt   sql.NullTime   `json:"password_changed_at"`
	PasswordExpiresAt   sql.NullTime   `json:"password_expires_at"`
	DepartmentID        sql.NullInt64  `json:"department_id"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	CreatedBy           sql.NullInt64  `json:"created_by"`
	UpdatedBy           sql.NullInt64  `json:"updated_by"`
	// xo fields
	_exists, _deleted bool
}

// Exists returns true when the [EnhancedUser] exists in the database.
func (u *EnhancedUser) Exists() bool {
	return u._exists
}

// Deleted returns true when the [EnhancedUser] has been marked for deletion
// from the database.
func (u *EnhancedUser) Deleted() bool {
	return u._deleted
}

// SetAsExists sets the _exists flag to true
func (u *EnhancedUser) SetAsExists() {
	u._exists = true
}

// SetAsDeleted sets the _deleted flag to true
func (u *EnhancedUser) SetAsDeleted() {
	u._deleted = true
}

// Insert inserts the [EnhancedUser] to the database.
func (u *EnhancedUser) Insert(ctx context.Context, db DB) error {
	switch {
	case u._exists: // already exists
		return logerror(&ErrInsertFailed{ErrAlreadyExists})
	case u._deleted: // deleted
		return logerror(&ErrInsertFailed{ErrMarkedForDeletion})
	}
	// insert (primary key generated and returned by database)
	const sqlstr = `INSERT INTO public.users (` +
		`user_name, user_pass, user_employee, email, first_name, last_name, ` +
		`password_hash, password_salt, is_active, is_locked, failed_login_attempts, ` +
		`last_login_at, password_changed_at, password_expires_at, department_id, ` +
		`created_at, updated_at, created_by, updated_by` +
		`) VALUES (` +
		`$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19` +
		`) RETURNING user_id`
	// run
	logf(sqlstr, u.UserName, u.UserPass, u.UserEmployee, u.Email, u.FirstName, u.LastName,
		u.PasswordHash, u.PasswordSalt, u.IsActive, u.IsLocked, u.FailedLoginAttempts,
		u.LastLoginAt, u.PasswordChangedAt, u.PasswordExpiresAt, u.DepartmentID,
		u.CreatedAt, u.UpdatedAt, u.CreatedBy, u.UpdatedBy)
	if err := db.QueryRowContext(ctx, sqlstr, u.UserName, u.UserPass, u.UserEmployee, u.Email, u.FirstName, u.LastName,
		u.PasswordHash, u.PasswordSalt, u.IsActive, u.IsLocked, u.FailedLoginAttempts,
		u.LastLoginAt, u.PasswordChangedAt, u.PasswordExpiresAt, u.DepartmentID,
		u.CreatedAt, u.UpdatedAt, u.CreatedBy, u.UpdatedBy).Scan(&u.UserID); err != nil {
		return logerror(err)
	}
	// set exists
	u._exists = true
	return nil
}

// Update updates an [EnhancedUser] in the database.
func (u *EnhancedUser) Update(ctx context.Context, db DB) error {
	switch {
	case !u._exists: // doesn't exist
		return logerror(&ErrUpdateFailed{ErrDoesNotExist})
	case u._deleted: // deleted
		return logerror(&ErrUpdateFailed{ErrMarkedForDeletion})
	}
	// update with composite primary key
	const sqlstr = `UPDATE public.users SET ` +
		`user_name = $1, user_pass = $2, user_employee = $3, email = $4, first_name = $5, last_name = $6, ` +
		`password_hash = $7, password_salt = $8, is_active = $9, is_locked = $10, failed_login_attempts = $11, ` +
		`last_login_at = $12, password_changed_at = $13, password_expires_at = $14, department_id = $15, ` +
		`updated_at = $16, updated_by = $17 ` +
		`WHERE user_id = $18`
	// run
	logf(sqlstr, u.UserName, u.UserPass, u.UserEmployee, u.Email, u.FirstName, u.LastName,
		u.PasswordHash, u.PasswordSalt, u.IsActive, u.IsLocked, u.FailedLoginAttempts,
		u.LastLoginAt, u.PasswordChangedAt, u.PasswordExpiresAt, u.DepartmentID,
		u.UpdatedAt, u.UpdatedBy, u.UserID)
	if _, err := db.ExecContext(ctx, sqlstr, u.UserName, u.UserPass, u.UserEmployee, u.Email, u.FirstName, u.LastName,
		u.PasswordHash, u.PasswordSalt, u.IsActive, u.IsLocked, u.FailedLoginAttempts,
		u.LastLoginAt, u.PasswordChangedAt, u.PasswordExpiresAt, u.DepartmentID,
		u.UpdatedAt, u.UpdatedBy, u.UserID); err != nil {
		return logerror(err)
	}
	return nil
}

// Save saves the [EnhancedUser] to the database.
func (u *EnhancedUser) Save(ctx context.Context, db DB) error {
	if u.Exists() {
		return u.Update(ctx, db)
	}
	return u.Insert(ctx, db)
}

// Delete deletes the [EnhancedUser] from the database.
func (u *EnhancedUser) Delete(ctx context.Context, db DB) error {
	switch {
	case !u._exists: // doesn't exist
		return nil
	case u._deleted: // deleted
		return nil
	}
	// delete with single primary key
	const sqlstr = `DELETE FROM public.users ` +
		`WHERE user_id = $1`
	// run
	logf(sqlstr, u.UserID)
	if _, err := db.ExecContext(ctx, sqlstr, u.UserID); err != nil {
		return logerror(err)
	}
	// set deleted
	u._deleted = true
	return nil
}

// EnhancedUserByUserID retrieves a row from 'public.users' as an [EnhancedUser].
func EnhancedUserByUserID(ctx context.Context, db DB, userID int) (*EnhancedUser, error) {
	// query
	const sqlstr = `SELECT ` +
		`user_id, user_name, user_pass, user_employee, email, first_name, last_name, ` +
		`password_hash, password_salt, is_active, is_locked, failed_login_attempts, ` +
		`last_login_at, password_changed_at, password_expires_at, department_id, ` +
		`created_at, updated_at, created_by, updated_by ` +
		`FROM public.users ` +
		`WHERE user_id = $1`
	// run
	logf(sqlstr, userID)
	u := EnhancedUser{
		_exists: true,
	}
	if err := db.QueryRowContext(ctx, sqlstr, userID).Scan(&u.UserID, &u.UserName, &u.UserPass, &u.UserEmployee,
		&u.Email, &u.FirstName, &u.LastName, &u.PasswordHash, &u.PasswordSalt, &u.IsActive, &u.IsLocked,
		&u.FailedLoginAttempts, &u.LastLoginAt, &u.PasswordChangedAt, &u.PasswordExpiresAt, &u.DepartmentID,
		&u.CreatedAt, &u.UpdatedAt, &u.CreatedBy, &u.UpdatedBy); err != nil {
		return nil, logerror(err)
	}
	return &u, nil
}

// EnhancedUserByUsername retrieves a row from 'public.users' as an [EnhancedUser] by username.
func EnhancedUserByUsername(ctx context.Context, db DB, username string) (*EnhancedUser, error) {
	// query
	const sqlstr = `SELECT ` +
		`user_id, user_name, user_pass, user_employee, email, first_name, last_name, ` +
		`password_hash, password_salt, is_active, is_locked, failed_login_attempts, ` +
		`last_login_at, password_changed_at, password_expires_at, department_id, ` +
		`created_at, updated_at, created_by, updated_by ` +
		`FROM public.users ` +
		`WHERE user_name = $1`
	// run
	logf(sqlstr, username)
	u := EnhancedUser{
		_exists: true,
	}
	if err := db.QueryRowContext(ctx, sqlstr, username).Scan(&u.UserID, &u.UserName, &u.UserPass, &u.UserEmployee,
		&u.Email, &u.FirstName, &u.LastName, &u.PasswordHash, &u.PasswordSalt, &u.IsActive, &u.IsLocked,
		&u.FailedLoginAttempts, &u.LastLoginAt, &u.PasswordChangedAt, &u.PasswordExpiresAt, &u.DepartmentID,
		&u.CreatedAt, &u.UpdatedAt, &u.CreatedBy, &u.UpdatedBy); err != nil {
		return nil, logerror(err)
	}
	return &u, nil
}

// GetAllEnhancedUsers retrieves all enhanced users from the database.
func GetAllEnhancedUsers(ctx context.Context, db DB) ([]*EnhancedUser, error) {
	// query
	const sqlstr = `SELECT ` +
		`user_id, user_name, user_pass, user_employee, email, first_name, last_name, ` +
		`password_hash, password_salt, is_active, is_locked, failed_login_attempts, ` +
		`last_login_at, password_changed_at, password_expires_at, department_id, ` +
		`created_at, updated_at, created_by, updated_by ` +
		`FROM public.users ` +
		`ORDER BY user_name`
	// run
	logf(sqlstr)
	rows, err := db.QueryContext(ctx, sqlstr)
	if err != nil {
		return nil, logerror(err)
	}
	defer rows.Close()
	// process
	var res []*EnhancedUser
	for rows.Next() {
		u := EnhancedUser{
			_exists: true,
		}
		// scan
		if err := rows.Scan(&u.UserID, &u.UserName, &u.UserPass, &u.UserEmployee,
			&u.Email, &u.FirstName, &u.LastName, &u.PasswordHash, &u.PasswordSalt, &u.IsActive, &u.IsLocked,
			&u.FailedLoginAttempts, &u.LastLoginAt, &u.PasswordChangedAt, &u.PasswordExpiresAt, &u.DepartmentID,
			&u.CreatedAt, &u.UpdatedAt, &u.CreatedBy, &u.UpdatedBy); err != nil {
			return nil, logerror(err)
		}
		res = append(res, &u)
	}
	if err := rows.Err(); err != nil {
		return nil, logerror(err)
	}
	return res, nil
}

// FullName returns the full name of the user
func (u *EnhancedUser) FullName() string {
	if u.FirstName.Valid && u.LastName.Valid {
		return u.FirstName.String + " " + u.LastName.String
	}
	if u.FirstName.Valid {
		return u.FirstName.String
	}
	if u.LastName.Valid {
		return u.LastName.String
	}
	if u.UserName.Valid {
		return u.UserName.String
	}
	return "Unknown User"
}

// GetEmail returns the email address
func (u *EnhancedUser) GetEmail() string {
	if u.Email.Valid {
		return u.Email.String
	}
	return ""
}

// GetPhone returns the phone number (if available)
func (u *EnhancedUser) GetPhone() string {
	// Phone field doesn't exist in EnhancedUser, return empty string
	return ""
}
