package main

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "pwaiswa")
	dbName := getEnv("DB_NAME", "ios")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

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

	// Create VHF Lab Technician user
	fmt.Println("\n1. Creating VHF Lab Technician user...")
	vhfLabUserID, err := createUser(db, "vhf_lab", "VHF Lab Technician", "vhf_lab@example.com", "password123")
	if err != nil {
		log.Printf("Error creating VHF lab user: %v", err)
	} else {
		fmt.Printf("✓ Created VHF Lab Technician user (ID: %d)\n", vhfLabUserID)

		// Assign VHF lab technician role
		if err := assignRoleToUser(db, vhfLabUserID, "vhf_lab_technician"); err != nil {
			log.Printf("Error assigning role to VHF lab user: %v", err)
		} else {
			fmt.Println("✓ Assigned vhf_lab_technician role")
		}
	}

	// Create VHF Data Entry user
	fmt.Println("\n2. Creating VHF Data Entry user...")
	vhfDataUserID, err := createUser(db, "vhf_data", "VHF Data Entry", "vhf_data@example.com", "password123")
	if err != nil {
		log.Printf("Error creating VHF data user: %v", err)
	} else {
		fmt.Printf("✓ Created VHF Data Entry user (ID: %d)\n", vhfDataUserID)

		// Assign VHF data entry role
		if err := assignRoleToUser(db, vhfDataUserID, "vhf_data_entry"); err != nil {
			log.Printf("Error assigning role to VHF data user: %v", err)
		} else {
			fmt.Println("✓ Assigned vhf_data_entry role")
		}
	}

	// Create MPOX Case Manager user
	fmt.Println("\n3. Creating MPOX Case Manager user...")
	mpoxManagerUserID, err := createUser(db, "mpox_manager", "MPOX Case Manager", "mpox_manager@example.com", "password123")
	if err != nil {
		log.Printf("Error creating MPOX manager user: %v", err)
	} else {
		fmt.Printf("✓ Created MPOX Case Manager user (ID: %d)\n", mpoxManagerUserID)

		// Assign MPOX case manager role
		if err := assignRoleToUser(db, mpoxManagerUserID, "mpox_case_manager"); err != nil {
			log.Printf("Error assigning role to MPOX manager user: %v", err)
		} else {
			fmt.Println("✓ Assigned mpox_case_manager role")
		}
	}

	// Create MPOX Data Entry user
	fmt.Println("\n4. Creating MPOX Data Entry user...")
	mpoxDataUserID, err := createUser(db, "mpox_data", "MPOX Data Entry", "mpox_data@example.com", "password123")
	if err != nil {
		log.Printf("Error creating MPOX data user: %v", err)
	} else {
		fmt.Printf("✓ Created MPOX Data Entry user (ID: %d)\n", mpoxDataUserID)

		// Assign MPOX data entry role
		if err := assignRoleToUser(db, mpoxDataUserID, "mpox_data_entry"); err != nil {
			log.Printf("Error assigning role to MPOX data user: %v", err)
		} else {
			fmt.Println("✓ Assigned mpox_data_entry role")
		}
	}

	// Create Outbreak Viewer user
	fmt.Println("\n5. Creating Outbreak Viewer user...")
	outbreakViewerUserID, err := createUser(db, "outbreak_viewer", "Outbreak Viewer", "outbreak_viewer@example.com", "password123")
	if err != nil {
		log.Printf("Error creating outbreak viewer user: %v", err)
	} else {
		fmt.Printf("✓ Created Outbreak Viewer user (ID: %d)\n", outbreakViewerUserID)

		// Assign outbreak viewer role
		if err := assignRoleToUser(db, outbreakViewerUserID, "outbreak_viewer"); err != nil {
			log.Printf("Error assigning role to outbreak viewer user: %v", err)
		} else {
			fmt.Println("✓ Assigned outbreak_viewer role")
		}
	}

	// Create Outbreak Manager user
	fmt.Println("\n6. Creating Outbreak Manager user...")
	outbreakManagerUserID, err := createUser(db, "outbreak_manager", "Outbreak Manager", "outbreak_manager@example.com", "password123")
	if err != nil {
		log.Printf("Error creating outbreak manager user: %v", err)
	} else {
		fmt.Printf("✓ Created Outbreak Manager user (ID: %d)\n", outbreakManagerUserID)

		// Assign outbreak manager role
		if err := assignRoleToUser(db, outbreakManagerUserID, "outbreak_manager"); err != nil {
			log.Printf("Error assigning role to outbreak manager user: %v", err)
		} else {
			fmt.Println("✓ Assigned outbreak_manager role")
		}
	}

	// Assign users to specific outbreaks
	fmt.Println("\n7. Assigning users to outbreaks...")

	// Get outbreak IDs
	vhfOutbreakID, err := getOutbreakID(db, "vhf")
	if err != nil {
		log.Printf("Error getting VHF outbreak ID: %v", err)
	} else {
		// Assign VHF users to VHF outbreak
		if vhfLabUserID > 0 {
			if err := assignUserToOutbreak(db, vhfLabUserID, vhfOutbreakID, 1); err != nil {
				log.Printf("Error assigning VHF lab user to outbreak: %v", err)
			} else {
				fmt.Println("✓ Assigned VHF Lab Technician to VHF outbreak")
			}
		}
		if vhfDataUserID > 0 {
			if err := assignUserToOutbreak(db, vhfDataUserID, vhfOutbreakID, 1); err != nil {
				log.Printf("Error assigning VHF data user to outbreak: %v", err)
			} else {
				fmt.Println("✓ Assigned VHF Data Entry to VHF outbreak")
			}
		}
	}

	mpoxOutbreakID, err := getOutbreakID(db, "mpox")
	if err != nil {
		log.Printf("Error getting MPOX outbreak ID: %v", err)
	} else {
		// Assign MPOX users to MPOX outbreak
		if mpoxManagerUserID > 0 {
			if err := assignUserToOutbreak(db, mpoxManagerUserID, mpoxOutbreakID, 1); err != nil {
				log.Printf("Error assigning MPOX manager user to outbreak: %v", err)
			} else {
				fmt.Println("✓ Assigned MPOX Case Manager to MPOX outbreak")
			}
		}
		if mpoxDataUserID > 0 {
			if err := assignUserToOutbreak(db, mpoxDataUserID, mpoxOutbreakID, 1); err != nil {
				log.Printf("Error assigning MPOX data user to outbreak: %v", err)
			} else {
				fmt.Println("✓ Assigned MPOX Data Entry to MPOX outbreak")
			}
		}
	}

	fmt.Println("\n✓ Setup completed successfully!")
	fmt.Println("\nUser credentials:")
	fmt.Println("VHF Lab Technician: vhf_lab / password123")
	fmt.Println("VHF Data Entry: vhf_data / password123")
	fmt.Println("MPOX Case Manager: mpox_manager / password123")
	fmt.Println("MPOX Data Entry: mpox_data / password123")
	fmt.Println("Outbreak Viewer: outbreak_viewer / password123")
	fmt.Println("Outbreak Manager: outbreak_manager / password123")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func hashPasswordSHA1(password string) string {
	h := sha1.New()
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}

func createUser(db *sql.DB, username, fullName, email, password string) (int, error) {
	// Check if user already exists
	var userID int
	err := db.QueryRow("SELECT user_id FROM users WHERE user_name = $1", username).Scan(&userID)
	if err == nil {
		// User exists, return existing ID
		return userID, nil
	}

	// Split full name into first and last name
	names := strings.Split(fullName, " ")
	firstName := names[0]
	lastName := ""
	if len(names) > 1 {
		lastName = strings.Join(names[1:], " ")
	}

	// Create new user
	query := `
		INSERT INTO users (user_name, first_name, last_name, email, user_pass, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
		RETURNING user_id
	`

	// Simple password hash (in production, use proper hashing)
	passwordHash := hashPasswordSHA1(password)

	err = db.QueryRow(query, username, firstName, lastName, email, passwordHash).Scan(&userID)
	return userID, err
}

func assignRoleToUser(db *sql.DB, userID int, roleName string) error {
	// Get role ID
	var roleID int
	err := db.QueryRow("SELECT id FROM roles WHERE name = $1", roleName).Scan(&roleID)
	if err != nil {
		return fmt.Errorf("role %s not found: %v", roleName, err)
	}

	// Check if assignment already exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM user_roles WHERE user_id = $1 AND role_id = $2)", userID, roleID).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return nil // Already assigned
	}

	// Assign role
	_, err = db.Exec("INSERT INTO user_roles (user_id, role_id, created_at) VALUES ($1, $2, NOW())", userID, roleID)
	return err
}

func getOutbreakID(db *sql.DB, outbreakType string) (int, error) {
	var outbreakID int
	err := db.QueryRow("SELECT id FROM outbreaks WHERE outbreak_type = $1 AND status = 'active' LIMIT 1", outbreakType).Scan(&outbreakID)
	return outbreakID, err
}

func assignUserToOutbreak(db *sql.DB, userID, outbreakID, assignedBy int) error {
	// Check if assignment already exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM user_outbreaks WHERE user_id = $1 AND outbreak_id = $2 AND is_active = true)", userID, outbreakID).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return nil // Already assigned
	}

	// Assign user to outbreak
	_, err = db.Exec(`
		INSERT INTO user_outbreaks (user_id, outbreak_id, assigned_by, is_active, assigned_at)
		VALUES ($1, $2, $3, true, NOW())
	`, userID, outbreakID, assignedBy)
	return err
}
