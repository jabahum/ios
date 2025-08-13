package models

import (
	"context"
	"database/sql"
	"time"
)

// LabUrineType represents a row from 'public.lab_urine_types'.
type LabUrineType struct {
	ID          int            `json:"id"`          // id
	Name        string         `json:"name"`        // name
	Description sql.NullString `json:"description"` // description
	CreatedAt   time.Time      `json:"created_at"`  // created_at
	UpdatedAt   time.Time      `json:"updated_at"`  // updated_at
	// xo fields
	_exists, _deleted bool
}

// Exists returns true when the [LabUrineType] exists in the database.
func (lut *LabUrineType) Exists() bool {
	return lut._exists
}

// Deleted returns true when the [LabUrineType] has been marked for deletion
// from the database.
func (lut *LabUrineType) Deleted() bool {
	return lut._deleted
}

// Insert inserts the [LabUrineType] to the database.
func (lut *LabUrineType) Insert(ctx context.Context, db DB) error {
	switch {
	case lut._exists: // already exists
		return logerror(&ErrInsertFailed{ErrAlreadyExists})
	case lut._deleted: // deleted
		return logerror(&ErrInsertFailed{ErrMarkedForDeletion})
	}
	// insert (primary key generated and returned by database)
	const sqlstr = `INSERT INTO public.lab_urine_types (` +
		`name, description, created_at, updated_at` +
		`) VALUES (` +
		`$1, $2, $3, $4` +
		`) RETURNING id`
	// run
	logf(sqlstr, lut.Name, lut.Description, lut.CreatedAt, lut.UpdatedAt)
	if err := db.QueryRowContext(ctx, sqlstr, lut.Name, lut.Description, lut.CreatedAt, lut.UpdatedAt).Scan(&lut.ID); err != nil {
		return logerror(err)
	}
	// set exists
	lut._exists = true
	return nil
}

// Update updates a [LabUrineType] in the database.
func (lut *LabUrineType) Update(ctx context.Context, db DB) error {
	switch {
	case !lut._exists: // doesn't exist
		return logerror(&ErrUpdateFailed{ErrDoesNotExist})
	case lut._deleted: // deleted
		return logerror(&ErrUpdateFailed{ErrMarkedForDeletion})
	}
	// update with composite primary key
	const sqlstr = `UPDATE public.lab_urine_types SET ` +
		`name = $1, description = $2, updated_at = $3` +
		` WHERE id = $4`
	// run
	logf(sqlstr, lut.Name, lut.Description, lut.UpdatedAt, lut.ID)
	if _, err := db.ExecContext(ctx, sqlstr, lut.Name, lut.Description, lut.UpdatedAt, lut.ID); err != nil {
		return logerror(err)
	}
	return nil
}

// Save saves the [LabUrineType] to the database.
func (lut *LabUrineType) Save(ctx context.Context, db DB) error {
	if lut.Exists() {
		return lut.Update(ctx, db)
	}
	return lut.Insert(ctx, db)
}

// Upsert performs an upsert for [LabUrineType].
func (lut *LabUrineType) Upsert(ctx context.Context, db DB) error {
	switch {
	case lut._deleted: // deleted
		return logerror(&ErrUpsertFailed{ErrMarkedForDeletion})
	}
	// upsert
	const sqlstr = `INSERT INTO public.lab_urine_types (` +
		`id, name, description, created_at, updated_at` +
		`) VALUES (` +
		`$1, $2, $3, $4, $5` +
		`)` +
		` ON CONFLICT (id) DO ` +
		`UPDATE SET ` +
		`name = EXCLUDED.name, description = EXCLUDED.description, updated_at = EXCLUDED.updated_at`
	// run
	logf(sqlstr, lut.ID, lut.Name, lut.Description, lut.CreatedAt, lut.UpdatedAt)
	if _, err := db.ExecContext(ctx, sqlstr, lut.ID, lut.Name, lut.Description, lut.CreatedAt, lut.UpdatedAt); err != nil {
		return logerror(err)
	}
	// set exists
	lut._exists = true
	return nil
}

// Delete deletes the [LabUrineType] from the database.
func (lut *LabUrineType) Delete(ctx context.Context, db DB) error {
	switch {
	case !lut._exists: // doesn't exist
		return nil
	case lut._deleted: // deleted
		return nil
	}
	// delete with single primary key
	const sqlstr = `DELETE FROM public.lab_urine_types WHERE id = $1`
	// run
	logf(sqlstr, lut.ID)
	if _, err := db.ExecContext(ctx, sqlstr, lut.ID); err != nil {
		return logerror(err)
	}
	// set deleted
	lut._deleted = true
	return nil
}

// LabUrineTypeByID retrieves a row from 'public.lab_urine_types' as a [LabUrineType].
func LabUrineTypeByID(ctx context.Context, db DB, id int) (*LabUrineType, error) {
	// query
	const sqlstr = `SELECT ` +
		`id, name, description, created_at, updated_at ` +
		`FROM public.lab_urine_types ` +
		`WHERE id = $1`
	// run
	logf(sqlstr, id)
	lut := LabUrineType{
		_exists: true,
	}
	if err := db.QueryRowContext(ctx, sqlstr, id).Scan(&lut.ID, &lut.Name, &lut.Description, &lut.CreatedAt, &lut.UpdatedAt); err != nil {
		return nil, logerror(err)
	}
	return &lut, nil
}

// LabUrineTypes retrieves all rows from 'public.lab_urine_types' as a [LabUrineType] slice.
func LabUrineTypes(ctx context.Context, db DB) ([]*LabUrineType, error) {
	// query
	const sqlstr = `SELECT ` +
		`id, name, description, created_at, updated_at ` +
		`FROM public.lab_urine_types`
	// run
	logf(sqlstr)
	rows, err := db.QueryContext(ctx, sqlstr)
	if err != nil {
		return nil, logerror(err)
	}
	defer rows.Close()
	// process
	var res []*LabUrineType
	for rows.Next() {
		lut := LabUrineType{
			_exists: true,
		}
		// scan
		if err := rows.Scan(&lut.ID, &lut.Name, &lut.Description, &lut.CreatedAt, &lut.UpdatedAt); err != nil {
			return nil, logerror(err)
		}
		res = append(res, &lut)
	}
	if err := rows.Err(); err != nil {
		return nil, logerror(err)
	}
	return res, nil
}
