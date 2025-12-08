package controllers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pina-colada-co/agent-go/internal/middleware"
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
