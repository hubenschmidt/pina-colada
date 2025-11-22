from sqlalchemy import Column, BigInteger, ForeignKey, Table
from models import Base

# Join table for Asset <-> Tag many-to-many relationship
AssetTag = Table(
    "AssetTag",
    Base.metadata,
    Column("asset_id", BigInteger, ForeignKey("Asset.id", ondelete="CASCADE"), primary_key=True),
    Column("tag_id", BigInteger, ForeignKey("Tag.id", ondelete="CASCADE"), primary_key=True),
)
