package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// Inventory API handlers - these extend the existing inventory handler with API endpoints

// HandlerInventoryDashboardAPI returns inventory dashboard data as JSON
func (h *InventoryHandler) HandlerInventoryDashboardAPI(c *fiber.Ctx) error {
	stats, err := h.getInventoryStats()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error loading inventory stats"})
	}
	
	return c.JSON(fiber.Map{
		"message": "Inventory dashboard API",
		"stats": stats,
	})
}

// HandlerGetInventoryItemAPI returns a single inventory item as JSON
func (h *InventoryHandler) HandlerGetInventoryItemAPI(c *fiber.Ctx) error {
	itemID := c.Params("id")
	item, err := h.getInventoryItemByID(itemID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Inventory item not found"})
	}
	
    return c.JSON(fiber.Map{
		"inventory_item": item,
	})
}

// HandlerInventoryItemSaveAPI creates a new inventory item via API
func (h *InventoryHandler) HandlerInventoryItemSaveAPI(c *fiber.Ctx) error {
	var item InventoryItem
	if err := c.BodyParser(&item); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	
	if err := h.createInventoryItem(&item); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error saving inventory item"})
	}
	
	return c.Status(201).JSON(fiber.Map{"message": "Inventory item saved successfully", "item": item})
}

// HandlerInventoryItemUpdateAPI updates an existing inventory item via API
func (h *InventoryHandler) HandlerInventoryItemUpdateAPI(c *fiber.Ctx) error {
	itemID := c.Params("id")
	
	var item InventoryItem
	if err := c.BodyParser(&item); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	
	item.ID = parseInventoryInt(itemID)
	if err := h.updateInventoryItem(&item); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error updating inventory item"})
	}
	
	return c.JSON(fiber.Map{"message": "Inventory item updated successfully", "item": item})
}

// HandlerInventoryItemDeleteAPI deletes an inventory item via API
func (h *InventoryHandler) HandlerInventoryItemDeleteAPI(c *fiber.Ctx) error {
	itemID := c.Params("id")
	
	// For now, we'll just mark it as inactive instead of deleting
	_, err := h.db.Exec("UPDATE inventory_items SET is_active = false WHERE id = $1", itemID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error deleting inventory item"})
	}
	
    return c.JSON(fiber.Map{"message": "Inventory item deleted successfully"})
}

// HandlerGetInventoryStockAPI returns stock level for a specific item
func (h *InventoryHandler) HandlerGetInventoryStockAPI(c *fiber.Ctx) error {
	stockID := c.Params("id")
	
	var stock InventoryStockLevel
	query := `
		SELECT id, item_id, site_id, quantity, last_updated
		FROM inventory_stock_levels
		WHERE id = $1
	`
	err := h.db.QueryRow(query, stockID).Scan(&stock.ID, &stock.ItemID, &stock.SiteID, &stock.Quantity, &stock.LastUpdated)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Stock level not found"})
	}
	
    return c.JSON(fiber.Map{
		"stock": stock,
	})
}

// HandlerInventoryStockSaveAPI creates a new stock transaction
func (h *InventoryHandler) HandlerInventoryStockSaveAPI(c *fiber.Ctx) error {
	var transaction InventoryTransaction
	if err := c.BodyParser(&transaction); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	
	if err := h.createInventoryTransaction(&transaction); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error saving stock transaction"})
	}
	
	return c.Status(201).JSON(fiber.Map{"message": "Stock level saved successfully", "transaction": transaction})
}

// HandlerInventoryStockUpdateAPI updates stock level
func (h *InventoryHandler) HandlerInventoryStockUpdateAPI(c *fiber.Ctx) error {
	stockID := c.Params("id")
	
	var data struct {
		Quantity float64 `json:"quantity"`
	}
	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	
	_, err := h.db.Exec("UPDATE inventory_stock_levels SET quantity = $1, last_updated = NOW() WHERE id = $2", data.Quantity, stockID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error updating stock level"})
	}
	
	return c.JSON(fiber.Map{"message": "Stock level updated successfully"})
}

// HandlerInventoryPurchaseOrdersAPI returns all purchase orders
func (h *InventoryHandler) HandlerInventoryPurchaseOrdersAPI(c *fiber.Ctx) error {
	query := `
		SELECT po.id, po.supplier_id, po.order_date, po.expected_delivery, po.status, po.notes,
		       s.name as supplier_name
		FROM inventory_purchase_orders po
		LEFT JOIN inventory_suppliers s ON po.supplier_id = s.id
		ORDER BY po.order_date DESC
	`
	
	rows, err := h.db.Query(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error loading purchase orders"})
	}
	defer rows.Close()
	
	var orders []fiber.Map
	for rows.Next() {
		var po InventoryPurchaseOrder
		var supplierName string
		err := rows.Scan(&po.ID, &po.SupplierID, &po.OrderDate, &po.ExpectedDelivery, &po.Status, &po.Notes, &supplierName)
		if err != nil {
			continue
		}
		orders = append(orders, fiber.Map{
			"id": po.ID,
			"supplier_id": po.SupplierID,
			"supplier_name": supplierName,
			"order_date": po.OrderDate,
			"expected_delivery": po.ExpectedDelivery,
			"status": po.Status,
			"notes": po.Notes,
		})
	}
	
	return c.JSON(fiber.Map{"purchase_orders": orders})
}

// HandlerGetPurchaseOrderAPI returns a single purchase order
func (h *InventoryHandler) HandlerGetPurchaseOrderAPI(c *fiber.Ctx) error {
	poID := c.Params("id")
	
	var po InventoryPurchaseOrder
	query := `SELECT id, supplier_id, order_date, expected_delivery, status, notes FROM inventory_purchase_orders WHERE id = $1`
	err := h.db.QueryRow(query, poID).Scan(&po.ID, &po.SupplierID, &po.OrderDate, &po.ExpectedDelivery, &po.Status, &po.Notes)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Purchase order not found"})
	}
	
    return c.JSON(fiber.Map{
		"purchase_order": po,
	})
}

// HandlerInventoryPurchaseOrderSaveAPI creates a new purchase order
func (h *InventoryHandler) HandlerInventoryPurchaseOrderSaveAPI(c *fiber.Ctx) error {
	var po InventoryPurchaseOrder
	if err := c.BodyParser(&po); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	
	if err := h.createPurchaseOrder(&po); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error saving purchase order"})
	}
	
	return c.Status(201).JSON(fiber.Map{"message": "Purchase order saved successfully", "purchase_order": po})
}

// HandlerPurchaseOrderUpdateAPI updates a purchase order
func (h *InventoryHandler) HandlerPurchaseOrderUpdateAPI(c *fiber.Ctx) error {
	poID := c.Params("id")
	
	var po InventoryPurchaseOrder
	if err := c.BodyParser(&po); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	
	query := `UPDATE inventory_purchase_orders SET supplier_id = $1, expected_delivery = $2, status = $3, notes = $4 WHERE id = $5`
	_, err := h.db.Exec(query, po.SupplierID, po.ExpectedDelivery, po.Status, po.Notes, poID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error updating purchase order"})
	}
	
	return c.JSON(fiber.Map{"message": "Purchase order updated successfully"})
}

// HandlerInventoryRequisitionsAPI returns all requisitions
func (h *InventoryHandler) HandlerInventoryRequisitionsAPI(c *fiber.Ctx) error {
	query := `
		SELECT r.id, r.site_id, r.request_date, r.priority, r.status, r.notes,
		       ts.name as site_name
		FROM inventory_requisitions r
		LEFT JOIN treatment_sites ts ON r.site_id = ts.id
		ORDER BY r.request_date DESC
	`
	
	rows, err := h.db.Query(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error loading requisitions"})
	}
	defer rows.Close()
	
	var requisitions []fiber.Map
	for rows.Next() {
		var req InventoryRequisition
		var siteName string
		err := rows.Scan(&req.ID, &req.SiteID, &req.RequestDate, &req.Priority, &req.Status, &req.Notes, &siteName)
		if err != nil {
			continue
		}
		requisitions = append(requisitions, fiber.Map{
			"id": req.ID,
			"site_id": req.SiteID,
			"site_name": siteName,
			"request_date": req.RequestDate,
			"priority": req.Priority,
			"status": req.Status,
			"notes": req.Notes,
		})
	}
	
	return c.JSON(fiber.Map{"requisitions": requisitions})
}

// HandlerGetRequisitionAPI returns a single requisition
func (h *InventoryHandler) HandlerGetRequisitionAPI(c *fiber.Ctx) error {
	reqID := c.Params("id")
	
	var req InventoryRequisition
	query := `SELECT id, site_id, request_date, priority, status, notes FROM inventory_requisitions WHERE id = $1`
	err := h.db.QueryRow(query, reqID).Scan(&req.ID, &req.SiteID, &req.RequestDate, &req.Priority, &req.Status, &req.Notes)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Requisition not found"})
	}
	
    return c.JSON(fiber.Map{
		"requisition": req,
	})
}

// HandlerInventoryRequisitionSaveAPI creates a new requisition
func (h *InventoryHandler) HandlerInventoryRequisitionSaveAPI(c *fiber.Ctx) error {
	var req InventoryRequisition
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	
	if err := h.createRequisition(&req); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error saving requisition"})
	}
	
	return c.Status(201).JSON(fiber.Map{"message": "Requisition saved successfully", "requisition": req})
}

// HandlerRequisitionUpdateAPI updates a requisition
func (h *InventoryHandler) HandlerRequisitionUpdateAPI(c *fiber.Ctx) error {
	reqID := c.Params("id")
	
	var req InventoryRequisition
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	
	query := `UPDATE inventory_requisitions SET site_id = $1, priority = $2, status = $3, notes = $4 WHERE id = $5`
	_, err := h.db.Exec(query, req.SiteID, req.Priority, req.Status, req.Notes, reqID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error updating requisition"})
	}
	
	return c.JSON(fiber.Map{"message": "Requisition updated successfully"})
}

// HandlerDonationsListAPI returns all donations as JSON
func (h *InventoryHandler) HandlerDonationsListAPI(c *fiber.Ctx) error {
	donations, err := h.getAllDonations()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error loading donations"})
	}
	
	return c.JSON(fiber.Map{"donations": donations})
}

// HandlerDonationViewAPI returns a single donation as JSON
func (h *InventoryHandler) HandlerDonationViewAPI(c *fiber.Ctx) error {
	donationID := parseInventoryInt(c.Params("id"))
	donation, err := h.getDonationByID(donationID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Donation not found"})
	}
	
    return c.JSON(fiber.Map{
		"donation": donation,
	})
}

// HandlerDonationSaveAPI creates a new donation via API
func (h *InventoryHandler) HandlerDonationSaveAPI(c *fiber.Ctx) error {
	var donation InventoryDonation
	if err := c.BodyParser(&donation); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	
	// Insert the donation (simplified version)
	query := `
		INSERT INTO inventory_donations (
			donor_id, donation_type_id, donation_date, received_date, 
			description, monetary_value, currency, outbreak_id, 
			treatment_site_id, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`
	
	err := h.db.QueryRow(
		query, donation.DonorID, donation.DonationTypeID, donation.DonationDate, donation.ReceivedDate,
		donation.Description, donation.MonetaryValue, donation.Currency, donation.OutbreakID, 
		donation.TreatmentSiteID, donation.Notes,
	).Scan(&donation.ID)
	
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error saving donation"})
	}
	
	return c.Status(201).JSON(fiber.Map{"message": "Donation saved successfully", "donation": donation})
}

// HandlerDonationUpdateAPI updates a donation via API
func (h *InventoryHandler) HandlerDonationUpdateAPI(c *fiber.Ctx) error {
	donationID := c.Params("id")
	
	var donation InventoryDonation
	if err := c.BodyParser(&donation); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	
	query := `
		UPDATE inventory_donations 
		SET donor_id = $1, donation_type_id = $2, donation_date = $3, received_date = $4,
		    description = $5, monetary_value = $6, currency = $7, outbreak_id = $8,
		    treatment_site_id = $9, notes = $10, donation_status = $11
		WHERE id = $12
	`
	
	_, err := h.db.Exec(
		query, donation.DonorID, donation.DonationTypeID, donation.DonationDate, donation.ReceivedDate,
		donation.Description, donation.MonetaryValue, donation.Currency, donation.OutbreakID,
		donation.TreatmentSiteID, donation.Notes, donation.DonationStatus, donationID,
	)
	
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error updating donation"})
	}
	
	return c.JSON(fiber.Map{"message": "Donation updated successfully"})
}

// HandlerDonationDeleteAPI deletes a donation via API
func (h *InventoryHandler) HandlerDonationDeleteAPI(c *fiber.Ctx) error {
	donationID := c.Params("id")
	
	// Soft delete by updating status
	_, err := h.db.Exec("UPDATE inventory_donations SET donation_status = 'cancelled' WHERE id = $1", donationID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error deleting donation"})
	}
	
    return c.JSON(fiber.Map{"message": "Donation deleted successfully"})
}

// HandlerDonorsListAPI returns all donors as JSON
func (h *InventoryHandler) HandlerDonorsListAPI(c *fiber.Ctx) error {
	donors, err := h.getAllDonors()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error loading donors"})
	}
	
	return c.JSON(fiber.Map{"donors": donors})
}

// HandlerGetDonorAPI returns a single donor as JSON
func (h *InventoryHandler) HandlerGetDonorAPI(c *fiber.Ctx) error {
	donorID := parseInventoryInt(c.Params("id"))
	donor, err := h.getDonorByID(donorID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Donor not found"})
	}
	
    return c.JSON(fiber.Map{
		"donor": donor,
	})
}

// HandlerDonorSaveAPI creates a new donor via API
func (h *InventoryHandler) HandlerDonorSaveAPI(c *fiber.Ctx) error {
	var donor InventoryDonor
	if err := c.BodyParser(&donor); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	
	query := `
		INSERT INTO inventory_donors (
			name, organization, contact_person, phone, email, address,
			donor_type, country, registration_number, tax_exempt, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`
	
	err := h.db.QueryRow(
		query, donor.Name, donor.Organization, donor.ContactPerson, donor.Phone, donor.Email, donor.Address,
		donor.DonorType, donor.Country, donor.RegistrationNumber, donor.TaxExempt, donor.Notes,
	).Scan(&donor.ID)
	
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error saving donor"})
	}
	
	return c.Status(201).JSON(fiber.Map{"message": "Donor saved successfully", "donor": donor})
}

// HandlerDonorUpdateAPI updates a donor via API
func (h *InventoryHandler) HandlerDonorUpdateAPI(c *fiber.Ctx) error {
	donorID := c.Params("id")
	
	var donor InventoryDonor
	if err := c.BodyParser(&donor); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	
	query := `
		UPDATE inventory_donors 
		SET name = $1, organization = $2, contact_person = $3, phone = $4, email = $5, 
		    address = $6, donor_type = $7, country = $8, registration_number = $9, 
		    tax_exempt = $10, notes = $11, status = $12
		WHERE id = $13
	`
	
	_, err := h.db.Exec(
		query, donor.Name, donor.Organization, donor.ContactPerson, donor.Phone, donor.Email,
		donor.Address, donor.DonorType, donor.Country, donor.RegistrationNumber,
		donor.TaxExempt, donor.Notes, donor.Status, donorID,
	)
	
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error updating donor"})
	}
	
	return c.JSON(fiber.Map{"message": "Donor updated successfully"})
}

// HandlerDonorDeleteAPI deletes a donor via API
func (h *InventoryHandler) HandlerDonorDeleteAPI(c *fiber.Ctx) error {
	donorID := c.Params("id")
	
	// Soft delete by updating status
	_, err := h.db.Exec("UPDATE inventory_donors SET status = 'inactive' WHERE id = $1", donorID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error deleting donor"})
	}
	
    return c.JSON(fiber.Map{"message": "Donor deleted successfully"})
}
