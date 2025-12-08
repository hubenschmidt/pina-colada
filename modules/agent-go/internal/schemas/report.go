package schemas

// ReportFilter represents a filter condition for report queries
type ReportFilter struct {
	Field    string `json:"field" validate:"required"`
	Operator string `json:"operator" validate:"required,oneof=eq neq gt gte lt lte contains starts_with is_null is_not_null in"`
	Value    any    `json:"value"`
}

// Aggregation represents an aggregation function for report queries
type Aggregation struct {
	Function string `json:"function" validate:"required,oneof=count sum avg min max"`
	Field    string `json:"field" validate:"required"`
	Alias    string `json:"alias" validate:"required"`
}

// ReportQueryRequest represents the request body for querying reports
type ReportQueryRequest struct {
	PrimaryEntity string          `json:"primary_entity" validate:"required,oneof=organizations individuals contacts leads notes"`
	Columns       []string        `json:"columns" validate:"required,min=1"`
	Joins         []string        `json:"joins"`
	Filters       []ReportFilter  `json:"filters"`
	GroupBy       []string        `json:"group_by"`
	Aggregations  []Aggregation   `json:"aggregations"`
	Limit         int             `json:"limit" validate:"omitempty,min=1,max=1000"`
	Offset        int             `json:"offset" validate:"omitempty,min=0"`
	ProjectID     *int64          `json:"project_id"`
}

// SavedReportCreate represents the request body for creating a saved report
type SavedReportCreate struct {
	Name            string         `json:"name" validate:"required"`
	Description     *string        `json:"description"`
	QueryDefinition map[string]any `json:"query_definition" validate:"required"`
	ProjectIDs      []int64        `json:"project_ids"`
}

// SavedReportUpdate represents the request body for updating a saved report
type SavedReportUpdate struct {
	Name            *string        `json:"name"`
	Description     *string        `json:"description"`
	QueryDefinition map[string]any `json:"query_definition"`
	ProjectIDs      []int64        `json:"project_ids"`
}
