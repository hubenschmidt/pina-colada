package repositories

import (
	"github.com/pina-colada-co/agent-go/internal/models"
	"gorm.io/gorm"
)

// RoleDTO represents role data for the service layer
type RoleDTO struct {
	ID          int64
	TenantID    *int64
	Name        string
	Description *string
}

// RoleWithPermissionsDTO includes permissions
type RoleWithPermissionsDTO struct {
	RoleDTO
	Permissions []PermissionDTO
}

// FullPermissionDTO includes ID for assignment
type FullPermissionDTO struct {
	ID          int64
	Resource    string
	Action      string
	Description *string
}

// UserRoleDTO represents a user's role assignment
type UserRoleDTO struct {
	UserID    int64
	FirstName *string
	LastName  *string
	Email     string
	RoleID    *int64
	RoleName  *string
}

// RoleRepository handles role data access
type RoleRepository struct {
	db *gorm.DB
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// FindAll returns all roles for a tenant (including system roles where tenant_id is null)
func (r *RoleRepository) FindAll(tenantID int64) ([]RoleDTO, error) {
	var roles []models.Role
	err := r.db.Where("tenant_id = ? OR tenant_id IS NULL", tenantID).
		Order("name").
		Find(&roles).Error
	if err != nil {
		return nil, err
	}

	result := make([]RoleDTO, len(roles))
	for i, role := range roles {
		result[i] = roleToDTO(&role)
	}
	return result, nil
}

// FindByID returns a role by ID
func (r *RoleRepository) FindByID(id int64) (*RoleDTO, error) {
	var role models.Role
	err := r.db.First(&role, id).Error
	if err != nil {
		return nil, err
	}
	dto := roleToDTO(&role)
	return &dto, nil
}

// Create creates a new role
func (r *RoleRepository) Create(tenantID int64, name string, description *string) (*RoleDTO, error) {
	role := models.Role{
		TenantID:    &tenantID,
		Name:        name,
		Description: description,
	}
	if err := r.db.Create(&role).Error; err != nil {
		return nil, err
	}
	dto := roleToDTO(&role)
	return &dto, nil
}

// Update updates an existing role
func (r *RoleRepository) Update(id int64, name string, description *string) (*RoleDTO, error) {
	var role models.Role
	if err := r.db.First(&role, id).Error; err != nil {
		return nil, err
	}

	role.Name = name
	role.Description = description
	if err := r.db.Save(&role).Error; err != nil {
		return nil, err
	}

	dto := roleToDTO(&role)
	return &dto, nil
}

// Delete deletes a role
func (r *RoleRepository) Delete(id int64) error {
	return r.db.Delete(&models.Role{}, id).Error
}

// GetAllPermissions returns all available permissions
func (r *RoleRepository) GetAllPermissions() ([]FullPermissionDTO, error) {
	var permissions []models.Permission
	err := r.db.Order("resource, action").Find(&permissions).Error
	if err != nil {
		return nil, err
	}

	result := make([]FullPermissionDTO, len(permissions))
	for i, p := range permissions {
		result[i] = FullPermissionDTO{
			ID:          p.ID,
			Resource:    p.Resource,
			Action:      p.Action,
			Description: p.Description,
		}
	}
	return result, nil
}

// GetRolePermissions returns permission IDs for a role
func (r *RoleRepository) GetRolePermissions(roleID int64) ([]int64, error) {
	var permissionIDs []int64
	err := r.db.Table(`"Role_Permission"`).
		Select("permission_id").
		Where("role_id = ?", roleID).
		Scan(&permissionIDs).Error
	return permissionIDs, err
}

// RolePermissionMapDTO represents a role's permission assignments
type RolePermissionMapDTO struct {
	RoleID        int64
	PermissionIDs []int64
}

// GetAllRolePermissions returns all role-permission mappings
func (r *RoleRepository) GetAllRolePermissions() (map[int64][]int64, error) {
	var mappings []struct {
		RoleID       int64 `gorm:"column:role_id"`
		PermissionID int64 `gorm:"column:permission_id"`
	}

	err := r.db.Table(`"Role_Permission"`).
		Select("role_id, permission_id").
		Scan(&mappings).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64][]int64)
	for _, m := range mappings {
		result[m.RoleID] = append(result[m.RoleID], m.PermissionID)
	}
	return result, nil
}

// SetRolePermissions replaces all permissions for a role
func (r *RoleRepository) SetRolePermissions(roleID int64, permissionIDs []int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete existing permissions
		if err := tx.Where("role_id = ?", roleID).Delete(&models.RolePermission{}).Error; err != nil {
			return err
		}

		// Insert new permissions
		for _, permID := range permissionIDs {
			rp := models.RolePermission{
				RoleID:       roleID,
				PermissionID: permID,
			}
			if err := tx.Create(&rp).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetTenantUserRoles returns all users in a tenant with their role assignments
func (r *RoleRepository) GetTenantUserRoles(tenantID int64) ([]UserRoleDTO, error) {
	var results []UserRoleDTO

	err := r.db.Table(`"User"`).
		Select(`"User".id as user_id, "User".first_name, "User".last_name, "User".email, "User_Role".role_id, "Role".name as role_name`).
		Joins(`LEFT JOIN "User_Role" ON "User_Role".user_id = "User".id`).
		Joins(`LEFT JOIN "Role" ON "Role".id = "User_Role".role_id`).
		Where(`"User".tenant_id = ?`, tenantID).
		Scan(&results).Error

	return results, err
}

// SetUserRole assigns a role to a user (replaces existing)
func (r *RoleRepository) SetUserRole(userID, roleID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete existing role assignment
		if err := tx.Where("user_id = ?", userID).Delete(&models.UserRole{}).Error; err != nil {
			return err
		}

		// Insert new role assignment
		ur := models.UserRole{
			UserID: userID,
			RoleID: roleID,
		}
		return tx.Create(&ur).Error
	})
}

// IsSystemRole checks if a role is a system role (tenant_id is null)
func (r *RoleRepository) IsSystemRole(roleID int64) (bool, error) {
	var role models.Role
	if err := r.db.First(&role, roleID).Error; err != nil {
		return false, err
	}
	return role.TenantID == nil, nil
}

func roleToDTO(role *models.Role) RoleDTO {
	return RoleDTO{
		ID:          role.ID,
		TenantID:    role.TenantID,
		Name:        role.Name,
		Description: role.Description,
	}
}
