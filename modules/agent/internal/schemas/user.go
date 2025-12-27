package schemas

// SetSelectedProjectRequest represents the request body for setting a user's selected project
type SetSelectedProjectRequest struct {
	ProjectID *int64 `json:"project_id"`
}
