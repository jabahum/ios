@echo off
echo Setting up basic inventory and RBAC data...
echo Using corrected SQL script that works with your actual database schema...
echo.

REM Check if Go is installed
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo Error: Go is not installed or not in PATH
    echo Please install Go from https://golang.org/dl/
    pause
    exit /b 1
)

REM Check if PostgreSQL driver is available
go mod tidy

REM Run the seed data script
echo Running seed data script...
go run scripts/run_seed_data.go

if %errorlevel% equ 0 (
    echo.
    echo ✅ Seed data setup completed successfully!
    echo.
    echo You can now:
    echo - Create users and assign roles
    echo - Create inventory items and manage categories  
    echo - Set up permissions for different user types
    echo - All dropdown lists should now be populated
) else (
    echo.
    echo ❌ Seed data setup failed!
    echo Please check your database connection settings.
)

echo.
pause
