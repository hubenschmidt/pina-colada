package services

import (
	"errors"

	"github.com/pina-colada-co/agent-go/internal/repositories"
	"github.com/pina-colada-co/agent-go/internal/serializers"
)

var (
	ErrRoleNotFound     = errors.New("role not found")
	ErrSystemRoleEdit   = errors.New("cannot modify system roles")
	ErrRoleNameRequired = errors.New("role name is required")
)

// RoleService handles role business logic
type RoleService struct {
	roleRepo *repositories.RoleRepository
}

// NewRoleService creates a new role service
func NewRoleService(roleRepo *repositories.RoleRepository) *RoleService {
	return &RoleService{roleRepo: roleRepo}
}

// GetRoles returns all roles for a tenant
func (s *RoleService) GetRoles(tenantID int64) ([]serializers.RoleResponse, error) {
	roles, err := s.roleRepo.FindAll(tenantID)
	if err != nil {
		return nil, err
	}

	result := make([]serializers.RoleResponse, len(roles))
	for i, r := range roles {
		result[i] = roleDTOToResponse(&r)
	}
	return result, nil
}

// GetRole returns a single role by ID
func (s *RoleService) GetRole(id int64) (*serializers.RoleResponse, error) {
	role, err := s.roleRepo.FindByID(id)
	if err != nil {
		return nil, ErrRoleNotFound
	}
	resp := roleDTOToResponse(role)
	return &resp, nil
}

// CreateRole creates a new role
func (s *RoleService) CreateRole(tenantID int64, name string, description *string) (*serializers.RoleResponse, error) {
	if name == "" {
		return nil, ErrRoleNameRequired
	}

	role, err := s.roleRepo.Create(tenantID, name, description)
	if err != nil {
		return nil, err
	}

	resp := roleDTOToResponse(role)
	return &resp, nil
}

// UpdateRole updates an existing role
func (s *RoleService) UpdateRole(id int64, name string, description *string) (*serializers.RoleResponse, error) {
	if name == "" {
		return nil, ErrRoleNameRequired
	}

	isSystem, err := s.roleRepo.IsSystemRole(id)
	if err != nil {
		return nil, ErrRoleNotFound
	}
	if isSystem {
		return nil, ErrSystemRoleEdit
	}

	role, err := s.roleRepo.Update(id, name, description)
	if err != nil {
		return nil, err
	}

	resp := roleDTOToResponse(role)
	return &resp, nil
}

// DeleteRole deletes a role
func (s *RoleService) DeleteRole(id int64) error {
	isSystem, err := s.roleRepo.IsSystemRole(id)
	if err != nil {
		return ErrRoleNotFound
	}
	if isSystem {
		return ErrSystemRoleEdit
	}

	return s.roleRepo.Delete(id)
}

// GetAllPermissions returns all available permissions
func (s *RoleService) GetAllPermissions() ([]serializers.PermissionResponse, error) {
	perms, err := s.roleRepo.GetAllPermissions()
	if err != nil {
		return nil, err
	}

	result := make([]serializers.PermissionResponse, len(perms))
	for i, p := range perms {
		result[i] = serializers.PermissionResponse{
			ID:          p.ID,
			Resource:    p.Resource,
			Action:      p.Action,
			Description: p.Description,
		}
	}
	return result, nil
}

// GetRolePermissions returns permission IDs for a role
func (s *RoleService) GetRolePermissions(roleID int64) ([]int64, error) {
	return s.roleRepo.GetRolePermissions(roleID)
}

// GetAllRolePermissions returns all role-permission mappings
func (s *RoleService) GetAllRolePermissions() (map[int64][]int64, error) {
	return s.roleRepo.GetAllRolePermissions()
}

// UpdateRolePermissions updates permissions for a role
func (s *RoleService) UpdateRolePermissions(roleID int64, permissionIDs []int64) error {
	isSystem, err := s.roleRepo.IsSystemRole(roleID)
	if err != nil {
		return ErrRoleNotFound
	}
	if isSystem {
		return ErrSystemRoleEdit
	}

	return s.roleRepo.SetRolePermissions(roleID, permissionIDs)
}

// GetTenantUserRoles returns all users with their role assignments
func (s *RoleService) GetTenantUserRoles(tenantID int64) ([]serializers.UserRoleResponse, error) {
	userRoles, err := s.roleRepo.GetTenantUserRoles(tenantID)
	if err != nil {
		return nil, err
	}

	result := make([]serializers.UserRoleResponse, len(userRoles))
	for i, ur := range userRoles {
		result[i] = serializers.UserRoleResponse{
			UserID:    ur.UserID,
			FirstName: ur.FirstName,
			LastName:  ur.LastName,
			Email:     ur.Email,
			RoleID:    ur.RoleID,
			RoleName:  ur.RoleName,
		}
	}
	return result, nil
}

// UpdateUserRole assigns a role to a user
func (s *RoleService) UpdateUserRole(userID, roleID int64) error {
	// Verify role exists
	_, err := s.roleRepo.FindByID(roleID)
	if err != nil {
		return ErrRoleNotFound
	}

	return s.roleRepo.SetUserRole(userID, roleID)
}

func roleDTOToResponse(r *repositories.RoleDTO) serializers.RoleResponse {
	return serializers.RoleResponse{
		ID:          r.ID,
		TenantID:    r.TenantID,
		Name:        r.Name,
		Description: r.Description,
		IsSystem:    r.TenantID == nil,
	}
}
