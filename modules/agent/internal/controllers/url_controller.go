package controllers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// URLResolver interface for resolving short codes to full URLs
type URLResolver interface {
	ResolveURL(code string) string
}

// URLController handles URL redirect requests
type URLController struct {
	resolver URLResolver
}

// NewURLController creates a new URL controller
func NewURLController(resolver URLResolver) *URLController {
	return &URLController{resolver: resolver}
}

// Redirect handles GET /u/{code} - redirects to full URL
func (c *URLController) Redirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		http.NotFound(w, r)
		return
	}

	fullURL := c.resolver.ResolveURL(code)
	if fullURL == "" {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, fullURL, http.StatusFound)
}
