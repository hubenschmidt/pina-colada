package controllers

import (
	"net/http"

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
