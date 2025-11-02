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

# Initialize serper for web search
try:
    serper = GoogleSerperAPIWrapper()
except Exception as e:
    logger.warning(f"Could not initialize GoogleSerperAPIWrapper: {e}")
    serper = None


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
    if serper:
        tool_search = Tool(
            name="search",
            func=serper.run,
            description="Use this tool when you want to get the results of an online web search. Useful for current information or job postings.",
        )
        tools.append(tool_search)

    # Wikipedia tool
    try:
        wikipedia = WikipediaAPIWrapper()
        wiki_tool = WikipediaQueryRun(api_wrapper=wikipedia)
        tools.append(wiki_tool)
    except Exception as e:
        logger.warning(f"Could not initialize Wikipedia tool: {e}")

    logger.info(f"Initialized {len(tools)} tools")
    return tools
