"""Service for building and executing dynamic report queries."""

import logging
from typing import Any, Dict, List, Optional
from io import BytesIO

from sqlalchemy import select, func
from sqlalchemy.orm import selectinload

from models.Organization import Organization
from models.Individual import Individual
from models.Contact import Contact
from models.Lead import Lead
from models.Account import Account
from models.Note import Note
from models.LeadProject import LeadProject
from lib.db import async_get_session

logger = logging.getLogger(__name__)

ENTITY_MAP = {
    "organizations": Organization,
    "individuals": Individual,
    "contacts": Contact,
    "leads": Lead,
    "notes": Note,
}

ENTITY_FIELDS = {
    "organizations": [
        "id", "name", "website", "phone", "employee_count", "description",
        "founding_year", "headquarters_city", "headquarters_state", "headquarters_country",
        "company_type", "linkedin_url", "crunchbase_url", "created_at", "updated_at"
    ],
    "individuals": [
        "id", "first_name", "last_name", "email", "phone", "title", "department",
        "seniority_level", "linkedin_url", "twitter_url", "github_url", "bio",
        "is_decision_maker", "created_at", "updated_at"
    ],
    "contacts": [
        "id", "first_name", "last_name", "email", "phone", "title", "department",
        "role", "is_primary", "notes", "created_at", "updated_at"
    ],
    "leads": [
        "id", "title", "description", "source", "type", "created_at", "updated_at"
    ],
    "notes": [
        "id", "entity_type", "entity_id", "content", "created_at", "updated_at"
    ],
}

JOIN_FIELDS = {
    "organizations": {
        "account": ["account.name"],
        "employee_count_range": ["employee_count_range.label"],
        "funding_stage": ["funding_stage.name"],
        "revenue_range": ["revenue_range.label"],
    },
    "individuals": {
        "account": ["account.name"],
    },
    "contacts": {
        "organizations": ["organization.name"],
        "individuals": ["individual.first_name", "individual.last_name"],
    },
    "leads": {
        "account": ["account.name"],
    },
    "notes": {},
}

# Define available joins between entities
# Format: { primary_entity: { join_name: { target_entity, join_condition_type } } }
ENTITY_JOINS = {
    "notes": {
        "leads": {"entity_type": "Lead", "target": Lead, "fields": ["lead.title", "lead.source", "lead.type"]},
        "organizations": {"entity_type": "Organization", "target": Organization, "fields": ["organization.name", "organization.website"]},
        "individuals": {"entity_type": "Individual", "target": Individual, "fields": ["individual.first_name", "individual.last_name", "individual.email"]},
        "contacts": {"entity_type": "Contact", "target": Contact, "fields": ["contact.first_name", "contact.last_name", "contact.email"]},
    },
    "leads": {
        "notes": {"target": Note, "fields": ["note.content", "note.created_at"]},
    },
    "organizations": {
        "notes": {"target": Note, "fields": ["note.content", "note.created_at"]},
    },
    "individuals": {
        "notes": {"target": Note, "fields": ["note.content", "note.created_at"]},
    },
    "contacts": {
        "notes": {"target": Note, "fields": ["note.content", "note.created_at"]},
    },
}

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

# Python-side operators for filtering joined data in memory
PYTHON_OPERATORS = {
    "eq": lambda val, filter_val: val == filter_val,
    "neq": lambda val, filter_val: val != filter_val,
    "gt": lambda val, filter_val: val is not None and val > filter_val,
    "gte": lambda val, filter_val: val is not None and val >= filter_val,
    "lt": lambda val, filter_val: val is not None and val < filter_val,
    "lte": lambda val, filter_val: val is not None and val <= filter_val,
    "contains": lambda val, filter_val: val is not None and str(filter_val).lower() in str(val).lower(),
    "starts_with": lambda val, filter_val: val is not None and str(val).lower().startswith(str(filter_val).lower()),
    "is_null": lambda val, filter_val: val is None or val == "",
    "is_not_null": lambda val, filter_val: val is not None and val != "",
    "in": lambda val, filter_val: val in (filter_val if isinstance(filter_val, list) else [filter_val]),
}


def get_available_fields(entity: str) -> Dict[str, Any]:
    """Get available fields for an entity including join fields."""
    base_fields = ENTITY_FIELDS.get(entity, [])
    joins = JOIN_FIELDS.get(entity, {})
    join_field_list = []
    for join_fields in joins.values():
        join_field_list.extend(join_fields)

    # Add entity join fields
    entity_joins = ENTITY_JOINS.get(entity, {})
    available_joins = []
    for join_name, join_config in entity_joins.items():
        available_joins.append({
            "name": join_name,
            "fields": join_config["fields"],
        })
        join_field_list.extend(join_config["fields"])

    return {
        "base": base_fields,
        "joins": join_field_list,
        "available_joins": available_joins,
    }


def _apply_tenant_filter(stmt, model, primary_entity: str, tenant_id: int):
    """Apply tenant filter based on entity type."""
    if primary_entity == "notes":
        return stmt.where(model.tenant_id == tenant_id)
    if primary_entity == "contacts":
        return stmt  # Contacts visible to all for now
    # Organizations, individuals, leads - filter through Account
    return stmt.join(Account, model.account_id == Account.id).where(Account.tenant_id == tenant_id)


def _get_joined_model_and_alias(primary_entity: str, join_name: str):
    """Get the target model for a join."""
    entity_joins = ENTITY_JOINS.get(primary_entity, {})
    join_config = entity_joins.get(join_name)
    if not join_config:
        return None, None
    return join_config.get("target"), join_config


def _format_value(val: Any) -> Any:
    """Format a value, converting datetime objects to ISO format."""
    if hasattr(val, "isoformat"):
        return val.isoformat()
    return val


def _get_direct_field(row, col: str) -> Any:
    """Get a direct field value from a row."""
    if not hasattr(row, col):
        return None
    return _format_value(getattr(row, col))


def _get_joined_field(joined_rows: Dict[str, Any], rel_name: str, field_name: str) -> Any:
    """Get a field value from joined rows."""
    joined_row = joined_rows.get(rel_name)
    if not joined_row or not hasattr(joined_row, field_name):
        return None
    return _format_value(getattr(joined_row, field_name))


def _get_relationship_field(row, rel_name: str, field_name: str) -> Any:
    """Get a field value from a model relationship."""
    rel_obj = getattr(row, rel_name, None)
    if not rel_obj:
        return None
    return _format_value(getattr(rel_obj, field_name, None))


def _extract_field_value(row, col: str, joined_rows: Dict[str, Any]):
    """Extract a field value from row or joined data."""
    if "." not in col:
        return _get_direct_field(row, col)

    rel_name, field_name = col.split(".", 1)

    # Check joined rows first, fall back to relationship
    joined_val = _get_joined_field(joined_rows, rel_name, field_name)
    if joined_val is not None:
        return joined_val
    return _get_relationship_field(row, rel_name, field_name)


def _is_valid_primary_filter(f: Dict[str, Any], model) -> bool:
    """Check if a filter is valid for the primary entity."""
    field = f.get("field")
    operator = f.get("operator")
    if not field or not operator:
        return False
    if "." in field:
        return False  # Join field filters handled separately
    return hasattr(model, field)


def _apply_primary_filters(stmt, model, filters: List[Dict[str, Any]]):
    """Apply filters on primary entity columns."""
    valid_filters = [f for f in filters if _is_valid_primary_filter(f, model)]
    for f in valid_filters:
        col = getattr(model, f["field"])
        op_func = OPERATORS.get(f["operator"])
        if op_func:
            stmt = stmt.where(op_func(col, f.get("value")))
    return stmt


def _row_passes_join_filters(row_dict: Dict[str, Any], join_filters: List[Dict[str, Any]]) -> bool:
    """Check if a row passes all join field filters."""
    valid_filters = [f for f in join_filters if f.get("field") and f.get("operator")]
    for f in valid_filters:
        field_value = row_dict.get(f["field"])
        op_func = PYTHON_OPERATORS.get(f["operator"])
        if op_func and not op_func(field_value, f.get("value")):
            return False
    return True


ENTITY_TYPE_MAP = {
    "leads": "Lead",
    "organizations": "Organization",
    "individuals": "Individual",
    "contacts": "Contact",
}


async def _fetch_polymorphic_join(session, row, target_model, join_config, join_name) -> tuple:
    """Fetch polymorphic join for notes -> entity."""
    if row.entity_type != join_config["entity_type"]:
        return None, None
    join_stmt = select(target_model).where(target_model.id == row.entity_id)
    join_result = await session.execute(join_stmt)
    singular_key = join_name.rstrip("s") if join_name.endswith("s") else join_name
    return singular_key, join_result.scalar_one_or_none()


async def _fetch_notes_join(session, row, primary_entity: str, tenant_id: int) -> tuple:
    """Fetch notes join for entity -> notes."""
    entity_type = ENTITY_TYPE_MAP.get(primary_entity)
    if not entity_type:
        return None, None
    join_stmt = (
        select(Note)
        .where(
            Note.tenant_id == tenant_id,
            Note.entity_type == entity_type,
            Note.entity_id == row.id,
        )
        .order_by(Note.created_at.desc())
        .limit(1)
    )
    join_result = await session.execute(join_stmt)
    return "note", join_result.scalar_one_or_none()


async def _fetch_single_join(
    session, row, join_name: str, target_model, join_config, primary_entity: str, tenant_id: int
) -> tuple:
    """Fetch a single join based on entity type."""
    # Notes use polymorphic join via entity_type/entity_id
    if primary_entity == "notes" and join_config.get("entity_type"):
        return await _fetch_polymorphic_join(session, row, target_model, join_config, join_name)

    if join_name == "notes":
        return await _fetch_notes_join(session, row, primary_entity, tenant_id)

    return None, None


async def _fetch_joined_rows(
    session, row, joins: List[str], primary_entity: str, tenant_id: int
) -> Dict[str, Any]:
    """Fetch all joined data for a row."""
    joined_rows = {}
    valid_joins = [
        (name, *_get_joined_model_and_alias(primary_entity, name))
        for name in joins
    ]
    valid_joins = [(name, model, config) for name, model, config in valid_joins if model]

    for join_name, target_model, join_config in valid_joins:
        key, value = await _fetch_single_join(
            session, row, join_name, target_model, join_config, primary_entity, tenant_id
        )
        if key:
            joined_rows[key] = value

    return joined_rows


async def execute_custom_query(
    tenant_id: int,
    primary_entity: str,
    columns: List[str],
    joins: Optional[List[str]] = None,
    filters: Optional[List[Dict[str, Any]]] = None,
    group_by: Optional[List[str]] = None,
    aggregations: Optional[List[Dict[str, Any]]] = None,
    limit: int = 100,
    offset: int = 0,
    project_id: Optional[int] = None,
) -> Dict[str, Any]:
    """Execute a custom report query and return results."""
    model = ENTITY_MAP.get(primary_entity)
    if not model:
        return {"error": f"Unknown entity: {primary_entity}", "data": [], "total": 0}

    joins = joins or []

    async with async_get_session() as session:
        stmt = select(model)
        stmt = _apply_tenant_filter(stmt, model, primary_entity, tenant_id)

        # Apply project filter for leads
        if project_id is not None and primary_entity == "leads":
            stmt = stmt.join(
                LeadProject, Lead.id == LeadProject.lead_id
            ).where(LeadProject.project_id == project_id)

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
        rows = result.scalars().all()

        # Fetch joined data for each row
        data = []
        for row in rows:
            joined_rows = await _fetch_joined_rows(
                session, row, joins, primary_entity, tenant_id
            )
            row_dict = {col: _extract_field_value(row, col, joined_rows) for col in columns}

            # Apply join field filters (post-fetch filtering)
            join_filters = [f for f in (filters or []) if "." in f.get("field", "")]
            if _row_passes_join_filters(row_dict, join_filters):
                data.append(row_dict)

        return {"data": data, "total": len(data), "limit": limit, "offset": offset}


async def get_lead_pipeline_report(
    tenant_id: int,
    date_from: Optional[str] = None,
    date_to: Optional[str] = None,
    project_id: Optional[int] = None
) -> Dict[str, Any]:
    """Generate lead pipeline canned report.

    If project_id is provided, filters leads to only those in the project.
    If project_id is None, returns data for all leads (global report).
    """
    async with async_get_session() as session:
        # Base query for leads in tenant
        base_stmt = (
            select(Lead)
            .join(Account, Lead.account_id == Account.id)
            .where(Account.tenant_id == tenant_id)
        )

        # Filter by project if provided
        if project_id is not None:
            base_stmt = base_stmt.join(
                LeadProject, Lead.id == LeadProject.lead_id
            ).where(LeadProject.project_id == project_id)

        if date_from:
            base_stmt = base_stmt.where(Lead.created_at >= date_from)
        if date_to:
            base_stmt = base_stmt.where(Lead.created_at <= date_to)

        result = await session.execute(base_stmt)
        leads = result.scalars().all()

        # Count by type
        type_counts = {}
        source_counts = {}
        total = len(leads)

        for lead in leads:
            lead_type = lead.type or "Unknown"
            type_counts[lead_type] = type_counts.get(lead_type, 0) + 1

            source = lead.source or "Unknown"
            source_counts[source] = source_counts.get(source, 0) + 1

        return {
            "total_leads": total,
            "by_type": type_counts,
            "by_source": source_counts,
        }


async def get_account_overview_report(tenant_id: int) -> Dict[str, Any]:
    """Generate account overview canned report."""
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


async def get_contact_coverage_report(tenant_id: int) -> Dict[str, Any]:
    """Generate contact coverage canned report."""
    async with async_get_session() as session:
        # Get all organizations for tenant with contact counts
        org_stmt = (
            select(Organization)
            .options(selectinload(Organization.contacts))
            .join(Account, Organization.account_id == Account.id)
            .where(Account.tenant_id == tenant_id)
        )
        result = await session.execute(org_stmt)
        orgs = result.scalars().all()

        total_orgs = len(orgs)
        total_contacts = 0
        orgs_with_zero_contacts = 0
        decision_makers = 0
        total_contact_count = 0

        coverage_data = []
        for org in orgs:
            contact_count = len(org.contacts)
            total_contacts += contact_count
            if contact_count == 0:
                orgs_with_zero_contacts += 1
            for contact in org.contacts:
                total_contact_count += 1
                if getattr(contact, "is_decision_maker", False):
                    decision_makers += 1
            coverage_data.append({
                "organization_id": org.id,
                "organization_name": org.name,
                "contact_count": contact_count,
            })

        avg_contacts = total_contacts / total_orgs if total_orgs > 0 else 0
        decision_maker_ratio = decision_makers / total_contact_count if total_contact_count > 0 else 0

        return {
            "total_organizations": total_orgs,
            "total_contacts": total_contacts,
            "average_contacts_per_org": round(avg_contacts, 2),
            "organizations_with_zero_contacts": orgs_with_zero_contacts,
            "decision_maker_count": decision_makers,
            "decision_maker_ratio": round(decision_maker_ratio, 4),
            "coverage_by_org": sorted(coverage_data, key=lambda x: x["contact_count"], reverse=True)[:20],
        }


def _format_person_name(first: Optional[str], last: Optional[str]) -> Optional[str]:
    """Format a person's name from first and last name fields."""
    first = first or ""
    last = last or ""
    name = f"{first} {last}".strip()
    return name or None


# Entity name extractors for polymorphic note lookups
ENTITY_NAME_EXTRACTORS = {
    "Organization": lambda entity: entity.name if entity else None,
    "Individual": lambda entity: _format_person_name(entity.first_name, entity.last_name) if entity else None,
    "Contact": lambda entity: _format_person_name(
        getattr(entity, "first_name", ""), getattr(entity, "last_name", "")
    ) if entity else None,
    "Lead": lambda entity: entity.title if entity else None,
}

ENTITY_MODEL_MAP = {
    "Organization": Organization,
    "Individual": Individual,
    "Contact": Contact,
    "Lead": Lead,
}


async def _get_entity_name_for_note(session, note) -> Optional[str]:
    """Get the entity name for a note using dict dispatch."""
    model = ENTITY_MODEL_MAP.get(note.entity_type)
    if not model:
        return None
    result = await session.execute(select(model).where(model.id == note.entity_id))
    entity = result.scalar_one_or_none()
    extractor = ENTITY_NAME_EXTRACTORS.get(note.entity_type, lambda e: None)
    return extractor(entity)


async def get_notes_activity_report(tenant_id: int, project_id: Optional[int] = None) -> Dict[str, Any]:
    """Generate notes activity canned report.

    If project_id is provided, filters to notes on leads in that project.
    If project_id is None, returns data for all notes (global report).
    """
    async with async_get_session() as session:
        # Get notes for tenant
        notes_stmt = select(Note).where(Note.tenant_id == tenant_id)

        # Filter by project - only include notes on Leads that are in the project
        if project_id is not None:
            # Subquery to get lead IDs in the project
            lead_ids_subq = (
                select(LeadProject.lead_id)
                .where(LeadProject.project_id == project_id)
                .scalar_subquery()
            )
            notes_stmt = notes_stmt.where(
                Note.entity_type == "Lead",
                Note.entity_id.in_(lead_ids_subq)
            )

        notes_stmt = notes_stmt.order_by(Note.created_at.desc())
        result = await session.execute(notes_stmt)
        notes = result.scalars().all()

        total_notes = len(notes)

        # Count by entity type
        by_entity_type = {}
        for note in notes:
            entity_type = note.entity_type or "Unknown"
            by_entity_type[entity_type] = by_entity_type.get(entity_type, 0) + 1

        # Get entity counts that have notes
        entities_with_notes = {}
        entity_ids_seen = {}
        for note in notes:
            key = f"{note.entity_type}:{note.entity_id}"
            if key not in entity_ids_seen:
                entity_ids_seen[key] = True
                entity_type = note.entity_type or "Unknown"
                entities_with_notes[entity_type] = entities_with_notes.get(entity_type, 0) + 1

        # Get recent 10 notes with entity names
        recent_notes = []
        for note in notes[:10]:
            entity_name = await _get_entity_name_for_note(session, note)
            recent_notes.append({
                "id": note.id,
                "entity_type": note.entity_type,
                "entity_id": note.entity_id,
                "entity_name": entity_name,
                "content": note.content[:100] + "..." if note.content and len(note.content) > 100 else note.content,
                "created_at": note.created_at.isoformat() if note.created_at else None,
            })

        return {
            "total_notes": total_notes,
            "by_entity_type": by_entity_type,
            "entities_with_notes": entities_with_notes,
            "recent_notes": recent_notes,
        }


def generate_excel_bytes(data: List[Dict[str, Any]], columns: List[str]) -> bytes:
    """Generate Excel file bytes from report data."""
    try:
        import openpyxl
        from openpyxl import Workbook
    except ImportError:
        raise ImportError("openpyxl is required for Excel export. Install with: pip install openpyxl")

    wb = Workbook()
    ws = wb.active
    ws.title = "Report"

    # Write headers
    for col_idx, col_name in enumerate(columns, 1):
        ws.cell(row=1, column=col_idx, value=col_name)

    # Write data
    for row_idx, row_data in enumerate(data, 2):
        for col_idx, col_name in enumerate(columns, 1):
            value = row_data.get(col_name, "")
            ws.cell(row=row_idx, column=col_idx, value=value)

    # Save to bytes
    output = BytesIO()
    wb.save(output)
    output.seek(0)
    return output.getvalue()
