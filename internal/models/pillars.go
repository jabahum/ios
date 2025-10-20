package models

import (
	"database/sql"
	"time"
)

// Pillar represents a pillar in the health system
type Pillar struct {
	ID              int64          `json:"id" db:"id"`
	Name            string         `json:"name" db:"name"`
	Description     sql.NullString `json:"description" db:"description"`
	PillarHeadID    sql.NullInt64  `json:"pillar_head_id" db:"pillar_head_id"`
	PillarHeadName  sql.NullString `json:"pillar_head_name" db:"pillar_head_name"`
	PillarHeadEmail sql.NullString `json:"pillar_head_email" db:"pillar_head_email"`
	PillarHeadPhone sql.NullString `json:"pillar_head_phone" db:"pillar_head_phone"`
	IsActive        bool           `json:"is_active" db:"is_active"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
	CreatedBy       sql.NullInt64  `json:"created_by" db:"created_by"`
	UpdatedBy       sql.NullInt64  `json:"updated_by" db:"updated_by"`
	// Related data
	PillarHead *EnhancedUser   `json:"pillar_head,omitempty"`
	Changes    []PillarChange  `json:"changes,omitempty"`
	Members    []RRTTeamMember `json:"members,omitempty"`
}

// PillarChange represents a change to a pillar
type PillarChange struct {
	ID           int64          `json:"id" db:"id"`
	PillarID     int64          `json:"pillar_id" db:"pillar_id"`
	ChangeType   string         `json:"change_type" db:"change_type"`
	OldValue     sql.NullString `json:"old_value" db:"old_value"`
	NewValue     sql.NullString `json:"new_value" db:"new_value"`
	ChangeReason sql.NullString `json:"change_reason" db:"change_reason"`
	ChangedBy    sql.NullInt64  `json:"changed_by" db:"changed_by"`
	ChangedAt    time.Time      `json:"changed_at" db:"changed_at"`
	Notes        sql.NullString `json:"notes" db:"notes"`
	// Related data
	Pillar        *Pillar       `json:"pillar,omitempty"`
	ChangedByUser *EnhancedUser `json:"changed_by_user,omitempty"`
}

// Helper methods for Pillar
func (p *Pillar) GetHeadName() string {
	if p.PillarHeadName.Valid {
		return p.PillarHeadName.String
	}
	if p.PillarHead != nil {
		return p.PillarHead.FullName()
	}
	return "Not Assigned"
}

func (p *Pillar) GetHeadEmail() string {
	if p.PillarHeadEmail.Valid {
		return p.PillarHeadEmail.String
	}
	if p.PillarHead != nil {
		return p.PillarHead.GetEmail()
	}
	return ""
}

func (p *Pillar) GetHeadPhone() string {
	if p.PillarHeadPhone.Valid {
		return p.PillarHeadPhone.String
	}
	if p.PillarHead != nil {
		return p.PillarHead.GetPhone()
	}
	return ""
}

// Helper methods for PillarChange
func (pc *PillarChange) GetChangeDescription() string {
	switch pc.ChangeType {
	case "head_change":
		return "Pillar head changed"
	case "status_change":
		return "Pillar status changed"
	case "info_update":
		return "Pillar information updated"
	default:
		return "Pillar changed"
	}
}
