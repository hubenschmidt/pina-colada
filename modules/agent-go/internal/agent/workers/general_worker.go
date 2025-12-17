package workers

import (
	"github.com/nlpodyssey/openai-agents-go/agents"
	"github.com/nlpodyssey/openai-agents-go/modelsettings"
	"github.com/pina-colada-co/agent-go/internal/agent/prompts"
	"github.com/pina-colada-co/agent-go/internal/agent/tools"
)

// NewGeneralWorker creates the general-purpose assistant agent.
// Has access to crm_lookup, crm_list, search_entity_documents, and read_document tools.
// Does NOT have access to job_search.
func NewGeneralWorker(model string, settings *modelsettings.ModelSettings, allTools []agents.Tool) *agents.Agent {
	workerTools := tools.FilterTools(allTools,
		"crm_lookup",
		"crm_list",
		"search_entity_documents",
		"read_document",
	)

	agent := agents.New("general_worker").
		WithInstructions(prompts.GeneralWorkerInstructions).
		WithModel(model).
		WithHandoffDescription("Handles general questions, conversation, resume analysis, and everything else. Can look up CRM records and documents.").
		WithTools(workerTools...)

	if settings != nil {
		agent = agent.WithModelSettings(*settings)
	}

	return agent
}
