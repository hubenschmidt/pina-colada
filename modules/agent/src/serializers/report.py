"""Serializers for report-related models."""


def saved_report_to_dict(report) -> dict:
    """Convert saved report model to response dict."""
    project_ids = [p.id for p in report.projects] if report.projects else []
    project_names = [p.name for p in report.projects] if report.projects else []
    return {
        "id": report.id,
        "name": report.name,
        "description": report.description,
        "query_definition": report.query_definition,
        "project_ids": project_ids,
        "project_names": project_names,
        "is_global": len(project_ids) == 0,
        "created_at": report.created_at.isoformat() if report.created_at else None,
        "updated_at": report.updated_at.isoformat() if report.updated_at else None,
    }
