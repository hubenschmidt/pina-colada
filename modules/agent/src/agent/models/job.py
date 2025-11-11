"""SQLAlchemy model for applied_jobs table."""

from datetime import datetime
from sqlalchemy import Column, String, DateTime, Text, CheckConstraint, func
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import declarative_base
import uuid

Base = declarative_base()


class AppliedJob(Base):
    """Applied job model."""
    
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

