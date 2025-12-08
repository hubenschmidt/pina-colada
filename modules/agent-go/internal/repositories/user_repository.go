package repositories

import (
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
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
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
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
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
	if err != nil {
		return nil, err
	}
	if user != nil {
		return user, nil
	}

	// Try by email (for seeded users without auth0_sub)
	user, err = r.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	if user != nil {
		// Update with auth0_sub
		if err := r.Update(user, map[string]interface{}{"auth0_sub": auth0Sub}); err != nil {
			return nil, err
		}
		return user, nil
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

func stringPtr(s string) *string {
	return &s
}
