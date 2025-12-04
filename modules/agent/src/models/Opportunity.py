from sqlalchemy import Column, Text, BigInteger, Numeric, Date, ForeignKey, SmallInteger, CheckConstraint
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

    # Relationships
    lead = relationship("Lead", back_populates="opportunity")
