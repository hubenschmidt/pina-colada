"""
Integration tests for AI chat via WebSocket.

Tests verify the agent can:
- Look up CRM entities (individuals, organizations, accounts)
- Search and retrieve documents
- Handle multi-word search queries

Run with: pytest src/agent/__tests__/test_ai_chat.py -vs
"""

import pytest

from agent.__tests__.conftest import send_query, QueryResult


def log_response(query: str, result: QueryResult) -> None:
    """Print query/response/tokens for debugging with -s flag."""
    print(f"\n{'='*60}")
    print(f"Q: {query}")
    print(f"A: {result.response[:300]}{'...' if len(result.response) > 300 else ''}")
    print(f"TOKENS: {result.total_tokens} total ({result.input_tokens} in, {result.output_tokens} out)")
    print(f"{'='*60}")


class TestCRMLookups:
    """Tests for CRM entity lookups."""

    @pytest.mark.asyncio
    async def test_lookup_individual(self, ws_connection):
        """Agent should find William Hubenschmidt's individual record."""
        query = "look up William Hubenschmidt individual record"
        result = await send_query(ws_connection, query)
        log_response(query, result)

        assert len(result.response) > 20, "Response too short"
        assert "william" in result.response.lower(), "Should mention William"
        assert "hubenschmidt" in result.response.lower(), "Should mention Hubenschmidt"
        assert "error" not in result.response.lower(), f"Got error: {result.response}"
        assert "failed" not in result.response.lower(), f"Got failure: {result.response}"

    @pytest.mark.asyncio
    async def test_lookup_organization(self, ws_connection):
        """Agent should find PinaColada organization."""
        query = "find information about PinaColada"
        result = await send_query(ws_connection, query)
        log_response(query, result)

        assert len(result.response) > 20, "Response too short"
        assert "pinacolada" in result.response.lower(), "Should mention PinaColada"
        assert "error" not in result.response.lower(), f"Got error: {result.response}"

    @pytest.mark.asyncio
    async def test_list_accounts(self, ws_connection):
        """Agent should list available accounts."""
        query = "what accounts do we have"
        result = await send_query(ws_connection, query)
        log_response(query, result)

        assert len(result.response) > 50, "Response too short for account list"
        assert "error" not in result.response.lower(), f"Got error: {result.response}"
        assert "failed" not in result.response.lower(), f"Got failure: {result.response}"


class TestDocumentOperations:
    """Tests for document search and retrieval."""

    @pytest.mark.asyncio
    async def test_search_documents_by_tag(self, ws_connection):
        """Agent should find documents by tag."""
        query = "find documents tagged with resume"
        result = await send_query(ws_connection, query)
        log_response(query, result)

        assert len(result.response) > 20, "Response too short"
        assert "resume" in result.response.lower(), "Should mention resume"
        assert "error" not in result.response.lower(), f"Got error: {result.response}"

    @pytest.mark.asyncio
    async def test_get_document_content(self, ws_connection):
        """Agent should retrieve document contents."""
        query = "show me the contents of William's resume"
        result = await send_query(ws_connection, query)
        log_response(query, result)

        assert len(result.response) > 100, "Response too short for resume content"
        assert "william" in result.response.lower(), "Should contain William's info"
        assert "error" not in result.response.lower(), f"Got error: {result.response}"


class TestSearchPatterns:
    """Tests for search query handling."""

    @pytest.mark.asyncio
    async def test_multiword_name_search(self, ws_connection):
        """Agent should handle multi-word names like 'William Hubenschmidt'."""
        query = "look up William Hubenschmidt"
        result = await send_query(ws_connection, query)
        log_response(query, result)

        assert "no individual" not in result.response.lower() or "found" in result.response.lower()
        assert "hubenschmidt" in result.response.lower()

    @pytest.mark.asyncio
    async def test_partial_name_search(self, ws_connection):
        """Agent should find records by partial name."""
        query = "look up Jennifer"
        result = await send_query(ws_connection, query)
        log_response(query, result)

        assert len(result.response) > 20, "Response too short"
        assert "jennifer" in result.response.lower(), "Should find Jennifer"


class TestFastPathTokenBudget:
    """Tests for fast-path token optimization (<500 tokens)."""

    @pytest.mark.asyncio
    async def test_individual_lookup_by_full_name(self, ws_connection):
        """Individual lookup by full name should use fast-path."""
        query = "look up William Hubenschmidt"
        result = await send_query(ws_connection, query)
        log_response(query, result)

        assert "hubenschmidt" in result.response.lower(), "Should find William"
        assert result.total_tokens < 500, f"Fast-path should use <500 tokens, got {result.total_tokens}"

    @pytest.mark.asyncio
    async def test_individual_lookup_by_first_name(self, ws_connection):
        """Individual lookup by first name should use fast-path."""
        query = "look up Jennifer"
        result = await send_query(ws_connection, query)
        log_response(query, result)

        assert "jennifer" in result.response.lower(), "Should find Jennifer"
        assert result.total_tokens < 500, f"Fast-path should use <500 tokens, got {result.total_tokens}"

    @pytest.mark.asyncio
    async def test_individual_lookup_imperative(self, ws_connection):
        """'find [name]' should use fast-path."""
        query = "find John Smith"
        result = await send_query(ws_connection, query)
        log_response(query, result)

        # May or may not find John Smith, but should use fast-path
        assert result.total_tokens < 500, f"Fast-path should use <500 tokens, got {result.total_tokens}"

    @pytest.mark.asyncio
    async def test_organization_lookup(self, ws_connection):
        """Organization lookup should use fast-path."""
        query = "look up PinaColada organization"
        result = await send_query(ws_connection, query)
        log_response(query, result)

        assert "pinacolada" in result.response.lower(), "Should find PinaColada"
        assert result.total_tokens < 500, f"Fast-path should use <500 tokens, got {result.total_tokens}"

    @pytest.mark.asyncio
    async def test_account_lookup_by_name(self, ws_connection):
        """Account lookup by name should use fast-path."""
        query = "look up CloudScale Systems account"
        result = await send_query(ws_connection, query)
        log_response(query, result)

        assert result.total_tokens < 500, f"Fast-path should use <500 tokens, got {result.total_tokens}"

    @pytest.mark.asyncio
    async def test_who_is_query(self, ws_connection):
        """'who is [name]' should use fast-path."""
        query = "who is William Hubenschmidt"
        result = await send_query(ws_connection, query)
        log_response(query, result)

        assert result.total_tokens < 500, f"Fast-path should use <500 tokens, got {result.total_tokens}"


class TestFullFlowTokenBudget:
    """Tests for full-flow token optimization (<1500 tokens target)."""

    @pytest.mark.asyncio
    async def test_list_all_accounts(self, ws_connection):
        """List accounts should stay under budget."""
        query = "what accounts do we have"
        result = await send_query(ws_connection, query)
        log_response(query, result)

        assert len(result.response) > 50, "Should return account list"
        assert result.total_tokens < 1500, f"Full-flow should use <1500 tokens, got {result.total_tokens}"

    @pytest.mark.asyncio
    async def test_list_all_organizations(self, ws_connection):
        """List organizations should stay under budget."""
        query = "list all organizations"
        result = await send_query(ws_connection, query)
        log_response(query, result)

        assert len(result.response) > 50, "Should return org list"
        assert result.total_tokens < 1500, f"Full-flow should use <1500 tokens, got {result.total_tokens}"

    @pytest.mark.asyncio
    async def test_list_all_contacts(self, ws_connection):
        """List contacts should stay under budget."""
        query = "show me all contacts"
        result = await send_query(ws_connection, query)
        log_response(query, result)

        assert len(result.response) > 20, "Should return contacts"
        assert result.total_tokens < 1500, f"Full-flow should use <1500 tokens, got {result.total_tokens}"

    @pytest.mark.asyncio
    async def test_search_documents(self, ws_connection):
        """Document search uses full flow (not optimized for fast-path yet)."""
        query = "find documents tagged with resume"
        result = await send_query(ws_connection, query)
        log_response(query, result)

        assert "resume" in result.response.lower(), "Should find resume docs"
        # Note: Document operations use full flow and may fetch content
        # Token optimization for documents is future work

    @pytest.mark.asyncio
    async def test_count_query(self, ws_connection):
        """Count queries should stay under budget."""
        query = "how many individuals do we have"
        result = await send_query(ws_connection, query)
        log_response(query, result)

        assert len(result.response) > 10, "Should return count"
        assert result.total_tokens < 1500, f"Full-flow should use <1500 tokens, got {result.total_tokens}"
