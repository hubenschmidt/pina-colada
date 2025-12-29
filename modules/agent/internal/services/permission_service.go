package services

import (
	"context"
	"strings"

	"agent/internal/middleware"
	"agent/internal/repositories"
)

// PermissionService handles permission business logic
type PermissionService struct {
	permRepo *repositories.PermissionRepository
}

// NewPermissionService creates a new permission service
func NewPermissionService(permRepo *repositories.PermissionRepository) *PermissionService {
	return &PermissionService{permRepo: permRepo}
}

// HasPermission checks if a user has a specific permission
func (s *PermissionService) HasPermission(userID int64, resource, action string) (bool, error) {
	return s.permRepo.HasPermission(userID, resource, action)
}

// GetUserPermissions returns all permissions for a user as "resource:action" strings
func (s *PermissionService) GetUserPermissions(userID int64) ([]string, error) {
	perms, err := s.permRepo.GetUserPermissions(userID)
	if err != nil {
		return nil, err
	}

	result := make([]string, len(perms))
	for i, p := range perms {
		result[i] = p.Resource + ":" + p.Action
	}

	return result, nil
}

// CanAccess checks if the user in context has the specified permission
// Permission format: "resource:action" (e.g., "individual:read")
func (s *PermissionService) CanAccess(ctx context.Context, permission string) bool {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return false
	}

	parts := strings.SplitN(permission, ":", 2)
	if len(parts) != 2 {
		return false
	}

	has, err := s.HasPermission(userID, parts[0], parts[1])
	if err != nil {
		return false
	}

	return has
}
