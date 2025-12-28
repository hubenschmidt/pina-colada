package workers

import (
	"github.com/nlpodyssey/openai-agents-go/agents"
	"github.com/nlpodyssey/openai-agents-go/modelsettings"
	"github.com/pina-colada-co/agent-go/internal/agent/prompts"
	"github.com/pina-colada-co/agent-go/internal/agent/tools"
)

// NewJobSearchWorker creates the job search specialist agent.
// Has access to job_search, crm_lookup, search_entity_documents, and read_document tools.
func NewJobSearchWorker(model string, settings *modelsettings.ModelSettings, allTools []agents.Tool) *agents.Agent {
	workerTools := tools.FilterTools(allTools,
		"job_search",
		"crm_lookup",
		"search_entity_documents",
		"read_document",
		"send_email",
	)

	agent := agents.New("job_search").
		WithInstructions(prompts.JobSearchWorkerInstructions).
		WithModel(model).
		WithHandoffDescription("Searches for jobs using filtered web search that returns direct company career page URLs only. Can also look up CRM records and read documents like resumes.").
		WithTools(workerTools...)

	if settings != nil {
		agent = agent.WithModelSettings(*settings)
	}

	return agent
}
