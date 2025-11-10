"""Jira configuration management."""

import os
import logging
from typing import Optional
from dotenv import load_dotenv

load_dotenv(override=True)
logger = logging.getLogger(__name__)


def get_jira_url() -> Optional[str]:
    """Get Jira instance URL from environment."""
    url = os.getenv("JIRA_URL")
    if not url:
        logger.warning("JIRA_URL not set")
        return None
    return url.rstrip("/")


def get_jira_email() -> Optional[str]:
    """Get Jira email from environment."""
    email = os.getenv("JIRA_EMAIL")
    if not email:
        logger.warning("JIRA_EMAIL not set")
        return None
    return email


def get_jira_api_token() -> Optional[str]:
    """Get Jira API token from environment."""
    token = os.getenv("JIRA_API_TOKEN")
    if not token:
        logger.warning("JIRA_API_TOKEN not set")
        return None
    return token


def get_jira_project_key() -> Optional[str]:
    """Get default Jira project key from environment."""
    return os.getenv("JIRA_PROJECT_KEY", "PC")


def get_jira_board_id() -> Optional[int]:
    """Get default Jira board ID from environment."""
    board_id = os.getenv("JIRA_BOARD_ID")
    if not board_id:
        return None
    try:
        return int(board_id)
    except ValueError:
        logger.warning(f"Invalid JIRA_BOARD_ID: {board_id}")
        return None

