from sqlalchemy import Column, Text, DateTime, BigInteger, Boolean, ForeignKey, func
from sqlalchemy.orm import relationship
from models import Base

"""Contact model (links Individual to Organization)."""




class Contact(Base):
    """Contact SQLAlchemy model (links Individual to Organization)."""

    __tablename__ = "Contact"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    individual_id = Column(BigInteger, ForeignKey("Individual.id", ondelete="CASCADE"), nullable=False)
    organization_id = Column(BigInteger, ForeignKey("Organization.id", ondelete="SET NULL"), nullable=True)
    title = Column(Text, nullable=True)
    department = Column(Text, nullable=True)
    role = Column(Text, nullable=True)
    email = Column(Text, nullable=True)
    phone = Column(Text, nullable=True)
    is_primary = Column(Boolean, nullable=False, default=False)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    individual = relationship("Individual", back_populates="contacts")
    organization = relationship("Organization", back_populates="contacts")
