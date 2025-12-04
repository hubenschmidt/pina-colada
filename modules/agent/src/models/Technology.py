"""Technology model for tech stack lookup."""

from sqlalchemy import Column, Text, DateTime, BigInteger, ForeignKey, func, UniqueConstraint
from models import Base


class Technology(Base):
    """Technology SQLAlchemy model."""

    __tablename__ = "Technology"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    name = Column(Text, nullable=False)
    category = Column(Text, nullable=False)  # 'CRM', 'Marketing Automation', 'Cloud', 'Database', etc.
    vendor = Column(Text, nullable=True)     # 'Salesforce', 'HubSpot', 'AWS', etc.
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    __table_args__ = (
        UniqueConstraint('name', 'category', name='uq_technology_name_category'),
    )
