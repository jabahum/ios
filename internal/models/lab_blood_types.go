package models

import (
	"context"
	"database/sql"
	"time"
)

// LabBloodType represents a row from 'public.lab_blood_types'.
type LabBloodType struct {
	ID          int            `json:"id"`          // id
	Name        string         `json:"name"`        // name
	Description sql.NullString `json:"description"` // description
	Category    sql.NullString `json:"category"`    // category
	CreatedAt   time.Time      `json:"created_at"`  // created_at
	UpdatedAt   time.Time      `json:"updated_at"`  // updated_at
	// xo fields
	_exists, _deleted bool
}

// Exists returns true when the [LabBloodType] exists in the database.
func (lbt *LabBloodType) Exists() bool {
	return lbt._exists
}

// Deleted returns true when the [LabBloodType] has been marked for deletion
// from the database.
func (lbt *LabBloodType) Deleted() bool {
	return lbt._deleted
}

// Insert inserts the [LabBloodType] to the database.
func (lbt *LabBloodType) Insert(ctx context.Context, db DB) error {
	switch {
	case lbt._exists: // already exists
		return logerror(&ErrInsertFailed{ErrAlreadyExists})
	case lbt._deleted: // deleted
		return logerror(&ErrInsertFailed{ErrMarkedForDeletion})
	}
	// insert (primary key generated and returned by database)
	const sqlstr = `INSERT INTO public.lab_blood_types (` +
		`name, description, category, created_at, updated_at` +
		`) VALUES (` +
		`$1, $2, $3, $4, $5` +
		`) RETURNING id`
	// run
	logf(sqlstr, lbt.Name, lbt.Description, lbt.Category, lbt.CreatedAt, lbt.UpdatedAt)
	if err := db.QueryRowContext(ctx, sqlstr, lbt.Name, lbt.Description, lbt.Category, lbt.CreatedAt, lbt.UpdatedAt).Scan(&lbt.ID); err != nil {
		return logerror(err)
	}
	// set exists
	lbt._exists = true
	return nil
}

// Update updates a [LabBloodType] in the database.
func (lbt *LabBloodType) Update(ctx context.Context, db DB) error {
	switch {
	case !lbt._exists: // doesn't exist
		return logerror(&ErrUpdateFailed{ErrDoesNotExist})
	case lbt._deleted: // deleted
		return logerror(&ErrUpdateFailed{ErrMarkedForDeletion})
	}
	// update with composite primary key
	const sqlstr = `UPDATE public.lab_blood_types SET ` +
		`name = $1, description = $2, category = $3, updated_at = $4` +
		` WHERE id = $5`
	// run
	logf(sqlstr, lbt.Name, lbt.Description, lbt.Category, lbt.UpdatedAt, lbt.ID)
	if _, err := db.ExecContext(ctx, sqlstr, lbt.Name, lbt.Description, lbt.Category, lbt.UpdatedAt, lbt.ID); err != nil {
		return logerror(err)
	}
	return nil
}

// Save saves the [LabBloodType] to the database.
func (lbt *LabBloodType) Save(ctx context.Context, db DB) error {
	if lbt.Exists() {
		return lbt.Update(ctx, db)
	}
	return lbt.Insert(ctx, db)
}

// Upsert performs an upsert for [LabBloodType].
func (lbt *LabBloodType) Upsert(ctx context.Context, db DB) error {
	switch {
	case lbt._deleted: // deleted
		return logerror(&ErrUpsertFailed{ErrMarkedForDeletion})
	}
	// upsert
	const sqlstr = `INSERT INTO public.lab_blood_types (` +
		`id, name, description, category, created_at, updated_at` +
		`) VALUES (` +
		`$1, $2, $3, $4, $5, $6` +
		`)` +
		` ON CONFLICT (id) DO ` +
		`UPDATE SET ` +
		`name = EXCLUDED.name, description = EXCLUDED.description, category = EXCLUDED.category, updated_at = EXCLUDED.updated_at`
	// run
	logf(sqlstr, lbt.ID, lbt.Name, lbt.Description, lbt.Category, lbt.CreatedAt, lbt.UpdatedAt)
	if _, err := db.ExecContext(ctx, sqlstr, lbt.ID, lbt.Name, lbt.Description, lbt.Category, lbt.CreatedAt, lbt.UpdatedAt); err != nil {
		return logerror(err)
	}
	// set exists
	lbt._exists = true
	return nil
}

// Delete deletes the [LabBloodType] from the database.
func (lbt *LabBloodType) Delete(ctx context.Context, db DB) error {
	switch {
	case !lbt._exists: // doesn't exist
		return nil
	case lbt._deleted: // deleted
		return nil
	}
	// delete with single primary key
	const sqlstr = `DELETE FROM public.lab_blood_types WHERE id = $1`
	// run
	logf(sqlstr, lbt.ID)
	if _, err := db.ExecContext(ctx, sqlstr, lbt.ID); err != nil {
		return logerror(err)
	}
	// set deleted
	lbt._deleted = true
	return nil
}

// LabBloodTypeByID retrieves a row from 'public.lab_blood_types' as a [LabBloodType].
func LabBloodTypeByID(ctx context.Context, db DB, id int) (*LabBloodType, error) {
	// query
	const sqlstr = `SELECT ` +
		`id, name, description, category, created_at, updated_at ` +
		`FROM public.lab_blood_types ` +
		`WHERE id = $1`
	// run
	logf(sqlstr, id)
	lbt := LabBloodType{
		_exists: true,
	}
	if err := db.QueryRowContext(ctx, sqlstr, id).Scan(&lbt.ID, &lbt.Name, &lbt.Description, &lbt.Category, &lbt.CreatedAt, &lbt.UpdatedAt); err != nil {
		return nil, logerror(err)
	}
	return &lbt, nil
}

// LabBloodTypes retrieves all rows from 'public.lab_blood_types' as a [LabBloodType] slice.
func LabBloodTypes(ctx context.Context, db DB) ([]*LabBloodType, error) {
	// query
	const sqlstr = `SELECT ` +
		`id, name, description, category, created_at, updated_at ` +
		`FROM public.lab_blood_types`
	// run
	logf(sqlstr)
	rows, err := db.QueryContext(ctx, sqlstr)
	if err != nil {
		return nil, logerror(err)
	}
	defer rows.Close()
	// process
	var res []*LabBloodType
	for rows.Next() {
		lbt := LabBloodType{
			_exists: true,
		}
		// scan
		if err := rows.Scan(&lbt.ID, &lbt.Name, &lbt.Description, &lbt.Category, &lbt.CreatedAt, &lbt.UpdatedAt); err != nil {
			return nil, logerror(err)
		}
		res = append(res, &lbt)
	}
	if err := rows.Err(); err != nil {
		return nil, logerror(err)
	}
	return res, nil
}

// LabBloodTypesByCategory retrieves all rows from 'public.lab_blood_types' by category.
func LabBloodTypesByCategory(ctx context.Context, db DB, category string) ([]*LabBloodType, error) {
	// query
	const sqlstr = `SELECT ` +
		`id, name, description, category, created_at, updated_at ` +
		`FROM public.lab_blood_types ` +
		`WHERE category = $1`
	// run
	logf(sqlstr, category)
	rows, err := db.QueryContext(ctx, sqlstr, category)
	if err != nil {
		return nil, logerror(err)
	}
	defer rows.Close()
	// process
	var res []*LabBloodType
	for rows.Next() {
		lbt := LabBloodType{
			_exists: true,
		}
		// scan
		if err := rows.Scan(&lbt.ID, &lbt.Name, &lbt.Description, &lbt.Category, &lbt.CreatedAt, &lbt.UpdatedAt); err != nil {
			return nil, logerror(err)
		}
		res = append(res, &lbt)
	}
	if err := rows.Err(); err != nil {
		return nil, logerror(err)
	}
	return res, nil
}
