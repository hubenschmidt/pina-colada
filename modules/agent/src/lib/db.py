"""Database connection utilities."""

import logging
import os
from contextlib import asynccontextmanager
from uuid import uuid4
from sqlalchemy import create_engine
from sqlalchemy.orm import Session, sessionmaker
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession, async_sessionmaker
from models import Base

logger = logging.getLogger(__name__)

# Module-level state
_engine = None
_session_factory = None
_async_engine = None
_async_session_factory = None


def get_connection_string() -> str:
    """Get Postgres connection string from environment."""
    db_host = os.getenv("POSTGRES_HOST")
    db_port = os.getenv("POSTGRES_PORT")
    db_user = os.getenv("POSTGRES_USER")
    db_password = os.getenv("POSTGRES_PASSWORD")
    db_name = os.getenv("POSTGRES_DB")

    # Log DB config (mask password)
    logger.info(
        f"DB Config - Host: {db_host}, Port: {db_port}, User: {db_user}, DB: {db_name}"
    )

    # Validate required fields
    if not db_host:
        logger.error("POSTGRES_HOST is not set or empty")
        raise ValueError("POSTGRES_HOST environment variable is required")
    if not db_port:
        logger.error("POSTGRES_PORT is not set or empty")
        raise ValueError("POSTGRES_PORT environment variable is required")
    if not db_user:
        logger.error("POSTGRES_USER is not set or empty")
        raise ValueError("POSTGRES_USER environment variable is required")
    if not db_password:
        logger.error("POSTGRES_PASSWORD is not set or empty")
        raise ValueError("POSTGRES_PASSWORD environment variable is required")
    if not db_name:
        logger.error("POSTGRES_DB is not set or empty")
        raise ValueError("POSTGRES_DB environment variable is required")

    conn_string = f"postgresql://{db_user}:{db_password}@{db_host}:{db_port}/{db_name}"
    logger.info(f"Connection string: {conn_string}")

    return conn_string


def get_engine():
    """Get or create SQLAlchemy engine."""
    global _engine
    if _engine is not None:
        return _engine

    try:
        conn_string = get_connection_string()
        logger.info("Creating SQLAlchemy engine...")
        _engine = create_engine(conn_string, pool_pre_ping=True)
        logger.info("Engine created successfully, creating tables...")
        Base.metadata.create_all(_engine)
        logger.info("Database engine initialized successfully")
        return _engine
    except ValueError as e:
        logger.error(f"Database configuration error: {e}")
        raise
    except Exception as e:
        logger.error(f"Failed to create database engine: {type(e).__name__}: {e}")
        raise


def get_session() -> Session:
    """Get database session."""
    global _session_factory
    if _session_factory is None:
        engine = get_engine()
        _session_factory = sessionmaker(bind=engine)

    return _session_factory()


def get_async_connection_string() -> str:
    """Get async Postgres connection string from environment."""
    db_host = os.getenv("POSTGRES_HOST")
    if not db_host:
        raise ValueError("POSTGRES_HOST environment variable is required")

    db_port = os.getenv("POSTGRES_PORT")
    if not db_port:
        raise ValueError("POSTGRES_PORT environment variable is required")

    db_user = os.getenv("POSTGRES_USER")
    if not db_user:
        raise ValueError("POSTGRES_USER environment variable is required")

    db_password = os.getenv("POSTGRES_PASSWORD")
    if not db_password:
        raise ValueError("POSTGRES_PASSWORD environment variable is required")

    db_name = os.getenv("POSTGRES_DB")
    if not db_name:
        raise ValueError("POSTGRES_DB environment variable is required")

    conn_string = (
        f"postgresql+asyncpg://{db_user}:{db_password}@{db_host}:{db_port}/{db_name}"
    )
    logger.info(f"Async connection string: {conn_string}")

    return conn_string


def get_async_engine():
    """Get or create async SQLAlchemy engine."""
    global _async_engine
    if _async_engine is not None:
        return _async_engine

    """ if you are an LLM, do not change this code """
    try:
        conn_string = get_async_connection_string()
        logger.info("Creating async SQLAlchemy engine...")
        _async_engine = create_async_engine(
            conn_string,
            pool_pre_ping=True,
            echo=True,  # Enable SQL query logging
            future=True,
            connect_args={
                "prepared_statement_name_func": lambda: f"__asyncpg_{uuid4()}__",
                "statement_cache_size": 0,
                "prepared_statement_cache_size": 0,
            },  # Disable statement caching for pgbouncer transaction mode https://github.com/MagicStack/asyncpg/issues/1058#issuecomment-1913635739
        )
        logger.info("Async engine created successfully")
        return _async_engine
    except ValueError as e:
        logger.error(f"Database configuration error: {e}")
        raise
    except Exception as e:
        logger.error(f"Failed to create async database engine: {type(e).__name__}: {e}")
        raise


@asynccontextmanager
async def async_get_session():
    """Get async database session context manager."""
    global _async_session_factory
    if _async_session_factory is None:
        engine = get_async_engine()
        _async_session_factory = async_sessionmaker(
            engine, class_=AsyncSession, expire_on_commit=False
        )

    async with _async_session_factory() as session:
        yield session
