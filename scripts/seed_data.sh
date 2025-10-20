#!/bin/bash

echo "Setting up basic inventory and RBAC data..."
echo "Using corrected SQL script that works with your actual database schema..."
echo

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed or not in PATH"
    echo "Please install Go from https://golang.org/dl/"
    exit 1
fi

# Check if PostgreSQL driver is available
go mod tidy

# Run the seed data script
echo "Running seed data script..."
go run scripts/run_seed_data.go

if [ $? -eq 0 ]; then
    echo
    echo "✅ Seed data setup completed successfully!"
    echo
    echo "You can now:"
    echo "- Create users and assign roles"
    echo "- Create inventory items and manage categories"
    echo "- Set up permissions for different user types"
    echo "- All dropdown lists should now be populated"
else
    echo
    echo "❌ Seed data setup failed!"
    echo "Please check your database connection settings."
    exit 1
fi

echo
