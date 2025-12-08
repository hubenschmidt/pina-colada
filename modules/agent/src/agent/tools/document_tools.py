"""
Document Tools for AI workers.

Provides:
- search_documents: Find documents by query, tags, or entity
- get_document_content: Load document text content (with PDF extraction)
"""

import io
import logging
from functools import wraps
from typing import Optional, List

from cachetools import TTLCache
from pydantic import BaseModel, Field
from pypdf import PdfReader

from lib.tenant_context import get_tenant_id
from lib.storage import get_storage
from repositories.document_repository import find_documents_by_tenant, find_document_by_id

logger = logging.getLogger(__name__)

# Max chars to return (token optimization)
MAX_CONTENT_CHARS = 15000

# TTL Caches for tool results
_document_content_cache: TTLCache = TTLCache(maxsize=100, ttl=300)  # 5 min
_search_cache: TTLCache = TTLCache(maxsize=200, ttl=120)  # 2 min


def async_ttl_cache(cache: TTLCache, key_fn):
    """Decorator for async functions with TTL cache."""
    def decorator(func):
        @wraps(func)
        async def wrapper(*args, **kwargs):
            key = key_fn(*args, **kwargs)
            if key in cache:
                logger.debug(f"Cache HIT: {func.__name__}({key})")
                return cache[key]
            result = await func(*args, **kwargs)
            cache[key] = result
            logger.debug(f"Cache MISS: {func.__name__}({key}) - cached")
            return result
        return wrapper
    return decorator


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

async def _get_default_tenant_id() -> Optional[int]:
    """Get default tenant ID (pinacolada) when no context available."""
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


async def _search_documents_cached(
    tenant_id: int,
    query: Optional[str],
    tags: Optional[str],
    entity_type: Optional[str],
    entity_id: Optional[int],
    limit: int,
) -> str:
    """Cached search implementation."""
    # Check cache first
    cache_key = (tenant_id, query, tags, entity_type, entity_id, limit)
    if cache_key in _search_cache:
        logger.debug(f"Cache HIT: search_documents({cache_key})")
        return _search_cache[cache_key]

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
            result = f"No documents found ({filter_str})"
        else:
            formatted = []
            for doc in documents:
                desc = f" - {doc.description}" if doc.description else ""
                formatted.append(f"- id={doc.id}: {doc.filename}{desc} ({doc.content_type})")
            result = f"Found {len(documents)} of {total} documents:\n" + "\n".join(formatted)

        # Cache the result
        _search_cache[cache_key] = result
        logger.debug(f"Cache MISS: search_documents({cache_key}) - cached")
        return result

    except Exception as e:
        logger.error(f"Document search failed: {e}")
        return f"Document search failed: {e}"


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
        tenant_id = await _get_default_tenant_id()
    if not tenant_id:
        return "Error: No tenant context available. Cannot search documents."

    return await _search_documents_cached(tenant_id, query, tags, entity_type, entity_id, limit)


async def _get_document_content_cached(tenant_id: int, document_id: int) -> str:
    """Cached document content retrieval."""
    # Check cache first
    cache_key = (tenant_id, document_id)
    if cache_key in _document_content_cache:
        logger.debug(f"Cache HIT: get_document_content({cache_key})")
        return _document_content_cache[cache_key]

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

        result = f"=== {document.filename} ===\n\n{content}"

        # Cache the result
        _document_content_cache[cache_key] = result
        logger.debug(f"Cache MISS: get_document_content({cache_key}) - cached")
        return result

    except Exception as e:
        logger.error(f"Failed to get document content: {e}")
        return f"Failed to get document content: {e}"


async def get_document_content(document_id: int) -> str:
    """Get the text content of a document by ID."""
    tenant_id = get_tenant_id()
    if not tenant_id:
        tenant_id = await _get_default_tenant_id()
    if not tenant_id:
        return "Error: No tenant context available. Cannot fetch document."

    return await _get_document_content_cached(tenant_id, document_id)


# --- Simple Wrappers for SDK ---

async def read_document(file_path: str) -> str:
    """Read a document by filename or ID. Wrapper for SDK tools."""
    # Try to parse as document ID
    try:
        doc_id = int(file_path)
        return await get_document_content(doc_id)
    except ValueError:
        pass

    # Search by filename
    results = await search_documents(query=file_path, limit=1)
    if "No documents found" in results:
        return f"Document not found: {file_path}"

    # Try to extract ID from search results
    import re
    match = re.search(r"id=(\d+)", results)
    if match:
        doc_id = int(match.group(1))
        return await get_document_content(doc_id)

    return results


async def list_documents() -> str:
    """List available documents. Wrapper for SDK tools."""
    return await search_documents(limit=20)


async def search_entity_documents(entity_type: str, entity_id: int) -> str:
    """Search for documents linked to a specific entity.

    Args:
        entity_type: Type of entity (Individual, Organization, Account, Contact)
        entity_id: ID of the entity
    """
    return await search_documents(entity_type=entity_type, entity_id=entity_id, limit=20)
