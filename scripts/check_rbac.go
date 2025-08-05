package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection
	dsn := "host=localhost port=5432 user=postgres password=pwaiswa dbname=ios sslmode=disable"

	// Connect to database
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Error pinging database:", err)
	}
	fmt.Println("✓ Connected to database successfully")

	// Check RBAC tables
	tables := []string{"roles", "permissions", "user_roles", "role_permissions"}

	for _, table := range tables {
		fmt.Printf("\n1. Checking %s table structure...\n", table)

		// Check if table exists
		var exists bool
		query := `SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)`

		err := db.QueryRow(query, table).Scan(&exists)
		if err != nil {
			fmt.Printf("    ✗ Error checking if table '%s' exists: %v\n", table, err)
			continue
		}

		if !exists {
			fmt.Printf("    ✗ Table '%s' does not exist\n", table)
			continue
		}

		fmt.Printf("    ✓ Table '%s' exists\n", table)

		// Get table structure
		columnsQuery := `
			SELECT column_name, data_type, is_nullable, column_default
			FROM information_schema.columns 
			WHERE table_schema = 'public' 
			AND table_name = $1 
			ORDER BY ordinal_position
		`

		rows, err := db.Query(columnsQuery, table)
		if err != nil {
			fmt.Printf("    ✗ Error getting columns for '%s': %v\n", table, err)
			continue
		}
		defer rows.Close()

		fmt.Printf("    Columns in '%s':\n", table)
		for rows.Next() {
			var columnName, dataType, isNullable, columnDefault sql.NullString
			err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault)
			if err != nil {
				fmt.Printf("      ✗ Error scanning column: %v\n", err)
				continue
			}

			nullable := "YES"
			if isNullable.String == "NO" {
				nullable = "NO"
			}

			defaultVal := "NULL"
			if columnDefault.Valid {
				defaultVal = columnDefault.String
			}

			fmt.Printf("      - %s (%s, nullable: %s, default: %s)\n",
				columnName.String, dataType.String, nullable, defaultVal)
		}

		// Get row count
		var count int
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
		err = db.QueryRow(countQuery).Scan(&count)
		if err != nil {
			fmt.Printf("    ✗ Error getting row count for '%s': %v\n", table, err)
			continue
		}
		fmt.Printf("    ✓ Table '%s' has %d rows\n", table)

		// Get sample data
		if count > 0 {
			sampleQuery := fmt.Sprintf("SELECT * FROM %s LIMIT 3", table)
			sampleRows, err := db.Query(sampleQuery)
			if err != nil {
				fmt.Printf("    ✗ Error getting sample data for '%s': %v\n", table, err)
				continue
			}
			defer sampleRows.Close()

			columns, err := sampleRows.Columns()
			if err != nil {
				fmt.Printf("    ✗ Error getting columns for sample data: %v\n", err)
				continue
			}

			fmt.Printf("    Sample data from '%s':\n", table)
			rowNum := 1
			for sampleRows.Next() {
				// Create a slice of interface{} to hold the values
				values := make([]interface{}, len(columns))
				valuePtrs := make([]interface{}, len(columns))
				for i := range values {
					valuePtrs[i] = &values[i]
				}

				err := sampleRows.Scan(valuePtrs...)
				if err != nil {
					fmt.Printf("      ✗ Error scanning sample row: %v\n", err)
					continue
				}

				fmt.Printf("      Row %d:\n", rowNum)
				for i, col := range columns {
					val := values[i]
					if val == nil {
						fmt.Printf("        %s: NULL\n", col)
					} else {
						fmt.Printf("        %s: %v\n", col, val)
					}
				}
				rowNum++
			}
		}
	}

	fmt.Println("\n✓ RBAC table structure check completed!")
}
