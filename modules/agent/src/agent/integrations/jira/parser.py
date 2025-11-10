"""Ticket parsing functionality."""

import logging
from typing import Dict, Any, Optional

logger = logging.getLogger(__name__)


def _extract_text_from_atlassian_doc(doc: Dict[str, Any]) -> str:
    """Extract plain text from Atlassian Document Format."""
    if not isinstance(doc, dict):
        return ""
    
    content = doc.get("content", [])
    if not isinstance(content, list):
        return ""
    
    text_parts = []
    for item in content:
        if not isinstance(item, dict):
            continue
        
        item_type = item.get("type")
        if item_type == "paragraph":
            para_content = item.get("content", [])
            for para_item in para_content:
                if isinstance(para_item, dict) and para_item.get("type") == "text":
                    text_parts.append(para_item.get("text", ""))
        elif item_type == "text":
            text_parts.append(item.get("text", ""))
    
    return "\n".join(text_parts).strip()


def _get_field(issue: Dict[str, Any], field_path: str) -> Any:
    """Get nested field value from issue."""
    if not issue or not field_path:
        return None
    
    parts = field_path.split(".")
    value = issue.get("fields", {})
    
    for part in parts:
        if not isinstance(value, dict):
            return None
        value = value.get(part)
    
    return value


def parse_ticket(issue: Dict[str, Any]) -> Dict[str, Any]:
    """Parse Jira issue into structured format."""
    if not issue:
        return {}
    
    key = issue.get("key", "")
    summary = _get_field(issue, "summary") or ""
    description_raw = _get_field(issue, "description")
    description = _extract_text_from_atlassian_doc(description_raw) if description_raw else ""
    
    status = _get_field(issue, "status.name") or ""
    priority = _get_field(issue, "priority.name") or ""
    labels = _get_field(issue, "labels") or []
    components = _get_field(issue, "components") or []
    component_names = [c.get("name", "") for c in components if isinstance(c, dict)]
    assignee = _get_field(issue, "assignee")
    assignee_name = assignee.get("displayName", "") if isinstance(assignee, dict) else ""
    
    return {
        "key": key,
        "summary": summary,
        "description": description,
        "status": status,
        "priority": priority,
        "labels": labels if isinstance(labels, list) else [],
        "components": component_names,
        "assignee": assignee_name,
        "raw": issue,
    }

