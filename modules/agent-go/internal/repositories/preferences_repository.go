package repositories

import (
	"errors"

	"github.com/pina-colada-co/agent-go/internal/models"
	"gorm.io/gorm"
)

type PreferencesRepository struct {
	db *gorm.DB
}

func NewPreferencesRepository(db *gorm.DB) *PreferencesRepository {
	return &PreferencesRepository{db: db}
}

// UserPrefsDTO is the data returned for user preferences
type UserPrefsDTO struct {
	Theme    *string
	Timezone *string
}

// TenantPrefsDTO is the data returned for tenant preferences
type TenantPrefsDTO struct {
	Theme string
}

// UserTenantDTO contains user and tenant info
type UserTenantDTO struct {
	UserID   int64
	TenantID *int64
}

func (r *PreferencesRepository) GetUserPreferences(userID int64) (*UserPrefsDTO, error) {
	var prefs models.UserPreferences
	err := r.db.Where("user_id = ?", userID).First(&prefs).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &UserPrefsDTO{Theme: prefs.Theme, Timezone: prefs.Timezone}, nil
}

func (r *PreferencesRepository) FindOrCreateUserPreferences(userID int64) (*UserPrefsDTO, error) {
	var prefs models.UserPreferences
	err := r.db.Where("user_id = ?", userID).First(&prefs).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		prefs = models.UserPreferences{UserID: userID}
		if err := r.db.Create(&prefs).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return &UserPrefsDTO{Theme: prefs.Theme, Timezone: prefs.Timezone}, nil
}

func (r *PreferencesRepository) UpdateUserPreferences(userID int64, updates map[string]interface{}) error {
	return r.db.Model(&models.UserPreferences{}).Where("user_id = ?", userID).Updates(updates).Error
}

func (r *PreferencesRepository) GetTenantPreferences(tenantID int64) (*TenantPrefsDTO, error) {
	var prefs models.TenantPreferences
	err := r.db.Where("tenant_id = ?", tenantID).First(&prefs).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &TenantPrefsDTO{Theme: prefs.Theme}, nil
}

func (r *PreferencesRepository) GetUserWithTenant(userID int64) (*UserTenantDTO, error) {
	var user models.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &UserTenantDTO{UserID: user.ID, TenantID: user.TenantID}, nil
}

// FindOrCreateTenantPreferences finds or creates tenant preferences
func (r *PreferencesRepository) FindOrCreateTenantPreferences(tenantID int64) (*TenantPrefsDTO, error) {
	var prefs models.TenantPreferences
	err := r.db.Where("tenant_id = ?", tenantID).First(&prefs).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		prefs = models.TenantPreferences{TenantID: tenantID, Theme: "light"}
		if err := r.db.Create(&prefs).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return &TenantPrefsDTO{Theme: prefs.Theme}, nil
}

// UpdateTenantPreferences updates tenant preferences
func (r *PreferencesRepository) UpdateTenantPreferences(tenantID int64, updates map[string]interface{}) error {
	return r.db.Model(&models.TenantPreferences{}).Where("tenant_id = ?", tenantID).Updates(updates).Error
}
