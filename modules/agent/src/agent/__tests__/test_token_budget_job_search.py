"""
Token budget tests for job search workflow.
Target: <= 10k tokens for full workflow (stretch goal: 5k).

Run with: pytest src/agent/__tests__/test_token_budget_job_search.py -vs
"""

import pytest
import re
from agent.__tests__.conftest import send_query, QueryResult


# Token budget constants
BUDGET_CRM_LOOKUP = 1000
BUDGET_DOC_LIST = 1000
BUDGET_JOB_SEARCH_5 = 6000
BUDGET_TOTAL_WORKFLOW = 10000

# Job board domains to exclude
EXCLUDED_DOMAINS = ["linkedin.com", "indeed.com", "glassdoor.com", "ziprecruiter.com"]


def log_tokens(step: str, result: QueryResult) -> None:
    """Log step name and token usage."""
    print(f"\n{step}: {result.total_tokens} tokens ({result.input_tokens} in, {result.output_tokens} out)")
    print(f"   Response preview: {result.response[:200]}...")


def extract_urls(text: str) -> list[str]:
    """Extract URLs from response text."""
    return re.findall(r"https?://[^\s<>\"'{}|\\^`\[\]]+", text)


def filter_direct_urls(urls: list[str]) -> list[str]:
    """Filter out job board URLs, keep direct company links."""
    return [u for u in urls if not any(d in u.lower() for d in EXCLUDED_DOMAINS)]


class TestTokenBudgetJobSearch:
    """Token budget tests - target <=10k for full workflow."""

    @pytest.mark.asyncio
    async def test_crm_lookup_token_budget(self, ws_connection):
        """CRM lookup should use <1000 tokens."""
        result = await send_query(ws_connection, "look up William Hubenschmidt Individual record")
        log_tokens("CRM lookup", result)

        assert result.total_tokens < BUDGET_CRM_LOOKUP, f"CRM lookup: {result.total_tokens} > {BUDGET_CRM_LOOKUP}"
        assert "hubenschmidt" in result.response.lower(), "Should find William Hubenschmidt"

    @pytest.mark.asyncio
    async def test_doc_list_token_budget(self, ws_connection):
        """Document list should use <1000 tokens."""
        result = await send_query(ws_connection, "list documents for Individual 1")
        log_tokens("Doc list", result)

        assert result.total_tokens < BUDGET_DOC_LIST, f"Doc list: {result.total_tokens} > {BUDGET_DOC_LIST}"
        assert "resume" in result.response.lower() or "document" in result.response.lower()

    @pytest.mark.asyncio
    async def test_job_search_5_results_token_budget(self, ws_connection):
        """Job search with 5 results should use <6000 tokens."""
        result = await send_query(
            ws_connection,
            "search for 5 Senior Software Engineer jobs at NYC startups - direct company career page links only, no job boards",
        )
        log_tokens("Job search (5 results)", result)

        assert result.total_tokens < BUDGET_JOB_SEARCH_5, f"Job search: {result.total_tokens} > {BUDGET_JOB_SEARCH_5}"

        urls = extract_urls(result.response)
        direct_urls = filter_direct_urls(urls)
        assert len(direct_urls) >= 3, f"Need >=3 direct URLs, got {len(direct_urls)}: {direct_urls}"

    @pytest.mark.asyncio
    async def test_full_workflow_token_budget(self, ws_connection):
        """Full 3-step workflow should use <=10k tokens total."""
        total_tokens = 0

        # Step 1: CRM lookup
        r1 = await send_query(ws_connection, "look up William Hubenschmidt Individual record")
        log_tokens("Step 1 - CRM lookup", r1)
        total_tokens += r1.total_tokens
        assert "hubenschmidt" in r1.response.lower()

        # Step 2: List documents for that Individual
        r2 = await send_query(ws_connection, "list documents for Individual 1")
        log_tokens("Step 2 - Doc list", r2)
        total_tokens += r2.total_tokens

        # Step 3: Job search with 5 results
        r3 = await send_query(
            ws_connection,
            "search for 5 Senior Software Engineer jobs at NYC startups matching the resume - direct links only",
        )
        log_tokens("Step 3 - Job search", r3)
        total_tokens += r3.total_tokens

        urls = extract_urls(r3.response)
        direct_urls = filter_direct_urls(urls)

        print(f"\n{'='*50}")
        print(f"TOTAL: {total_tokens} tokens (budget: {BUDGET_TOTAL_WORKFLOW})")
        print(f"Direct URLs found: {len(direct_urls)}")
        print(f"{'='*50}")

        assert total_tokens <= BUDGET_TOTAL_WORKFLOW, f"Total {total_tokens} > {BUDGET_TOTAL_WORKFLOW}"
        assert len(direct_urls) >= 3, f"Need >=3 direct URLs, got {len(direct_urls)}"
