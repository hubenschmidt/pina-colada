"""Model pricing constants and cost calculation utilities.

Note: OpenAI/Anthropic don't provide pricing APIs, so this is best-effort.
Unknown models return None for cost, displayed as "N/A" in the UI.
"""

# Pricing per 1M tokens (USD) - update manually when providers change prices
# Source: https://openai.com/api/pricing/ and https://www.anthropic.com/pricing
MODEL_PRICING = {
    # OpenAI models
    "gpt-4o": {"input": 2.50, "output": 10.00},
    "gpt-4o-mini": {"input": 0.15, "output": 0.60},
    "gpt-4-turbo": {"input": 10.00, "output": 30.00},
    "gpt-4.1": {"input": 2.00, "output": 8.00},
    "gpt-5.1": {"input": 5.00, "output": 20.00},
    "gpt-5-mini": {"input": 0.30, "output": 1.20},
    "gpt-5-nano": {"input": 0.10, "output": 0.40},
    "gpt-5-pro": {"input": 10.00, "output": 40.00},
    # Anthropic models
    "claude-3-5-sonnet-20241022": {"input": 3.00, "output": 15.00},
    "claude-3-5-haiku-20241022": {"input": 0.80, "output": 4.00},
    "claude-haiku-4-5-20251001": {"input": 0.80, "output": 4.00},
    "claude-3-opus-20240229": {"input": 15.00, "output": 75.00},
}

# Cost tier thresholds (USD)
COST_TIER_LOW = 0.01
COST_TIER_MODERATE = 0.10


def calculate_cost(model_name: str, input_tokens: int, output_tokens: int) -> float | None:
    """Calculate cost in USD for given token usage. Returns None if model pricing unknown."""
    pricing = MODEL_PRICING.get(model_name)
    if pricing is None:
        return None
    return (input_tokens * pricing["input"] + output_tokens * pricing["output"]) / 1_000_000


def get_cost_tier(cost: float | None) -> str:
    """Return 'low', 'moderate', 'high', or 'unknown' based on cost."""
    if cost is None:
        return "unknown"
    if cost < COST_TIER_LOW:
        return "low"
    if cost < COST_TIER_MODERATE:
        return "moderate"
    return "high"


def get_model_pricing(model_name: str) -> dict | None:
    """Get pricing info for a model, or None if unknown."""
    return MODEL_PRICING.get(model_name)
