"""Serializers for organization-related models."""

from serializers.common import contact_to_dict
from serializers.account import get_account_relationships


def technology_to_dict(org_tech) -> dict:
    """Convert organization technology ORM to response dictionary."""
    return {
        "organization_id": org_tech.organization_id,
        "technology_id": org_tech.technology_id,
        "technology": {
            "id": org_tech.technology.id,
            "name": org_tech.technology.name,
            "category": org_tech.technology.category,
            "vendor": org_tech.technology.vendor,
        } if org_tech.technology else None,
        "detected_at": org_tech.detected_at.isoformat() if org_tech.detected_at else None,
        "source": org_tech.source,
        "confidence": float(org_tech.confidence) if org_tech.confidence else None,
        "created_at": org_tech.created_at.isoformat() if org_tech.created_at else None,
        "updated_at": org_tech.updated_at.isoformat() if org_tech.updated_at else None,
    }


def funding_round_to_dict(fr) -> dict:
    """Convert funding round ORM to response dictionary."""
    return {
        "id": fr.id,
        "organization_id": fr.organization_id,
        "round_type": fr.round_type,
        "amount": fr.amount,
        "announced_date": fr.announced_date.isoformat() if fr.announced_date else None,
        "lead_investor": fr.lead_investor,
        "source_url": fr.source_url,
        "created_at": fr.created_at.isoformat() if fr.created_at else None,
        "updated_at": fr.updated_at.isoformat() if fr.updated_at else None,
    }


def signal_to_dict(s) -> dict:
    """Convert signal ORM to response dictionary."""
    return {
        "id": s.id,
        "account_id": s.account_id,
        "signal_type": s.signal_type,
        "headline": s.headline,
        "description": s.description,
        "signal_date": s.signal_date.isoformat() if s.signal_date else None,
        "source": s.source,
        "source_url": s.source_url,
        "sentiment": s.sentiment,
        "relevance_score": float(s.relevance_score) if s.relevance_score else None,
        "created_at": s.created_at.isoformat() if s.created_at else None,
        "updated_at": s.updated_at.isoformat() if s.updated_at else None,
    }


def org_to_list_response(org) -> dict:
    """Convert organization ORM to response dictionary - optimized for list/table view."""
    industries = []
    if org.account and org.account.industries:
        industries = [ind.name for ind in org.account.industries]

    return {
        "id": org.id,
        "name": org.name,
        "website": org.website,
        "industries": industries,
        "employee_count_range": org.employee_count_range.label if org.employee_count_range else None,
        "funding_stage": org.funding_stage.label if org.funding_stage else None,
        "description": org.description,
    }


def org_to_search_response(org) -> dict:
    """Convert organization ORM to response dictionary - minimal for search/autocomplete."""
    industries = []
    if org.account and org.account.industries:
        industries = [ind.name for ind in org.account.industries]

    return {
        "id": org.id,
        "name": org.name,
        "industries": industries,
    }


def org_to_detail_response(org, contacts=None, include_research=False) -> dict:
    """Convert organization ORM to full response dictionary - for detail/edit views."""
    industries = []
    if org.account and org.account.industries:
        industries = [ind.name for ind in org.account.industries]

    projects = []
    if org.account and org.account.projects:
        projects = [{"id": p.id, "name": p.name} for p in org.account.projects]

    result = {
        "id": org.id,
        "account_id": org.account_id,
        "name": org.name,
        "website": org.website,
        "phone": org.phone,
        "industries": industries,
        "projects": projects,
        "employee_count": org.employee_count,
        "employee_count_range_id": org.employee_count_range_id,
        "employee_count_range": org.employee_count_range.label if org.employee_count_range else None,
        "funding_stage_id": org.funding_stage_id,
        "funding_stage": org.funding_stage.label if org.funding_stage else None,
        "description": org.description,
        "revenue_range_id": org.revenue_range_id,
        "revenue_range": org.revenue_range.label if org.revenue_range else None,
        "founding_year": org.founding_year,
        "headquarters_city": org.headquarters_city,
        "headquarters_state": org.headquarters_state,
        "headquarters_country": org.headquarters_country,
        "company_type": org.company_type,
        "linkedin_url": org.linkedin_url,
        "crunchbase_url": org.crunchbase_url,
        "created_at": org.created_at.isoformat() if org.created_at else None,
        "updated_at": org.updated_at.isoformat() if org.updated_at else None,
    }

    if contacts is not None:
        result["contacts"] = [contact_to_dict(c) for c in contacts]

    # Get relationships from Account level
    if org.account:
        result["relationships"] = get_account_relationships(org.account, org.account_id)
    else:
        result["relationships"] = []

    if include_research:
        result["technologies"] = [technology_to_dict(t) for t in (org.technologies or [])]
        result["funding_rounds"] = [funding_round_to_dict(fr) for fr in (org.funding_rounds or [])]
        # Signals are now on the Account level
        account_signals = org.account.signals if org.account else []
        result["signals"] = [signal_to_dict(s) for s in (account_signals or [])]

    return result
