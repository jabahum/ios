package handlers

// APIUserCreateRequest is the JSON body for POST /api/users (roles applied in the same request).
// Department is not set here; link via employee or PUT /api/users/:id when needed.
type APIUserCreateRequest struct {
	Username  string `json:"username" example:"jdoe"`
	Email     string `json:"email" example:"jdoe@example.com"`
	FirstName string `json:"first_name" example:"Jane"`
	LastName  string `json:"last_name" example:"Doe"`
	Password  string `json:"password" example:"ChangeMe123!"`
	RoleIDs   []int  `json:"role_ids" example:"2,3"`
	IsActive  bool   `json:"is_active" example:"true"`
}

// APIEmployeeWriteRequest documents POST /api/employees and PUT /api/employees/{id}.
// JSON maps to models.Employee (omit employee_id on create; id comes from the path on update).
type APIEmployeeWriteRequest struct {
	EmployeeFname string `json:"employee_fname" example:"Jane"`
	EmployeeLname string `json:"employee_lname" example:"Doe"`
	EmployeeSex   string `json:"employee_sex" example:"F"`
	EmployeeEmail string `json:"employee_email" example:"jane@example.com"`
	EmployeePhone string `json:"employee_phone"`
	EmployeeCadre string `json:"employee_cadre"`
	Facility      int64  `json:"facility" example:"1"`
	AFIFacility   string `json:"afi_facility"`
	AFIRegion     string `json:"afi_region"`
	AFIDistrict   string `json:"afi_district"`
}

// APIDepartmentWriteRequest is the JSON body for POST /api/departments and PUT /api/departments/{id}.
type APIDepartmentWriteRequest struct {
	Name        string `json:"name" example:"Surveillance"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active" example:"true"`
}

// APIUserUpdateRequest is the JSON body for PUT /api/users/:id (optional role_ids replaces all roles when non-empty).
type APIUserUpdateRequest struct {
	Email        string `json:"email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	DepartmentID int    `json:"department_id"`
	IsActive     bool   `json:"is_active"`
	RoleIDs      []int  `json:"role_ids"`
}

// APIOutbreakCreateRequest documents POST /api/outbreaks.
type APIOutbreakCreateRequest struct {
	Name             string `json:"name" example:"Ebola response"`
	Description      string `json:"description"`
	StartDate        string `json:"start_date" example:"2025-01-15"`
	EndDate          string `json:"end_date"`
	Status           string `json:"status" example:"active"`
	OutbreakType     string `json:"outbreak_type" example:"vhf"`
	OutbreakCategory string `json:"outbreak_category"`
}

// APIOutbreakAssignRequest documents POST /api/outbreaks/assign.
type APIOutbreakAssignRequest struct {
	OutbreakID int64 `json:"outbreak_id" example:"1"`
	UserID     int64 `json:"user_id" example:"42"`
}

// --- Resource management JSON bodies (/api/resource-management/*) ---

// APIResourceManagementPillarWrite is the body for POST/PUT /api/resource-management/pillars.
type APIResourceManagementPillarWrite struct {
	Name            string `json:"name" example:"Surveillance"`
	Description     string `json:"description"`
	PillarHeadID    int64  `json:"pillar_head_id"`
	PillarHeadName  string `json:"pillar_head_name"`
	PillarHeadEmail string `json:"pillar_head_email"`
	PillarHeadPhone string `json:"pillar_head_phone"`
	IsActive        bool   `json:"is_active" example:"true"`
}

// APIResourceManagementRRTTeamWrite is the body for POST/PUT /api/resource-management/rrt-teams.
type APIResourceManagementRRTTeamWrite struct {
	TeamName        string   `json:"team_name" example:"RRT Central"`
	TeamCode        string   `json:"team_code" example:"RRT-C-01"`
	TeamType        string   `json:"team_type" example:"medical"`
	TeamLeadName    string   `json:"team_lead_name" example:"Dr. A"`
	TeamLeadPhone   string   `json:"team_lead_phone"`
	TeamLeadEmail   string   `json:"team_lead_email"`
	TeamSize        int      `json:"team_size" example:"5"`
	Specializations []string `json:"specializations"`
	BaseLocation    string   `json:"base_location"`
	IsActive        bool     `json:"is_active" example:"true"`
}

// APIResourceManagementRRTDeploymentWrite is the body for POST/PUT /api/resource-management/rrt-deployments.
type APIResourceManagementRRTDeploymentWrite struct {
	TeamID             int64  `json:"team_id" example:"1"`
	OutbreakID         int64  `json:"outbreak_id" example:"1"`
	DeploymentDate     string `json:"deployment_date" example:"2025-03-01"`
	ExpectedReturnDate string `json:"expected_return_date"`
	ActualReturnDate   string `json:"actual_return_date"`
	DeploymentStatus   string `json:"deployment_status" example:"deployed"`
	DeploymentPurpose  string `json:"deployment_purpose"`
	AssignedVehicle    string `json:"assigned_vehicle"`
	AssignedDriver     string `json:"assigned_driver"`
	DeploymentNotes    string `json:"deployment_notes"`
}

// APIResourceManagementResourceWrite is the body for POST/PUT /api/resource-management/resources.
type APIResourceManagementResourceWrite struct {
	Name          string `json:"name" example:"Surgical gloves"`
	Description   string `json:"description"`
	ResourceCode  string `json:"resource_code"`
	CategoryID    int64  `json:"category_id" example:"1"`
	UnitOfMeasure string `json:"unit_of_measure" example:"box"`
	IsConsumable  bool   `json:"is_consumable" example:"true"`
	HasExpiry     bool   `json:"has_expiry"`
	ShelfLifeDays int64  `json:"shelf_life_days"`
	IsCritical    bool   `json:"is_critical"`
	IsActive      bool   `json:"is_active" example:"true"`
}

// APIResourceManagementRequisitionWrite is the body for POST/PUT /api/resource-management/requisitions.
type APIResourceManagementRequisitionWrite struct {
	RequisitionNumber string `json:"requisition_number" example:"REQ-2025-001"`
	OutbreakID        int64  `json:"outbreak_id" example:"1"`
	DeploymentID      int64  `json:"deployment_id"`
	RequestedBy       int64  `json:"requested_by"`
	RequiredDate      string `json:"required_date" example:"2025-03-15"`
	Priority          string `json:"priority" example:"normal"`
	Status            string `json:"status" example:"pending"`
	Notes             string `json:"notes"`
}

// APIResourceManagementActivityLogWrite is the body for POST/PUT /api/resource-management/activity-logs.
type APIResourceManagementActivityLogWrite struct {
	DeploymentID        int64  `json:"deployment_id" example:"1"`
	ActivityType        string `json:"activity_type" example:"investigation"`
	ActivityDate        string `json:"activity_date" example:"2025-03-01"`
	StartTime           string `json:"start_time" example:"09:00"`
	EndTime             string `json:"end_time" example:"17:00"`
	Location            string `json:"location"`
	ParticipantsCount   int64  `json:"participants_count"`
	ActivityDescription string `json:"activity_description"`
	Outcomes            string `json:"outcomes"`
	Challenges          string `json:"challenges"`
	Recommendations     string `json:"recommendations"`
	ResourcesUsed       string `json:"resources_used"`
}

// APIResourceCategoryWrite is the body for POST /api/resource-management/resource-categories.
type APIResourceCategoryWrite struct {
	Name         string `json:"name" example:"PPE"`
	Description  string `json:"description"`
	CategoryType string `json:"category_type" example:"ppe"`
	IsActive     bool   `json:"is_active" example:"true"`
}
