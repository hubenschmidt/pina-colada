package serializers

import (
	"time"

	"agent/internal/models"
)

// ContactAccountBrief represents a linked account for a contact
type ContactAccountBrief struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// ContactResponse represents a contact
type ContactResponse struct {
	ID         int64                 `json:"id"`
	FirstName  string                `json:"first_name"`
	LastName   string                `json:"last_name"`
	Email      *string               `json:"email"`
	Phone      *string               `json:"phone"`
	Title      *string               `json:"title"`
	Department *string               `json:"department"`
	Role       *string               `json:"role"`
	IsPrimary  bool                  `json:"is_primary"`
	Notes      *string               `json:"notes"`
	Accounts   []ContactAccountBrief `json:"accounts"`
	CreatedAt  time.Time             `json:"created_at"`
	UpdatedAt  time.Time             `json:"updated_at"`
}

// ContactToResponse converts Contact model to response
func ContactToResponse(contact *models.Contact) ContactResponse {
	accounts := make([]ContactAccountBrief, 0, len(contact.Accounts))
	for _, acc := range contact.Accounts {
		accountType := "unknown"
		if len(acc.Organizations) > 0 {
			accountType = "organization"
		}
		if accountType == "unknown" && len(acc.Individuals) > 0 {
			accountType = "individual"
		}
		accounts = append(accounts, ContactAccountBrief{
			ID:   acc.ID,
			Name: acc.Name,
			Type: accountType,
		})
	}

	return ContactResponse{
		ID:         contact.ID,
		FirstName:  contact.FirstName,
		LastName:   contact.LastName,
		Email:      contact.Email,
		Phone:      contact.Phone,
		Title:      contact.Title,
		Department: contact.Department,
		Role:       contact.Role,
		IsPrimary:  contact.IsPrimary,
		Notes:      contact.Notes,
		Accounts:   accounts,
		CreatedAt:  contact.CreatedAt,
		UpdatedAt:  contact.UpdatedAt,
	}
}
