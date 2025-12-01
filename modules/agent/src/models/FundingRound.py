"""FundingRound model for historical funding events."""

from sqlalchemy import Column, Text, DateTime, BigInteger, Date, func, ForeignKey
from sqlalchemy.orm import relationship

from models import Base


class FundingRound(Base):
    """FundingRound SQLAlchemy model."""

    __tablename__ = "Funding_Round"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    organization_id = Column(BigInteger, ForeignKey("Organization.id", ondelete="CASCADE"), nullable=False)
    round_type = Column(Text, nullable=False)    # 'Pre-Seed', 'Seed', 'Series A', 'Series B', etc.
    amount = Column(BigInteger, nullable=True)   # USD cents
    announced_date = Column(Date, nullable=True)
    lead_investor = Column(Text, nullable=True)
    source_url = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)

    # Relationships
    organization = relationship("Organization", back_populates="funding_rounds")
