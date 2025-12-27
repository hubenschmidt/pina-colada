package schemas

// TenantCreate represents the request body for creating a tenant
type TenantCreate struct {
	Name string  `json:"name" validate:"required"`
	Slug *string `json:"slug"`
	Plan string  `json:"plan" validate:"omitempty,oneof=free starter pro enterprise"`
}
