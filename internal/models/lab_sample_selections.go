package models

import (
	"context"
	"database/sql"
	"time"
)

// LabSampleSelection represents a row from 'public.lab_sample_selections'.
type LabSampleSelection struct {
	ID             int            `json:"id"`               // id
	LabID          int            `json:"lab_id"`           // lab_id
	SampleType     string         `json:"sample_type"`      // sample_type
	SelectedTypeID sql.NullInt64  `json:"selected_type_id"` // selected_type_id
	OtherSpecify   sql.NullString `json:"other_specify"`    // other_specify
	CreatedAt      time.Time      `json:"created_at"`       // created_at
	UpdatedAt      time.Time      `json:"updated_at"`       // updated_at
	// xo fields
	_exists, _deleted bool
}

// Exists returns true when the [LabSampleSelection] exists in the database.
func (lss *LabSampleSelection) Exists() bool {
	return lss._exists
}

// Deleted returns true when the [LabSampleSelection] has been marked for deletion
// from the database.
func (lss *LabSampleSelection) Deleted() bool {
	return lss._deleted
}

// Insert inserts the [LabSampleSelection] to the database.
func (lss *LabSampleSelection) Insert(ctx context.Context, db DB) error {
	switch {
	case lss._exists: // already exists
		return logerror(&ErrInsertFailed{ErrAlreadyExists})
	case lss._deleted: // deleted
		return logerror(&ErrInsertFailed{ErrMarkedForDeletion})
	}
	// insert (primary key generated and returned by database)
	const sqlstr = `INSERT INTO public.lab_sample_selections (` +
		`lab_id, sample_type, selected_type_id, other_specify, created_at, updated_at` +
		`) VALUES (` +
		`$1, $2, $3, $4, $5, $6` +
		`) RETURNING id`
	// run
	logf(sqlstr, lss.LabID, lss.SampleType, lss.SelectedTypeID, lss.OtherSpecify, lss.CreatedAt, lss.UpdatedAt)
	if err := db.QueryRowContext(ctx, sqlstr, lss.LabID, lss.SampleType, lss.SelectedTypeID, lss.OtherSpecify, lss.CreatedAt, lss.UpdatedAt).Scan(&lss.ID); err != nil {
		return logerror(err)
	}
	// set exists
	lss._exists = true
	return nil
}

// Update updates a [LabSampleSelection] in the database.
func (lss *LabSampleSelection) Update(ctx context.Context, db DB) error {
	switch {
	case !lss._exists: // doesn't exist
		return logerror(&ErrUpdateFailed{ErrDoesNotExist})
	case lss._deleted: // deleted
		return logerror(&ErrUpdateFailed{ErrMarkedForDeletion})
	}
	// update with composite primary key
	const sqlstr = `UPDATE public.lab_sample_selections SET ` +
		`lab_id = $1, sample_type = $2, selected_type_id = $3, other_specify = $4, updated_at = $5` +
		` WHERE id = $6`
	// run
	logf(sqlstr, lss.LabID, lss.SampleType, lss.SelectedTypeID, lss.OtherSpecify, lss.UpdatedAt, lss.ID)
	if _, err := db.ExecContext(ctx, sqlstr, lss.LabID, lss.SampleType, lss.SelectedTypeID, lss.OtherSpecify, lss.UpdatedAt, lss.ID); err != nil {
		return logerror(err)
	}
	return nil
}

// Save saves the [LabSampleSelection] to the database.
func (lss *LabSampleSelection) Save(ctx context.Context, db DB) error {
	if lss.Exists() {
		return lss.Update(ctx, db)
	}
	return lss.Insert(ctx, db)
}

// Upsert performs an upsert for [LabSampleSelection].
func (lss *LabSampleSelection) Upsert(ctx context.Context, db DB) error {
	switch {
	case lss._deleted: // deleted
		return logerror(&ErrUpsertFailed{ErrMarkedForDeletion})
	}
	// upsert
	const sqlstr = `INSERT INTO public.lab_sample_selections (` +
		`id, lab_id, sample_type, selected_type_id, other_specify, created_at, updated_at` +
		`) VALUES (` +
		`$1, $2, $3, $4, $5, $6, $7` +
		`)` +
		` ON CONFLICT (id) DO ` +
		`UPDATE SET ` +
		`lab_id = EXCLUDED.lab_id, sample_type = EXCLUDED.sample_type, selected_type_id = EXCLUDED.selected_type_id, other_specify = EXCLUDED.other_specify, updated_at = EXCLUDED.updated_at`
	// run
	logf(sqlstr, lss.ID, lss.LabID, lss.SampleType, lss.SelectedTypeID, lss.OtherSpecify, lss.CreatedAt, lss.UpdatedAt)
	if _, err := db.ExecContext(ctx, sqlstr, lss.ID, lss.LabID, lss.SampleType, lss.SelectedTypeID, lss.OtherSpecify, lss.CreatedAt, lss.UpdatedAt); err != nil {
		return logerror(err)
	}
	// set exists
	lss._exists = true
	return nil
}

// Delete deletes the [LabSampleSelection] from the database.
func (lss *LabSampleSelection) Delete(ctx context.Context, db DB) error {
	switch {
	case !lss._exists: // doesn't exist
		return nil
	case lss._deleted: // deleted
		return nil
	}
	// delete with single primary key
	const sqlstr = `DELETE FROM public.lab_sample_selections WHERE id = $1`
	// run
	logf(sqlstr, lss.ID)
	if _, err := db.ExecContext(ctx, sqlstr, lss.ID); err != nil {
		return logerror(err)
	}
	// set deleted
	lss._deleted = true
	return nil
}

// LabSampleSelectionByID retrieves a row from 'public.lab_sample_selections' as a [LabSampleSelection].
func LabSampleSelectionByID(ctx context.Context, db DB, id int) (*LabSampleSelection, error) {
	// query
	const sqlstr = `SELECT ` +
		`id, lab_id, sample_type, selected_type_id, other_specify, created_at, updated_at ` +
		`FROM public.lab_sample_selections ` +
		`WHERE id = $1`
	// run
	logf(sqlstr, id)
	lss := LabSampleSelection{
		_exists: true,
	}
	if err := db.QueryRowContext(ctx, sqlstr, id).Scan(&lss.ID, &lss.LabID, &lss.SampleType, &lss.SelectedTypeID, &lss.OtherSpecify, &lss.CreatedAt, &lss.UpdatedAt); err != nil {
		return nil, logerror(err)
	}
	return &lss, nil
}

// LabSampleSelectionsByLabID retrieves all rows from 'public.lab_sample_selections' by lab_id.
func LabSampleSelectionsByLabID(ctx context.Context, db DB, labID int) ([]*LabSampleSelection, error) {
	// query
	const sqlstr = `SELECT ` +
		`id, lab_id, sample_type, selected_type_id, other_specify, created_at, updated_at ` +
		`FROM public.lab_sample_selections ` +
		`WHERE lab_id = $1`
	// run
	logf(sqlstr, labID)
	rows, err := db.QueryContext(ctx, sqlstr, labID)
	if err != nil {
		return nil, logerror(err)
	}
	defer rows.Close()
	// process
	var res []*LabSampleSelection
	for rows.Next() {
		lss := LabSampleSelection{
			_exists: true,
		}
		// scan
		if err := rows.Scan(&lss.ID, &lss.LabID, &lss.SampleType, &lss.SelectedTypeID, &lss.OtherSpecify, &lss.CreatedAt, &lss.UpdatedAt); err != nil {
			return nil, logerror(err)
		}
		res = append(res, &lss)
	}
	if err := rows.Err(); err != nil {
		return nil, logerror(err)
	}
	return res, nil
}

// LabSampleSelectionsBySampleType retrieves all rows from 'public.lab_sample_selections' by sample_type.
func LabSampleSelectionsBySampleType(ctx context.Context, db DB, sampleType string) ([]*LabSampleSelection, error) {
	// query
	const sqlstr = `SELECT ` +
		`id, lab_id, sample_type, selected_type_id, other_specify, created_at, updated_at ` +
		`FROM public.lab_sample_selections ` +
		`WHERE sample_type = $1`
	// run
	logf(sqlstr, sampleType)
	rows, err := db.QueryContext(ctx, sqlstr, sampleType)
	if err != nil {
		return nil, logerror(err)
	}
	defer rows.Close()
	// process
	var res []*LabSampleSelection
	for rows.Next() {
		lss := LabSampleSelection{
			_exists: true,
		}
		// scan
		if err := rows.Scan(&lss.ID, &lss.LabID, &lss.SampleType, &lss.SelectedTypeID, &lss.OtherSpecify, &lss.CreatedAt, &lss.UpdatedAt); err != nil {
			return nil, logerror(err)
		}
		res = append(res, &lss)
	}
	if err := rows.Err(); err != nil {
		return nil, logerror(err)
	}
	return res, nil
}
