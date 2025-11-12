"""Data models for jobs (extends Lead via Joined Table Inheritance).

SQLAlchemy model for database persistence (unavoidable OOP requirement).
Functional TypedDict models for business logic.
"""

from datetime import datetime
from typing import TypedDict, Optional
from sqlalchemy import Column, Text, DateTime, BigInteger, ForeignKey, func
from sqlalchemy.orm import relationship

from models import Base


# SQLAlchemy model (OOP required for ORM)
class Job(Base):
    """Job SQLAlchemy model (extends Lead via Joined Table Inheritance)."""

    __tablename__ = "Job"

    id = Column(BigInteger, ForeignKey("Lead.id", ondelete="CASCADE"), primary_key=True)
    organization_id = Column(BigInteger, ForeignKey("Organization.id", ondelete="CASCADE"), nullable=False)
    job_title = Column(Text, nullable=False)
    job_url = Column(Text, nullable=True)
    notes = Column(Text, nullable=True)
    resume_date = Column(DateTime(timezone=True), nullable=True)
    salary_range = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    lead = relationship("Lead", back_populates="job")
    organization = relationship("Organization", back_populates="jobs")


# Functional data models (TypedDict)
class JobData(TypedDict, total=False):
    """Functional job data model for business logic."""
    id: int
    organization_id: int
    organization: Optional[dict]  # Nested OrganizationData when loaded
    job_title: str
    job_url: Optional[str]
    notes: Optional[str]
    resume_date: Optional[datetime]
    salary_range: Optional[str]
    created_at: Optional[datetime]
    updated_at: Optional[datetime]
    # Lead fields (when joined)
    lead: Optional[dict]  # Nested LeadData when loaded
    deal_id: Optional[int]
    title: Optional[str]
    description: Optional[str]
    source: Optional[str]
    current_status_id: Optional[int]
    current_status: Optional[dict]
    owner_user_id: Optional[int]


class JobCreateData(TypedDict, total=False):
    """Functional job creation data model.

    Note: When creating a Job, you must also create the corresponding Lead record first,
    or provide lead data to create both atomically.
    """
    # Job-specific fields
    organization_id: int
    job_title: str
    job_url: Optional[str]
    notes: Optional[str]
    resume_date: Optional[datetime]
    salary_range: Optional[str]
    # Lead fields (for atomic creation)
    deal_id: int
    title: str
    description: Optional[str]
    source: Optional[str]
    current_status_id: Optional[int]
    owner_user_id: Optional[int]


class JobUpdateData(TypedDict, total=False):
    """Functional job update data model."""
    organization_id: Optional[int]
    job_title: Optional[str]
    job_url: Optional[str]
    notes: Optional[str]
    resume_date: Optional[datetime]
    salary_range: Optional[str]
    # Lead fields (for updating the parent Lead)
    title: Optional[str]
    description: Optional[str]
    source: Optional[str]
    current_status_id: Optional[int]
    owner_user_id: Optional[int]


# Conversion functions
def orm_to_dict(job: Job, include_lead: bool = True) -> JobData:
    """Convert SQLAlchemy model to functional dict."""
    from models.Organization import orm_to_dict as org_to_dict
    from models.Lead import orm_to_dict as lead_to_dict

    result = JobData(
        id=job.id,
        organization_id=job.organization_id,
        job_title=job.job_title or "",
        job_url=job.job_url,
        notes=job.notes,
        resume_date=job.resume_date,
        salary_range=job.salary_range,
        created_at=job.created_at,
        updated_at=job.updated_at
    )

    # Include organization if relationship is loaded
    if job.organization:
        result["organization"] = org_to_dict(job.organization)

    # Include lead if relationship is loaded and requested
    if include_lead and job.lead:
        result["lead"] = lead_to_dict(job.lead)
        # Also flatten key lead fields to top level for convenience
        result["deal_id"] = job.lead.deal_id
        result["title"] = job.lead.title
        result["description"] = job.lead.description
        result["source"] = job.lead.source
        result["current_status_id"] = job.lead.current_status_id
        result["owner_user_id"] = job.lead.owner_user_id
        if job.lead.current_status:
            from models.Status import orm_to_dict as status_to_dict
            result["current_status"] = status_to_dict(job.lead.current_status)

    return result


def dict_to_orm(data: JobCreateData) -> Job:
    """Convert functional dict to SQLAlchemy model.

    Note: This only creates the Job record. The corresponding Lead record
    must be created separately or you should use a service function to
    create both atomically.
    """
    return Job(
        organization_id=data.get("organization_id"),
        job_title=data.get("job_title", ""),
        job_url=data.get("job_url"),
        notes=data.get("notes"),
        resume_date=data.get("resume_date"),
        salary_range=data.get("salary_range")
    )


def update_orm_from_dict(job: Job, data: JobUpdateData) -> Job:
    """Update SQLAlchemy model from functional dict."""
    if "organization_id" in data and data["organization_id"] is not None:
        job.organization_id = data["organization_id"]
    if "job_title" in data and data["job_title"] is not None:
        job.job_title = data["job_title"]
    if "job_url" in data:
        job.job_url = data["job_url"]
    if "notes" in data:
        job.notes = data["notes"]
    if "resume_date" in data:
        job.resume_date = data["resume_date"]
    if "salary_range" in data:
        job.salary_range = data["salary_range"]
    return job
