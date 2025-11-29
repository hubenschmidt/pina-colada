"""Service for building and executing dynamic report queries."""

import logging
from typing import Any, Dict, List, Optional
from io import BytesIO

from sqlalchemy import select, func, and_, or_, text
from sqlalchemy.orm import selectinload

from models.Organization import Organization
from models.Individual import Individual
from models.Contact import Contact
from models.Lead import Lead
from models.Account import Account
from lib.db import async_get_session

logger = logging.getLogger(__name__)

ENTITY_MAP = {
    "organizations": Organization,
    "individuals": Individual,
    "contacts": Contact,
    "leads": Lead,
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


def get_available_fields(entity: str) -> Dict[str, List[str]]:
    """Get available fields for an entity including join fields."""
    base_fields = ENTITY_FIELDS.get(entity, [])
    joins = JOIN_FIELDS.get(entity, {})
    join_field_list = []
    for join_fields in joins.values():
        join_field_list.extend(join_fields)
    return {
        "base": base_fields,
        "joins": join_field_list,
    }


async def execute_custom_query(
    tenant_id: int,
    primary_entity: str,
    columns: List[str],
    filters: Optional[List[Dict[str, Any]]] = None,
    group_by: Optional[List[str]] = None,
    aggregations: Optional[List[Dict[str, Any]]] = None,
    limit: int = 100,
    offset: int = 0,
) -> Dict[str, Any]:
    """Execute a custom report query and return results."""
    model = ENTITY_MAP.get(primary_entity)
    if not model:
        return {"error": f"Unknown entity: {primary_entity}", "data": [], "total": 0}

    async with async_get_session() as session:
        # Build base query
        stmt = select(model)

        # Apply tenant filter through Account relationship
        if primary_entity == "organizations":
            stmt = stmt.join(Account, model.account_id == Account.id).where(Account.tenant_id == tenant_id)
        elif primary_entity == "individuals":
            stmt = stmt.join(Account, model.account_id == Account.id).where(Account.tenant_id == tenant_id)
        elif primary_entity == "leads":
            stmt = stmt.join(Account, model.account_id == Account.id).where(Account.tenant_id == tenant_id)
        elif primary_entity == "contacts":
            # Contacts don't have direct tenant relationship, filter through organizations
            stmt = stmt.where(True)  # Contacts visible to all for now

        # Load relationships for join fields
        if primary_entity == "organizations":
            stmt = stmt.options(
                selectinload(Organization.account),
                selectinload(Organization.employee_count_range),
                selectinload(Organization.funding_stage),
                selectinload(Organization.revenue_range),
            )
        elif primary_entity == "individuals":
            stmt = stmt.options(selectinload(Individual.account))

        # Apply filters
        if filters:
            for f in filters:
                field = f.get("field")
                operator = f.get("operator")
                value = f.get("value")
                if not field or not operator:
                    continue
                if not hasattr(model, field):
                    continue
                col = getattr(model, field)
                op_func = OPERATORS.get(operator)
                if op_func:
                    stmt = stmt.where(op_func(col, value))

        # Get total count before pagination
        count_stmt = select(func.count()).select_from(stmt.subquery())
        count_result = await session.execute(count_stmt)
        total = count_result.scalar() or 0

        # Apply pagination
        stmt = stmt.limit(limit).offset(offset)

        # Execute query
        result = await session.execute(stmt)
        rows = result.scalars().all()

        # Transform to dicts with selected columns
        data = []
        for row in rows:
            row_dict = {}
            for col in columns:
                if "." in col:
                    # Handle join field
                    parts = col.split(".")
                    rel_name = parts[0]
                    field_name = parts[1]
                    rel_obj = getattr(row, rel_name, None)
                    row_dict[col] = getattr(rel_obj, field_name, None) if rel_obj else None
                elif hasattr(row, col):
                    val = getattr(row, col)
                    # Convert datetime to ISO string
                    if hasattr(val, "isoformat"):
                        val = val.isoformat()
                    row_dict[col] = val
            data.append(row_dict)

        return {"data": data, "total": total, "limit": limit, "offset": offset}


async def get_lead_pipeline_report(tenant_id: int, date_from: Optional[str] = None, date_to: Optional[str] = None) -> Dict[str, Any]:
    """Generate lead pipeline canned report."""
    async with async_get_session() as session:
        # Base query for leads in tenant
        base_stmt = (
            select(Lead)
            .join(Account, Lead.account_id == Account.id)
            .where(Account.tenant_id == tenant_id)
        )

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
