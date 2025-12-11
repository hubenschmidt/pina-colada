package agent

import (
	"github.com/pina-colada-co/agent-go/internal/agent/prompts"
	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
)

// NewTriageAgent creates the triage router that delegates to specialist workers
func NewTriageAgent(m model.LLM, workers []adkagent.Agent) (adkagent.Agent, error) {
	return llmagent.New(llmagent.Config{
		Name:        "triage",
		Model:       m,
		Description: "Routes requests to appropriate specialist workers",
		Instruction: prompts.TriageInstructions,
		SubAgents:   workers,
	})
}
