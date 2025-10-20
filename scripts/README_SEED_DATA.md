# Seed Data Setup for Inventory and RBAC

This directory contains scripts to populate your database with essential data for inventory management and role-based access control (RBAC).

**IMPORTANT**: This script has been corrected to work with your actual database schema. It will check if tables exist before trying to insert data, so it's safe to run even if some tables don't exist yet.

## What This Script Does

The seed data script will create:

### Inventory Management
- **10 Inventory Categories** (PPE, Medical Supplies, Medications, etc.)
- **8 Inventory Suppliers** (Ministry of Health, WHO, UNICEF, etc.)
- **8 Treatment Sites** (Hospitals and Health Centers)
- **10 Sample Inventory Items** (N95 masks, gloves, test kits, etc.)
- **10 Inventory Settings** (stock thresholds, alerts, etc.)

### RBAC (Role-Based Access Control)
- **12 User Roles** (super_admin, admin, manager, data_entry, etc.)
- **50+ Permissions** (create, read, update, delete for different resources)
- **Role-Permission Assignments** (pre-configured role permissions)
- **10 Departments** (Administration, Surveillance, Laboratory, etc.)

## How to Run

### Option 1: Using the Batch Script (Windows)
```bash
scripts/seed_data.bat
```

### Option 2: Using the Shell Script (Linux/Mac)
```bash
scripts/seed_data.sh
```

### Option 3: Manual Execution
```bash
go run scripts/run_seed_data.go
```

## Database Configuration

The script will use these environment variables (with defaults):

- `DB_HOST` (default: localhost)
- `DB_PORT` (default: 5432)
- `DB_USER` (default: postgres)
- `DB_PASSWORD` (default: password)
- `DB_NAME` (default: ios)

You can set these before running the script:

```bash
# Windows
set DB_HOST=your_host
set DB_PASSWORD=your_password
scripts/seed_data.bat

# Linux/Mac
export DB_HOST=your_host
export DB_PASSWORD=your_password
scripts/seed_data.sh
```

## What You'll Get

After running the script, you'll have:

### For User Management:
- Pre-defined roles that you can assign to users
- Permissions that control what users can do
- Department structure for organizing users

### For Inventory Management:
- Categories for organizing inventory items
- Suppliers for procurement
- Treatment sites for distribution
- Sample items to get you started

### For Dropdown Lists:
- All dropdown lists in your application will be populated
- You can create users and assign roles
- You can create inventory items and assign categories
- You can set up permissions for different user types

## Verification

The script will show you a summary of what was created:

```
✅ Inventory Categories: 10 records
✅ Inventory Suppliers: 8 records
✅ Treatment Sites: 8 records
✅ Departments: 10 records
✅ Roles: 12 records
✅ Permissions: 50+ records
✅ Role-Permission Assignments: 200+ records
✅ Inventory Items: 10 records
✅ Inventory Settings: 10 records
```

## Troubleshooting

### Database Connection Issues
- Make sure PostgreSQL is running
- Check your database credentials
- Ensure the database exists
- Verify network connectivity

### Permission Issues
- Make sure your database user has CREATE/INSERT permissions
- Check if tables already exist (script handles conflicts gracefully)

### Go Module Issues
- Run `go mod tidy` to ensure dependencies are available
- Make sure you're in the project root directory

## Next Steps

After running the seed data:

1. **Create Users**: Go to your user management interface and create users
2. **Assign Roles**: Assign appropriate roles to users
3. **Create Inventory Items**: Add more inventory items as needed
4. **Set Permissions**: Adjust role permissions as required
5. **Test Dropdowns**: Verify that all dropdown lists are populated

## Customization

You can modify `scripts/seed_basic_data.sql` to:
- Add more categories, suppliers, or treatment sites
- Create additional roles and permissions
- Add more sample inventory items
- Adjust role-permission assignments

The script uses `ON CONFLICT DO NOTHING` so it's safe to run multiple times.
