"""Repository layer for user data access."""

from typing import Optional, Dict, Any
from sqlalchemy import select
from models.User import User
from lib.db import get_session


def find_user_by_auth0_sub(auth0_sub: str) -> Optional[User]:
    """Find user by Auth0 subject ID."""
    session = get_session()
    try:
        stmt = select(User).where(User.auth0_sub == auth0_sub)
        return session.execute(stmt).scalar_one_or_none()
    finally:
        session.close()


def find_user_by_id(user_id: int) -> Optional[User]:
    """Find user by ID."""
    session = get_session()
    try:
        stmt = select(User).where(User.id == user_id)
        return session.execute(stmt).scalar_one_or_none()
    finally:
        session.close()


def create_user(data: Dict[str, Any]) -> User:
    """Create a new user."""
    session = get_session()
    try:
        user = User(**data)
        session.add(user)
        session.commit()
        session.refresh(user)
        return user
    finally:
        session.close()


def update_user(user_id: int, data: Dict[str, Any]) -> Optional[User]:
    """Update user by ID."""
    session = get_session()
    try:
        user = session.get(User, user_id)
        if not user:
            return None

        for key, value in data.items():
            if hasattr(user, key):
                setattr(user, key, value)

        session.commit()
        session.refresh(user)
        return user
    finally:
        session.close()
