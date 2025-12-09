"""Repository layer for job data access."""

import logging
from typing import Any, Dict, List, Optional
from sqlalchemy import delete, func as sql_func, or_, select
from sqlalchemy.orm import joinedload
from lib.db import async_get_session
from models.Account import Account
from models.Deal import Deal
from models.Individual import Individual
from models.Job import Job
from models.Lead import Lead
from models.LeadProject import LeadProject
from models.Organization import Organization
from models.Project import Project
from models.Status import Status
from models.Contact import Contact
from schemas.job import JobCreate, JobUpdate

__all__ = ["JobCreate", "JobUpdate"]

logger = logging.getLogger(__name__)


def _update_lead_status(lead: Lead, data: Dict[str, Any]) -> None:
    """Update lead status from data. Expects current_status_id to be resolved by service."""
    if "current_status_id" not in data:
        return
    if data["current_status_id"] is None:
        return
    lead.current_status_id = data["current_status_id"]


def _update_lead_source(lead: Lead, data: Dict[str, Any]) -> None:
    """Update lead source if provided."""
    if "source" not in data:
        return
    if data["source"] is None:
        return
    lead.source = data["source"]


async def _load_job_with_relationships(session, job_id: int) -> Job:
    """Load job with all relationships eagerly."""
    stmt = select(Job).options(
        joinedload(Job.lead).joinedload(Lead.current_status),
        joinedload(Job.lead).joinedload(Lead.tenant),
        joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.tenant),
        joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.current_status),
        joinedload(Job.lead).joinedload(Lead.account).joinedload(Account.organizations),
        joinedload(Job.lead).joinedload(Lead.account).joinedload(Account.individuals),
        joinedload(Job.lead).joinedload(Lead.account).joinedload(Account.contacts),
        joinedload(Job.lead).joinedload(Lead.account).joinedload(Account.industries),
        joinedload(Job.lead).joinedload(Lead.projects),
        joinedload(Job.salary_range_ref)
    ).where(Job.id == job_id)
    result = await session.execute(stmt)
    return result.unique().scalar_one()


async def find_all_jobs(
    page: int = 1,
    page_size: int = 50,
    search: Optional[str] = None,
    order_by: str = "updated_at",
    order: str = "DESC",
    tenant_id: Optional[int] = None,
    project_id: Optional[int] = None
) -> tuple[List[Job], int]:
    """Find jobs with pagination, filtering, and sorting at database level.
    
    Optimized for list view - only loads essential relationships:
    - Lead status (for status display)
    - Account organizations/individuals (for company name)
    - Salary range (for salary display)
    """
    async with async_get_session() as session:
        # Base query with minimal relationships needed for list view
        stmt = select(Job).options(
            joinedload(Job.lead).joinedload(Lead.current_status),
            joinedload(Job.lead).joinedload(Lead.account).joinedload(Account.organizations),
            joinedload(Job.lead).joinedload(Lead.account).joinedload(Account.individuals),
            joinedload(Job.salary_range_ref)
        ).join(Lead).outerjoin(Account, Lead.account_id == Account.id).outerjoin(Organization, Account.id == Organization.account_id)

        # Filter by tenant
        if tenant_id is not None:
            stmt = stmt.where(Lead.tenant_id == tenant_id)

        # Filter by project (via many-to-many junction table)
        # If project_id is provided, filter by that project; otherwise return all jobs
        if project_id is not None:
            stmt = stmt.join(LeadProject, Lead.id == LeadProject.lead_id).where(LeadProject.project_id == project_id)

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
            "updated_at": Lead.updated_at,
            "application_date": Lead.created_at,
            "company": Organization.name,
            "job_title": Job.job_title,
            "resume": Job.resume_date,
        }
        sort_column = sort_map.get(order_by, Lead.updated_at)
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
            joinedload(Job.lead).joinedload(Lead.tenant),
            joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.tenant),
            joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.current_status),
            joinedload(Job.lead).joinedload(Lead.account).joinedload(Account.organizations),
            joinedload(Job.lead).joinedload(Lead.account).joinedload(Account.individuals),
            joinedload(Job.lead).joinedload(Lead.account).joinedload(Account.contacts),
            joinedload(Job.lead).joinedload(Lead.account).joinedload(Account.industries),
            joinedload(Job.lead).joinedload(Lead.projects),
            joinedload(Job.salary_range_ref)
        ).where(Job.id == job_id)
        result = await session.execute(stmt)
        return result.unique().scalar_one_or_none()


async def create_job(data: Dict[str, Any]) -> Job:
    """Create a new job (with Lead parent).

    Note: This handles the joined table inheritance by creating both Lead and Job records.
    Expects account_id, deal_id, and current_status_id to be resolved by service layer.
    """
    async with async_get_session() as session:
        try:
            account_id = data.get("account_id")
            tenant_id = data.get("tenant_id")
            deal_id = data.get("deal_id")
            status_id = data.get("current_status_id")

            if not account_id:
                raise ValueError("account_id is required")
            if not deal_id:
                raise ValueError("deal_id is required")

            # Create Lead first with account_id
            lead_data: Dict[str, Any] = {
                "deal_id": deal_id,
                "type": "Job",
                "source": data.get("source", "manual"),
                "current_status_id": status_id,
                "account_id": account_id,
                "tenant_id": tenant_id,
                "created_by": data.get("created_by"),
                "updated_by": data.get("updated_by"),
            }

            lead = Lead(**lead_data)
            session.add(lead)
            await session.flush()  # Get the lead.id

            # Handle project_ids (many-to-many) - insert directly into junction table
            project_ids = data.get("project_ids") or []
            for pid in project_ids:
                lead_project = LeadProject(lead_id=lead.id, project_id=pid)
                session.add(lead_project)

            # Create Job with same ID as Lead
            job = Job(
                id=lead.id,
                job_title=data.get("job_title", ""),
                job_url=data.get("job_url"),
                description=data.get("description"),
                resume_date=data.get("resume_date"),
                salary_range=data.get("salary_range"),
                salary_range_id=data.get("salary_range_id"),
            )

            session.add(job)
            await session.commit()
            return await _load_job_with_relationships(session, job.id)
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to create job: {e}")
            raise


async def _sync_account_from_org(session, job: Job, data: Dict[str, Any]) -> None:
    """Sync account_id from organization if organization_id is provided."""
    org_id = data.get("organization_id")
    if not org_id:
        return
    org = await session.get(Organization, org_id)
    if not org or not org.account_id:
        return
    job.lead.account_id = org.account_id
    data["account_id"] = org.account_id


async def update_job(job_id: int, data: Dict[str, Any]) -> Optional[Job]:
    """Update an existing job.

    Note: Updates both Job and Lead records as needed.
    """
    async with async_get_session() as session:
        try:
            # Eagerly load the job with its lead and account relationships
            stmt = select(Job).options(
                joinedload(Job.lead).joinedload(Lead.account).joinedload(Account.organizations),
                joinedload(Job.lead).joinedload(Lead.account).joinedload(Account.individuals)
            ).where(Job.id == job_id)
            result = await session.execute(stmt)
            job = result.unique().scalar_one_or_none()

            if not job:
                return None

            # Update Job fields
            if "job_title" in data and data["job_title"] is not None:
                job.job_title = data["job_title"]
            if "job_url" in data:
                job.job_url = data["job_url"]
            if "description" in data:
                job.description = data["description"]
            if "resume_date" in data:
                job.resume_date = data["resume_date"]
            if "salary_range" in data:
                job.salary_range = data["salary_range"]
            if "salary_range_id" in data:
                job.salary_range_id = data["salary_range_id"]

            # Update Lead fields if provided
            if not job.lead:
                await session.commit()
                return await _load_job_with_relationships(session, job.id)

            # Handle organization/account update
            await _sync_account_from_org(session, job, data)

            _update_lead_status(job.lead, data)
            _update_lead_source(job.lead, data)

            # Update project_ids if provided (many-to-many)
            if "project_ids" in data:
                # Delete existing links
                await session.execute(
                    delete(LeadProject).where(LeadProject.lead_id == job.lead.id)
                )
                # Insert new links
                for pid in (data["project_ids"] or []):
                    lead_project = LeadProject(lead_id=job.lead.id, project_id=pid)
                    session.add(lead_project)

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
            joinedload(Job.lead).joinedload(Lead.tenant),
            joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.tenant),
            joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.current_status),
            joinedload(Job.lead).joinedload(Lead.account).joinedload(Account.organizations),
            joinedload(Job.lead).joinedload(Lead.account).joinedload(Account.individuals),
            joinedload(Job.lead).joinedload(Lead.account).joinedload(Account.contacts),
            joinedload(Job.lead).joinedload(Lead.account).joinedload(Account.industries),
            joinedload(Job.lead).joinedload(Lead.projects),
            joinedload(Job.salary_range_ref)
        ).join(Lead).outerjoin(Account, Lead.account_id == Account.id).outerjoin(Organization, Account.id == Organization.account_id).where(
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
            joinedload(Job.lead).joinedload(Lead.tenant),
            joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.tenant),
            joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.current_status),
            joinedload(Job.lead).joinedload(Lead.account).joinedload(Account.organizations),
            joinedload(Job.lead).joinedload(Lead.account).joinedload(Account.individuals),
            joinedload(Job.lead).joinedload(Lead.account).joinedload(Account.contacts),
            joinedload(Job.lead).joinedload(Lead.account).joinedload(Account.industries),
            joinedload(Job.lead).joinedload(Lead.projects),
            joinedload(Job.salary_range_ref)
        ).join(Lead).join(Status).where(Lead.current_status_id.isnot(None))

        if status_names:
            stmt = stmt.where(Status.name.in_(status_names))

        stmt = stmt.order_by(Lead.created_at.desc())
        result = await session.execute(stmt)
        return list(result.unique().scalars().all())
