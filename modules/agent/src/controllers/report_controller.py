"""Controller layer for report routing to services."""

import logging
from typing import Optional, Dict, Any, List

from fastapi import Request, HTTPException

from lib.decorators import handle_http_exceptions
from serializers.report import saved_report_to_dict
from services.saved_report_service import (
    get_saved_reports,
    get_saved_report_by_id,
    create_report,
    update_report,
    delete_report,
    ReportQueryRequest,
    SavedReportCreate,
    SavedReportUpdate,
)
from schemas.report import SavedReportCreate, SavedReportUpdate
from services.report_builder import (
    execute_custom_query,
    get_lead_pipeline_report,
    get_account_overview_report,
    get_contact_coverage_report,
    get_notes_activity_report,
    get_user_audit_report,
    get_available_fields,
    generate_excel_bytes,
)

# Re-export for routes
__all__ = ["ReportQueryRequest", "SavedReportCreate", "SavedReportUpdate"]

# --- Canned Reports ---
@handle_http_exceptions
async def get_lead_pipeline(
    request: Request,
    date_from: Optional[str],
    date_to: Optional[str],
    project_id: Optional[int],
) -> Dict[str, Any]:
    """Get lead pipeline canned report."""
    tenant_id = request.state.tenant_id
    return await get_lead_pipeline_report(tenant_id, date_from, date_to, project_id)


@handle_http_exceptions
async def get_account_overview(request: Request) -> Dict[str, Any]:
    """Get account overview canned report."""
    tenant_id = request.state.tenant_id
    return await get_account_overview_report(tenant_id)


@handle_http_exceptions
async def get_contact_coverage(request: Request) -> Dict[str, Any]:
    """Get contact coverage canned report."""
    tenant_id = request.state.tenant_id
    return await get_contact_coverage_report(tenant_id)


@handle_http_exceptions
async def get_notes_activity(request: Request, project_id: Optional[int]) -> Dict[str, Any]:
    """Get notes activity canned report."""
    tenant_id = request.state.tenant_id
    return await get_notes_activity_report(tenant_id, project_id)


@handle_http_exceptions
async def get_user_audit(request: Request, user_id: Optional[int]) -> Dict[str, Any]:
    """Get user audit canned report."""
    tenant_id = request.state.tenant_id
    return await get_user_audit_report(tenant_id, user_id)


# --- Custom Report Execution ---

def get_entity_fields(entity: str) -> Dict[str, Any]:
    """Get available fields for an entity."""
    fields = get_available_fields(entity)
    if not fields["base"]:
        raise HTTPException(status_code=404, detail=f"Unknown entity: {entity}")
    return fields


@handle_http_exceptions
async def preview_custom_report(request: Request, query: ReportQueryRequest) -> Dict[str, Any]:
    """Preview a custom report with limited rows."""
    tenant_id = request.state.tenant_id
    query.limit = min(query.limit, 100)
    filters = [f.model_dump() for f in query.filters] if query.filters else None
    return await execute_custom_query(
        tenant_id=tenant_id,
        primary_entity=query.primary_entity,
        columns=query.columns,
        joins=query.joins,
        filters=filters,
        group_by=query.group_by,
        limit=query.limit,
        offset=query.offset,
        project_id=query.project_id,
    )


@handle_http_exceptions
async def run_custom_report(request: Request, query: ReportQueryRequest) -> Dict[str, Any]:
    """Run a custom report with full results."""
    tenant_id = request.state.tenant_id
    filters = [f.model_dump() for f in query.filters] if query.filters else None
    return await execute_custom_query(
        tenant_id=tenant_id,
        primary_entity=query.primary_entity,
        columns=query.columns,
        joins=query.joins,
        filters=filters,
        group_by=query.group_by,
        limit=query.limit,
        offset=query.offset,
        project_id=query.project_id,
    )


@handle_http_exceptions
async def export_custom_report(request: Request, query: ReportQueryRequest) -> tuple:
    """Export a custom report to Excel. Returns (excel_bytes, columns)."""
    tenant_id = request.state.tenant_id
    query.limit = 10000
    query.offset = 0
    filters = [f.model_dump() for f in query.filters] if query.filters else None
    result = await execute_custom_query(
        tenant_id=tenant_id,
        primary_entity=query.primary_entity,
        columns=query.columns,
        joins=query.joins,
        filters=filters,
        group_by=query.group_by,
        limit=query.limit,
        offset=query.offset,
        project_id=query.project_id,
    )
    if "error" in result:
        raise HTTPException(status_code=400, detail=result["error"])

    excel_bytes = generate_excel_bytes(result["data"], query.columns)
    return excel_bytes


# --- Saved Reports CRUD ---

@handle_http_exceptions
async def list_saved_reports(
    request: Request,
    project_id: Optional[int],
    include_global: bool,
    search: Optional[str],
    page: int,
    limit: int,
    sort_by: str,
    order: str,
) -> Dict[str, Any]:
    """List saved reports for the tenant with pagination and search."""
    tenant_id = request.state.tenant_id
    result = await get_saved_reports(
        tenant_id,
        project_id,
        include_global,
        search=search,
        page=page,
        limit=limit,
        sort_by=sort_by,
        sort_direction=order
    )
    return {
        "items": [saved_report_to_dict(r) for r in result["items"]],
        "total": result["total"],
        "currentPage": result["page"],
        "totalPages": result["total_pages"],
        "pageSize": result["limit"]
    }


@handle_http_exceptions
async def create_saved_report_controller(request: Request, data: SavedReportCreate) -> Dict[str, Any]:
    """Create a new saved report."""
    tenant_id = request.state.tenant_id
    user_id = getattr(request.state, "user_id", None)
    report = await create_report(
        {
            "tenant_id": tenant_id,
            "name": data.name,
            "description": data.description,
            "query_definition": data.query_definition,
            "created_by": user_id,
        },
        project_ids=data.project_ids
    )
    return saved_report_to_dict(report)


@handle_http_exceptions
async def get_saved_report(request: Request, report_id: int) -> Dict[str, Any]:
    """Get a saved report by ID."""
    tenant_id = request.state.tenant_id
    report = await get_saved_report_by_id(report_id, tenant_id)
    if not report:
        raise HTTPException(status_code=404, detail="Report not found")
    return saved_report_to_dict(report)


@handle_http_exceptions
async def update_saved_report_controller(request: Request, report_id: int, data: SavedReportUpdate) -> Dict[str, Any]:
    """Update a saved report."""
    tenant_id = request.state.tenant_id
    project_ids = data.project_ids
    update_data = {
        k: v for k, v in data.model_dump().items()
        if v is not None and k != "project_ids"
    }
    if not update_data and project_ids is None:
        raise HTTPException(status_code=400, detail="No fields to update")
    report = await update_report(report_id, tenant_id, update_data, project_ids)
    if not report:
        raise HTTPException(status_code=404, detail="Report not found")
    return saved_report_to_dict(report)


@handle_http_exceptions
async def delete_saved_report_controller(request: Request, report_id: int) -> Dict[str, Any]:
    """Delete a saved report."""
    tenant_id = request.state.tenant_id
    deleted = await delete_report(report_id, tenant_id)
    if not deleted:
        raise HTTPException(status_code=404, detail="Report not found")
    return {"message": "Report deleted"}
