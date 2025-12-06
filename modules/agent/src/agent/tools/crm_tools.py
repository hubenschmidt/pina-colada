"""
CRM Tools for AI-assisted CRM worker.

Provides:
- Service-layer lookup tools (preferred)
- SQL fallback for complex queries
- Web search reused from worker_tools
"""

import logging
from typing import Optional, List

from langchain_core.tools import Tool, StructuredTool
from pydantic import BaseModel, Field

from services.account_service import search_accounts
from services.organization_service import search_organizations
from services.individual_service import search_individuals
from services.contact_service import search_contacts
from services.reasoning_service import get_reasoning_tables
from lib.tenant_context import get_tenant_id

logger = logging.getLogger(__name__)


async def _get_tenant_with_fallback() -> int | None:
    """Get tenant_id from context, or fallback to default 'pinacolada' tenant."""
    tenant_id = get_tenant_id()
    if tenant_id:
        return tenant_id
    # Fallback: get default tenant
    from lib.db import async_get_session
    from sqlalchemy import text
    try:
        async with async_get_session() as session:
            result = await session.execute(text("SELECT id FROM \"Tenant\" WHERE slug = 'pinacolada' LIMIT 1"))
            row = result.fetchone()
            return row[0] if row else None
    except Exception as e:
        logger.warning(f"Failed to get default tenant: {e}")
        return None


# --- Pydantic Input Schemas ---

class EntityLookupInput(BaseModel):
    """Search input for CRM entity lookups. Tenant is auto-injected."""
    query: str = Field(description="Search term for entity lookup")




class SQLQueryInput(BaseModel):
    query: str = Field(description="SQL SELECT query to execute")
    reasoning: str = Field(description="Explanation of what this query retrieves")


# --- Lookup Functions ---

async def lookup_account(query: str) -> str:
    """Look up accounts by name. Tenant scoped automatically."""
    tenant_id = await _get_tenant_with_fallback()
    try:
        results = await search_accounts(query, tenant_id)
        if not results:
            return f"No accounts found matching '{query}'"
        formatted = []
        for acc in results[:10]:
            orgs = [o.name for o in acc.organizations[:3]] if acc.organizations else []
            indivs = [f"{i.first_name} {i.last_name}" for i in acc.individuals[:3]] if acc.individuals else []
            formatted.append(f"- {acc.name} (id={acc.id}, orgs={orgs}, individuals={indivs})")
        return f"Found {len(results)} accounts:\n" + "\n".join(formatted)
    except Exception as e:
        logger.error(f"Account lookup failed: {e}")
        return f"Account lookup failed: {e}"


async def lookup_organization(query: str) -> str:
    """Look up organizations by name. Tenant scoped automatically."""
    tenant_id = await _get_tenant_with_fallback()
    try:
        results = await search_organizations(query, tenant_id)
        if not results:
            return f"No organizations found matching '{query}'"
        formatted = []
        for org in results[:10]:
            formatted.append(f"- {org.name} (id={org.id}, website={org.website or 'N/A'})")
        return f"Found {len(results)} organizations:\n" + "\n".join(formatted)
    except Exception as e:
        logger.error(f"Organization lookup failed: {e}")
        return f"Organization lookup failed: {e}"


async def lookup_individual(query: str) -> str:
    """Look up individuals by name or email. Tenant scoped automatically."""
    logger.info(f"🔍 lookup_individual called with query='{query}'")
    tenant_id = await _get_tenant_with_fallback()
    logger.info(f"🔍 lookup_individual using tenant_id={tenant_id}")
    try:
        results = await search_individuals(query, tenant_id)
        logger.info(f"🔍 lookup_individual found {len(results) if results else 0} results")
        if not results:
            return f"No individuals found matching '{query}'"
        formatted = []
        for ind in results[:10]:
            name = f"{ind.first_name or ''} {ind.last_name or ''}".strip()
            formatted.append(f"- {name} (id={ind.id}, email={ind.email or 'N/A'}, title={ind.title or 'N/A'})")
        return f"Found {len(results)} individuals:\n" + "\n".join(formatted)
    except Exception as e:
        logger.error(f"Individual lookup failed: {e}")
        return f"Individual lookup failed: {e}"


async def lookup_contact(query: str) -> str:
    """Look up contacts. Tenant scoped automatically."""
    tenant_id = await _get_tenant_with_fallback()
    try:
        results = await search_contacts(query, tenant_id)
        if not results:
            return f"No contacts found matching '{query}'"
        formatted = []
        for contact in results[:10]:
            formatted.append(f"- Contact id={contact.id}")
        return f"Found {len(results)} contacts:\n" + "\n".join(formatted)
    except Exception as e:
        logger.error(f"Contact lookup failed: {e}")
        return f"Contact lookup failed: {e}"




async def execute_crm_query(query: str, reasoning: str) -> str:
    """
    Execute a read-only SQL query against CRM tables.
    Only SELECT queries on tables registered in the reasoning table are allowed.
    """
    # Validate query is SELECT only
    query_upper = query.strip().upper()
    if not query_upper.startswith("SELECT"):
        return "Error: Only SELECT queries are allowed. Use service tools for writes."

    # Check for dangerous keywords
    dangerous = ["INSERT", "UPDATE", "DELETE", "DROP", "ALTER", "TRUNCATE", "CREATE"]
    for keyword in dangerous:
        if keyword in query_upper:
            return f"Error: {keyword} operations are not allowed. This tool is read-only."

    # Validate tables against reasoning registry
    try:
        allowed_tables = await get_reasoning_tables("crm")
        allowed_lower = [t.lower() for t in allowed_tables]

        # Simple table extraction (could be improved with proper SQL parsing)
        from_idx = query_upper.find("FROM")
        if from_idx == -1:
            return "Error: Could not parse FROM clause in query"

        # Extract table names after FROM (simplified)
        after_from = query[from_idx + 4:].strip()
        # Take first word as table name (simplified parsing)
        table_name = after_from.split()[0].strip('"').strip("'").lower()

        if table_name not in allowed_lower:
            return f"Error: Table '{table_name}' is not in the allowed CRM tables. Allowed: {', '.join(allowed_tables)}"

    except Exception as e:
        logger.error(f"Table validation failed: {e}")
        return f"Error validating tables: {e}"

    # Fix PostgreSQL case sensitivity by quoting table names
    # PostgreSQL lowercases unquoted identifiers, but our tables use PascalCase
    import re
    fixed_query = query
    for table in allowed_tables:
        # Replace unquoted table name with quoted version (case-insensitive)
        pattern = rf'\b{re.escape(table)}\b(?!["\'])'
        fixed_query = re.sub(pattern, f'"{table}"', fixed_query, flags=re.IGNORECASE)

    # Execute query
    try:
        from lib.db import async_get_session
        from sqlalchemy import text

        async with async_get_session() as session:
            result = await session.execute(text(fixed_query))
            rows = result.fetchall()

            if not rows:
                return f"Query returned no results.\nReasoning: {reasoning}"

            # Format results
            columns = result.keys()
            formatted = []
            for row in rows[:50]:  # Limit to 50 rows
                row_dict = dict(zip(columns, row))
                formatted.append(str(row_dict))

            output = f"Query returned {len(rows)} rows (showing first {min(len(rows), 50)}):\n"
            output += "\n".join(formatted)
            output += f"\n\nReasoning: {reasoning}"
            return output

    except Exception as e:
        logger.error(f"SQL query execution failed: {e}")
        return f"Query execution failed: {e}"


# --- Tool Factory ---

async def get_crm_tools() -> List:
    """Return CRM-specific tools for the CRM worker."""
    tools = []

    # Account lookup
    tools.append(StructuredTool.from_function(
        func=lambda **kwargs: None,
        coroutine=lookup_account,
        name="lookup_account",
        description="Look up CRM accounts by name. Returns account details with linked organizations and individuals.",
        args_schema=EntityLookupInput,
    ))

    # Organization lookup
    tools.append(StructuredTool.from_function(
        func=lambda **kwargs: None,
        coroutine=lookup_organization,
        name="lookup_organization",
        description="Look up organizations (companies) by name. Returns org details including website.",
        args_schema=EntityLookupInput,
    ))

    # Individual lookup
    tools.append(StructuredTool.from_function(
        func=lambda **kwargs: None,
        coroutine=lookup_individual,
        name="lookup_individual",
        description="Look up individuals (people) by name or email. Returns contact details.",
        args_schema=EntityLookupInput,
    ))

    # Contact lookup
    tools.append(StructuredTool.from_function(
        func=lambda **kwargs: None,
        coroutine=lookup_contact,
        name="lookup_contact",
        description="Look up contacts linking individuals to organizations.",
        args_schema=EntityLookupInput,
    ))

    # SQL fallback (use for deals, leads, and complex queries)
    tools.append(StructuredTool.from_function(
        func=lambda **kwargs: None,
        coroutine=execute_crm_query,
        name="execute_crm_query",
        description="Execute a read-only SQL SELECT query against CRM tables. Use only for complex joins/aggregations that service tools can't handle. Must explain reasoning.",
        args_schema=SQLQueryInput,
    ))

    logger.info(f"Initialized {len(tools)} CRM tools")
    return tools
