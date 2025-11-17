"""MCP Playwright client for LangGraph - provides browser automation tools."""

import logging
import asyncio
from typing import Any, Dict, List, Optional
from langchain_core.tools import tool

logger = logging.getLogger(__name__)


class PlaywrightMCPClient:
    """Persistent MCP client for Playwright automation."""

    def __init__(self):
        self.session = None
        self._context_stack = []
        self._initialized = False
        self._lock = asyncio.Lock()

    async def initialize(self):
        """Initialize connection to Playwright MCP server."""
        async with self._lock:
            if self._initialized:
                return

            logger.info("Starting Playwright MCP server...")

            try:
                from mcp import StdioServerParameters
                from mcp.client.stdio import stdio_client
                from mcp.client.session import ClientSession

                playwright_params = StdioServerParameters(
                    command="npx",
                    args=["@playwright/mcp@latest", "--isolated", "--no-sandbox"],
                    env={"PLAYWRIGHT_BROWSER": "chromium"},
                )

                # Store context managers for proper cleanup
                stdio_ctx = stdio_client(playwright_params)
                read, write = await stdio_ctx.__aenter__()
                self._context_stack.append(("stdio", stdio_ctx))

                # Create session
                session_ctx = ClientSession(read, write)
                self.session = await session_ctx.__aenter__()
                self._context_stack.append(("session", session_ctx))

                # Initialize session
                await self.session.initialize()

                logger.info("✓ Playwright MCP server connected")
                self._initialized = True

            except Exception as e:
                logger.error(f"Failed to initialize Playwright MCP: {e}")
                logger.exception("Full traceback:")
                await self.close()
                raise

    async def call_tool(self, tool_name: str, arguments: Dict[str, Any]) -> Any:
        """Call a Playwright MCP tool."""
        if not self._initialized:
            await self.initialize()

        try:
            logger.info(f"Calling Playwright tool: {tool_name} with args: {arguments}")
            result = await self.session.call_tool(tool_name, arguments=arguments)
            logger.info(
                f"Tool {tool_name} returned: {result.content[:200] if len(str(result.content)) > 200 else result.content}"
            )
            return result.content
        except Exception as e:
            logger.error(f"Error calling tool {tool_name}: {e}")
            logger.exception("Full traceback:")
            raise

    async def close(self):
        """Close MCP session and cleanup context managers."""
        logger.info("Closing Playwright MCP client...")

        # Close contexts in reverse order
        while self._context_stack:
            ctx_type, ctx = self._context_stack.pop()
            try:
                await ctx.__aexit__(None, None, None)
                logger.info(f"Closed {ctx_type} context")
            except Exception as e:
                logger.error(f"Error closing {ctx_type}: {e}")

        self.session = None
        self._initialized = False
        logger.info("✓ Playwright MCP client closed")


# Global MCP client instance
_mcp_client = PlaywrightMCPClient()


# Define key browser tools as LangChain tools
@tool
async def browser_navigate(url: str) -> str:
    """
    Navigate to a URL in the browser.

    Args:
        url: The URL to navigate to

    Returns:
        Success message
    """
    result = await _mcp_client.call_tool("browser_navigate", {"url": url})
    return str(result)


@tool
async def browser_click(element: str, ref: str) -> str:
    """
    Click an element on the page.

    Args:
        element: Human-readable description of the element
        ref: Exact element reference from page snapshot

    Returns:
        Success message
    """
    result = await _mcp_client.call_tool(
        "browser_click", {"element": element, "ref": ref}
    )
    return str(result)


@tool
async def browser_type(element: str, ref: str, text: str, submit: bool = False) -> str:
    """
    Type text into an element.

    Args:
        element: Human-readable description of the element
        ref: Exact element reference from page snapshot
        text: Text to type
        submit: Whether to press Enter after typing

    Returns:
        Success message
    """
    result = await _mcp_client.call_tool(
        "browser_type", {"element": element, "ref": ref, "text": text, "submit": submit}
    )
    return str(result)


@tool
async def browser_snapshot() -> str:
    """
    Capture accessibility snapshot of the current page.

    Returns page structure that you can use to find element references for clicking/typing.

    Returns:
        Page snapshot with element references
    """
    result = await _mcp_client.call_tool("browser_snapshot", {})
    return str(result)


@tool
async def browser_take_screenshot(filename: Optional[str] = None) -> str:
    """
    Take a screenshot of the current page.

    Args:
        filename: Optional filename to save screenshot

    Returns:
        HTTP URL to access the screenshot
    """
    args = {}
    if filename:
        args["filename"] = filename

    result = await _mcp_client.call_tool("browser_take_screenshot", args)
    file_path = str(result)

    # Convert file path to HTTP URL
    # e.g., /tmp/playwright-mcp-output/123/screenshot.png -> http://localhost:8000/screenshots/123/screenshot.png
    if "/tmp/playwright-mcp-output/" in file_path:
        relative_path = file_path.replace("/tmp/playwright-mcp-output/", "")
        http_url = f"http://localhost:8000/screenshots/{relative_path}"
        return http_url

    return file_path


@tool
async def browser_wait_for(
    time: Optional[float] = None, text: Optional[str] = None
) -> str:
    """
    Wait for a condition.

    Args:
        time: Seconds to wait
        text: Text to wait for to appear

    Returns:
        Success message
    """
    args = {}
    if time is not None:
        args["time"] = time
    if text is not None:
        args["text"] = text

    result = await _mcp_client.call_tool("browser_wait_for", args)
    return str(result)


# Export tools list
PLAYWRIGHT_MCP_TOOLS = [
    browser_navigate,
    browser_click,
    browser_type,
    browser_snapshot,
    browser_take_screenshot,
    browser_wait_for,
]


async def init_playwright_mcp():
    """Initialize Playwright MCP client on startup."""
    await _mcp_client.initialize()


async def cleanup_playwright_mcp():
    """Cleanup Playwright MCP client on shutdown."""
    await _mcp_client.close()
