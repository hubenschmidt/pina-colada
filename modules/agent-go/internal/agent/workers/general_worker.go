package workers

import (
	"github.com/pina-colada-co/agent-go/internal/agent/prompts"
	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
)

// NewGeneralWorker creates the general-purpose assistant agent
func NewGeneralWorker(m model.LLM) (adkagent.Agent, error) {
	return llmagent.New(llmagent.Config{
		Name:        "general_worker",
		Model:       m,
		Description: "Handles general questions, conversation, resume analysis, and everything else",
		Instruction: prompts.GeneralWorkerInstructions,
	})
}
