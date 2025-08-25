package main

import (
	"case/internal/reports"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection string - adjust as needed
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://username:password@localhost:5432/dbname?sslmode=disable"
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("✅ Database connection successful")

	// Create test filters
	filters := reports.ReportFilters{
		StartDate:  "2024-01-01",
		EndDate:    "2024-12-31",
		ReportType: "indicators",
	}

	fmt.Println("🧪 Testing indicators report functions...")

	// Test new admissions daily
	fmt.Println("\n📊 Testing new admissions daily...")
	newAdmissions, err := reports.GetNewAdmissionsDaily(db, filters)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Success: %+v\n", newAdmissions)
	}

	// Test cumulative confirmed cases
	fmt.Println("\n📊 Testing cumulative confirmed cases...")
	confirmed, err := reports.GetCumulativeConfirmedCases(db, filters)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Success: %+v\n", confirmed)
	}

	// Test cumulative suspected cases
	fmt.Println("\n📊 Testing cumulative suspected cases...")
	suspected, err := reports.GetCumulativeSuspectedCases(db, filters)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Success: %+v\n", suspected)
	}

	// Test cumulative deaths
	fmt.Println("\n📊 Testing cumulative deaths...")
	deaths, err := reports.GetCumulativeDeaths(db, filters)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Success: %+v\n", deaths)
	}

	// Test case fatality rate
	fmt.Println("\n📊 Testing case fatality rate...")
	cfr, err := reports.GetCaseFatalityRate(db, filters)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Success: %+v\n", cfr)
	}

	// Test current admissions
	fmt.Println("\n📊 Testing current admissions...")
	current, err := reports.GetCurrentAdmissions(db, filters)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Success: %+v\n", current)
	}

	// Test cumulative discharges
	fmt.Println("\n📊 Testing cumulative discharges...")
	discharges, err := reports.GetCumulativeDischarges(db, filters)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Success: %+v\n", discharges)
	}

	// Test severe cases
	fmt.Println("\n📊 Testing severe cases...")
	severe, err := reports.GetSevereCasesAdmitted(db, filters)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Success: %+v\n", severe)
	}

	// Test critical cases
	fmt.Println("\n📊 Testing critical cases...")
	critical, err := reports.GetCriticalCasesAdmitted(db, filters)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Success: %+v\n", critical)
	}

	// Test cases by sex
	fmt.Println("\n📊 Testing cases by sex...")
	sex, err := reports.GetCasesBySex(db, filters)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Success: %+v\n", sex)
	}

	// Test cases by age group
	fmt.Println("\n📊 Testing cases by age group...")
	age, err := reports.GetCasesByAgeGroup(db, filters)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Success: %+v\n", age)
	}

	// Test cases by location
	fmt.Println("\n📊 Testing cases by location...")
	location, err := reports.GetCasesByLocation(db, filters)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Success: %+v\n", location)
	}

	// Test HCW infections
	fmt.Println("\n📊 Testing HCW infections...")
	hcw, err := reports.GetHCWInfections(db, filters)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Success: %+v\n", hcw)
	}

	fmt.Println("\n🎉 Indicators report testing completed!")
}
