from sqlalchemy import Column, BigInteger, String, Text, DateTime, ForeignKey, Table, func
from models import Base


# Polymorphic join table for tagging any entity
EntityTag = Table(
    "Entity_Tag",
    Base.metadata,
    Column("tag_id", BigInteger, ForeignKey("Tag.id", ondelete="CASCADE"), primary_key=True),
    Column("entity_type", Text, primary_key=True),
    Column("entity_id", BigInteger, primary_key=True),
    Column("created_at", DateTime(timezone=True), server_default=func.now(), nullable=False),
    Column("updated_at", DateTime(timezone=True), server_default=func.now(), nullable=False),
    Column("created_by", BigInteger, ForeignKey("User.id"), nullable=False),
    Column("updated_by", BigInteger, ForeignKey("User.id"), nullable=False),
)


class Tag(Base):
    """Unique tags for categorizing any entity."""

    __tablename__ = "Tag"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    name = Column(String(100), nullable=False, unique=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)
    created_by = Column(BigInteger, ForeignKey("User.id"), nullable=False)
    updated_by = Column(BigInteger, ForeignKey("User.id"), nullable=False)
