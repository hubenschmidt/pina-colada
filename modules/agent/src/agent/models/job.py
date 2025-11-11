"""Data models for applied jobs.

SQLAlchemy model for database persistence (unavoidable OOP requirement).
Functional TypedDict models for business logic.
"""

from datetime import datetime
from typing import TypedDict, Optional
from sqlalchemy import Column, String, DateTime, Text, CheckConstraint, func
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import declarative_base
import uuid

Base = declarative_base()


# SQLAlchemy model (required for ORM - unavoidable OOP)
class AppliedJob(Base):
    """Applied job SQLAlchemy model for database persistence."""
    
    __tablename__ = "applied_jobs"
    
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    company = Column(String, nullable=False)
    job_title = Column(String, nullable=False)
    application_date = Column(DateTime, server_default=func.now())
    status = Column(String, default="applied")
    job_url = Column(Text, nullable=True)
    location = Column(Text, nullable=True)
    salary_range = Column(Text, nullable=True)
    notes = Column(Text, nullable=True)
    source = Column(String, default="manual")
    
    __table_args__ = (
        CheckConstraint("status IN ('applied', 'interviewing', 'rejected', 'offer', 'accepted')", name='check_status'),
        CheckConstraint("source IN ('manual', 'agent')", name='check_source'),
    )
    created_at = Column(DateTime, server_default=func.now())
    updated_at = Column(DateTime, server_default=func.now(), onupdate=func.now())


# Functional data models (TypedDict)
class JobData(TypedDict, total=False):
    """Functional job data model for business logic."""
    id: str
    company: str
    job_title: str
    application_date: Optional[datetime]
    status: str
    job_url: Optional[str]
    location: Optional[str]
    salary_range: Optional[str]
    notes: Optional[str]
    source: str
    created_at: Optional[datetime]
    updated_at: Optional[datetime]


class JobCreateData(TypedDict, total=False):
    """Functional job creation data model."""
    company: str
    job_title: str
    job_url: Optional[str]
    location: Optional[str]
    salary_range: Optional[str]
    notes: Optional[str]
    status: str
    source: str


class JobUpdateData(TypedDict, total=False):
    """Functional job update data model."""
    company: Optional[str]
    job_title: Optional[str]
    job_url: Optional[str]
    location: Optional[str]
    salary_range: Optional[str]
    notes: Optional[str]
    status: Optional[str]
    source: Optional[str]


# Conversion functions
def orm_to_dict(job: AppliedJob) -> JobData:
    """Convert SQLAlchemy model to functional dict."""
    return JobData(
        id=str(job.id) if job.id else "",
        company=job.company or "",
        job_title=job.job_title or "",
        application_date=job.application_date,
        status=job.status or "applied",
        job_url=job.job_url,
        location=job.location,
        salary_range=job.salary_range,
        notes=job.notes,
        source=job.source or "manual",
        created_at=job.created_at,
        updated_at=job.updated_at
    )


def dict_to_orm(data: JobCreateData) -> AppliedJob:
    """Convert functional dict to SQLAlchemy model."""
    return AppliedJob(
        company=data.get("company", ""),
        job_title=data.get("job_title", ""),
        job_url=data.get("job_url"),
        location=data.get("location"),
        salary_range=data.get("salary_range"),
        notes=data.get("notes"),
        status=data.get("status", "applied"),
        source=data.get("source", "manual")
    )


def update_orm_from_dict(job: AppliedJob, data: JobUpdateData) -> AppliedJob:
    """Update SQLAlchemy model from functional dict."""
    if "company" in data and data["company"] is not None:
        job.company = data["company"]
    if "job_title" in data and data["job_title"] is not None:
        job.job_title = data["job_title"]
    if "job_url" in data:
        job.job_url = data["job_url"]
    if "location" in data:
        job.location = data["location"]
    if "salary_range" in data:
        job.salary_range = data["salary_range"]
    if "notes" in data:
        job.notes = data["notes"]
    if "status" in data and data["status"] is not None:
        job.status = data["status"]
    if "source" in data and data["source"] is not None:
        job.source = data["source"]
    return job
