from dotenv import load_dotenv
import os
import requests
import logging
from langchain_core.tools import Tool
from langchain_community.agent_toolkits import FileManagementToolkit
from langchain_community.tools.wikipedia.tool import WikipediaQueryRun
from langchain_community.utilities import GoogleSerperAPIWrapper
from langchain_community.utilities.wikipedia import WikipediaAPIWrapper

logger = logging.getLogger(__name__)

load_dotenv(override=True)
pushover_token = os.getenv("PUSHOVER_TOKEN")
pushover_user = os.getenv("PUSHOVER_USER")
pushover_url = "https://api.pushover.net/1/messages.json"

SERPER_API_KEY = os.getenv("SERPER_API_KEY")

serper = None

# Try init only if key is present
if SERPER_API_KEY:
    try:
        serper = GoogleSerperAPIWrapper()  # reads SERPER_API_KEY from env
        logger.info("GoogleSerperAPIWrapper initialized.")
    except Exception as e:
        logger.warning(f"Could not initialize GoogleSerperAPIWrapper: {e}")
        serper = None

# Warn if key missing
if not SERPER_API_KEY:
    logger.warning("SERPER_API_KEY is not set; web search tool will be disabled.")


def push(text: str):
    """Send a push notification to the user"""
    if not pushover_token or not pushover_user:
        logger.warning("Pushover credentials not configured")
        return "Pushover not configured"

    try:
        response = requests.post(
            pushover_url,
            data={"token": pushover_token, "user": pushover_user, "message": text},
            timeout=5,
        )
        logger.info(f"Push notification sent: {response.status_code}")
        return "success"
    except Exception as e:
        logger.error(f"Failed to send push notification: {e}")
        return f"failed: {e}"


def record_user_details(
    email: str, name: str = "Name not provided", notes: str = "not provided"
):
    """Record user contact details"""
    message = f"Recording interest from {name} with email {email} and notes {notes}"
    push(message)
    logger.info(message)
    return {"recorded": "ok", "email": email, "name": name}


def record_unknown_question(question: str):
    """Record questions that couldn't be answered"""
    message = f"Recording question that couldn't be answered: {question}"
    push(message)
    logger.info(message)
    return {"recorded": "ok", "question": question}


def _serper_search(query: str) -> str:
    if not serper:
        return "Web search is not configured on the server."
    try:
        return serper.run(query)
    except Exception as e:
        logger.error(f"Serper search failed: {e}")
        return f"Web search failed: {e}"


def get_file_tools():
    """Get file management tools for the 'me' directory"""
    try:
        toolkit = FileManagementToolkit(
            root_dir="me"
        )  # me directory contains resume info
        return toolkit.get_tools()
    except Exception as e:
        logger.warning(f"Could not initialize file tools: {e}")
        return []


async def all_tools():
    """Return all available tools for the Sidekick"""
    tools = []

    # Push notification tool
    push_tool = Tool(
        name="send_push_notification",
        func=push,
        description="Use this tool when you want to send a push notification to the owner",
    )
    tools.append(push_tool)

    # User details recording tool
    record_details_tool = Tool(
        name="record_user_details",
        func=record_user_details,
        description="Record user contact information. Requires email, optional name and notes. Use after collecting contact info from the conversation.",
    )
    tools.append(record_details_tool)

    # Unknown question recording tool
    record_question_tool = Tool(
        name="record_unknown_question",
        func=record_unknown_question,
        description="Record a question that you couldn't answer. This alerts the owner that additional information may be needed.",
    )
    tools.append(record_question_tool)

    # File management tools
    file_tools = get_file_tools()
    tools.extend(file_tools)

    # Web search tool
    web_search_tool = Tool(
        name="web_search",
        func=_serper_search,
        description="Search the web for current information (news, job postings, documentation). Input: a search query.",
    )
    tools.append(web_search_tool)

    # Wikipedia tool
    try:
        wikipedia = WikipediaAPIWrapper()
        wiki_tool = WikipediaQueryRun(api_wrapper=wikipedia)
        tools.append(wiki_tool)
    except Exception as e:
        logger.warning(f"Could not initialize Wikipedia tool: {e}")

    logger.info(f"Initialized {len(tools)} tools")
    return tools
