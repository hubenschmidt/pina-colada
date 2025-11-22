# Specialized Evaluators - Technical Specification

## Context

Previously, all workers (worker, job_search, cover_letter_writer, scraper) routed to a single evaluator. This created issues:

- Job search outputs evaluated with generic criteria
- Scraper results evaluated without understanding data extraction quality
- Cover letters evaluated without career-specific context
- Single evaluator required bloated system prompt to handle all cases

**Goal:** Keep context length lean for efficient token usage. Each specialized evaluator only includes criteria relevant to its task type, avoiding unnecessary prompt bloat.

**Solution:** Route workers to specialized evaluators based on task type via graph edges, with prompts centralized in code.

---

## Architecture Overview

### High-Level Flow

```
Worker → Specialized Evaluator → [Retry Worker | END]

Routing:
- worker → general_evaluator
- job_search → career_evaluator
- cover_letter_writer → career_evaluator
- scraper → scraper_evaluator
```

### Centralized Prompts in Code

**Design Principle:** Core agent behavior stays in code (protected IP).

This approach:
- Protects core application logic from API exposure
- Provides single source of truth for all prompts
- Easy to review and update in one place

#### Prompt Directory Structure

Base system prompts are centralized in `src/agent/prompts/` directory. This provides a single source of truth for all prompts while keeping them protected from API exposure.

```
src/agent/prompts/
├── __init__.py
├── evaluator_prompts.py    # career, scraper, general evaluator prompts
├── worker_prompts.py       # job_search, scraper, worker, cover_letter prompts
└── orchestrator_prompts.py # orchestrator/router prompt
```

**Example: evaluator_prompts.py**
```python
# src/agent/prompts/evaluator_prompts.py

# --- Career Evaluator Prompt Sections ---

CAREER_BASE = """
You evaluate whether the Assistant's response meets career-related success criteria.

Respond with JSON: feedback, success_criteria_met, user_input_needed.
Be slightly lenient—approve unless clearly inadequate.
""".strip()

CAREER_CRITICAL_DISTINCTION = """
CRITICAL DISTINCTION:
- JOB SEARCH RESULTS: External job postings are VALID even if not in resume
- RESUME DATA: User's work history must match resume exactly
""".strip()

CAREER_CHECKS = """
CAREER-SPECIFIC CHECKS:
- Job links must be direct posting URLs
- Cover letters must be tailored to job
""".strip()


def build_career_evaluator_prompt(resume_context: str) -> str:
    """Build career evaluator system prompt."""
    sections = [CAREER_BASE]

    if resume_context:
        sections.append(
            "RESUME_CONTEXT (read-only factual background):\n"
            "```resume\n"
            f"{resume_context.strip()}\n"
            "```"
        )
        sections.append(CAREER_CRITICAL_DISTINCTION)

    sections.append(CAREER_CHECKS)

    return "\n\n".join(sections)


# --- Scraper Evaluator ---

SCRAPER_BASE = """
You evaluate whether the Assistant's web scraping response meets the success criteria.
...
""".strip()

def build_scraper_evaluator_prompt() -> str:
    """Build scraper evaluator system prompt."""
    # ... similar pattern ...


# --- General Evaluator ---

def build_general_evaluator_prompt(resume_context: str) -> str:
    """Build general evaluator system prompt."""
    # ... similar pattern ...
```

**Usage in evaluators:**
```python
# In career_evaluator.py - now just imports and uses
from agent.prompts.evaluator_prompts import build_career_evaluator_prompt

def create_career_evaluator_node():
    # ... setup code ...

    def evaluator_node(state):
        resume_context = state.get("resume_context", "")
        system_message = build_career_evaluator_prompt(resume_context)
        # ... rest of evaluation logic ...
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

**Used by:** job_search, cover_letter_writer

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

### 5. State Extension

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
    "job_search",
    route_from_job_search,
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
    "job_search": "job_search",
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
├── prompts/                           # Centralized prompt definitions
│   ├── __init__.py
│   ├── evaluator_prompts.py           # career, scraper, general evaluator prompts
│   ├── worker_prompts.py              # job_search, scraper, worker, cover_letter prompts
│   └── orchestrator_prompts.py        # orchestrator/router prompt
├── evaluators/                        # Specialized evaluators
│   ├── __init__.py                    # Package exports
│   ├── base_evaluator.py              # Shared logic
│   ├── career_evaluator.py            # Evaluator logic (imports from prompts/)
│   ├── scraper_evaluator.py           # Evaluator logic (imports from prompts/)
│   └── general_evaluator.py           # Evaluator logic (imports from prompts/)
├── workers/
│   ├── worker.py
│   ├── job_search.py
│   ├── cover_letter_writer.py
│   ├── scraper.py
│   └── evaluator.py                   # DEPRECATED - kept for reference
└── orchestrator.py                    # Multi-evaluator wiring
```

---

## Benefits

1. **Task-specific evaluation**: Each worker type gets criteria appropriate to its output
2. **Centralized prompts**: Single source of truth in `agent/prompts/` directory
3. **Shared infrastructure**: Base evaluator reduces code duplication
4. **Extensible**: Add new evaluator types by creating new file
5. **Better feedback loops**: Workers receive domain-specific improvement suggestions
6. **Maintainable**: Easy to review and update all prompts in one place

---

## Future Enhancements

- **Dynamic evaluator selection**: Use LLM to choose evaluator based on task content
- **Evaluator metrics**: Track approval rates per evaluator type
- **Custom retry limits**: Different retry thresholds per evaluator
- **Evaluator chaining**: Multiple evaluators for complex outputs

---

## Migration/Cleanup

When implementing this spec, perform the following changes:

### Files to Delete
- `modules/agent/src/agent/config/evaluator_config.yaml` - Redundant; routing defined by graph edges
- `modules/agent/src/agent/config/default_system_prompts.yml` - Not needed; core prompts stay in code

### Code to Remove from `orchestrator.py`
- `load_evaluator_config()` function - Not used
- `get_evaluator_for_worker()` function - Not used
- `load_default_prompts()` function - Not needed
- `get_system_prompt()` function - Not needed

### New Directory to Create
- `modules/agent/src/agent/prompts/` - Centralized prompt definitions
  - `__init__.py`
  - `evaluator_prompts.py` - Extract prompts from career/scraper/general_evaluator.py
  - `worker_prompts.py` - Extract prompts from worker files
  - `orchestrator_prompts.py` - Extract prompt from orchestrator/router

### Files to Delete (from previous implementation attempt)
- `models/NodeConfig.py` - Not needed for this spec
- `repositories/node_config_repository.py` - Not needed for this spec
- `controllers/node_config_controller.py` - Not needed for this spec
- `api/routes/node_configs.py` - Not needed for this spec
- Revert changes to `models/__init__.py`, `api/routes/__init__.py`, `server.py`

### Evaluator Refactoring
- Remove `_build_*_system_prompt()` functions from evaluator files
- Import prompt builders from `agent.prompts.evaluator_prompts`

---

## References

- [LLM Prompt Management in 2025: A Practical Playbook - DEV Community](https://dev.to/kamya_shah_3f4a20d6f64092/llm-prompt-management-in-2025-a-practical-playbook-for-scale-quality-and-speed-59bi)
- [Prompt Engineering & Management in Production - ZenML](https://www.zenml.io/blog/prompt-engineering-management-in-production-practical-lessons-from-the-llmops-database)
- [Best Tools for Creating System Prompts - PromptLayer](https://blog.promptlayer.com/the-best-tools-for-creating-system-prompts/)
