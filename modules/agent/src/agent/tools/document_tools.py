"""
Document Tools for AI workers.

Provides:
- search_documents: Find documents by query, tags, or entity
- get_document_content: Load document text content (with PDF extraction)
"""

import io
import logging
from typing import Optional, List

from langchain_core.tools import StructuredTool
from pydantic import BaseModel, Field
from pypdf import PdfReader

from lib.tenant_context import get_tenant_id
from lib.storage import get_storage
from repositories.document_repository import find_documents_by_tenant, find_document_by_id

logger = logging.getLogger(__name__)

# Max chars to return (token optimization)
MAX_CONTENT_CHARS = 15000


def _extract_pdf_text(content_bytes: bytes) -> str:
    """Extract text from PDF bytes."""
    try:
        reader = PdfReader(io.BytesIO(content_bytes))
        parts = []
        for page in reader.pages:
            text = page.extract_text() or ""
            if text.strip():
                parts.append(text.strip())
        return "\n\n".join(parts)
    except Exception as e:
        logger.warning(f"PDF extraction failed: {e}")
        return ""


# --- Pydantic Input Schemas ---

class SearchDocumentsInput(BaseModel):
    """Search for documents by query and filters."""
    query: Optional[str] = Field(default=None, description="Search term to match in filename")
    tags: Optional[str] = Field(default=None, description="Comma-separated tags to filter by (e.g., 'resume,cover-letter')")
    entity_type: Optional[str] = Field(default=None, description="Entity type to filter by (Individual, Organization, etc.)")
    entity_id: Optional[int] = Field(default=None, description="Entity ID to filter by (requires entity_type)")
    limit: int = Field(default=10, description="Max documents to return")


class GetDocumentContentInput(BaseModel):
    """Get the content of a specific document."""
    document_id: int = Field(description="The ID of the document to fetch")


# --- Tool Functions ---

async def search_documents(
    query: Optional[str] = None,
    tags: Optional[str] = None,
    entity_type: Optional[str] = None,
    entity_id: Optional[int] = None,
    limit: int = 10,
) -> str:
    """Search for documents. Returns document metadata (id, filename, description)."""
    tenant_id = get_tenant_id()
    if not tenant_id:
        # Fallback: get default tenant (pinacolada)
        from lib.db import async_get_session
        from sqlalchemy import select, text
        try:
            async with async_get_session() as session:
                result = await session.execute(text("SELECT id FROM \"Tenant\" WHERE slug = 'pinacolada' LIMIT 1"))
                row = result.fetchone()
                tenant_id = row[0] if row else None
        except Exception as e:
            logger.warning(f"Failed to get default tenant: {e}")
    if not tenant_id:
        return "Error: No tenant context available. Cannot search documents."

    try:
        tag_list = tags.split(",") if tags else None

        documents, total = await find_documents_by_tenant(
            tenant_id=tenant_id,
            entity_type=entity_type,
            entity_id=entity_id,
            page=1,
            page_size=limit,
            order_by="updated_at",
            order="DESC",
            search=query,
            tags=tag_list,
        )

        if not documents:
            filters = []
            if query:
                filters.append(f"query='{query}'")
            if tags:
                filters.append(f"tags={tags}")
            if entity_type:
                filters.append(f"entity_type={entity_type}")
            filter_str = ", ".join(filters) if filters else "no filters"
            return f"No documents found ({filter_str})"

        formatted = []
        for doc in documents:
            desc = f" - {doc.description}" if doc.description else ""
            formatted.append(f"- id={doc.id}: {doc.filename}{desc} ({doc.content_type})")

        return f"Found {len(documents)} of {total} documents:\n" + "\n".join(formatted)

    except Exception as e:
        logger.error(f"Document search failed: {e}")
        return f"Document search failed: {e}"


async def get_document_content(document_id: int) -> str:
    """Get the text content of a document by ID."""
    tenant_id = get_tenant_id()
    if not tenant_id:
        # Fallback: get default tenant (pinacolada)
        from lib.db import async_get_session
        from sqlalchemy import text
        try:
            async with async_get_session() as session:
                result = await session.execute(text("SELECT id FROM \"Tenant\" WHERE slug = 'pinacolada' LIMIT 1"))
                row = result.fetchone()
                tenant_id = row[0] if row else None
        except Exception as e:
            logger.warning(f"Failed to get default tenant: {e}")
    if not tenant_id:
        return "Error: No tenant context available. Cannot fetch document."

    try:
        document = await find_document_by_id(document_id, tenant_id)
        if not document:
            return f"Document {document_id} not found"

        # Download content
        storage = get_storage()
        content_bytes = await storage.download(document.storage_path)

        # Extract text based on content type
        content_type = document.content_type or ""

        if "pdf" in content_type:
            # PDF: extract text using pypdf
            content = _extract_pdf_text(content_bytes)
            if not content:
                return f"Document {document_id} ({document.filename}): Could not extract text from PDF"
        elif any(t in content_type for t in ["text/", "application/json"]):
            # Text files: decode directly
            try:
                content = content_bytes.decode("utf-8")
            except UnicodeDecodeError:
                content = content_bytes.decode("latin-1")
        else:
            return f"Document {document_id} ({document.filename}): Unsupported type ({content_type})"

        # Truncate for token optimization
        if len(content) > MAX_CONTENT_CHARS:
            content = content[:MAX_CONTENT_CHARS] + f"\n\n[Truncated - {len(content)} total chars]"

        return f"=== {document.filename} ===\n\n{content}"

    except Exception as e:
        logger.error(f"Failed to get document content: {e}")
        return f"Failed to get document content: {e}"


# --- Tool Factory ---

async def get_document_tools() -> List:
    """Return document tools for workers."""
    tools = []

    tools.append(StructuredTool.from_function(
        func=lambda **kwargs: None,
        coroutine=search_documents,
        name="search_documents",
        description="Search for documents by filename, tags, or linked entity. Returns document metadata including IDs.",
        args_schema=SearchDocumentsInput,
    ))

    tools.append(StructuredTool.from_function(
        func=lambda **kwargs: None,
        coroutine=get_document_content,
        name="get_document_content",
        description="Get the text content of a document by its ID. Use search_documents first to find document IDs.",
        args_schema=GetDocumentContentInput,
    ))

    logger.info(f"Initialized {len(tools)} document tools")
    return tools
