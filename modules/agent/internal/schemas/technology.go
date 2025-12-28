package schemas

// TechnologyCreate represents the request body for creating a technology
type TechnologyCreate struct {
	Name     string  `json:"name" validate:"required"`
	Category string  `json:"category" validate:"required"`
	Vendor   *string `json:"vendor"`
}
