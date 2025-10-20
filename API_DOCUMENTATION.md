# IOS API Documentation

## Overview

The Integrated Outbreak Surveillance (IOS) system provides a comprehensive RESTful API for managing disease surveillance, case management, laboratory data, inventory, and resource management.

## Accessing API Documentation

### Swagger UI

Once the application is running, you can access the interactive API documentation at:

- **Swagger UI**: `http://localhost:3000/swagger/index.html`
- **Alternative URL**: `http://localhost:3000/api/docs/index.html`

### API Specifications

- **OpenAPI JSON**: `http://localhost:3000/swagger/doc.json`
- **OpenAPI YAML**: `http://localhost:3000/swagger/swagger.yaml`

## Authentication

Most API endpoints require authentication via session cookies. To authenticate:

1. **POST** to `/login` with credentials:
```json
{
  "username": "your_username",
  "password": "your_password"
}
```

2. The server will create a session cookie that must be included in subsequent requests.

## API Categories

### 1. Authentication
- `POST /login` - User login
- `GET /logout` - User logout
- `POST /api/auth/change-password` - Change password

### 2. Locations
- `GET /api/locations/districts` - Get all districts
- `GET /api/locations/subcounties/{district_id}` - Get subcounties by district
- `GET /api/locations/parishes/{subcounty_id}` - Get parishes by subcounty
- `GET /api/locations/villages/{parish_id}` - Get villages by parish

### 3. Outbreaks
- `GET /api/outbreaks` - List all outbreaks
- `GET /api/outbreaks/{id}` - Get outbreak by ID
- `POST /api/outbreaks` - Create new outbreak
- `PUT /api/outbreaks/{id}` - Update outbreak
- `DELETE /api/outbreaks/{id}` - Delete outbreak
- `POST /api/outbreaks/{id}/close` - Close outbreak

### 4. Facilities
- `GET /api/facilities` - List all facilities
- `GET /api/facilities/{id}` - Get facility by ID
- `POST /api/facilities` - Create facility
- `PUT /api/facilities/{id}` - Update facility

### 5. VHF (Viral Hemorrhagic Fever)
- `GET /api/vhf/patients` - List VHF cases
- `GET /api/vhf/patients/{id}` - Get VHF case details
- `POST /api/vhf/patients` - Create VHF case
- `PUT /api/vhf/patients/{id}` - Update VHF case
- `GET /api/vhf/cif/{id}` - Get VHF CIF by ID
- `GET /api/vhf/cif?case_code={code}` - Get VHF CIF by case code

### 6. Mpox
- `GET /api/mpox/patients` - List Mpox cases
- `GET /api/mpox/patients/{id}` - Get Mpox case details
- `POST /api/mpox/patients` - Create Mpox case
- `PUT /api/mpox/patients/{id}` - Update Mpox case

### 7. Measles
- `GET /api/measles/patients` - List Measles cases
- `GET /api/measles/patients/{id}` - Get Measles case details
- `POST /api/measles/patients` - Create Measles case

### 8. Polio
- `GET /api/polio/patients` - List Polio cases
- `GET /api/polio/patients/{id}` - Get Polio case details
- `POST /api/polio/patients` - Create Polio case

### 9. Users
- `GET /api/users` - List all users
- `GET /api/users/{id}` - Get user by ID
- `POST /api/users` - Create user
- `PUT /api/users/{id}` - Update user
- `DELETE /api/users/{id}` - Delete user

### 10. Employees
- `GET /api/employees` - List employees
- `GET /api/employees/{id}` - Get employee by ID
- `POST /api/employees` - Create employee
- `PUT /api/employees/{id}` - Update employee
- `DELETE /api/employees/{id}` - Delete employee

### 11. Inventory
- `GET /api/inventory/items` - List inventory items
- `GET /api/inventory/items/{id}` - Get item by ID
- `POST /api/inventory/items` - Create item
- `GET /api/inventory/stock-levels` - Get stock levels
- `GET /api/inventory/low-stock` - Get low stock items
- `POST /api/inventory/stock` - Add/update stock
- `GET /api/inventory/donations` - List donations
- `POST /api/inventory/donations` - Create donation

### 12. Resource Management
- `GET /api/pillars` - List pillars
- `GET /resource-management/rrt-teams` - List RRT teams
- `GET /resource-management/rrt-deployments` - List deployments
- `GET /resource-management/deployment-proposals` - List deployment proposals
- `POST /resource-management/deployment-proposals/save` - Create proposal
- `POST /resource-management/deployment-proposals/approve/{id}` - Approve proposal
- `POST /resource-management/deployment-proposals/reject/{id}` - Reject proposal

### 13. Laboratory
- `GET /api/laboratory` - List lab tests
- `GET /api/laboratory/{id}` - Get lab test by ID
- `POST /api/laboratory` - Create lab test
- `GET /api/lab/blood-types` - Get blood sample types
- `GET /api/lab/swab-types` - Get swab sample types
- `GET /api/lab/urine-types` - Get urine sample types

### 14. Reports
- `GET /reports/quick-stats` - Get dashboard statistics
- `GET /reports/chart-data/{type}` - Get chart data
- `GET /reports/table-data` - Get table data
- `POST /reports/export` - Export report
- `GET /reports/vhf-stats` - VHF statistics
- `GET /reports/demographics-stats` - Demographics statistics

### 15. Alerts
- `GET /api/alerts` - List alerts
- `GET /api/alerts/{id}` - Get alert by ID
- `POST /api/alerts` - Create alert
- `PUT /api/alerts/{id}` - Update alert
- `DELETE /api/alerts/{id}` - Delete alert

### 16. RBAC (Role-Based Access Control)
- `GET /api/rbac/roles` - List roles
- `GET /api/rbac/roles/{id}` - Get role by ID
- `POST /api/rbac/roles` - Create role
- `PUT /api/rbac/roles/{id}` - Update role
- `DELETE /api/rbac/roles/{id}` - Delete role
- `GET /api/rbac/permissions` - List permissions
- `POST /api/rbac/user-roles` - Assign role to user
- `DELETE /api/rbac/user-roles/{user_id}/{role_id}` - Remove role from user

## Permissions

All API endpoints are protected by Role-Based Access Control (RBAC). Users must have the appropriate permissions to access each endpoint.

### Permission Format
Permissions follow the format: `resource:action`

Examples:
- `vhf_patients:read` - View VHF patients
- `vhf_patients:create` - Create VHF patients
- `outbreaks:update` - Update outbreaks
- `resource_management:read` - View resource management

### Common Resources
- `vhf_patients` - VHF case management
- `outbreaks` - Outbreak management
- `facilities` - Facility management
- `users` - User management
- `employees` - Employee management
- `inventory` - Inventory management
- `resource_management` - Resource/RRT management
- `laboratory` - Laboratory management
- `reports` - Reporting and analytics
- `alerts` - Alert management
- `admin` - Administrative functions

### Common Actions
- `read` - View/list resources
- `create` - Create new resources
- `update` - Modify existing resources
- `delete` - Remove resources

## Response Formats

### Success Response
```json
{
  "status": "success",
  "data": { ... }
}
```

### Error Response
```json
{
  "error": "Error message",
  "details": "Additional error details"
}
```

### Paginated Response
```json
{
  "data": [ ... ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "pages": 5
  }
}
```

## Common Query Parameters

- `page` - Page number for pagination (default: 1)
- `limit` - Items per page (default: 20)
- `outbreak_id` - Filter by outbreak
- `status` - Filter by status
- `district_id` - Filter by district

## Example API Calls

### Get All Outbreaks
```bash
curl -X GET http://localhost:3000/api/outbreaks \
  -H "Cookie: fiber_sess=your_session_cookie"
```

### Create VHF Case
```bash
curl -X POST http://localhost:3000/api/vhf/patients \
  -H "Content-Type: application/json" \
  -H "Cookie: fiber_sess=your_session_cookie" \
  -d '{
    "case_code": "VHF-2025-001",
    "first_name": "John",
    "last_name": "Doe",
    "age": 35,
    "gender": "Male",
    "outbreak_id": 1
  }'
```

### Get Inventory Low Stock
```bash
curl -X GET http://localhost:3000/api/inventory/low-stock \
  -H "Cookie: fiber_sess=your_session_cookie"
```

## Updating Swagger Documentation

After making changes to API endpoints or adding new ones:

1. Add Swagger annotations to your handler functions in `internal/handlers/api_docs.go`
2. Regenerate documentation:
```bash
cd cmd/web
swag init
```
3. Rebuild the application:
```bash
go build -o web.exe
```

## API Versioning

The current API is version 1.0. Future versions will be accessible via `/api/v2/...` etc.

## Rate Limiting

Currently, there are no rate limits enforced. This may be added in future versions.

## Support

For API support or to report issues, contact the IOS development team.

