package services

import (
	"errors"
	"regexp"
	"strings"

	"github.com/pina-colada-co/agent-go/internal/repositories"
	"github.com/pina-colada-co/agent-go/internal/schemas"
)

var ErrUserNotFound = errors.New("user not found")

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

// GetOrCreateUser finds or creates a user by Auth0 sub - implements middleware.UserLoader
func (s *AuthService) GetOrCreateUser(auth0Sub, email string) (int64, *int64, error) {
	user, err := s.userRepo.GetOrCreate(auth0Sub, email)
	if err != nil {
		return 0, nil, err
	}
	return user.ID, user.TenantID, nil
}

// GetOrCreateUserFull finds or creates a user by Auth0 sub and returns full response
func (s *AuthService) GetOrCreateUserFull(auth0Sub, email string) (*UserResponse, error) {
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
	user, err := s.GetOrCreateUserFull(auth0Sub, email)
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

// GetUserTenantByEmail returns tenant info for a user by email
func (s *AuthService) GetUserTenantByEmail(email string) (*UserTenantResponse, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	tenants, err := s.GetUserTenants(user.ID)
	if err != nil {
		return nil, err
	}

	return &UserTenantResponse{
		UserID:   user.ID,
		Email:    user.Email,
		TenantID: user.TenantID,
		Tenants:  tenants,
	}, nil
}

// UserTenantResponse represents the /users/{email}/tenant response
type UserTenantResponse struct {
	UserID   int64        `json:"user_id"`
	Email    string       `json:"email"`
	TenantID *int64       `json:"tenant_id"`
	Tenants  []TenantInfo `json:"tenants"`
}

var ErrProjectNotFound = errors.New("project not found")
var ErrProjectNotInTenant = errors.New("project does not belong to tenant")

// SetSelectedProject sets the user's selected project
func (s *AuthService) SetSelectedProject(userID int64, tenantID int64, projectID *int64) (*int64, error) {
	if projectID != nil {
		belongs, err := s.userRepo.ProjectBelongsToTenant(*projectID, tenantID)
		if err != nil {
			return nil, err
		}
		if !belongs {
			return nil, ErrProjectNotInTenant
		}
	}

	if err := s.userRepo.SetSelectedProject(userID, projectID); err != nil {
		return nil, err
	}

	return projectID, nil
}

// TenantCreateResponse represents the response for tenant creation
type TenantCreateResponse struct {
	Tenant TenantDetail `json:"tenant"`
}

// TenantDetail contains tenant details in response
type TenantDetail struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Plan string `json:"plan"`
	Role string `json:"role"`
}

// CreateTenant creates a new tenant and assigns the user as owner
func (s *AuthService) CreateTenant(userID int64, input schemas.TenantCreate) (*TenantCreateResponse, error) {
	slug := s.generateSlug(input.Name, input.Slug)
	plan := input.Plan
	if plan == "" {
		plan = "free"
	}

	repoInput := repositories.TenantCreateInput{
		UserID: userID,
		Name:   input.Name,
		Slug:   slug,
		Plan:   plan,
	}

	result, err := s.userRepo.CreateTenantWithUser(repoInput)
	if err != nil {
		return nil, err
	}

	return &TenantCreateResponse{
		Tenant: TenantDetail{
			ID:   result.TenantID,
			Name: result.TenantName,
			Slug: result.TenantSlug,
			Plan: result.TenantPlan,
			Role: result.RoleName,
		},
	}, nil
}

func (s *AuthService) generateSlug(name string, providedSlug *string) string {
	if providedSlug != nil && *providedSlug != "" {
		return *providedSlug
	}
	// Convert to lowercase, replace non-alphanumeric with hyphens, trim
	slug := strings.ToLower(name)
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug = re.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}

// TenantUserResponse represents a user in the tenant for API responses
type TenantUserResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// GetTenantUsers returns all users belonging to a tenant
func (s *AuthService) GetTenantUsers(tenantID int64) ([]TenantUserResponse, error) {
	users, err := s.userRepo.GetTenantUsers(tenantID)
	if err != nil {
		return nil, err
	}

	result := make([]TenantUserResponse, len(users))
	for i, u := range users {
		name := strings.TrimSpace(u.FirstName + " " + u.LastName)
		if name == "" {
			name = u.Email
		}
		result[i] = TenantUserResponse{
			ID:   u.ID,
			Name: name,
		}
	}

	return result, nil
}
