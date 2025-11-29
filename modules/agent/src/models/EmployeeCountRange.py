"""EmployeeCountRange model for organization size brackets."""

from sqlalchemy import Column, BigInteger, Text, Integer, DateTime
from sqlalchemy.sql import func
from models import Base


class EmployeeCountRange(Base):
    """Lookup table for employee count brackets (e.g., 1-10, 11-50)."""

    __tablename__ = "EmployeeCountRange"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    label = Column(Text, nullable=False, unique=True)
    min_value = Column(Integer, nullable=True)
    max_value = Column(Integer, nullable=True)
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
