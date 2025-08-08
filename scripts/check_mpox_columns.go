package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// Connect to database
	db, err := sql.Open("postgres", "host=localhost user=postgres password=postgres dbname=ios sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Get all columns from mpox_onset_vitals
	rows, err := db.Query("SELECT column_name FROM information_schema.columns WHERE table_name = 'mpox_onset_vitals' ORDER BY ordinal_position")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Columns in mpox_onset_vitals table:")
	count := 0
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%d. %s\n", count+1, columnName)
		count++
	}
	fmt.Printf("\nTotal columns: %d\n", count)
}
