package repositories

import (
	"errors"

	apperrors "github.com/pina-colada-co/agent-go/internal/errors"
	"github.com/pina-colada-co/agent-go/internal/models"
	"gorm.io/gorm"
)

// UserRepository handles user data access
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByAuth0Sub finds a user by Auth0 subject ID
func (r *UserRepository) FindByAuth0Sub(auth0Sub string) (*models.User, error) {
	var user models.User
	err := r.db.Where("auth0_sub = ?", auth0Sub).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(id int64) (*models.User, error) {
	var user models.User
	err := r.db.
		Preload("Preferences").
		First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create creates a new user
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// Update updates a user
func (r *UserRepository) Update(user *models.User, updates map[string]interface{}) error {
	return r.db.Model(user).Updates(updates).Error
}

// GetOrCreate finds or creates a user by Auth0 sub
func (r *UserRepository) GetOrCreate(auth0Sub, email string) (*models.User, error) {
	// First try by auth0_sub
	user, err := r.FindByAuth0Sub(auth0Sub)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}
	if user != nil {
		return user, nil
	}

	// Try by email (for seeded users without auth0_sub)
	user, err = r.FindByEmail(email)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}
	if user != nil {
		return r.linkAuth0Sub(user, auth0Sub)
	}

	// Create new user
	newUser := &models.User{
		Auth0Sub: &auth0Sub,
		Email:    email,
		Status:   "active",
	}
	if err := r.Create(newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

// GetUserTenantsWithRoles returns tenants and roles for a user
func (r *UserRepository) GetUserTenantsWithRoles(userID int64) ([]TenantWithRole, error) {
	var results []TenantWithRole

	err := r.db.Table(`"User_Role"`).
		Select(`"Tenant".id as tenant_id, "Tenant".name as tenant_name, "Tenant".slug as tenant_slug, "Role".name as role_name`).
		Joins(`JOIN "Role" ON "Role".id = "User_Role".role_id`).
		Joins(`JOIN "Tenant" ON "Tenant".id = "Role".tenant_id`).
		Where(`"User_Role".user_id = ?`, userID).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// TenantWithRole represents a tenant with its role
type TenantWithRole struct {
	TenantID   int64  `json:"tenant_id"`
	TenantName string `json:"tenant_name"`
	TenantSlug string `json:"tenant_slug"`
	RoleName   string `json:"role_name"`
}

// SetSelectedProject updates user's selected project
func (r *UserRepository) SetSelectedProject(userID int64, projectID *int64) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("selected_project_id", projectID).Error
}

// HasRole checks if a user has a specific role (across all tenants)
func (r *UserRepository) HasRole(userID int64, roleName string) (bool, error) {
	var count int64
	err := r.db.Table(`"User_Role"`).
		Joins(`JOIN "Role" ON "Role".id = "User_Role".role_id`).
		Where(`"User_Role".user_id = ? AND LOWER("Role".name) = LOWER(?)`, userID, roleName).
		Count(&count).Error
	return count > 0, err
}

// GetUserGlobalRoles returns global role names for a user (roles with NULL tenant_id)
func (r *UserRepository) GetUserGlobalRoles(userID int64) ([]string, error) {
	var roles []string
	err := r.db.Table(`"Role"`).
		Select(`"Role".name`).
		Joins(`JOIN "User_Role" ON "User_Role".role_id = "Role".id`).
		Where(`"User_Role".user_id = ? AND "Role".tenant_id IS NULL`, userID).
		Pluck("name", &roles).Error
	return roles, err
}

// ProjectBelongsToTenant checks if a project belongs to a tenant
func (r *UserRepository) ProjectBelongsToTenant(projectID int64, tenantID int64) (bool, error) {
	var count int64
	err := r.db.Table(`"Project"`).Where("id = ? AND tenant_id = ?", projectID, tenantID).Count(&count).Error
	return count > 0, err
}

// TenantCreateInput contains data needed to create a tenant
type TenantCreateInput struct {
	UserID int64
	Name   string
	Slug   string
	Plan   string
}

// TenantCreateResult contains the result of creating a tenant
type TenantCreateResult struct {
	TenantID   int64
	TenantName string
	TenantSlug string
	TenantPlan string
	RoleName   string
}

// CreateTenantWithUser creates a new tenant and assigns the user as owner
func (r *UserRepository) CreateTenantWithUser(input TenantCreateInput) (*TenantCreateResult, error) {
	// Check if tenant with slug already exists
	var existing models.Tenant
	err := r.db.Where("slug = ?", input.Slug).First(&existing).Error
	if err == nil {
		// Tenant exists - check if user already has a role
		return r.assignUserToExistingTenant(input.UserID, &existing)
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Create new tenant with owner role
	return r.createNewTenantWithOwner(input)
}

func (r *UserRepository) assignUserToExistingTenant(userID int64, tenant *models.Tenant) (*TenantCreateResult, error) {
	// Check if user already has role in this tenant
	var existingRole models.Role
	err := r.db.Table(`"Role"`).
		Joins(`JOIN "User_Role" ON "User_Role".role_id = "Role".id`).
		Where(`"Role".tenant_id = ? AND "User_Role".user_id = ?`, tenant.ID, userID).
		First(&existingRole).Error

	if err == nil {
		// User already has role - return existing
		return &TenantCreateResult{
			TenantID:   tenant.ID,
			TenantName: tenant.Name,
			TenantSlug: tenant.Slug,
			TenantPlan: tenant.Plan,
			RoleName:   existingRole.Name,
		}, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Find owner role for this tenant
	var ownerRole models.Role
	err = r.db.Where("tenant_id = ? AND name = ?", tenant.ID, "Owner").First(&ownerRole).Error
	if err != nil {
		return nil, err
	}

	// Assign user as owner
	userRole := models.UserRole{UserID: userID, RoleID: ownerRole.ID}
	if err := r.db.Create(&userRole).Error; err != nil {
		return nil, err
	}

	return &TenantCreateResult{
		TenantID:   tenant.ID,
		TenantName: tenant.Name,
		TenantSlug: tenant.Slug,
		TenantPlan: tenant.Plan,
		RoleName:   ownerRole.Name,
	}, nil
}

func (r *UserRepository) createNewTenantWithOwner(input TenantCreateInput) (*TenantCreateResult, error) {
	var result *TenantCreateResult

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Create tenant
		tenant := &models.Tenant{
			Name: input.Name,
			Slug: input.Slug,
			Plan: input.Plan,
		}
		if err := tx.Create(tenant).Error; err != nil {
			return err
		}

		// Create Account for default organization
		account := &models.Account{
			Name:     input.Name,
			TenantID: &tenant.ID,
		}
		if err := tx.Create(account).Error; err != nil {
			return err
		}

		// Create default Organization
		org := &models.Organization{
			AccountID: &account.ID,
			Name:      input.Name,
		}
		if err := tx.Create(org).Error; err != nil {
			return err
		}

		// Create default roles for tenant
		roles := []models.Role{
			{TenantID: &tenant.ID, Name: "Owner"},
			{TenantID: &tenant.ID, Name: "Admin"},
			{TenantID: &tenant.ID, Name: "Member"},
		}
		for i := range roles {
			if err := tx.Create(&roles[i]).Error; err != nil {
				return err
			}
		}

		// Assign user as owner (first role)
		userRole := models.UserRole{UserID: input.UserID, RoleID: roles[0].ID}
		if err := tx.Create(&userRole).Error; err != nil {
			return err
		}

		result = &TenantCreateResult{
			TenantID:   tenant.ID,
			TenantName: tenant.Name,
			TenantSlug: tenant.Slug,
			TenantPlan: tenant.Plan,
			RoleName:   roles[0].Name,
		}
		return nil
	})

	return result, err
}

// TenantUser represents a user in a tenant
type TenantUser struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// GetTenantUsers returns all users belonging to a tenant
func (r *UserRepository) GetTenantUsers(tenantID int64) ([]TenantUser, error) {
	var users []TenantUser

	err := r.db.Table(`"User"`).
		Select(`DISTINCT "User".id, "User".email, "User".first_name, "User".last_name`).
		Joins(`JOIN "User_Role" ON "User_Role".user_id = "User".id`).
		Joins(`JOIN "Role" ON "Role".id = "User_Role".role_id`).
		Where(`"Role".tenant_id = ?`, tenantID).
		Order(`"User".first_name, "User".last_name`).
		Scan(&users).Error

	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) linkAuth0Sub(user *models.User, auth0Sub string) (*models.User, error) {
	if err := r.Update(user, map[string]interface{}{"auth0_sub": auth0Sub}); err != nil {
		return nil, err
	}
	return user, nil
}
