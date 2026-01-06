package workers

import (
	"github.com/nlpodyssey/openai-agents-go/agents"
	"github.com/nlpodyssey/openai-agents-go/modelsettings"
	"agent/internal/agent/prompts"
	"agent/internal/agent/tools"
)

// NewCRMWorker creates the CRM specialist agent.
// Has access to CRM read/write tools and document tools.
func NewCRMWorker(model string, settings *modelsettings.ModelSettings, allTools []agents.Tool) *agents.Agent {
	workerTools := tools.FilterTools(allTools,
		"crm_lookup",
		"crm_list",
		"crm_propose_create",
		"crm_propose_batch_create",
		"crm_propose_update",
		"crm_propose_batch_update",
		"crm_propose_bulk_update_all",
		"crm_propose_delete",
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
