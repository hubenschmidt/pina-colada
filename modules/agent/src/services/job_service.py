"""Service layer for job business logic."""

import logging
from typing import Dict, List, Optional, Any
from datetime import datetime
from fastapi import HTTPException
from serializers.base import model_to_dict
from lib.validators import validate_phone
from lib.date_utils import parse_date_input
from repositories.job_repository import (
    find_all_jobs,
    create_job as create_job_repo,
    find_job_by_id,
    update_job as update_job_repo,
    delete_job as delete_job_repo,
    find_all_statuses,
    find_jobs_with_status,
    JobCreate,
    JobUpdate,
)

# Re-export Pydantic models for controllers
__all__ = ["JobCreate", "JobUpdate"]
from repositories.organization_repository import get_or_create_organization
from repositories.individual_repository import get_or_create_individual
from repositories.contact_repository import get_or_create_contact
from repositories.job_repository import find_all_jobs as find_all_jobs_repo
from repositories.deal_repository import get_or_create_deal
from repositories.status_repository import find_status_by_name
from repositories.industry_repository import find_industry_by_name

logger = logging.getLogger(__name__)


async def _resolve_individual_account(job_data: Dict[str, Any], tenant_id: Optional[str]) -> tuple[int, str]:
    """Resolve account for Individual account type."""
    contact_name = job_data.get("contact_name")
    if not contact_name:
        raise HTTPException(
            status_code=400, detail="contact_name is required for Individual account type"
        )
    parts = contact_name.strip().split(" ", 1)
    first_name = parts[0]
    last_name = parts[1] if len(parts) > 1 else ""
    individual = await get_or_create_individual(first_name, last_name, tenant_id)
    if not individual.account_id:
        raise HTTPException(
            status_code=400, detail=f"Individual {contact_name} has no account"
        )
    return individual.account_id, contact_name


async def _resolve_organization_account(job_data: Dict[str, Any], tenant_id: Optional[str]) -> tuple[int, str]:
    """Resolve account for Organization account type."""
    organization_name = job_data.get("account") or job_data.get("organization_name")
    if not organization_name:
        raise HTTPException(status_code=400, detail="account is required")
    industry_ids = job_data.get("industry_ids")
    industry_input = job_data.get("industry")
    if not industry_ids and industry_input:
        industry_names = industry_input if isinstance(industry_input, list) else [industry_input]
        industry_ids = [
            ind.id for name in industry_names if name
            for ind in [await find_industry_by_name(name)] if ind
        ]
    org = await get_or_create_organization(organization_name, tenant_id, industry_ids if industry_ids else None)
    if not org.account_id:
        raise HTTPException(
            status_code=400, detail=f"Organization {organization_name} has no account"
        )
    return org.account_id, org.name


def _validate_contact_phone(phone: Optional[str], first_name: str, last_name: str) -> Optional[str]:
    """Validate and return phone, or raise HTTPException."""
    if not phone:
        return None
    try:
        return validate_phone(phone)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=f"Contact {first_name} {last_name}: {str(e)}")


async def _process_single_contact(
    contact_data: Dict[str, Any],
    organization_id: Optional[int],
    tenant_id: Optional[str],
    is_primary_default: bool
) -> None:
    """Process a single contact for a job's account. Returns early if invalid."""
    first_name = contact_data.get("first_name", "").strip()
    last_name = contact_data.get("last_name", "").strip()
    if not first_name or not last_name:
        return

    contact_phone = _validate_contact_phone(contact_data.get("phone"), first_name, last_name)

    # Use is_primary from contact_data if provided, otherwise use default
    is_primary = contact_data.get("is_primary", is_primary_default)

    # Use provided individual_id if available (e.g., from search selection)
    individual_id = contact_data.get("individual_id")

    # Use existing contact_id if provided (for updates)
    contact_id = contact_data.get("id")

    await get_or_create_contact(
        contact_id=contact_id,
        individual_id=individual_id,
        organization_id=organization_id,
        first_name=first_name,
        last_name=last_name,
        email=contact_data.get("email"),
        phone=contact_phone,
        title=contact_data.get("title"),
        is_primary=is_primary
    )


def _parse_resume_date(resume_str: str | None) -> tuple[bool, datetime | None]:
    """Parse resume date string. Returns (should_update, parsed_value)."""
    if resume_str is None:
        return False, None
    if resume_str == "":
        return True, None
    return True, parse_date_input(resume_str)


def _normalize_identifier(company: str, title: str) -> str:
    """Normalize job identifier for comparison."""
    return f"{company.lower().strip()}|{title.lower().strip()}"


def _extract_account_info(organizations: list, individuals: list) -> tuple[str, str, str | None]:
    """Extract company, account_type, and contact_name from account data."""
    if organizations:
        return organizations[0].get("name", "Unknown Company"), "Organization", None
    if individuals:
        ind = individuals[0]
        first_name = ind.get("first_name", "")
        last_name = ind.get("last_name", "")
        contact_name = f"{first_name} {last_name}".strip()
        return contact_name, "Individual", contact_name
    return "Unknown Company", "Organization", None


def _map_to_dict(job) -> Dict[str, str]:
    """Map database model to dictionary for API compatibility."""
    job_dict = model_to_dict(job, include_relationships=True)

    lead = job_dict.get("lead", {})
    account = lead.get("account", {})
    organizations = account.get("organizations", [])
    individuals = account.get("individuals", [])
    company, account_type, contact_name = _extract_account_info(organizations, individuals)

    # Extract status name from current_status (backward compatibility)
    status = job_dict.get("current_status", {}).get("name", "Applied")

    created_at = job_dict.get("created_at")
    date_applied = str(created_at)[:10] if created_at else "Not specified"

    result = {
        "company": company,
        "title": job_dict.get("job_title", ""),
        "date_applied": date_applied,
        "link": job_dict.get("job_url") or "",
        "status": status,
        "salary_range": job_dict.get("salary_range") or "",
        "notes": job_dict.get("notes") or "",
        "source": lead.get("source", "manual"),
        "id": str(job_dict.get("id", "")),
        "account_type": account_type,
    }

    if contact_name:
        result["contact_name"] = contact_name

    return result


async def get_all_jobs(refresh: bool = False) -> List[Dict[str, str]]:
    """Get all jobs as dictionaries."""
    jobs, _ = await find_all_jobs()
    # Map to internal format with company/date_applied keys for consistency
    details = [_map_to_dict(job) for job in jobs]
    return details


async def get_applied_jobs_only(refresh: bool = False) -> List[Dict[str, str]]:
    """Get only jobs with status 'Applied'."""
    all_jobs = await get_all_jobs(refresh=refresh)
    applied_jobs = [
        job for job in all_jobs if job.get("status", "") == "Applied"
    ]
    return applied_jobs


async def get_applied_identifiers(refresh: bool = False) -> set[str]:
    """Get set of normalized job identifiers (only status='applied')."""
    applied_jobs = await get_applied_jobs_only(refresh=refresh)
    return {
        _normalize_identifier(j["company"], j["title"])
        for j in applied_jobs
        if j.get("title")
    }


async def fetch_applied_jobs(refresh: bool = False) -> set[str]:
    """Fetch set of normalized job identifiers (for compatibility)."""
    return await get_applied_identifiers(refresh=refresh)


async def get_jobs_details(refresh: bool = False) -> List[Dict[str, str]]:
    """Get detailed list of applied jobs (status='applied' only, for compatibility)."""
    return await get_applied_jobs_only(refresh=refresh)


async def is_job_applied(company: str, title: str) -> bool:
    """Check if a job has been applied to (status='applied' only)."""
    applied_jobs = await get_applied_jobs_only()
    identifier = _normalize_identifier(company, title)
    applied_identifiers = {
        _normalize_identifier(j["company"], j["title"])
        for j in applied_jobs
        if j.get("title")
    }
    return identifier in applied_identifiers


async def filter_jobs(jobs: List[Dict[str, str]]) -> List[Dict[str, str]]:
    """Filter out jobs that have already been applied to (status='applied' only)."""
    identifiers = await get_applied_identifiers()

    if not identifiers:
        logger.warning("No applied jobs (status='applied') found - returning all jobs")
        return jobs

    filtered = []
    for job in jobs:
        is_applied = await is_job_applied(job.get("company", ""), job.get("title", ""))
        if not is_applied:
            filtered.append(job)

    return filtered


async def add_job(
    organization_name: str,
    job_title: str,
    project_ids: List[int],
    job_url: str = "",
    salary_range: str = "",
    notes: str = "",
    status: str = "applied",
    source: str = "agent",
) -> Optional[Dict[str, str]]:
    """Add a new job application.

    Args:
        organization_name: Company name
        job_title: Job title
        project_ids: List of project IDs (required, must have at least one)
        job_url: Optional job posting URL
        salary_range: Optional salary range
        notes: Optional notes
        status: Job status (default: applied)
        source: Source of job (default: agent)
    """
    if not project_ids:
        raise HTTPException(
            status_code=400,
            detail="At least one project must be specified for a job"
        )

    # Get or create organization
    org = await get_or_create_organization(organization_name)

    data: Dict[str, Any] = {
        "organization_id": org.id,
        "job_title": job_title,
        "job_url": job_url or None,
        "salary_range": salary_range or None,
        "notes": notes or None,
        "status": status,  # Will be converted to status_id in repository
        "source": source,
        "project_ids": project_ids,
    }

    created = await create_job(data)
    return _map_to_dict(created)


async def add_applied_job(
    company: str,
    job_title: str,
    project_ids: List[int],
    job_url: str = "",
    location: str = "",  # Deprecated, kept for backward compatibility but ignored
    salary_range: str = "",
    notes: str = "",
    status: str = "applied",
    source: str = "agent",
) -> Optional[Dict[str, str]]:
    """Add a new job application (backward compatibility wrapper)."""
    return await add_job(
        organization_name=company,
        job_title=job_title,
        project_ids=project_ids,
        job_url=job_url,
        salary_range=salary_range,
        notes=notes,
        status=status,
        source=source,
    )


def _fuzzy_match_company(search_company: str, db_company: str) -> bool:
    """Check if company names match using fuzzy matching."""
    if not search_company or not db_company:
        return False

    search_norm = search_company.lower().strip()
    db_norm = db_company.lower().strip()

    # Exact match
    if search_norm == db_norm:
        return True

    # Remove common suffixes for comparison
    search_clean = (
        search_norm.replace(" inc", "")
        .replace(" llc", "")
        .replace(" corp", "")
        .replace(" ltd", "")
        .strip()
    )
    db_clean = (
        db_norm.replace(" inc", "")
        .replace(" llc", "")
        .replace(" corp", "")
        .replace(" ltd", "")
        .strip()
    )

    if search_clean == db_clean:
        return True

    # One is substring of the other (require at least 4 chars to avoid false matches)
    if len(search_clean) < 4 or len(db_clean) < 4:
        return False

    return search_clean in db_clean or db_clean in search_clean


def _matches_job(job, company: str, job_title: str) -> bool:
    """Check if job matches company and title using fuzzy matching."""
    job_dict = model_to_dict(job, include_relationships=True)

    # Get company from Lead.account.organizations
    job_company = ""
    lead = job_dict.get("lead", {})
    account = lead.get("account", {})
    organizations = account.get("organizations", [])
    if organizations and len(organizations) > 0:
        job_company = organizations[0].get("name", "")

    if not _fuzzy_match_company(company, job_company):
        return False

    job_title_lower = job_title.lower().strip()
    db_title_lower = job_dict.get("job_title", "").lower()
    return job_title_lower in db_title_lower or db_title_lower in job_title_lower


async def get_jobs_paginated(
    page: int, limit: int, order_by: str, order: str, search: Optional[str] = None, tenant_id: Optional[int] = None, project_id: Optional[int] = None
) -> tuple[List[Any], int]:
    """Get all jobs with search, sorting, and pagination logic.

    Args:
        page: Page number
        limit: Items per page
        order_by: Field to sort by
        order: Sort direction (ASC/DESC)
        search: Optional search query
        tenant_id: Optional tenant ID for filtering
        project_id: Optional project ID for filtering

    Returns:
        Tuple of (paginated_jobs, total_count)
    """
    return await find_all_jobs_repo(
        page=page,
        page_size=limit,
        search=search,
        order_by=order_by,
        order=order,
        tenant_id=tenant_id,
        project_id=project_id
    )


async def create_job(job_data: Dict[str, Any]) -> Any:
    """Create a new job.

    Handles:
    - Organization or Individual lookup/creation based on account_type
    - Date parsing
    - Field validation
    - Repository creation

    Args:
        job_data: Dictionary of job fields

    Returns:
        Created Job ORM object

    Raises:
        HTTPException: If required fields are missing
    """
    account_type = job_data.get("account_type", "Organization")
    tenant_id = job_data.get("tenant_id")
    resolve_account = _resolve_individual_account if account_type == "Individual" else _resolve_organization_account
    account_id, account_name = await resolve_account(job_data, tenant_id)

    # Parse resume date
    resume_obj = parse_date_input(job_data.get("resume"))

    # Resolve deal_id
    deal_id = job_data.get("deal_id")
    if not deal_id:
        deal = await get_or_create_deal("Job Search 2025")
        deal_id = deal.id

    # Resolve status_id from status name
    status_id = job_data.get("current_status_id")
    if not status_id:
        status_name = job_data.get("status", "applied")
        status = await find_status_by_name(status_name, "job")
        status_id = status.id if status else None

    user_id = job_data.get("user_id")
    data: Dict[str, Any] = {
        "account_id": account_id,
        "account_name": account_name,
        "deal_id": deal_id,
        "current_status_id": status_id,
        "job_title": job_data.get("job_title", ""),
        "job_url": job_data.get("job_url"),
        "salary_range": job_data.get("salary_range"),  # Legacy field
        "salary_range_id": job_data.get("salary_range_id"),
        "notes": job_data.get("notes"),
        "source": job_data.get("source", "manual"),
        "tenant_id": tenant_id,
        "resume_date": resume_obj,
        "project_ids": job_data.get("project_ids") or [],
        "created_by": user_id,
        "updated_by": user_id,
    }

    # Validate that at least one project is specified
    if not data["project_ids"]:
        raise HTTPException(
            status_code=400,
            detail="At least one project must be specified for a job"
        )

    created = await create_job_repo(data)

    # Process contacts if provided
    contacts = job_data.get("contacts") or []
    organization_id = None
    if account_type == "Organization":
        # Get org_id for linking contacts
        organization_name = job_data.get("account") or job_data.get("organization_name")
        if organization_name:
            org = await get_or_create_organization(organization_name, tenant_id)
            organization_id = org.id

    for idx, contact_data in enumerate(contacts):
        await _process_single_contact(contact_data, organization_id, tenant_id, is_primary_default=(idx == 0))

    return created


async def get_job(job_id: str) -> Any:
    """Get a job by ID.

    Args:
        job_id: Job ID (string from route)

    Returns:
        Job ORM object

    Raises:
        HTTPException: If job_id is invalid or job not found
    """
    try:
        job_id_int = int(job_id)
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid job ID format")

    job = await find_job_by_id(job_id_int)

    if not job:
        raise HTTPException(status_code=404, detail="Job not found")

    return job


async def delete_job(job_id: str) -> bool:
    """Delete a job by ID.

    Args:
        job_id: Job ID (string from route)

    Returns:
        True if deleted

    Raises:
        HTTPException: If job_id is invalid or job not found
    """
    try:
        job_id_int = int(job_id)
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid job ID format")

    deleted = await delete_job_repo(job_id_int)

    if not deleted:
        raise HTTPException(status_code=404, detail="Job not found")

    return deleted


async def get_statuses() -> List[Any]:
    """Get all job statuses.

    Returns:
        List of Status ORM objects
    """
    return await find_all_statuses()


async def get_jobs_with_status(status_names: Optional[List[str]] = None) -> List[Any]:
    """Get jobs filtered by status names.

    Args:
        status_names: Optional list of status names to filter by

    Returns:
        List of Job ORM objects
    """
    return await find_jobs_with_status(status_names)


async def get_recent_resume_date() -> Optional[str]:
    """Get the most recent resume date from all jobs.

    Returns:
        Resume date string (YYYY-MM-DD) or None
    """
    all_jobs, _ = await find_all_jobs_repo()

    # Filter jobs with resume dates and sort by Lead created_at DESC
    jobs_with_resume = [job for job in all_jobs if job.resume_date]
    if not jobs_with_resume:
        return None

    # Sort by Lead created_at (application date) DESC to get most recent application
    jobs_with_resume.sort(
        key=lambda j: j.lead.created_at if j.lead else datetime.min, reverse=True
    )

    # Get the resume date from the most recently applied job
    most_recent_job = jobs_with_resume[0]
    if most_recent_job.resume_date:
        return most_recent_job.resume_date.isoformat().split("T")[0]

    return None


async def update_job(job_id: str, job_data: Dict[str, Any]) -> Any:
    """Update a job with the provided data.

    Handles:
    - ID validation
    - Organization lookup/creation
    - Date parsing
    - Field validation
    - Repository update
    - Error handling

    Args:
        job_id: Job ID to update (string from route)
        job_data: Dictionary of fields to update

    Returns:
        Updated Job ORM object

    Raises:
        HTTPException: If job_id is invalid or job not found
    """
    # Validate job_id
    try:
        job_id_int = int(job_id)
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid job ID format")

    data: Dict[str, Any] = {}
    user_id = job_data.get("user_id")
    if user_id:
        data["updated_by"] = user_id

    # Handle organization update
    organization_name = job_data.get("account") or job_data.get("organization_name")
    if organization_name:
        org = await get_or_create_organization(organization_name)
        data["organization_id"] = org.id

    # Update simple fields
    allowed_fields = [
        "job_title",
        "job_url",
        "salary_range",  # Legacy field
        "salary_range_id",
        "notes",
        "source",
        "project_ids",
    ]
    data.update({k: job_data[k] for k in allowed_fields if k in job_data})

    # Resolve status name to status_id
    if "current_status_id" in job_data:
        data["current_status_id"] = job_data["current_status_id"]

    status_name = job_data.get("status")
    if "current_status_id" not in job_data and status_name:
        status = await find_status_by_name(status_name, "job")
        data["current_status_id"] = status.id if status else None

    # Handle resume_date parsing
    should_update, resume_date = _parse_resume_date(job_data.get("resume"))
    if should_update:
        data["resume_date"] = resume_date

    # Update via repository
    updated = await update_job_repo(job_id_int, data)

    # Handle not found
    if not updated:
        raise HTTPException(status_code=404, detail="Job not found")

    # Handle contacts update if provided
    contacts = job_data.get("contacts")
    if contacts is None:
        return updated

    # Get organization_id and tenant_id from the job's lead account
    organization_id = None
    tenant_id = None
    if not updated.lead or not updated.lead.account:
        for idx, contact_data in enumerate(contacts):
            await _process_single_contact(contact_data, organization_id, tenant_id, is_primary_default=(idx == 0))
        return updated

    organizations = updated.lead.account.organizations
    organization_id = organizations[0].id if organizations else None
    tenant_id = updated.lead.account.tenant_id
    for idx, contact_data in enumerate(contacts):
        await _process_single_contact(contact_data, organization_id, tenant_id, is_primary_default=(idx == 0))
    return updated


async def update_job_by_company(
    company: str,
    job_title: str,
    status: Optional[str] = None,
    job_url: Optional[str] = None,
    notes: Optional[str] = None,
) -> Optional[Dict[str, str]]:
    """
    Find and update a job by company name and job title (fuzzy matching).
    Returns updated job or None if not found.
    """
    # Get all jobs and find matching one
    all_jobs, _ = await find_all_jobs()
    matching_job = next(
        (job for job in all_jobs if _matches_job(job, company, job_title)), None
    )

    if not matching_job:
        return None

    # Build update data
    update_data: Dict[str, Any] = {
        k: v
        for k, v in [
            ("status", status),
            ("job_url", job_url),
            ("notes", notes),
        ]
        if v is not None
    }

    # Update the job
    updated_job = await update_job(str(matching_job.id), update_data)

    if not updated_job:
        return None

    return _map_to_dict(updated_job)
