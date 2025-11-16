"""Database connection utilities."""

import logging
import os
from sqlalchemy import create_engine
from sqlalchemy.orm import Session, sessionmaker
from models import Base

logger = logging.getLogger(__name__)

# Module-level state
_engine = None
_session_factory = None


def get_connection_string() -> str:
    """Get Postgres connection string from environment."""
    db_host = os.getenv("POSTGRES_HOST")
    db_port = os.getenv("POSTGRES_PORT")
    db_user = os.getenv("POSTGRES_USER")
    db_password = os.getenv("POSTGRES_PASSWORD")
    db_name = os.getenv("POSTGRES_DB")

    return f"postgresql://{db_user}:{db_password}@{db_host}:{db_port}/{db_name}"


def get_engine():
    """Get or create SQLAlchemy engine."""
    global _engine
    if _engine is not None:
        return _engine

    conn_string = get_connection_string()
    _engine = create_engine(conn_string, pool_pre_ping=True)
    Base.metadata.create_all(_engine)
    return _engine


def get_session() -> Session:
    """Get database session."""
    global _session_factory
    if _session_factory is None:
        engine = get_engine()
        _session_factory = sessionmaker(bind=engine)

    return _session_factory()
