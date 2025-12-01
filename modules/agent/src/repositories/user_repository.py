"""Repository layer for user data access."""

from typing import Optional, Dict, Any
from sqlalchemy import select
from models.User import User
from models.Individual import Individual
from models.Account import Account
from lib.db import async_get_session


async def find_user_by_auth0_sub(auth0_sub: str) -> Optional[User]:
    """Find user by Auth0 subject ID."""
    async with async_get_session() as session:
        stmt = select(User).where(User.auth0_sub == auth0_sub)
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def find_user_by_id(user_id: int) -> Optional[User]:
    """Find user by ID."""
    async with async_get_session() as session:
        stmt = select(User).where(User.id == user_id)
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def create_user(data: Dict[str, Any]) -> User:
    """Create a new user with associated Individual and Account."""
    async with async_get_session() as session:
        # Create Account for the Individual
        account = Account(name=data.get("email", "User Account"))
        session.add(account)
        await session.flush()

        # Create Individual profile
        email = data.get("email", "")
        individual = Individual(
            account_id=account.id,
            first_name=data.get("first_name", ""),
            last_name=data.get("last_name", ""),
            email=email,
        )
        session.add(individual)
        await session.flush()

        # Create User linked to Individual
        user = User(**data)
        user.individual_id = individual.id
        session.add(user)

        await session.commit()
        await session.refresh(user)
        return user


async def find_user_by_email(email: str) -> Optional[User]:
    """Find user by email."""
    async with async_get_session() as session:
        stmt = select(User).where(User.email == email)
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def update_user(user_id: int, data: Dict[str, Any]) -> Optional[User]:
    """Update user by ID."""
    async with async_get_session() as session:
        user = await session.get(User, user_id)
        if not user:
            return None

        for key, value in data.items():
            if hasattr(user, key):
                setattr(user, key, value)

        await session.commit()
        await session.refresh(user)
        return user
