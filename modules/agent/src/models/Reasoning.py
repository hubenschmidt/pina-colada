from sqlalchemy import Column, Text, DateTime, BigInteger, UniqueConstraint, func
from sqlalchemy.dialects.postgresql import JSONB
from models import Base


class Reasoning(Base):
    """Schema registry for AI agents - maps reasoning contexts to database tables."""

    __tablename__ = "Reasoning"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    type = Column(Text, nullable=False)  # 'crm', 'finance', etc.
    table_name = Column(Text, nullable=False)
    description = Column(Text, nullable=True)
    schema_hint = Column(JSONB, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)

    __table_args__ = (
        UniqueConstraint("type", "table_name", name="uq_reasoning_type_table"),
    )
