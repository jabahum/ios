package models

import (
	"database/sql"
	"time"
)

// RRTTeamMember represents a team member in the RRT system
type RRTTeamMember struct {
	ID                  int64          `json:"id" db:"id"`
	FirstName           string         `json:"first_name" db:"first_name"`
	LastName            string         `json:"last_name" db:"last_name"`
	Email               sql.NullString `json:"email" db:"email"`
	Phone               sql.NullString `json:"phone" db:"phone"`
	NationalID          sql.NullString `json:"national_id" db:"national_id"`
	EmployeeID          sql.NullString `json:"employee_id" db:"employee_id"`
	Organization        sql.NullString `json:"organization" db:"organization"`
	Position            sql.NullString `json:"position" db:"position"`
	Qualifications      sql.NullString `json:"qualifications" db:"qualifications"`
	Specializations     []string       `json:"specializations" db:"specializations"`
	Certifications      sql.NullString `json:"certifications" db:"certifications"`
	ExperienceYears     int            `json:"experience_years" db:"experience_years"`
	IsDriver            bool           `json:"is_driver" db:"is_driver"`
	DriverLicense       sql.NullString `json:"driver_license" db:"driver_license"`
	DriverLicenseExpiry sql.NullTime   `json:"driver_license_expiry" db:"driver_license_expiry"`
	IsActive            bool           `json:"is_active" db:"is_active"`
	PillarID            sql.NullInt64  `json:"pillar_id" db:"pillar_id"`
	CreatedAt           time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at" db:"updated_at"`
	CreatedBy           sql.NullInt64  `json:"created_by" db:"created_by"`
	// Related data
	Pillar *Pillar `json:"pillar,omitempty"`
}

// RRTTeamMemberAssignment represents a team member's assignment to a team
type RRTTeamMemberAssignment struct {
	ID              int64          `json:"id" db:"id"`
	TeamID          int64          `json:"team_id" db:"team_id"`
	MemberID        int64          `json:"member_id" db:"member_id"`
	Role            string         `json:"role" db:"role"`
	StartDate       time.Time      `json:"start_date" db:"start_date"`
	EndDate         sql.NullTime   `json:"end_date" db:"end_date"`
	IsActive        bool           `json:"is_active" db:"is_active"`
	AssignmentNotes sql.NullString `json:"assignment_notes" db:"assignment_notes"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	CreatedBy       sql.NullInt64  `json:"created_by" db:"created_by"`
	// Related data
	Team   *RRTTeam       `json:"team,omitempty"`
	Member *RRTTeamMember `json:"member,omitempty"`
}

// RRTDeploymentProposal represents a proposed team deployment
type RRTDeploymentProposal struct {
	ID                      int64          `json:"id" db:"id"`
	ProposalNumber          string         `json:"proposal_number" db:"proposal_number"`
	OutbreakID              int64          `json:"outbreak_id" db:"outbreak_id"`
	ProposedBy              int64          `json:"proposed_by" db:"proposed_by"`
	ProposedDate            time.Time      `json:"proposed_date" db:"proposed_date"`
	DeploymentPurpose       string         `json:"deployment_purpose" db:"deployment_purpose"`
	ProposedTeamComposition sql.NullString `json:"proposed_team_composition" db:"proposed_team_composition"`
	RequiredSkills          []string       `json:"required_skills" db:"required_skills"`
	DeploymentDurationDays  sql.NullInt64  `json:"deployment_duration_days" db:"deployment_duration_days"`
	ExpectedStartDate       sql.NullTime   `json:"expected_start_date" db:"expected_start_date"`
	ExpectedEndDate         sql.NullTime   `json:"expected_end_date" db:"expected_end_date"`
	SpecialRequirements     sql.NullString `json:"special_requirements" db:"special_requirements"`
	Justification           string         `json:"justification" db:"justification"`
	Status                  string         `json:"status" db:"status"`
	ReviewedBy              sql.NullInt64  `json:"reviewed_by" db:"reviewed_by"`
	ReviewedAt              sql.NullTime   `json:"reviewed_at" db:"reviewed_at"`
	ReviewNotes             sql.NullString `json:"review_notes" db:"review_notes"`
	RejectionReason         sql.NullString `json:"rejection_reason" db:"rejection_reason"`
	CreatedAt               time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at" db:"updated_at"`
	// Related data
	Outbreak *Outbreak                     `json:"outbreak,omitempty"`
	Members  []RRTDeploymentProposalMember `json:"members,omitempty"`
}

// RRTDeploymentProposalMember represents a specific member in a deployment proposal
type RRTDeploymentProposalMember struct {
	ID                  int64          `json:"id" db:"id"`
	ProposalID          int64          `json:"proposal_id" db:"proposal_id"`
	MemberID            int64          `json:"member_id" db:"member_id"`
	ProposedRole        string         `json:"proposed_role" db:"proposed_role"`
	IsEssential         bool           `json:"is_essential" db:"is_essential"`
	AlternativeMemberID sql.NullInt64  `json:"alternative_member_id" db:"alternative_member_id"`
	Notes               sql.NullString `json:"notes" db:"notes"`
	CreatedAt           time.Time      `json:"created_at" db:"created_at"`
	// Related data
	Member            *RRTTeamMember `json:"member,omitempty"`
	AlternativeMember *RRTTeamMember `json:"alternative_member,omitempty"`
}

// RRTDeploymentExtension represents a deployment extension request
type RRTDeploymentExtension struct {
	ID              int64          `json:"id" db:"id"`
	DeploymentID    int64          `json:"deployment_id" db:"deployment_id"`
	ExtensionReason string         `json:"extension_reason" db:"extension_reason"`
	OriginalEndDate time.Time      `json:"original_end_date" db:"original_end_date"`
	NewEndDate      time.Time      `json:"new_end_date" db:"new_end_date"`
	RequestedBy     int64          `json:"requested_by" db:"requested_by"`
	RequestedDate   time.Time      `json:"requested_date" db:"requested_date"`
	ApprovedBy      sql.NullInt64  `json:"approved_by" db:"approved_by"`
	ApprovedDate    sql.NullTime   `json:"approved_date" db:"approved_date"`
	Status          string         `json:"status" db:"status"`
	ApprovalNotes   sql.NullString `json:"approval_notes" db:"approval_notes"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	// Related data
	Deployment *RRTDeployment `json:"deployment,omitempty"`
}

// RRTFieldRoleAssignment represents additional roles assigned in the field
type RRTFieldRoleAssignment struct {
	ID               int64          `json:"id" db:"id"`
	DeploymentID     int64          `json:"deployment_id" db:"deployment_id"`
	MemberID         int64          `json:"member_id" db:"member_id"`
	AdditionalRole   string         `json:"additional_role" db:"additional_role"`
	AssignmentDate   time.Time      `json:"assignment_date" db:"assignment_date"`
	EndDate          sql.NullTime   `json:"end_date" db:"end_date"`
	AssignedBy       int64          `json:"assigned_by" db:"assigned_by"`
	AssignmentReason sql.NullString `json:"assignment_reason" db:"assignment_reason"`
	IsActive         bool           `json:"is_active" db:"is_active"`
	CreatedAt        time.Time      `json:"created_at" db:"created_at"`
	// Related data
	Deployment *RRTDeployment `json:"deployment,omitempty"`
	Member     *RRTTeamMember `json:"member,omitempty"`
}

// Helper methods for RRTTeamMember
func (m *RRTTeamMember) FullName() string {
	return m.FirstName + " " + m.LastName
}

func (m *RRTTeamMember) IsAvailableForDeployment() bool {
	return m.IsActive && (m.DriverLicenseExpiry.Valid == false || m.DriverLicenseExpiry.Time.After(time.Now()))
}

// Helper methods for RRTTeamMemberAssignment
func (a *RRTTeamMemberAssignment) IsCurrentAssignment() bool {
	now := time.Now()
	return a.IsActive && a.StartDate.Before(now) && (a.EndDate.Valid == false || a.EndDate.Time.After(now))
}

// Helper methods for RRTDeploymentProposal
func (p *RRTDeploymentProposal) IsPending() bool {
	return p.Status == "pending"
}

func (p *RRTDeploymentProposal) IsApproved() bool {
	return p.Status == "approved"
}

func (p *RRTDeploymentProposal) IsRejected() bool {
	return p.Status == "rejected"
}

// Helper methods for RRTDeploymentExtension
func (e *RRTDeploymentExtension) IsPending() bool {
	return e.Status == "pending"
}

func (e *RRTDeploymentExtension) IsApproved() bool {
	return e.Status == "approved"
}

func (e *RRTDeploymentExtension) IsRejected() bool {
	return e.Status == "rejected"
}
