package repositories

import (
	"github.com/pina-colada-co/agent-go/internal/models"
	"gorm.io/gorm"
)

type AccountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// AccountSearchResult represents an account search result
type AccountSearchResult struct {
	ID        int64
	AccountID int64
	Name      string
	Type      string
}

func (r *AccountRepository) Search(query string, tenantID *int64, limit int) ([]AccountSearchResult, error) {
	var accounts []models.Account
	q := r.db.Preload("Organizations").Preload("Individuals").
		Where("LOWER(name) LIKE LOWER(?)", "%"+query+"%")
	if tenantID != nil {
		q = q.Where("tenant_id = ?", *tenantID)
	}
	if err := q.Order("name").Limit(limit).Find(&accounts).Error; err != nil {
		return nil, err
	}

	results := make([]AccountSearchResult, len(accounts))
	for i, a := range accounts {
		accountType, entityID := getAccountTypeAndEntityID(a)
		results[i] = AccountSearchResult{
			ID:        entityID,
			AccountID: a.ID,
			Name:      a.Name,
			Type:      accountType,
		}
	}
	return results, nil
}

func getAccountTypeAndEntityID(account models.Account) (string, int64) {
	if len(account.Organizations) > 0 {
		return "organization", account.Organizations[0].ID
	}
	if len(account.Individuals) > 0 {
		return "individual", account.Individuals[0].ID
	}
	return "unknown", account.ID
}

// CreateRelationship creates a relationship between two accounts
func (r *AccountRepository) CreateRelationship(fromAccountID, toAccountID, userID int64, relationshipType, notes *string) (*models.AccountRelationship, error) {
	rel := &models.AccountRelationship{
		FromAccountID:    fromAccountID,
		ToAccountID:      toAccountID,
		RelationshipType: relationshipType,
		Notes:            notes,
		CreatedBy:        userID,
		UpdatedBy:        userID,
	}
	if err := r.db.Create(rel).Error; err != nil {
		return nil, err
	}
	return rel, nil
}

// DeleteRelationship deletes a relationship
func (r *AccountRepository) DeleteRelationship(fromAccountID, relationshipID int64) error {
	return r.db.Where("id = ? AND (from_account_id = ? OR to_account_id = ?)", relationshipID, fromAccountID, fromAccountID).
		Delete(&models.AccountRelationship{}).Error
}
