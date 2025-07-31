package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
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
	log.Printf("DEBUG: HandlerInventoryDashboard called - START")
	defer log.Printf("DEBUG: HandlerInventoryDashboard called - END")

	data := NewTemplateDataWithDB(c, h.store, h.db)
	log.Printf("DEBUG: TemplateData created, user authenticated: %v", data.IsAuthenticated)

	// Check if user has inventory access (temporarily disabled for public access)
	/*
		if !h.hasInventoryAccess(c) {
			return c.Status(403).SendString("Access denied. You don't have permission to access inventory management.")
		}
	*/

	// Get summary statistics
	stats, err := h.getInventoryStats()
	if err != nil {
		log.Printf("Error getting inventory stats: %v", err)
		stats = &InventoryStats{}
	}

	data.InventoryStats = stats
	log.Printf("DEBUG: About to render inventory_dashboard template")

	return GenerateHTML(c, h.db, data, "inventory_dashboard")
}

// hasInventoryAccess checks if the current user has permission to access inventory
func (h *InventoryHandler) hasInventoryAccess(c *fiber.Ctx) bool {
	// Get user session
	sess, err := h.store.Get(c)
	if err != nil {
		log.Printf("Error getting session: %v", err)
		return false
	}

	userID := sess.Get("user_id")
	if userID == nil {
		log.Printf("No user_id in session")
		return false
	}

	// Get user's primary role
	role, err := h.getUserPrimaryRole(userID.(int))
	if err != nil {
		log.Printf("Error getting user role: %v", err)
		return false
	}

	// Define roles that have inventory access
	inventoryRoles := []string{
		"super_admin",
		"admin",
		"outbreak_manager",
		"case_manager",
		"outbreak_viewer",
		"inventory_manager",
		"logistics_coordinator",
	}

	// Check if user's role is in the allowed list
	for _, allowedRole := range inventoryRoles {
		if role == allowedRole {
			log.Printf("User %d with role %s has inventory access", userID, role)
			return true
		}
	}

	log.Printf("User %d with role %s denied inventory access", userID, role)
	return false
}

// getUserPrimaryRole gets the primary role for a user
func (h *InventoryHandler) getUserPrimaryRole(userID int) (string, error) {
	var role string
	query := `
		SELECT r.name 
		FROM user_roles ur 
		JOIN roles r ON ur.role_id = r.id 
		WHERE ur.user_id = $1 
		ORDER BY r.priority ASC 
		LIMIT 1
	`
	err := h.db.QueryRow(query, userID).Scan(&role)
	return role, err
}

// HandlerInventoryItemsList displays all inventory items
func (h *InventoryHandler) HandlerInventoryItemsList(c *fiber.Ctx) error {
	// Check if user has inventory access
	if !h.hasInventoryAccess(c) {
		return c.Status(403).SendString("Access denied. You don't have permission to access inventory management.")
	}

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
	_, err = h.getAllTreatmentSites()
	if err != nil {
		log.Printf("Error getting treatment sites: %v", err)
	}

	data.InventoryCategories = categories
	data.InventorySuppliers = suppliers
	// Note: TreatmentSites not available in TemplateData yet

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
	_, err = h.getAllTreatmentSites()
	if err != nil {
		log.Printf("Error getting treatment sites: %v", err)
	}

	data.InventoryItems = items
	// Note: TreatmentSites not available in TemplateData yet

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
	_, err := h.getAllTreatmentSites()
	if err != nil {
		log.Printf("Error getting treatment sites: %v", err)
	}

	// Get items for dropdown
	items, err := h.getAllInventoryItems()
	if err != nil {
		log.Printf("Error getting inventory items: %v", err)
		items = []*InventoryItem{}
	}

	// Note: TreatmentSites not available in TemplateData yet
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

// ==================== DONATION MANAGEMENT HANDLERS ====================

// HandlerDonationsList displays all donations
func (h *InventoryHandler) HandlerDonationsList(c *fiber.Ctx) error {
	// Check if user has inventory access
	if !h.hasInventoryAccess(c) {
		return c.Status(403).SendString("Access denied. You don't have permission to access inventory management.")
	}

	data := NewTemplateData(c, h.store)

	donations, err := h.getAllDonations()
	if err != nil {
		log.Printf("Error getting donations: %v", err)
		return c.Status(500).SendString("Error loading donations")
	}

	data.Donations = donations
	return GenerateHTML(c, h.db, data, "inventory_donations_list")
}

// HandlerDonationForm displays the donation form
func (h *InventoryHandler) HandlerDonationForm(c *fiber.Ctx) error {
	// Check if user has inventory access
	if !h.hasInventoryAccess(c) {
		return c.Status(403).SendString("Access denied. You don't have permission to access inventory management.")
	}

	data := NewTemplateData(c, h.store)

	// Get donors, donation types, outbreaks, and treatment sites for dropdowns
	donors, err := h.getAllDonors()
	if err != nil {
		log.Printf("Error getting donors: %v", err)
	}
	data.Donors = donors

	donationTypes, err := h.getAllDonationTypes()
	if err != nil {
		log.Printf("Error getting donation types: %v", err)
	}
	data.DonationTypes = donationTypes

	// Note: Outbreaks and TreatmentSites are not available in TemplateData yet
	// We'll add them later if needed

	// Get items for donation items
	items, err := h.getAllInventoryItems()
	if err != nil {
		log.Printf("Error getting items: %v", err)
	}
	data.InventoryItems = items

	return GenerateHTML(c, h.db, data, "inventory_donation_form")
}

// HandlerDonationSave saves a new donation
func (h *InventoryHandler) HandlerDonationSave(c *fiber.Ctx) error {
	// Check if user has inventory access
	if !h.hasInventoryAccess(c) {
		return c.Status(403).SendString("Access denied. You don't have permission to access inventory management.")
	}

	// Parse form data
	donorID := parseInventoryInt(c.FormValue("donor_id"))
	donationTypeID := parseInventoryInt(c.FormValue("donation_type_id"))
	donationDate := parseInventoryDate(c.FormValue("donation_date"))
	receivedDate := parseInventoryDate(c.FormValue("received_date"))
	description := c.FormValue("description")
	monetaryValue := parseInventoryFloat(c.FormValue("monetary_value"))
	currency := c.FormValue("currency")
	if currency == "" {
		currency = "USD"
	}
	outbreakID := parseInventoryInt(c.FormValue("outbreak_id"))
	treatmentSiteID := parseInventoryInt(c.FormValue("treatment_site_id"))
	notes := c.FormValue("notes")

	// Get current user ID
	sess, err := h.store.Get(c)
	if err != nil {
		return c.Status(500).SendString("Session error")
	}
	userID := sess.Get("user_id").(int)

	// Insert donation
	query := `
		INSERT INTO inventory_donations (
			donor_id, donation_type_id, donation_date, received_date, 
			description, monetary_value, currency, outbreak_id, 
			treatment_site_id, received_by, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`

	var donationID int
	err = h.db.QueryRow(
		query, donorID, donationTypeID, donationDate, receivedDate,
		description, monetaryValue, currency, outbreakID, treatmentSiteID, userID, notes,
	).Scan(&donationID)

	if err != nil {
		log.Printf("Error saving donation: %v", err)
		return c.Status(500).SendString("Error saving donation")
	}

	// Handle donation items (for in-kind donations)
	itemIDs := c.FormValue("item_ids")
	quantities := c.FormValue("quantities")
	units := c.FormValue("units")
	estimatedValues := c.FormValue("estimated_values")
	conditionStatuses := c.FormValue("condition_statuses")
	expiryDates := c.FormValue("expiry_dates")
	batchNumbers := c.FormValue("batch_numbers")
	serialNumbers := c.FormValue("serial_numbers")
	itemNotes := c.FormValue("item_notes")

	if itemIDs != "" {
		itemIDList := strings.Split(itemIDs, ",")
		quantityList := strings.Split(quantities, ",")
		unitList := strings.Split(units, ",")
		valueList := strings.Split(estimatedValues, ",")
		conditionList := strings.Split(conditionStatuses, ",")
		expiryList := strings.Split(expiryDates, ",")
		batchList := strings.Split(batchNumbers, ",")
		serialList := strings.Split(serialNumbers, ",")
		noteList := strings.Split(itemNotes, ",")

		for i, itemIDStr := range itemIDList {
			if itemIDStr == "" {
				continue
			}

			itemID := parseInventoryInt(itemIDStr)
			quantity := parseInventoryFloat(quantityList[i])
			unit := unitList[i]
			estimatedValue := parseInventoryFloat(valueList[i])
			conditionStatus := conditionList[i]
			expiryDate := parseInventoryDate(expiryList[i])
			batchNumber := batchList[i]
			serialNumber := serialList[i]
			itemNote := noteList[i]

			// Insert donation item
			itemQuery := `
				INSERT INTO inventory_donation_items (
					donation_id, item_id, quantity, unit, estimated_value,
					condition_status, expiry_date, batch_number, serial_number, notes
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			`
			_, err = h.db.Exec(
				itemQuery, donationID, itemID, quantity, unit, estimatedValue,
				conditionStatus, expiryDate, batchNumber, serialNumber, itemNote,
			)
			if err != nil {
				log.Printf("Error saving donation item: %v", err)
			}
		}
	}

	return c.Redirect("/inventory/donations")
}

// HandlerDonationView displays a specific donation
func (h *InventoryHandler) HandlerDonationView(c *fiber.Ctx) error {
	// Check if user has inventory access
	if !h.hasInventoryAccess(c) {
		return c.Status(403).SendString("Access denied. You don't have permission to access inventory management.")
	}

	donationID := parseInventoryInt(c.Params("id"))
	if donationID == 0 {
		return c.Status(400).SendString("Invalid donation ID")
	}

	data := NewTemplateData(c, h.store)

	donation, err := h.getDonationByID(donationID)
	if err != nil {
		log.Printf("Error getting donation: %v", err)
		return c.Status(500).SendString("Error loading donation")
	}

	data.Donation = donation
	return GenerateHTML(c, h.db, data, "inventory_donation_view")
}

// HandlerDonorsList displays all donors
func (h *InventoryHandler) HandlerDonorsList(c *fiber.Ctx) error {
	// Check if user has inventory access
	if !h.hasInventoryAccess(c) {
		return c.Status(403).SendString("Access denied. You don't have permission to access inventory management.")
	}

	data := NewTemplateData(c, h.store)

	donors, err := h.getAllDonors()
	if err != nil {
		log.Printf("Error getting donors: %v", err)
		return c.Status(500).SendString("Error loading donors")
	}

	data.Donors = donors
	return GenerateHTML(c, h.db, data, "inventory_donors_list")
}

// HandlerDonorForm displays the donor form
func (h *InventoryHandler) HandlerDonorForm(c *fiber.Ctx) error {
	// Check if user has inventory access
	if !h.hasInventoryAccess(c) {
		return c.Status(403).SendString("Access denied. You don't have permission to access inventory management.")
	}

	data := NewTemplateData(c, h.store)
	return GenerateHTML(c, h.db, data, "inventory_donor_form")
}

// HandlerDonorSave saves a new donor
func (h *InventoryHandler) HandlerDonorSave(c *fiber.Ctx) error {
	// Check if user has inventory access
	if !h.hasInventoryAccess(c) {
		return c.Status(403).SendString("Access denied. You don't have permission to access inventory management.")
	}

	name := c.FormValue("name")
	organization := c.FormValue("organization")
	contactPerson := c.FormValue("contact_person")
	phone := c.FormValue("phone")
	email := c.FormValue("email")
	address := c.FormValue("address")
	donorType := c.FormValue("donor_type")
	country := c.FormValue("country")
	registrationNumber := c.FormValue("registration_number")
	taxExempt := c.FormValue("tax_exempt") == "on"
	notes := c.FormValue("notes")

	query := `
		INSERT INTO inventory_donors (
			name, organization, contact_person, phone, email, address,
			donor_type, country, registration_number, tax_exempt, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := h.db.Exec(
		query, name, organization, contactPerson, phone, email, address,
		donorType, country, registrationNumber, taxExempt, notes,
	)

	if err != nil {
		log.Printf("Error saving donor: %v", err)
		return c.Status(500).SendString("Error saving donor")
	}

	return c.Redirect("/inventory/donors")
}

// ==================== DONATION DATA ACCESS METHODS ====================

// getAllDonations retrieves all donations with related data
func (h *InventoryHandler) getAllDonations() ([]*DonationSummary, error) {
	query := `
		SELECT 
			d.id, d.donation_date, d.received_date, d.donation_status,
			dt.name as donation_type, dt.is_monetary,
			dr.name as donor_name, dr.organization as donor_organization, dr.donor_type,
			o.name as outbreak_name, ts.name as treatment_site_name,
			u.username as received_by_user,
			d.monetary_value, d.currency,
			COUNT(di.id) as item_count,
			SUM(di.quantity) as total_quantity,
			SUM(di.estimated_value) as total_estimated_value
		FROM inventory_donations d
		LEFT JOIN inventory_donation_types dt ON d.donation_type_id = dt.id
		LEFT JOIN inventory_donors dr ON d.donor_id = dr.id
		LEFT JOIN outbreaks o ON d.outbreak_id = o.id
		LEFT JOIN treatment_sites ts ON d.treatment_site_id = ts.id
		LEFT JOIN users u ON d.received_by = u.id
		LEFT JOIN inventory_donation_items di ON d.id = di.donation_id
		GROUP BY d.id, dt.name, dt.is_monetary, dr.name, dr.organization, dr.donor_type, 
				 o.name, ts.name, u.username, d.monetary_value, d.currency
		ORDER BY d.donation_date DESC
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var donations []*DonationSummary
	for rows.Next() {
		var d DonationSummary
		err := rows.Scan(
			&d.DonationID, &d.DonationDate, &d.ReceivedDate, &d.DonationStatus,
			&d.DonationType, &d.IsMonetary,
			&d.DonorName, &d.DonorOrganization, &d.DonorType,
			&d.OutbreakName, &d.TreatmentSiteName,
			&d.ReceivedByUser,
			&d.MonetaryValue, &d.Currency,
			&d.ItemCount, &d.TotalQuantity, &d.TotalEstimatedValue,
		)
		if err != nil {
			return nil, err
		}
		donations = append(donations, &d)
	}

	return donations, nil
}

// getDonationByID retrieves a specific donation with all related data
func (h *InventoryHandler) getDonationByID(donationID int) (*InventoryDonation, error) {
	// Get main donation data
	query := `
		SELECT 
			d.id, d.donor_id, d.donation_type_id, d.donation_date, d.received_date,
			d.description, d.monetary_value, d.currency, d.donation_status,
			d.outbreak_id, d.treatment_site_id, d.received_by, d.notes,
			d.created_at, d.updated_at
		FROM inventory_donations d
		WHERE d.id = $1
	`

	var donation InventoryDonation
	err := h.db.QueryRow(query, donationID).Scan(
		&donation.ID, &donation.DonorID, &donation.DonationTypeID, &donation.DonationDate, &donation.ReceivedDate,
		&donation.Description, &donation.MonetaryValue, &donation.Currency, &donation.DonationStatus,
		&donation.OutbreakID, &donation.TreatmentSiteID, &donation.ReceivedBy, &donation.Notes,
		&donation.CreatedAt, &donation.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Get donor information
	donor, err := h.getDonorByID(donation.DonorID)
	if err == nil {
		donation.Donor = donor
	}

	// Get donation type information
	donationType, err := h.getDonationTypeByID(donation.DonationTypeID)
	if err == nil {
		donation.DonationType = donationType
	}

	// Get donation items
	items, err := h.getDonationItemsByDonationID(donationID)
	if err == nil {
		donation.Items = items
	}

	return &donation, nil
}

// getAllDonors retrieves all donors
func (h *InventoryHandler) getAllDonors() ([]*InventoryDonor, error) {
	query := `
		SELECT id, name, organization, contact_person, phone, email, address,
			   donor_type, country, registration_number, tax_exempt, status, notes,
			   created_at, updated_at
		FROM inventory_donors
		ORDER BY name
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var donors []*InventoryDonor
	for rows.Next() {
		var d InventoryDonor
		err := rows.Scan(
			&d.ID, &d.Name, &d.Organization, &d.ContactPerson, &d.Phone, &d.Email, &d.Address,
			&d.DonorType, &d.Country, &d.RegistrationNumber, &d.TaxExempt, &d.Status, &d.Notes,
			&d.CreatedAt, &d.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		donors = append(donors, &d)
	}

	return donors, nil
}

// getDonorByID retrieves a specific donor
func (h *InventoryHandler) getDonorByID(donorID int) (*InventoryDonor, error) {
	query := `
		SELECT id, name, organization, contact_person, phone, email, address,
			   donor_type, country, registration_number, tax_exempt, status, notes,
			   created_at, updated_at
		FROM inventory_donors
		WHERE id = $1
	`

	var donor InventoryDonor
	err := h.db.QueryRow(query, donorID).Scan(
		&donor.ID, &donor.Name, &donor.Organization, &donor.ContactPerson, &donor.Phone, &donor.Email, &donor.Address,
		&donor.DonorType, &donor.Country, &donor.RegistrationNumber, &donor.TaxExempt, &donor.Status, &donor.Notes,
		&donor.CreatedAt, &donor.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &donor, nil
}

// getAllDonationTypes retrieves all donation types
func (h *InventoryHandler) getAllDonationTypes() ([]*InventoryDonationType, error) {
	query := `
		SELECT id, name, description, is_monetary, status, created_at
		FROM inventory_donation_types
		WHERE status = 'active'
		ORDER BY name
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []*InventoryDonationType
	for rows.Next() {
		var t InventoryDonationType
		err := rows.Scan(
			&t.ID, &t.Name, &t.Description, &t.IsMonetary, &t.Status, &t.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		types = append(types, &t)
	}

	return types, nil
}

// getDonationTypeByID retrieves a specific donation type
func (h *InventoryHandler) getDonationTypeByID(typeID int) (*InventoryDonationType, error) {
	query := `
		SELECT id, name, description, is_monetary, status, created_at
		FROM inventory_donation_types
		WHERE id = $1
	`

	var donationType InventoryDonationType
	err := h.db.QueryRow(query, typeID).Scan(
		&donationType.ID, &donationType.Name, &donationType.Description, &donationType.IsMonetary, &donationType.Status, &donationType.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &donationType, nil
}

// getDonationItemsByDonationID retrieves all items for a specific donation
func (h *InventoryHandler) getDonationItemsByDonationID(donationID int) ([]*InventoryDonationItem, error) {
	query := `
		SELECT di.id, di.donation_id, di.item_id, di.quantity, di.unit,
			   di.estimated_value, di.condition_status, di.expiry_date,
			   di.batch_number, di.serial_number, di.notes, di.created_at
		FROM inventory_donation_items di
		WHERE di.donation_id = $1
		ORDER BY di.id
	`

	rows, err := h.db.Query(query, donationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*InventoryDonationItem
	for rows.Next() {
		var item InventoryDonationItem
		err := rows.Scan(
			&item.ID, &item.DonationID, &item.ItemID, &item.Quantity, &item.Unit,
			&item.EstimatedValue, &item.ConditionStatus, &item.ExpiryDate,
			&item.BatchNumber, &item.SerialNumber, &item.Notes, &item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Get item details
		inventoryItem, err := h.getInventoryItemByID(strconv.Itoa(item.ItemID))
		if err == nil {
			item.Item = inventoryItem
		}

		items = append(items, &item)
	}

	return items, nil
}

// getAllOutbreaks retrieves all outbreaks for dropdown
func (h *InventoryHandler) getAllOutbreaks() ([]map[string]interface{}, error) {
	query := `
		SELECT id, name, outbreak_type, start_date, end_date, status
		FROM outbreaks
		ORDER BY name
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var outbreaks []map[string]interface{}
	for rows.Next() {
		var id int
		var name, outbreakType, status string
		var startDate, endDate sql.NullTime

		err := rows.Scan(&id, &name, &outbreakType, &startDate, &endDate, &status)
		if err != nil {
			return nil, err
		}

		outbreak := map[string]interface{}{
			"id":            id,
			"name":          name,
			"outbreak_type": outbreakType,
			"start_date":    startDate,
			"end_date":      endDate,
			"status":        status,
		}
		outbreaks = append(outbreaks, outbreak)
	}

	return outbreaks, nil
}

// getAllTreatmentSites retrieves all treatment sites for dropdown
func (h *InventoryHandler) getAllTreatmentSites() ([]map[string]interface{}, error) {
	query := `
		SELECT id, name, location, contact_person, phone, email, status
		FROM treatment_sites
		ORDER BY name
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sites []map[string]interface{}
	for rows.Next() {
		var id int
		var name, location, contactPerson, phone, email, status string

		err := rows.Scan(&id, &name, &location, &contactPerson, &phone, &email, &status)
		if err != nil {
			return nil, err
		}

		site := map[string]interface{}{
			"id":             id,
			"name":           name,
			"location":       location,
			"contact_person": contactPerson,
			"phone":          phone,
			"email":          email,
			"status":         status,
		}
		sites = append(sites, site)
	}

	return sites, nil
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
			SUM(CASE WHEN sl.current_quantity <= i.minimum_stock_level THEN 1 ELSE 0 END) as low_stock_count,
			SUM(sl.current_quantity * i.unit_cost) as total_value
		FROM inventory_items i
		LEFT JOIN inventory_stock_levels sl ON i.id = sl.item_id
		WHERE i.is_active = true
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
		JOIN treatment_sites s ON a.site_id = s.id
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
		JOIN treatment_sites s ON t.site_id = s.id
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
	// Donation-related fields
	IsDonated           bool            `json:"is_donated"`
	DonorID             sql.NullInt64   `json:"donor_id"`
	DonationDate        sql.NullTime    `json:"donation_date"`
	EstimatedDonorValue sql.NullFloat64 `json:"estimated_donor_value"`
	ConditionOnReceipt  string          `json:"condition_on_receipt"`
	DonorName           string          `json:"donor_name,omitempty"`
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

// Donation-related structures
type InventoryDonor struct {
	ID                 int       `json:"id"`
	Name               string    `json:"name"`
	Organization       string    `json:"organization"`
	ContactPerson      string    `json:"contact_person"`
	Phone              string    `json:"phone"`
	Email              string    `json:"email"`
	Address            string    `json:"address"`
	DonorType          string    `json:"donor_type"`
	Country            string    `json:"country"`
	RegistrationNumber string    `json:"registration_number"`
	TaxExempt          bool      `json:"tax_exempt"`
	Status             string    `json:"status"`
	Notes              string    `json:"notes"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type InventoryDonationType struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsMonetary  bool      `json:"is_monetary"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type InventoryDonation struct {
	ID              int             `json:"id"`
	DonorID         int             `json:"donor_id"`
	DonationTypeID  int             `json:"donation_type_id"`
	DonationDate    time.Time       `json:"donation_date"`
	ReceivedDate    sql.NullTime    `json:"received_date"`
	Description     string          `json:"description"`
	MonetaryValue   sql.NullFloat64 `json:"monetary_value"`
	Currency        string          `json:"currency"`
	DonationStatus  string          `json:"donation_status"`
	OutbreakID      sql.NullInt64   `json:"outbreak_id"`
	TreatmentSiteID sql.NullInt64   `json:"treatment_site_id"`
	ReceivedBy      sql.NullInt64   `json:"received_by"`
	Notes           string          `json:"notes"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	// Related data
	Donor        *InventoryDonor          `json:"donor,omitempty"`
	DonationType *InventoryDonationType   `json:"donation_type,omitempty"`
	Items        []*InventoryDonationItem `json:"items,omitempty"`
}

type InventoryDonationItem struct {
	ID              int             `json:"id"`
	DonationID      int             `json:"donation_id"`
	ItemID          int             `json:"item_id"`
	Quantity        float64         `json:"quantity"`
	Unit            string          `json:"unit"`
	EstimatedValue  sql.NullFloat64 `json:"estimated_value"`
	ConditionStatus string          `json:"condition_status"`
	ExpiryDate      sql.NullTime    `json:"expiry_date"`
	BatchNumber     string          `json:"batch_number"`
	SerialNumber    string          `json:"serial_number"`
	Notes           string          `json:"notes"`
	CreatedAt       time.Time       `json:"created_at"`
	// Related data
	Item *InventoryItem `json:"item,omitempty"`
}

type InventoryDonationAcknowledgment struct {
	ID                 int           `json:"id"`
	DonationID         int           `json:"donation_id"`
	AcknowledgmentType string        `json:"acknowledgment_type"`
	SentDate           sql.NullTime  `json:"sent_date"`
	SentBy             sql.NullInt64 `json:"sent_by"`
	RecipientName      string        `json:"recipient_name"`
	RecipientEmail     string        `json:"recipient_email"`
	AcknowledgmentText string        `json:"acknowledgment_text"`
	Status             string        `json:"status"`
	CreatedAt          time.Time     `json:"created_at"`
}

type DonationSummary struct {
	DonationID          int             `json:"donation_id"`
	DonationDate        time.Time       `json:"donation_date"`
	ReceivedDate        sql.NullTime    `json:"received_date"`
	DonationStatus      string          `json:"donation_status"`
	DonationType        string          `json:"donation_type"`
	IsMonetary          bool            `json:"is_monetary"`
	DonorName           string          `json:"donor_name"`
	DonorOrganization   string          `json:"donor_organization"`
	DonorType           string          `json:"donor_type"`
	OutbreakName        string          `json:"outbreak_name"`
	TreatmentSiteName   string          `json:"treatment_site_name"`
	ReceivedByUser      string          `json:"received_by_user"`
	MonetaryValue       sql.NullFloat64 `json:"monetary_value"`
	Currency            string          `json:"currency"`
	ItemCount           int             `json:"item_count"`
	TotalQuantity       sql.NullFloat64 `json:"total_quantity"`
	TotalEstimatedValue sql.NullFloat64 `json:"total_estimated_value"`
}

type DonationStatistics struct {
	TotalDonations     int     `json:"total_donations"`
	TotalMonetaryValue float64 `json:"total_monetary_value"`
	TotalInKindValue   float64 `json:"total_in_kind_value"`
	DonorCount         int     `json:"donor_count"`
	TopDonorName       string  `json:"top_donor_name"`
	TopDonorValue      float64 `json:"top_donor_value"`
}
