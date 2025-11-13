"""Repository layer for job data access."""

import logging
from typing import List, Optional
from sqlalchemy import select, func as sql_func
from sqlalchemy.orm import joinedload
from models.Job import Job, JobCreateData, JobUpdateData
from models.Lead import Lead, LeadCreateData
from models.Status import Status
from models.Organization import Organization
from models.Deal import Deal
from lib.db import get_session
from repositories.deal_repository import get_or_create_deal
from repositories.status_repository import find_status_by_name

logger = logging.getLogger(__name__)


def find_all_jobs() -> List[Job]:
    """Find all jobs with related data."""
    session = get_session()
    try:
        stmt = select(Job).options(
            joinedload(Job.lead).joinedload(Lead.current_status),
            joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.tenant),
            joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.current_status),
            joinedload(Job.organization).joinedload(Organization.tenant)
        ).join(Lead).order_by(Lead.created_at.desc())
        return list(session.execute(stmt).unique().scalars().all())
    finally:
        session.close()


def count_jobs() -> int:
    """Count total jobs."""
    session = get_session()
    try:
        stmt = select(sql_func.count(Job.id))
        return session.execute(stmt).scalar() or 0
    finally:
        session.close()


def find_job_by_id(job_id: int) -> Optional[Job]:
    """Find job by ID with related data."""
    session = get_session()
    try:
        stmt = select(Job).options(
            joinedload(Job.lead).joinedload(Lead.current_status),
            joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.tenant),
            joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.current_status),
            joinedload(Job.organization).joinedload(Organization.tenant)
        ).where(Job.id == job_id)
        return session.execute(stmt).unique().scalar_one_or_none()
    finally:
        session.close()


def create_job(data: JobCreateData) -> Job:
    """Create a new job (with Lead parent).

    Note: This handles the joined table inheritance by creating both Lead and Job records.
    """
    session = get_session()
    try:
        # Get organization_id
        organization_id = data.get("organization_id")
        if not organization_id:
            raise ValueError("organization_id is required")

        # Get or create default deal if not provided
        deal_id = data.get("deal_id")
        if not deal_id:
            # Use default "Job Search" deal
            deal = get_or_create_deal("Job Search 2025")
            deal_id = deal.id

        # Get status_id from status name if provided
        status_id = data.get("current_status_id")
        if not status_id and "status" in data:
            # Map old status strings to Status records
            status_name = data["status"].title()  # "applied" → "Applied"
            status = find_status_by_name(status_name, "job")
            if status:
                status_id = status.id

        # Build title from organization and job_title
        org = session.get(Organization, organization_id)
        title = f"{org.name if org else 'Unknown'} - {data.get('job_title', 'Job')}"

        # Create Lead first
        lead_data: LeadCreateData = {
            "deal_id": deal_id,
            "type": "Job",
            "title": title,
            "description": data.get("notes"),
            "source": data.get("source", "manual"),
            "current_status_id": status_id
        }

        lead = Lead(**lead_data)
        session.add(lead)
        session.flush()  # Get the lead.id

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
        session.commit()

        # Eagerly load all relationships before closing session
        stmt = select(Job).options(
            joinedload(Job.lead).joinedload(Lead.current_status),
            joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.tenant),
            joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.current_status),
            joinedload(Job.organization).joinedload(Organization.tenant)
        ).where(Job.id == job.id)

        loaded_job = session.execute(stmt).unique().scalar_one()
        return loaded_job
    except Exception as e:
        session.rollback()
        logger.error(f"Failed to create job: {e}")
        raise
    finally:
        session.close()


def update_job(job_id: int, data: JobUpdateData) -> Optional[Job]:
    """Update an existing job.

    Note: Updates both Job and Lead records as needed.
    """
    session = get_session()
    try:
        job = session.get(Job, job_id)
        if not job:
            return None

        # Update Job fields
        if "job_title" in data and data["job_title"] is not None:
            job.job_title = data["job_title"]
        if "job_url" in data:
            job.job_url = data["job_url"]
        if "notes" in data:
            job.notes = data["notes"]
            # Also update Lead.description
            if job.lead:
                job.lead.description = data["notes"]
        if "resume_date" in data:
            job.resume_date = data["resume_date"]
        if "salary_range" in data:
            job.salary_range = data["salary_range"]
        if "organization_id" in data and data["organization_id"] is not None:
            job.organization_id = data["organization_id"]

        # Update Lead fields if provided
        if job.lead:
            if "current_status_id" in data:
                job.lead.current_status_id = data["current_status_id"]
            elif "status" in data and data["status"]:
                # Map status string to Status ID
                status_name = data["status"].title()
                status = find_status_by_name(status_name, "job")
                if status:
                    job.lead.current_status_id = status.id

            if "source" in data and data["source"] is not None:
                job.lead.source = data["source"]

            # Update title if job_title or organization changed
            if "job_title" in data or "organization_id" in data:
                org = session.get(Organization, job.organization_id)
                job.lead.title = f"{org.name if org else 'Unknown'} - {job.job_title}"

        session.commit()

        # Eagerly load all relationships before closing session
        stmt = select(Job).options(
            joinedload(Job.lead).joinedload(Lead.current_status),
            joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.tenant),
            joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.current_status),
            joinedload(Job.organization).joinedload(Organization.tenant)
        ).where(Job.id == job.id)

        loaded_job = session.execute(stmt).unique().scalar_one()
        return loaded_job
    except Exception as e:
        session.rollback()
        logger.error(f"Failed to update job: {e}")
        raise
    finally:
        session.close()


def delete_job(job_id: int) -> bool:
    """Delete a job by ID (cascades to Lead)."""
    session = get_session()
    try:
        job = session.get(Job, job_id)
        if not job:
            return False

        # Delete the Lead (Job will be cascade-deleted)
        if job.lead:
            session.delete(job.lead)
        else:
            session.delete(job)

        session.commit()
        return True
    except Exception as e:
        session.rollback()
        logger.error(f"Failed to delete job: {e}")
        raise
    finally:
        session.close()


def find_job_by_company_and_title(company: str, title: str) -> Optional[Job]:
    """Find job by company name and title."""
    session = get_session()
    try:
        stmt = select(Job).options(
            joinedload(Job.lead).joinedload(Lead.current_status),
            joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.tenant),
            joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.current_status),
            joinedload(Job.organization).joinedload(Organization.tenant)
        ).join(Organization).join(Lead).where(
            sql_func.lower(Organization.name).contains(sql_func.lower(company.strip())),
            sql_func.lower(Job.job_title).contains(sql_func.lower(title.strip()))
        )
        return session.execute(stmt).unique().scalar_one_or_none()
    finally:
        session.close()


def find_all_statuses() -> List[Status]:
    """Find all job statuses."""
    session = get_session()
    try:
        stmt = select(Status).where(Status.category == "job").order_by(Status.name)
        return list(session.execute(stmt).scalars().all())
    finally:
        session.close()


def find_jobs_with_status(status_names: Optional[List[str]] = None) -> List[Job]:
    """Find jobs filtered by status names."""
    session = get_session()
    try:
        stmt = select(Job).options(
            joinedload(Job.lead).joinedload(Lead.current_status),
            joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.tenant),
            joinedload(Job.lead).joinedload(Lead.deal).joinedload(Deal.current_status),
            joinedload(Job.organization).joinedload(Organization.tenant)
        ).join(Lead).join(Status).where(Lead.current_status_id.isnot(None))

        if status_names:
            stmt = stmt.where(Status.name.in_(status_names))

        stmt = stmt.order_by(Lead.created_at.desc())
        return list(session.execute(stmt).unique().scalars().all())
    finally:
        session.close()
