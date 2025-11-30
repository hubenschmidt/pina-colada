"""Repository layer for document data access."""

import logging
from typing import List, Optional, Dict, Any

from sqlalchemy import select, and_, delete, insert, func
from sqlalchemy.orm import joinedload

from models.Document import Document
from models.EntityAsset import EntityAsset
from models.Tag import Tag, EntityTag
from lib.db import async_get_session

logger = logging.getLogger(__name__)


async def find_documents_by_tenant(
    tenant_id: int,
    entity_type: Optional[str] = None,
    entity_id: Optional[int] = None,
    page: int = 1,
    page_size: int = 50,
    order_by: str = "updated_at",
    order: str = "DESC",
) -> tuple[List[Document], int]:
    """Find documents for a tenant with pagination and sorting, optionally filtered by entity."""
    async with async_get_session() as session:
        # Base query
        base_stmt = select(Document).where(Document.tenant_id == tenant_id)

        if entity_type and entity_id:
            # Join with EntityAsset to filter by entity
            base_stmt = (
                select(Document)
                .join(EntityAsset, EntityAsset.c.asset_id == Document.id)
                .where(
                    and_(
                        Document.tenant_id == tenant_id,
                        EntityAsset.c.entity_type == entity_type,
                        EntityAsset.c.entity_id == entity_id,
                    )
                )
            )

        # Count total
        count_stmt = select(func.count()).select_from(base_stmt.alias())
        count_result = await session.execute(count_stmt)
        total = count_result.scalar() or 0

        # Apply sorting
        sort_column = getattr(Document, order_by, Document.updated_at)
        if order.upper() == "ASC":
            base_stmt = base_stmt.order_by(sort_column.asc())
        else:
            base_stmt = base_stmt.order_by(sort_column.desc())

        # Apply pagination
        offset = (page - 1) * page_size
        stmt = base_stmt.offset(offset).limit(page_size)
        
        result = await session.execute(stmt)
        return list(result.scalars().all()), total


async def find_document_by_id(document_id: int, tenant_id: int) -> Optional[Document]:
    """Find a document by ID within a tenant."""
    async with async_get_session() as session:
        stmt = select(Document).where(
            and_(Document.id == document_id, Document.tenant_id == tenant_id)
        )
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def create_document(data: Dict[str, Any]) -> Document:
    """Create a new document."""
    async with async_get_session() as session:
        try:
            document = Document(**data)
            session.add(document)
            await session.commit()
            await session.refresh(document)
            return document
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to create document: {e}")
            raise


async def update_document(
    document_id: int, tenant_id: int, data: Dict[str, Any]
) -> Optional[Document]:
    """Update a document."""
    async with async_get_session() as session:
        try:
            stmt = select(Document).where(
                and_(Document.id == document_id, Document.tenant_id == tenant_id)
            )
            result = await session.execute(stmt)
            document = result.scalar_one_or_none()
            if not document:
                return None
            for key, value in data.items():
                if hasattr(document, key):
                    setattr(document, key, value)
            await session.commit()
            await session.refresh(document)
            return document
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to update document: {e}")
            raise


async def delete_document(document_id: int, tenant_id: int) -> Optional[str]:
    """Delete a document and return storage_path for cleanup."""
    async with async_get_session() as session:
        try:
            stmt = select(Document).where(
                and_(Document.id == document_id, Document.tenant_id == tenant_id)
            )
            result = await session.execute(stmt)
            document = result.scalar_one_or_none()
            if not document:
                return None
            storage_path = document.storage_path
            await session.delete(document)
            await session.commit()
            return storage_path
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to delete document: {e}")
            raise


async def link_document_to_entity(
    document_id: int, entity_type: str, entity_id: int
) -> bool:
    """Link a document to an entity via EntityAsset."""
    async with async_get_session() as session:
        try:
            stmt = insert(EntityAsset).values(
                asset_id=document_id,
                entity_type=entity_type,
                entity_id=entity_id,
            )
            await session.execute(stmt)
            await session.commit()
            return True
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to link document: {e}")
            raise


async def unlink_document_from_entity(
    document_id: int, entity_type: str, entity_id: int
) -> bool:
    """Unlink a document from an entity."""
    async with async_get_session() as session:
        try:
            stmt = delete(EntityAsset).where(
                and_(
                    EntityAsset.c.asset_id == document_id,
                    EntityAsset.c.entity_type == entity_type,
                    EntityAsset.c.entity_id == entity_id,
                )
            )
            result = await session.execute(stmt)
            await session.commit()
            return result.rowcount > 0
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to unlink document: {e}")
            raise


async def get_document_entities(document_id: int) -> List[Dict[str, Any]]:
    """Get all entities linked to a document."""
    async with async_get_session() as session:
        stmt = select(EntityAsset).where(EntityAsset.c.asset_id == document_id)
        result = await session.execute(stmt)
        return [
            {"entity_type": row.entity_type, "entity_id": row.entity_id}
            for row in result.fetchall()
        ]


async def get_document_tags(document_id: int) -> List[str]:
    """Get all tags for a document."""
    async with async_get_session() as session:
        stmt = (
            select(Tag.name)
            .join(EntityTag, EntityTag.c.tag_id == Tag.id)
            .where(
                and_(
                    EntityTag.c.entity_type == "Asset",
                    EntityTag.c.entity_id == document_id,
                )
            )
            .order_by(Tag.name)
        )
        result = await session.execute(stmt)
        return [row[0] for row in result.fetchall()]
