"""Serializers for partnership-related models."""

from typing import Dict, Any

from lib.date_utils import format_date, format_datetime, format_display_date
from serializers.base import model_to_dict


def extract_company_info(organizations: list, individuals: list) -> tuple[str, str]:
    """Extract company name and type from account data."""
    if organizations:
        return organizations[0].get("name", ""), "Organization"
    if individuals:
        ind = individuals[0]
        first_name = ind.get("first_name", "")
        last_name = ind.get("last_name", "")
        return f"{last_name}, {first_name}".strip(", "), "Individual"
    return "", "Organization"


def get_account_contacts(partnership) -> list:
    """Get contacts from partnership's account."""
    if not partnership.lead or not partnership.lead.account:
        return []

    if partnership.lead.account.organizations:
        return partnership.lead.account.organizations[0].contacts or []

    if partnership.lead.account.individuals:
        return partnership.lead.account.individuals[0].contacts or []

    return []


def build_contact_dict(contact) -> dict:
    """Build contact dictionary from ORM contact."""
    first_name = contact.first_name or ""
    last_name = contact.last_name or ""

    if not first_name and contact.individuals:
        first_name = contact.individuals[0].first_name or ""
    if not last_name and contact.individuals:
        last_name = contact.individuals[0].last_name or ""

    return {
        "id": contact.id,
        "first_name": first_name,
        "last_name": last_name,
        "email": contact.email or "",
        "phone": contact.phone or "",
        "title": contact.title,
        "is_primary": contact.is_primary,
        "created_at": contact.created_at.isoformat() if contact.created_at else None,
        "updated_at": contact.updated_at.isoformat() if contact.updated_at else None,
    }


def get_industries(partnership) -> list:
    """Get industry names from partnership's account."""
    if not partnership.lead or not partnership.lead.account or not partnership.lead.account.industries:
        return []
    return [ind.name for ind in partnership.lead.account.industries]


def partnership_to_list_response(partnership) -> Dict[str, Any]:
    """Convert partnership ORM to dict - optimized for list/table view."""
    p_dict = model_to_dict(partnership, include_relationships=True)

    lead = p_dict.get("lead") or {}
    account = lead.get("account") or {}
    company, _ = extract_company_info(
        account.get("organizations") or [],
        account.get("individuals") or [],
    )

    status = "Exploring"
    if partnership.lead and partnership.lead.current_status:
        status = partnership.lead.current_status.name

    return {
        "id": str(p_dict.get("id", "")),
        "account": company,
        "partnership_name": p_dict.get("partnership_name", ""),
        "partnership_type": p_dict.get("partnership_type"),
        "status": status,
        "start_date": format_date(p_dict.get("start_date")),
        "formatted_start_date": format_display_date(p_dict.get("start_date")),
        "end_date": format_date(p_dict.get("end_date")),
        "formatted_end_date": format_display_date(p_dict.get("end_date")),
        "updated_at": format_datetime(p_dict.get("updated_at")),
        "formatted_updated_at": format_display_date(p_dict.get("updated_at")),
    }


def partnership_to_detail_response(partnership) -> Dict[str, Any]:
    """Convert partnership ORM to response dictionary - full detail view."""
    p_dict = model_to_dict(partnership, include_relationships=True)

    lead = p_dict.get("lead") or {}
    account = lead.get("account") or {}
    company, company_type = extract_company_info(
        account.get("organizations") or [],
        account.get("individuals") or [],
    )

    contacts = [build_contact_dict(c) for c in get_account_contacts(partnership)]

    status = "Exploring"
    if partnership.lead and partnership.lead.current_status:
        status = partnership.lead.current_status.name

    created_at = p_dict.get("created_at", "")

    project_ids = []
    if partnership.lead and partnership.lead.projects:
        project_ids = [p.id for p in partnership.lead.projects]

    return {
        "id": str(p_dict.get("id", "")),
        "account": company,
        "account_type": company_type,
        "title": lead.get("title", ""),
        "partnership_name": p_dict.get("partnership_name", ""),
        "partnership_type": p_dict.get("partnership_type"),
        "start_date": format_date(p_dict.get("start_date")),
        "formatted_start_date": format_display_date(p_dict.get("start_date")),
        "end_date": format_date(p_dict.get("end_date")),
        "formatted_end_date": format_display_date(p_dict.get("end_date")),
        "description": p_dict.get("description"),
        "status": status,
        "source": lead.get("source", "manual"),
        "created_at": format_datetime(created_at),
        "updated_at": format_datetime(p_dict.get("updated_at")),
        "contacts": contacts,
        "industry": get_industries(partnership),
        "project_ids": project_ids,
    }
