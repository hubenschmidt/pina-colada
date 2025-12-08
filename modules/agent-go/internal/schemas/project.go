package schemas

// ProjectCreate represents the request body for creating a project
type ProjectCreate struct {
	Name            string  `json:"name" validate:"required"`
	Description     *string `json:"description"`
	Status          *string `json:"status"`
	CurrentStatusID *int64  `json:"current_status_id"`
	StartDate       *string `json:"start_date"`
	EndDate         *string `json:"end_date"`
}

// ProjectUpdate represents the request body for updating a project
type ProjectUpdate struct {
	Name            *string `json:"name"`
	Description     *string `json:"description"`
	Status          *string `json:"status"`
	CurrentStatusID *int64  `json:"current_status_id"`
	StartDate       *string `json:"start_date"`
	EndDate         *string `json:"end_date"`
}
