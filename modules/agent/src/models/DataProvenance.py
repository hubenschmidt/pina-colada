"""DataProvenance model for field-level tracking."""

from sqlalchemy import Column, Text, DateTime, BigInteger, Numeric, ForeignKey, func, CheckConstraint
from sqlalchemy.dialects.postgresql import JSONB
from sqlalchemy.orm import relationship

from models import Base


class DataProvenance(Base):
    """DataProvenance SQLAlchemy model."""

    __tablename__ = "Data_Provenance"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    entity_type = Column(Text, nullable=False)     # 'Organization', 'Individual'
    entity_id = Column(BigInteger, nullable=False)
    field_name = Column(Text, nullable=False)      # 'revenue_range_id', 'seniority_level', etc.
    source = Column(Text, nullable=False)          # 'clearbit', 'apollo', 'linkedin', 'agent', 'manual'
    source_url = Column(Text, nullable=True)
    confidence = Column(Numeric(3, 2), nullable=True)  # 0.00 to 1.00
    verified_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    verified_by = Column(BigInteger, ForeignKey("User.id", ondelete="SET NULL"), nullable=True)  # NULL = AI agent
    raw_value = Column(JSONB, nullable=True)       # Original value from source
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)
    created_by = Column(BigInteger, ForeignKey("User.id"), nullable=False)
    updated_by = Column(BigInteger, ForeignKey("User.id"), nullable=False)

    # Relationships
    verifier = relationship("User", foreign_keys=[verified_by])

    __table_args__ = (
        CheckConstraint('confidence >= 0 AND confidence <= 1', name='confidence_range'),
    )
