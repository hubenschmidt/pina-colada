"""RevenueRange model for money brackets (salary, deal value, etc.)."""

from sqlalchemy import Column, Text, DateTime, BigInteger, Integer, func
from models import Base


class RevenueRange(Base):
    """RevenueRange SQLAlchemy model."""

    __tablename__ = "RevenueRange"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    category = Column(Text, nullable=False)
    label = Column(Text, nullable=False)
    min_value = Column(Integer, nullable=True)
    max_value = Column(Integer, nullable=True)
    display_order = Column(Integer, nullable=False, default=0)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)
