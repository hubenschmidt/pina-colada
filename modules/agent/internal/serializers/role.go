package serializers

// RoleResponse represents a role in API responses
type RoleResponse struct {
	ID          int64   `json:"id"`
	TenantID    *int64  `json:"tenant_id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	IsSystem    bool    `json:"is_system"`
}

// PermissionResponse represents a permission in API responses
type PermissionResponse struct {
	ID          int64   `json:"id"`
	Resource    string  `json:"resource"`
	Action      string  `json:"action"`
	Description *string `json:"description,omitempty"`
}

// UserRoleResponse represents a user's role assignment
type UserRoleResponse struct {
	UserID    int64   `json:"user_id"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Email     string  `json:"email"`
	RoleID    *int64  `json:"role_id"`
	RoleName  *string `json:"role_name"`
}

// RolePermissionsResponse returns permission IDs for a role
type RolePermissionsResponse struct {
	RoleID        int64   `json:"role_id"`
	PermissionIDs []int64 `json:"permission_ids"`
}
