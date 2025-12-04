"""EntityAsset model - polymorphic junction for linking assets to entities."""

from sqlalchemy import Column, BigInteger, Text, DateTime, ForeignKey, Table, func
from models import Base

# Polymorphic join table for linking any asset to any entity
EntityAsset = Table(
    "Entity_Asset",
    Base.metadata,
    Column("asset_id", BigInteger, ForeignKey("Asset.id", ondelete="CASCADE"), primary_key=True),
    Column("entity_type", Text, primary_key=True),  # 'Project', 'Lead', 'Account', etc.
    Column("entity_id", BigInteger, primary_key=True),
    Column("created_at", DateTime(timezone=True), server_default=func.now(), nullable=False),
    Column("updated_at", DateTime(timezone=True), server_default=func.now(), nullable=False),
    Column("created_by", BigInteger, ForeignKey("User.id"), nullable=False),
    Column("updated_by", BigInteger, ForeignKey("User.id"), nullable=False),
)
