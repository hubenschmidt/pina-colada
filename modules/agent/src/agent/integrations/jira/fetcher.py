"""Ticket fetching functionality."""

import logging
from typing import Dict, Any, Optional, List
from agent.integrations.jira.client import get_board_issues, search_issues, get_issue
from agent.integrations.jira.config import get_jira_board_id, get_jira_project_key

logger = logging.getLogger(__name__)


def _extract_issues(response: Optional[Dict[str, Any]]) -> List[Dict[str, Any]]:
    """Extract issues list from API response."""
    if not response:
        return []
    
    issues = response.get("issues", [])
    if not isinstance(issues, list):
        return []
    
    return issues


def fetch_tickets(
    board_id: Optional[int] = None,
    project_key: Optional[str] = None,
    status: Optional[str] = None,
    max_results: int = 50,
) -> List[Dict[str, Any]]:
    """Fetch tickets from Jira."""
    if board_id:
        response = get_board_issues(board_id, status)
        return _extract_issues(response)
    
    if not project_key:
        project_key = get_jira_project_key()
    
    if not project_key:
        logger.error("No board_id or project_key provided")
        return []
    
    jql = f"project = {project_key}"
    if status:
        jql += f" AND status = {status}"
    
    response = search_issues(jql, max_results)
    return _extract_issues(response)


def fetch_ticket_by_key(issue_key: str) -> Optional[Dict[str, Any]]:
    """Fetch a single ticket by issue key."""
    if not issue_key:
        return None
    
    return get_issue(issue_key)


def fetch_tickets_by_jql(jql: str, max_results: int = 50) -> List[Dict[str, Any]]:
    """Fetch tickets using custom JQL query."""
    if not jql:
        return []
    
    response = search_issues(jql, max_results)
    return _extract_issues(response)

