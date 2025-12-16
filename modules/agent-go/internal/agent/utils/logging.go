package utils

import (
	"log"
	"time"
)

// Logging functions that match the Python agent's hooks.py patterns

// LogAgentStart logs when an agent begins execution
func LogAgentStart(agentName string) {
	log.Printf("▶️  AGENT START: %s", agentName)
}

// LogAgentEnd logs when an agent completes execution with output and token usage
func LogAgentEnd(agentName string, output string, inputTokens, outputTokens, totalTokens int32) {
	log.Printf("✅ AGENT END: %s → %s", agentName, output)
	if totalTokens > 0 {
		log.Printf("📊 TOKENS [%s]: in=%d out=%d total=%d", agentName, inputTokens, outputTokens, totalTokens)
	}
}

// LogAgentEndWithDuration logs agent completion with timing
func LogAgentEndWithDuration(agentName string, output string, inputTokens, outputTokens, totalTokens int32, duration time.Duration) {
	log.Printf("✅ AGENT END: %s (%dms) → %s", agentName, duration.Milliseconds(), output)
	if totalTokens > 0 {
		log.Printf("📊 TOKENS [%s]: in=%d out=%d total=%d", agentName, inputTokens, outputTokens, totalTokens)
	}
}

// LogHandoff logs when control transfers between agents
func LogHandoff(fromAgent, toAgent string) {
	log.Printf("🔀 HANDOFF: %s → %s", fromAgent, toAgent)
}

// LogHandoffWithDuration logs handoff with timing for the completing agent
func LogHandoffWithDuration(fromAgent, toAgent string, duration time.Duration) {
	log.Printf("🔀 HANDOFF: %s (%dms) → %s", fromAgent, duration.Milliseconds(), toAgent)
}

// LogHandoffWithTokens logs handoff with timing and token usage
func LogHandoffWithTokens(fromAgent, toAgent string, duration time.Duration, inputTokens, outputTokens, totalTokens int32) {
	log.Printf("🔀 HANDOFF: %s (%dms, %d tokens) → %s", fromAgent, duration.Milliseconds(), totalTokens, toAgent)
	if totalTokens > 0 {
		log.Printf("📊 TOKENS [%s]: in=%d out=%d total=%d", fromAgent, inputTokens, outputTokens, totalTokens)
	}
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
