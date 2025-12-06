"""
Integration tests for job search workflow via WebSocket.

Tests the full agent flow:
1. Look up individual record
2. List linked documents
3. Conduct job search based on resume

Run with: pytest src/agent/workers/career/__tests__/integration/ -vs
"""

import pytest
import re

from agent.__tests__.conftest import send_query


def log_response(query: str, response: str, max_len: int = 500) -> None:
    """Print query/response for debugging with -s flag."""
    print(f"\n{'='*60}")
    print(f"Q: {query}")
    print(f"A: {response[:max_len]}{'...' if len(response) > max_len else ''}")
    print(f"{'='*60}")


def extract_urls(text: str) -> list[str]:
    """Extract URLs from text."""
    url_pattern = r'https?://[^\s<>"{}|\\^`\[\]]+'
    return re.findall(url_pattern, text)


class TestJobSearchWorkflow:
    """Integration tests for the full job search workflow."""

    @pytest.mark.asyncio
    async def test_lookup_individual_without_entity_keyword(self, ws_connection):
        """Agent should find Individual record even without saying 'individual'."""
        query = "look up William Hubenschmidt"
        response = await send_query(ws_connection, query)
        log_response(query, response)

        assert len(response) > 20, "Response too short"
        assert "william" in response.lower(), "Should mention William"
        assert "hubenschmidt" in response.lower(), "Should mention Hubenschmidt"
        assert "error" not in response.lower(), f"Got error: {response}"

    @pytest.mark.asyncio
    async def test_list_documents_for_individual(self, ws_connection):
        """Agent should list documents linked to William's individual record."""
        # First establish context
        await send_query(ws_connection, "look up William Hubenschmidt")

        # Now ask for documents
        query = "list the documents linked to his record"
        response = await send_query(ws_connection, query)
        log_response(query, response)

        assert len(response) > 50, "Response too short for document list"
        assert "resume" in response.lower(), "Should mention resume"
        assert "error" not in response.lower(), f"Got error: {response}"

    @pytest.mark.asyncio
    async def test_job_search_with_resume_context(self, ws_connection):
        """Agent should search for jobs based on resume in NYC startups."""
        # Establish context
        await send_query(ws_connection, "look up William Hubenschmidt")
        await send_query(ws_connection, "list documents linked to his record")

        # Job search
        query = "search for software engineering jobs matching his resume at NYC startups"
        response = await send_query(ws_connection, query)
        log_response(query, response, max_len=1000)

        assert len(response) > 100, "Response too short"
        assert "error" not in response.lower(), f"Got error: {response}"

        # Should have job-related content
        job_keywords = ["job", "position", "role", "engineer", "software", "startup"]
        has_job_content = any(kw in response.lower() for kw in job_keywords)
        assert has_job_content, f"No job-related content: {response[:200]}"

        # Check for URLs
        urls = extract_urls(response)
        print(f"\nFound {len(urls)} URLs in response")
        for url in urls[:5]:
            print(f"  - {url}")


class TestJobSearchDirect:
    """Direct job search tests without resume context."""

    @pytest.mark.asyncio
    async def test_job_search_returns_urls(self, ws_connection):
        """Job search should return actual job postings with URLs."""
        query = "search for senior software engineer jobs at startups in NYC"
        response = await send_query(ws_connection, query)
        log_response(query, response, max_len=800)

        assert len(response) > 50, "Response too short"
        assert "error" not in response.lower(), f"Got error: {response}"

        # Check for URLs
        urls = extract_urls(response)
        print(f"\nFound {len(urls)} URLs")
        assert len(urls) > 0, "Should return at least one URL"


class TestDocumentRetrieval:
    """Tests for document retrieval flow."""

    @pytest.mark.asyncio
    async def test_get_resume_content(self, ws_connection):
        """Agent should retrieve resume content with proper context."""
        # Establish context
        await send_query(ws_connection, "look up William Hubenschmidt")
        await send_query(ws_connection, "what documents are linked to his record?")

        # Get resume content
        query = "show me the contents of the resume"
        response = await send_query(ws_connection, query)
        log_response(query, response, max_len=800)

        assert len(response) > 100, "Response too short"
        resume_keywords = ["experience", "education", "skill", "engineer", "software", "work"]
        has_resume_content = any(kw in response.lower() for kw in resume_keywords)
        assert has_resume_content, f"No resume content: {response[:300]}"

    @pytest.mark.asyncio
    async def test_documents_linked_to_individual(self, ws_connection):
        """Agent should find documents linked to individual."""
        query = "what documents are linked to William Hubenschmidt?"
        response = await send_query(ws_connection, query)
        log_response(query, response)

        assert len(response) > 50, "Response too short"
        assert "resume" in response.lower(), "Should find resume"
        assert "error" not in response.lower(), f"Got error: {response}"
