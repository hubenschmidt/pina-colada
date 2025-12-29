package controllers

import (
	"encoding/json"
	"net/http"

	"agent/internal/schemas"
	"agent/internal/services"
)

type LookupController struct {
	lookupService *services.LookupService
}

func NewLookupController(lookupService *services.LookupService) *LookupController {
	return &LookupController{lookupService: lookupService}
}

func (c *LookupController) GetIndustries(w http.ResponseWriter, r *http.Request) {
	industries, err := c.lookupService.GetIndustries()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, industries)
}

func (c *LookupController) GetEmployeeCountRanges(w http.ResponseWriter, r *http.Request) {
	ranges, err := c.lookupService.GetEmployeeCountRanges()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ranges)
}

func (c *LookupController) GetRevenueRanges(w http.ResponseWriter, r *http.Request) {
	ranges, err := c.lookupService.GetRevenueRanges()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ranges)
}

func (c *LookupController) GetFundingStages(w http.ResponseWriter, r *http.Request) {
	stages, err := c.lookupService.GetFundingStages()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, stages)
}

func (c *LookupController) GetTaskStatuses(w http.ResponseWriter, r *http.Request) {
	statuses, err := c.lookupService.GetTaskStatuses()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, statuses)
}

func (c *LookupController) GetTaskPriorities(w http.ResponseWriter, r *http.Request) {
	priorities, err := c.lookupService.GetTaskPriorities()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, priorities)
}

func (c *LookupController) GetSalaryRanges(w http.ResponseWriter, r *http.Request) {
	ranges, err := c.lookupService.GetSalaryRanges()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ranges)
}

func (c *LookupController) CreateIndustry(w http.ResponseWriter, r *http.Request) {
	var input schemas.IndustryCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if input.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	industry, err := c.lookupService.CreateIndustry(input.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, industry)
}
