package services

// ModelPricing holds input and output cost per 1M tokens
type ModelPricing struct {
	InputPerMillion  float64
	OutputPerMillion float64
}

// modelPricing maps model names to their pricing (per 1M tokens in USD)
// Source: https://platform.openai.com/docs/pricing (OpenAI)
// Source: https://www.anthropic.com/pricing (Anthropic)
var modelPricing = map[string]ModelPricing{
	// OpenAI GPT-5 family
	"gpt-5.2":     {InputPerMillion: 1.75, OutputPerMillion: 14.00},
	"gpt-5.1":     {InputPerMillion: 1.25, OutputPerMillion: 10.00},
	"gpt-5":       {InputPerMillion: 1.25, OutputPerMillion: 10.00},
	"gpt-5-mini":  {InputPerMillion: 0.25, OutputPerMillion: 2.00},
	"gpt-5-nano":  {InputPerMillion: 0.05, OutputPerMillion: 0.40},
	"gpt-5.2-pro": {InputPerMillion: 21.00, OutputPerMillion: 168.00},
	"gpt-5-pro":   {InputPerMillion: 15.00, OutputPerMillion: 120.00},

	// OpenAI GPT-4.1 family
	"gpt-4.1":      {InputPerMillion: 2.00, OutputPerMillion: 8.00},
	"gpt-4.1-mini": {InputPerMillion: 0.40, OutputPerMillion: 1.60},
	"gpt-4.1-nano": {InputPerMillion: 0.10, OutputPerMillion: 0.40},

	// OpenAI GPT-4o family
	"gpt-4o":      {InputPerMillion: 2.50, OutputPerMillion: 10.00},
	"gpt-4o-mini": {InputPerMillion: 0.15, OutputPerMillion: 0.60},

	// OpenAI o-series (reasoning models)
	"o1":      {InputPerMillion: 15.00, OutputPerMillion: 60.00},
	"o1-mini": {InputPerMillion: 1.10, OutputPerMillion: 4.40},
	"o1-pro":  {InputPerMillion: 150.00, OutputPerMillion: 600.00},
	"o3":      {InputPerMillion: 2.00, OutputPerMillion: 8.00},
	"o3-mini": {InputPerMillion: 1.10, OutputPerMillion: 4.40},
	"o3-pro":  {InputPerMillion: 20.00, OutputPerMillion: 80.00},
	"o4-mini": {InputPerMillion: 1.10, OutputPerMillion: 4.40},

	// Anthropic Claude Opus
	"claude-opus-4-5-20251101": {InputPerMillion: 5.00, OutputPerMillion: 25.00},

	// Anthropic Claude Sonnet
	"claude-sonnet-4-5-20250929": {InputPerMillion: 3.00, OutputPerMillion: 15.00},

	// Anthropic Claude Haiku
	"claude-haiku-4-5-20251001": {InputPerMillion: 1.00, OutputPerMillion: 5.00},
}

// CalculateCost returns estimated cost in USD for the given model and token counts
func CalculateCost(model string, inputTokens, outputTokens int32) float64 {
	pricing, ok := modelPricing[model]
	if !ok {
		return 0
	}

	inputCost := float64(inputTokens) * pricing.InputPerMillion / 1_000_000
	outputCost := float64(outputTokens) * pricing.OutputPerMillion / 1_000_000

	return inputCost + outputCost
}

// GetModelPricing returns the pricing for a model, or nil if not found
func GetModelPricing(model string) *ModelPricing {
	pricing, ok := modelPricing[model]
	if !ok {
		return nil
	}
	return &pricing
}
