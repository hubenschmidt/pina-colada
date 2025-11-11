"""LeadStatus model for job lead qualification statuses."""

from datetime import datetime
from typing import TypedDict, Optional, List
from sqlalchemy import Column, String, Text, DateTime, func
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import relationship
import uuid

from agent.models import Base


class LeadStatus(Base):
    """LeadStatus SQLAlchemy model for lead qualification statuses."""
    
    __tablename__ = "LeadStatus"
    
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    name = Column(String, nullable=False, unique=True)
    description = Column(Text, nullable=True)
    created_at = Column(DateTime, server_default=func.now())
    
    # Relationship: one LeadStatus can have many Jobs
    jobs = relationship("Job", back_populates="lead_status")


# Functional data models (TypedDict)
class LeadStatusData(TypedDict, total=False):
    """Functional lead status data model."""
    id: str
    name: str
    description: Optional[str]
    created_at: Optional[datetime]


# Conversion functions
def orm_to_dict(lead_status: LeadStatus) -> LeadStatusData:
    """Convert SQLAlchemy model to functional dict."""
    return LeadStatusData(
        id=str(lead_status.id) if lead_status.id else "",
        name=lead_status.name or "",
        description=lead_status.description,
        created_at=lead_status.created_at
    )

