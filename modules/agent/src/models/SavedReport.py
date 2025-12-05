"""SavedReport model for storing custom report definitions."""

from sqlalchemy import Column, Text, BigInteger, ForeignKey, DateTime, func
from sqlalchemy.dialects.postgresql import JSONB
from sqlalchemy.orm import relationship
from models import Base


class SavedReport(Base):
    """SavedReport SQLAlchemy model for custom report definitions.

    Reports can be scoped to multiple projects via the SavedReportProject junction table.
    No projects = global report (visible regardless of selected project)
    One or more projects = project-specific report (visible when any of those projects is selected)
    """

    __tablename__ = "Saved_Report"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    tenant_id = Column(BigInteger, ForeignKey("Tenant.id", ondelete="CASCADE"), nullable=False)
    name = Column(Text, nullable=False)
    description = Column(Text, nullable=True)
    query_definition = Column(JSONB, nullable=False)
    created_by = Column(BigInteger, ForeignKey("Individual.id", ondelete="SET NULL"), nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    tenant = relationship("Tenant", foreign_keys=[tenant_id])
    projects = relationship("Project", secondary="Saved_Report_Project", backref="saved_reports")
