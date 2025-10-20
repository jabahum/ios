package models

import (
	"context"
	"database/sql"
	"time"
)

// RRTTeam represents a Rapid Response Team
type RRTTeam struct {
	ID              int64          `json:"id" db:"id"`
	TeamName        string         `json:"team_name" db:"team_name"`
	TeamCode        string         `json:"team_code" db:"team_code"`
	TeamType        string         `json:"team_type" db:"team_type"`
	TeamLeadName    string         `json:"team_lead_name" db:"team_lead_name"`
	TeamLeadPhone   sql.NullString `json:"team_lead_phone" db:"team_lead_phone"`
	TeamLeadEmail   sql.NullString `json:"team_lead_email" db:"team_lead_email"`
	TeamSize        int            `json:"team_size" db:"team_size"`
	Specializations []string       `json:"specializations" db:"specializations"`
	BaseLocation    sql.NullString `json:"base_location" db:"base_location"`
	IsActive        bool           `json:"is_active" db:"is_active"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
	CreatedBy       sql.NullInt64  `json:"created_by" db:"created_by"`
}

// RRTDeployment represents a team deployment to an outbreak
type RRTDeployment struct {
	ID                 int64          `json:"id" db:"id"`
	TeamID             int64          `json:"team_id" db:"team_id"`
	OutbreakID         int64          `json:"outbreak_id" db:"outbreak_id"`
	DeploymentDate     time.Time      `json:"deployment_date" db:"deployment_date"`
	ExpectedReturnDate sql.NullTime   `json:"expected_return_date" db:"expected_return_date"`
	ActualReturnDate   sql.NullTime   `json:"actual_return_date" db:"actual_return_date"`
	DeploymentStatus   string         `json:"deployment_status" db:"deployment_status"`
	DeploymentPurpose  sql.NullString `json:"deployment_purpose" db:"deployment_purpose"`
	AssignedVehicle    sql.NullString `json:"assigned_vehicle" db:"assigned_vehicle"`
	AssignedDriver     sql.NullString `json:"assigned_driver" db:"assigned_driver"`
	DeploymentNotes    sql.NullString `json:"deployment_notes" db:"deployment_notes"`
	CreatedAt          time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at" db:"updated_at"`
	CreatedBy          sql.NullInt64  `json:"created_by" db:"created_by"`
	// Related data
	Team     *RRTTeam  `json:"team,omitempty"`
	Outbreak *Outbreak `json:"outbreak,omitempty"`
}

// ResourceCategory represents categories for resources
type ResourceCategory struct {
	ID           int64          `json:"id" db:"id"`
	Name         string         `json:"name" db:"name"`
	Description  sql.NullString `json:"description" db:"description"`
	CategoryType string         `json:"category_type" db:"category_type"`
	IsActive     bool           `json:"is_active" db:"is_active"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
	CreatedBy    sql.NullInt64  `json:"created_by" db:"created_by"`
}

// Resource represents a resource item
type Resource struct {
	ID            int64          `json:"id" db:"id"`
	Name          string         `json:"name" db:"name"`
	Description   sql.NullString `json:"description" db:"description"`
	ResourceCode  sql.NullString `json:"resource_code" db:"resource_code"`
	CategoryID    int64          `json:"category_id" db:"category_id"`
	UnitOfMeasure string         `json:"unit_of_measure" db:"unit_of_measure"`
	IsConsumable  bool           `json:"is_consumable" db:"is_consumable"`
	HasExpiry     bool           `json:"has_expiry" db:"has_expiry"`
	ShelfLifeDays sql.NullInt64  `json:"shelf_life_days" db:"shelf_life_days"`
	IsCritical    bool           `json:"is_critical" db:"is_critical"`
	IsActive      bool           `json:"is_active" db:"is_active"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at" db:"updated_at"`
	CreatedBy     sql.NullInt64  `json:"created_by" db:"created_by"`
	// Related data
	Category *ResourceCategory `json:"category,omitempty"`
}

// StorageLocation represents storage facilities
type StorageLocation struct {
	ID                  int64          `json:"id" db:"id"`
	LocationName        string         `json:"location_name" db:"location_name"`
	LocationCode        sql.NullString `json:"location_code" db:"location_code"`
	LocationType        string         `json:"location_type" db:"location_type"`
	Address             sql.NullString `json:"address" db:"address"`
	ContactPerson       sql.NullString `json:"contact_person" db:"contact_person"`
	ContactPhone        sql.NullString `json:"contact_phone" db:"contact_phone"`
	ContactEmail        sql.NullString `json:"contact_email" db:"contact_email"`
	CapacityDescription sql.NullString `json:"capacity_description" db:"capacity_description"`
	IsActive            bool           `json:"is_active" db:"is_active"`
	CreatedAt           time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at" db:"updated_at"`
	CreatedBy           sql.NullInt64  `json:"created_by" db:"created_by"`
}

// Donor represents source entities for resources
type Donor struct {
	ID            int64          `json:"id" db:"id"`
	DonorName     string         `json:"donor_name" db:"donor_name"`
	DonorType     string         `json:"donor_type" db:"donor_type"`
	ContactPerson sql.NullString `json:"contact_person" db:"contact_person"`
	ContactPhone  sql.NullString `json:"contact_phone" db:"contact_phone"`
	ContactEmail  sql.NullString `json:"contact_email" db:"contact_email"`
	Address       sql.NullString `json:"address" db:"address"`
	IsActive      bool           `json:"is_active" db:"is_active"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at" db:"updated_at"`
	CreatedBy     sql.NullInt64  `json:"created_by" db:"created_by"`
}

// StockLedgerEntry represents a stock movement entry
type StockLedgerEntry struct {
	ID              int64           `json:"id" db:"id"`
	ResourceID      int64           `json:"resource_id" db:"resource_id"`
	LocationID      int64           `json:"location_id" db:"location_id"`
	TransactionType string          `json:"transaction_type" db:"transaction_type"`
	Quantity        float64         `json:"quantity" db:"quantity"`
	UnitCost        sql.NullFloat64 `json:"unit_cost" db:"unit_cost"`
	DonorID         sql.NullInt64   `json:"donor_id" db:"donor_id"`
	OutbreakID      sql.NullInt64   `json:"outbreak_id" db:"outbreak_id"`
	DeploymentID    sql.NullInt64   `json:"deployment_id" db:"deployment_id"`
	ReferenceNumber sql.NullString  `json:"reference_number" db:"reference_number"`
	TransactionDate time.Time       `json:"transaction_date" db:"transaction_date"`
	ExpiryDate      sql.NullTime    `json:"expiry_date" db:"expiry_date"`
	BatchNumber     sql.NullString  `json:"batch_number" db:"batch_number"`
	Notes           sql.NullString  `json:"notes" db:"notes"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	CreatedBy       sql.NullInt64   `json:"created_by" db:"created_by"`
	// Related data
	Resource   *Resource        `json:"resource,omitempty"`
	Location   *StorageLocation `json:"location,omitempty"`
	Donor      *Donor           `json:"donor,omitempty"`
	Outbreak   *Outbreak        `json:"outbreak,omitempty"`
	Deployment *RRTDeployment   `json:"deployment,omitempty"`
}

// Requisition represents a resource requisition request
type Requisition struct {
	ID                int64          `json:"id" db:"id"`
	RequisitionNumber string         `json:"requisition_number" db:"requisition_number"`
	OutbreakID        int64          `json:"outbreak_id" db:"outbreak_id"`
	DeploymentID      sql.NullInt64  `json:"deployment_id" db:"deployment_id"`
	RequestedBy       int64          `json:"requested_by" db:"requested_by"`
	RequestedDate     time.Time      `json:"requested_date" db:"requested_date"`
	RequiredDate      sql.NullTime   `json:"required_date" db:"required_date"`
	Priority          string         `json:"priority" db:"priority"`
	Status            string         `json:"status" db:"status"`
	ApprovedBy        sql.NullInt64  `json:"approved_by" db:"approved_by"`
	ApprovedDate      sql.NullTime   `json:"approved_date" db:"approved_date"`
	RejectionReason   sql.NullString `json:"rejection_reason" db:"rejection_reason"`
	DispatchDate      sql.NullTime   `json:"dispatch_date" db:"dispatch_date"`
	ReceivedDate      sql.NullTime   `json:"received_date" db:"received_date"`
	Notes             sql.NullString `json:"notes" db:"notes"`
	CreatedAt         time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at" db:"updated_at"`
	// Related data
	Outbreak   *Outbreak         `json:"outbreak,omitempty"`
	Deployment *RRTDeployment    `json:"deployment,omitempty"`
	Items      []RequisitionItem `json:"items,omitempty"`
}

// RequisitionItem represents items in a requisition
type RequisitionItem struct {
	ID                 int64           `json:"id" db:"id"`
	RequisitionID      int64           `json:"requisition_id" db:"requisition_id"`
	ResourceID         int64           `json:"resource_id" db:"resource_id"`
	QuantityRequested  float64         `json:"quantity_requested" db:"quantity_requested"`
	QuantityApproved   sql.NullFloat64 `json:"quantity_approved" db:"quantity_approved"`
	QuantityDispatched sql.NullFloat64 `json:"quantity_dispatched" db:"quantity_dispatched"`
	QuantityReceived   sql.NullFloat64 `json:"quantity_received" db:"quantity_received"`
	UnitCost           sql.NullFloat64 `json:"unit_cost" db:"unit_cost"`
	Notes              sql.NullString  `json:"notes" db:"notes"`
	CreatedAt          time.Time       `json:"created_at" db:"created_at"`
	// Related data
	Resource *Resource `json:"resource,omitempty"`
}

// Dispatch represents a resource dispatch
type Dispatch struct {
	ID                     int64          `json:"id" db:"id"`
	DispatchNumber         string         `json:"dispatch_number" db:"dispatch_number"`
	RequisitionID          int64          `json:"requisition_id" db:"requisition_id"`
	FromLocationID         int64          `json:"from_location_id" db:"from_location_id"`
	ToLocationID           sql.NullInt64  `json:"to_location_id" db:"to_location_id"`
	ToDeploymentID         sql.NullInt64  `json:"to_deployment_id" db:"to_deployment_id"`
	DispatchedBy           int64          `json:"dispatched_by" db:"dispatched_by"`
	DispatchDate           time.Time      `json:"dispatch_date" db:"dispatch_date"`
	ExpectedDeliveryDate   sql.NullTime   `json:"expected_delivery_date" db:"expected_delivery_date"`
	ActualDeliveryDate     sql.NullTime   `json:"actual_delivery_date" db:"actual_delivery_date"`
	TransportMethod        sql.NullString `json:"transport_method" db:"transport_method"`
	VehicleDetails         sql.NullString `json:"vehicle_details" db:"vehicle_details"`
	DriverName             sql.NullString `json:"driver_name" db:"driver_name"`
	DriverPhone            sql.NullString `json:"driver_phone" db:"driver_phone"`
	DeliveryNotes          sql.NullString `json:"delivery_notes" db:"delivery_notes"`
	AcknowledgmentReceived bool           `json:"acknowledgment_received" db:"acknowledgment_received"`
	AcknowledgmentDate     sql.NullTime   `json:"acknowledgment_date" db:"acknowledgment_date"`
	Status                 string         `json:"status" db:"status"`
	CreatedAt              time.Time      `json:"created_at" db:"created_at"`
	// Related data
	Requisition *Requisition   `json:"requisition,omitempty"`
	Items       []DispatchItem `json:"items,omitempty"`
}

// DispatchItem represents items in a dispatch
type DispatchItem struct {
	ID                 int64           `json:"id" db:"id"`
	DispatchID         int64           `json:"dispatch_id" db:"dispatch_id"`
	ResourceID         int64           `json:"resource_id" db:"resource_id"`
	QuantityDispatched float64         `json:"quantity_dispatched" db:"quantity_dispatched"`
	BatchNumber        sql.NullString  `json:"batch_number" db:"batch_number"`
	ExpiryDate         sql.NullTime    `json:"expiry_date" db:"expiry_date"`
	UnitCost           sql.NullFloat64 `json:"unit_cost" db:"unit_cost"`
	Notes              sql.NullString  `json:"notes" db:"notes"`
	CreatedAt          time.Time       `json:"created_at" db:"created_at"`
	// Related data
	Resource *Resource `json:"resource,omitempty"`
}

// ActivityLog represents RRT activity logging
type ActivityLog struct {
	ID                  int64          `json:"id" db:"id"`
	DeploymentID        int64          `json:"deployment_id" db:"deployment_id"`
	ActivityType        string         `json:"activity_type" db:"activity_type"`
	ActivityDate        time.Time      `json:"activity_date" db:"activity_date"`
	StartTime           sql.NullTime   `json:"start_time" db:"start_time"`
	EndTime             sql.NullTime   `json:"end_time" db:"end_time"`
	Location            sql.NullString `json:"location" db:"location"`
	ParticipantsCount   sql.NullInt64  `json:"participants_count" db:"participants_count"`
	ActivityDescription sql.NullString `json:"activity_description" db:"activity_description"`
	Outcomes            sql.NullString `json:"outcomes" db:"outcomes"`
	Challenges          sql.NullString `json:"challenges" db:"challenges"`
	Recommendations     sql.NullString `json:"recommendations" db:"recommendations"`
	ResourcesUsed       sql.NullString `json:"resources_used" db:"resources_used"`
	CreatedAt           time.Time      `json:"created_at" db:"created_at"`
	CreatedBy           sql.NullInt64  `json:"created_by" db:"created_by"`
	// Related data
	Deployment   *RRTDeployment        `json:"deployment,omitempty"`
	Participants []ActivityParticipant `json:"participants,omitempty"`
}

// ActivityParticipant represents participants in activities
type ActivityParticipant struct {
	ID              int64          `json:"id" db:"id"`
	ActivityID      int64          `json:"activity_id" db:"activity_id"`
	ParticipantName string         `json:"participant_name" db:"participant_name"`
	ParticipantType sql.NullString `json:"participant_type" db:"participant_type"`
	Organization    sql.NullString `json:"organization" db:"organization"`
	ContactPhone    sql.NullString `json:"contact_phone" db:"contact_phone"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
}

// GeneratedSitRep represents auto-generated situation reports
type GeneratedSitRep struct {
	ID                int64          `json:"id" db:"id"`
	OutbreakID        int64          `json:"outbreak_id" db:"outbreak_id"`
	ReportPeriodStart time.Time      `json:"report_period_start" db:"report_period_start"`
	ReportPeriodEnd   time.Time      `json:"report_period_end" db:"report_period_end"`
	ReportType        string         `json:"report_type" db:"report_type"`
	ReportTitle       sql.NullString `json:"report_title" db:"report_title"`
	ReportContent     sql.NullString `json:"report_content" db:"report_content"`
	GeneratedBy       sql.NullInt64  `json:"generated_by" db:"generated_by"`
	GeneratedAt       time.Time      `json:"generated_at" db:"generated_at"`
	ApprovedBy        sql.NullInt64  `json:"approved_by" db:"approved_by"`
	ApprovedAt        sql.NullTime   `json:"approved_at" db:"approved_at"`
	Status            string         `json:"status" db:"status"`
	FilePath          sql.NullString `json:"file_path" db:"file_path"`
	CreatedAt         time.Time      `json:"created_at" db:"created_at"`
	// Related data
	Outbreak *Outbreak `json:"outbreak,omitempty"`
}

// OutbreakClosureCriteria represents criteria for outbreak closure
type OutbreakClosureCriteria struct {
	ID                   int64          `json:"id" db:"id"`
	OutbreakID           int64          `json:"outbreak_id" db:"outbreak_id"`
	NoNewCasesDays       int            `json:"no_new_cases_days" db:"no_new_cases_days"`
	AllTeamsDemobilized  bool           `json:"all_teams_demobilized" db:"all_teams_demobilized"`
	AllItemsReturned     bool           `json:"all_items_returned" db:"all_items_returned"`
	FinalSitRepGenerated bool           `json:"final_sitrep_generated" db:"final_sitrep_generated"`
	ClosureApproved      bool           `json:"closure_approved" db:"closure_approved"`
	ClosureDate          sql.NullTime   `json:"closure_date" db:"closure_date"`
	ClosureNotes         sql.NullString `json:"closure_notes" db:"closure_notes"`
	CreatedAt            time.Time      `json:"created_at" db:"created_at"`
	CreatedBy            sql.NullInt64  `json:"created_by" db:"created_by"`
	// Related data
	Outbreak *Outbreak `json:"outbreak,omitempty"`
}

// Resource Management Service for business logic
type ResourceManagementService struct {
	db *sql.DB
}

// NewResourceManagementService creates a new service instance
func NewResourceManagementService(db *sql.DB) *ResourceManagementService {
	return &ResourceManagementService{db: db}
}

// GetRRTTeams retrieves all active RRT teams
func (s *ResourceManagementService) GetRRTTeams(ctx context.Context) ([]*RRTTeam, error) {
	query := `SELECT id, team_name, team_code, team_type, team_lead_name, team_lead_phone, 
	          team_lead_email, team_size, specializations, base_location, is_active, 
	          created_at, updated_at, created_by 
	          FROM rrt_teams WHERE is_active = true ORDER BY team_name`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*RRTTeam
	for rows.Next() {
		var team RRTTeam
		err := rows.Scan(
			&team.ID, &team.TeamName, &team.TeamCode, &team.TeamType,
			&team.TeamLeadName, &team.TeamLeadPhone, &team.TeamLeadEmail,
			&team.TeamSize, &team.Specializations, &team.BaseLocation,
			&team.IsActive, &team.CreatedAt, &team.UpdatedAt, &team.CreatedBy,
		)
		if err != nil {
			return nil, err
		}
		teams = append(teams, &team)
	}
	return teams, nil
}

// GetActiveDeployments retrieves all active deployments for an outbreak
func (s *ResourceManagementService) GetActiveDeployments(ctx context.Context, outbreakID int64) ([]*RRTDeployment, error) {
	query := `SELECT d.id, d.team_id, d.outbreak_id, d.deployment_date, d.expected_return_date,
	          d.actual_return_date, d.deployment_status, d.deployment_purpose, d.assigned_vehicle,
	          d.assigned_driver, d.deployment_notes, d.created_at, d.updated_at, d.created_by,
	          t.team_name, t.team_code, t.team_type, t.team_lead_name
	          FROM rrt_deployments d
	          JOIN rrt_teams t ON d.team_id = t.id
	          WHERE d.outbreak_id = $1 AND d.deployment_status IN ('deployed', 'extended')
	          ORDER BY d.deployment_date DESC`

	rows, err := s.db.QueryContext(ctx, query, outbreakID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []*RRTDeployment
	for rows.Next() {
		var deployment RRTDeployment
		var team RRTTeam
		err := rows.Scan(
			&deployment.ID, &deployment.TeamID, &deployment.OutbreakID,
			&deployment.DeploymentDate, &deployment.ExpectedReturnDate,
			&deployment.ActualReturnDate, &deployment.DeploymentStatus,
			&deployment.DeploymentPurpose, &deployment.AssignedVehicle,
			&deployment.AssignedDriver, &deployment.DeploymentNotes,
			&deployment.CreatedAt, &deployment.UpdatedAt, &deployment.CreatedBy,
			&team.TeamName, &team.TeamCode, &team.TeamType, &team.TeamLeadName,
		)
		if err != nil {
			return nil, err
		}
		deployment.Team = &team
		deployments = append(deployments, &deployment)
	}
	return deployments, nil
}

// GetResourceStockLevel retrieves current stock level for a resource at a location
func (s *ResourceManagementService) GetResourceStockLevel(ctx context.Context, resourceID, locationID int64) (float64, error) {
	query := `
		SELECT COALESCE(SUM(
			CASE 
				WHEN transaction_type IN ('received', 'returned') THEN quantity
				WHEN transaction_type IN ('allocated', 'used', 'transferred', 'expired', 'damaged') THEN -quantity
				ELSE 0
			END
		), 0) as current_stock
		FROM stock_ledger 
		WHERE resource_id = $1 AND location_id = $2`

	var stock float64
	err := s.db.QueryRowContext(ctx, query, resourceID, locationID).Scan(&stock)
	return stock, err
}

// GetOutbreakResourceSummary retrieves resource utilization summary for an outbreak
func (s *ResourceManagementService) GetOutbreakResourceSummary(ctx context.Context, outbreakID int64) (map[string]interface{}, error) {
	query := `
		SELECT 
			r.name as resource_name,
			rc.name as category_name,
			SUM(sl.quantity) as total_allocated,
			COUNT(DISTINCT sl.deployment_id) as deployments_count
		FROM stock_ledger sl
		JOIN resources r ON sl.resource_id = r.id
		JOIN resource_categories rc ON r.category_id = rc.id
		WHERE sl.outbreak_id = $1 AND sl.transaction_type = 'allocated'
		GROUP BY r.name, rc.name
		ORDER BY total_allocated DESC`

	rows, err := s.db.QueryContext(ctx, query, outbreakID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summary := make(map[string]interface{})
	var resources []map[string]interface{}

	for rows.Next() {
		var resourceName, categoryName string
		var totalAllocated float64
		var deploymentsCount int

		err := rows.Scan(&resourceName, &categoryName, &totalAllocated, &deploymentsCount)
		if err != nil {
			return nil, err
		}

		resources = append(resources, map[string]interface{}{
			"resource_name":     resourceName,
			"category_name":     categoryName,
			"total_allocated":   totalAllocated,
			"deployments_count": deploymentsCount,
		})
	}

	summary["resources"] = resources
	return summary, nil
}
