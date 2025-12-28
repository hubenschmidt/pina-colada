package schemas

// TaskCreate represents the request body for creating a task
type TaskCreate struct {
	Title                  string   `json:"title" validate:"required"`
	Description            *string  `json:"description"`
	TaskableType           *string  `json:"taskable_type" validate:"omitempty,oneof=Deal Lead Job Project Organization Individual Contact"`
	TaskableID             *int64   `json:"taskable_id"`
	CurrentStatusID        *int64   `json:"current_status_id"`
	PriorityID             *int64   `json:"priority_id"`
	StartDate              *string  `json:"start_date"`
	DueDate                *string  `json:"due_date"`
	EstimatedHours         *float64 `json:"estimated_hours" validate:"omitempty,min=0"`
	ActualHours            *float64 `json:"actual_hours" validate:"omitempty,min=0"`
	Complexity             *int16   `json:"complexity" validate:"omitempty,oneof=1 2 3 5 8 13 21"`
	SortOrder              *int     `json:"sort_order"`
	AssignedToIndividualID *int64   `json:"assigned_to_individual_id"`
}

// TaskUpdate represents the request body for updating a task
type TaskUpdate struct {
	Title                  *string  `json:"title"`
	Description            *string  `json:"description"`
	TaskableType           *string  `json:"taskable_type" validate:"omitempty,oneof=Deal Lead Job Project Organization Individual Contact"`
	TaskableID             *int64   `json:"taskable_id"`
	CurrentStatusID        *int64   `json:"current_status_id"`
	PriorityID             *int64   `json:"priority_id"`
	StartDate              *string  `json:"start_date"`
	DueDate                *string  `json:"due_date"`
	EstimatedHours         *float64 `json:"estimated_hours" validate:"omitempty,min=0"`
	ActualHours            *float64 `json:"actual_hours" validate:"omitempty,min=0"`
	Complexity             *int16   `json:"complexity" validate:"omitempty,oneof=1 2 3 5 8 13 21"`
	SortOrder              *int     `json:"sort_order"`
	CompletedAt            *string  `json:"completed_at"`
	AssignedToIndividualID *int64   `json:"assigned_to_individual_id"`
}
