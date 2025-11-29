"""SavedReportProject junction table for many-to-many relationship."""

from sqlalchemy import Column, BigInteger, ForeignKey, DateTime
from sqlalchemy.sql import func

from models import Base


class SavedReportProject(Base):
    """Junction table linking SavedReport to Project (many-to-many)."""

    __tablename__ = "SavedReportProject"

    saved_report_id = Column(
        BigInteger,
        ForeignKey("SavedReport.id", ondelete="CASCADE"),
        primary_key=True
    )
    project_id = Column(
        BigInteger,
        ForeignKey("Project.id", ondelete="CASCADE"),
        primary_key=True
    )
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
