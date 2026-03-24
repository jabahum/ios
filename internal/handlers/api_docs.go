package handlers

// Authentication API

// HandlerLoginSubmit godoc
// @Summary User login
// @Description Authenticate user and create session
// @Tags Authentication
// @Accept  json
// @Produce  json
// @Param   credentials  body  object{username=string,password=string}  true  "User credentials"
// @Success 200 {object} map[string]interface{} "Login successful"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /login [post]

// HandlerLoginOut godoc
// @Summary User logout
// @Description End user session
// @Tags Authentication
// @Success 302 "Redirect to login"
// @Router /logout [get]

// Locations API

// HandlerGetDistricts godoc
// @Summary Get all districts
// @Description Retrieve list of all districts
// @Tags Locations
// @Produce  json
// @Success 200 {array} map[string]interface{} "List of districts"
// @Failure 500 {object} map[string]string "Server error"
// @Router /api/locations/districts [get]

// HandlerGetSubcountiesByDistrict godoc
// @Summary Get subcounties by district
// @Description Retrieve subcounties for a specific district
// @Tags Locations
// @Produce  json
// @Param   district_id  path  int  true  "District ID"
// @Success 200 {array} map[string]interface{} "List of subcounties"
// @Failure 500 {object} map[string]string "Server error"
// @Router /api/locations/subcounties/{district_id} [get]

// HandlerGetParishesBySubcounty godoc
// @Summary Get parishes by subcounty
// @Description Retrieve parishes for a specific subcounty
// @Tags Locations
// @Produce  json
// @Param   subcounty_id  path  int  true  "Subcounty ID"
// @Success 200 {array} map[string]interface{} "List of parishes"
// @Failure 500 {object} map[string]string "Server error"
// @Router /api/locations/parishes/{subcounty_id} [get]

// HandlerGetVillagesByParish godoc
// @Summary Get villages by parish
// @Description Retrieve villages for a specific parish
// @Tags Locations
// @Produce  json
// @Param   parish_id  path  int  true  "Parish ID"
// @Success 200 {array} map[string]interface{} "List of villages"
// @Failure 500 {object} map[string]string "Server error"
// @Router /api/locations/villages/{parish_id} [get]

// Outbreaks API

// HandlerGetOutbreaksAPI godoc
// @Summary Get all outbreaks
// @Description Retrieve list of all outbreaks
// @Tags Outbreaks
// @Produce  json
// @Security SessionAuth
// @Success 200 {array} map[string]interface{} "List of outbreaks"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /api/outbreaks [get]

// HandlerOutbreakListAPI godoc
// @Summary Get outbreak list (paginated)
// @Description Retrieve paginated list of outbreaks
// @Tags Outbreaks
// @Produce  json
// @Security SessionAuth
// @Param   page  query  int  false  "Page number"
// @Param   limit  query  int  false  "Items per page"
// @Success 200 {object} map[string]interface{} "Paginated outbreak list"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /api/outbreaks [get]

// Facilities API

// HandlerGetFacilities godoc
// @Summary Get all facilities
// @Description Retrieve list of all health facilities
// @Tags Facilities
// @Produce  json
// @Param   district_id  query  int  false  "Filter by district"
// @Success 200 {array} map[string]interface{} "List of facilities"
// @Failure 500 {object} map[string]string "Server error"
// @Router /api/facilities [get]

// VHF API

// HandlerVHFListAPI godoc
// @Summary Get VHF cases
// @Description Retrieve list of VHF cases
// @Tags VHF
// @Produce  json
// @Security SessionAuth
// @Param   outbreak_id  query  int  false  "Filter by outbreak"
// @Param   status  query  string  false  "Filter by status"
// @Success 200 {array} map[string]interface{} "List of VHF cases"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /api/vhf/patients [get]

// HandlerVHFViewAPI godoc
// @Summary Get VHF case by ID
// @Description Retrieve detailed information about a specific VHF case
// @Tags VHF
// @Produce  json
// @Security SessionAuth
// @Param   id  path  int  true  "Case ID"
// @Success 200 {object} map[string]interface{} "VHF case details"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 500 {object} map[string]string "Server error"
// @Router /api/vhf/patients/{id} [get]

// HandlerVHFPatientSubmitAPI godoc
// @Summary Create VHF case
// @Description Create a new VHF case/patient record
// @Tags VHF
// @Accept  json
// @Produce  json
// @Security SessionAuth
// @Param   patient  body  object  true  "VHF patient data"
// @Success 201 {object} map[string]interface{} "Case created"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /api/vhf/patients [post]

// Users API

// HandlerGetUsers godoc
// @Summary Get all users
// @Description Retrieve list of all users
// @Tags Users
// @Produce  json
// @Security SessionAuth
// @Success 200 {array} map[string]interface{} "List of users"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /api/users [get]

// HandlerUserSubmitAPI godoc
// @Summary Create user
// @Description Create a new user account
// @Tags Users
// @Accept  json
// @Produce  json
// @Security SessionAuth
// @Param   user  body  object  true  "User data"
// @Success 201 {object} map[string]interface{} "User created"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /api/users [post]

// Employees API

// HandlerEmployeeListAPI godoc
// @Summary Get employees
// @Description Retrieve list of employees
// @Tags Employees
// @Produce  json
// @Security SessionAuth
// @Success 200 {array} map[string]interface{} "List of employees"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /api/employees [get]

// HandlerGetEmployeeAPI godoc
// @Summary Get employee by ID
// @Description Retrieve employee details
// @Tags Employees
// @Produce  json
// @Security SessionAuth
// @Param   id  path  int  true  "Employee ID"
// @Success 200 {object} map[string]interface{} "Employee details"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 500 {object} map[string]string "Server error"
// @Router /api/employees/{id} [get]

// Inventory API

// HandlerInventoryAPIItems godoc
// @Summary Get inventory items
// @Description Retrieve list of inventory items
// @Tags Inventory
// @Produce  json
// @Security SessionAuth
// @Success 200 {array} map[string]interface{} "List of inventory items"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /api/inventory/items [get]

// HandlerInventoryAPIStockLevels godoc
// @Summary Get stock levels
// @Description Retrieve current stock levels for all items
// @Tags Inventory
// @Produce  json
// @Security SessionAuth
// @Success 200 {array} map[string]interface{} "Stock levels"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /api/inventory/stock-levels [get]

// HandlerInventoryAPILowStock godoc
// @Summary Get low stock items
// @Description Retrieve items with low stock levels
// @Tags Inventory
// @Produce  json
// @Security SessionAuth
// @Success 200 {array} map[string]interface{} "Low stock items"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /api/inventory/low-stock [get]

// Resource Management API

// HandlerPillarsAPI godoc
// @Summary Get pillars
// @Description Retrieve list of all pillars as `{ "pillars": [...] }`
// @Tags Resource Management
// @Produce  json
// @Security SessionAuth
// @Success 200 {object} map[string]interface{} "pillars array"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /api/pillars [get]

// Reports API

// GetQuickStats godoc
// @Summary Get quick statistics
// @Description Retrieve dashboard quick statistics
// @Tags Reports
// @Produce  json
// @Security SessionAuth
// @Success 200 {object} map[string]interface{} "Statistics"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /reports/quick-stats [get]

// GetChartData godoc
// @Summary Get chart data
// @Description Retrieve data for charts
// @Tags Reports
// @Produce  json
// @Security SessionAuth
// @Param   type  path  string  true  "Chart type"
// @Success 200 {object} map[string]interface{} "Chart data"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /reports/chart-data/{type} [get]

// RBAC API

// HandlerGetRoles godoc
// @Summary Get all roles
// @Description Retrieve list of all roles
// @Tags RBAC
// @Produce  json
// @Security SessionAuth
// @Success 200 {array} map[string]interface{} "List of roles"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /api/rbac/roles [get]

// HandlerGetPermissions godoc
// @Summary Get all permissions
// @Description Retrieve list of all permissions
// @Tags RBAC
// @Produce  json
// @Security SessionAuth
// @Success 200 {array} map[string]interface{} "List of permissions"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /api/rbac/permissions [get]

// Alerts API

// HandlerAlertsAPI godoc
// @Summary Get alerts
// @Description Retrieve list of alerts
// @Tags Alerts
// @Produce  json
// @Security SessionAuth
// @Param   page  query  int  false  "Page number"
// @Param   limit  query  int  false  "Items per page"
// @Success 200 {object} map[string]interface{} "Paginated alerts"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /api/alerts [get]
