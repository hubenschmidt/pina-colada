"""Example usage of Jira integration."""

import logging
from agent.integrations.jira import fetch_tickets, parse_ticket
from agent.integrations.jira.config import get_jira_board_id

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


def example_fetch_and_parse():
    """Example: Fetch and parse tickets from board."""
    board_id = get_jira_board_id()
    if not board_id:
        logger.error("JIRA_BOARD_ID not configured")
        return
    
    tickets = fetch_tickets(board_id=board_id, status="To Do")
    logger.info(f"Fetched {len(tickets)} tickets")
    
    for ticket_raw in tickets:
        parsed = parse_ticket(ticket_raw)
        logger.info(f"Ticket {parsed['key']}: {parsed['summary']}")
        logger.info(f"  Status: {parsed['status']}")
        logger.info(f"  Description: {parsed['description'][:100]}...")


if __name__ == "__main__":
    example_fetch_and_parse()

