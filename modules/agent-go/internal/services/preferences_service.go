package services

import (
	"errors"

	"github.com/pina-colada-co/agent-go/internal/models"
	"github.com/pina-colada-co/agent-go/internal/repositories"
)

var ErrPreferencesUserNotFound = errors.New("user not found")

type PreferencesService struct {
	prefsRepo *repositories.PreferencesRepository
}

func NewPreferencesService(prefsRepo *repositories.PreferencesRepository) *PreferencesService {
	return &PreferencesService{prefsRepo: prefsRepo}
}

type UserPreferencesResponse struct {
	Theme          *string `json:"theme"`
	Timezone       string  `json:"timezone"`
	EffectiveTheme string  `json:"effective_theme"`
	CanEditTenant  bool    `json:"can_edit_tenant"`
}

func (s *PreferencesService) GetUserPreferences(userID int64) (*UserPreferencesResponse, error) {
	user, tenant, err := s.prefsRepo.GetUserWithTenant(userID)
	if err != nil {
		return nil, ErrPreferencesUserNotFound
	}

	prefs, err := s.prefsRepo.GetUserPreferences(userID)
	if err != nil {
		return nil, err
	}

	// Create prefs if not exists
	if prefs == nil {
		prefs = &models.UserPreferences{UserID: userID}
		if err := s.prefsRepo.CreateUserPreferences(prefs); err != nil {
			return nil, err
		}
	}

	effectiveTheme := s.resolveTheme(prefs, tenant)
	timezone := "America/New_York"
	if prefs.Timezone != nil {
		timezone = *prefs.Timezone
	}

	_ = user // suppress unused warning for now
	return &UserPreferencesResponse{
		Theme:          prefs.Theme,
		Timezone:       timezone,
		EffectiveTheme: effectiveTheme,
		CanEditTenant:  false, // TODO: check user roles
	}, nil
}

func (s *PreferencesService) UpdateUserPreferences(userID int64, theme, timezone *string) (*UserPreferencesResponse, error) {
	prefs, err := s.prefsRepo.GetUserPreferences(userID)
	if err != nil {
		return nil, err
	}

	if prefs == nil {
		prefs = &models.UserPreferences{UserID: userID}
		if err := s.prefsRepo.CreateUserPreferences(prefs); err != nil {
			return nil, err
		}
	}

	updates := make(map[string]interface{})
	if theme != nil {
		updates["theme"] = *theme
	}
	if timezone != nil {
		updates["timezone"] = *timezone
	}

	if len(updates) > 0 {
		if err := s.prefsRepo.UpdateUserPreferences(userID, updates); err != nil {
			return nil, err
		}
	}

	return s.GetUserPreferences(userID)
}

func (s *PreferencesService) resolveTheme(prefs *models.UserPreferences, tenant *models.Tenant) string {
	if prefs != nil && prefs.Theme != nil {
		return *prefs.Theme
	}

	if tenant != nil {
		tenantPrefs, _ := s.prefsRepo.GetTenantPreferences(tenant.ID)
		if tenantPrefs != nil && tenantPrefs.Theme != "" {
			return tenantPrefs.Theme
		}
	}

	return "light"
}
