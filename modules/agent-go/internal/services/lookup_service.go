package services

import (
	"github.com/pina-colada-co/agent-go/internal/models"
	"github.com/pina-colada-co/agent-go/internal/repositories"
)

type LookupService struct {
	lookupRepo *repositories.LookupRepository
}

func NewLookupService(lookupRepo *repositories.LookupRepository) *LookupService {
	return &LookupService{lookupRepo: lookupRepo}
}

func (s *LookupService) GetIndustries() ([]models.Industry, error) {
	return s.lookupRepo.FindAllIndustries()
}

func (s *LookupService) GetEmployeeCountRanges() ([]models.EmployeeCountRange, error) {
	return s.lookupRepo.FindAllEmployeeCountRanges()
}

func (s *LookupService) GetRevenueRanges() ([]models.RevenueRange, error) {
	return s.lookupRepo.FindAllRevenueRanges()
}

func (s *LookupService) GetFundingStages() ([]models.FundingStage, error) {
	return s.lookupRepo.FindAllFundingStages()
}

func (s *LookupService) GetTaskStatuses() ([]models.Status, error) {
	return s.lookupRepo.FindTaskStatuses()
}

func (s *LookupService) GetTaskPriorities() ([]models.Status, error) {
	return s.lookupRepo.FindTaskPriorities()
}

func (s *LookupService) GetSalaryRanges() ([]models.SalaryRange, error) {
	return s.lookupRepo.FindAllSalaryRanges()
}
