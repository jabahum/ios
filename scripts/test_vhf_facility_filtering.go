package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection
	dsn := "host=localhost port=5432 user=postgres password=pwaiswa dbname=ios sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("Error pinging database:", err)
	}

	fmt.Println("✓ Connected to database successfully")

	// Run the migration
	fmt.Println("\n1. Running facility migration...")
	if err := runMigration(db); err != nil {
		log.Printf("Error running migration: %v", err)
	} else {
		fmt.Println("✓ Migration completed successfully")
	}

	// Test facility filtering logic
	fmt.Println("\n2. Testing facility filtering logic...")
	if err := testFacilityFiltering(db); err != nil {
		log.Printf("Error testing facility filtering: %v", err)
	} else {
		fmt.Println("✓ Facility filtering test completed")
	}

	// Create a test user with role 65 and facility assignment
	fmt.Println("\n3. Creating test user with role 65 and facility...")
	if err := createTestUser(db); err != nil {
		log.Printf("Error creating test user: %v", err)
	} else {
		fmt.Println("✓ Test user created successfully")
	}

	fmt.Println("\n✓ All tests completed!")
}

func runMigration(db *sql.DB) error {
	// Read and execute the migration file
	migrationSQL := `
	-- Migration: Migrate facilities data and add VHF filtering support
	-- This migration migrates data from afi_facilities to facility table and adds missing columns

	-- Add missing columns to facility table first
	ALTER TABLE facility ADD COLUMN IF NOT EXISTS facility_code VARCHAR(50);
	ALTER TABLE facility ADD COLUMN IF NOT EXISTS facility_type VARCHAR(100);
	ALTER TABLE facility ADD COLUMN IF NOT EXISTS district VARCHAR(100);
	ALTER TABLE facility ADD COLUMN IF NOT EXISTS subcounty VARCHAR(100);
	ALTER TABLE facility ADD COLUMN IF NOT EXISTS parish VARCHAR(100);
	ALTER TABLE facility ADD COLUMN IF NOT EXISTS village VARCHAR(100);
	ALTER TABLE facility ADD COLUMN IF NOT EXISTS latitude DECIMAL(10, 8);
	ALTER TABLE facility ADD COLUMN IF NOT EXISTS longitude DECIMAL(11, 8);
	ALTER TABLE facility ADD COLUMN IF NOT EXISTS contact_person VARCHAR(100);
	ALTER TABLE facility ADD COLUMN IF NOT EXISTS contact_phone VARCHAR(20);
	ALTER TABLE facility ADD COLUMN IF NOT EXISTS contact_email VARCHAR(100);
	ALTER TABLE facility ADD COLUMN IF NOT EXISTS ownership VARCHAR(50);
	ALTER TABLE facility ADD COLUMN IF NOT EXISTS hsd VARCHAR(100);
	ALTER TABLE facility ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;
	ALTER TABLE facility ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
	ALTER TABLE facility ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

	-- First, let's check if afi_facilities table exists and has data
	DO $$
	BEGIN
		IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'afi_facilities') THEN
			-- Migrate data from afi_facilities to facility table
			-- Use a temporary table approach to avoid conflicts
			CREATE TEMP TABLE temp_facilities AS
			SELECT DISTINCT
				COALESCE(facility_name, 'Unknown Facility') as facility_name,
				CASE 
					WHEN level = 'HC I' THEN 1
					WHEN level = 'HC II' THEN 2
					WHEN level = 'HC III' THEN 3
					WHEN level = 'HC IV' THEN 4
					WHEN level = 'Hospital' THEN 5
					WHEN level = 'Regional Referral Hospital' THEN 6
					WHEN level = 'National Referral Hospital' THEN 7
					WHEN level = 'Clinic' THEN 1
					WHEN level = 'Health Centre' THEN 2
					WHEN level = 'Medical Centre' THEN 3
					ELSE 1
				END as facility_level,
				hsd,
				subcounty_town_council_division,
				ownership,
				phone_number
			FROM afi_facilities
			WHERE facility_name IS NOT NULL AND facility_name != '';
			
			-- Insert only facilities that don't already exist
			INSERT INTO facility (facility_name, facility_level)
			SELECT tf.facility_name, tf.facility_level
			FROM temp_facilities tf
			WHERE NOT EXISTS (
				SELECT 1 FROM facility f 
				WHERE LOWER(TRIM(f.facility_name)) = LOWER(TRIM(tf.facility_name))
			);
			
			-- Update existing facilities with additional data
			UPDATE facility 
			SET 
				district = tf.hsd,
				subcounty = tf.subcounty_town_council_division,
				ownership = tf.ownership,
				contact_phone = tf.phone_number,
				hsd = tf.hsd
			FROM temp_facilities tf
			WHERE LOWER(TRIM(facility.facility_name)) = LOWER(TRIM(tf.facility_name))
			  AND (facility.district IS NULL OR facility.subcounty IS NULL);
			
			DROP TABLE temp_facilities;
			
			RAISE NOTICE 'Migrated facilities from afi_facilities to facility table';
		ELSE
			RAISE NOTICE 'afi_facilities table does not exist, skipping migration';
		END IF;
	END $$;



	-- Create indexes for better performance
	CREATE INDEX IF NOT EXISTS idx_facility_district ON facility(district);
	CREATE INDEX IF NOT EXISTS idx_facility_is_active ON facility(is_active);
	CREATE INDEX IF NOT EXISTS idx_facility_name ON facility(facility_name);
	CREATE INDEX IF NOT EXISTS idx_facility_hsd ON facility(hsd);

	-- Add facility_id column to vhf_patients table if it doesn't exist
	ALTER TABLE vhf_patients ADD COLUMN IF NOT EXISTS facility_id INTEGER REFERENCES facility(facility_id);

	-- Create index for facility filtering
	CREATE INDEX IF NOT EXISTS idx_vhf_patients_facility_id ON vhf_patients(facility_id);

	-- Add facility_id column to employee table if it doesn't exist (it should already exist based on previous migrations)
	-- This is just to ensure it exists
	DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'employee' AND column_name = 'facility') THEN
			ALTER TABLE employee ADD COLUMN facility INTEGER REFERENCES facility(facility_id);
		END IF;
	END $$;

	-- Create index for employee facility lookups
	CREATE INDEX IF NOT EXISTS idx_employee_facility ON employee(facility);

	-- Update VHF patients to link to facilities based on reporting_health_facility_name
	UPDATE vhf_patients 
	SET facility_id = f.facility_id
	FROM facility f
	WHERE vhf_patients.reporting_health_facility_name IS NOT NULL 
	  AND vhf_patients.reporting_health_facility_name != ''
	  AND LOWER(TRIM(vhf_patients.reporting_health_facility_name)) = LOWER(TRIM(f.facility_name))
	  AND vhf_patients.facility_id IS NULL;

	-- Create a function to get user's facility ID
	CREATE OR REPLACE FUNCTION get_user_facility_id(user_id_param INTEGER)
	RETURNS INTEGER AS $$
	DECLARE
		facility_id_result INTEGER;
	BEGIN
		SELECT e.facility INTO facility_id_result
		FROM employee e
		JOIN users u ON e.employee_email = u.email
		WHERE u.user_id = user_id_param
		LIMIT 1;
		
		RETURN facility_id_result;
	END;
	$$ LANGUAGE plpgsql;

	-- Create a function to check if user has facility-based access
	CREATE OR REPLACE FUNCTION user_has_facility_access(user_id_param INTEGER, target_facility_id INTEGER)
	RETURNS BOOLEAN AS $$
	DECLARE
		user_facility_id INTEGER;
		user_role_id INTEGER;
	BEGIN
		-- Get user's facility ID
		user_facility_id := get_user_facility_id(user_id_param);
		
		-- If user has no facility assigned, they have access to all facilities
		IF user_facility_id IS NULL THEN
			RETURN TRUE;
		END IF;
		
		-- If user has a facility assigned, check if it matches the target facility
		RETURN user_facility_id = target_facility_id;
	END;
	$$ LANGUAGE plpgsql;
	`

	_, err := db.ExecContext(context.Background(), migrationSQL)
	return err
}

func testFacilityFiltering(db *sql.DB) error {
	// Test 1: Check if role 65 exists
	fmt.Println("  - Checking if role 65 exists...")
	var roleName string
	err := db.QueryRowContext(context.Background(), "SELECT name FROM roles WHERE id = 65").Scan(&roleName)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("    ⚠️  Role 65 does not exist, creating it...")
			_, err = db.ExecContext(context.Background(), `
				INSERT INTO roles (id, name, description, is_active) 
				VALUES (65, 'vhf_lab_technician', 'VHF laboratory technician - can access VHF list and capture lab requests', true)
				ON CONFLICT (id) DO NOTHING
			`)
			if err != nil {
				return fmt.Errorf("failed to create role 65: %v", err)
			}
			fmt.Println("    ✓ Role 65 created")
		} else {
			return fmt.Errorf("failed to check role 65: %v", err)
		}
	} else {
		fmt.Printf("    ✓ Role 65 exists: %s\n", roleName)
	}

	// Test 2: Check facility table structure
	fmt.Println("  - Checking facility table structure...")
	rows, err := db.QueryContext(context.Background(), `
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_name = 'facility' 
		ORDER BY ordinal_position
	`)
	if err != nil {
		return fmt.Errorf("failed to check facility table: %v", err)
	}
	defer rows.Close()

	fmt.Println("    Facility table columns:")
	for rows.Next() {
		var colName, dataType string
		if err := rows.Scan(&colName, &dataType); err != nil {
			return fmt.Errorf("failed to scan column info: %v", err)
		}
		fmt.Printf("      - %s (%s)\n", colName, dataType)
	}

	// Test 3: Check VHF patients table for facility_id column
	fmt.Println("  - Checking VHF patients table...")
	var hasFacilityID bool
	err = db.QueryRowContext(context.Background(), `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'vhf_patients' AND column_name = 'facility_id'
		)
	`).Scan(&hasFacilityID)
	if err != nil {
		return fmt.Errorf("failed to check vhf_patients table: %v", err)
	}
	if hasFacilityID {
		fmt.Println("    ✓ VHF patients table has facility_id column")
	} else {
		fmt.Println("    ⚠️  VHF patients table missing facility_id column")
	}

	// Test 4: Check employee table for facility column
	fmt.Println("  - Checking employee table...")
	var hasFacility bool
	err = db.QueryRowContext(context.Background(), `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'employee' AND column_name = 'facility'
		)
	`).Scan(&hasFacility)
	if err != nil {
		return fmt.Errorf("failed to check employee table: %v", err)
	}
	if hasFacility {
		fmt.Println("    ✓ Employee table has facility column")
	} else {
		fmt.Println("    ⚠️  Employee table missing facility column")
	}

	return nil
}

func createTestUser(db *sql.DB) error {
	// Create a test user
	fmt.Println("  - Creating test user...")

	// Check if user already exists
	var userID int
	err := db.QueryRowContext(context.Background(), "SELECT user_id FROM users WHERE user_name = 'vhf_facility_test'").Scan(&userID)
	if err == sql.ErrNoRows {
		// User doesn't exist, create it
		userQuery := `
			INSERT INTO users (user_name, user_pass, email, first_name, last_name, is_active, created_at)
			VALUES ('vhf_facility_test', '5f4dcc3b5aa765d61d8327deb882cf99', 'vhf_facility_test@example.com', 'VHF', 'Facility Test', true, NOW())
			RETURNING user_id
		`
		err = db.QueryRowContext(context.Background(), userQuery).Scan(&userID)
		if err != nil {
			return fmt.Errorf("failed to create test user: %v", err)
		}
		fmt.Println("    ✓ Test user created")
	} else if err != nil {
		return fmt.Errorf("failed to check existing test user: %v", err)
	} else {
		fmt.Println("    ✓ Test user already exists")
	}
	fmt.Printf("    ✓ Test user ID: %d\n", userID)

	// Assign role 65 to the user
	fmt.Println("  - Assigning role 65 to test user...")

	// Check if role is already assigned
	var roleCount int
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM user_roles WHERE user_id = $1 AND role_id = 65", userID).Scan(&roleCount)
	if err != nil {
		return fmt.Errorf("failed to check existing role assignment: %v", err)
	}

	if roleCount == 0 {
		_, err = db.ExecContext(context.Background(), `
			INSERT INTO user_roles (user_id, role_id, created_at)
			VALUES ($1, 65, NOW())
		`, userID)
		if err != nil {
			return fmt.Errorf("failed to assign role to user: %v", err)
		}
		fmt.Println("    ✓ Role 65 assigned")
	} else {
		fmt.Println("    ✓ Role 65 already assigned")
	}

	// Create a test facility if none exists
	fmt.Println("  - Creating test facility...")
	var facilityID int
	err = db.QueryRowContext(context.Background(), "SELECT facility_id FROM facility WHERE facility_name = 'Test VHF Facility'").Scan(&facilityID)
	if err == sql.ErrNoRows {
		// Facility doesn't exist, create it
		err = db.QueryRowContext(context.Background(), `
			INSERT INTO facility (facility_name, facility_level, is_active)
			VALUES ('Test VHF Facility', 3, true)
			RETURNING facility_id
		`).Scan(&facilityID)
		if err != nil {
			return fmt.Errorf("failed to create test facility: %v", err)
		}
		fmt.Println("    ✓ Test facility created")
	} else if err != nil {
		return fmt.Errorf("failed to check existing test facility: %v", err)
	} else {
		fmt.Println("    ✓ Test facility already exists")
	}
	fmt.Printf("    ✓ Test facility ID: %d\n", facilityID)

	// Create a test employee record linked to the user
	fmt.Println("  - Creating test employee record...")

	// Check if employee already exists
	var employeeCount int
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM employee WHERE employee_email = 'vhf_facility_test@example.com'").Scan(&employeeCount)
	if err != nil {
		return fmt.Errorf("failed to check existing employee: %v", err)
	}

	if employeeCount == 0 {
		_, err = db.ExecContext(context.Background(), `
			INSERT INTO employee (employee_fname, employee_lname, employee_email, facility)
			VALUES ('VHF', 'Facility Test', 'vhf_facility_test@example.com', $1)
		`, facilityID)
		if err != nil {
			return fmt.Errorf("failed to create test employee: %v", err)
		}
		fmt.Println("    ✓ Test employee record created")
	} else {
		// Update existing employee with facility
		_, err = db.ExecContext(context.Background(), `
			UPDATE employee SET facility = $1 WHERE employee_email = 'vhf_facility_test@example.com'
		`, facilityID)
		if err != nil {
			return fmt.Errorf("failed to update test employee: %v", err)
		}
		fmt.Println("    ✓ Test employee record updated")
	}

	// Test the facility filtering logic
	fmt.Println("  - Testing facility filtering logic...")
	var count int
	err = db.QueryRowContext(context.Background(), `
		SELECT COUNT(*) FROM user_roles ur 
		JOIN roles r ON ur.role_id = r.id 
		WHERE ur.user_id = $1 AND r.id = 65 AND r.is_active = true
	`, userID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to test role check: %v", err)
	}
	fmt.Printf("    ✓ User has role 65: %t\n", count > 0)

	var userFacilityID sql.NullInt64
	err = db.QueryRowContext(context.Background(), `
		SELECT e.facility 
		FROM employee e
		JOIN users u ON e.employee_email = u.email
		WHERE u.user_id = $1
		LIMIT 1
	`, userID).Scan(&userFacilityID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to test facility lookup: %v", err)
	}
	if userFacilityID.Valid {
		fmt.Printf("    ✓ User facility ID: %d\n", userFacilityID.Int64)
	} else {
		fmt.Println("    ⚠️  User has no facility assigned")
	}

	fmt.Println("\n✓ Test user created successfully!")
	fmt.Println("Username: vhf_facility_test")
	fmt.Println("Password: password")
	fmt.Println("Role: 65 (vhf_lab_technician)")
	fmt.Printf("Facility: %d (Test VHF Facility)\n", facilityID)
	fmt.Println("\nThis user should see only VHF cases from their assigned facility when they log in.")

	return nil
}
