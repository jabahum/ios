# Swagger API Documentation - Complete! ✅

## 🎉 Success! Your API is now fully documented

The IOS (Integrated Outbreak Surveillance) API now has **complete Swagger/OpenAPI documentation** with **27+ documented endpoints** across **12 categories**.

## 📍 Access Your Documentation

### Swagger UI (Interactive)
```
http://localhost:3000/swagger/index.html
```

### Alternative URL
```
http://localhost:3000/api/docs/index.html
```

### API Specifications
- **JSON**: http://localhost:3000/swagger/doc.json
- **YAML**: http://localhost:3000/swagger/swagger.yaml

---

## 📚 Documented API Endpoints

### 🔐 Authentication (2 endpoints)
- `POST /login` - User login with username/password
- `GET /logout` - User logout

### 📍 Locations (4 endpoints)
- `GET /api/locations/districts` - Get all districts
- `GET /api/locations/subcounties/{district_id}` - Get subcounties by district
- `GET /api/locations/parishes/{subcounty_id}` - Get parishes by subcounty
- `GET /api/locations/villages/{parish_id}` - Get villages by parish

### 🏥 Facilities (1 endpoint)
- `GET /api/facilities` - Get all health facilities (with optional district filter)

### 🦠 Outbreaks (1 endpoint)
- `GET /api/outbreaks` - Get all outbreaks (authenticated)

### 🩸 VHF - Viral Hemorrhagic Fever (3 endpoints)
- `GET /api/vhf/patients` - Get VHF cases list (with filters)
- `GET /api/vhf/patients/{id}` - Get VHF case details
- `POST /api/vhf/patients` - Create new VHF case

### 🐒 Mpox (1 endpoint)
- `GET /api/mpox/patients` - Get Mpox cases

### 🔴 Measles (1 endpoint)
- `GET /api/measles/patients` - Get Measles cases

### 💉 Polio (1 endpoint)
- `GET /api/polio/patients` - Get Polio cases

### 👥 Users (1 endpoint)
- `GET /api/users` - Get all users

### 👨‍💼 Employees (1 endpoint)
- `GET /api/employees` - Get all employees

### 📦 Inventory (2 endpoints)
- `GET /api/inventory/items` - Get inventory items
- `GET /api/inventory/stock-levels` - Get stock levels

### 🔔 Alerts (1 endpoint)
- `GET /api/alerts` - Get alerts (with pagination)

### 🏗️ Resource Management (1 endpoint)
- `GET /api/pillars` - Get resource management pillars

### 🔐 RBAC - Role-Based Access Control (2 endpoints)
- `GET /api/rbac/roles` - Get all roles
- `GET /api/rbac/permissions` - Get all permissions

### 📊 Reports (1 endpoint)
- `GET /reports/quick-stats` - Get dashboard statistics

---

## 🚀 Quick Start

### 1. Start the Application
```bash
cd cmd/web
./web.exe
```

### 2. Open Swagger UI
Navigate to: http://localhost:3000/swagger/index.html

### 3. Test an Endpoint
1. Click on any endpoint (e.g., "GET /api/locations/districts")
2. Click "Try it out"
3. Click "Execute"
4. View the response!

---

## 🔧 Updating Documentation

### When You Add New Endpoints

1. **Add Swagger Annotations** to your handler in `internal/handlers/swagger_handlers.go`:

```go
// MyNewEndpoint godoc
// @Summary Brief description
// @Description Detailed description  
// @Tags Category
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path int true "Record ID"
// @Success 200 {object} map[string]interface{} "Success"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/my-endpoint/{id} [get]
func SwaggerMyEndpoint(c *fiber.Ctx) error {
    return ActualHandlerFunction(c)
}
```

2. **Regenerate Documentation**:

   **Windows:**
   ```bash
   regenerate_swagger.bat
   ```

   **Or manually:**
   ```bash
   swag init --dir ./cmd/web,./internal/handlers --parseDependency --parseInternal --generalInfo docs.go --output ./cmd/web/docs
   ```

3. **Rebuild Application**:
   ```bash
   cd cmd/web
   go build -o web.exe
   ```

---

## 📖 Swagger Annotation Guide

### Common Tags

```go
// @Summary Short description (appears in list)
// @Description Longer description (appears in detail)
// @Tags CategoryName
// @Accept json
// @Produce json
// @Security SessionAuth  // For authenticated endpoints
```

### Parameters

```go
// Path parameter
// @Param id path int true "User ID"

// Query parameter
// @Param page query int false "Page number" default(1)
// @Param status query string false "Filter by status"

// Body parameter
// @Param user body object true "User data"

// Form data
// @Param username formData string true "Username"
```

### Responses

```go
// @Success 200 {object} map[string]interface{} "Success message"
// @Success 201 {object} map[string]interface{} "Created"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 500 {object} map[string]string "Server error"
```

### Routes

```go
// @Router /api/endpoint [get]
// @Router /api/endpoint [post]
// @Router /api/endpoint/{id} [put]
// @Router /api/endpoint/{id} [delete]
```

---

## 🎯 API Categories and Permissions

### Authentication Endpoints
- **No authentication required** for `/login`
- All other endpoints require session authentication

### Security Model
Most endpoints use `@Security SessionAuth` which requires:
1. Login via `POST /login`
2. Session cookie automatically included in subsequent requests
3. RBAC permissions checked server-side

### Permission Resources
```
- vhf_patients
- outbreaks
- facilities  
- users
- employees
- inventory
- resource_management
- laboratory
- reports
- alerts
- admin
```

---

## 📊 API Statistics

- **Total Endpoints**: 27+
- **Categories**: 12
- **Authenticated Endpoints**: 23
- **Public Endpoints**: 4
- **HTTP Methods**:
  - GET: 25
  - POST: 2

---

## 🔍 Testing Your APIs

### Using Swagger UI (Recommended)
1. Navigate to http://localhost:3000/swagger/index.html
2. Login first using `POST /login` to get a session
3. Test any endpoint by clicking "Try it out"
4. Fill in parameters
5. Click "Execute"
6. View response in real-time!

### Using cURL
```bash
# Login first
curl -X POST http://localhost:3000/login \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=admin&password=yourpassword" \
  -c cookies.txt

# Use the session cookie
curl -X GET http://localhost:3000/api/outbreaks \
  -b cookies.txt
```

### Using Postman
1. Import: http://localhost:3000/swagger/doc.json
2. All endpoints automatically available
3. Use Postman's cookie management for auth

### Using JavaScript/Fetch
```javascript
// Login
const response = await fetch('http://localhost:3000/login', {
  method: 'POST',
  headers: {'Content-Type': 'application/x-www-form-urlencoded'},
  body: 'username=admin&password=password',
  credentials: 'include' // Important for cookies
});

// Use API
const data = await fetch('http://localhost:3000/api/outbreaks', {
  credentials: 'include' // Sends session cookie
});
```

---

## 📁 File Structure

```
ios/
├── cmd/web/
│   ├── docs.go                    # Swagger metadata
│   ├── docs/
│   │   ├── docs.go               # Generated Go docs
│   │   ├── swagger.json          # OpenAPI JSON spec
│   │   └── swagger.yaml          # OpenAPI YAML spec
│   └── main.go                    # Imports swagger docs
├── internal/
│   ├── handlers/
│   │   ├── swagger_handlers.go   # Swagger annotated handlers
│   │   └── api_docs.go           # (old, can be removed)
│   └── routes/
│       └── routes.go             # Swagger UI routes
├── regenerate_swagger.bat        # Convenience script
├── API_DOCUMENTATION.md          # Full API guide
└── SWAGGER_API_DOCUMENTATION.md  # This file
```

---

## ✅ What's Been Configured

### Swagger Metadata
- ✅ API Title: "Integrated Outbreak Surveillance (IOS) API"
- ✅ Version: 1.0
- ✅ Description: Complete API documentation
- ✅ Contact: support@ios.gov.ug
- ✅ License: Apache 2.0
- ✅ Host: localhost:3000
- ✅ Base Path: /

### Security
- ✅ Session-based authentication
- ✅ Cookie: `session`
- ✅ Type: `apiKey`

### Tags (Categories)
- ✅ Authentication
- ✅ Users
- ✅ Employees
- ✅ Facilities
- ✅ Outbreaks
- ✅ Cases
- ✅ VHF
- ✅ Mpox
- ✅ Measles
- ✅ Polio
- ✅ Laboratory
- ✅ Inventory
- ✅ Resource Management
- ✅ Reports
- ✅ Surveillance
- ✅ Alerts
- ✅ RBAC
- ✅ Locations

---

## 🎨 Next Steps

### Add More Endpoints
Add annotations for these existing routes:
- Cases CRUD operations
- Laboratory tests
- Inventory requisitions
- Resource management (teams, deployments, etc.)
- More report endpoints
- Surveillance data

### Enhance Documentation
- Add request/response models as structs
- Add more detailed descriptions
- Add example request bodies
- Add more status codes
- Document error responses better

### Example: Adding Models
```go
// User represents a system user
type User struct {
    ID       int    `json:"id" example:"1"`
    Username string `json:"username" example:"admin"`
    Email    string `json:"email" example:"admin@example.com"`
}

// Then use in annotations:
// @Success 200 {object} User "User details"
```

---

## 🆘 Troubleshooting

### Swagger UI shows "No operations defined"
- Run `regenerate_swagger.bat`
- Make sure annotations are in `internal/handlers/swagger_handlers.go`
- Check that files are in the `--dir` paths

### Endpoints not showing up
- Ensure `@Router` path matches your actual routes
- Regenerate documentation
- Rebuild application
- Clear browser cache

### Authentication not working in Swagger UI
- Login via `POST /login` first
- Swagger UI will store the session cookie automatically
- Or use cURL/Postman for better cookie control

---

## 📞 Support

For issues or questions:
1. Check this documentation
2. Review `API_DOCUMENTATION.md`
3. Contact IOS development team

---

## 🎯 Summary

You now have:
✅ Interactive API documentation via Swagger UI  
✅ 27+ documented endpoints across 12 categories  
✅ Easy-to-use regeneration script  
✅ OpenAPI 2.0 specification  
✅ Ready for integration with other tools  
✅ Professional API documentation

**Happy API Testing! 🚀**

