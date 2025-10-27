package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type Config struct {
	Ux string `json:"Ux"`
	Px string `json:"Px"`
	Dx string `json:"Dx"`
}

func main() {
	// Load config
	configData, err := os.ReadFile("cmd/web/config.json")
	if err != nil {
		log.Fatal("Error reading config file:", err)
	}

	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		log.Fatal("Error parsing config:", err)
	}

	// Test connection string
	connStr := fmt.Sprintf("host=db port=5432 user=%s password=%s dbname=%s sslmode=disable",
		config.Ux, config.Px, config.Dx)

	fmt.Printf("Testing connection to: %s@db:5432/%s\n", config.Ux, config.Dx)

	// Open connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error opening database:", err)
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(3 * time.Minute)

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("Testing database ping...")
	if err := db.PingContext(ctx); err != nil {
		fmt.Println("connstr:", connStr)
		log.Fatal("Database ping failed:", err)
	}
	fmt.Println("✓ Database ping successful")

	// Test simple query
	fmt.Println("Testing simple query...")
	var result int
	if err := db.QueryRowContext(ctx, "SELECT 1").Scan(&result); err != nil {
		log.Fatal("Simple query failed:", err)
	}
	fmt.Printf("✓ Simple query successful, result: %d\n", result)

	// Test users table query
	fmt.Println("Testing users table query...")
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		log.Fatal("Users table query failed:", err)
	}
	fmt.Printf("✓ Users table query successful, count: %d\n", count)

	// Test multiple connections
	fmt.Println("Testing multiple connections...")
	for i := 0; i < 5; i++ {
		var id int
		if err := db.QueryRowContext(ctx, "SELECT 1").Scan(&id); err != nil {
			log.Printf("Connection test %d failed: %v\n", i+1, err)
		} else {
			fmt.Printf("✓ Connection test %d successful\n", i+1)
		}
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("All database connection tests passed!")
}
