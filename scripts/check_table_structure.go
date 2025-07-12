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

	// Check afi_facilities table structure
	fmt.Println("\n1. Checking afi_facilities table structure...")
	if err := checkTableStructure(db, "afi_facilities"); err != nil {
		log.Printf("Error checking afi_facilities: %v", err)
	}

	// Check facility table structure
	fmt.Println("\n2. Checking facility table structure...")
	if err := checkTableStructure(db, "facility"); err != nil {
		log.Printf("Error checking facility: %v", err)
	}

	// Check if afi_facilities has data
	fmt.Println("\n3. Checking afi_facilities data...")
	if err := checkTableData(db, "afi_facilities"); err != nil {
		log.Printf("Error checking afi_facilities data: %v", err)
	}

	// Check if facility has data
	fmt.Println("\n4. Checking facility data...")
	if err := checkTableData(db, "facility"); err != nil {
		log.Printf("Error checking facility data: %v", err)
	}

	fmt.Println("\n✓ Table structure check completed!")
}

func checkTableStructure(db *sql.DB, tableName string) error {
	// Check if table exists
	var exists bool
	err := db.QueryRowContext(context.Background(), `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = $1
		)
	`, tableName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if table exists: %v", err)
	}

	if !exists {
		fmt.Printf("    ⚠️  Table '%s' does not exist\n", tableName)
		return nil
	}

	fmt.Printf("    ✓ Table '%s' exists\n", tableName)

	// Get table columns
	rows, err := db.QueryContext(context.Background(), `
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns 
		WHERE table_name = $1 
		ORDER BY ordinal_position
	`, tableName)
	if err != nil {
		return fmt.Errorf("failed to get table columns: %v", err)
	}
	defer rows.Close()

	fmt.Printf("    Columns in '%s':\n", tableName)
	for rows.Next() {
		var colName, dataType, isNullable, columnDefault sql.NullString
		if err := rows.Scan(&colName, &dataType, &isNullable, &columnDefault); err != nil {
			return fmt.Errorf("failed to scan column info: %v", err)
		}
		fmt.Printf("      - %s (%s, nullable: %s", colName.String, dataType.String, isNullable.String)
		if columnDefault.Valid {
			fmt.Printf(", default: %s", columnDefault.String)
		}
		fmt.Printf(")\n")
	}

	return nil
}

func checkTableData(db *sql.DB, tableName string) error {
	// Check if table exists first
	var exists bool
	err := db.QueryRowContext(context.Background(), `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = $1
		)
	`, tableName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if table exists: %v", err)
	}

	if !exists {
		fmt.Printf("    ⚠️  Table '%s' does not exist, skipping data check\n", tableName)
		return nil
	}

	// Get row count
	var count int
	err = db.QueryRowContext(context.Background(), fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to get row count: %v", err)
	}

	fmt.Printf("    ✓ Table '%s' has %d rows\n", tableName, count)

	// Show sample data (first 5 rows)
	if count > 0 {
		fmt.Printf("    Sample data from '%s':\n", tableName)
		rows, err := db.QueryContext(context.Background(), fmt.Sprintf("SELECT * FROM %s LIMIT 5", tableName))
		if err != nil {
			return fmt.Errorf("failed to get sample data: %v", err)
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("failed to get columns: %v", err)
		}

		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		rowCount := 0
		for rows.Next() {
			err := rows.Scan(valuePtrs...)
			if err != nil {
				return fmt.Errorf("failed to scan row: %v", err)
			}

			fmt.Printf("      Row %d:\n", rowCount+1)
			for i, col := range columns {
				val := values[i]
				if val == nil {
					fmt.Printf("        %s: NULL\n", col)
				} else {
					fmt.Printf("        %s: %v\n", col, val)
				}
			}
			rowCount++
		}
	}

	return nil
}
