package schemas

// UpdateUserPreferencesRequest represents the request body for updating user preferences
type UpdateUserPreferencesRequest struct {
	Theme    *string `json:"theme"`
	Timezone *string `json:"timezone"`
}

// UpdateTenantPreferencesRequest represents the request body for updating tenant preferences
type UpdateTenantPreferencesRequest struct {
	Theme string `json:"theme" validate:"required"`
}
