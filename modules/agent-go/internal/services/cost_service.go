package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// CostService handles provider costs business logic
type CostService struct{}

// NewCostService creates a new costs service
func NewCostService() *CostService {
	return &CostService{}
}

// ProviderCosts represents costs from a single provider
type ProviderCosts struct {
	Provider  string  `json:"provider"`
	Spend     float64 `json:"spend"`
	Currency  string  `json:"currency"`
	Period    string  `json:"period"`
	Scope     string  `json:"scope,omitempty"`
	FetchedAt string  `json:"fetched_at"`
}

// CombinedCostsResponse represents combined costs from all providers
type CombinedCostsResponse struct {
	Period     string         `json:"period"`
	FetchedAt  string         `json:"fetched_at"`
	OpenAI     *ProviderCosts `json:"openai"`
	Anthropic  *ProviderCosts `json:"anthropic"`
	TotalSpend float64        `json:"total_spend"`
}

// OrgCostsResponse represents org-wide costs
type OrgCostsResponse struct {
	Period     string         `json:"period"`
	FetchedAt  string         `json:"fetched_at"`
	Anthropic  *ProviderCosts `json:"anthropic"`
	TotalSpend float64        `json:"total_spend"`
}

// GetCombinedCosts fetches costs from all providers
func (s *CostService) GetCombinedCosts(period string) (*CombinedCostsResponse, error) {
	if period == "" {
		period = "monthly"
	}

	openaiCosts := s.fetchOpenAICosts(period)
	anthropicCosts := s.fetchAnthropicCosts(period)

	totalSpend := 0.0
	if openaiCosts != nil {
		totalSpend += openaiCosts.Spend
	}
	if anthropicCosts != nil {
		totalSpend += anthropicCosts.Spend
	}

	return &CombinedCostsResponse{
		Period:     period,
		FetchedAt:  time.Now().UTC().Format(time.RFC3339),
		OpenAI:     openaiCosts,
		Anthropic:  anthropicCosts,
		TotalSpend: totalSpend,
	}, nil
}

// GetOrgCosts fetches org-wide costs
func (s *CostService) GetOrgCosts(period string) (*OrgCostsResponse, error) {
	if period == "" {
		period = "monthly"
	}

	anthropicCosts := s.fetchAnthropicCosts(period)

	totalSpend := 0.0
	if anthropicCosts != nil {
		totalSpend = anthropicCosts.Spend
	}

	return &OrgCostsResponse{
		Period:     period,
		FetchedAt:  time.Now().UTC().Format(time.RFC3339),
		Anthropic:  anthropicCosts,
		TotalSpend: totalSpend,
	}, nil
}

var costPeriodOffsets = map[string][3]int{
	"daily":     {0, 0, -1},
	"weekly":    {0, 0, -7},
	"quarterly": {0, -3, 0},
	"annual":    {-1, 0, 0},
	"monthly":   {0, 0, -30},
}

func (s *CostService) getPeriodTimestamps(period string) (time.Time, time.Time) {
	now := time.Now().UTC()
	offset, ok := costPeriodOffsets[period]
	if !ok {
		offset = costPeriodOffsets["monthly"]
	}
	return now.AddDate(offset[0], offset[1], offset[2]), now
}

func (s *CostService) fetchOpenAICosts(period string) *ProviderCosts {
	adminKey := os.Getenv("OPENAI_ADMIN_KEY")
	if adminKey == "" {
		return nil
	}

	start, end := s.getPeriodTimestamps(period)

	url := fmt.Sprintf(
		"https://api.openai.com/v1/organization/costs?start_time=%d&end_time=%d&bucket_width=1d&limit=180",
		start.Unix(), end.Unix(),
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Authorization", "Bearer "+adminKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var data struct {
		Data []struct {
			Results []struct {
				Amount struct {
					Value float64 `json:"value"`
				} `json:"amount"`
			} `json:"results"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil
	}

	totalSpend := 0.0
	for _, bucket := range data.Data {
		for _, result := range bucket.Results {
			totalSpend += result.Amount.Value
		}
	}

	return &ProviderCosts{
		Provider:  "openai",
		Spend:     totalSpend,
		Currency:  "USD",
		Period:    period,
		FetchedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func (s *CostService) fetchAnthropicCosts(period string) *ProviderCosts {
	adminKey := os.Getenv("ANTHROPIC_ADMIN_KEY")
	if adminKey == "" {
		return nil
	}

	start, end := s.getPeriodTimestamps(period)
	startingAt := start.Format("2006-01-02T15:04:05Z")
	endingAt := end.Format("2006-01-02T15:04:05Z")

	url := fmt.Sprintf(
		"https://api.anthropic.com/v1/organizations/cost_report?starting_at=%s&ending_at=%s",
		startingAt, endingAt,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("x-api-key", adminKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var data struct {
		Data []struct {
			Results []struct {
				Amount string `json:"amount"`
			} `json:"results"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil
	}

	totalSpend := 0.0
	for _, bucket := range data.Data {
		for _, result := range bucket.Results {
			var amount float64
			fmt.Sscanf(result.Amount, "%f", &amount)
			totalSpend += amount
		}
	}

	return &ProviderCosts{
		Provider:  "anthropic",
		Spend:     totalSpend,
		Currency:  "USD",
		Period:    period,
		Scope:     "organization",
		FetchedAt: time.Now().UTC().Format(time.RFC3339),
	}
}
