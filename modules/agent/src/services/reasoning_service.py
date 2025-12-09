"""Service layer for reasoning schema registry (RAG context)."""

from typing import List, Dict, Any

from repositories.reasoning_repository import (
    find_reasoning_table_names,
    find_reasoning_by_type,
)


async def get_reasoning_tables(reasoning_type: str) -> List[str]:
    """Get all table names for a reasoning context."""
    return await find_reasoning_table_names(reasoning_type)


async def get_reasoning_schema(reasoning_type: str) -> List[Dict[str, Any]]:
    """Get tables with descriptions and schema hints for richer context."""
    records = await find_reasoning_by_type(reasoning_type)
    return [
        {
            "table": r.table_name,
            "description": r.description,
            "hints": r.schema_hint,
        }
        for r in records
    ]


def format_schema_context(schema_data: List[Dict[str, Any]]) -> str:
    """Format schema data for LLM prompt injection."""
    if not schema_data:
        return "No schema context available."

    lines = ["CRM DATABASE SCHEMA:"]
    for item in schema_data:
        line = f"- {item['table']}: {item['description'] or 'No description'}"
        if item["hints"]:
            key_fields = item["hints"].get("key_fields", [])
            if key_fields:
                line += f" [fields: {', '.join(key_fields)}]"
            if item["hints"].get("polymorphic"):
                line += " (polymorphic)"
        lines.append(line)
    return "\n".join(lines)
