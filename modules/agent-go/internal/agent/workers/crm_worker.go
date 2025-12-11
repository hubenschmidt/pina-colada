package workers

import (
	"github.com/pina-colada-co/agent-go/internal/agent/prompts"
	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"
)

// NewCRMWorker creates the CRM specialist agent
func NewCRMWorker(m model.LLM, tools []tool.Tool) (adkagent.Agent, error) {
	return llmagent.New(llmagent.Config{
		Name:        "crm_worker",
		Model:       m,
		Description: "Handles CRM data lookups - contacts, individuals, organizations, accounts",
		Instruction: prompts.CRMWorkerInstructions,
		Tools:       tools,
	})
}
