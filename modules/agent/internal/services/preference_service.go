package services

import (
	"errors"

	"agent/internal/repositories"
)

var ErrPreferencesUserNotFound = errors.New("user not found")

type PreferenceService struct {
	prefsRepo *repositories.PreferenceRepository
	userRepo  *repositories.UserRepository
}

func NewPreferenceService(prefsRepo *repositories.PreferenceRepository, userRepo *repositories.UserRepository) *PreferenceService {
	return &PreferenceService{prefsRepo: prefsRepo, userRepo: userRepo}
}

type UserPreferencesResponse struct {
	Theme          *string `json:"theme"`
	Timezone       string  `json:"timezone"`
	EffectiveTheme string  `json:"effective_theme"`
	CanEditTenant  bool    `json:"can_edit_tenant"`
}

func (s *PreferenceService) GetUserPreferences(userID int64) (*UserPreferencesResponse, error) {
	userTenant, err := s.prefsRepo.GetUserWithTenant(userID)
	if err != nil {
		return nil, ErrPreferencesUserNotFound
	}

	prefs, err := s.prefsRepo.FindOrCreateUserPreferences(userID)
	if err != nil {
		return nil, err
	}

	effectiveTheme := s.resolveTheme(prefs.Theme, userTenant.TenantID)
	timezone := "America/New_York"
	if prefs.Timezone != nil {
		timezone = *prefs.Timezone
	}

	isAdmin, _ := s.userRepo.HasRole(userID, "admin")
	isOwner, _ := s.userRepo.HasRole(userID, "owner")

	return &UserPreferencesResponse{
		Theme:          prefs.Theme,
		Timezone:       timezone,
		EffectiveTheme: effectiveTheme,
		CanEditTenant:  isAdmin || isOwner,
	}, nil
}

func (s *PreferenceService) UpdateUserPreferences(userID int64, theme, timezone *string) (*UserPreferencesResponse, error) {
	// Ensure prefs exist
	_, err := s.prefsRepo.FindOrCreateUserPreferences(userID)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	if theme != nil {
		updates["theme"] = *theme
	}
	if timezone != nil {
		updates["timezone"] = *timezone
	}

	var updateErr error
	if len(updates) > 0 {
		updateErr = s.prefsRepo.UpdateUserPreferences(userID, updates)
	}
	if updateErr != nil {
		return nil, updateErr
	}

	return s.GetUserPreferences(userID)
}

func (s *PreferenceService) resolveTheme(userTheme *string, tenantID *int64) string {
	if userTheme != nil {
		return *userTheme
	}

	if tenantID == nil {
		return "light"
	}

	tenantPrefs, _ := s.prefsRepo.GetTenantPreferences(*tenantID)
	if tenantPrefs != nil && tenantPrefs.Theme != "" {
		return tenantPrefs.Theme
	}

	return "light"
}

var ErrTenantNotSet = errors.New("tenant not set")

// TenantPreferencesResponse represents the tenant preferences response
type TenantPreferencesResponse struct {
	Theme string `json:"theme"`
}

// GetTenantPreferences returns tenant preferences
func (s *PreferenceService) GetTenantPreferences(tenantID int64) (*TenantPreferencesResponse, error) {
	prefs, err := s.prefsRepo.FindOrCreateTenantPreferences(tenantID)
	if err != nil {
		return nil, err
	}

	return &TenantPreferencesResponse{
		Theme: prefs.Theme,
	}, nil
}

// UpdateTenantPreferences updates tenant preferences
func (s *PreferenceService) UpdateTenantPreferences(tenantID int64, theme string) (*TenantPreferencesResponse, error) {
	// Ensure prefs exist
	_, err := s.prefsRepo.FindOrCreateTenantPreferences(tenantID)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{"theme": theme}
	if err := s.prefsRepo.UpdateTenantPreferences(tenantID, updates); err != nil {
		return nil, err
	}

	return s.GetTenantPreferences(tenantID)
}
