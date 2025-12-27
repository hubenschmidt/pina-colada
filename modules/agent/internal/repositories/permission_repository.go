package repositories

import (
	"gorm.io/gorm"
)

// PermissionDTO represents permission data for the service layer
type PermissionDTO struct {
	Resource string
	Action   string
}

// PermissionRepository handles permission data access
type PermissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

// GetUserPermissions returns all permissions for a user via their roles
// Joins User_Role -> Role_Permission -> Permission
func (r *PermissionRepository) GetUserPermissions(userID int64) ([]PermissionDTO, error) {
	var permissions []PermissionDTO

	err := r.db.Table(`"Permission"`).
		Select(`"Permission".resource, "Permission".action`).
		Joins(`JOIN "Role_Permission" ON "Role_Permission".permission_id = "Permission".id`).
		Joins(`JOIN "User_Role" ON "User_Role".role_id = "Role_Permission".role_id`).
		Where(`"User_Role".user_id = ?`, userID).
		Scan(&permissions).Error

	if err != nil {
		return nil, err
	}

	return permissions, nil
}

// HasPermission checks if a user has a specific permission
func (r *PermissionRepository) HasPermission(userID int64, resource, action string) (bool, error) {
	var count int64

	err := r.db.Table(`"Permission"`).
		Joins(`JOIN "Role_Permission" ON "Role_Permission".permission_id = "Permission".id`).
		Joins(`JOIN "User_Role" ON "User_Role".role_id = "Role_Permission".role_id`).
		Where(`"User_Role".user_id = ? AND "Permission".resource = ? AND "Permission".action = ?`, userID, resource, action).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
