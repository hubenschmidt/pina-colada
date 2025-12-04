from sqlalchemy import Column, Text, DateTime, BigInteger, Numeric, Date, ForeignKey, SmallInteger, CheckConstraint, func
from sqlalchemy.orm import relationship
from models import Base

class Opportunity(Base):
    """Opportunity SQLAlchemy model (extends Lead)."""

    __tablename__ = "Opportunity"

    __table_args__ = (
        CheckConstraint('probability >= 0 AND probability <= 100', name='opportunity_probability_check'),
    )

    id = Column(BigInteger, ForeignKey("Lead.id", ondelete="CASCADE"), primary_key=True)
    opportunity_name = Column(Text, nullable=False)
    estimated_value = Column(Numeric(18, 2), nullable=True)
    probability = Column(SmallInteger, nullable=True)
    expected_close_date = Column(Date, nullable=True)
    description = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)
    created_by = Column(BigInteger, ForeignKey("User.id"), nullable=False)
    updated_by = Column(BigInteger, ForeignKey("User.id"), nullable=False)

    # Relationships
    lead = relationship("Lead", back_populates="opportunity")
