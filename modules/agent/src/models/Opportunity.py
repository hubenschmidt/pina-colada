from decimal import Decimal
from sqlalchemy import Column, Text, DateTime, BigInteger, Numeric, Date, ForeignKey, func
from sqlalchemy.orm import relationship
from models import Base

"""Opportunity model (extends Lead via Joined Table Inheritance)."""




class Opportunity(Base):
    """Opportunity SQLAlchemy model (extends Lead)."""

    __tablename__ = "Opportunity"

    id = Column(BigInteger, ForeignKey("Lead.id", ondelete="CASCADE"), primary_key=True)
    organization_id = Column(BigInteger, ForeignKey("Organization.id", ondelete="CASCADE"), nullable=False)
    opportunity_name = Column(Text, nullable=False)
    estimated_value = Column(Numeric(18, 2), nullable=True)
    probability = Column(Numeric(5, 2), nullable=True)
    expected_close_date = Column(Date, nullable=True)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    lead = relationship("Lead", back_populates="opportunity")
    organization = relationship("Organization", back_populates="opportunities")
