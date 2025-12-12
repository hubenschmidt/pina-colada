package workers

import (
	"github.com/pina-colada-co/agent-go/internal/agent/prompts"
	"github.com/pina-colada-co/agent-go/internal/tools"
	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"
)

// NewJobSearchWorker creates the job search specialist agent with Serper-based search
// Uses Serper API with domain exclusions to return direct company career page URLs
// Also has access to CRM tools for looking up resumes and contact data
func NewJobSearchWorker(m model.LLM, serperAPIKey string, crmTools []tool.Tool) (adkagent.Agent, error) {
	// Create Serper tools for job search
	serperTools := tools.NewSerperTools(serperAPIKey)
	serperToolList, err := serperTools.BuildTools()
	if err != nil {
		return nil, err
	}

	// Combine serper tools with CRM tools
	allTools := append(serperToolList, crmTools...)

	return llmagent.New(llmagent.Config{
		Name:        "job_search",
		Model:       m,
		Description: "Searches for jobs using filtered web search that returns direct company career page URLs only. Can also look up CRM records and read documents like resumes.",
		Instruction: prompts.JobSearchWorkerInstructions,
		Tools:       allTools,
	})
}
