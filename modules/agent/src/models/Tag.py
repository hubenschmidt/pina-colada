from sqlalchemy import Column, BigInteger, String
from sqlalchemy.orm import relationship
from models import Base


class Tag(Base):
    """Unique tags for categorizing assets."""

    __tablename__ = "Tag"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    name = Column(String(100), nullable=False, unique=True)

    assets = relationship("Asset", secondary="AssetTag", back_populates="tags")
