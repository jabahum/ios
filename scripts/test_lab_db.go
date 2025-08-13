package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"case/internal/models"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection string
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:pwaiswa@localhost:5432/ios?sslmode=disable"
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	fmt.Println("✅ Connected to database successfully!")

	// Test LabBloodTypes function
	fmt.Println("\n🧪 Testing LabBloodTypes function:")
	ctx := context.Background()
	bloodTypes, err := models.LabBloodTypes(ctx, db)
	if err != nil {
		fmt.Printf("❌ Error calling LabBloodTypes: %v\n", err)
		return
	}
	fmt.Printf("✅ LabBloodTypes returned %d records\n", len(bloodTypes))
	if len(bloodTypes) > 0 {
		fmt.Printf("   First record: %+v\n", bloodTypes[0])
	}

	// Test LabSwabTypes function
	fmt.Println("\n🧪 Testing LabSwabTypes function:")
	swabTypes, err := models.LabSwabTypes(ctx, db)
	if err != nil {
		fmt.Printf("❌ Error calling LabSwabTypes: %v\n", err)
		return
	}
	fmt.Printf("✅ LabSwabTypes returned %d records\n", len(swabTypes))
	if len(swabTypes) > 0 {
		fmt.Printf("   First record: %+v\n", swabTypes[0])
	}

	// Test LabUrineTypes function
	fmt.Println("\n🧪 Testing LabUrineTypes function:")
	urineTypes, err := models.LabUrineTypes(ctx, db)
	if err != nil {
		fmt.Printf("❌ Error calling LabUrineTypes: %v\n", err)
		return
	}
	fmt.Printf("✅ LabUrineTypes returned %d records\n", len(urineTypes))
	if len(urineTypes) > 0 {
		fmt.Printf("   First record: %+v\n", urineTypes[0])
	}

	fmt.Println("\n✅ All database tests completed!")
}
