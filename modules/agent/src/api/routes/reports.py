"""Routes for reports API endpoints."""

from typing import Optional
from io import BytesIO

from fastapi import APIRouter, Request, Query
from fastapi.responses import StreamingResponse

from controllers.report_controller import (
    create_saved_report_controller,
    delete_saved_report_controller,
    export_custom_report,
    get_account_overview,
    get_contact_coverage,
    get_entity_fields,
    get_lead_pipeline,
    get_notes_activity,
    get_saved_report,
    get_user_audit,
    list_saved_reports,
    preview_custom_report,
    run_custom_report,
    update_saved_report_controller,
)
from schemas.report import ReportQueryRequest, SavedReportCreate, SavedReportUpdate


router = APIRouter(prefix="/reports", tags=["reports"])


# --- Canned Reports ---

@router.get("/canned/lead-pipeline")
async def get_lead_pipeline_route(
    request: Request,
    date_from: Optional[str] = None,
    date_to: Optional[str] = None,
    project_id: Optional[int] = None,
):
    """Get lead pipeline canned report."""
    return await get_lead_pipeline(request, date_from, date_to, project_id)


@router.get("/canned/account-overview")
async def get_account_overview_route(request: Request):
    """Get account overview canned report."""
    return await get_account_overview(request)


@router.get("/canned/contact-coverage")
async def get_contact_coverage_route(request: Request):
    """Get contact coverage canned report."""
    return await get_contact_coverage(request)


@router.get("/canned/notes-activity")
async def get_notes_activity_route(request: Request, project_id: Optional[int] = None):
    """Get notes activity canned report."""
    return await get_notes_activity(request, project_id)


@router.get("/canned/user-audit")
async def get_user_audit_route(request: Request, user_id: Optional[int] = None):
    """Get user audit canned report."""
    return await get_user_audit(request, user_id)


# --- Custom Report Execution ---

@router.get("/fields/{entity}")
async def get_entity_fields_route(request: Request, entity: str):
    """Get available fields for an entity."""
    return get_entity_fields(entity)


@router.post("/custom/preview")
async def preview_custom_report_route(request: Request, query: ReportQueryRequest):
    """Preview a custom report with limited rows."""
    return await preview_custom_report(request, query)


@router.post("/custom/run")
async def run_custom_report_route(request: Request, query: ReportQueryRequest):
    """Run a custom report with full results."""
    return await run_custom_report(request, query)


@router.post("/custom/export")
async def export_custom_report_route(request: Request, query: ReportQueryRequest):
    """Export a custom report to Excel."""
    excel_bytes = await export_custom_report(request, query)
    return StreamingResponse(
        BytesIO(excel_bytes),
        media_type="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
        headers={"Content-Disposition": "attachment; filename=report.xlsx"}
    )


# --- Saved Reports CRUD ---

@router.get("/saved")
async def list_saved_reports_route(
    request: Request,
    project_id: Optional[int] = None,
    include_global: bool = True,
    q: Optional[str] = None,
    page: int = Query(1, ge=1),
    limit: int = Query(50, ge=1, le=100),
    sort_by: str = Query("updated_at"),
    order: str = Query("DESC", pattern="^(ASC|DESC)$")
):
    """List saved reports for the tenant with pagination and search."""
    return await list_saved_reports(request, project_id, include_global, q, page, limit, sort_by, order)


@router.post("/saved")
async def create_saved_report_route(request: Request, data: SavedReportCreate):
    """Create a new saved report."""
    return await create_saved_report_controller(request, data)


@router.get("/saved/{report_id}")
async def get_saved_report_route(request: Request, report_id: int):
    """Get a saved report by ID."""
    return await get_saved_report(request, report_id)


@router.put("/saved/{report_id}")
async def update_saved_report_route(request: Request, report_id: int, data: SavedReportUpdate):
    """Update a saved report."""
    return await update_saved_report_controller(request, report_id, data)


@router.delete("/saved/{report_id}")
async def delete_saved_report_route(request: Request, report_id: int):
    """Delete a saved report."""
    return await delete_saved_report_controller(request, report_id)
