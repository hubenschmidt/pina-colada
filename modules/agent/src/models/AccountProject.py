"""AccountProject model - junction table for Account-Project association."""

from sqlalchemy import Column, BigInteger, ForeignKey, DateTime
from sqlalchemy.sql import func
from models import Base


class AccountProject(Base):
    """Junction table for Account-Project many-to-many relationship.

    Used for reporting purposes only - does not affect scope/filtering.
    """
    __tablename__ = "Account_Project"

    account_id = Column(
        BigInteger,
        ForeignKey("Account.id", ondelete="CASCADE"),
        primary_key=True
    )
    project_id = Column(
        BigInteger,
        ForeignKey("Project.id", ondelete="CASCADE"),
        primary_key=True
    )
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)
    created_by = Column(BigInteger, ForeignKey("User.id"), nullable=False)
    updated_by = Column(BigInteger, ForeignKey("User.id"), nullable=False)
