package serializers

import (
	"time"

	"gorm.io/datatypes"
)

// ProposalResponse represents an agent proposal
type ProposalResponse struct {
	ID               int64          `json:"id,omitempty"`
	TenantID         int64          `json:"tenant_id,omitempty"`
	ProposedByID     int64          `json:"proposed_by_id,omitempty"`
	EntityType       string         `json:"entity_type"`
	EntityID         *int64         `json:"entity_id,omitempty"`
	Operation        string         `json:"operation"`
	Payload          datatypes.JSON `json:"payload,omitempty"`
	Status           string         `json:"status"`
	ValidationErrors datatypes.JSON `json:"validation_errors,omitempty"`
	Message          string         `json:"message,omitempty"`
	ReviewedByID     *int64         `json:"reviewed_by_id,omitempty"`
	ReviewedAt       *time.Time     `json:"reviewed_at,omitempty"`
	ExecutedAt       *time.Time     `json:"executed_at,omitempty"`
	ErrorMessage     *string        `json:"error_message,omitempty"`
	CreatedAt        time.Time      `json:"created_at,omitempty"`
	UpdatedAt        time.Time      `json:"updated_at,omitempty"`
}

// BulkProposalResponse represents results of bulk operations
type BulkProposalResponse struct {
	Succeeded []ProposalResponse `json:"succeeded"`
	Failed    []BulkError        `json:"failed"`
}

// BulkError represents an error from bulk operations
type BulkError struct {
	ID    int64  `json:"id"`
	Error string `json:"error"`
}

// ApprovalConfigResponse represents entity approval config
type ApprovalConfigResponse struct {
	EntityType       string `json:"entity_type"`
	RequiresApproval bool   `json:"requires_approval"`
}
