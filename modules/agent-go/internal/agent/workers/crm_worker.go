package workers

import (
	"github.com/nlpodyssey/openai-agents-go/agents"
	"github.com/nlpodyssey/openai-agents-go/modelsettings"
	"github.com/pina-colada-co/agent-go/internal/agent/prompts"
	"github.com/pina-colada-co/agent-go/internal/agent/tools"
)

// NewCRMWorker creates the CRM specialist agent.
// Has access to crm_lookup, crm_list, crm_statuses, search_entity_documents, and read_document tools.
func NewCRMWorker(model string, settings *modelsettings.ModelSettings, allTools []agents.Tool) *agents.Agent {
	workerTools := tools.FilterTools(allTools,
		"crm_lookup",
		"crm_list",
		"crm_statuses",
		"search_entity_documents",
		"read_document",
	)

	agent := agents.New("crm_worker").
		WithInstructions(prompts.CRMWorkerInstructions).
		WithModel(model).
		WithHandoffDescription("Handles CRM data lookups - contacts, individuals, organizations, accounts, and job leads/pipeline").
		WithTools(workerTools...)

	if settings != nil {
		agent = agent.WithModelSettings(*settings)
	}

	return agent
}
