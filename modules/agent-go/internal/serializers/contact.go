package serializers

import (
	"time"

	"github.com/pina-colada-co/agent-go/internal/models"
)

// ContactResponse represents a contact
type ContactResponse struct {
	ID         int64     `json:"id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      *string   `json:"email"`
	Phone      *string   `json:"phone"`
	Title      *string   `json:"title"`
	Department *string   `json:"department"`
	Role       *string   `json:"role"`
	IsPrimary  bool      `json:"is_primary"`
	Notes      *string   `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ContactToResponse converts Contact model to response
func ContactToResponse(contact *models.Contact) ContactResponse {
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
		CreatedAt:  contact.CreatedAt,
		UpdatedAt:  contact.UpdatedAt,
	}
}
