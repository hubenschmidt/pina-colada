from sqlalchemy import Column, BigInteger, String, Text, DateTime, ForeignKey, Table, func
from models import Base


# Polymorphic join table for tagging any entity
EntityTag = Table(
    "EntityTag",
    Base.metadata,
    Column("tag_id", BigInteger, ForeignKey("Tag.id", ondelete="CASCADE"), primary_key=True),
    Column("entity_type", Text, primary_key=True),
    Column("entity_id", BigInteger, primary_key=True),
    Column("created_at", DateTime(timezone=True), server_default=func.now(), nullable=False),
)


class Tag(Base):
    """Unique tags for categorizing any entity."""

    __tablename__ = "Tag"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    name = Column(String(100), nullable=False, unique=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)
