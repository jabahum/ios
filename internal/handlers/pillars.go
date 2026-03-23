package handlers

import (
	"case/internal/models"
	"database/sql"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// PillarsHandler handles all pillar operations
type PillarsHandler struct {
	db    *sql.DB
	store *session.Store
}

// NewPillarsHandler creates a new pillars handler
func NewPillarsHandler(db *sql.DB, store *session.Store) *PillarsHandler {
	return &PillarsHandler{db: db, store: store}
}

// HandlerPillarsList displays all pillars
func (h *PillarsHandler) HandlerPillarsList(c *fiber.Ctx) error {
	data := NewTemplateDataWithDB(c, h.store, h.db)

	pillars, err := h.getAllPillars()
	if err != nil {
		log.Printf("Error getting pillars: %v", err)
		pillars = []*models.Pillar{}
	}

	data.Pillars = pillars
	return GenerateHTML(c, h.db, data, "pillars_list")
}

// HandlerPillarForm displays the form for adding/editing pillars
func (h *PillarsHandler) HandlerPillarForm(c *fiber.Ctx) error {
	data := NewTemplateDataWithDB(c, h.store, h.db)

	// If editing, get the pillar
	pillarID := c.Params("id")
	if pillarID != "" {
		pillar, err := h.getPillarByID(pillarID)
		if err != nil {
			log.Printf("Error getting pillar: %v", err)
		} else {
			data.Pillar = pillar
		}
	}

	// Get users for pillar head dropdown
	users, err := h.getAllEnhancedUsers()
	if err != nil {
		log.Printf("Error getting users: %v", err)
		users = []*models.EnhancedUser{}
	}
	data.Users = users

	return GenerateHTML(c, h.db, data, "pillar_form")
}

// HandlerPillarSave handles saving pillar data
func (h *PillarsHandler) HandlerPillarSave(c *fiber.Ctx) error {
	pillarID := c.FormValue("id")

	// Get current user ID from session
	userID := GetCurrentUser(c, h.store)
	if userID == 0 {
		return c.Redirect("/login")
	}

	// Parse pillar head ID
	pillarHeadID := parseRRTNullInt64(c.FormValue("pillar_head_id"))

	pillar := &models.Pillar{
		Name:            c.FormValue("name"),
		Description:     parseRRTNullString(c.FormValue("description")),
		PillarHeadID:    pillarHeadID,
		PillarHeadName:  parseRRTNullString(c.FormValue("pillar_head_name")),
		PillarHeadEmail: parseRRTNullString(c.FormValue("pillar_head_email")),
		PillarHeadPhone: parseRRTNullString(c.FormValue("pillar_head_phone")),
		IsActive:        c.FormValue("is_active") == "true",
		UpdatedBy:       sql.NullInt64{Int64: int64(userID), Valid: true},
	}

	var err error
	if pillarID != "" {
		pillar.ID, _ = strconv.ParseInt(pillarID, 10, 64)
		err = h.updatePillar(pillar)
	} else {
		err = h.createPillar(pillar)
	}

	if err != nil {
		log.Printf("Error saving pillar: %v", err)
		return c.Status(500).SendString("Error saving pillar")
	}

	return c.Redirect("/resource-management/pillars")
}

// HandlerPillarChanges displays pillar change history
func (h *PillarsHandler) HandlerPillarChanges(c *fiber.Ctx) error {
	data := NewTemplateDataWithDB(c, h.store, h.db)

	pillarID := c.Params("id")
	changes, err := h.getPillarChanges(pillarID)
	if err != nil {
		log.Printf("Error getting pillar changes: %v", err)
		changes = []*models.PillarChange{}
	}

	data.PillarChanges = changes
	return GenerateHTML(c, h.db, data, "pillar_changes")
}

// API endpoint to get pillars for dropdown
func (h *PillarsHandler) HandlerPillarsAPI(c *fiber.Ctx) error {
	pillars, err := h.getAllPillars()
	if err != nil {
		log.Printf("Error getting pillars: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get pillars"})
	}

	return c.JSON(pillars)
}

// Helper methods for database operations
func (h *PillarsHandler) getAllPillars() ([]*models.Pillar, error) {
	// Join users with user_name + email only (first_name/last_name are not present on all deployed DBs).
	query := `
		SELECT p.id, p.name, p.description, p.pillar_head_id, p.pillar_head_name, 
		       p.pillar_head_email, p.pillar_head_phone, p.is_active, p.created_at, 
		       p.updated_at, p.created_by, p.updated_by,
		       u.user_name, u.email
		FROM pillars p
		LEFT JOIN users u ON p.pillar_head_id = u.user_id
		ORDER BY p.name
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pillars []*models.Pillar
	for rows.Next() {
		pillar := &models.Pillar{}
		var headUserName, headEmail sql.NullString

		err := rows.Scan(
			&pillar.ID, &pillar.Name, &pillar.Description, &pillar.PillarHeadID,
			&pillar.PillarHeadName, &pillar.PillarHeadEmail, &pillar.PillarHeadPhone,
			&pillar.IsActive, &pillar.CreatedAt, &pillar.UpdatedAt,
			&pillar.CreatedBy, &pillar.UpdatedBy,
			&headUserName, &headEmail,
		)
		if err != nil {
			return nil, err
		}

		if headUserName.Valid || headEmail.Valid {
			pillar.PillarHead = &models.EnhancedUser{
				UserName: headUserName,
				Email:    headEmail,
			}
			if headUserName.Valid {
				pillar.PillarHead.FirstName = sql.NullString{String: headUserName.String, Valid: true}
			}
		}

		pillars = append(pillars, pillar)
	}

	return pillars, nil
}

func (h *PillarsHandler) getPillarByID(id string) (*models.Pillar, error) {
	query := `
		SELECT p.id, p.name, p.description, p.pillar_head_id, p.pillar_head_name, 
		       p.pillar_head_email, p.pillar_head_phone, p.is_active, p.created_at, 
		       p.updated_at, p.created_by, p.updated_by,
		       u.user_name, u.email
		FROM pillars p
		LEFT JOIN users u ON p.pillar_head_id = u.user_id
		WHERE p.id = $1
	`

	pillar := &models.Pillar{}
	var headUserName, headEmail sql.NullString

	err := h.db.QueryRow(query, id).Scan(
		&pillar.ID, &pillar.Name, &pillar.Description, &pillar.PillarHeadID,
		&pillar.PillarHeadName, &pillar.PillarHeadEmail, &pillar.PillarHeadPhone,
		&pillar.IsActive, &pillar.CreatedAt, &pillar.UpdatedAt,
		&pillar.CreatedBy, &pillar.UpdatedBy,
		&headUserName, &headEmail,
	)
	if err != nil {
		return nil, err
	}

	if headUserName.Valid || headEmail.Valid {
		pillar.PillarHead = &models.EnhancedUser{
			UserName: headUserName,
			Email:    headEmail,
		}
		if headUserName.Valid {
			pillar.PillarHead.FirstName = sql.NullString{String: headUserName.String, Valid: true}
		}
	}

	return pillar, nil
}

func (h *PillarsHandler) createPillar(pillar *models.Pillar) error {
	query := `
		INSERT INTO pillars (name, description, pillar_head_id, pillar_head_name, 
		                     pillar_head_email, pillar_head_phone, is_active, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var id int64
	err := h.db.QueryRow(query,
		pillar.Name, pillar.Description, pillar.PillarHeadID,
		pillar.PillarHeadName, pillar.PillarHeadEmail, pillar.PillarHeadPhone,
		pillar.IsActive, pillar.CreatedBy,
	).Scan(&id)

	if err != nil {
		return err
	}

	pillar.ID = id
	return nil
}

func (h *PillarsHandler) updatePillar(pillar *models.Pillar) error {
	// First, get the old values for change tracking
	oldPillar, err := h.getPillarByID(strconv.FormatInt(pillar.ID, 10))
	if err != nil {
		return err
	}

	// Update the pillar
	query := `
		UPDATE pillars 
		SET name = $1, description = $2, pillar_head_id = $3, pillar_head_name = $4,
		    pillar_head_email = $5, pillar_head_phone = $6, is_active = $7, 
		    updated_at = CURRENT_TIMESTAMP, updated_by = $8
		WHERE id = $9
	`

	_, err = h.db.Exec(query,
		pillar.Name, pillar.Description, pillar.PillarHeadID,
		pillar.PillarHeadName, pillar.PillarHeadEmail, pillar.PillarHeadPhone,
		pillar.IsActive, pillar.UpdatedBy, pillar.ID,
	)
	if err != nil {
		return err
	}

	// Create change records for significant changes
	err = h.createPillarChangeRecord(pillar, oldPillar)
	return err
}

// createPillarChangeRecord creates change records for pillar updates
func (h *PillarsHandler) createPillarChangeRecord(newPillar, oldPillar *models.Pillar) error {
	// Check for name changes
	if oldPillar.Name != newPillar.Name {
		err := h.insertPillarChange(newPillar.ID, "name_change", oldPillar.Name, newPillar.Name, "Pillar name updated", newPillar.UpdatedBy, "")
		if err != nil {
			return err
		}
	}

	// Check for description changes
	oldDesc := ""
	newDesc := ""
	if oldPillar.Description.Valid {
		oldDesc = oldPillar.Description.String
	}
	if newPillar.Description.Valid {
		newDesc = newPillar.Description.String
	}
	if oldDesc != newDesc {
		err := h.insertPillarChange(newPillar.ID, "description_change", oldDesc, newDesc, "Pillar description updated", newPillar.UpdatedBy, "")
		if err != nil {
			return err
		}
	}

	// Check for pillar head changes
	oldHeadID := int64(0)
	newHeadID := int64(0)
	if oldPillar.PillarHeadID.Valid {
		oldHeadID = oldPillar.PillarHeadID.Int64
	}
	if newPillar.PillarHeadID.Valid {
		newHeadID = newPillar.PillarHeadID.Int64
	}
	if oldHeadID != newHeadID {
		oldHeadName := ""
		newHeadName := ""
		if oldPillar.PillarHeadName.Valid {
			oldHeadName = oldPillar.PillarHeadName.String
		}
		if newPillar.PillarHeadName.Valid {
			newHeadName = newPillar.PillarHeadName.String
		}
		err := h.insertPillarChange(newPillar.ID, "head_change", oldHeadName, newHeadName, "Pillar head changed", newPillar.UpdatedBy, "")
		if err != nil {
			return err
		}
	}

	// Check for status changes
	if oldPillar.IsActive != newPillar.IsActive {
		oldStatus := "Inactive"
		newStatus := "Inactive"
		if oldPillar.IsActive {
			oldStatus = "Active"
		}
		if newPillar.IsActive {
			newStatus = "Active"
		}
		err := h.insertPillarChange(newPillar.ID, "status_change", oldStatus, newStatus, "Pillar status changed", newPillar.UpdatedBy, "")
		if err != nil {
			return err
		}
	}

	return nil
}

// insertPillarChange inserts a single change record
func (h *PillarsHandler) insertPillarChange(pillarID int64, changeType, oldValue, newValue, reason string, changedBy sql.NullInt64, notes string) error {
	query := `
		INSERT INTO pillar_changes (pillar_id, change_type, old_value, new_value, change_reason, changed_by, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := h.db.Exec(query, pillarID, changeType, oldValue, newValue, reason, changedBy, notes)
	return err
}

func (h *PillarsHandler) getPillarChanges(pillarID string) ([]*models.PillarChange, error) {
	query := `
		SELECT pc.id, pc.pillar_id, pc.change_type, pc.old_value, pc.new_value,
		       pc.change_reason, pc.changed_by, pc.changed_at, pc.notes,
		       u.user_name
		FROM pillar_changes pc
		LEFT JOIN users u ON pc.changed_by = u.user_id
		WHERE pc.pillar_id = $1
		ORDER BY pc.changed_at DESC
	`

	rows, err := h.db.Query(query, pillarID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var changes []*models.PillarChange
	for rows.Next() {
		change := &models.PillarChange{}
		var changedByUserName sql.NullString

		err := rows.Scan(
			&change.ID, &change.PillarID, &change.ChangeType, &change.OldValue,
			&change.NewValue, &change.ChangeReason, &change.ChangedBy,
			&change.ChangedAt, &change.Notes, &changedByUserName,
		)
		if err != nil {
			return nil, err
		}

		if changedByUserName.Valid {
			change.ChangedByUser = &models.EnhancedUser{
				UserName:  changedByUserName,
				FirstName: sql.NullString{String: changedByUserName.String, Valid: true},
			}
		}

		changes = append(changes, change)
	}

	return changes, nil
}

func (h *PillarsHandler) deletePillarByID(id int64) error {
	_, err := h.db.Exec(`DELETE FROM pillars WHERE id = $1`, id)
	return err
}

func (h *PillarsHandler) getAllEnhancedUsers() ([]*models.EnhancedUser, error) {
	query := `
		SELECT user_id, user_name, email, is_active
		FROM users
		WHERE is_active = true
		ORDER BY user_name
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.EnhancedUser
	for rows.Next() {
		user := &models.EnhancedUser{}
		var userName sql.NullString
		err := rows.Scan(
			&user.UserID, &userName, &user.Email, &user.IsActive,
		)
		if err != nil {
			return nil, err
		}

		// Set the user name as the display name
		if userName.Valid {
			user.FirstName = sql.NullString{String: userName.String, Valid: true}
		}

		users = append(users, user)
	}

	return users, nil
}
