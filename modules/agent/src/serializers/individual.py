"""Serializers for individual-related models."""

from serializers.common import contact_to_dict
from serializers.account import get_account_relationships


def get_industries(ind) -> list:
    """Extract industry names from individual's account."""
    if not ind.account or not ind.account.industries:
        return []
    return [industry.name for industry in ind.account.industries]


def get_projects(ind) -> list:
    """Extract projects info from individual's account."""
    if not ind.account or not ind.account.projects:
        return []
    return [{"id": p.id, "name": p.name} for p in ind.account.projects]


def ind_to_list_response(ind) -> dict:
    """Convert individual ORM to response dictionary - optimized for list/table view."""
    return {
        "id": ind.id,
        "first_name": ind.first_name,
        "last_name": ind.last_name,
        "title": ind.title,
        "industries": get_industries(ind),
        "email": ind.email,
        "phone": ind.phone,
        "updated_at": ind.updated_at.isoformat() if ind.updated_at else None,
    }


def ind_to_search_response(ind) -> dict:
    """Convert individual ORM to response dictionary - minimal for search/autocomplete."""
    return {
        "id": ind.id,
        "first_name": ind.first_name,
        "last_name": ind.last_name,
        "email": ind.email,
    }


def ind_to_detail_response(ind, contacts=None, include_research=False) -> dict:
    """Convert individual ORM to full response dictionary - for detail/edit views."""
    result = {
        "id": ind.id,
        "account_id": ind.account_id,
        "first_name": ind.first_name,
        "last_name": ind.last_name,
        "email": ind.email,
        "phone": ind.phone,
        "linkedin_url": ind.linkedin_url,
        "title": ind.title,
        "description": ind.description,
        "industries": get_industries(ind),
        "projects": get_projects(ind),
        "twitter_url": ind.twitter_url,
        "github_url": ind.github_url,
        "bio": ind.bio,
        "seniority_level": ind.seniority_level,
        "department": ind.department,
        "is_decision_maker": ind.is_decision_maker,
        "reports_to_id": ind.reports_to_id,
        "created_at": ind.created_at.isoformat() if ind.created_at else None,
        "updated_at": ind.updated_at.isoformat() if ind.updated_at else None,
    }

    if contacts is not None:
        result["contacts"] = [contact_to_dict(c) for c in contacts]

    # Get relationships from Account level
    if ind.account:
        result["relationships"] = get_account_relationships(ind.account, ind.account_id)
    else:
        result["relationships"] = []

    if include_research:
        if ind.reports_to:
            result["reports_to"] = {
                "id": ind.reports_to.id,
                "first_name": ind.reports_to.first_name,
                "last_name": ind.reports_to.last_name,
            }
        result["direct_reports"] = [
            {"id": dr.id, "first_name": dr.first_name, "last_name": dr.last_name, "title": dr.title}
            for dr in (ind.direct_reports or [])
        ]

    return result
