package models

import (
	"context"
	"database/sql"
	"time"
)

// LabSwabType represents a row from 'public.lab_swab_types'.
type LabSwabType struct {
	ID          int            `json:"id"`          // id
	Name        string         `json:"name"`        // name
	Description sql.NullString `json:"description"` // description
	CreatedAt   time.Time      `json:"created_at"`  // created_at
	UpdatedAt   time.Time      `json:"updated_at"`  // updated_at
	// xo fields
	_exists, _deleted bool
}

// Exists returns true when the [LabSwabType] exists in the database.
func (lst *LabSwabType) Exists() bool {
	return lst._exists
}

// Deleted returns true when the [LabSwabType] has been marked for deletion
// from the database.
func (lst *LabSwabType) Deleted() bool {
	return lst._deleted
}

// Insert inserts the [LabSwabType] to the database.
func (lst *LabSwabType) Insert(ctx context.Context, db DB) error {
	switch {
	case lst._exists: // already exists
		return logerror(&ErrInsertFailed{ErrAlreadyExists})
	case lst._deleted: // deleted
		return logerror(&ErrInsertFailed{ErrMarkedForDeletion})
	}
	// insert (primary key generated and returned by database)
	const sqlstr = `INSERT INTO public.lab_swab_types (` +
		`name, description, created_at, updated_at` +
		`) VALUES (` +
		`$1, $2, $3, $4` +
		`) RETURNING id`
	// run
	logf(sqlstr, lst.Name, lst.Description, lst.CreatedAt, lst.UpdatedAt)
	if err := db.QueryRowContext(ctx, sqlstr, lst.Name, lst.Description, lst.CreatedAt, lst.UpdatedAt).Scan(&lst.ID); err != nil {
		return logerror(err)
	}
	// set exists
	lst._exists = true
	return nil
}

// Update updates a [LabSwabType] in the database.
func (lst *LabSwabType) Update(ctx context.Context, db DB) error {
	switch {
	case !lst._exists: // doesn't exist
		return logerror(&ErrUpdateFailed{ErrDoesNotExist})
	case lst._deleted: // deleted
		return logerror(&ErrUpdateFailed{ErrMarkedForDeletion})
	}
	// update with composite primary key
	const sqlstr = `UPDATE public.lab_swab_types SET ` +
		`name = $1, description = $2, updated_at = $3` +
		` WHERE id = $4`
	// run
	logf(sqlstr, lst.Name, lst.Description, lst.UpdatedAt, lst.ID)
	if _, err := db.ExecContext(ctx, sqlstr, lst.Name, lst.Description, lst.UpdatedAt, lst.ID); err != nil {
		return logerror(err)
	}
	return nil
}

// Save saves the [LabSwabType] to the database.
func (lst *LabSwabType) Save(ctx context.Context, db DB) error {
	if lst.Exists() {
		return lst.Update(ctx, db)
	}
	return lst.Insert(ctx, db)
}

// Upsert performs an upsert for [LabSwabType].
func (lst *LabSwabType) Upsert(ctx context.Context, db DB) error {
	switch {
	case lst._deleted: // deleted
		return logerror(&ErrUpsertFailed{ErrMarkedForDeletion})
	}
	// upsert
	const sqlstr = `INSERT INTO public.lab_swab_types (` +
		`id, name, description, created_at, updated_at` +
		`) VALUES (` +
		`$1, $2, $3, $4, $5` +
		`)` +
		` ON CONFLICT (id) DO ` +
		`UPDATE SET ` +
		`name = EXCLUDED.name, description = EXCLUDED.description, updated_at = EXCLUDED.updated_at`
	// run
	logf(sqlstr, lst.ID, lst.Name, lst.Description, lst.CreatedAt, lst.UpdatedAt)
	if _, err := db.ExecContext(ctx, sqlstr, lst.ID, lst.Name, lst.Description, lst.CreatedAt, lst.UpdatedAt); err != nil {
		return logerror(err)
	}
	// set exists
	lst._exists = true
	return nil
}

// Delete deletes the [LabSwabType] from the database.
func (lst *LabSwabType) Delete(ctx context.Context, db DB) error {
	switch {
	case !lst._exists: // doesn't exist
		return nil
	case lst._deleted: // deleted
		return nil
	}
	// delete with single primary key
	const sqlstr = `DELETE FROM public.lab_swab_types WHERE id = $1`
	// run
	logf(sqlstr, lst.ID)
	if _, err := db.ExecContext(ctx, sqlstr, lst.ID); err != nil {
		return logerror(err)
	}
	// set deleted
	lst._deleted = true
	return nil
}

// LabSwabTypeByID retrieves a row from 'public.lab_swab_types' as a [LabSwabType].
func LabSwabTypeByID(ctx context.Context, db DB, id int) (*LabSwabType, error) {
	// query
	const sqlstr = `SELECT ` +
		`id, name, description, created_at, updated_at ` +
		`FROM public.lab_swab_types ` +
		`WHERE id = $1`
	// run
	logf(sqlstr, id)
	lst := LabSwabType{
		_exists: true,
	}
	if err := db.QueryRowContext(ctx, sqlstr, id).Scan(&lst.ID, &lst.Name, &lst.Description, &lst.CreatedAt, &lst.UpdatedAt); err != nil {
		return nil, logerror(err)
	}
	return &lst, nil
}

// LabSwabTypes retrieves all rows from 'public.lab_swab_types' as a [LabSwabType] slice.
func LabSwabTypes(ctx context.Context, db DB) ([]*LabSwabType, error) {
	// query
	const sqlstr = `SELECT ` +
		`id, name, description, created_at, updated_at ` +
		`FROM public.lab_swab_types`
	// run
	logf(sqlstr)
	rows, err := db.QueryContext(ctx, sqlstr)
	if err != nil {
		return nil, logerror(err)
	}
	defer rows.Close()
	// process
	var res []*LabSwabType
	for rows.Next() {
		lst := LabSwabType{
			_exists: true,
		}
		// scan
		if err := rows.Scan(&lst.ID, &lst.Name, &lst.Description, &lst.CreatedAt, &lst.UpdatedAt); err != nil {
			return nil, logerror(err)
		}
		res = append(res, &lst)
	}
	if err := rows.Err(); err != nil {
		return nil, logerror(err)
	}
	return res, nil
}
