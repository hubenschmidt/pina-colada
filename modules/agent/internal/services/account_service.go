package services

import (
	"agent/internal/repositories"
)

type AccountService struct {
	accountRepo *repositories.AccountRepository
}

func NewAccountService(accountRepo *repositories.AccountRepository) *AccountService {
	return &AccountService{accountRepo: accountRepo}
}

type AccountSearchResult struct {
	ID        int64  `json:"id"`
	AccountID int64  `json:"account_id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
}

func (s *AccountService) SearchAccounts(query string, tenantID *int64, limit int) ([]AccountSearchResult, error) {
	if limit <= 0 {
		limit = 20
	}
	repoResults, err := s.accountRepo.Search(query, tenantID, limit)
	if err != nil {
		return nil, err
	}

	results := make([]AccountSearchResult, len(repoResults))
	for i, r := range repoResults {
		results[i] = AccountSearchResult{
			ID:        r.ID,
			AccountID: r.AccountID,
			Name:      r.Name,
			Type:      r.Type,
		}
	}
	return results, nil
}

// RelationshipCreateInput holds data for creating a relationship
type RelationshipCreateInput struct {
	ToAccountID      int64   `json:"to_account_id"`
	RelationshipType *string `json:"relationship_type"`
	Notes            *string `json:"notes"`
}

// RelationshipResponse represents a created relationship
type RelationshipResponse struct {
	ID          int64 `json:"id"`
	ToAccountID int64 `json:"to_account_id"`
}

// CreateRelationship creates a relationship between accounts
func (s *AccountService) CreateRelationship(fromAccountID int64, input RelationshipCreateInput, userID int64) (*RelationshipResponse, error) {
	rel, err := s.accountRepo.CreateRelationship(fromAccountID, input.ToAccountID, userID, input.RelationshipType, input.Notes)
	if err != nil {
		return nil, err
	}
	return &RelationshipResponse{
		ID:          rel.ID,
		ToAccountID: rel.ToAccountID,
	}, nil
}

// DeleteRelationship deletes a relationship
func (s *AccountService) DeleteRelationship(fromAccountID, relationshipID int64) error {
	return s.accountRepo.DeleteRelationship(fromAccountID, relationshipID)
}
