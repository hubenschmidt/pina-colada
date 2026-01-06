package services

import (
	"errors"
	"strings"
	"time"

	apperrors "agent/internal/errors"
	"agent/internal/repositories"
	"agent/internal/schemas"
	"agent/internal/serializers"
)

var ErrJobNotFound = errors.New("job not found")
var ErrJobTitleRequired = errors.New("job_title is required")
var ErrAccountRequired = errors.New("account is required")

type JobService struct {
	jobRepo    *repositories.JobRepository
	orgRepo    *repositories.OrganizationRepository
	indRepo    *repositories.IndividualRepository
	lookupRepo *repositories.LookupRepository
}

func NewJobService(
	jobRepo *repositories.JobRepository,
	orgRepo *repositories.OrganizationRepository,
	indRepo *repositories.IndividualRepository,
	lookupRepo *repositories.LookupRepository,
) *JobService {
	return &JobService{
		jobRepo:    jobRepo,
		orgRepo:    orgRepo,
		indRepo:    indRepo,
		lookupRepo: lookupRepo,
	}
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

	accountID, err := s.resolveAccount(input, tenantID, userID)
	if err != nil {
		return nil, err
	}

	statusID, err := s.resolveStatusID(input.Status)
	if err != nil {
		return nil, err
	}

	repoInput := repositories.JobCreateInput{
		TenantID:        tenantID,
		UserID:          userID,
		JobTitle:        input.JobTitle,
		Description:     input.Description,
		Source:          input.Source,
		JobURL:          input.JobURL,
		SalaryRange:     input.SalaryRange,
		SalaryRangeID:   input.SalaryRangeID,
		ResumeDate:      parseDate(input.Resume),
		ProjectIDs:      input.ProjectIDs,
		AccountID:       accountID,
		CurrentStatusID: statusID,
	}

	jobID, err := s.jobRepo.Create(repoInput)
	if err != nil {
		return nil, err
	}

	return s.GetJob(jobID)
}

// resolveStatusID resolves a status name to its ID
func (s *JobService) resolveStatusID(statusName string) (*int64, error) {
	if statusName == "" {
		return nil, nil
	}
	status, err := s.lookupRepo.FindStatusByName(statusName, "job")
	if errors.Is(err, apperrors.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &status.ID, nil
}

// resolveAccount resolves the account based on account_type and account name
func (s *JobService) resolveAccount(input schemas.JobCreate, tenantID *int64, userID int64) (*int64, error) {
	if input.Account == nil || *input.Account == "" {
		return nil, nil
	}

	if input.AccountType == "Individual" {
		return s.resolveIndividualAccount(*input.Account, tenantID, userID)
	}

	return s.resolveOrganizationAccount(*input.Account, tenantID, userID)
}

func (s *JobService) resolveIndividualAccount(name string, tenantID *int64, userID int64) (*int64, error) {
	firstName, lastName := parseFullName(name)
	ind, err := s.indRepo.FindOrCreateByName(firstName, lastName, tenantID, userID)
	if err != nil {
		return nil, err
	}
	return ind.AccountID, nil
}

func (s *JobService) resolveOrganizationAccount(name string, tenantID *int64, userID int64) (*int64, error) {
	org, err := s.orgRepo.FindOrCreateByName(name, tenantID, userID)
	if err != nil {
		return nil, err
	}
	return org.AccountID, nil
}

func parseFullName(name string) (firstName, lastName string) {
	parts := strings.SplitN(strings.TrimSpace(name), " ", 2)
	firstName = parts[0]
	if len(parts) > 1 {
		lastName = parts[1]
	}
	return
}

func (s *JobService) UpdateJob(id int64, input schemas.JobUpdate, tenantID *int64, userID int64) (*serializers.JobDetailResponse, error) {
	job, err := s.jobRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, ErrJobNotFound
	}

	err = s.applyJobUpdates(id, input)
	if err != nil {
		return nil, err
	}

	err = s.applyLeadUpdates(id, input, tenantID, userID)
	if err != nil {
		return nil, err
	}

	err = s.applyProjectUpdates(id, input.ProjectIDs)
	if err != nil {
		return nil, err
	}

	return s.GetJob(id)
}

func (s *JobService) applyJobUpdates(id int64, input schemas.JobUpdate) error {
	updates := buildJobUpdates(input)
	if len(updates) == 0 {
		return nil
	}
	return s.jobRepo.Update(id, updates)
}

func (s *JobService) applyLeadUpdates(id int64, input schemas.JobUpdate, tenantID *int64, userID int64) error {
	updates := buildLeadUpdates(input, userID)

	if err := s.addAccountUpdate(updates, input.Account, tenantID, userID); err != nil {
		return err
	}

	if err := s.addStatusUpdate(updates, input.Status); err != nil {
		return err
	}

	if input.LeadStatusID != nil && *input.LeadStatusID != "" {
		updates["current_status_id"] = *input.LeadStatusID
	}

	if len(updates) == 0 {
		return nil
	}
	return s.jobRepo.UpdateLead(id, updates)
}

func (s *JobService) addAccountUpdate(updates map[string]interface{}, account *string, tenantID *int64, userID int64) error {
	if account == nil {
		return nil
	}
	if *account == "" {
		return nil
	}
	org, err := s.orgRepo.FindOrCreateByName(*account, tenantID, userID)
	if err != nil {
		return err
	}
	if org.AccountID == nil {
		return nil
	}
	updates["account_id"] = *org.AccountID
	return nil
}

func (s *JobService) addStatusUpdate(updates map[string]interface{}, status *string) error {
	if status == nil || *status == "" {
		return nil
	}
	st, err := s.lookupRepo.FindStatusByName(*status, "job")
	if err != nil {
		return err
	}
	if st == nil {
		return nil
	}
	updates["current_status_id"] = st.ID
	return nil
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

	if input.Source != nil {
		updates["source"] = *input.Source
	}
	updates["updated_by"] = userID

	return updates
}

// GetRecentResumeDate returns the most recent resume date as a string
func (s *JobService) GetRecentResumeDate() (*string, error) {
	resumeDate, err := s.jobRepo.GetRecentResumeDate()
	if errors.Is(err, apperrors.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	dateStr := resumeDate.Format("2006-01-02")
	return &dateStr, nil
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

// GetLeadStatuses returns all job-related statuses
func (s *JobService) GetLeadStatuses() ([]serializers.StatusResponse, error) {
	statuses, err := s.jobRepo.FindAllStatuses()
	if err != nil {
		return nil, err
	}

	resp := make([]serializers.StatusResponse, len(statuses))
	for i := range statuses {
		resp[i] = serializers.StatusToResponse(&statuses[i])
	}
	return resp, nil
}

// GetLeads returns job leads optionally filtered by status names
func (s *JobService) GetLeads(statusNames []string, tenantID *int64) ([]serializers.JobDetailResponse, error) {
	jobs, err := s.jobRepo.FindJobsByStatusNames(statusNames, tenantID)
	if err != nil {
		return nil, err
	}

	resp := make([]serializers.JobDetailResponse, len(jobs))
	for i := range jobs {
		resp[i] = serializers.JobToDetailResponse(&jobs[i])
		if jobs[i].Lead != nil && jobs[i].Lead.CurrentStatus != nil {
			resp[i].LeadStatus = serializers.StatusToResponse(jobs[i].Lead.CurrentStatus)
		}
	}
	return resp, nil
}

// MarkLeadAsApplied marks a job lead as applied
func (s *JobService) MarkLeadAsApplied(jobID int64, userID int64) (*serializers.JobDetailResponse, error) {
	job, err := s.jobRepo.FindByID(jobID)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, ErrJobNotFound
	}

	if err := s.jobRepo.UpdateJobStatus(jobID, "applied", userID); err != nil {
		return nil, err
	}

	return s.GetJob(jobID)
}

// MarkLeadAsDoNotApply marks a job lead as do not apply
func (s *JobService) MarkLeadAsDoNotApply(jobID int64, userID int64) (*serializers.JobDetailResponse, error) {
	job, err := s.jobRepo.FindByID(jobID)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, ErrJobNotFound
	}

	if err := s.jobRepo.UpdateJobStatus(jobID, "do_not_apply", userID); err != nil {
		return nil, err
	}

	return s.GetJob(jobID)
}
