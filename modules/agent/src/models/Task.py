from sqlalchemy import Column, Text, DateTime, BigInteger, ForeignKey, Date, func
from sqlalchemy.orm import relationship
from models import Base

"""Data models for tasks.

SQLAlchemy model for database persistence (unavoidable OOP requirement).
Functional TypedDict models for business logic.
"""




# SQLAlchemy model (OOP required for ORM)
class Task(Base):
    """Task SQLAlchemy model."""

    __tablename__ = "Task"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    taskable_type = Column(Text, nullable=True)  # 'Deal', 'Lead', 'Job', 'Project', 'Organization', 'Individual', 'Contact'
    taskable_id = Column(BigInteger, nullable=True)
    title = Column(Text, nullable=False)
    description = Column(Text, nullable=True)
    current_status_id = Column(BigInteger, ForeignKey("Status.id", ondelete="SET NULL"), nullable=True)
    priority_id = Column(BigInteger, ForeignKey("Status.id", ondelete="SET NULL"), nullable=True)
    due_date = Column(Date, nullable=True)
    completed_at = Column(DateTime(timezone=True), nullable=True)
    assigned_to_individual_id = Column(BigInteger, ForeignKey("Individual.id", ondelete="SET NULL"), nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    current_status = relationship("Status", back_populates="tasks_status", foreign_keys=[current_status_id])
    priority = relationship("Status", back_populates="tasks_priority", foreign_keys=[priority_id])
