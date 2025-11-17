# Scraper Integration - Technical Specification

## Context

**Interview Prep for:** Integrations Engineer @ Capitalize (https://www.builtinnyc.com/job/integrations-engineer/7310182)
**Interview Date:** Monday
**Goal:** Build proof-of-concept demo emulating Capitalize's 401k rollover automation using LangGraph worker architecture

### Capitalize's Problem Space

Capitalize automates 401k rollovers by:
1. Embedding their app in iframe within financial advisor/bank portals
2. Making REST API calls to 401k platforms (when APIs exist)
3. Using web scraping/browser automation for platforms without APIs
4. Handling multi-step forms, authentication, and data extraction

**Next-gen requirement:** Intelligent web scraper that goes beyond traditional REST integrations

---

## Architecture Overview

### High-Level Flow

```
User Request → Router → ScraperWorker → Scraper Tools → Evaluator → Response
                              ↓
                    LLM Decision Engine
                    (static vs browser)
                              ↓
                    ┌─────────┴─────────┐
                    ↓                   ↓
            BeautifulSoup          Playwright
            (static HTML)      (browser automation)
```

### Integration with Existing System

**Current Pattern:** Orchestrator → Router → [Worker|JobHunter|CoverLetterWriter] → Tools → Evaluator

**New Addition:** ScraperWorker joins as 4th specialized worker type

```
START
  ↓
Router (agent_router.py)
  ├─ Job search keywords → JobHunter
  ├─ Cover letter keywords → CoverLetterWriter
  ├─ Scrape/automate keywords → ScraperWorker ← NEW
  └─ General → Worker
  ↓
ScraperWorker
  ├─ Analyzes target URL/page structure
  ├─ LLM decides: static scraping vs browser automation
  ├─ Executes scraping strategy
  ├─ Extracts/submits data
  └─ Returns results
  ↓
Evaluator
  ├─ Validates scraping success
  ├─ Checks if data extraction complete
  └─ Routes: retry or END
```

---

## Component Specifications

### 1. ScraperWorker (`modules/agent/src/agent/workers/scraper.py`)

**Factory Function Pattern:**

```python
def create_scraper_node(websocket_sender=None):
    """
    Creates scraper worker node following existing JobHunter pattern.

    Closes over:
    - LLM instance (GPT-5-chat-latest, temp=0.7)
    - Scraper tools
    - WebSocket sender for streaming
    """

    llm = ChatOpenAI(model="gpt-5-chat-latest", temperature=0.7)
    llm_with_tools = llm.bind_tools(SCRAPER_TOOLS)

    system_prompt = """You are a web scraping specialist.
    Analyze pages and determine optimal scraping strategy:
    - Static HTML → use scrape_static_page
    - JavaScript-heavy → use automate_browser
    - Forms → use fill_form with intelligent field detection

    For 401k rollover simulation:
    1. Navigate multi-step flows
    2. Extract account information
    3. Fill rollover forms
    4. Confirm submission
    """

    async def scraper_node(state: State) -> dict:
        # Trim messages, build prompt, call LLM
        # Route to tools or return results
        # Stream via WebSocket
        pass

    return scraper_node
```

**State Management:**

Extends existing `State` TypedDict:
- `messages`: Conversation history
- `scraping_context`: Dict with URL, page type, extracted data
- `scraping_strategy`: "static" | "browser" | "hybrid"
- `form_data`: Data to submit in forms
- `success_criteria`: "data extracted" | "form submitted" | "multi-step complete"

---

### 2. Scraper Tools (`modules/agent/src/agent/tools/scraper_tools.py`)

#### Tool 1: `analyze_page_structure`

```python
@tool
async def analyze_page_structure(url: str) -> dict:
    """
    LLM analyzes page HTML to determine scraping strategy.

    Args:
        url: Target URL to analyze

    Returns:
        {
            "page_type": "static" | "dynamic" | "form",
            "requires_js": bool,
            "form_fields": [...],
            "recommended_strategy": "static" | "browser"
        }
    """
    # Fetch page, get HTML structure
    # Use LLM to analyze (GPT-4o mini for speed)
    # Return recommendation
```

#### Tool 2: `scrape_static_page`

```python
@tool
async def scrape_static_page(url: str, selectors: dict) -> dict:
    """
    Traditional scraping with requests + BeautifulSoup.

    Args:
        url: Target URL
        selectors: CSS/XPath selectors for data extraction

    Returns:
        {
            "success": bool,
            "data": {...extracted data...},
            "error": str | None
        }
    """
    # requests.get(url)
    # BeautifulSoup parsing
    # CSS selector extraction
```

#### Tool 3: `automate_browser`

```python
@tool
async def automate_browser(
    url: str,
    actions: list[dict],
    headless: bool = True
) -> dict:
    """
    Playwright browser automation for JS-heavy sites.

    Args:
        url: Starting URL
        actions: List of actions [
            {"type": "click", "selector": "..."},
            {"type": "fill", "selector": "...", "value": "..."},
            {"type": "wait", "selector": "..."},
            {"type": "extract", "selector": "..."}
        ]

    Returns:
        {
            "success": bool,
            "data": {...extracted data...},
            "screenshots": [...paths...],
            "error": str | None
        }
    """
    # Launch Playwright browser
    # Execute action sequence
    # Handle waits, navigation
    # Take screenshots for debugging
```

#### Tool 4: `fill_form`

```python
@tool
async def fill_form(
    url: str,
    form_data: dict,
    multi_step: bool = False
) -> dict:
    """
    Intelligent form filling with field detection.

    Args:
        url: Form URL
        form_data: {field_name: value, ...}
        multi_step: Whether to handle multi-page forms

    Returns:
        {
            "success": bool,
            "submitted": bool,
            "confirmation_data": {...},
            "error": str | None
        }
    """
    # Detect form fields (name, id, type)
    # Map form_data to fields intelligently
    # Fill and submit
    # Handle multi-step navigation
```

---

### 3. Mock 401k Endpoint (`modules/agent/src/api/routes/mock/401k_rollover.py`)

**Purpose:** Simulated 401k provider for demo

**Endpoints:**

```python
# GET /api/mock/401k-rollover
# Returns HTML login page

# POST /api/mock/401k-rollover/login
# Body: {username, password}
# Returns: session token + redirect to accounts page

# GET /api/mock/401k-rollover/accounts
# Headers: Authorization: Bearer <token>
# Returns: HTML with account list

# POST /api/mock/401k-rollover/accounts/{id}/rollover
# Body: {destination_account, amount, confirm}
# Returns: confirmation page

# GET /api/mock/401k-rollover/confirmation/{id}
# Returns: rollover confirmation details
```

**Multi-Step Flow:**

1. Login page (form with username/password)
2. Account selection (list of 401k accounts)
3. Rollover form (destination, amount, checkboxes)
4. Confirmation page (success message + details)

**Implementation:**

```python
from fastapi import APIRouter, HTTPException
from fastapi.responses import HTMLResponse

router = APIRouter(prefix="/api/mock/401k-rollover")

# In-memory session store for demo
sessions = {}
accounts_db = {
    "user123": [
        {"id": "401k-1", "balance": 50000, "provider": "OldCorp 401k"},
        {"id": "401k-2", "balance": 25000, "provider": "StartupCo 401k"}
    ]
}

@router.get("/", response_class=HTMLResponse)
async def login_page():
    return """<html>...</html>"""  # Login form

@router.post("/login")
async def login(username: str, password: str):
    # Simplified auth
    # Return session token

@router.get("/accounts", response_class=HTMLResponse)
async def accounts_page(authorization: str):
    # Validate token
    # Return HTML with account cards

# ... remaining endpoints
```

---

### 4. Router Update (`modules/agent/src/agent/routers/agent_router.py`)

**Add Scraper Detection:**

```python
def route_to_agent(state: State) -> Literal["worker", "job_hunter", "cover_letter_writer", "scraper"]:
    """Route to appropriate worker based on message content."""

    messages = state.get("messages", [])
    if not messages:
        return "worker"

    last_message = messages[-1].content.lower()

    # Existing routing...
    if any(kw in last_message for kw in ["job", "search", "apply"]):
        return "job_hunter"
    if any(kw in last_message for kw in ["cover letter", "application"]):
        return "cover_letter_writer"

    # NEW: Scraper routing
    if any(kw in last_message for kw in [
        "scrape", "automate", "extract data", "fill form",
        "401k", "rollover", "browser automation"
    ]):
        return "scraper"

    return "worker"
```

---

### 5. Frontend Demo UI

**New Page:** `/demo/scraper`

**Layout:**

```
┌─────────────────────────────────────────────┐
│  Scraper Demo - 401k Rollover Automation   │
├──────────────────┬──────────────────────────┤
│                  │                          │
│  Mock 401k Site  │   Chat Interface         │
│  (iframe)        │   "Automate rollover     │
│                  │    for account 401k-1"   │
│  [Login Form]    │                          │
│                  │   ScraperWorker:         │
│                  │   → Analyzing page...    │
│                  │   → Using browser auto   │
│                  │   → Filling form...      │
│                  │   ✓ Rollover complete    │
│                  │                          │
└──────────────────┴──────────────────────────┘
```

**Features:**

- Split view: iframe with mock site + chat interface
- Real-time scraper actions displayed in chat
- Ability to trigger automation via natural language
- Visual feedback (screenshots from Playwright)

---

## Implementation Checklist

### Phase 1: Foundation (Day 1)

- [ ] Install dependencies (`playwright`, `beautifulsoup4`, `lxml`)
- [ ] Run `playwright install chromium`
- [ ] Create `modules/agent/src/api/routes/mock/` directory
- [ ] Implement mock 401k endpoint with all 4 pages
- [ ] Test mock endpoint manually in browser

### Phase 2: Tools (Day 2)

- [ ] Create `tools/scraper_tools.py`
- [ ] Implement `analyze_page_structure` (use GPT-4o mini)
- [ ] Implement `scrape_static_page` (requests + BS4)
- [ ] Implement `automate_browser` (Playwright)
- [ ] Implement `fill_form` (Playwright + intelligent field detection)
- [ ] Unit test each tool against mock endpoint

### Phase 3: Worker (Day 3)

- [ ] Create `workers/scraper.py`
- [ ] Implement `create_scraper_node` factory function
- [ ] Add system prompt for scraping strategy
- [ ] Integrate tools with worker LLM
- [ ] Test worker in isolation

### Phase 4: Integration (Day 4)

- [ ] Update `routers/agent_router.py` with scraper routing
- [ ] Add scraper node to orchestrator graph
- [ ] Update evaluator to handle scraper success criteria
- [ ] End-to-end test via WebSocket API

### Phase 5: Frontend Demo (Day 5)

- [ ] Create demo page component
- [ ] Add iframe for mock 401k site
- [ ] Connect chat interface to scraper worker
- [ ] Add visual feedback for scraper actions
- [ ] Polish UI for interview demo

### Phase 6: Interview Prep (Day 6-7)

- [ ] Document architecture decisions
- [ ] Prepare talking points about Capitalize's challenges
- [ ] Practice demo flow
- [ ] Prepare questions about their integrations stack

---

## Interview Talking Points

### Architecture Decisions

**Why LangGraph?**
- State management for multi-step flows
- Natural checkpointing/resume capability
- LLM-powered decision making (static vs browser)

**Why Hybrid Scraping?**
- Static scraping is faster, lower resource usage
- Browser automation handles JS rendering, complex interactions
- LLM analyzes page and chooses optimal strategy

**Scalability Considerations:**
- Worker pool pattern for parallel scraping jobs
- Playwright browser instances are resource-intensive → need orchestration
- Caching strategies for repeated scraping of same sites

**Error Handling:**
- Evaluator provides feedback loop for retries
- Screenshot capture for debugging failed automations
- Graceful degradation: browser automation → static → manual intervention

### Questions for Capitalize Team

1. How do you handle rate limiting across different 401k providers?
2. What's your strategy for detecting page structure changes?
3. Do you use headless browsers in production? How do you manage resource usage?
4. How do you handle CAPTCHA or 2FA in automation flows?
5. What's your approach to maintaining integrations as provider sites change?
6. How do you balance automation vs human-in-the-loop for edge cases?

---

## Dependencies

```txt
# Add to modules/agent/requirements.txt
playwright==1.41.0
beautifulsoup4==4.12.3
lxml==5.1.0
requests==2.31.0
```

```bash
# Post-install
playwright install chromium
```

---

## Success Metrics for Demo

- [ ] Mock 401k site renders correctly
- [ ] ScraperWorker correctly routes to appropriate tool
- [ ] Successfully automates full rollover flow (login → select → submit)
- [ ] Chat interface shows real-time scraping progress
- [ ] Evaluator validates success and provides feedback
- [ ] Can demo end-to-end in < 3 minutes

---

## Timeline

**Today (Day 1):** Mock endpoint + dependencies
**Tomorrow (Day 2-3):** Tools + worker implementation
**Weekend (Day 4-5):** Integration + frontend
**Sunday (Day 6):** Polish + interview prep
**Monday:** Interview 🚀
