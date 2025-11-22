"""Industry model for organization classification."""

from sqlalchemy import Column, Text, DateTime, BigInteger, ForeignKey, Table, func
from sqlalchemy.orm import relationship

from models import Base


# Join table for many-to-many relationship
Organization_Industry = Table(
    "Organization_Industry",
    Base.metadata,
    Column("organization_id", BigInteger, ForeignKey("Organization.id", ondelete="CASCADE"), primary_key=True),
    Column("industry_id", BigInteger, ForeignKey("Industry.id", ondelete="CASCADE"), primary_key=True),
)


class Industry(Base):
    """Industry SQLAlchemy model."""

    __tablename__ = "Industry"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    name = Column(Text, nullable=False, unique=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    organizations = relationship("Organization", secondary=Organization_Industry, back_populates="industries")
