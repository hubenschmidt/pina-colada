package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pina-colada-co/agent-go/internal/middleware"
	"github.com/pina-colada-co/agent-go/internal/schemas"
	"github.com/pina-colada-co/agent-go/internal/services"
)

// AuthController handles auth HTTP requests
type AuthController struct {
	authService *services.AuthService
}

// NewAuthController creates a new auth controller
func NewAuthController(authService *services.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

// GetMe handles GET /auth/me
func (c *AuthController) GetMe(w http.ResponseWriter, r *http.Request) {
	auth0Sub, ok := middleware.GetAuth0Sub(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	email, _ := middleware.GetEmail(r.Context())

	result, err := c.authService.GetMe(auth0Sub, email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetUserTenant handles GET /users/{email}/tenant
func (c *AuthController) GetUserTenant(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	if email == "" {
		writeError(w, http.StatusBadRequest, "email is required")
		return
	}

	result, err := c.authService.GetUserTenantByEmail(email)
	if errors.Is(err, services.ErrUserNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// SetSelectedProjectRequest is the request body for setting selected project
type SetSelectedProjectRequest struct {
	ProjectID *int64 `json:"project_id"`
}

// SetSelectedProject handles PUT /users/me/selected-project
func (c *AuthController) SetSelectedProject(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeError(w, http.StatusBadRequest, "tenant not set")
		return
	}

	var req SetSelectedProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := c.authService.SetSelectedProject(userID, tenantID, req.ProjectID)
	if errors.Is(err, services.ErrProjectNotInTenant) {
		writeError(w, http.StatusForbidden, "project does not belong to tenant")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]*int64{"selected_project_id": result})
}

// GetTenantUsers handles GET /users
func (c *AuthController) GetTenantUsers(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeError(w, http.StatusBadRequest, "tenant not set")
		return
	}

	users, err := c.authService.GetTenantUsers(tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, users)
}

// CreateTenant handles POST /auth/tenant/create
func (c *AuthController) CreateTenant(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var input schemas.TenantCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if input.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	result, err := c.authService.CreateTenant(userID, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, result)
}
