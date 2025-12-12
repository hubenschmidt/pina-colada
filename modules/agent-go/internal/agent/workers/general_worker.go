package workers

import (
	"github.com/pina-colada-co/agent-go/internal/agent/prompts"
	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"
)

// NewGeneralWorker creates the general-purpose assistant agent
// Has access to CRM tools for looking up records and documents
func NewGeneralWorker(m model.LLM, crmTools []tool.Tool) (adkagent.Agent, error) {
	return llmagent.New(llmagent.Config{
		Name:        "general_worker",
		Model:       m,
		Description: "Handles general questions, conversation, resume analysis, and everything else. Can look up CRM records and documents.",
		Instruction: prompts.GeneralWorkerInstructions,
		Tools:       crmTools,
	})
}
