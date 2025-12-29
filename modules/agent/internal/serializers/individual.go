package serializers

import (
	"time"

	"agent/internal/models"
)

// IndividualListResponse represents an individual in list view
type IndividualListResponse struct {
	ID              int64     `json:"id"`
	FirstName       string    `json:"first_name"`
	LastName        string    `json:"last_name"`
	Email           *string   `json:"email"`
	Phone           *string   `json:"phone"`
	Title           *string   `json:"title"`
	SeniorityLevel  *string   `json:"seniority_level"`
	Department      *string   `json:"department"`
	IsDecisionMaker *bool     `json:"is_decision_maker"`
	Industries      []string  `json:"industries"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// RelationshipBrief represents a relationship for display
type RelationshipBrief struct {
	ID               int64   `json:"id"`
	AccountID        int64   `json:"account_id"`
	Name             string  `json:"name"`
	Type             string  `json:"type"`
	RelationshipID   int64   `json:"relationship_id"`
	RelationshipType *string `json:"relationship_type"`
}

// IndividualDetailResponse represents an individual in detail view
type IndividualDetailResponse struct {
	IndividualListResponse
	AccountID     *int64              `json:"account_id"`
	LinkedInURL   *string             `json:"linkedin_url"`
	TwitterURL    *string             `json:"twitter_url"`
	GithubURL     *string             `json:"github_url"`
	Bio           *string             `json:"bio"`
	Description   *string             `json:"description"`
	Relationships []RelationshipBrief `json:"relationships"`
	Contacts      []ContactBrief      `json:"contacts"`
}

// IndividualToListResponse converts Individual model to list response
func IndividualToListResponse(ind *models.Individual) IndividualListResponse {
	resp := IndividualListResponse{
		ID:              ind.ID,
		FirstName:       ind.FirstName,
		LastName:        ind.LastName,
		Email:           ind.Email,
		Phone:           ind.Phone,
		Title:           ind.Title,
		SeniorityLevel:  ind.SeniorityLevel,
		Department:      ind.Department,
		IsDecisionMaker: ind.IsDecisionMaker,
		CreatedAt:       ind.CreatedAt,
		UpdatedAt:       ind.UpdatedAt,
	}

	// Get industries from Account
	if ind.Account != nil {
		resp.Industries = make([]string, len(ind.Account.Industries))
		for i, industry := range ind.Account.Industries {
			resp.Industries[i] = industry.Name
		}
	}

	return resp
}

// IndividualToDetailResponse converts Individual model to detail response
func IndividualToDetailResponse(ind *models.Individual) IndividualDetailResponse {
	resp := IndividualDetailResponse{
		IndividualListResponse: IndividualToListResponse(ind),
		AccountID:              ind.AccountID,
		LinkedInURL:            ind.LinkedInURL,
		TwitterURL:             ind.TwitterURL,
		GithubURL:              ind.GithubURL,
		Bio:                    ind.Bio,
		Description:            ind.Description,
		Relationships:          []RelationshipBrief{},
		Contacts:               []ContactBrief{},
	}

	// Get relationships and contacts from Account
	if ind.Account != nil {
		resp.Relationships = getAccountRelationships(ind.Account, ind.Account.ID)
		resp.Contacts = getAccountContacts(ind.Account)
	}

	return resp
}

// getAccountContacts extracts contacts from an account
func getAccountContacts(account *models.Account) []ContactBrief {
	contacts := make([]ContactBrief, len(account.Contacts))
	for i, c := range account.Contacts {
		contacts[i] = ContactBrief{
			ID:        c.ID,
			FirstName: c.FirstName,
			LastName:  c.LastName,
			Email:     c.Email,
			Phone:     c.Phone,
			Title:     c.Title,
		}
	}
	return contacts
}

// getAccountRelationships extracts relationships from an account
func getAccountRelationships(account *models.Account, ownerAccountID int64) []RelationshipBrief {
	relationships := []RelationshipBrief{}
	seen := make(map[string]bool)

	// Outgoing relationships
	for _, rel := range account.OutgoingRelationships {
		addRelationshipIfNew(rel.ToAccount, rel.ID, rel.RelationshipType, seen, &relationships)
	}

	// Incoming relationships
	for _, rel := range account.IncomingRelationships {
		addRelationshipIfNew(rel.FromAccount, rel.ID, rel.RelationshipType, seen, &relationships)
	}

	return relationships
}

func addRelationshipIfNew(relAccount *models.Account, relID int64, relType *string, seen map[string]bool, relationships *[]RelationshipBrief) {
	if relAccount == nil {
		return
	}
	accountType, entityID := getAccountTypeAndEntityID(*relAccount)
	key := accountType + string(rune(entityID))
	if seen[key] {
		return
	}
	seen[key] = true
	*relationships = append(*relationships, RelationshipBrief{
		ID:               entityID,
		AccountID:        relAccount.ID,
		Name:             relAccount.Name,
		Type:             accountType,
		RelationshipID:   relID,
		RelationshipType: relType,
	})
}

// getAccountTypeAndEntityID determines the type and entity ID from an account
func getAccountTypeAndEntityID(account models.Account) (string, int64) {
	if len(account.Organizations) > 0 {
		return "organization", account.Organizations[0].ID
	}
	if len(account.Individuals) > 0 {
		return "individual", account.Individuals[0].ID
	}
	return "unknown", account.ID
}
