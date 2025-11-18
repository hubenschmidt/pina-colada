"""Repository layer for job data access."""
from typing import Dict, Any, Optional

import logging
from typing import List, Optional
from sqlalchemy import select, func as sql_func, or_
from sqlalchemy.orm import joinedload
from models.Job import Job
from models.Lead import Lead
from models.Status import Status
from models.Organization import Organization
from models.Deal import Deal
from lib.db import async_get_session
from repositories.deal_repository import get_or_create_deal
from repositories.status_repository import find_status_by_name

logger = logging.getLogger(__name__)


async def _get_status_id_from_name(status_name: str) -> Optional[int]:
    """Get status ID from status name."""
    status = await find_status_by_name(status_name, "job")
    if not status:
        return None
    return status.id


def _update_job_notes(job: Job, notes: str) -> None:
    """Update job notes and lead description."""
    job.notes = notes
    if not job.lead:
        return
    job.lead.description = notes


async def _update_lead_status(lead: Lead, data: Dict[str, Any]) -> None:
    """Update lead status from data."""
    if "current_status_id" in data:
        lead.current_status_id = data["current_status_id"]
        return

    if "status" not in data:
        return

    if not data["status"]:
        return

    status_id = await _get_status_id_from_name(data["status"])

    if not status_id:
        logger.warning(f"Could not find status ID for name: {data['status']}")
        return

    lead.current_status_id = status_id


def _update_lead_source(lead: Lead, data: Dict[str, Any]) -> None:
    """Update lead source if provided."""
    if "source" not in data:
        return
    if data["source"] is None:
        return
    lead.source = data["source"]


async def _update_lead_title_if_needed(session, job: Job, data: Dict[str, Any]) -> None:
    """Update lead title if job_title or organization changed."""
    if "job_title" not in data and "organization_id" not in data:
        return

    org = await session.get(Organization, job.organization_id)
    job.lead.title = f"{org.name if org else 'Unknown'} - {job.job_title}"


async def _load_job_with_relationships(session, job_id: int) -> Job:
    """Load job with all relationships eagerly."""
    stmt = select(Job).options(
        joinedload(Job.lead).joinedload(Lead.current_status),
        joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.tenant),
        joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.current_status),
        joinedload(Job.organization).joinedload(Organization.tenant)
    ).where(Job.id == job_id)
    result = await session.execute(stmt)
    return result.unique().scalar_one()


async def find_all_jobs(
    page: int = 1,
    page_size: int = 50,
    search: Optional[str] = None,
    order_by: str = "date",
    order: str = "DESC"
) -> tuple[List[Job], int]:
    """Find jobs with pagination, filtering, and sorting at database level."""
    async with async_get_session() as session:
        # Base query with relationships
        stmt = select(Job).options(
            joinedload(Job.lead).joinedload(Lead.current_status),
            joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.tenant),
            joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.current_status),
            joinedload(Job.organization).joinedload(Organization.tenant)
        ).join(Lead).join(Organization)

        # Apply search filter at DB level
        if search and search.strip():
            search_lower = search.strip().lower()
            stmt = stmt.where(
                or_(
                    sql_func.lower(Organization.name).contains(search_lower),
                    sql_func.lower(Job.job_title).contains(search_lower)
                )
            )

        # Get total count before pagination
        count_stmt = select(sql_func.count()).select_from(stmt.alias())
        count_result = await session.execute(count_stmt)
        total_count = count_result.scalar() or 0

        # Apply sorting at DB level
        sort_map = {
            "date": Lead.created_at,
            "application_date": Lead.created_at,
            "company": Organization.name,
            "job_title": Job.job_title,
            "resume": Job.resume_date,
        }
        sort_column = sort_map.get(order_by, Lead.created_at)
        stmt = stmt.order_by(sort_column.desc() if order.upper() == "DESC" else sort_column.asc())

        # Apply pagination at DB level
        offset = (page - 1) * page_size
        stmt = stmt.limit(page_size).offset(offset)

        # Execute query
        result = await session.execute(stmt)
        jobs = list(result.unique().scalars().all())

        return jobs, total_count


async def count_jobs() -> int:
    """Count total jobs."""
    async with async_get_session() as session:
        stmt = select(sql_func.count(Job.id))
        result = await session.execute(stmt)
        return result.scalar() or 0


async def find_job_by_id(job_id: int) -> Optional[Job]:
    """Find job by ID with related data."""
    async with async_get_session() as session:
        stmt = select(Job).options(
            joinedload(Job.lead).joinedload(Lead.current_status),
            joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.tenant),
            joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.current_status),
            joinedload(Job.organization).joinedload(Organization.tenant)
        ).where(Job.id == job_id)
        result = await session.execute(stmt)
        return result.unique().scalar_one_or_none()


async def create_job(data: Dict[str, Any]) -> Job:
    """Create a new job (with Lead parent).

    Note: This handles the joined table inheritance by creating both Lead and Job records.
    """
    async with async_get_session() as session:
        try:
            # Get organization_id
            organization_id = data.get("organization_id")
            if not organization_id:
                raise ValueError("organization_id is required")

            # Get or create default deal if not provided
            deal_id = data.get("deal_id")
            if not deal_id:
                # Use default "Job Search" deal
                deal = await get_or_create_deal("Job Search 2025")
                deal_id = deal.id

            # Get status_id from status name if provided
            status_id = data.get("current_status_id")
            if not status_id and "status" in data:
                status_id = await _get_status_id_from_name(data["status"])

            # Build title from organization and job_title
            org = await session.get(Organization, organization_id)
            title = f"{org.name if org else 'Unknown'} - {data.get('job_title', 'Job')}"

            # Create Lead first
            lead_data: Dict[str, Any] = {
                "deal_id": deal_id,
                "type": "Job",
                "title": title,
                "description": data.get("notes"),
                "source": data.get("source", "manual"),
                "current_status_id": status_id
            }

            lead = Lead(**lead_data)
            session.add(lead)
            await session.flush()  # Get the lead.id

            # Create Job with same ID as Lead
            job = Job(
                id=lead.id,
                organization_id=organization_id,
                job_title=data.get("job_title", ""),
                job_url=data.get("job_url"),
                notes=data.get("notes"),
                resume_date=data.get("resume_date"),
                salary_range=data.get("salary_range")
            )

            session.add(job)
            await session.commit()
            return await _load_job_with_relationships(session, job.id)
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to create job: {e}")
            raise


async def update_job(job_id: int, data: Dict[str, Any]) -> Optional[Job]:
    """Update an existing job.

    Note: Updates both Job and Lead records as needed.
    """
    async with async_get_session() as session:
        try:
            # Eagerly load the job with its lead relationship
            stmt = select(Job).options(joinedload(Job.lead)).where(Job.id == job_id)
            result = await session.execute(stmt)
            job = result.unique().scalar_one_or_none()

            if not job:
                return None

            # Update Job fields
            if "job_title" in data and data["job_title"] is not None:
                job.job_title = data["job_title"]
            if "job_url" in data:
                job.job_url = data["job_url"]
            if "notes" in data:
                _update_job_notes(job, data["notes"])
            if "resume_date" in data:
                job.resume_date = data["resume_date"]
            if "salary_range" in data:
                job.salary_range = data["salary_range"]
            if "organization_id" in data and data["organization_id"] is not None:
                job.organization_id = data["organization_id"]

            # Update Lead fields if provided
            if not job.lead:
                await session.commit()
                return await _load_job_with_relationships(session, job.id)

            await _update_lead_status(job.lead, data)
            _update_lead_source(job.lead, data)
            await _update_lead_title_if_needed(session, job, data)

            await session.commit()
            return await _load_job_with_relationships(session, job.id)
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to update job: {e}")
            raise


async def delete_job(job_id: int) -> bool:
    """Delete a job by ID."""
    async with async_get_session() as session:
        try:
            stmt = select(Job).where(Job.id == job_id)
            result = await session.execute(stmt)
            job = result.scalar_one_or_none()
            if not job:
                return False

            await session.delete(job)
            await session.commit()
            return True
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to delete job: {e}")
            raise


async def find_job_by_company_and_title(company: str, title: str) -> Optional[Job]:
    """Find job by company name and title."""
    async with async_get_session() as session:
        stmt = select(Job).options(
            joinedload(Job.lead).joinedload(Lead.current_status),
            joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.tenant),
            joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.current_status),
            joinedload(Job.organization).joinedload(Organization.tenant)
        ).join(Organization).join(Lead).where(
            sql_func.lower(Organization.name).contains(sql_func.lower(company.strip())),
            sql_func.lower(Job.job_title).contains(sql_func.lower(title.strip()))
        )
        result = await session.execute(stmt)
        return result.unique().scalar_one_or_none()


async def find_all_statuses() -> List[Status]:
    """Find all job statuses."""
    async with async_get_session() as session:
        stmt = select(Status).where(Status.category == "job").order_by(Status.name)
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def find_jobs_with_status(status_names: Optional[List[str]] = None) -> List[Job]:
    """Find jobs filtered by status names."""
    async with async_get_session() as session:
        stmt = select(Job).options(
            joinedload(Job.lead).joinedload(Lead.current_status),
            joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.tenant),
            joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.current_status),
            joinedload(Job.organization).joinedload(Organization.tenant)
        ).join(Lead).join(Status).where(Lead.current_status_id.isnot(None))

        if status_names:
            stmt = stmt.where(Status.name.in_(status_names))

        stmt = stmt.order_by(Lead.created_at.desc())
        result = await session.execute(stmt)
        return list(result.unique().scalars().all())
