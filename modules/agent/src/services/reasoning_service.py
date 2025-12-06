"""Service layer for reasoning schema registry (RAG context)."""

from typing import List, Dict, Any

from sqlalchemy import select

from lib.db import async_get_session
from models.Reasoning import Reasoning


async def get_reasoning_tables(reasoning_type: str) -> List[str]:
    """Get all table names for a reasoning context."""
    async with async_get_session() as session:
        stmt = select(Reasoning.table_name).where(Reasoning.type == reasoning_type)
        result = await session.execute(stmt)
        return [row[0] for row in result.fetchall()]


async def get_reasoning_schema(reasoning_type: str) -> List[Dict[str, Any]]:
    """Get tables with descriptions and schema hints for richer context."""
    async with async_get_session() as session:
        stmt = select(Reasoning).where(Reasoning.type == reasoning_type)
        result = await session.execute(stmt)
        return [
            {
                "table": r.table_name,
                "description": r.description,
                "hints": r.schema_hint,
            }
            for r in result.scalars().all()
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
