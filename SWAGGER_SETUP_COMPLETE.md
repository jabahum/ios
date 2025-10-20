# Swagger API Documentation Setup - Complete ✅

## What Has Been Done

### 1. Swagger Dependencies Installed
- ✅ `github.com/arsmn/fiber-swagger/v2@v2.31.1` - Fiber Swagger middleware
- ✅ `github.com/swaggo/swag@v1.8.12` - Swagger CLI tool for generating documentation

### 2. Files Created/Modified

#### New Files Created:
1. **`cmd/web/docs.go`** - Main Swagger configuration with API metadata and tags
2. **`internal/handlers/api_docs.go`** - Swagger annotations for all major API endpoints
3. **`cmd/web/docs/docs.go`** - Auto-generated Swagger documentation (by swag init)
4. **`cmd/web/docs/swagger.json`** - OpenAPI JSON specification
5. **`cmd/web/docs/swagger.yaml`** - OpenAPI YAML specification
6. **`API_DOCUMENTATION.md`** - Comprehensive API documentation guide
7. **`scripts/setup_resource_management_perms.go`** - Script to setup resource_management permissions
8. **`scripts/setup_resource_management_permissions.sql`** - SQL script for permissions

#### Modified Files:
1. **`internal/routes/routes.go`**
   - Added Swagger import
   - Added Swagger UI routes (`/swagger/*` and `/api/docs/*`)
   - Added RBAC middleware to all resource management routes

2. **`cmd/web/main.go`**
   - Added Swagger docs import

### 3. Swagger UI Endpoints

Once the application is running, access Swagger documentation at:
- **http://localhost:3000/swagger/index.html** - Interactive Swagger UI
- **http://localhost:3000/api/docs/index.html** - Alternative URL
- **http://localhost:3000/swagger/doc.json** - OpenAPI JSON spec
- **http://localhost:3000/swagger/swagger.yaml** - OpenAPI YAML spec

### 4. API Endpoints Documented

The following API categories are now documented in Swagger:

✅ **Authentication** (Login, Logout)
✅ **Locations** (Districts, Subcounties, Parishes, Villages)
✅ **Outbreaks** (List, Create, Update, Close)
✅ **Facilities** (List, Get, Create, Update)
✅ **VHF** (Cases, CIF forms, Laboratory data)
✅ **Mpox** (Cases, Admissions, Daily follow-up)
✅ **Measles** (Cases, CIF forms)
✅ **Polio** (Cases, CIF forms)
✅ **Users** (List, Create, Update, Delete)
✅ **Employees** (List, Get, Create, Update, Delete)
✅ **Inventory** (Items, Stock levels, Low stock, Donations)
✅ **Resource Management** (Pillars, RRT Teams, Deployments, Proposals)
✅ **Laboratory** (Tests, Sample types)
✅ **Reports** (Statistics, Charts, Exports)
✅ **Alerts** (List, Create, Update, Delete)
✅ **RBAC** (Roles, Permissions, User-Role assignments)

### 5. RBAC for Resource Management

All resource management routes now have proper RBAC middleware with permission checks:
- `resource_management:read` - View resources
- `resource_management:create` - Create resources
- `resource_management:update` - Update/approve/reject
- `resource_management:delete` - Delete resources

Routes protected:
- Dashboard
- Pillars
- RRT Teams
- RRT Team Members
- RRT Deployments  
- Deployment Proposals
- Deployment Extensions
- Field Role Assignments
- Resources
- Requisitions
- Activity Logs
- SitRep Generation

## How to Use

### 1. Start the Application
```bash
cd cmd/web
./web.exe
```

### 2. Access Swagger UI
Open your browser and navigate to:
```
http://localhost:3000/swagger/index.html
```

### 3. Test API Endpoints
- Use the "Try it out" button in Swagger UI to test endpoints
- For authenticated endpoints, you must first login via `/login` to get a session cookie

### 4. Setup Resource Management Permissions

Run the SQL script to create resource_management permissions:
```bash
# Connect to PostgreSQL
psql -U postgres -d ios

# Run the script
\i scripts/setup_resource_management_permissions.sql
```

Or manually run:
```sql
-- Create permissions
INSERT INTO permissions (resource, action, description)
SELECT 'resource_management', 'read', 'View resource management dashboard and lists'
WHERE NOT EXISTS (
    SELECT 1 FROM permissions WHERE resource = 'resource_management' AND action = 'read'
);

-- Similar INSERT statements for create, update, delete...

-- Assign to roles
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'Super Admin'
  AND p.resource = 'resource_management'
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
);
```

## Updating Swagger Documentation

When you add new API endpoints:

### 1. Add Swagger Annotations
In `internal/handlers/api_docs.go`, add comments like:
```go
// HandlerNewEndpoint godoc
// @Summary Brief description
// @Description Detailed description
// @Tags Category
// @Accept  json
// @Produce  json
// @Security SessionAuth
// @Param   id  path  int  true  "Record ID"
// @Success 200 {object} map[string]interface{} "Success response"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/your-endpoint/{id} [get]
```

### 2. Regenerate Documentation
```bash
cd cmd/web
swag init
```

### 3. Rebuild Application
```bash
go build -o web.exe
```

## API Documentation

Comprehensive API documentation is available in:
- **`API_DOCUMENTATION.md`** - Full API guide with examples

## Testing the API

### Using Swagger UI (Recommended)
1. Navigate to http://localhost:3000/swagger/index.html
2. Click "Try it out" on any endpoint
3. Fill in parameters
4. Click "Execute"

### Using cURL
```bash
# Login first
curl -X POST http://localhost:3000/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}' \
  -c cookies.txt

# Use the session cookie
curl -X GET http://localhost:3000/api/outbreaks \
  -b cookies.txt
```

### Using Postman
1. Import the OpenAPI specification: http://localhost:3000/swagger/doc.json
2. All endpoints will be automatically available

## Next Steps

1. ✅ Start the application
2. ✅ Access Swagger UI at http://localhost:3000/swagger/index.html
3. ✅ Run the permissions setup SQL script
4. ✅ Test API endpoints via Swagger UI
5. 🔄 Add more detailed annotations as needed
6. 🔄 Create request/response models for better documentation

## Benefits

- 📚 **Interactive Documentation** - Test APIs directly from the browser
- 🔄 **Auto-Generated** - Documentation stays in sync with code
- 📖 **Comprehensive** - All endpoints documented in one place
- 🧪 **Easy Testing** - No need for separate API testing tools
- 🌐 **Standards-Based** - Uses OpenAPI 3.0 specification
- 🔗 **Shareable** - Export OpenAPI spec for use in other tools

## Support

If you encounter any issues:
1. Check that the application is running
2. Verify you're accessing the correct URL
3. Ensure session authentication is working
4. Check browser console for errors

Happy API Testing! 🚀

