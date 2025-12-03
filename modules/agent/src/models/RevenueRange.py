"""RevenueRange model for company revenue brackets."""

from sqlalchemy import Column, Text, DateTime, BigInteger, Integer, ForeignKey, func

from models import Base


class RevenueRange(Base):
    """RevenueRange SQLAlchemy model."""

    __tablename__ = "Revenue_Range"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    label = Column(Text, nullable=False, unique=True)
    min_value = Column(BigInteger, nullable=True)  # USD, NULL = unbounded
    max_value = Column(BigInteger, nullable=True)  # USD, NULL = unbounded
    display_order = Column(Integer, nullable=False, default=0)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)
    created_by = Column(BigInteger, ForeignKey("User.id"), nullable=False)
    updated_by = Column(BigInteger, ForeignKey("User.id"), nullable=False)
