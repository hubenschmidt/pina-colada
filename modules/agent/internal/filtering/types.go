package filtering

import (
	"sync"
)

// JobResult represents a job search result from any source (Serper, RemoteOK, Remotive, etc.)
type JobResult struct {
	Title                string
	Company              string
	URL                  string
	Snippet              string
	DatePosted           string
	DatePostedConfidence string
	FullText             string
}

// AgentReview represents a single result review from the LLM
type AgentReview struct {
	Index    int    `json:"index"`
	Approved bool   `json:"approved"`
	Reason   string `json:"reason"`
}

// AgentReviewResponse is the JSON response from the LLM
type AgentReviewResponse struct {
	Reviews []AgentReview `json:"reviews"`
}

// ReviewedJobResult extends JobResult with agent analysis
type ReviewedJobResult struct {
	JobResult
	Approved bool
	Reason   string
}

// DedupData holds preloaded data for deduplication
type DedupData struct {
	mu            sync.RWMutex
	URLSource     map[string]string // URL -> source ("Job", "Proposal", "Rejected")
	JobIndex      map[string]bool   // normalized "company|title" -> exists (for jobs)
	ProposalIndex map[string]bool   // normalized "company|title" -> exists (for proposals)
}

// NewDedupData creates an initialized DedupData
func NewDedupData() *DedupData {
	return &DedupData{
		URLSource:     make(map[string]string),
		JobIndex:      make(map[string]bool),
		ProposalIndex: make(map[string]bool),
	}
}
