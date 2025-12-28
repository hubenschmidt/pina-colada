package schemas

// IndustryCreate represents the request body for creating an industry
type IndustryCreate struct {
	Name string `json:"name" validate:"required"`
}
