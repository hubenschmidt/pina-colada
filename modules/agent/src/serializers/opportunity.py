"""Serializers for opportunity-related models."""

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


def get_account_contacts(opp) -> list:
    """Get contacts from opportunity's account."""
    if not opp.lead or not opp.lead.account:
        return []
    return opp.lead.account.contacts or []


def build_contact_dict(contact) -> dict:
    """Build contact dictionary from ORM contact."""
    return {
        "id": contact.id,
        "first_name": contact.first_name or "",
        "last_name": contact.last_name or "",
        "email": contact.email or "",
        "phone": contact.phone or "",
        "title": contact.title,
        "is_primary": contact.is_primary,
        "created_at": contact.created_at.isoformat() if contact.created_at else None,
        "updated_at": contact.updated_at.isoformat() if contact.updated_at else None,
    }


def get_industries(opp) -> list:
    """Get industry names from opportunity's account."""
    if not opp.lead or not opp.lead.account or not opp.lead.account.industries:
        return []
    return [ind.name for ind in opp.lead.account.industries]


def opportunity_to_list_response(opp) -> Dict[str, Any]:
    """Convert opportunity ORM to dict - optimized for list/table view."""
    opp_dict = model_to_dict(opp, include_relationships=True)

    lead = opp_dict.get("lead") or {}
    account = lead.get("account") or {}
    company, _ = extract_company_info(
        account.get("organizations") or [],
        account.get("individuals") or [],
    )

    status = "Qualifying"
    if opp.lead and opp.lead.current_status:
        status = opp.lead.current_status.name

    return {
        "id": str(opp_dict.get("id", "")),
        "account": company,
        "opportunity_name": opp_dict.get("opportunity_name", ""),
        "estimated_value": float(opp_dict.get("estimated_value")) if opp_dict.get("estimated_value") else None,
        "probability": float(opp_dict.get("probability")) if opp_dict.get("probability") else None,
        "expected_close_date": format_date(opp_dict.get("expected_close_date")),
        "formatted_expected_close_date": format_display_date(opp_dict.get("expected_close_date")),
        "status": status,
        "updated_at": format_datetime(opp_dict.get("updated_at")),
        "formatted_updated_at": format_display_date(opp_dict.get("updated_at")),
    }


def opportunity_to_detail_response(opp) -> Dict[str, Any]:
    """Convert opportunity ORM to response dictionary - full detail view."""
    opp_dict = model_to_dict(opp, include_relationships=True)

    lead = opp_dict.get("lead") or {}
    account = lead.get("account") or {}
    company, company_type = extract_company_info(
        account.get("organizations") or [],
        account.get("individuals") or [],
    )

    contacts = [build_contact_dict(c) for c in get_account_contacts(opp)]

    status = "Qualifying"
    if opp.lead and opp.lead.current_status:
        status = opp.lead.current_status.name

    project_ids = []
    if opp.lead and opp.lead.projects:
        project_ids = [p.id for p in opp.lead.projects]

    return {
        "id": str(opp_dict.get("id", "")),
        "account": company,
        "account_type": company_type,
        "title": lead.get("title", ""),
        "opportunity_name": opp_dict.get("opportunity_name", ""),
        "estimated_value": float(opp_dict.get("estimated_value")) if opp_dict.get("estimated_value") else None,
        "probability": float(opp_dict.get("probability")) if opp_dict.get("probability") else None,
        "expected_close_date": format_date(opp_dict.get("expected_close_date")),
        "formatted_expected_close_date": format_display_date(opp_dict.get("expected_close_date")),
        "description": opp_dict.get("description"),
        "status": status,
        "source": lead.get("source", "manual"),
        "created_at": format_datetime(opp_dict.get("created_at")),
        "updated_at": format_datetime(opp_dict.get("updated_at")),
        "contacts": contacts,
        "industry": get_industries(opp),
        "project_ids": project_ids,
    }
