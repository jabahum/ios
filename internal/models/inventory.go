package models

import (
	"context"
	"database/sql"
	"time"
)

// InventoryCategory represents a category of inventory items
type InventoryCategory struct {
	ID               int64          `json:"id"`
	Name             string         `json:"name"`
	Description      sql.NullString `json:"description"`
	ParentCategoryID sql.NullInt64  `json:"parent_category_id"`
	IsActive         bool           `json:"is_active"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	CreatedBy        sql.NullInt64  `json:"created_by"`
	UpdatedBy        sql.NullInt64  `json:"updated_by"`
	// Related data
	ParentCategory *InventoryCategory   `json:"parent_category,omitempty"`
	SubCategories  []*InventoryCategory `json:"sub_categories,omitempty"`
}

// InventoryItem represents an inventory item
type InventoryItem struct {
	ID                  int64           `json:"id"`
	Name                string          `json:"name"`
	Description         sql.NullString  `json:"description"`
	CategoryID          int64           `json:"category_id"`
	ItemCode            sql.NullString  `json:"item_code"`
	Barcode             sql.NullString  `json:"barcode"`
	UnitOfMeasure       string          `json:"unit_of_measure"`
	MinimumStockLevel   int             `json:"minimum_stock_level"`
	MaximumStockLevel   sql.NullInt64   `json:"maximum_stock_level"`
	ReorderPoint        int             `json:"reorder_point"`
	UnitCost            sql.NullFloat64 `json:"unit_cost"`
	IsActive            bool            `json:"is_active"`
	IsCritical          bool            `json:"is_critical"`
	RequiresColdStorage bool            `json:"requires_cold_storage"`
	ExpiryTracking      bool            `json:"expiry_tracking"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
	CreatedBy           sql.NullInt64   `json:"created_by"`
	UpdatedBy           sql.NullInt64   `json:"updated_by"`
	// Related data
	Category *InventoryCategory `json:"category,omitempty"`
}

// InventorySupplier represents a supplier/vendor
type InventorySupplier struct {
	ID            int64          `json:"id"`
	Name          string         `json:"name"`
	ContactPerson sql.NullString `json:"contact_person"`
	Email         sql.NullString `json:"email"`
	Phone         sql.NullString `json:"phone"`
	Address       sql.NullString `json:"address"`
	TaxID         sql.NullString `json:"tax_id"`
	PaymentTerms  sql.NullString `json:"payment_terms"`
	IsActive      bool           `json:"is_active"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	CreatedBy     sql.NullInt64  `json:"created_by"`
	UpdatedBy     sql.NullInt64  `json:"updated_by"`
}

// TreatmentSite represents a treatment site/facility
type TreatmentSite struct {
	ID              int64          `json:"id"`
	Name            string         `json:"name"`
	FacilityID      sql.NullInt64  `json:"facility_id"`
	OutbreakID      sql.NullInt64  `json:"outbreak_id"`
	SiteType        sql.NullString `json:"site_type"`
	LocationAddress sql.NullString `json:"location_address"`
	ContactPerson   sql.NullString `json:"contact_person"`
	Phone           sql.NullString `json:"phone"`
	Email           sql.NullString `json:"email"`
	Capacity        sql.NullInt64  `json:"capacity"`
	IsActive        bool           `json:"is_active"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	CreatedBy       sql.NullInt64  `json:"created_by"`
	UpdatedBy       sql.NullInt64  `json:"updated_by"`
	// Related data
	Facility *Facility `json:"facility,omitempty"`
	Outbreak *Outbreak `json:"outbreak,omitempty"`
}

// InventoryStockLevel represents stock levels at a specific site
type InventoryStockLevel struct {
	ID                int64         `json:"id"`
	ItemID            int64         `json:"item_id"`
	SiteID            int64         `json:"site_id"`
	CurrentQuantity   int           `json:"current_quantity"`
	ReservedQuantity  int           `json:"reserved_quantity"`
	AvailableQuantity int           `json:"available_quantity"`
	LastUpdated       time.Time     `json:"last_updated"`
	UpdatedBy         sql.NullInt64 `json:"updated_by"`
	// Related data
	Item *InventoryItem `json:"item,omitempty"`
	Site *TreatmentSite `json:"site,omitempty"`
}

// InventoryTransaction represents an inventory transaction
type InventoryTransaction struct {
	ID              int64           `json:"id"`
	TransactionType string          `json:"transaction_type"` // IN, OUT, ADJUSTMENT, TRANSFER
	ItemID          int64           `json:"item_id"`
	FromSiteID      sql.NullInt64   `json:"from_site_id"`
	ToSiteID        sql.NullInt64   `json:"to_site_id"`
	Quantity        int             `json:"quantity"`
	UnitCost        sql.NullFloat64 `json:"unit_cost"`
	TotalCost       sql.NullFloat64 `json:"total_cost"`
	ReferenceNumber sql.NullString  `json:"reference_number"`
	SupplierID      sql.NullInt64   `json:"supplier_id"`
	TransactionDate time.Time       `json:"transaction_date"`
	ExpiryDate      sql.NullTime    `json:"expiry_date"`
	BatchNumber     sql.NullString  `json:"batch_number"`
	Notes           sql.NullString  `json:"notes"`
	CreatedAt       time.Time       `json:"created_at"`
	CreatedBy       sql.NullInt64   `json:"created_by"`
	// Related data
	Item     *InventoryItem     `json:"item,omitempty"`
	FromSite *TreatmentSite     `json:"from_site,omitempty"`
	ToSite   *TreatmentSite     `json:"to_site,omitempty"`
	Supplier *InventorySupplier `json:"supplier,omitempty"`
}

// InventoryPurchaseOrder represents a purchase order
type InventoryPurchaseOrder struct {
	ID                   int64          `json:"id"`
	PONumber             string         `json:"po_number"`
	SupplierID           int64          `json:"supplier_id"`
	SiteID               int64          `json:"site_id"`
	OutbreakID           sql.NullInt64  `json:"outbreak_id"`
	OrderDate            time.Time      `json:"order_date"`
	ExpectedDeliveryDate sql.NullTime   `json:"expected_delivery_date"`
	Status               string         `json:"status"` // DRAFT, SUBMITTED, APPROVED, RECEIVED, CANCELLED
	TotalAmount          float64        `json:"total_amount"`
	Notes                sql.NullString `json:"notes"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	CreatedBy            sql.NullInt64  `json:"created_by"`
	UpdatedBy            sql.NullInt64  `json:"updated_by"`
	ApprovedBy           sql.NullInt64  `json:"approved_by"`
	ApprovedAt           sql.NullTime   `json:"approved_at"`
	// Related data
	Supplier *InventorySupplier `json:"supplier,omitempty"`
	Site     *TreatmentSite     `json:"site,omitempty"`
	Outbreak *Outbreak          `json:"outbreak,omitempty"`
	Items    []*InventoryPOItem `json:"items,omitempty"`
}

// InventoryPOItem represents an item in a purchase order
type InventoryPOItem struct {
	ID               int64           `json:"id"`
	POID             int64           `json:"po_id"`
	ItemID           int64           `json:"item_id"`
	QuantityOrdered  int             `json:"quantity_ordered"`
	QuantityReceived int             `json:"quantity_received"`
	UnitCost         sql.NullFloat64 `json:"unit_cost"`
	TotalCost        sql.NullFloat64 `json:"total_cost"`
	Notes            sql.NullString  `json:"notes"`
	// Related data
	Item *InventoryItem `json:"item,omitempty"`
}

// InventoryRequisition represents a requisition
type InventoryRequisition struct {
	ID               int64          `json:"id"`
	ReqNumber        string         `json:"req_number"`
	RequestingSiteID int64          `json:"requesting_site_id"`
	ApprovingSiteID  int64          `json:"approving_site_id"`
	OutbreakID       sql.NullInt64  `json:"outbreak_id"`
	RequestDate      time.Time      `json:"request_date"`
	RequiredDate     sql.NullTime   `json:"required_date"`
	Priority         string         `json:"priority"` // LOW, NORMAL, HIGH, URGENT
	Status           string         `json:"status"`   // DRAFT, SUBMITTED, APPROVED, FULFILLED, CANCELLED
	TotalAmount      float64        `json:"total_amount"`
	Notes            sql.NullString `json:"notes"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	CreatedBy        sql.NullInt64  `json:"created_by"`
	UpdatedBy        sql.NullInt64  `json:"updated_by"`
	ApprovedBy       sql.NullInt64  `json:"approved_by"`
	ApprovedAt       sql.NullTime   `json:"approved_at"`
	// Related data
	RequestingSite *TreatmentSite      `json:"requesting_site,omitempty"`
	ApprovingSite  *TreatmentSite      `json:"approving_site,omitempty"`
	Outbreak       *Outbreak           `json:"outbreak,omitempty"`
	Items          []*InventoryReqItem `json:"items,omitempty"`
}

// InventoryReqItem represents an item in a requisition
type InventoryReqItem struct {
	ID                int64           `json:"id"`
	ReqID             int64           `json:"req_id"`
	ItemID            int64           `json:"item_id"`
	QuantityRequested int             `json:"quantity_requested"`
	QuantityApproved  int             `json:"quantity_approved"`
	QuantityIssued    int             `json:"quantity_issued"`
	UnitCost          sql.NullFloat64 `json:"unit_cost"`
	TotalCost         sql.NullFloat64 `json:"total_cost"`
	Notes             sql.NullString  `json:"notes"`
	// Related data
	Item *InventoryItem `json:"item,omitempty"`
}

// InventoryAlert represents an inventory alert
type InventoryAlert struct {
	ID             int64         `json:"id"`
	AlertType      string        `json:"alert_type"` // LOW_STOCK, EXPIRY, OVERSTOCK, REORDER
	ItemID         int64         `json:"item_id"`
	SiteID         int64         `json:"site_id"`
	AlertMessage   string        `json:"alert_message"`
	Severity       string        `json:"severity"` // LOW, MEDIUM, HIGH, CRITICAL
	IsActive       bool          `json:"is_active"`
	IsAcknowledged bool          `json:"is_acknowledged"`
	AcknowledgedBy sql.NullInt64 `json:"acknowledged_by"`
	AcknowledgedAt sql.NullTime  `json:"acknowledged_at"`
	CreatedAt      time.Time     `json:"created_at"`
	// Related data
	Item *InventoryItem `json:"item,omitempty"`
	Site *TreatmentSite `json:"site,omitempty"`
}

// InventoryReport represents an inventory report
type InventoryReport struct {
	ID          int64          `json:"id"`
	ReportType  string         `json:"report_type"`
	SiteID      sql.NullInt64  `json:"site_id"`
	OutbreakID  sql.NullInt64  `json:"outbreak_id"`
	ReportDate  time.Time      `json:"report_date"`
	ReportData  sql.NullString `json:"report_data"` // JSON data
	GeneratedBy sql.NullInt64  `json:"generated_by"`
	GeneratedAt time.Time      `json:"generated_at"`
	// Related data
	Site     *TreatmentSite `json:"site,omitempty"`
	Outbreak *Outbreak      `json:"outbreak,omitempty"`
}

// InventorySetting represents an inventory setting
type InventorySetting struct {
	ID                 int64          `json:"id"`
	SettingKey         string         `json:"setting_key"`
	SettingValue       sql.NullString `json:"setting_value"`
	SettingDescription sql.NullString `json:"setting_description"`
	IsActive           bool           `json:"is_active"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	UpdatedBy          sql.NullInt64  `json:"updated_by"`
}

// InventoryService provides methods for inventory management
type InventoryService struct {
	db *sql.DB
}

// NewInventoryService creates a new inventory service
func NewInventoryService(db *sql.DB) *InventoryService {
	return &InventoryService{db: db}
}

// GetInventoryCategories retrieves all active inventory categories
func (s *InventoryService) GetInventoryCategories(ctx context.Context) ([]*InventoryCategory, error) {
	query := `
		SELECT id, name, description, parent_category_id, is_active, 
		       created_at, updated_at, created_by, updated_by
		FROM inventory_categories 
		WHERE is_active = true 
		ORDER BY name
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*InventoryCategory
	for rows.Next() {
		var cat InventoryCategory
		err := rows.Scan(
			&cat.ID, &cat.Name, &cat.Description, &cat.ParentCategoryID,
			&cat.IsActive, &cat.CreatedAt, &cat.UpdatedAt, &cat.CreatedBy, &cat.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		categories = append(categories, &cat)
	}
	return categories, nil
}

// GetInventoryItems retrieves all active inventory items
func (s *InventoryService) GetInventoryItems(ctx context.Context) ([]*InventoryItem, error) {
	query := `
		SELECT i.id, i.name, i.description, i.category_id, i.item_code, i.barcode,
		       i.unit_of_measure, i.minimum_stock_level, i.maximum_stock_level,
		       i.reorder_point, i.unit_cost, i.is_active, i.is_critical,
		       i.requires_cold_storage, i.expiry_tracking, i.created_at, i.updated_at,
		       i.created_by, i.updated_by
		FROM inventory_items i
		WHERE i.is_active = true
		ORDER BY i.name
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*InventoryItem
	for rows.Next() {
		var item InventoryItem
		err := rows.Scan(
			&item.ID, &item.Name, &item.Description, &item.CategoryID, &item.ItemCode,
			&item.Barcode, &item.UnitOfMeasure, &item.MinimumStockLevel, &item.MaximumStockLevel,
			&item.ReorderPoint, &item.UnitCost, &item.IsActive, &item.IsCritical,
			&item.RequiresColdStorage, &item.ExpiryTracking, &item.CreatedAt, &item.UpdatedAt,
			&item.CreatedBy, &item.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	return items, nil
}

// GetTreatmentSites retrieves all active treatment sites
func (s *InventoryService) GetTreatmentSites(ctx context.Context) ([]*TreatmentSite, error) {
	query := `
		SELECT id, name, facility_id, outbreak_id, site_type, location_address,
		       contact_person, phone, email, capacity, is_active, created_at, updated_at,
		       created_by, updated_by
		FROM treatment_sites
		WHERE is_active = true
		ORDER BY name
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sites []*TreatmentSite
	for rows.Next() {
		var site TreatmentSite
		err := rows.Scan(
			&site.ID, &site.Name, &site.FacilityID, &site.OutbreakID, &site.SiteType,
			&site.LocationAddress, &site.ContactPerson, &site.Phone, &site.Email,
			&site.Capacity, &site.IsActive, &site.CreatedAt, &site.UpdatedAt,
			&site.CreatedBy, &site.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		sites = append(sites, &site)
	}
	return sites, nil
}

// GetStockLevels retrieves stock levels for a specific site
func (s *InventoryService) GetStockLevels(ctx context.Context, siteID int64) ([]*InventoryStockLevel, error) {
	query := `
		SELECT sl.id, sl.item_id, sl.site_id, sl.current_quantity, sl.reserved_quantity,
		       sl.available_quantity, sl.last_updated, sl.updated_by,
		       i.name as item_name, i.item_code, i.unit_of_measure, i.minimum_stock_level,
		       i.is_critical, c.name as category_name
		FROM inventory_stock_levels sl
		JOIN inventory_items i ON sl.item_id = i.id
		JOIN inventory_categories c ON i.category_id = c.id
		WHERE sl.site_id = $1 AND i.is_active = true
		ORDER BY c.name, i.name
	`
	rows, err := s.db.QueryContext(ctx, query, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stockLevels []*InventoryStockLevel
	for rows.Next() {
		var sl InventoryStockLevel
		var itemName, itemCode, unitOfMeasure, categoryName sql.NullString
		var minimumStockLevel int
		var isCritical bool

		err := rows.Scan(
			&sl.ID, &sl.ItemID, &sl.SiteID, &sl.CurrentQuantity, &sl.ReservedQuantity,
			&sl.AvailableQuantity, &sl.LastUpdated, &sl.UpdatedBy,
			&itemName, &itemCode, &unitOfMeasure, &minimumStockLevel, &isCritical, &categoryName,
		)
		if err != nil {
			return nil, err
		}

		sl.Item = &InventoryItem{
			Name:              itemName.String,
			ItemCode:          itemCode,
			UnitOfMeasure:     unitOfMeasure.String,
			MinimumStockLevel: minimumStockLevel,
			IsCritical:        isCritical,
		}
		stockLevels = append(stockLevels, &sl)
	}
	return stockLevels, nil
}

// GetActiveAlerts retrieves active alerts for a site
func (s *InventoryService) GetActiveAlerts(ctx context.Context, siteID int64) ([]*InventoryAlert, error) {
	query := `
		SELECT a.id, a.alert_type, a.item_id, a.site_id, a.alert_message, a.severity,
		       a.is_active, a.is_acknowledged, a.acknowledged_by, a.acknowledged_at, a.created_at,
		       i.name as item_name, i.item_code
		FROM inventory_alerts a
		JOIN inventory_items i ON a.item_id = i.id
		WHERE a.site_id = $1 AND a.is_active = true
		ORDER BY a.severity DESC, a.created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []*InventoryAlert
	for rows.Next() {
		var alert InventoryAlert
		var itemName, itemCode sql.NullString

		err := rows.Scan(
			&alert.ID, &alert.AlertType, &alert.ItemID, &alert.SiteID, &alert.AlertMessage,
			&alert.Severity, &alert.IsActive, &alert.IsAcknowledged, &alert.AcknowledgedBy,
			&alert.AcknowledgedAt, &alert.CreatedAt, &itemName, &itemCode,
		)
		if err != nil {
			return nil, err
		}

		alert.Item = &InventoryItem{
			Name:     itemName.String,
			ItemCode: itemCode,
		}
		alerts = append(alerts, &alert)
	}
	return alerts, nil
}

// CreateTransaction creates a new inventory transaction
func (s *InventoryService) CreateTransaction(ctx context.Context, tx *InventoryTransaction) error {
	query := `
		INSERT INTO inventory_transactions (
			transaction_type, item_id, from_site_id, to_site_id, quantity,
			unit_cost, total_cost, reference_number, supplier_id, transaction_date,
			expiry_date, batch_number, notes, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id
	`

	var id int64
	err := s.db.QueryRowContext(ctx, query,
		tx.TransactionType, tx.ItemID, tx.FromSiteID, tx.ToSiteID, tx.Quantity,
		tx.UnitCost, tx.TotalCost, tx.ReferenceNumber, tx.SupplierID, tx.TransactionDate,
		tx.ExpiryDate, tx.BatchNumber, tx.Notes, tx.CreatedBy,
	).Scan(&id)

	if err != nil {
		return err
	}

	tx.ID = id
	return nil
}

// GetTransactionHistory retrieves transaction history for an item at a site
func (s *InventoryService) GetTransactionHistory(ctx context.Context, itemID, siteID int64, limit int) ([]*InventoryTransaction, error) {
	query := `
		SELECT id, transaction_type, item_id, from_site_id, to_site_id, quantity,
		       unit_cost, total_cost, reference_number, supplier_id, transaction_date,
		       expiry_date, batch_number, notes, created_at, created_by
		FROM inventory_transactions
		WHERE item_id = $1 AND (from_site_id = $2 OR to_site_id = $2)
		ORDER BY transaction_date DESC
		LIMIT $3
	`
	rows, err := s.db.QueryContext(ctx, query, itemID, siteID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*InventoryTransaction
	for rows.Next() {
		var tx InventoryTransaction
		err := rows.Scan(
			&tx.ID, &tx.TransactionType, &tx.ItemID, &tx.FromSiteID, &tx.ToSiteID,
			&tx.Quantity, &tx.UnitCost, &tx.TotalCost, &tx.ReferenceNumber, &tx.SupplierID,
			&tx.TransactionDate, &tx.ExpiryDate, &tx.BatchNumber, &tx.Notes,
			&tx.CreatedAt, &tx.CreatedBy,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, &tx)
	}
	return transactions, nil
}
