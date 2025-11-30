"""Routes for reports API endpoints."""

from typing import Optional, List, Any, Literal

from fastapi import APIRouter, Request, HTTPException, Query
from fastapi.responses import StreamingResponse
from pydantic import BaseModel
from io import BytesIO

from lib.auth import require_auth
from lib.error_logging import log_errors
from repositories.saved_report_repository import (
    find_all_saved_reports,
    find_saved_report_by_id,
    create_saved_report,
    update_saved_report,
    delete_saved_report,
)
from services.report_builder import (
    execute_custom_query,
    get_lead_pipeline_report,
    get_account_overview_report,
    get_contact_coverage_report,
    get_notes_activity_report,
    get_available_fields,
    generate_excel_bytes,
)


router = APIRouter(prefix="/reports", tags=["reports"])


# --- Pydantic Models ---

class ReportFilter(BaseModel):
    field: str
    operator: Literal["eq", "neq", "gt", "gte", "lt", "lte", "contains", "starts_with", "is_null", "is_not_null", "in"]
    value: Any = None


class Aggregation(BaseModel):
    function: Literal["count", "sum", "avg", "min", "max"]
    field: str
    alias: str


class ReportQueryRequest(BaseModel):
    primary_entity: Literal["organizations", "individuals", "contacts", "leads", "notes"]
    columns: List[str]
    joins: Optional[List[str]] = None
    filters: Optional[List[ReportFilter]] = None
    group_by: Optional[List[str]] = None
    aggregations: Optional[List[Aggregation]] = None
    limit: int = 100
    offset: int = 0
    project_id: Optional[int] = None  # Filter data by project scope


class SavedReportCreate(BaseModel):
    name: str
    description: Optional[str] = None
    query_definition: dict
    project_ids: Optional[List[int]] = None  # Empty/None = global report


class SavedReportUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    query_definition: Optional[dict] = None
    project_ids: Optional[List[int]] = None  # Empty = global, None = don't update


# --- Helper Functions ---

def _saved_report_to_dict(report) -> dict:
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
        "created_by": report.created_by,
        "creator_name": f"{report.creator.first_name} {report.creator.last_name}" if report.creator else None,
        "created_at": report.created_at.isoformat() if report.created_at else None,
        "updated_at": report.updated_at.isoformat() if report.updated_at else None,
    }


# --- Canned Reports ---

@router.get("/canned/lead-pipeline")
@log_errors
@require_auth
async def get_lead_pipeline(
    request: Request,
    date_from: Optional[str] = None,
    date_to: Optional[str] = None,
    project_id: Optional[int] = None,
):
    """Get lead pipeline canned report.

    If project_id is provided, filters to leads in that project.
    If project_id is None, returns global report for all leads.
    """
    tenant_id = request.state.tenant_id
    return await get_lead_pipeline_report(tenant_id, date_from, date_to, project_id)


@router.get("/canned/account-overview")
@log_errors
@require_auth
async def get_account_overview(request: Request):
    """Get account overview canned report."""
    tenant_id = request.state.tenant_id
    return await get_account_overview_report(tenant_id)


@router.get("/canned/contact-coverage")
@log_errors
@require_auth
async def get_contact_coverage(request: Request):
    """Get contact coverage canned report."""
    tenant_id = request.state.tenant_id
    return await get_contact_coverage_report(tenant_id)


@router.get("/canned/notes-activity")
@log_errors
@require_auth
async def get_notes_activity(request: Request, project_id: Optional[int] = None):
    """Get notes activity canned report.

    If project_id is provided, filters to notes on leads in that project.
    If project_id is None, returns global report for all notes.
    """
    tenant_id = request.state.tenant_id
    return await get_notes_activity_report(tenant_id, project_id)


# --- Custom Report Execution ---

@router.get("/fields/{entity}")
@log_errors
@require_auth
async def get_entity_fields(request: Request, entity: str):
    """Get available fields for an entity."""
    fields = get_available_fields(entity)
    if not fields["base"]:
        raise HTTPException(status_code=404, detail=f"Unknown entity: {entity}")
    return fields


@router.post("/custom/preview")
@log_errors
@require_auth
async def preview_custom_report(request: Request, query: ReportQueryRequest):
    """Preview a custom report with limited rows."""
    tenant_id = request.state.tenant_id
    query.limit = min(query.limit, 100)
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
    return result


@router.post("/custom/run")
@log_errors
@require_auth
async def run_custom_report(request: Request, query: ReportQueryRequest):
    """Run a custom report with full results."""
    tenant_id = request.state.tenant_id
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
    return result


@router.post("/custom/export")
@log_errors
@require_auth
async def export_custom_report(request: Request, query: ReportQueryRequest):
    """Export a custom report to Excel."""
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
    return StreamingResponse(
        BytesIO(excel_bytes),
        media_type="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
        headers={"Content-Disposition": "attachment; filename=report.xlsx"}
    )


# --- Saved Reports CRUD ---

@router.get("/saved")
@log_errors
@require_auth
async def list_saved_reports(
    request: Request,
    project_id: Optional[int] = None,
    include_global: bool = True,
    q: Optional[str] = None,
    page: int = Query(1, ge=1),
    limit: int = Query(50, ge=1, le=100),
    sort_by: str = Query("updated_at"),
    order: str = Query("DESC", pattern="^(ASC|DESC)$")
):
    """List saved reports for the tenant with pagination and search.

    Args:
        project_id: Filter by project ID. If None, only global reports are returned.
        include_global: If True, also include global reports (project_id IS NULL).
        q: Search string for name/description
        page: Page number (1-indexed)
        limit: Items per page
        sort_by: Column to sort by
        order: ASC or DESC
    """
    tenant_id = request.state.tenant_id
    result = await find_all_saved_reports(
        tenant_id,
        project_id,
        include_global,
        search=q,
        page=page,
        limit=limit,
        sort_by=sort_by,
        sort_direction=order
    )
    return {
        "items": [_saved_report_to_dict(r) for r in result["items"]],
        "total": result["total"],
        "currentPage": result["page"],
        "totalPages": result["total_pages"],
        "pageSize": result["limit"]
    }


@router.post("/saved")
@log_errors
@require_auth
async def create_saved_report_route(request: Request, data: SavedReportCreate):
    """Create a new saved report."""
    tenant_id = request.state.tenant_id
    user_id = getattr(request.state, "user_id", None)
    report = await create_saved_report(
        {
            "tenant_id": tenant_id,
            "name": data.name,
            "description": data.description,
            "query_definition": data.query_definition,
            "created_by": user_id,
        },
        project_ids=data.project_ids
    )
    return _saved_report_to_dict(report)


@router.get("/saved/{report_id}")
@log_errors
@require_auth
async def get_saved_report_route(request: Request, report_id: int):
    """Get a saved report by ID."""
    tenant_id = request.state.tenant_id
    report = await find_saved_report_by_id(report_id, tenant_id)
    if not report:
        raise HTTPException(status_code=404, detail="Report not found")
    return _saved_report_to_dict(report)


@router.put("/saved/{report_id}")
@log_errors
@require_auth
async def update_saved_report_route(request: Request, report_id: int, data: SavedReportUpdate):
    """Update a saved report."""
    tenant_id = request.state.tenant_id
    # Separate project_ids from other update data
    project_ids = data.project_ids
    update_data = {
        k: v for k, v in data.model_dump().items()
        if v is not None and k != "project_ids"
    }
    # Allow empty update_data if only project_ids is being updated
    if not update_data and project_ids is None:
        raise HTTPException(status_code=400, detail="No fields to update")
    report = await update_saved_report(report_id, tenant_id, update_data, project_ids)
    if not report:
        raise HTTPException(status_code=404, detail="Report not found")
    return _saved_report_to_dict(report)


@router.delete("/saved/{report_id}")
@log_errors
@require_auth
async def delete_saved_report_route(request: Request, report_id: int):
    """Delete a saved report."""
    tenant_id = request.state.tenant_id
    deleted = await delete_saved_report(report_id, tenant_id)
    if not deleted:
        raise HTTPException(status_code=404, detail="Report not found")
    return {"message": "Report deleted"}
