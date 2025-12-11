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
func NewJobSearchWorker(m model.LLM, additionalTools []tool.Tool) (adkagent.Agent, error) {
	// Start with Google Search as the primary tool
	tools := []tool.Tool{
		geminitool.GoogleSearch{},
	}

	// Add any additional tools (like CRM lookup for resume)
	tools = append(tools, additionalTools...)

	return llmagent.New(llmagent.Config{
		Name:        "job_search",
		Model:       m,
		Description: "Searches for jobs, finds company career pages, and helps with job applications. Has access to Google Search to find direct company career URLs.",
		Instruction: prompts.JobSearchWorkerInstructions,
		Tools:       tools,
	})
}
