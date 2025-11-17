"""Scraper tools for web automation and data extraction."""

import asyncio
import logging
from typing import Any, Dict, List, Optional

import requests
from bs4 import BeautifulSoup
from langchain_core.tools import tool
from langchain_openai import ChatOpenAI
from playwright.async_api import async_playwright, Browser, Page

logger = logging.getLogger(__name__)


@tool
async def scrape_static_page(
    url: str, selectors: Optional[Dict[str, str]] = None
) -> dict:
    """
    Traditional web scraping using requests + BeautifulSoup.

    Best for static HTML pages without JavaScript rendering.

    Args:
        url: Target URL to scrape
        selectors: Optional dict of {label: css_selector} to extract specific data

    Returns:
        {
            "success": bool,
            "data": {extracted data},
            "html_preview": "first 500 chars of HTML",
            "error": str | None
        }
    """
    try:
        logger.info(f"Scraping static page: {url}")

        # Fetch the page
        response = requests.get(url, timeout=10)
        response.raise_for_status()

        # Parse HTML
        soup = BeautifulSoup(response.text, "lxml")

        # Extract data based on selectors
        extracted_data = {}

        if selectors:
            for label, selector in selectors.items():
                elements = soup.select(selector)
                if elements:
                    if len(elements) == 1:
                        extracted_data[label] = elements[0].get_text(strip=True)
                    else:
                        extracted_data[label] = [
                            el.get_text(strip=True) for el in elements
                        ]
                else:
                    extracted_data[label] = None
        else:
            # Extract common elements if no selectors provided
            extracted_data = {
                "title": soup.title.string if soup.title else None,
                "headings": [
                    h.get_text(strip=True)
                    for h in soup.find_all(["h1", "h2", "h3"])[:5]
                ],
                "paragraphs": [p.get_text(strip=True) for p in soup.find_all("p")[:3]],
            }

        logger.info(f"Successfully scraped {len(extracted_data)} data points")

        return {
            "success": True,
            "data": extracted_data,
            "html_preview": response.text[:500],
            "error": None,
        }

    except Exception as e:
        logger.error(f"Error scraping static page: {e}")
        return {"success": False, "data": {}, "html_preview": None, "error": str(e)}


# Export all tools (kept for backwards compatibility, but now empty)
# Use PLAYWRIGHT_MCP_TOOLS from mcp_playwright.py instead
SCRAPER_TOOLS = [scrape_static_page]
