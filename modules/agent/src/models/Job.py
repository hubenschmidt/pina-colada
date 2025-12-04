from sqlalchemy import Column, Text, DateTime, BigInteger, ForeignKey
from sqlalchemy.orm import relationship
from models import Base

"""Data models for jobs (extends Lead via Joined Table Inheritance).

SQLAlchemy model for database persistence (unavoidable OOP requirement).
Functional TypedDict models for business logic.
"""




# SQLAlchemy model (OOP required for ORM)
class Job(Base):
    """Job SQLAlchemy model (extends Lead via Joined Table Inheritance)."""

    __tablename__ = "Job"

    id = Column(BigInteger, ForeignKey("Lead.id", ondelete="CASCADE"), primary_key=True)
    job_title = Column(Text, nullable=False)
    job_url = Column(Text, nullable=True)
    description = Column(Text, nullable=True)
    resume_date = Column(DateTime(timezone=True), nullable=True)
    salary_range = Column(Text, nullable=True)  # Legacy field, kept for backwards compat
    salary_range_id = Column(BigInteger, ForeignKey("Salary_Range.id", ondelete="SET NULL"), nullable=True)

    # Relationships
    lead = relationship("Lead", back_populates="job")
    salary_range_ref = relationship("SalaryRange")
