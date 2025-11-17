"""Test scraper worker in isolation."""
import pytest
from agent.workers.scraper import create_scraper_node, route_from_scraper
from agent.tools.scraper_tools import SCRAPER_TOOLS
from agent.orchestrator import _trim_messages
from langchain_core.messages import HumanMessage


@pytest.mark.asyncio
async def test_scraper_worker_invocation():
    """Test scraper worker responds to scraping request."""
    resume_context = "Will Hubenschmidt - Software Engineer with expertise in web automation"

    scraper_node = await create_scraper_node(
        tools=SCRAPER_TOOLS,
        resume_context_concise=resume_context,
        trim_messages_fn=_trim_messages,
    )

    state = {
        "resume_name": "Will",
        "success_criteria": "Automate login to the 401k provider and extract account balances",
        "messages": [
            HumanMessage(content="Please scrape the mock 401k site at http://localhost:8000/mocks/401k-rollover/ and login with demo/demo123")
        ],
        "feedback_on_work": None,
    }

    result = scraper_node(state)
    response_message = result["messages"][0]

    assert response_message.content is not None
    assert len(response_message.content) > 0
    assert response_message.tool_calls is not None


@pytest.mark.asyncio
async def test_scraper_worker_routing():
    """Test scraper worker routes correctly."""
    resume_context = "Will Hubenschmidt - Software Engineer"

    scraper_node = await create_scraper_node(
        tools=SCRAPER_TOOLS,
        resume_context_concise=resume_context,
        trim_messages_fn=_trim_messages,
    )

    state = {
        "resume_name": "Will",
        "success_criteria": "Scrape data from a website",
        "messages": [
            HumanMessage(content="Scrape http://localhost:8000/mocks/401k-rollover/")
        ],
        "feedback_on_work": None,
    }

    result = scraper_node(state)
    state["messages"].append(result["messages"][0])

    route = route_from_scraper(state)

    # Should route to tools if there are tool calls
    if result["messages"][0].tool_calls:
        assert route == "tools"
    else:
        assert route == "evaluator"
