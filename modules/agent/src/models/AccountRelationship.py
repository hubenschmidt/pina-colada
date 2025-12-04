"""AccountRelationship model for linking accounts to each other."""

from sqlalchemy import Column, Text, DateTime, BigInteger, ForeignKey, func
from sqlalchemy.orm import relationship
from models import Base


class AccountRelationship(Base):
    """Relationship between two accounts (individuals or organizations)."""

    __tablename__ = "Account_Relationship"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    from_account_id = Column(BigInteger, ForeignKey("Account.id", ondelete="CASCADE"), nullable=False)
    to_account_id = Column(BigInteger, ForeignKey("Account.id", ondelete="CASCADE"), nullable=False)
    relationship_type = Column(Text, nullable=True)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)
    created_by = Column(BigInteger, ForeignKey("User.id"), nullable=False)
    updated_by = Column(BigInteger, ForeignKey("User.id"), nullable=False)

    # Relationships
    from_account = relationship("Account", foreign_keys=[from_account_id], backref="outgoing_relationships")
    to_account = relationship("Account", foreign_keys=[to_account_id], backref="incoming_relationships")
