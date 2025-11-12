"""Data models for applied jobs.

SQLAlchemy model for database persistence (unavoidable OOP requirement).
Functional TypedDict models for business logic.
"""

from datetime import datetime
from typing import TypedDict, Optional
from sqlalchemy import Column, String, DateTime, Text, CheckConstraint, func, ForeignKey
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import relationship
import uuid

from models import Base


# SQLAlchemy model (required for ORM - unavoidable OOP)
class Job(Base):
    """Job SQLAlchemy model for database persistence."""
    
    __tablename__ = "Job"  # Capitalized table name (quoted in SQL)
    
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    company = Column(String, nullable=False)
    job_title = Column(String, nullable=False)
    date = Column(DateTime, server_default=func.now())  # Renamed from application_date
    status = Column(String, default="applied")
    job_url = Column(Text, nullable=True)
    notes = Column(Text, nullable=True)
    resume = Column(DateTime, nullable=True)  # New column
    salary_range = Column(Text, nullable=True)
    source = Column(String, default="manual")
    lead_status_id = Column(UUID(as_uuid=True), ForeignKey("LeadStatus.id"), nullable=True)
    
    # Relationship: many Jobs can have one LeadStatus
    lead_status = relationship("LeadStatus", back_populates="jobs")
    
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
    date: Optional[datetime]  # Renamed from application_date
    status: str
    job_url: Optional[str]
    notes: Optional[str]
    resume: Optional[datetime]  # New column
    salary_range: Optional[str]
    source: str
    lead_status_id: Optional[str]
    lead_status: Optional[dict]  # Nested LeadStatusData when loaded
    created_at: Optional[datetime]
    updated_at: Optional[datetime]


class JobCreateData(TypedDict, total=False):
    """Functional job creation data model."""
    company: str
    job_title: str
    date: Optional[datetime]  # Renamed from application_date
    job_url: Optional[str]
    notes: Optional[str]
    resume: Optional[datetime]  # New column
    salary_range: Optional[str]
    status: str
    source: str
    lead_status_id: Optional[str]


class JobUpdateData(TypedDict, total=False):
    """Functional job update data model."""
    company: Optional[str]
    job_title: Optional[str]
    date: Optional[datetime]  # Renamed from application_date
    job_url: Optional[str]
    notes: Optional[str]
    resume: Optional[datetime]  # New column
    salary_range: Optional[str]
    status: Optional[str]
    source: Optional[str]
    lead_status_id: Optional[str]


# Conversion functions
def orm_to_dict(job: Job) -> JobData:
    """Convert SQLAlchemy model to functional dict."""
    from models.LeadStatus import orm_to_dict as lead_status_to_dict
    
    result = JobData(
        id=str(job.id) if job.id else "",
        company=job.company or "",
        job_title=job.job_title or "",
        date=job.date,  # Renamed from application_date
        status=job.status or "applied",
        job_url=job.job_url,
        notes=job.notes,
        resume=job.resume,  # New column
        salary_range=job.salary_range,
        source=job.source or "manual",
        lead_status_id=str(job.lead_status_id) if job.lead_status_id else None,
        created_at=job.created_at,
        updated_at=job.updated_at
    )
    
    # Include lead_status if relationship is loaded
    if job.lead_status:
        result["lead_status"] = lead_status_to_dict(job.lead_status)
    
    return result


def dict_to_orm(data: JobCreateData) -> Job:
    """Convert functional dict to SQLAlchemy model."""
    import uuid
    lead_status_id = data.get("lead_status_id")
    return Job(
        company=data.get("company", ""),
        job_title=data.get("job_title", ""),
        date=data.get("date"),  # Renamed from application_date
        job_url=data.get("job_url"),
        notes=data.get("notes"),
        resume=data.get("resume"),  # New column
        salary_range=data.get("salary_range"),
        status=data.get("status", "applied"),
        source=data.get("source", "manual"),
        lead_status_id=uuid.UUID(lead_status_id) if lead_status_id else None
    )


def update_orm_from_dict(job: Job, data: JobUpdateData) -> Job:
    """Update SQLAlchemy model from functional dict."""
    import uuid
    if "company" in data and data["company"] is not None:
        job.company = data["company"]
    if "job_title" in data and data["job_title"] is not None:
        job.job_title = data["job_title"]
    if "date" in data:
        job.date = data["date"]  # Renamed from application_date
    if "job_url" in data:
        job.job_url = data["job_url"]
    if "notes" in data:
        job.notes = data["notes"]
    if "resume" in data:
        job.resume = data["resume"]  # New column
    if "salary_range" in data:
        job.salary_range = data["salary_range"]
    if "status" in data and data["status"] is not None:
        job.status = data["status"]
    if "source" in data and data["source"] is not None:
        job.source = data["source"]
    if "lead_status_id" in data:
        job.lead_status_id = uuid.UUID(data["lead_status_id"]) if data["lead_status_id"] else None
    return job

