"""Industry model for account classification."""

from sqlalchemy import Column, Text, DateTime, BigInteger, ForeignKey, Table, func
from sqlalchemy.orm import relationship

from models import Base


# Join table for many-to-many relationship
Account_Industry = Table(
    "Account_Industry",
    Base.metadata,
    Column("account_id", BigInteger, ForeignKey("Account.id", ondelete="CASCADE"), primary_key=True),
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
    accounts = relationship("Account", secondary=Account_Industry, back_populates="industries")
