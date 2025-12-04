from sqlalchemy import Column, Text, DateTime, BigInteger, ForeignKey, func
from sqlalchemy.orm import relationship
from models import Base


class IndividualRelationship(Base):
    """Relationship between two individuals."""

    __tablename__ = "Individual_Relationship"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    from_individual_id = Column(BigInteger, ForeignKey("Individual.id", ondelete="CASCADE"), nullable=False)
    to_individual_id = Column(BigInteger, ForeignKey("Individual.id", ondelete="CASCADE"), nullable=False)
    relationship_type = Column(Text, nullable=True)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    from_individual = relationship("Individual", foreign_keys=[from_individual_id], backref="outgoing_relationships")
    to_individual = relationship("Individual", foreign_keys=[to_individual_id], backref="incoming_relationships")
