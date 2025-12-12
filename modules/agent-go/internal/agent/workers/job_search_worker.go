package workers

import (
	"github.com/pina-colada-co/agent-go/internal/agent/prompts"
	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/geminitool"
)

// NewJobSearchWorker creates the job search specialist agent with Google Search
// Note: GoogleSearch cannot be mixed with custom function tools, so we only use GoogleSearch here
func NewJobSearchWorker(m model.LLM, additionalTools []tool.Tool) (adkagent.Agent, error) {
	// GoogleSearch is a special grounding tool - use it alone
	tools := []tool.Tool{
		geminitool.GoogleSearch{},
	}

	// NOTE: We intentionally do NOT add additionalTools here because
	// geminitool.GoogleSearch cannot be combined with functiontool tools
	// The job search worker should delegate CRM lookups back to the CRM worker
	_ = additionalTools // suppress unused warning

	return llmagent.New(llmagent.Config{
		Name:        "job_search",
		Model:       m,
		Description: "Searches for jobs, finds company career pages, and helps with job applications. Has access to Google Search to find direct company career URLs.",
		Instruction: prompts.JobSearchWorkerInstructions,
		Tools:       tools,
	})
}
