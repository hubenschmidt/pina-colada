package services

import (
	"errors"
	"time"

	"github.com/pina-colada-co/agent-go/internal/models"
	"github.com/pina-colada-co/agent-go/internal/repositories"
	"github.com/pina-colada-co/agent-go/internal/schemas"
	"github.com/pina-colada-co/agent-go/internal/serializers"
)

var ErrJobNotFound = errors.New("job not found")
var ErrJobTitleRequired = errors.New("job_title is required")

type JobService struct {
	jobRepo *repositories.JobRepository
}

func NewJobService(jobRepo *repositories.JobRepository) *JobService {
	return &JobService{jobRepo: jobRepo}
}

func (s *JobService) GetJobs(page, pageSize int, orderBy, order, search string, tenantID, projectID *int64) (*serializers.PagedResponse, error) {
	params := repositories.NewPaginationParams(page, pageSize, orderBy, order)
	params.Search = search

	result, err := s.jobRepo.FindAll(params, tenantID, projectID)
	if err != nil {
		return nil, err
	}

	items := make([]serializers.JobListResponse, len(result.Items))
	for i := range result.Items {
		items[i] = serializers.JobToListResponse(&result.Items[i])
	}

	resp := serializers.NewPagedResponse(items, result.TotalCount, result.Page, result.PageSize, result.TotalPages)
	return &resp, nil
}

func (s *JobService) GetJob(id int64) (*serializers.JobDetailResponse, error) {
	job, err := s.jobRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, ErrJobNotFound
	}

	resp := serializers.JobToDetailResponse(job)
	return &resp, nil
}

func (s *JobService) DeleteJob(id int64) error {
	job, err := s.jobRepo.FindByID(id)
	if err != nil {
		return err
	}
	if job == nil {
		return ErrJobNotFound
	}

	return s.jobRepo.Delete(id)
}

func (s *JobService) CreateJob(input schemas.JobCreate, tenantID *int64, userID int64) (*serializers.JobDetailResponse, error) {
	if input.JobTitle == "" {
		return nil, ErrJobTitleRequired
	}

	lead := &models.Lead{
		TenantID:    tenantID,
		Type:        "Job",
		Title:       input.JobTitle,
		Description: input.Description,
		Source:      &input.Source,
		CreatedBy:   userID,
		UpdatedBy:   userID,
	}

	job := &models.Job{
		JobTitle:      input.JobTitle,
		Description:   input.Description,
		JobURL:        input.JobURL,
		SalaryRange:   input.SalaryRange,
		SalaryRangeID: input.SalaryRangeID,
		ResumeDate:    parseDate(input.Resume),
	}

	err := s.jobRepo.Create(job, lead, input.ProjectIDs)
	if err != nil {
		return nil, err
	}

	return s.GetJob(job.ID)
}

func (s *JobService) UpdateJob(id int64, input schemas.JobUpdate, userID int64) (*serializers.JobDetailResponse, error) {
	job, err := s.jobRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, ErrJobNotFound
	}

	err = s.applyJobUpdates(job, input)
	if err != nil {
		return nil, err
	}

	err = s.applyLeadUpdates(id, input, userID)
	if err != nil {
		return nil, err
	}

	err = s.applyProjectUpdates(id, input.ProjectIDs)
	if err != nil {
		return nil, err
	}

	return s.GetJob(id)
}

func (s *JobService) applyJobUpdates(job *models.Job, input schemas.JobUpdate) error {
	updates := buildJobUpdates(input)
	if len(updates) == 0 {
		return nil
	}
	return s.jobRepo.Update(job, updates)
}

func (s *JobService) applyLeadUpdates(id int64, input schemas.JobUpdate, userID int64) error {
	updates := buildLeadUpdates(input, userID)
	if len(updates) == 0 {
		return nil
	}
	return s.jobRepo.UpdateLead(id, updates)
}

func (s *JobService) applyProjectUpdates(id int64, projectIDs []int64) error {
	if projectIDs == nil {
		return nil
	}
	return s.jobRepo.UpdateProjectAssociations(id, projectIDs)
}

func buildJobUpdates(input schemas.JobUpdate) map[string]interface{} {
	updates := make(map[string]interface{})

	if input.JobTitle != nil {
		updates["job_title"] = *input.JobTitle
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.JobURL != nil {
		updates["job_url"] = *input.JobURL
	}
	if input.SalaryRange != nil {
		updates["salary_range"] = *input.SalaryRange
	}
	if input.SalaryRangeID != nil {
		updates["salary_range_id"] = *input.SalaryRangeID
	}
	if t := parseDate(input.Resume); t != nil {
		updates["resume_date"] = t
	}

	return updates
}

func buildLeadUpdates(input schemas.JobUpdate, userID int64) map[string]interface{} {
	updates := make(map[string]interface{})

	if input.JobTitle != nil {
		updates["title"] = *input.JobTitle
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.Source != nil {
		updates["source"] = *input.Source
	}
	updates["updated_by"] = userID

	return updates
}

func parseDate(dateStr *string) *time.Time {
	if dateStr == nil || *dateStr == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *dateStr)
	if err != nil {
		return nil
	}
	return &t
}
