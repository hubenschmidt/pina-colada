"""Serializers for project-related models."""


def project_to_response(project, deals_count: int = 0, leads_count: int = 0) -> dict:
    """Convert project to dict."""
    return {
        "id": project.id,
        "name": project.name,
        "description": project.description,
        "status": project.status,
        "current_status_id": project.current_status_id,
        "start_date": str(project.start_date) if project.start_date else None,
        "end_date": str(project.end_date) if project.end_date else None,
        "created_at": project.created_at.isoformat() if project.created_at else None,
        "updated_at": project.updated_at.isoformat() if project.updated_at else None,
        "deals_count": deals_count,
        "leads_count": leads_count,
    }


def lead_to_response(lead) -> dict:
    """Convert lead to response dict."""
    return {
        "id": lead.id,
        "title": lead.title,
        "type": lead.type,
        "description": lead.description,
        "source": lead.source,
        "current_status": lead.current_status.name if lead.current_status else None,
        "account_name": lead.account.name if lead.account else None,
        "created_at": lead.created_at.isoformat() if lead.created_at else None,
    }


def deal_to_response(deal) -> dict:
    """Convert deal to response dict."""
    return {
        "id": deal.id,
        "name": deal.name,
        "description": deal.description,
        "current_status": deal.current_status.name if deal.current_status else None,
        "value_amount": float(deal.value_amount) if deal.value_amount else None,
        "value_currency": deal.value_currency,
        "created_at": deal.created_at.isoformat() if deal.created_at else None,
    }
