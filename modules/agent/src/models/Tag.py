from sqlalchemy import Column, BigInteger, String, Text, DateTime, ForeignKey, func
from sqlalchemy.orm import relationship
from models import Base


class EntityTag(Base):
    """Polymorphic junction for tagging any entity."""

    __tablename__ = "Entity_Tag"

    tag_id = Column(BigInteger, ForeignKey("Tag.id", ondelete="CASCADE"), primary_key=True)
    entity_type = Column(Text, primary_key=True)
    entity_id = Column(BigInteger, primary_key=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    tag = relationship("Tag", back_populates="entity_links")


class Tag(Base):
    """Unique tags for categorizing any entity."""

    __tablename__ = "Tag"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    name = Column(String(100), nullable=False, unique=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    entity_links = relationship("EntityTag", back_populates="tag")
