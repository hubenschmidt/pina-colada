"""Jira REST API client."""

import logging
import time
from typing import Dict, Any, Optional, List
import requests
from agent.integrations.jira.config import (
    get_jira_url,
    get_jira_email,
    get_jira_api_token,
)

logger = logging.getLogger(__name__)

# Rate limiting: Jira allows 100 requests per minute per user
RATE_LIMIT_REQUESTS = 100
RATE_LIMIT_WINDOW = 60
_request_times: List[float] = []


def _rate_limit():
    """Enforce rate limiting by sleeping if needed."""
    now = time.time()
    _request_times.append(now)
    
    # Keep only requests within the window
    cutoff = now - RATE_LIMIT_WINDOW
    _request_times[:] = [t for t in _request_times if t > cutoff]
    
    if len(_request_times) >= RATE_LIMIT_REQUESTS:
        sleep_time = RATE_LIMIT_WINDOW - (now - _request_times[0]) + 1
        if sleep_time > 0:
            logger.info(f"Rate limit reached, sleeping {sleep_time:.1f}s")
            time.sleep(sleep_time)


def _get_auth() -> Optional[tuple]:
    """Get authentication tuple (email, token)."""
    email = get_jira_email()
    token = get_jira_api_token()
    
    if not email or not token:
        return None
    
    return (email, token)


def _make_request(method: str, endpoint: str, **kwargs) -> Optional[Dict[str, Any]]:
    """Make authenticated request to Jira API."""
    url = get_jira_url()
    if not url:
        logger.error("JIRA_URL not configured")
        return None
    
    auth = _get_auth()
    if not auth:
        logger.error("Jira credentials not configured")
        return None
    
    _rate_limit()
    
    full_url = f"{url}/rest/api/3/{endpoint.lstrip('/')}"
    headers = kwargs.pop("headers", {})
    headers.setdefault("Accept", "application/json")
    
    try:
        response = requests.request(
            method,
            full_url,
            auth=auth,
            headers=headers,
            timeout=30,
            **kwargs,
        )
        response.raise_for_status()
        return response.json()
    except requests.exceptions.RequestException as e:
        logger.error(f"Jira API request failed: {e}")
        return None


def get_issue(issue_key: str) -> Optional[Dict[str, Any]]:
    """Fetch a single issue by key."""
    if not issue_key:
        return None
    
    return _make_request("GET", f"issue/{issue_key}")


def search_issues(jql: str, max_results: int = 50) -> Optional[Dict[str, Any]]:
    """Search issues using JQL."""
    if not jql:
        return None
    
    params = {
        "jql": jql,
        "maxResults": min(max_results, 100),
        "fields": "summary,description,status,priority,labels,components,assignee",
    }
    
    return _make_request("GET", "search", params=params)


def get_board_issues(board_id: int, status: Optional[str] = None) -> Optional[Dict[str, Any]]:
    """Fetch issues from a board."""
    if not board_id:
        return None
    
    jql = f"board = {board_id}"
    if status:
        jql += f" AND status = {status}"
    
    return search_issues(jql)


def update_issue_status(issue_key: str, status_id: str) -> Optional[Dict[str, Any]]:
    """Update issue status."""
    if not issue_key or not status_id:
        return None
    
    data = {"transition": {"id": status_id}}
    return _make_request("POST", f"issue/{issue_key}/transitions", json=data)


def add_comment(issue_key: str, body: str) -> Optional[Dict[str, Any]]:
    """Add comment to issue."""
    if not issue_key or not body:
        return None
    
    data = {"body": {"type": "doc", "version": 1, "content": [{"type": "paragraph", "content": [{"type": "text", "text": body}]}]}}
    return _make_request("POST", f"issue/{issue_key}/comment", json=data)

