package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection
	db, err := sql.Open("postgres", "postgres://postgres:pwaiswa@localhost:5432/ios?sslmode=disable")
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Error pinging database:", err)
	}

	fmt.Println("Connected to database successfully")

	// Initialize RBAC system
	if err := initializeRBAC(db); err != nil {
		log.Fatal("Error initializing RBAC:", err)
	}

	fmt.Println("RBAC system initialized successfully")
}

func initializeRBAC(db *sql.DB) error {
	ctx := db.QueryRowContext

	// 1. Insert default departments
	fmt.Println("Creating departments...")
	departments := []struct {
		name        string
		description string
	}{
		{"Administration", "System administration and management"},
		{"Surveillance", "Disease surveillance and monitoring"},
		{"Laboratory", "Laboratory services and testing"},
		{"Clinical", "Clinical services and patient care"},
		{"Data Management", "Data entry and management"},
		{"Reports", "Reporting and analytics"},
	}

	for _, dept := range departments {
		_, err := db.ExecContext(ctx, `
			INSERT INTO departments (name, description) 
			VALUES ($1, $2) 
			ON CONFLICT (name) DO NOTHING
		`, dept.name, dept.description)
		if err != nil {
			return fmt.Errorf("error creating department %s: %v", dept.name, err)
		}
	}

	// 2. Insert default roles
	fmt.Println("Creating roles...")
	roles := []struct {
		name        string
		description string
	}{
		{"super_admin", "Super Administrator with full system access"},
		{"admin", "Administrator with management access"},
		{"manager", "Department manager with oversight access"},
		{"data_entry", "Data entry personnel"},
		{"viewer", "Read-only access to reports and data"},
		{"lab_technician", "Laboratory technician with lab-specific access"},
		{"surveillance_officer", "Surveillance officer with monitoring access"},
	}

	for _, role := range roles {
		_, err := db.ExecContext(ctx, `
			INSERT INTO roles (name, description) 
			VALUES ($1, $2) 
			ON CONFLICT (name) DO NOTHING
		`, role.name, role.description)
		if err != nil {
			return fmt.Errorf("error creating role %s: %v", role.name, err)
		}
	}

	// 3. Insert permissions
	fmt.Println("Creating permissions...")
	permissions := []struct {
		name        string
		description string
		resource    string
		action      string
	}{
		// VHF Patients permissions
		{"Create VHF Patients", "Create new VHF patient records", "vhf_patients", "create"},
		{"View VHF Patients", "View VHF patient records", "vhf_patients", "read"},
		{"Update VHF Patients", "Update VHF patient records", "vhf_patients", "update"},
		{"Delete VHF Patients", "Delete VHF patient records", "vhf_patients", "delete"},
		{"Export VHF Patients", "Export VHF patient data", "vhf_patients", "export"},

		// Users permissions
		{"Create Users", "Create new user accounts", "users", "create"},
		{"View Users", "View user accounts", "users", "read"},
		{"Update Users", "Update user accounts", "users", "update"},
		{"Delete Users", "Delete user accounts", "users", "delete"},

		// Reports permissions
		{"View Reports", "View system reports", "reports", "read"},
		{"Export Reports", "Export report data", "reports", "export"},

		// Outbreaks permissions
		{"Create Outbreaks", "Create new outbreak records", "outbreaks", "create"},
		{"View Outbreaks", "View outbreak records", "outbreaks", "read"},
		{"Update Outbreaks", "Update outbreak records", "outbreaks", "update"},
		{"Delete Outbreaks", "Delete outbreak records", "outbreaks", "delete"},

		// Employees permissions
		{"Create Employees", "Create new employee records", "employees", "create"},
		{"View Employees", "View employee records", "employees", "read"},
		{"Update Employees", "Update employee records", "employees", "update"},
		{"Delete Employees", "Delete employee records", "employees", "delete"},

		// Facilities permissions
		{"Create Facilities", "Create new facility records", "facilities", "create"},
		{"View Facilities", "View facility records", "facilities", "read"},
		{"Update Facilities", "Update facility records", "facilities", "update"},
		{"Delete Facilities", "Delete facility records", "facilities", "delete"},

		// Laboratory permissions
		{"Create Lab Results", "Create laboratory test results", "laboratory", "create"},
		{"View Lab Results", "View laboratory test results", "laboratory", "read"},
		{"Update Lab Results", "Update laboratory test results", "laboratory", "update"},
		{"Delete Lab Results", "Delete laboratory test results", "laboratory", "delete"},

		// Surveillance permissions
		{"Create Surveillance Data", "Create surveillance data", "surveillance", "create"},
		{"View Surveillance Data", "View surveillance data", "surveillance", "read"},
		{"Update Surveillance Data", "Update surveillance data", "surveillance", "update"},
		{"Delete Surveillance Data", "Delete surveillance data", "surveillance", "delete"},
	}

	for _, perm := range permissions {
		_, err := db.ExecContext(ctx, `
			INSERT INTO permissions (name, description, resource, action) 
			VALUES ($1, $2, $3, $4) 
			ON CONFLICT (resource, action) DO NOTHING
		`, perm.name, perm.description, perm.resource, perm.action)
		if err != nil {
			return fmt.Errorf("error creating permission %s: %v", perm.name, err)
		}
	}

	// 4. Assign permissions to roles
	fmt.Println("Assigning permissions to roles...")

	// Super Admin gets all permissions
	_, err := db.ExecContext(ctx, `
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT r.id, p.id FROM roles r, permissions p
		WHERE r.name = 'super_admin'
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("error assigning permissions to super_admin: %v", err)
	}

	// Admin gets most permissions except super admin functions
	_, err = db.ExecContext(ctx, `
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT r.id, p.id FROM roles r, permissions p
		WHERE r.name = 'admin' AND p.resource != 'users'
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("error assigning permissions to admin: %v", err)
	}

	// Manager gets read and update permissions for most resources
	_, err = db.ExecContext(ctx, `
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT r.id, p.id FROM roles r, permissions p
		WHERE r.name = 'manager' AND p.action IN ('read', 'update', 'export')
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("error assigning permissions to manager: %v", err)
	}

	// Data Entry gets create and read permissions
	_, err = db.ExecContext(ctx, `
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT r.id, p.id FROM roles r, permissions p
		WHERE r.name = 'data_entry' AND p.action IN ('create', 'read')
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("error assigning permissions to data_entry: %v", err)
	}

	// Viewer gets only read permissions
	_, err = db.ExecContext(ctx, `
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT r.id, p.id FROM roles r, permissions p
		WHERE r.name = 'viewer' AND p.action = 'read'
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("error assigning permissions to viewer: %v", err)
	}

	// Lab Technician gets lab-specific permissions
	_, err = db.ExecContext(ctx, `
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT r.id, p.id FROM roles r, permissions p
		WHERE r.name = 'lab_technician' AND p.resource = 'laboratory'
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("error assigning permissions to lab_technician: %v", err)
	}

	// Surveillance Officer gets surveillance permissions
	_, err = db.ExecContext(ctx, `
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT r.id, p.id FROM roles r, permissions p
		WHERE r.name = 'surveillance_officer' AND p.resource = 'surveillance'
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("error assigning permissions to surveillance_officer: %v", err)
	}

	// 5. Create default admin user if it doesn't exist
	fmt.Println("Creating default admin user...")
	var adminUserID int
	err = db.QueryRowContext(ctx, `
		SELECT user_id FROM users WHERE user_name = 'admin'
	`).Scan(&adminUserID)

	if err == sql.ErrNoRows {
		// Create admin user
		err = db.QueryRowContext(ctx, `
			INSERT INTO users (user_name, user_pass, email, first_name, last_name, is_active, created_at)
			VALUES ('admin', '5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8', 'admin@system.local', 'System', 'Administrator', true, $1)
			RETURNING user_id
		`, time.Now()).Scan(&adminUserID)
		if err != nil {
			return fmt.Errorf("error creating admin user: %v", err)
		}
		fmt.Println("Created admin user with ID:", adminUserID)
	} else if err != nil {
		return fmt.Errorf("error checking for admin user: %v", err)
	} else {
		fmt.Println("Admin user already exists with ID:", adminUserID)
	}

	// 6. Assign super_admin role to admin user
	fmt.Println("Assigning super_admin role to admin user...")
	_, err = db.ExecContext(ctx, `
		INSERT INTO user_roles (user_id, role_id)
		SELECT $1, r.id FROM roles r WHERE r.name = 'super_admin'
		ON CONFLICT (user_id, role_id) DO NOTHING
	`, adminUserID)
	if err != nil {
		return fmt.Errorf("error assigning super_admin role: %v", err)
	}

	// 7. Assign default roles to existing users
	fmt.Println("Assigning default roles to existing users...")
	_, err = db.ExecContext(ctx, `
		INSERT INTO user_roles (user_id, role_id)
		SELECT u.user_id, r.id 
		FROM users u, roles r 
		WHERE r.name = 'data_entry' 
		  AND u.user_id NOT IN (SELECT user_id FROM user_roles WHERE role_id = r.id)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("error assigning default roles: %v", err)
	}

	fmt.Println("RBAC initialization completed successfully!")
	return nil
}
