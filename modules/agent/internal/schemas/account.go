package schemas

// AccountRelationshipCreate represents the request body for creating an account relationship
type AccountRelationshipCreate struct {
	ToAccountID      int64   `json:"to_account_id" validate:"required"`
	RelationshipType *string `json:"relationship_type"`
	Notes            *string `json:"notes"`
}
