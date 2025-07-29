package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// InventoryHandler handles all inventory-related operations
type InventoryHandler struct {
	db    *sql.DB
	store *session.Store
}

// NewInventoryHandler creates a new inventory handler
func NewInventoryHandler(db *sql.DB, store *session.Store) *InventoryHandler {
	return &InventoryHandler{db: db, store: store}
}

// HandlerInventoryDashboard displays the main inventory dashboard
func (h *InventoryHandler) HandlerInventoryDashboard(c *fiber.Ctx) error {
	data := NewTemplateData(c, h.store)

	// Get summary statistics
	stats, err := h.getInventoryStats()
	if err != nil {
		log.Printf("Error getting inventory stats: %v", err)
		stats = &InventoryStats{}
	}

	data.InventoryStats = stats

	return GenerateHTML(c, h.db, data, "inventory_dashboard")
}

// HandlerInventoryItemsList displays all inventory items
func (h *InventoryHandler) HandlerInventoryItemsList(c *fiber.Ctx) error {
	data := NewTemplateData(c, h.store)

	items, err := h.getAllInventoryItems()
	if err != nil {
		log.Printf("Error getting inventory items: %v", err)
		return c.Status(500).SendString("Error loading inventory items")
	}

	data.InventoryItems = items
	return GenerateHTML(c, h.db, data, "inventory_items_list")
}

// HandlerInventoryItemForm displays the form for adding/editing inventory items
func (h *InventoryHandler) HandlerInventoryItemForm(c *fiber.Ctx) error {
	data := NewTemplateData(c, h.store)

	// Get categories for dropdown
	categories, err := h.getAllCategories()
	if err != nil {
		log.Printf("Error getting categories: %v", err)
		categories = []*InventoryCategory{}
	}

	// Get suppliers for dropdown
	suppliers, err := h.getAllSuppliers()
	if err != nil {
		log.Printf("Error getting suppliers: %v", err)
		suppliers = []*InventorySupplier{}
	}

	// Get treatment sites for dropdown
	sites, err := h.getAllTreatmentSites()
	if err != nil {
		log.Printf("Error getting treatment sites: %v", err)
		sites = []*InventoryTreatmentSite{}
	}

	data.InventoryCategories = categories
	data.InventorySuppliers = suppliers
	data.InventoryTreatmentSites = sites

	// If editing, get the item
	itemID := c.Params("id")
	if itemID != "" {
		item, err := h.getInventoryItemByID(itemID)
		if err != nil {
			log.Printf("Error getting inventory item: %v", err)
		} else {
			data.InventoryItem = item
		}
	}

	return GenerateHTML(c, h.db, data, "inventory_item_form")
}

// HandlerInventoryItemSave saves an inventory item
func (h *InventoryHandler) HandlerInventoryItemSave(c *fiber.Ctx) error {
	itemID := c.FormValue("id")

	item := &InventoryItem{
		Name:        c.FormValue("name"),
		Description: c.FormValue("description"),
		CategoryID:  parseInventoryInt(c.FormValue("category_id")),
		SupplierID:  parseInventoryInt(c.FormValue("supplier_id")),
		Unit:        c.FormValue("unit"),
		MinStock:    parseInventoryFloat(c.FormValue("min_stock")),
		MaxStock:    parseInventoryFloat(c.FormValue("max_stock")),
		UnitCost:    parseInventoryFloat(c.FormValue("unit_cost")),
		Status:      c.FormValue("status"),
	}

	var err error
	if itemID == "" {
		// Create new item
		err = h.createInventoryItem(item)
	} else {
		// Update existing item
		item.ID = parseInventoryInt(itemID)
		err = h.updateInventoryItem(item)
	}

	if err != nil {
		log.Printf("Error saving inventory item: %v", err)
		return c.Status(500).SendString("Error saving inventory item")
	}

	return c.Redirect("/inventory/items")
}

// HandlerInventoryStockForm displays the stock management form
func (h *InventoryHandler) HandlerInventoryStockForm(c *fiber.Ctx) error {
	data := NewTemplateData(c, h.store)

	// Get items for dropdown
	items, err := h.getAllInventoryItems()
	if err != nil {
		log.Printf("Error getting inventory items: %v", err)
		items = []*InventoryItem{}
	}

	// Get treatment sites for dropdown
	sites, err := h.getAllTreatmentSites()
	if err != nil {
		log.Printf("Error getting treatment sites: %v", err)
		sites = []*InventoryTreatmentSite{}
	}

	data.InventoryItems = items
	data.InventoryTreatmentSites = sites

	return GenerateHTML(c, h.db, data, "inventory_stock_form")
}

// HandlerInventoryStockSave saves stock level updates
func (h *InventoryHandler) HandlerInventoryStockSave(c *fiber.Ctx) error {
	transaction := &InventoryTransaction{
		ItemID:          parseInventoryInt(c.FormValue("item_id")),
		SiteID:          parseInventoryInt(c.FormValue("site_id")),
		TransactionType: c.FormValue("transaction_type"),
		Quantity:        parseInventoryFloat(c.FormValue("quantity")),
		UnitCost:        parseInventoryFloat(c.FormValue("unit_cost")),
		Reason:          c.FormValue("reason"),
		Notes:           c.FormValue("notes"),
		TransactionDate: time.Now(),
	}

	err := h.createInventoryTransaction(transaction)
	if err != nil {
		log.Printf("Error saving stock transaction: %v", err)
		return c.Status(500).SendString("Error saving stock transaction")
	}

	return c.Redirect("/inventory/stock")
}

// HandlerInventoryPurchaseOrderForm displays the purchase order form
func (h *InventoryHandler) HandlerInventoryPurchaseOrderForm(c *fiber.Ctx) error {
	data := NewTemplateData(c, h.store)

	// Get suppliers for dropdown
	suppliers, err := h.getAllSuppliers()
	if err != nil {
		log.Printf("Error getting suppliers: %v", err)
		suppliers = []*InventorySupplier{}
	}

	// Get items for dropdown
	items, err := h.getAllInventoryItems()
	if err != nil {
		log.Printf("Error getting inventory items: %v", err)
		items = []*InventoryItem{}
	}

	data.InventorySuppliers = suppliers
	data.InventoryItems = items

	return GenerateHTML(c, h.db, data, "inventory_purchase_order_form")
}

// HandlerInventoryPurchaseOrderSave saves a purchase order
func (h *InventoryHandler) HandlerInventoryPurchaseOrderSave(c *fiber.Ctx) error {
	po := &InventoryPurchaseOrder{
		SupplierID:       parseInventoryInt(c.FormValue("supplier_id")),
		OrderDate:        time.Now(),
		ExpectedDelivery: parseInventoryDate(c.FormValue("expected_delivery")),
		Status:           "pending",
		Notes:            c.FormValue("notes"),
	}

	err := h.createPurchaseOrder(po)
	if err != nil {
		log.Printf("Error saving purchase order: %v", err)
		return c.Status(500).SendString("Error saving purchase order")
	}

	return c.Redirect("/inventory/purchase-orders")
}

// HandlerInventoryRequisitionForm displays the requisition form
func (h *InventoryHandler) HandlerInventoryRequisitionForm(c *fiber.Ctx) error {
	data := NewTemplateData(c, h.store)

	// Get treatment sites for dropdown
	sites, err := h.getAllTreatmentSites()
	if err != nil {
		log.Printf("Error getting treatment sites: %v", err)
		sites = []*InventoryTreatmentSite{}
	}

	// Get items for dropdown
	items, err := h.getAllInventoryItems()
	if err != nil {
		log.Printf("Error getting inventory items: %v", err)
		items = []*InventoryItem{}
	}

	data.InventoryTreatmentSites = sites
	data.InventoryItems = items

	return GenerateHTML(c, h.db, data, "inventory_requisition_form")
}

// HandlerInventoryRequisitionSave saves a requisition
func (h *InventoryHandler) HandlerInventoryRequisitionSave(c *fiber.Ctx) error {
	req := &InventoryRequisition{
		SiteID:      parseInventoryInt(c.FormValue("site_id")),
		RequestDate: time.Now(),
		Priority:    c.FormValue("priority"),
		Status:      "pending",
		Notes:       c.FormValue("notes"),
	}

	err := h.createRequisition(req)
	if err != nil {
		log.Printf("Error saving requisition: %v", err)
		return c.Status(500).SendString("Error saving requisition")
	}

	return c.Redirect("/inventory/requisitions")
}

// HandlerInventoryReports displays inventory reports
func (h *InventoryHandler) HandlerInventoryReports(c *fiber.Ctx) error {
	data := NewTemplateData(c, h.store)

	// Get stock levels report
	stockLevels, err := h.getStockLevelsReport()
	if err != nil {
		log.Printf("Error getting stock levels report: %v", err)
		stockLevels = []*StockLevelReport{}
	}

	// Get transaction history report
	transactions, err := h.getTransactionHistoryReport()
	if err != nil {
		log.Printf("Error getting transaction history report: %v", err)
		transactions = []*TransactionReport{}
	}

	data.StockLevelReports = stockLevels
	data.TransactionReports = transactions

	return GenerateHTML(c, h.db, data, "inventory_reports")
}

// API Handlers for AJAX calls

// HandlerInventoryAPIItems returns inventory items as JSON
func (h *InventoryHandler) HandlerInventoryAPIItems(c *fiber.Ctx) error {
	items, err := h.getAllInventoryItems()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error loading inventory items"})
	}

	return c.JSON(items)
}

// HandlerInventoryAPIStockLevels returns stock levels as JSON
func (h *InventoryHandler) HandlerInventoryAPIStockLevels(c *fiber.Ctx) error {
	siteID := c.Query("site_id")

	levels, err := h.getStockLevelsBySite(siteID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error loading stock levels"})
	}

	return c.JSON(levels)
}

// HandlerInventoryAPILowStock returns low stock alerts as JSON
func (h *InventoryHandler) HandlerInventoryAPILowStock(c *fiber.Ctx) error {
	alerts, err := h.getLowStockAlerts()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error loading low stock alerts"})
	}

	return c.JSON(alerts)
}

// Database helper methods

func (h *InventoryHandler) getInventoryStats() (*InventoryStats, error) {
	query := `
		SELECT 
			COUNT(DISTINCT i.id) as total_items,
			COUNT(DISTINCT sl.id) as total_stock_levels,
			SUM(CASE WHEN sl.quantity <= i.min_stock THEN 1 ELSE 0 END) as low_stock_count,
			SUM(sl.quantity * i.unit_cost) as total_value
		FROM inventory_items i
		LEFT JOIN inventory_stock_levels sl ON i.id = sl.item_id
		WHERE i.status = 'active'
	`

	var stats InventoryStats
	err := h.db.QueryRow(query).Scan(
		&stats.TotalItems,
		&stats.TotalStockLevels,
		&stats.LowStockCount,
		&stats.TotalValue,
	)

	return &stats, err
}

func (h *InventoryHandler) getLowStockAlerts() ([]*InventoryAlert, error) {
	query := `
		SELECT 
			a.id, a.item_id, a.site_id, a.alert_type, a.message, a.created_at, a.status,
			i.name as item_name,
			s.name as site_name
		FROM inventory_alerts a
		JOIN inventory_items i ON a.item_id = i.id
		JOIN inventory_treatment_sites s ON a.site_id = s.id
		WHERE a.status = 'active'
		ORDER BY a.created_at DESC
		LIMIT 10
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []*InventoryAlert
	for rows.Next() {
		var alert InventoryAlert
		err := rows.Scan(
			&alert.ID, &alert.ItemID, &alert.SiteID, &alert.AlertType, &alert.Message,
			&alert.CreatedAt, &alert.Status, &alert.ItemName, &alert.SiteName,
		)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, &alert)
	}

	return alerts, nil
}

func (h *InventoryHandler) getRecentTransactions() ([]*InventoryTransaction, error) {
	query := `
		SELECT 
			t.id, t.item_id, t.site_id, t.transaction_type, t.quantity, t.unit_cost,
			t.reason, t.notes, t.transaction_date,
			i.name as item_name,
			s.name as site_name
		FROM inventory_transactions t
		JOIN inventory_items i ON t.item_id = i.id
		JOIN inventory_treatment_sites s ON t.site_id = s.id
		ORDER BY t.transaction_date DESC
		LIMIT 10
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*InventoryTransaction
	for rows.Next() {
		var trans InventoryTransaction
		err := rows.Scan(
			&trans.ID, &trans.ItemID, &trans.SiteID, &trans.TransactionType,
			&trans.Quantity, &trans.UnitCost, &trans.Reason, &trans.Notes,
			&trans.TransactionDate, &trans.ItemName, &trans.SiteName,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, &trans)
	}

	return transactions, nil
}

func (h *InventoryHandler) getAllInventoryItems() ([]*InventoryItem, error) {
	query := `
		SELECT 
			i.id, i.name, i.description, i.category_id, i.supplier_id, i.unit,
			i.min_stock, i.max_stock, i.unit_cost, i.status, i.created_at,
			c.name as category_name,
			s.name as supplier_name
		FROM inventory_items i
		LEFT JOIN inventory_categories c ON i.category_id = c.id
		LEFT JOIN inventory_suppliers s ON i.supplier_id = s.id
		ORDER BY i.name
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*InventoryItem
	for rows.Next() {
		var item InventoryItem
		err := rows.Scan(
			&item.ID, &item.Name, &item.Description, &item.CategoryID, &item.SupplierID,
			&item.Unit, &item.MinStock, &item.MaxStock, &item.UnitCost, &item.Status,
			&item.CreatedAt, &item.CategoryName, &item.SupplierName,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	return items, nil
}

func (h *InventoryHandler) getInventoryItemByID(id string) (*InventoryItem, error) {
	query := `
		SELECT 
			i.id, i.name, i.description, i.category_id, i.supplier_id, i.unit,
			i.min_stock, i.max_stock, i.unit_cost, i.status, i.created_at,
			c.name as category_name,
			s.name as supplier_name
		FROM inventory_items i
		LEFT JOIN inventory_categories c ON i.category_id = c.id
		LEFT JOIN inventory_suppliers s ON i.supplier_id = s.id
		WHERE i.id = $1
	`

	var item InventoryItem
	err := h.db.QueryRow(query, id).Scan(
		&item.ID, &item.Name, &item.Description, &item.CategoryID, &item.SupplierID,
		&item.Unit, &item.MinStock, &item.MaxStock, &item.UnitCost, &item.Status,
		&item.CreatedAt, &item.CategoryName, &item.SupplierName,
	)

	return &item, err
}

func (h *InventoryHandler) createInventoryItem(item *InventoryItem) error {
	query := `
		INSERT INTO inventory_items (name, description, category_id, supplier_id, unit, min_stock, max_stock, unit_cost, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		RETURNING id
	`

	return h.db.QueryRow(query,
		item.Name, item.Description, item.CategoryID, item.SupplierID,
		item.Unit, item.MinStock, item.MaxStock, item.UnitCost, item.Status,
	).Scan(&item.ID)
}

func (h *InventoryHandler) updateInventoryItem(item *InventoryItem) error {
	query := `
		UPDATE inventory_items 
		SET name = $1, description = $2, category_id = $3, supplier_id = $4, unit = $5,
			min_stock = $6, max_stock = $7, unit_cost = $8, status = $9
		WHERE id = $10
	`

	_, err := h.db.Exec(query,
		item.Name, item.Description, item.CategoryID, item.SupplierID,
		item.Unit, item.MinStock, item.MaxStock, item.UnitCost, item.Status, item.ID,
	)
	return err
}

func (h *InventoryHandler) getAllCategories() ([]*InventoryCategory, error) {
	query := `SELECT id, name, description, status FROM inventory_categories ORDER BY name`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*InventoryCategory
	for rows.Next() {
		var cat InventoryCategory
		err := rows.Scan(&cat.ID, &cat.Name, &cat.Description, &cat.Status)
		if err != nil {
			return nil, err
		}
		categories = append(categories, &cat)
	}

	return categories, nil
}

func (h *InventoryHandler) getAllSuppliers() ([]*InventorySupplier, error) {
	query := `SELECT id, name, contact_person, phone, email, address, status FROM inventory_suppliers ORDER BY name`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suppliers []*InventorySupplier
	for rows.Next() {
		var sup InventorySupplier
		err := rows.Scan(&sup.ID, &sup.Name, &sup.ContactPerson, &sup.Phone, &sup.Email, &sup.Address, &sup.Status)
		if err != nil {
			return nil, err
		}
		suppliers = append(suppliers, &sup)
	}

	return suppliers, nil
}

func (h *InventoryHandler) getAllTreatmentSites() ([]*InventoryTreatmentSite, error) {
	query := `SELECT id, name, location, contact_person, phone, email, status FROM inventory_treatment_sites ORDER BY name`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sites []*InventoryTreatmentSite
	for rows.Next() {
		var site InventoryTreatmentSite
		err := rows.Scan(&site.ID, &site.Name, &site.Location, &site.ContactPerson, &site.Phone, &site.Email, &site.Status)
		if err != nil {
			return nil, err
		}
		sites = append(sites, &site)
	}

	return sites, nil
}

func (h *InventoryHandler) createInventoryTransaction(trans *InventoryTransaction) error {
	query := `
		INSERT INTO inventory_transactions (item_id, site_id, transaction_type, quantity, unit_cost, reason, notes, transaction_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	err := h.db.QueryRow(query,
		trans.ItemID, trans.SiteID, trans.TransactionType, trans.Quantity,
		trans.UnitCost, trans.Reason, trans.Notes, trans.TransactionDate,
	).Scan(&trans.ID)

	if err != nil {
		return err
	}

	// Update stock level
	return h.updateStockLevel(trans.ItemID, trans.SiteID, trans.TransactionType, trans.Quantity)
}

func (h *InventoryHandler) updateStockLevel(itemID, siteID int, transactionType string, quantity float64) error {
	// Get current stock level
	var currentQuantity float64
	query := `SELECT quantity FROM inventory_stock_levels WHERE item_id = $1 AND site_id = $2`
	err := h.db.QueryRow(query, itemID, siteID).Scan(&currentQuantity)

	if err == sql.ErrNoRows {
		// Create new stock level
		if transactionType == "in" {
			insertQuery := `INSERT INTO inventory_stock_levels (item_id, site_id, quantity) VALUES ($1, $2, $3)`
			_, err = h.db.Exec(insertQuery, itemID, siteID, quantity)
		}
		return err
	} else if err != nil {
		return err
	}

	// Update existing stock level
	var newQuantity float64
	switch transactionType {
	case "in":
		newQuantity = currentQuantity + quantity
	case "out":
		newQuantity = currentQuantity - quantity
	default:
		return fmt.Errorf("invalid transaction type: %s", transactionType)
	}

	updateQuery := `UPDATE inventory_stock_levels SET quantity = $1 WHERE item_id = $2 AND site_id = $3`
	_, err = h.db.Exec(updateQuery, newQuantity, itemID, siteID)

	return err
}

func (h *InventoryHandler) createPurchaseOrder(po *InventoryPurchaseOrder) error {
	query := `
		INSERT INTO inventory_purchase_orders (supplier_id, order_date, expected_delivery, status, notes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	return h.db.QueryRow(query,
		po.SupplierID, po.OrderDate, po.ExpectedDelivery, po.Status, po.Notes,
	).Scan(&po.ID)
}

func (h *InventoryHandler) createRequisition(req *InventoryRequisition) error {
	query := `
		INSERT INTO inventory_requisitions (site_id, request_date, priority, status, notes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	return h.db.QueryRow(query,
		req.SiteID, req.RequestDate, req.Priority, req.Status, req.Notes,
	).Scan(&req.ID)
}

func (h *InventoryHandler) getStockLevelsBySite(siteID string) ([]*InventoryStockLevel, error) {
	query := `
		SELECT 
			sl.id, sl.item_id, sl.site_id, sl.quantity, sl.last_updated,
			i.name as item_name,
			s.name as site_name
		FROM inventory_stock_levels sl
		JOIN inventory_items i ON sl.item_id = i.id
		JOIN inventory_treatment_sites s ON sl.site_id = s.id
		WHERE sl.site_id = $1
		ORDER BY i.name
	`

	rows, err := h.db.Query(query, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var levels []*InventoryStockLevel
	for rows.Next() {
		var level InventoryStockLevel
		err := rows.Scan(
			&level.ID, &level.ItemID, &level.SiteID, &level.Quantity, &level.LastUpdated,
			&level.ItemName, &level.SiteName,
		)
		if err != nil {
			return nil, err
		}
		levels = append(levels, &level)
	}

	return levels, nil
}

func (h *InventoryHandler) getStockLevelsReport() ([]*StockLevelReport, error) {
	query := `
		SELECT 
			i.name as item_name,
			c.name as category_name,
			s.name as site_name,
			COALESCE(sl.quantity, 0) as current_stock,
			i.min_stock,
			i.max_stock,
			i.unit_cost,
			(COALESCE(sl.quantity, 0) * i.unit_cost) as total_value
		FROM inventory_items i
		LEFT JOIN inventory_categories c ON i.category_id = c.id
		CROSS JOIN inventory_treatment_sites s
		LEFT JOIN inventory_stock_levels sl ON i.id = sl.item_id AND s.id = sl.site_id
		WHERE i.status = 'active' AND s.status = 'active'
		ORDER BY i.name, s.name
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*StockLevelReport
	for rows.Next() {
		var report StockLevelReport
		err := rows.Scan(
			&report.ItemName, &report.CategoryName, &report.SiteName,
			&report.CurrentStock, &report.MinStock, &report.MaxStock,
			&report.UnitCost, &report.TotalValue,
		)
		if err != nil {
			return nil, err
		}
		reports = append(reports, &report)
	}

	return reports, nil
}

func (h *InventoryHandler) getTransactionHistoryReport() ([]*TransactionReport, error) {
	query := `
		SELECT 
			t.transaction_date,
			i.name as item_name,
			s.name as site_name,
			t.transaction_type,
			t.quantity,
			t.unit_cost,
			(t.quantity * t.unit_cost) as total_cost,
			t.reason,
			t.notes
		FROM inventory_transactions t
		JOIN inventory_items i ON t.item_id = i.id
		JOIN inventory_treatment_sites s ON t.site_id = s.id
		ORDER BY t.transaction_date DESC
		LIMIT 100
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*TransactionReport
	for rows.Next() {
		var report TransactionReport
		err := rows.Scan(
			&report.TransactionDate, &report.ItemName, &report.SiteName,
			&report.TransactionType, &report.Quantity, &report.UnitCost,
			&report.TotalCost, &report.Reason, &report.Notes,
		)
		if err != nil {
			return nil, err
		}
		reports = append(reports, &report)
	}

	return reports, nil
}

// Helper functions for parsing form data
func parseInventoryInt(s string) int {
	if s == "" {
		return 0
	}
	i, _ := strconv.Atoi(s)
	return i
}

func parseInventoryFloat(s string) float64 {
	if s == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func parseInventoryDate(s string) sql.NullTime {
	if s == "" {
		return sql.NullTime{Valid: false}
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: t, Valid: true}
}

// Inventory model structs
type InventoryItem struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	CategoryID   int       `json:"category_id"`
	SupplierID   int       `json:"supplier_id"`
	Unit         string    `json:"unit"`
	MinStock     float64   `json:"min_stock"`
	MaxStock     float64   `json:"max_stock"`
	UnitCost     float64   `json:"unit_cost"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	CategoryName string    `json:"category_name"`
	SupplierName string    `json:"supplier_name"`
}

type InventoryCategory struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type InventorySupplier struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	ContactPerson string `json:"contact_person"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Address       string `json:"address"`
	Status        string `json:"status"`
}

type InventoryTreatmentSite struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Location      string `json:"location"`
	ContactPerson string `json:"contact_person"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Status        string `json:"status"`
}

type InventoryTransaction struct {
	ID              int       `json:"id"`
	ItemID          int       `json:"item_id"`
	SiteID          int       `json:"site_id"`
	TransactionType string    `json:"transaction_type"`
	Quantity        float64   `json:"quantity"`
	UnitCost        float64   `json:"unit_cost"`
	Reason          string    `json:"reason"`
	Notes           string    `json:"notes"`
	TransactionDate time.Time `json:"transaction_date"`
	ItemName        string    `json:"item_name"`
	SiteName        string    `json:"site_name"`
}

type InventoryStockLevel struct {
	ID          int       `json:"id"`
	ItemID      int       `json:"item_id"`
	SiteID      int       `json:"site_id"`
	Quantity    float64   `json:"quantity"`
	LastUpdated time.Time `json:"last_updated"`
	ItemName    string    `json:"item_name"`
	SiteName    string    `json:"site_name"`
}

type InventoryPurchaseOrder struct {
	ID               int          `json:"id"`
	SupplierID       int          `json:"supplier_id"`
	OrderDate        time.Time    `json:"order_date"`
	ExpectedDelivery sql.NullTime `json:"expected_delivery"`
	Status           string       `json:"status"`
	Notes            string       `json:"notes"`
}

type InventoryRequisition struct {
	ID          int       `json:"id"`
	SiteID      int       `json:"site_id"`
	RequestDate time.Time `json:"request_date"`
	Priority    string    `json:"priority"`
	Status      string    `json:"status"`
	Notes       string    `json:"notes"`
}

type InventoryAlert struct {
	ID        int       `json:"id"`
	ItemID    int       `json:"item_id"`
	SiteID    int       `json:"site_id"`
	AlertType string    `json:"alert_type"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"`
	ItemName  string    `json:"item_name"`
	SiteName  string    `json:"site_name"`
}

// Report structs
type InventoryStats struct {
	TotalItems       int     `json:"total_items"`
	TotalStockLevels int     `json:"total_stock_levels"`
	LowStockCount    int     `json:"low_stock_count"`
	TotalValue       float64 `json:"total_value"`
}

type StockLevelReport struct {
	ItemName     string  `json:"item_name"`
	CategoryName string  `json:"category_name"`
	SiteName     string  `json:"site_name"`
	CurrentStock float64 `json:"current_stock"`
	MinStock     float64 `json:"min_stock"`
	MaxStock     float64 `json:"max_stock"`
	UnitCost     float64 `json:"unit_cost"`
	TotalValue   float64 `json:"total_value"`
}

type TransactionReport struct {
	TransactionDate time.Time `json:"transaction_date"`
	ItemName        string    `json:"item_name"`
	SiteName        string    `json:"site_name"`
	TransactionType string    `json:"transaction_type"`
	Quantity        float64   `json:"quantity"`
	UnitCost        float64   `json:"unit_cost"`
	TotalCost       float64   `json:"total_cost"`
	Reason          string    `json:"reason"`
	Notes           string    `json:"notes"`
}
