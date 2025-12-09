package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pina-colada-co/agent-go/internal/schemas"
	"github.com/pina-colada-co/agent-go/internal/services"
)

// TechnologyController handles technology HTTP requests
type TechnologyController struct {
	techService *services.TechnologyService
}

// NewTechnologyController creates a new technology controller
func NewTechnologyController(techService *services.TechnologyService) *TechnologyController {
	return &TechnologyController{techService: techService}
}

// GetTechnologies handles GET /technologies
func (c *TechnologyController) GetTechnologies(w http.ResponseWriter, r *http.Request) {
	var category *string
	if cat := r.URL.Query().Get("category"); cat != "" {
		category = &cat
	}

	techs, err := c.techService.GetAllTechnologies(category)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, techs)
}

// GetTechnology handles GET /technologies/{id}
func (c *TechnologyController) GetTechnology(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid technology ID")
		return
	}

	tech, err := c.techService.GetTechnology(id)
	if errors.Is(err, services.ErrTechnologyNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, tech)
}

// CreateTechnology handles POST /technologies
func (c *TechnologyController) CreateTechnology(w http.ResponseWriter, r *http.Request) {
	var input schemas.TechnologyCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if input.Name == "" || input.Category == "" {
		writeError(w, http.StatusBadRequest, "name and category are required")
		return
	}

	tech, err := c.techService.CreateTechnology(input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, tech)
}

func parseTechIDParam(r *http.Request, param string) (int64, error) {
	return parseIDParam(r, param)
}

// Note: parseIDParam is defined in job_controller.go and should be shared
func init() {
	// ensure chi is imported for URLParam
	_ = chi.URLParam
}
