package schemas

// ContactCreate represents the request body for creating a contact
type ContactCreate struct {
	FirstName  string  `json:"first_name" validate:"required"`
	LastName   string  `json:"last_name" validate:"required"`
	Email      *string `json:"email" validate:"omitempty,email"`
	Phone      *string `json:"phone" validate:"omitempty,e164"`
	Title      *string `json:"title"`
	Department *string `json:"department"`
	Role       *string `json:"role"`
	Notes      *string `json:"notes"`
	AccountIDs []int64 `json:"account_ids"`
	IsPrimary  bool    `json:"is_primary"`
}

// ContactUpdate represents the request body for updating a contact
type ContactUpdate struct {
	FirstName  *string `json:"first_name"`
	LastName   *string `json:"last_name"`
	Email      *string `json:"email" validate:"omitempty,email"`
	Phone      *string `json:"phone" validate:"omitempty,e164"`
	Title      *string `json:"title"`
	Department *string `json:"department"`
	Role       *string `json:"role"`
	Notes      *string `json:"notes"`
	AccountIDs []int64 `json:"account_ids"`
	IsPrimary  *bool   `json:"is_primary"`
}
