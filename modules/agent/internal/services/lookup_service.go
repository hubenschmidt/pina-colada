package services

import (
	"agent/internal/repositories"
	"agent/internal/serializers"
)

type LookupService struct {
	lookupRepo *repositories.LookupRepository
}

func NewLookupService(lookupRepo *repositories.LookupRepository) *LookupService {
	return &LookupService{lookupRepo: lookupRepo}
}

func (s *LookupService) GetIndustries() ([]serializers.IndustryResponse, error) {
	industries, err := s.lookupRepo.FindAllIndustries()
	if err != nil {
		return nil, err
	}
	return serializers.IndustriesToResponse(industries), nil
}

func (s *LookupService) GetEmployeeCountRanges() ([]serializers.EmployeeCountRangeResponse, error) {
	ranges, err := s.lookupRepo.FindAllEmployeeCountRanges()
	if err != nil {
		return nil, err
	}
	return serializers.EmployeeCountRangesToResponse(ranges), nil
}

func (s *LookupService) GetRevenueRanges() ([]serializers.RevenueRangeResponse, error) {
	ranges, err := s.lookupRepo.FindAllRevenueRanges()
	if err != nil {
		return nil, err
	}
	return serializers.RevenueRangesToResponse(ranges), nil
}

func (s *LookupService) GetFundingStages() ([]serializers.FundingStageResponse, error) {
	stages, err := s.lookupRepo.FindAllFundingStages()
	if err != nil {
		return nil, err
	}
	return serializers.FundingStagesToResponse(stages), nil
}

func (s *LookupService) GetTaskStatuses() ([]serializers.StatusResponse, error) {
	statuses, err := s.lookupRepo.FindTaskStatuses()
	if err != nil {
		return nil, err
	}
	resp := make([]serializers.StatusResponse, len(statuses))
	for i := range statuses {
		resp[i] = serializers.StatusToResponse(&statuses[i])
	}
	return resp, nil
}

func (s *LookupService) GetTaskPriorities() ([]serializers.StatusResponse, error) {
	priorities, err := s.lookupRepo.FindTaskPriorities()
	if err != nil {
		return nil, err
	}
	resp := make([]serializers.StatusResponse, len(priorities))
	for i := range priorities {
		resp[i] = serializers.StatusToResponse(&priorities[i])
	}
	return resp, nil
}

func (s *LookupService) GetSalaryRanges() ([]serializers.SalaryRangeResponse, error) {
	ranges, err := s.lookupRepo.FindAllSalaryRanges()
	if err != nil {
		return nil, err
	}
	return serializers.SalaryRangesToResponse(ranges), nil
}

func (s *LookupService) CreateIndustry(name string) (*serializers.IndustryResponse, error) {
	industry, err := s.lookupRepo.CreateIndustry(name)
	if err != nil {
		return nil, err
	}
	resp := serializers.IndustryToResponse(industry)
	return &resp, nil
}
