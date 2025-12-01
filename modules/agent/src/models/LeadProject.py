"""LeadProject junction table for many-to-many relationship."""

from sqlalchemy import Column, BigInteger, DateTime, ForeignKey, func
from models import Base


class LeadProject(Base):
    """Junction table for Lead-Project many-to-many relationship."""

    __tablename__ = "Lead_Project"

    lead_id = Column(BigInteger, ForeignKey("Lead.id", ondelete="CASCADE"), primary_key=True)
    project_id = Column(BigInteger, ForeignKey("Project.id", ondelete="CASCADE"), primary_key=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
