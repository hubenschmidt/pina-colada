package controllers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pina-colada-co/agent-go/internal/middleware"
	"github.com/pina-colada-co/agent-go/internal/serializers"
	"github.com/pina-colada-co/agent-go/internal/services"
)

// TaskController handles task HTTP requests
type TaskController struct {
	taskService *services.TaskService
}

// NewTaskController creates a new task controller
func NewTaskController(taskService *services.TaskService) *TaskController {
	return &TaskController{taskService: taskService}
}

// GetTasks handles GET /tasks
func (c *TaskController) GetTasks(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	orderBy := r.URL.Query().Get("orderBy")
	order := r.URL.Query().Get("order")
	search := r.URL.Query().Get("search")

	var tenantID *int64
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = &tid
	}

	var taskableType *string
	if tt := r.URL.Query().Get("taskable_type"); tt != "" {
		taskableType = &tt
	}

	var taskableID *int64
	if tid := r.URL.Query().Get("taskable_id"); tid != "" {
		if id, err := strconv.ParseInt(tid, 10, 64); err == nil {
			taskableID = &id
		}
	}

	result, err := c.taskService.GetTasks(page, limit, orderBy, order, search, tenantID, taskableType, taskableID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetTask handles GET /tasks/{id}
func (c *TaskController) GetTask(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task ID")
		return
	}

	result, err := c.taskService.GetTask(id)
	if err != nil {
		if err.Error() == "task not found" {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// DeleteTask handles DELETE /tasks/{id}
func (c *TaskController) DeleteTask(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task ID")
		return
	}

	if err := c.taskService.DeleteTask(id); err != nil {
		if err.Error() == "task not found" {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, serializers.SuccessResponse{Success: true})
}
