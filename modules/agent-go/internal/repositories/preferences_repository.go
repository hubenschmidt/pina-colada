package repositories

import (
	"github.com/pina-colada-co/agent-go/internal/models"
	"gorm.io/gorm"
)

type PreferencesRepository struct {
	db *gorm.DB
}

func NewPreferencesRepository(db *gorm.DB) *PreferencesRepository {
	return &PreferencesRepository{db: db}
}

func (r *PreferencesRepository) GetUserPreferences(userID int64) (*models.UserPreferences, error) {
	var prefs models.UserPreferences
	err := r.db.Where("user_id = ?", userID).First(&prefs).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &prefs, err
}

func (r *PreferencesRepository) CreateUserPreferences(prefs *models.UserPreferences) error {
	return r.db.Create(prefs).Error
}

func (r *PreferencesRepository) UpdateUserPreferences(userID int64, updates map[string]interface{}) error {
	return r.db.Model(&models.UserPreferences{}).Where("user_id = ?", userID).Updates(updates).Error
}

func (r *PreferencesRepository) GetTenantPreferences(tenantID int64) (*models.TenantPreferences, error) {
	var prefs models.TenantPreferences
	err := r.db.Where("tenant_id = ?", tenantID).First(&prefs).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &prefs, err
}

func (r *PreferencesRepository) GetUserWithTenant(userID int64) (*models.User, *models.Tenant, error) {
	var user models.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return nil, nil, err
	}

	var tenant *models.Tenant
	if user.TenantID != nil {
		var t models.Tenant
		if err := r.db.First(&t, *user.TenantID).Error; err == nil {
			tenant = &t
		}
	}

	return &user, tenant, nil
}
