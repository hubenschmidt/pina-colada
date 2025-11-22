# Browser Select Tool Specification

**Status**: Proposed
**Date**: 2025-11-17
**Author**: System

## Problem Statement

### Current Issue
The Playwright MCP integration requires multiple graph iterations to interact with HTML `<select>` dropdown elements, and **ultimately fails to select options correctly**. Observed behavior from 401k rollover demo:

**Iteration Log** (graph-log.txt):
- **Iteration 4** (line 34): `browser_type("Destination Account dropdown", "e21", "New IRA Account")` → **Failed** (cannot type into `<select>`)
- **Iteration 6** (line 46): `browser_click("Destination Account dropdown", "e21")` → Opens dropdown
- **Iteration 8** (line 58): `browser_click("New IRA Account option", "e21")` → **Failed silently** (clicked dropdown again, not option)
- **Iteration 13** (line 82): `browser_click("Submit Rollover Request", "e28")` → **Form validation failed** (no option selected)
- **Result**: Form never submitted, confirmation page never reached, screenshot shows incomplete form

**Critical Failure Mode**:
The snapshot provided the same ref (`e21`) for both the dropdown and its options, causing the agent to click the dropdown repeatedly instead of selecting an option. The form submission was blocked by validation, but the agent didn't understand why and took a screenshot of the failed state.

**Impact**:
- 3 iterations wasted attempting dropdown selection
- Form submission failure (validation prevented submit)
- ~6k tokens wasted on failed attempts
- ~10 seconds of latency for a failed operation
- LLM cannot complete task due to tool limitations
- User sees "stuck" behavior and incomplete demo

### Root Cause
**Primary Issue**: Playwright MCP snapshot returns identical refs for `<select>` and `<option>` elements, making proper selection impossible via clicks.

**Secondary Issue**: No tool directly supports `<select>` element interaction. Playwright's native API has `selectOption()` for this purpose, but it's not exposed through MCP.

**Current Tools**:
```python
browser_navigate(url)
browser_click(element, ref)
browser_type(element, ref, text, submit)
browser_snapshot()
browser_take_screenshot(filename)
browser_wait_for(time, text)
```

**Evidence from logs** (line 83):
```yaml
- <changed> combobox "Destination Account" [active] [ref=e21]:
  - option "Select destination..." [selected]  # Still on default!
  - option "New IRA Account"
  - option "Existing IRA"
  - option "New Employer 401(k)"
```
After 3 click attempts, the dropdown remains on "Select destination...", proving selection failed.

## Proposed Solution

Add a `browser_select` tool that wraps Playwright's `selectOption()` method for efficient dropdown interaction.

### Tool Signature
```python
@tool
async def browser_select(element: str, ref: str, value: str) -> str:
    """
    Select an option from a dropdown (<select> element).

    Args:
        element: Human-readable description of the select element
        ref: Exact element reference from page snapshot
        value: Option value or text to select

    Returns:
        Success message
    """
```

### Usage Example
**Before** (Failed after 3+ iterations):
1. `browser_type("Destination Account", "e21", "New IRA Account")` → Error: can't type
2. `browser_click("Destination Account dropdown", "e21")` → Opens dropdown
3. `browser_click("New IRA Account option", "e21")` → **Clicks dropdown again** (same ref!)
4. Form submission fails validation → Task cannot complete

**After** (1 iteration, succeeds):
1. `browser_select("Destination Account dropdown", "e21", "New IRA Account")` → Option selected, form valid

## Technical Design

### Implementation Location
File: `/home/hubenschmidt/pina-colada-co/modules/agent/src/agent/tools/mcp_playwright.py`

### Code Structure
```python
@tool
async def browser_select(element: str, ref: str, value: str) -> str:
    """
    Select an option from a dropdown (<select> element).

    Args:
        element: Human-readable description of the select element
        ref: Exact element reference from page snapshot
        value: Option value or text to select

    Returns:
        Success message
    """
    result = await _mcp_client.call_tool(
        "browser_select_option",
        {"element": element, "ref": ref, "value": value}
    )
    return str(result)
```

### Integration Points

#### 1. Update PLAYWRIGHT_MCP_TOOLS List
```python
PLAYWRIGHT_MCP_TOOLS = [
    browser_navigate,
    browser_click,
    browser_type,
    browser_select,  # NEW
    browser_snapshot,
    browser_take_screenshot,
    browser_wait_for,
]
```

#### 2. Update Scraper Prompt
File: `/home/hubenschmidt/pina-colada-co/modules/agent/src/agent/workers/scraper.py`

Add to tool list in `_build_scraper_prompt()`:
```python
def _build_scraper_prompt(state: Dict[str, Any]) -> str:
    return f"""You are a browser automation agent. Your job is to USE TOOLS, not describe them.

TASK: {state['success_criteria']}

YOU MUST CALL THESE TOOLS TO COMPLETE THE TASK:
- browser_navigate(url)
- browser_snapshot()
- browser_type(element, ref, text, submit)
- browser_click(element, ref)
- browser_select(element, ref, value)  # NEW: For dropdown/select elements
- browser_wait_for(time)
- browser_take_screenshot(filename)

DO NOT respond with text explanations. DO NOT say what you "will" do.
IMMEDIATELY CALL browser_navigate to start. Then call browser_snapshot.
Use refs from snapshot output for browser_click, browser_type, and browser_select.
Use browser_select for <select> dropdown elements instead of clicking.

Date: {datetime.now().strftime("%Y-%m-%d %H:%M:%S")}
"""
```

## MCP Server Compatibility

### Check if MCP Supports selectOption
The Playwright MCP server (`@playwright/mcp@latest`) may or may not expose a `browser_select_option` tool.

**Option A**: If supported
- Use the approach above directly
- MCP server handles the Playwright API call

**Option B**: If NOT supported (Fallback)
- Tool calls `browser_click` twice internally (click select → click option)
- Requires snapshot parsing to find option ref
- More complex but still reduces LLM iterations

### Verification Method
```python
# During init, list available tools
async def initialize(self):
    # ... existing initialization code ...

    # After session.initialize()
    tools_list = await self.session.list_tools()
    available_tools = [tool.name for tool in tools_list.tools]
    logger.info(f"Available MCP tools: {available_tools}")

    if "browser_select_option" not in available_tools:
        logger.warning("browser_select_option not available, using fallback")
```

## Implementation Steps

1. **Phase 1: Add Tool (mcp_playwright.py)**
   - Add `browser_select()` function with `@tool` decorator
   - Add to `PLAYWRIGHT_MCP_TOOLS` list
   - Test MCP server compatibility

2. **Phase 2: Update Scraper Prompt (scraper.py)**
   - Add `browser_select` to tool list in prompt
   - Add usage guidance for select elements

3. **Phase 3: Testing**
   - Run 401k rollover demo (Browser Automation)
   - Verify "Destination Account" dropdown uses `browser_select`
   - Confirm reduction from 3 iterations to 1

4. **Phase 4: Documentation**
   - Update scraper integration spec
   - Document dropdown handling best practices

## Testing Plan

### Test Case: 401k Rollover Demo
**Setup**:
- Navigate to `https://api.pinacolada.co/mocks/401k-rollover/`
- Login with demo/demo123
- Click "Initiate Rollover" for OldCorp 401(k)

**Expected Behavior**:
```
Iteration N: scraper calls browser_select("Destination Account", "e21", "New IRA Account")
Iteration N+1: scraper proceeds to confirmation checkbox
```

**Success Criteria**:
- Dropdown selection completes in 1 iteration (down from 3)
- No errors in MCP communication
- Correct option selected in form

### Verification Commands
```bash
# Check logs for browser_select calls
grep "browser_select" graph-log.txt

# Count iterations for dropdown interaction
grep -A5 "Destination Account" graph-log.txt | grep "iteration"
```

## Alternative Approaches

### Alternative 1: Improve Prompt Only
**Approach**: Update scraper prompt with explicit dropdown handling instructions
```
For <select> dropdowns:
1. Call browser_click on the select element
2. Call browser_snapshot to see options
3. Call browser_click on the desired option
```

**Pros**: No code changes
**Cons**: Still requires 3 iterations, LLM may not follow consistently

### Alternative 2: Create Composite Tool
**Approach**: `browser_select_and_confirm()` that does select + screenshot verification

**Pros**: Ensures selection worked
**Cons**: More complex, less composable

### Alternative 3: Use browser_type with Tab
**Approach**: Type option text + Tab to navigate select options

**Pros**: May work with existing tools
**Cons**: Unreliable, doesn't work on all browsers/dropdowns

## Recommended Approach
**Primary**: Implement `browser_select()` tool as specified above
**Fallback**: If MCP doesn't support selectOption, use Alternative 1 (prompt improvement) as temporary measure

## Success Metrics
- **Task completion**: Currently fails → Successfully completes
- **Dropdown interaction**: 3+ failed iterations → 1 successful iteration (100% reduction in waste)
- **Token usage**: ~6k tokens wasted → ~2k tokens (67% reduction)
- **Latency**: ~10 seconds failed → ~2 seconds success (80% improvement)
- **Form submission**: Validation failure → Success
- **Demo completion**: Stuck on form → Reaches confirmation page
- **LLM success rate**: 0% (cannot complete) → 95%+ first-try success

## References
- Playwright MCP: https://github.com/playwright/playwright/tree/main/packages/playwright-mcp
- Playwright selectOption API: https://playwright.dev/docs/api/class-locator#locator-select-option
- Current implementation: `/modules/agent/src/agent/tools/mcp_playwright.py`
- Scraper worker: `/modules/agent/src/agent/workers/scraper.py`
- Demo logs: `/graph-log.txt` (iterations 4, 6, 8)
