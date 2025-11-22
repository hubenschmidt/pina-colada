# Specialized Evaluators - Technical Specification

## Context

Previously, all workers (worker, job_hunter, cover_letter_writer, scraper) routed to a single evaluator. This created issues:

- Job search outputs evaluated with generic criteria
- Scraper results evaluated without understanding data extraction quality
- Cover letters evaluated without career-specific context
- Single evaluator required bloated system prompt to handle all cases

**Goal:** Keep context length lean for efficient token usage. Each specialized evaluator only includes criteria relevant to its task type, avoiding unnecessary prompt bloat.

**Solution:** Route workers to specialized evaluators based on task type via external YAML config.

---

## Architecture Overview

### High-Level Flow

```
Worker → Specialized Evaluator → [Retry Worker | END]

Routing:
- worker → general_evaluator
- job_hunter → career_evaluator
- cover_letter_writer → career_evaluator
- scraper → scraper_evaluator
```

### Configuration-Driven Routing

```yaml
# config/evaluator_config.yaml
evaluators:
  career:
    workers: [job_hunter, cover_letter_writer]
  scraper:
    workers: [scraper]
  general:
    workers: [worker]
```

---

## Component Specifications

### 1. Base Evaluator (`workers/evaluators/base_evaluator.py`)

Shared evaluation logic with customizable system prompts.

```python
async def create_base_evaluator_node(
    evaluator_name: str,
    build_system_prompt: Callable[[str], str],
):
    """
    Factory function for evaluator nodes.

    Args:
        evaluator_name: Name for logging
        build_system_prompt: Function(resume_context) -> system_prompt

    Returns:
        evaluator_node function
    """
```

**Features:**
- Retry loop detection and forced approval
- Resume context injection
- Structured output (feedback, success_criteria_met, user_input_needed)
- Langfuse tracing integration

---

### 2. Career Evaluator (`workers/evaluators/career_evaluator.py`)

**Used by:** job_hunter, cover_letter_writer

**Evaluation Criteria:**
- Job search results: Validates external postings, direct URLs
- Cover letters: Checks professional tone, job tailoring, experience highlighting
- Resume accuracy: Distinguishes between job postings and user's work history

**System Prompt Additions:**
```
CRITICAL DISTINCTION:
- JOB SEARCH RESULTS: External job postings are VALID even if not in resume
- RESUME DATA: User's work history must match resume exactly

CAREER-SPECIFIC CHECKS:
- Job links must be direct posting URLs
- Cover letters must be tailored to job
```

---

### 3. Scraper Evaluator (`workers/evaluators/scraper_evaluator.py`)

**Used by:** scraper

**Evaluation Criteria:**
- Data extraction accuracy and completeness
- Output format (JSON, CSV)
- Error handling
- Rate limiting compliance

**System Prompt Additions:**
```
SCRAPER-SPECIFIC CHECKS:
- Data extraction accuracy
- Completeness of required fields
- Usable output format
- Navigation/timeout handling

COMMON ISSUES:
- Missing or null fields
- Malformed values
- Truncated data
```

---

### 4. General Evaluator (`workers/evaluators/general_evaluator.py`)

**Used by:** worker

**Evaluation Criteria:**
- Factual accuracy
- Completeness of response
- Clarity and structure
- Relevance to request

---

### 5. Config Loader (`orchestrator.py`)

```python
def load_evaluator_config() -> Dict[str, Any]:
    """Load evaluator routing configuration from YAML"""
    config_path = Path(__file__).parent / "config" / "evaluator_config.yaml"
    with open(config_path) as f:
        return yaml.safe_load(f)

def get_evaluator_for_worker(config: Dict[str, Any], worker_name: str) -> str:
    """Get the evaluator type for a given worker"""
    for evaluator_type, evaluator_config in config.get("evaluators", {}).items():
        if worker_name in evaluator_config.get("workers", []):
            return evaluator_type
    return "general"
```

---

### 6. State Extension

```python
class State(TypedDict):
    # ... existing fields ...
    evaluator_type: Optional[str]  # career, scraper, or general
```

---

## Graph Wiring

### Node Addition

```python
graph_builder.add_node("career_evaluator", career_evaluator)
graph_builder.add_node("scraper_evaluator", scraper_evaluator)
graph_builder.add_node("general_evaluator", general_evaluator)
```

### Conditional Edges

```python
# Worker to evaluator routing
graph_builder.add_conditional_edges(
    "worker",
    route_from_worker,
    {"tools": "tools", "evaluator": "general_evaluator"},
)

graph_builder.add_conditional_edges(
    "job_hunter",
    route_from_job_hunter,
    {"tools": "tools", "evaluator": "career_evaluator"},
)

graph_builder.add_conditional_edges(
    "scraper",
    route_from_scraper,
    {"tools": "tools", "evaluator": "scraper_evaluator"},
)

graph_builder.add_edge("cover_letter_writer", "career_evaluator")

# All evaluators share same return routing
evaluator_routing = {
    "worker": "worker",
    "job_hunter": "job_hunter",
    "scraper": "scraper",
    "cover_letter_writer": "cover_letter_writer",
    "END": END,
}

graph_builder.add_conditional_edges("career_evaluator", route_from_evaluator, evaluator_routing)
graph_builder.add_conditional_edges("scraper_evaluator", route_from_evaluator, evaluator_routing)
graph_builder.add_conditional_edges("general_evaluator", route_from_evaluator, evaluator_routing)
```

---

## File Structure

```
modules/agent/src/agent/
├── config/
│   └── evaluator_config.yaml      # Worker-to-evaluator mapping
├── workers/
│   ├── evaluators/
│   │   ├── __init__.py            # Package exports
│   │   ├── base_evaluator.py      # Shared logic
│   │   ├── career_evaluator.py    # Job/cover letter evaluation
│   │   ├── scraper_evaluator.py   # Web scraping evaluation
│   │   └── general_evaluator.py   # Default evaluation
│   ├── worker.py
│   ├── job_hunter.py
│   ├── cover_letter_writer.py
│   ├── scraper.py
│   └── evaluator.py               # DEPRECATED - kept for reference
└── orchestrator.py                # Updated with multi-evaluator wiring
```

---

## Benefits

1. **Task-specific evaluation**: Each worker type gets criteria appropriate to its output
2. **Configurable routing**: YAML config allows easy changes without code edits
3. **Shared infrastructure**: Base evaluator reduces code duplication
4. **Extensible**: Add new evaluator types by creating new file + config entry
5. **Better feedback loops**: Workers receive domain-specific improvement suggestions

---

## Future Enhancements

- **Dynamic evaluator selection**: Use LLM to choose evaluator based on task content
- **Evaluator metrics**: Track approval rates per evaluator type
- **Custom retry limits**: Different retry thresholds per evaluator
- **Evaluator chaining**: Multiple evaluators for complex outputs
