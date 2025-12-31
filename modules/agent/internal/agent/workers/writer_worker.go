package workers

import (
	"strings"
	"time"

	"github.com/nlpodyssey/openai-agents-go/agents"
	"github.com/nlpodyssey/openai-agents-go/modelsettings"

	"agent/internal/agent/prompts"
	"agent/internal/agent/tools"
)

// NewWriterWorker creates the document writer agent.
// Has access to crm_lookup, search_entity_documents, read_document, and crm_propose_create tools.
func NewWriterWorker(model string, settings *modelsettings.ModelSettings, allTools []agents.Tool) *agents.Agent {
	workerTools := tools.FilterTools(allTools,
		"crm_lookup",
		"search_entity_documents",
		"read_document",
		"crm_propose_create",
	)

	instructions := strings.Replace(prompts.WriterWorkerInstructions, "{{DATE}}", time.Now().Format("January 2, 2006"), 1)

	agent := agents.New("writer_worker").
		WithInstructions(instructions).
		WithModel(model).
		WithHandoffDescription("Generates documents (cover letters, emails, proposals) using existing examples as templates.").
		WithTools(workerTools...)

	if settings == nil {
		return agent
	}

	return agent.WithModelSettings(*settings)
}
