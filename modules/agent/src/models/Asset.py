from sqlalchemy import Column, Text, DateTime, BigInteger, String, func
from sqlalchemy.orm import relationship
from models import Base


class Asset(Base):
    """Polymorphic content storage for AI agent context."""

    __tablename__ = "Asset"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    entity_type = Column(String(50), nullable=False)  # 'lead', 'job', 'organization', 'user'
    entity_id = Column(BigInteger, nullable=False)
    content = Column(Text, nullable=False)
    content_type = Column(String(100), nullable=False)  # MIME type
    filename = Column(String(255), nullable=True)
    checksum = Column(String(64), nullable=True)  # SHA-256 for deduplication
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    tags = relationship("Tag", secondary="AssetTag", back_populates="assets")
