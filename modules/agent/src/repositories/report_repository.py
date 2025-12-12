"""Repository layer for report query execution."""

import logging
from typing import Any, Dict, List, Optional

from sqlalchemy import select, func
from sqlalchemy.orm import selectinload

from lib.db import async_get_session
from models import Base
from models.Organization import Organization
from models.Individual import Individual
from models.Contact import Contact
from models.Lead import Lead
from models.Account import Account
from models.Note import Note
from models.LeadProject import LeadProject
from models.User import User

logger = logging.getLogger(__name__)

# Entity name to model mapping for dynamic queries
ENTITY_MODEL_MAP = {
    "organizations": Organization,
    "individuals": Individual,
    "contacts": Contact,
    "leads": Lead,
    "notes": Note,
}

# Polymorphic entity type to model mapping
POLYMORPHIC_MODEL_MAP = {
    "Organization": Organization,
    "Individual": Individual,
    "Contact": Contact,
    "Lead": Lead,
}

# SQL operators for query building
OPERATORS = {
    "eq": lambda col, val: col == val,
    "neq": lambda col, val: col != val,
    "gt": lambda col, val: col > val,
    "gte": lambda col, val: col >= val,
    "lt": lambda col, val: col < val,
    "lte": lambda col, val: col <= val,
    "contains": lambda col, val: col.ilike(f"%{val}%"),
    "starts_with": lambda col, val: col.ilike(f"{val}%"),
    "is_null": lambda col, val: col.is_(None),
    "is_not_null": lambda col, val: col.isnot(None),
    "in": lambda col, val: col.in_(val if isinstance(val, list) else [val]),
}


def get_entity_model(entity_name: str):
    """Get the SQLAlchemy model for an entity name."""
    return ENTITY_MODEL_MAP.get(entity_name)


def get_polymorphic_model(entity_type: str):
    """Get the SQLAlchemy model for a polymorphic entity type."""
    return POLYMORPHIC_MODEL_MAP.get(entity_type)


def _apply_tenant_filter(stmt, model, primary_entity: str, tenant_id: int):
    """Apply tenant filter based on entity type."""
    if primary_entity == "notes":
        return stmt.where(model.tenant_id == tenant_id)
    if primary_entity == "contacts":
        return stmt
    return stmt.join(Account, model.account_id == Account.id).where(Account.tenant_id == tenant_id)


def _apply_primary_filters(stmt, model, filters: List[Dict[str, Any]]):
    """Apply filters on primary entity columns."""
    for f in filters:
        field = f.get("field")
        operator = f.get("operator")
        if not field or not operator or "." in field:
            continue
        if not hasattr(model, field):
            continue
        col = getattr(model, field)
        op_func = OPERATORS.get(operator)
        if op_func:
            stmt = stmt.where(op_func(col, f.get("value")))
    return stmt


async def execute_report_query(
    tenant_id: int,
    primary_entity: str,
    filters: Optional[List[Dict[str, Any]]] = None,
    limit: int = 100,
    offset: int = 0,
    project_id: Optional[int] = None,
) -> tuple:
    """Execute a report query and return rows and total count."""
    model = get_entity_model(primary_entity)
    if not model:
        return [], 0

    async with async_get_session() as session:
        stmt = select(model)
        stmt = _apply_tenant_filter(stmt, model, primary_entity, tenant_id)

        # Apply project filter for leads
        if project_id is not None and primary_entity == "leads":
            stmt = stmt.join(LeadProject, Lead.id == LeadProject.lead_id).where(
                LeadProject.project_id == project_id
            )

        # Load relationships for built-in join fields
        if primary_entity == "organizations":
            stmt = stmt.options(
                selectinload(Organization.account),
                selectinload(Organization.employee_count_range),
                selectinload(Organization.funding_stage),
                selectinload(Organization.revenue_range),
            )
        if primary_entity == "individuals":
            stmt = stmt.options(selectinload(Individual.account))

        # Apply filters on primary entity
        stmt = _apply_primary_filters(stmt, model, filters or [])

        # Get total count before pagination
        count_stmt = select(func.count()).select_from(stmt.subquery())
        count_result = await session.execute(count_stmt)
        total = count_result.scalar() or 0

        stmt = stmt.limit(limit).offset(offset)
        result = await session.execute(stmt)
        rows = list(result.scalars().all())

        return rows, total


async def fetch_polymorphic_entity(entity_type: str, entity_id: int):
    """Fetch a polymorphic entity by type and ID."""
    model = get_polymorphic_model(entity_type)
    if not model:
        return None

    async with async_get_session() as session:
        stmt = select(model).where(model.id == entity_id)
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def fetch_notes_for_entity(tenant_id: int, entity_type: str, entity_id: int):
    """Fetch notes for an entity."""
    async with async_get_session() as session:
        stmt = (
            select(Note)
            .where(
                Note.tenant_id == tenant_id,
                Note.entity_type == entity_type,
                Note.entity_id == entity_id,
            )
            .order_by(Note.created_at.desc())
            .limit(1)
        )
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def get_leads_for_pipeline(
    tenant_id: int,
    date_from: Optional[str] = None,
    date_to: Optional[str] = None,
    project_id: Optional[int] = None,
) -> List:
    """Get leads for pipeline report."""
    async with async_get_session() as session:
        stmt = (
            select(Lead)
            .join(Account, Lead.account_id == Account.id)
            .where(Account.tenant_id == tenant_id)
        )

        if project_id is not None:
            stmt = stmt.join(LeadProject, Lead.id == LeadProject.lead_id).where(
                LeadProject.project_id == project_id
            )

        if date_from:
            stmt = stmt.where(Lead.created_at >= date_from)
        if date_to:
            stmt = stmt.where(Lead.created_at <= date_to)

        result = await session.execute(stmt)
        return list(result.scalars().all())


async def get_account_overview_counts(tenant_id: int) -> Dict[str, Any]:
    """Get counts for account overview report."""
    async with async_get_session() as session:
        # Count organizations
        org_stmt = (
            select(func.count(Organization.id))
            .join(Account, Organization.account_id == Account.id)
            .where(Account.tenant_id == tenant_id)
        )
        org_result = await session.execute(org_stmt)
        org_count = org_result.scalar() or 0

        # Count individuals
        ind_stmt = (
            select(func.count(Individual.id))
            .join(Account, Individual.account_id == Account.id)
            .where(Account.tenant_id == tenant_id)
        )
        ind_result = await session.execute(ind_stmt)
        ind_count = ind_result.scalar() or 0

        # Organizations by country
        country_stmt = (
            select(Organization.headquarters_country, func.count(Organization.id))
            .join(Account, Organization.account_id == Account.id)
            .where(Account.tenant_id == tenant_id)
            .group_by(Organization.headquarters_country)
        )
        country_result = await session.execute(country_stmt)
        by_country = {row[0] or "Unknown": row[1] for row in country_result.all()}

        # Organizations by company type
        type_stmt = (
            select(Organization.company_type, func.count(Organization.id))
            .join(Account, Organization.account_id == Account.id)
            .where(Account.tenant_id == tenant_id)
            .group_by(Organization.company_type)
        )
        type_result = await session.execute(type_stmt)
        by_type = {row[0] or "Unknown": row[1] for row in type_result.all()}

        return {
            "total_organizations": org_count,
            "total_individuals": ind_count,
            "organizations_by_country": by_country,
            "organizations_by_type": by_type,
        }


async def get_organizations_with_contacts(tenant_id: int) -> List:
    """Get organizations with their contacts for coverage report."""
    async with async_get_session() as session:
        stmt = (
            select(Organization)
            .options(selectinload(Organization.account).selectinload(Account.contacts))
            .join(Account, Organization.account_id == Account.id)
            .where(Account.tenant_id == tenant_id)
        )
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def get_notes_for_activity_report(
    tenant_id: int, project_id: Optional[int] = None
) -> List:
    """Get notes for activity report."""
    async with async_get_session() as session:
        stmt = select(Note).where(Note.tenant_id == tenant_id)

        if project_id is not None:
            lead_ids_subq = (
                select(LeadProject.lead_id)
                .where(LeadProject.project_id == project_id)
                .scalar_subquery()
            )
            stmt = stmt.where(Note.entity_type == "Lead", Note.entity_id.in_(lead_ids_subq))

        stmt = stmt.order_by(Note.created_at.desc())
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def get_entity_name_for_note(entity_type: str, entity_id: int) -> Optional[str]:
    """Get the entity name for a note."""
    entity = await fetch_polymorphic_entity(entity_type, entity_id)
    if not entity:
        return None

    if entity_type == "Organization":
        return entity.name
    elif entity_type == "Individual":
        first = entity.first_name or ""
        last = entity.last_name or ""
        return f"{first} {last}".strip() or None
    elif entity_type == "Contact":
        first = getattr(entity, "first_name", "") or ""
        last = getattr(entity, "last_name", "") or ""
        return f"{first} {last}".strip() or None
    elif entity_type == "Lead":
        return entity.title
    return None


async def get_lead_type_for_note(entity_type: str, entity_id: int) -> Optional[str]:
    """Get the lead type for a Lead note."""
    if entity_type != "Lead":
        return None
    lead = await fetch_polymorphic_entity("Lead", entity_id)
    return lead.type.lower() if lead and lead.type else None


async def get_user_by_id(user_id: int) -> Optional[Dict[str, Any]]:
    """Get user info for audit report."""
    async with async_get_session() as session:
        stmt = select(User).where(User.id == user_id)
        result = await session.execute(stmt)
        user = result.scalar_one_or_none()
        if user:
            return {
                "id": user.id,
                "email": user.email,
                "name": f"{user.first_name or ''} {user.last_name or ''}".strip() or user.email,
            }
        return None


def get_audit_models() -> list:
    """Dynamically discover all models with created_by/updated_by columns (not relationships)."""
    from sqlalchemy.orm import ColumnProperty
    audit_models = []
    for mapper in Base.registry.mappers:
        model = mapper.class_
        # Check if created_by and updated_by are actual columns, not relationships
        has_created_by_col = any(
            isinstance(prop, ColumnProperty) and prop.key == "created_by"
            for prop in mapper.iterate_properties
        )
        has_updated_by_col = any(
            isinstance(prop, ColumnProperty) and prop.key == "updated_by"
            for prop in mapper.iterate_properties
        )
        if has_created_by_col and has_updated_by_col:
            table_name = model.__name__
            audit_models.append((table_name, model))
    audit_models.sort(key=lambda x: x[0])
    return audit_models


async def get_audit_counts_by_table(
    tenant_id: int, user_id: Optional[int] = None
) -> tuple:
    """Get audit counts by table."""
    audit_models = get_audit_models()
    by_table = []
    total_created = 0
    total_updated = 0

    async with async_get_session() as session:
        for table_name, model in audit_models:
            # Count created_by
            created_stmt = select(func.count(model.id)).where(model.created_by != None)
            if hasattr(model, "tenant_id"):
                created_stmt = created_stmt.where(model.tenant_id == tenant_id)
            if user_id:
                created_stmt = created_stmt.where(model.created_by == user_id)

            created_result = await session.execute(created_stmt)
            created_count = created_result.scalar() or 0

            # Count updated_by
            updated_stmt = select(func.count(model.id)).where(model.updated_by != None)
            if hasattr(model, "tenant_id"):
                updated_stmt = updated_stmt.where(model.tenant_id == tenant_id)
            if user_id:
                updated_stmt = updated_stmt.where(model.updated_by == user_id)

            updated_result = await session.execute(updated_stmt)
            updated_count = updated_result.scalar() or 0

            if created_count > 0 or updated_count > 0:
                by_table.append({
                    "table": table_name,
                    "created_count": created_count,
                    "updated_count": updated_count,
                })
                total_created += created_count
                total_updated += updated_count

    return by_table, total_created, total_updated


async def get_recent_audit_activity(
    tenant_id: int, user_id: Optional[int] = None, limit_per_table: int = 5
) -> List[Dict[str, Any]]:
    """Get recent audit activity across all tables."""
    audit_models = get_audit_models()
    recent_activity = []

    async with async_get_session() as session:
        for table_name, model in audit_models:
            if not hasattr(model, "updated_at"):
                continue

            stmt = select(model).where(model.updated_by.isnot(None))
            if hasattr(model, "tenant_id"):
                stmt = stmt.where(model.tenant_id == tenant_id)
            if user_id:
                stmt = stmt.where(model.updated_by == user_id)
            stmt = stmt.order_by(model.updated_at.desc()).limit(limit_per_table)

            result = await session.execute(stmt)
            records = result.scalars().all()

            for record in records:
                display_name = _get_record_display_name(record, table_name)
                recent_activity.append({
                    "table": table_name,
                    "id": record.id,
                    "display_name": display_name,
                    "updated_by": record.updated_by,
                    "updated_at": record.updated_at.isoformat() if record.updated_at else None,
                })

    return recent_activity


def _get_record_display_name(record, table_name: str) -> str:
    """Get a display name for a record based on its type."""
    if table_name in ("Individual", "Contact"):
        first = getattr(record, "first_name", "") or ""
        last = getattr(record, "last_name", "") or ""
        return f"{first} {last}".strip() or f"#{record.id}"
    elif hasattr(record, "name"):
        return record.name or f"#{record.id}"
    elif hasattr(record, "title"):
        return record.title or f"#{record.id}"
    elif hasattr(record, "filename"):
        return record.filename or f"#{record.id}"
    elif hasattr(record, "content"):
        content = record.content or ""
        return content[:50] + "..." if len(content) > 50 else content or f"#{record.id}"
    return f"#{record.id}"
