package controllers

import (
	"net/http"

	"github.com/pina-colada-co/agent-go/internal/middleware"
	"github.com/pina-colada-co/agent-go/internal/services"
)

// ProjectController handles project HTTP requests
type ProjectController struct {
	projectService *services.ProjectService
}

// NewProjectController creates a new project controller
func NewProjectController(projectService *services.ProjectService) *ProjectController {
	return &ProjectController{projectService: projectService}
}

// GetProjects handles GET /projects
func (c *ProjectController) GetProjects(w http.ResponseWriter, r *http.Request) {
	var tenantID *int64
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = &tid
	}

	results, err := c.projectService.GetProjects(tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, results)
}
