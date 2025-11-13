"""Repository layer for user data access."""

from typing import Optional
from sqlalchemy import select
from models.User import User, UserCreateData, UserUpdateData, dict_to_orm, update_orm_from_dict
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


def create_user(data: UserCreateData) -> User:
    """Create a new user."""
    session = get_session()
    try:
        user = dict_to_orm(data)
        session.add(user)
        session.commit()
        session.refresh(user)
        return user
    finally:
        session.close()


def update_user(user_id: int, data: UserUpdateData) -> Optional[User]:
    """Update user by ID."""
    session = get_session()
    try:
        user = session.get(User, user_id)
        if not user:
            return None

        user = update_orm_from_dict(user, data)
        session.commit()
        session.refresh(user)
        return user
    finally:
        session.close()
