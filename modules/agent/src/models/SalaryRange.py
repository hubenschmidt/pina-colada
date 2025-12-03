"""SalaryRange model for job salary brackets."""

from sqlalchemy import Column, Text, DateTime, BigInteger, Integer, ForeignKey, func
from models import Base


class SalaryRange(Base):
    """SalaryRange SQLAlchemy model."""

    __tablename__ = "Salary_Range"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    label = Column(Text, nullable=False, unique=True)
    min_value = Column(Integer, nullable=True)
    max_value = Column(Integer, nullable=True)
    display_order = Column(Integer, nullable=False, default=0)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)
    created_by = Column(BigInteger, ForeignKey("User.id"), nullable=False)
    updated_by = Column(BigInteger, ForeignKey("User.id"), nullable=False)
