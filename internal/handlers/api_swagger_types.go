package handlers

// APIUserCreateRequest is the JSON body for POST /api/users (roles applied in the same request).
type APIUserCreateRequest struct {
	Username     string `json:"username" example:"jdoe"`
	Email        string `json:"email" example:"jdoe@example.com"`
	FirstName    string `json:"first_name" example:"Jane"`
	LastName     string `json:"last_name" example:"Doe"`
	Password     string `json:"password" example:"ChangeMe123!"`
	DepartmentID int    `json:"department_id" example:"1"`
	RoleIDs      []int  `json:"role_ids" example:"2,3"`
	IsActive     bool   `json:"is_active" example:"true"`
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
