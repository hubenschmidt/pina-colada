package agent

import (
	"github.com/pina-colada-co/agent-go/internal/agent/prompts"
	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"
)

// NewTriageAgent creates the triage router that delegates to specialist workers via SubAgents
func NewTriageAgent(m model.LLM, workers []adkagent.Agent) (adkagent.Agent, error) {
	return llmagent.New(llmagent.Config{
		Name:        "triage",
		Model:       m,
		Description: "Routes requests to appropriate specialist workers",
		Instruction: prompts.TriageInstructions,
		SubAgents:   workers,
	})
}

// NewTriageAgentWithTools creates a triage router using agenttool-wrapped workers
// This is required when mixing GoogleSearch with function tools (Gemini API limitation)
func NewTriageAgentWithTools(m model.LLM, agentTools []tool.Tool) (adkagent.Agent, error) {
	return llmagent.New(llmagent.Config{
		Name:        "root_agent",
		Model:       m,
		Description: "Routes requests to appropriate specialist workers",
		Instruction: prompts.TriageInstructionsWithTools,
		Tools:       agentTools,
	})
}
