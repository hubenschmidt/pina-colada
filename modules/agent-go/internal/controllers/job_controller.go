package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pina-colada-co/agent-go/internal/middleware"
	"github.com/pina-colada-co/agent-go/internal/schemas"
	"github.com/pina-colada-co/agent-go/internal/serializers"
	"github.com/pina-colada-co/agent-go/internal/services"
)

type JobController struct {
	jobService *services.JobService
}

func NewJobController(jobService *services.JobService) *JobController {
	return &JobController{jobService: jobService}
}

func (c *JobController) GetJobs(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	projectID := parseOptionalInt64(r.URL.Query().Get("projectId"))

	var tenantID *int64
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = &tid
	}

	result, err := c.jobService.GetJobs(
		page, limit,
		r.URL.Query().Get("orderBy"),
		r.URL.Query().Get("order"),
		r.URL.Query().Get("search"),
		tenantID, projectID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (c *JobController) GetJob(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid job ID")
		return
	}

	result, err := c.jobService.GetJob(id)
	writeJobResponse(w, http.StatusOK, result, err)
}

func (c *JobController) CreateJob(w http.ResponseWriter, r *http.Request) {
	var input schemas.JobCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var tenantID *int64
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = &tid
	}
	userID, _ := middleware.GetUserID(r.Context())

	result, err := c.jobService.CreateJob(input, tenantID, userID)
	writeJobResponse(w, http.StatusCreated, result, err)
}

func (c *JobController) UpdateJob(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid job ID")
		return
	}

	var input schemas.JobUpdate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, _ := middleware.GetUserID(r.Context())

	result, err := c.jobService.UpdateJob(id, input, userID)
	writeJobResponse(w, http.StatusOK, result, err)
}

func (c *JobController) DeleteJob(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid job ID")
		return
	}

	err = c.jobService.DeleteJob(id)
	if errors.Is(err, services.ErrJobNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, serializers.SuccessResponse{Success: true})
}

func writeJobResponse(w http.ResponseWriter, successStatus int, result *serializers.JobDetailResponse, err error) {
	if errors.Is(err, services.ErrJobNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if errors.Is(err, services.ErrJobTitleRequired) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, successStatus, result)
}

func parseIDParam(r *http.Request, param string) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, param), 10, 64)
}

func parseOptionalInt64(s string) *int64 {
	if s == "" {
		return nil
	}
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil
	}
	return &id
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, serializers.ErrorResponse{Error: message})
}
