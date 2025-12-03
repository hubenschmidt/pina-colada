from sqlalchemy import Column, Text, DateTime, Boolean, BigInteger, ForeignKey, func
from sqlalchemy.orm import relationship
from models import Base

"""Status model for central status/stage definitions."""




class Status(Base):
    """Status SQLAlchemy model for central status/stage definitions."""

    __tablename__ = "Status"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    name = Column(Text, nullable=False, unique=True)
    description = Column(Text, nullable=True)
    category = Column(Text, nullable=True)  # 'job', 'lead', 'deal', 'task_status', 'task_priority'
    is_terminal = Column(Boolean, default=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)
    created_by = Column(BigInteger, ForeignKey("User.id"), nullable=False)
    updated_by = Column(BigInteger, ForeignKey("User.id"), nullable=False)

    # Relationships
    leads = relationship("Lead", back_populates="current_status", foreign_keys="Lead.current_status_id")
    deals = relationship("Deal", back_populates="current_status")
    projects = relationship("Project", back_populates="current_status")
    tasks_status = relationship("Task", back_populates="current_status", foreign_keys="Task.current_status_id")
    tasks_priority = relationship("Task", back_populates="priority", foreign_keys="Task.priority_id")
