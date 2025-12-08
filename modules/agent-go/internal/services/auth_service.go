package services

import (
	"github.com/pina-colada-co/agent-go/internal/repositories"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo *repositories.UserRepository
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo *repositories.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

// UserResponse represents user info for API responses
type UserResponse struct {
	ID       int64   `json:"id"`
	Email    string  `json:"email"`
	Auth0Sub *string `json:"auth0_sub"`
	Status   string  `json:"status"`
	TenantID *int64  `json:"tenant_id"`
}

// TenantInfo represents tenant info for API responses
type TenantInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Role string `json:"role"`
}

// GetOrCreateUser finds or creates a user by Auth0 sub
func (s *AuthService) GetOrCreateUser(auth0Sub, email string) (*UserResponse, error) {
	user, err := s.userRepo.GetOrCreate(auth0Sub, email)
	if err != nil {
		return nil, err
	}

	return &UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		Auth0Sub: user.Auth0Sub,
		Status:   user.Status,
		TenantID: user.TenantID,
	}, nil
}

// GetUserTenants returns all tenants a user belongs to
func (s *AuthService) GetUserTenants(userID int64) ([]TenantInfo, error) {
	results, err := s.userRepo.GetUserTenantsWithRoles(userID)
	if err != nil {
		return nil, err
	}

	tenants := make([]TenantInfo, len(results))
	for i, r := range results {
		tenants[i] = TenantInfo{
			ID:   r.TenantID,
			Name: r.TenantName,
			Slug: r.TenantSlug,
			Role: r.RoleName,
		}
	}

	return tenants, nil
}

// MeResponse represents the /auth/me response
type MeResponse struct {
	User    UserResponse `json:"user"`
	Tenants []TenantInfo `json:"tenants"`
}

// GetMe returns current user with tenants
func (s *AuthService) GetMe(auth0Sub, email string) (*MeResponse, error) {
	user, err := s.GetOrCreateUser(auth0Sub, email)
	if err != nil {
		return nil, err
	}

	tenants, err := s.GetUserTenants(user.ID)
	if err != nil {
		return nil, err
	}

	return &MeResponse{
		User:    *user,
		Tenants: tenants,
	}, nil
}
