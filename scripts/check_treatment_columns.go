package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// Connect to database
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/ios?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Query to get column information
	query := `
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_name = 'treatment' 
		AND table_schema = 'public' 
		ORDER BY ordinal_position
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Treatment table columns:")
	count := 0
	for rows.Next() {
		var columnName, dataType string
		err := rows.Scan(&columnName, &dataType)
		if err != nil {
			log.Fatal(err)
		}
		count++
		fmt.Printf("%d. %s (%s)\n", count, columnName, dataType)
	}

	fmt.Printf("\nTotal columns: %d\n", count)
}
