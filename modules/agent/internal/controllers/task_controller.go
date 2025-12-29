package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"agent/internal/middleware"
	"agent/internal/schemas"
	"agent/internal/serializers"
	"agent/internal/services"
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
	scope := r.URL.Query().Get("scope")

	var tenantID *int64
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = &tid
	}

	var taskableType *string
	if tt := r.URL.Query().Get("taskable_type"); tt != "" {
		taskableType = &tt
	}

	taskableID := parseOptionalInt64(r.URL.Query().Get("taskable_id"))
	projectID := parseOptionalInt64(r.URL.Query().Get("projectId"))

	result, err := c.taskService.GetTasks(page, limit, orderBy, order, search, scope, tenantID, taskableType, taskableID, projectID)
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
	writeTaskResponse(w, http.StatusOK, result, err)
}

// DeleteTask handles DELETE /tasks/{id}
func (c *TaskController) DeleteTask(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task ID")
		return
	}

	err = c.taskService.DeleteTask(id)
	if errors.Is(err, services.ErrTaskNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, serializers.SuccessResponse{Success: true})
}

// CreateTask handles POST /tasks
func (c *TaskController) CreateTask(w http.ResponseWriter, r *http.Request) {
	var input schemas.TaskCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var tenantID *int64
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = &tid
	}

	userID, _ := middleware.GetUserID(r.Context())

	result, err := c.taskService.CreateTask(input, tenantID, userID)
	writeTaskResponse(w, http.StatusCreated, result, err)
}

// UpdateTask handles PUT /tasks/{id}
func (c *TaskController) UpdateTask(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task ID")
		return
	}

	var input schemas.TaskUpdate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, _ := middleware.GetUserID(r.Context())

	result, err := c.taskService.UpdateTask(id, input, userID)
	writeTaskResponse(w, http.StatusOK, result, err)
}

// GetTasksByEntity handles GET /tasks/entity/{entityType}/{entityID}
func (c *TaskController) GetTasksByEntity(w http.ResponseWriter, r *http.Request) {
	entityType := chi.URLParam(r, "entityType")
	entityIDStr := chi.URLParam(r, "entityID")

	entityID, err := strconv.ParseInt(entityIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid entity ID")
		return
	}

	var tenantID *int64
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = &tid
	}

	tasks, err := c.taskService.GetTasksByEntity(entityType, entityID, tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"items": tasks})
}

func writeTaskResponse(w http.ResponseWriter, successStatus int, result *serializers.TaskResponse, err error) {
	if errors.Is(err, services.ErrTaskNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if errors.Is(err, services.ErrTaskTitleRequired) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, successStatus, result)
}
