package agent

import "log"

// Logging functions that match the Python agent's hooks.py patterns

// LogAgentStart logs when an agent begins execution
func LogAgentStart(agentName string) {
	log.Printf("▶️  AGENT START: %s", agentName)
}

// LogAgentEnd logs when an agent completes execution with output and token usage
func LogAgentEnd(agentName string, output string, tokens *TokenUsage) {
	log.Printf("✅ AGENT END: %s → %s", agentName, output)
	if tokens != nil && tokens.Total > 0 {
		log.Printf("📊 TOKENS [%s]: in=%d out=%d total=%d", agentName, tokens.Input, tokens.Output, tokens.Total)
	}
}

// LogHandoff logs when control transfers between agents
func LogHandoff(fromAgent, toAgent string) {
	log.Printf("🔀 HANDOFF: %s → %s", fromAgent, toAgent)
}

// LogToolStart logs when a tool begins execution
func LogToolStart(toolName string) {
	log.Printf("🔧 TOOL START: %s", toolName)
}

// LogToolEnd logs when a tool completes execution with result
func LogToolEnd(toolName string, result string) {
	log.Printf("🔧 TOOL END: %s → %s", toolName, result)
}

// LogLLMStart logs when an LLM call begins (debug level)
func LogLLMStart(agentName string) {
	log.Printf("💭 LLM START: %s", agentName)
}

// LogLLMEnd logs when an LLM call completes (debug level)
func LogLLMEnd(agentName string) {
	log.Printf("💭 LLM END: %s", agentName)
}

// LogInfo logs general info messages
func LogInfo(format string, args ...interface{}) {
	log.Printf(format, args...)
}

// LogWarning logs warning messages
func LogWarning(format string, args ...interface{}) {
	log.Printf("⚠️  "+format, args...)
}

// LogError logs error messages
func LogError(format string, args ...interface{}) {
	log.Printf("❌ "+format, args...)
}
