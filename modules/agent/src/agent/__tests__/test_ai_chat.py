"""
Integration tests for AI chat via WebSocket.

Tests verify the agent can:
- Look up CRM entities (individuals, organizations, accounts)
- Search and retrieve documents
- Handle multi-word search queries

Run with: pytest src/agent/__tests__/test_ai_chat.py -vs
"""

import pytest

from agent.__tests__.conftest import send_query


def log_response(query: str, response: str) -> None:
    """Print query/response for debugging with -s flag."""
    print(f"\n{'='*60}")
    print(f"Q: {query}")
    print(f"A: {response[:300]}{'...' if len(response) > 300 else ''}")
    print(f"{'='*60}")


class TestCRMLookups:
    """Tests for CRM entity lookups."""

    @pytest.mark.asyncio
    async def test_lookup_individual(self, ws_connection):
        """Agent should find William Hubenschmidt's individual record."""
        query = "look up William Hubenschmidt individual record"
        response = await send_query(ws_connection, query)
        log_response(query, response)

        assert len(response) > 20, "Response too short"
        assert "william" in response.lower(), "Should mention William"
        assert "hubenschmidt" in response.lower(), "Should mention Hubenschmidt"
        assert "error" not in response.lower(), f"Got error: {response}"
        assert "failed" not in response.lower(), f"Got failure: {response}"

    @pytest.mark.asyncio
    async def test_lookup_organization(self, ws_connection):
        """Agent should find PinaColada organization."""
        query = "find information about PinaColada"
        response = await send_query(ws_connection, query)
        log_response(query, response)

        assert len(response) > 20, "Response too short"
        assert "pinacolada" in response.lower(), "Should mention PinaColada"
        assert "error" not in response.lower(), f"Got error: {response}"

    @pytest.mark.asyncio
    async def test_list_accounts(self, ws_connection):
        """Agent should list available accounts."""
        query = "what accounts do we have"
        response = await send_query(ws_connection, query)
        log_response(query, response)

        assert len(response) > 50, "Response too short for account list"
        assert "error" not in response.lower(), f"Got error: {response}"
        assert "failed" not in response.lower(), f"Got failure: {response}"


class TestDocumentOperations:
    """Tests for document search and retrieval."""

    @pytest.mark.asyncio
    async def test_search_documents_by_tag(self, ws_connection):
        """Agent should find documents by tag."""
        query = "find documents tagged with resume"
        response = await send_query(ws_connection, query)
        log_response(query, response)

        assert len(response) > 20, "Response too short"
        assert "resume" in response.lower(), "Should mention resume"
        assert "error" not in response.lower(), f"Got error: {response}"

    @pytest.mark.asyncio
    async def test_get_document_content(self, ws_connection):
        """Agent should retrieve document contents."""
        query = "show me the contents of William's resume"
        response = await send_query(ws_connection, query)
        log_response(query, response)

        assert len(response) > 100, "Response too short for resume content"
        assert "william" in response.lower(), "Should contain William's info"
        assert "error" not in response.lower(), f"Got error: {response}"


class TestSearchPatterns:
    """Tests for search query handling."""

    @pytest.mark.asyncio
    async def test_multiword_name_search(self, ws_connection):
        """Agent should handle multi-word names like 'William Hubenschmidt'."""
        query = "look up William Hubenschmidt"
        response = await send_query(ws_connection, query)
        log_response(query, response)

        assert "no individual" not in response.lower() or "found" in response.lower()
        assert "hubenschmidt" in response.lower()

    @pytest.mark.asyncio
    async def test_partial_name_search(self, ws_connection):
        """Agent should find records by partial name."""
        query = "look up Jennifer"
        response = await send_query(ws_connection, query)
        log_response(query, response)

        assert len(response) > 20, "Response too short"
        assert "jennifer" in response.lower(), "Should find Jennifer"
