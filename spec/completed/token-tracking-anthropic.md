# Token Tracking for Anthropic Streaming

## Problem
Token usage tracking returns zeros for Anthropic/Claude models when using the OpenAI-compatible endpoint via chat completions.

## Root Cause
The SDK's `GetStreamOptionsParam` function only auto-enables `include_usage` when `IsOpenAI(client)` returns true (checks for `api.openai.com`):

```go
if modelSettings.IncludeUsage.Valid() {
    options.IncludeUsage = param.NewOpt(modelSettings.IncludeUsage.Value)
} else if h.IsOpenAI(client) {  // <-- fails for api.anthropic.com
    options.IncludeUsage = param.NewOpt(true)
}
```

Since we use `https://api.anthropic.com/v1/`, the check fails and `include_usage` is not set in `stream_options`.

## Solution
Explicitly set `IncludeUsage: true` in model settings for all agents.

## Implementation

### File: `internal/agent/orchestrator.go`

**Change:** Add `IncludeUsage` to `buildModelSettingsWithOptions`

```go
func buildModelSettingsWithOptions(settings services.LLMSettings, allowParallelToolCalls bool) *modelsettings.ModelSettings {
    ms := &modelsettings.ModelSettings{}

    // Always enable usage tracking for streaming (required for Anthropic)
    ms.IncludeUsage = param.NewOpt(true)

    // ... existing settings ...
}
```

## Verification
1. Set model to Claude (e.g., `claude-opus-4-5-20251101`)
2. Send a message through the agent
3. Check logs for non-zero values in `📊 TOTAL TOKENS: in=X out=Y total=Z`

## Status
- [ ] Implement fix in `buildModelSettingsWithOptions`
- [ ] Test with Anthropic model
- [ ] Verify token counts appear in frontend
