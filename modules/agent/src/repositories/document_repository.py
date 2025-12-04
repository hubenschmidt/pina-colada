"""Repository layer for document data access."""

import logging
from typing import Any, Dict, List, Optional
from sqlalchemy import and_, delete, func, or_, select, update
from sqlalchemy.orm import joinedload
from lib.db import async_get_session
from models.Asset import Asset
from models.Contact import Contact
from models.Document import Document
from models.EntityAsset import EntityAsset
from models.Individual import Individual
from models.Lead import Lead
from models.Organization import Organization
from models.Project import Project
from models.Tag import EntityTag, Tag
from schemas.document import DocumentUpdate, EntityLink

__all__ = ["DocumentUpdate", "EntityLink"]

logger = logging.getLogger(__name__)


async def find_documents_by_tenant(
    tenant_id: int,
    entity_type: Optional[str] = None,
    entity_id: Optional[int] = None,
    page: int = 1,
    page_size: int = 50,
    order_by: str = "updated_at",
    order: str = "DESC",
    search: Optional[str] = None,
    tags: Optional[List[str]] = None,
    current_versions_only: bool = True,
) -> tuple[List[Document], int]:
    """Find documents for a tenant with pagination and sorting, optionally filtered by entity, tags, and search."""
    async with async_get_session() as session:
        # Base query - filter by current versions by default
        base_stmt = select(Document).where(Document.tenant_id == tenant_id)
        if current_versions_only:
            base_stmt = base_stmt.where(Document.is_current_version == True)

        if entity_type and entity_id:
            # Join with EntityAsset to filter by entity
            # Note: Don't filter by is_current_version here - entity links point
            # to specific versions, so we show the version that was actually linked
            conditions = [
                Document.tenant_id == tenant_id,
                EntityAsset.entity_type == entity_type,
                EntityAsset.entity_id == entity_id,
            ]
            base_stmt = (
                select(Document)
                .join(EntityAsset, EntityAsset.asset_id == Document.id)
                .where(and_(*conditions))
            )

        # Filter by tags if provided
        if tags and len(tags) > 0:
            base_stmt = (
                base_stmt.join(EntityTag, EntityTag.entity_id == Document.id)
                .join(Tag, Tag.id == EntityTag.tag_id)
                .where(
                    and_(
                        EntityTag.entity_type == "Asset",
                        Tag.name.in_(tags),
                    )
                )
            )

        # Apply search filter (filename)
        if search and search.strip():
            search_lower = search.strip().lower()
            base_stmt = base_stmt.where(
                func.lower(Document.filename).contains(search_lower)
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
        # Use unique() if we have joins (tags or entity filtering with joins) that might create duplicates
        if tags or (entity_type and entity_id):
            return list(result.unique().scalars().all()), total
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
            # Check if link already exists
            existing_stmt = select(func.count()).select_from(EntityAsset).where(
                and_(
                    EntityAsset.asset_id == document_id,
                    EntityAsset.entity_type == entity_type,
                    EntityAsset.entity_id == entity_id,
                )
            )
            result = await session.execute(existing_stmt)
            count = result.scalar() or 0
            if count > 0:
                return True  # Already linked

            link = EntityAsset(
                asset_id=document_id,
                entity_type=entity_type,
                entity_id=entity_id,
            )
            session.add(link)
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
                    EntityAsset.asset_id == document_id,
                    EntityAsset.entity_type == entity_type,
                    EntityAsset.entity_id == entity_id,
                )
            )
            result = await session.execute(stmt)
            await session.commit()
            return result.rowcount > 0
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to unlink document: {e}")
            raise


async def _get_entity_name(session, entity_type: str, entity_id: int) -> Optional[str]:
    """Get entity name by type and id."""
    name_queries = {
        "Organization": lambda eid: select(Organization.name).where(Organization.id == eid),
        "Project": lambda eid: select(Project.name).where(Project.id == eid),
        "Lead": lambda eid: select(Lead.title).where(Lead.id == eid),
    }
    composite_name_queries = {
        "Individual": lambda eid: select(Individual.first_name, Individual.last_name).where(Individual.id == eid),
        "Contact": lambda eid: select(Contact.first_name, Contact.last_name).where(Contact.id == eid),
    }

    if entity_type in name_queries:
        result = await session.execute(name_queries[entity_type](entity_id))
        return result.scalar()

    if entity_type in composite_name_queries:
        result = await session.execute(composite_name_queries[entity_type](entity_id))
        row = result.first()
        if not row:
            return None
        return f"{row[0]} {row[1]}"

    return None


async def get_document_entities(document_id: int) -> List[Dict[str, Any]]:
    """Get all entities linked to a document with their names."""
    async with async_get_session() as session:
        stmt = select(EntityAsset).where(EntityAsset.asset_id == document_id)
        result = await session.execute(stmt)
        rows = result.scalars().all()

        entities = []
        for row in rows:
            entity_type = row.entity_type
            entity_id = row.entity_id
            entity_name = await _get_entity_name(session, entity_type, entity_id)

            entities.append({
                "entity_type": entity_type,
                "entity_id": entity_id,
                "entity_name": entity_name or f"{entity_type} #{entity_id}",
            })

        return entities


async def get_document_tags(document_id: int) -> List[str]:
    """Get all tags for a document."""
    async with async_get_session() as session:
        stmt = (
            select(Tag.name)
            .join(EntityTag, EntityTag.tag_id == Tag.id)
            .where(
                and_(
                    EntityTag.entity_type == "Asset",
                    EntityTag.entity_id == document_id,
                )
            )
            .order_by(Tag.name)
        )
        result = await session.execute(stmt)
        return [row[0] for row in result.fetchall()]


async def get_documents_entities_batch(document_ids: List[int]) -> Dict[int, List[Dict[str, Any]]]:
    """Get entities for multiple documents in a single query.

    Returns a dict mapping document_id -> list of entity dicts.
    """
    if not document_ids:
        return {}

    async with async_get_session() as session:
        # Get all entity links for all documents
        stmt = select(EntityAsset).where(EntityAsset.asset_id.in_(document_ids))
        result = await session.execute(stmt)
        rows = result.scalars().all()

        # Group by document and collect unique entity lookups needed
        doc_entities: Dict[int, List[tuple]] = {}
        entity_lookups: Dict[tuple, str] = {}  # (type, id) -> name

        for row in rows:
            doc_id = row.asset_id
            entity_type = row.entity_type
            entity_id = row.entity_id

            if doc_id not in doc_entities:
                doc_entities[doc_id] = []
            doc_entities[doc_id].append((entity_type, entity_id))
            entity_lookups[(entity_type, entity_id)] = None

        # Batch fetch entity names by type
        for entity_type in ["Organization", "Project", "Lead", "Individual", "Contact"]:
            type_ids = [eid for (etype, eid) in entity_lookups.keys() if etype == entity_type]
            if not type_ids:
                continue

            if entity_type == "Organization":
                stmt = select(Organization.id, Organization.name).where(Organization.id.in_(type_ids))
            elif entity_type == "Project":
                stmt = select(Project.id, Project.name).where(Project.id.in_(type_ids))
            elif entity_type == "Lead":
                stmt = select(Lead.id, Lead.title).where(Lead.id.in_(type_ids))
            elif entity_type == "Individual":
                stmt = select(Individual.id, Individual.first_name, Individual.last_name).where(Individual.id.in_(type_ids))
            elif entity_type == "Contact":
                stmt = select(Contact.id, Contact.first_name, Contact.last_name).where(Contact.id.in_(type_ids))
            else:
                continue

            result = await session.execute(stmt)
            for row in result.fetchall():
                if entity_type in ["Individual", "Contact"]:
                    name = f"{row[1]} {row[2]}"
                else:
                    name = row[1]
                entity_lookups[(entity_type, row[0])] = name

        # Build final result
        result_dict: Dict[int, List[Dict[str, Any]]] = {doc_id: [] for doc_id in document_ids}
        for doc_id, entities in doc_entities.items():
            for entity_type, entity_id in entities:
                name = entity_lookups.get((entity_type, entity_id))
                result_dict[doc_id].append({
                    "entity_type": entity_type,
                    "entity_id": entity_id,
                    "entity_name": name or f"{entity_type} #{entity_id}",
                })

        return result_dict


async def get_documents_tags_batch(document_ids: List[int]) -> Dict[int, List[str]]:
    """Get tags for multiple documents in a single query.

    Returns a dict mapping document_id -> list of tag names.
    """
    if not document_ids:
        return {}

    async with async_get_session() as session:
        stmt = (
            select(EntityTag.entity_id, Tag.name)
            .join(Tag, Tag.id == EntityTag.tag_id)
            .where(
                and_(
                    EntityTag.entity_type == "Asset",
                    EntityTag.entity_id.in_(document_ids),
                )
            )
            .order_by(EntityTag.entity_id, Tag.name)
        )
        result = await session.execute(stmt)

        # Build result dict
        result_dict: Dict[int, List[str]] = {doc_id: [] for doc_id in document_ids}
        for row in result.fetchall():
            doc_id = row[0]
            tag_name = row[1]
            result_dict[doc_id].append(tag_name)

        return result_dict


async def find_document_versions(document_id: int, tenant_id: int) -> List[Document]:
    """Get all versions of a document (including itself).

    Returns versions ordered by version_number descending (newest first).
    Works whether document_id is the parent or a child version.
    """
    async with async_get_session() as session:
        # First, find the root document (parent_id is NULL)
        doc_stmt = select(Document).where(
            and_(Document.id == document_id, Document.tenant_id == tenant_id)
        )
        result = await session.execute(doc_stmt)
        document = result.scalar_one_or_none()
        if not document:
            return []

        # Get the root document ID
        root_id = document.parent_id if document.parent_id else document.id

        # Get all versions: root + all children
        stmt = (
            select(Document)
            .where(
                and_(
                    Document.tenant_id == tenant_id,
                    or_(
                        Document.id == root_id,
                        Document.parent_id == root_id,
                    ),
                )
            )
            .order_by(Document.version_number.desc())
        )
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def get_version_count(document_id: int, tenant_id: int) -> int:
    """Get the count of all versions for a document."""
    async with async_get_session() as session:
        # First, find the document
        doc_stmt = select(Document).where(
            and_(Document.id == document_id, Document.tenant_id == tenant_id)
        )
        result = await session.execute(doc_stmt)
        document = result.scalar_one_or_none()
        if not document:
            return 0

        # Get the root document ID
        root_id = document.parent_id if document.parent_id else document.id

        # Count all versions
        count_stmt = select(func.count()).select_from(Document).where(
            and_(
                Document.tenant_id == tenant_id,
                or_(
                    Document.id == root_id,
                    Document.parent_id == root_id,
                ),
            )
        )
        result = await session.execute(count_stmt)
        return result.scalar() or 0


async def find_existing_document_by_filename(
    tenant_id: int, filename: str, entity_type: str, entity_id: int
) -> Optional[Document]:
    """Check if a document with the same filename exists on an entity.

    Returns the current version of the document if found.
    """
    async with async_get_session() as session:
        stmt = (
            select(Document)
            .join(EntityAsset, EntityAsset.asset_id == Document.id)
            .where(
                and_(
                    Document.tenant_id == tenant_id,
                    Document.filename == filename,
                    Document.is_current_version == True,
                    EntityAsset.entity_type == entity_type,
                    EntityAsset.entity_id == entity_id,
                )
            )
        )
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def create_new_version(
    parent_id: int, tenant_id: int, user_id: int, storage_path: str,
    file_size: int, content_type: str
) -> Optional[Document]:
    """Create a new version of an existing document.

    1. Finds the parent document
    2. Marks the current version as not current
    3. Creates a new version with incremented version_number
    4. Copies entity links from parent to new version
    """
    async with async_get_session() as session:
        try:
            # Find the root document
            root_stmt = select(Document).where(
                and_(Document.id == parent_id, Document.tenant_id == tenant_id)
            )
            result = await session.execute(root_stmt)
            root_doc = result.scalar_one_or_none()
            if not root_doc:
                return None

            # Get actual root ID (in case parent_id is already a child)
            actual_root_id = root_doc.parent_id if root_doc.parent_id else root_doc.id

            # Get max version number
            max_version_stmt = (
                select(func.max(Document.version_number))
                .where(
                    and_(
                        Document.tenant_id == tenant_id,
                        or_(
                            Document.id == actual_root_id,
                            Document.parent_id == actual_root_id,
                        ),
                    )
                )
            )
            result = await session.execute(max_version_stmt)
            max_version = result.scalar() or 1

            # Mark all versions as not current (use Asset table directly for inheritance)
            update_stmt = (
                update(Asset)
                .where(
                    and_(
                        Asset.tenant_id == tenant_id,
                        or_(
                            Asset.id == actual_root_id,
                            Asset.parent_id == actual_root_id,
                        ),
                    )
                )
                .values(is_current_version=False)
            )
            await session.execute(update_stmt)

            # Get root document for metadata
            root_stmt = select(Document).where(Document.id == actual_root_id)
            result = await session.execute(root_stmt)
            root_doc = result.scalar_one()

            # Create new version
            new_version = Document(
                asset_type="document",
                tenant_id=tenant_id,
                user_id=user_id,
                filename=root_doc.filename,
                content_type=content_type,
                description=root_doc.description,
                storage_path=storage_path,
                file_size=file_size,
                parent_id=actual_root_id,
                version_number=max_version + 1,
                is_current_version=True,
            )
            session.add(new_version)
            await session.flush()

            # Note: Entity links are NOT copied to new version.
            # Each version maintains its own entity links, so entities
            # show the specific version that was linked to them.

            await session.commit()
            await session.refresh(new_version)
            return new_version

        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to create new version: {e}")
            raise


async def set_current_version(document_id: int, tenant_id: int) -> Optional[Document]:
    """Set a specific version as the current version.

    Marks all other versions as not current.
    """
    async with async_get_session() as session:
        try:
            # Find the document
            doc_stmt = select(Document).where(
                and_(Document.id == document_id, Document.tenant_id == tenant_id)
            )
            result = await session.execute(doc_stmt)
            document = result.scalar_one_or_none()
            if not document:
                return None

            # Get root ID
            root_id = document.parent_id if document.parent_id else document.id

            # Mark all versions as not current (use Asset table directly for inheritance)
            update_stmt = (
                update(Asset)
                .where(
                    and_(
                        Asset.tenant_id == tenant_id,
                        or_(
                            Asset.id == root_id,
                            Asset.parent_id == root_id,
                        ),
                    )
                )
                .values(is_current_version=False)
            )
            await session.execute(update_stmt)

            # Mark this version as current
            set_current_stmt = (
                update(Asset)
                .where(Asset.id == document_id)
                .values(is_current_version=True)
            )
            await session.execute(set_current_stmt)
            await session.commit()
            await session.refresh(document)
            return document

        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to set current version: {e}")
            raise
