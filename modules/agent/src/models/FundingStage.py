"""FundingStage model for organization funding stages."""

from sqlalchemy import Column, BigInteger, Text, Integer, DateTime
from sqlalchemy.sql import func
from models import Base


class FundingStage(Base):
    """Lookup table for funding stages (e.g., Seed, Series A)."""

    __tablename__ = "Funding_Stage"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    label = Column(Text, nullable=False, unique=True)
    display_order = Column(Integer, nullable=False, default=0)
    created_at = Column(
        DateTime(timezone=True), server_default=func.now(), nullable=False
    )
    updated_at = Column(
        DateTime(timezone=True),
        server_default=func.now(),
        onupdate=func.now(),
        nullable=False,
    )
