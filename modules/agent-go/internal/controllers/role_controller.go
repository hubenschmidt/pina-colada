package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pina-colada-co/agent-go/internal/middleware"
	"github.com/pina-colada-co/agent-go/internal/serializers"
	"github.com/pina-colada-co/agent-go/internal/services"
)

// RoleController handles role HTTP requests
type RoleController struct {
	roleService *services.RoleService
}

// NewRoleController creates a new role controller
func NewRoleController(roleService *services.RoleService) *RoleController {
	return &RoleController{roleService: roleService}
}

// List handles GET /admin/roles
func (c *RoleController) List(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeRoleError(w, http.StatusUnauthorized, "tenant not found")
		return
	}

	roles, err := c.roleService.GetRoles(tenantID)
	if err != nil {
		writeRoleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeRoleJSON(w, http.StatusOK, roles)
}

type createRoleRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

// Create handles POST /admin/roles
func (c *RoleController) Create(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeRoleError(w, http.StatusUnauthorized, "tenant not found")
		return
	}

	var req createRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeRoleError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	role, err := c.roleService.CreateRole(tenantID, req.Name, req.Description)
	if errors.Is(err, services.ErrRoleNameRequired) {
		writeRoleError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		writeRoleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeRoleJSON(w, http.StatusCreated, role)
}

type updateRoleRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

// Update handles PUT /admin/roles/{id}
func (c *RoleController) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeRoleError(w, http.StatusBadRequest, "invalid role id")
		return
	}

	var req updateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeRoleError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	role, err := c.roleService.UpdateRole(id, req.Name, req.Description)
	if errors.Is(err, services.ErrRoleNotFound) {
		writeRoleError(w, http.StatusNotFound, err.Error())
		return
	}
	if errors.Is(err, services.ErrSystemRoleEdit) {
		writeRoleError(w, http.StatusForbidden, err.Error())
		return
	}
	if errors.Is(err, services.ErrRoleNameRequired) {
		writeRoleError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		writeRoleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeRoleJSON(w, http.StatusOK, role)
}

// Delete handles DELETE /admin/roles/{id}
func (c *RoleController) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeRoleError(w, http.StatusBadRequest, "invalid role id")
		return
	}

	err = c.roleService.DeleteRole(id)
	if errors.Is(err, services.ErrRoleNotFound) {
		writeRoleError(w, http.StatusNotFound, err.Error())
		return
	}
	if errors.Is(err, services.ErrSystemRoleEdit) {
		writeRoleError(w, http.StatusForbidden, err.Error())
		return
	}
	if err != nil {
		writeRoleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetPermissions handles GET /admin/permissions
func (c *RoleController) GetPermissions(w http.ResponseWriter, r *http.Request) {
	perms, err := c.roleService.GetAllPermissions()
	if err != nil {
		writeRoleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeRoleJSON(w, http.StatusOK, perms)
}

// GetRolePermissions handles GET /admin/roles/{id}/permissions
func (c *RoleController) GetRolePermissions(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeRoleError(w, http.StatusBadRequest, "invalid role id")
		return
	}

	permIDs, err := c.roleService.GetRolePermissions(id)
	if err != nil {
		writeRoleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeRoleJSON(w, http.StatusOK, serializers.RolePermissionsResponse{
		RoleID:        id,
		PermissionIDs: permIDs,
	})
}

// GetAllRolePermissions handles GET /admin/role-permissions
func (c *RoleController) GetAllRolePermissions(w http.ResponseWriter, r *http.Request) {
	permMap, err := c.roleService.GetAllRolePermissions()
	if err != nil {
		writeRoleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeRoleJSON(w, http.StatusOK, permMap)
}

type updateRolePermissionsRequest struct {
	PermissionIDs []int64 `json:"permission_ids"`
}

// UpdateRolePermissions handles PUT /admin/roles/{id}/permissions
func (c *RoleController) UpdateRolePermissions(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeRoleError(w, http.StatusBadRequest, "invalid role id")
		return
	}

	var req updateRolePermissionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeRoleError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = c.roleService.UpdateRolePermissions(id, req.PermissionIDs)
	if errors.Is(err, services.ErrRoleNotFound) {
		writeRoleError(w, http.StatusNotFound, err.Error())
		return
	}
	if errors.Is(err, services.ErrSystemRoleEdit) {
		writeRoleError(w, http.StatusForbidden, err.Error())
		return
	}
	if err != nil {
		writeRoleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetUserRoles handles GET /admin/user-roles
func (c *RoleController) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeRoleError(w, http.StatusUnauthorized, "tenant not found")
		return
	}

	userRoles, err := c.roleService.GetTenantUserRoles(tenantID)
	if err != nil {
		writeRoleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeRoleJSON(w, http.StatusOK, userRoles)
}

type updateUserRoleRequest struct {
	RoleID int64 `json:"role_id"`
}

// UpdateUserRole handles PUT /admin/users/{id}/role
func (c *RoleController) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeRoleError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req updateUserRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeRoleError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = c.roleService.UpdateUserRole(userID, req.RoleID)
	if errors.Is(err, services.ErrRoleNotFound) {
		writeRoleError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeRoleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeRoleJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeRoleError(w http.ResponseWriter, status int, message string) {
	writeRoleJSON(w, status, serializers.ErrorResponse{Error: message})
}
