"""Node configuration model for storing system prompts."""

from sqlalchemy import Column, Text, DateTime, BigInteger, Boolean, ForeignKey, func
from models import Base


class NodeConfig(Base):
    """NodeConfig SQLAlchemy model for storing system prompts."""

    __tablename__ = "NodeConfig"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    node_type = Column(Text, nullable=False)  # evaluator, worker, orchestrator
    node_name = Column(Text, nullable=False)  # career_evaluator, job_hunter, etc.
    system_prompt = Column(Text, nullable=False)
    is_active = Column(Boolean, default=True, nullable=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)
    updated_by = Column(Text, nullable=False)


class NodeConfigHistory(Base):
    """NodeConfigHistory SQLAlchemy model for audit trail."""

    __tablename__ = "NodeConfigHistory"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    config_id = Column(BigInteger, ForeignKey("NodeConfig.id", ondelete="CASCADE"), nullable=False)
    node_type = Column(Text, nullable=False)
    node_name = Column(Text, nullable=False)
    system_prompt = Column(Text, nullable=False)
    changed_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    changed_by = Column(Text, nullable=False)
    change_type = Column(Text, nullable=False)  # create, update, delete
