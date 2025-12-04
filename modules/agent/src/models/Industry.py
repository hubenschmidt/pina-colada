"""Industry model for account classification."""

from sqlalchemy import Column, Text, DateTime, BigInteger, ForeignKey, func
from sqlalchemy.orm import relationship
from models import Base


class AccountIndustry(Base):
    """Junction table for Account-Industry many-to-many relationship."""

    __tablename__ = "Account_Industry"

    account_id = Column(BigInteger, ForeignKey("Account.id", ondelete="CASCADE"), primary_key=True)
    industry_id = Column(BigInteger, ForeignKey("Industry.id", ondelete="CASCADE"), primary_key=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)


class Industry(Base):
    """Industry SQLAlchemy model."""

    __tablename__ = "Industry"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    name = Column(Text, nullable=False, unique=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    accounts = relationship("Account", secondary="Account_Industry", back_populates="industries")
